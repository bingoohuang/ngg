// Package httpretty prints your HTTP requests pretty on your terminal screen.
// You can use this package both on the client-side and on the server-side.
//
// This package provides a better way to view HTTP traffic without httputil
// DumpRequest, DumpRequestOut, and DumpResponse heavy debugging functions.
//
// You can use the logger quickly to log requests you are opening. For example:
//
//	package main
//
//	import (
//		"fmt"
//		"net/http"
//		"os"
//
//		"github.com/bingoohuang/ngg/httpretty"
//	)
//
//	func main() {
//		logger := &httpretty.Logger{
//			Time:           true,
//			TLS:            true,
//			RequestHeader:  true,
//			RequestBody:    true,
//			ResponseHeader: true,
//			ResponseBody:   true,
//			Colors:         true,
//			Formatters:     []httpretty.Formatter{&httpretty.JSONFormatter{}},
//		}
//
//		http.DefaultClient.Transport = logger.RoundTripper(http.DefaultClient.Transport) // tip: you can use it on any *http.Client
//
//		if _, err := http.Get("https://www.google.com/"); err != nil {
//			fmt.Fprintf(os.Stderr, "%+v\n", err)
//			os.Exit(1)
//		}
//	}
//
// If you pass nil to the logger.RoundTripper it is going to fallback to http.DefaultTransport.
//
// You can use the logger quickly to log requests on your server. For example:
//
//	logger := &httpretty.Logger{
//		Time:           true,
//		TLS:            true,
//		RequestHeader:  true,
//		RequestBody:    true,
//		ResponseHeader: true,
//		ResponseBody:   true,
//	}
//
//	logger.Middleware(handler)
//
// Note: server logs don't include response headers set by the server.
// Client logs don't include request headers set by the HTTP client.
package httpretty

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/textproto"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/bingoohuang/ngg/httpretty/httpsnoop"
	"github.com/bingoohuang/ngg/httpretty/internal/color"
)

// Formatter can be used to format body.
//
// If the Format function returns an error, the content is printed in verbatim after a warning.
// Match receives a media type from the Content-Type field. The body is formatted if it returns true.
type Formatter interface {
	Match(mediatype string) bool
	Format(w io.Writer, src []byte) error
}

// WithHide can be used to protect a request from being exposed.
func WithHide(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextHide{}, struct{}{})
}

// Logger provides a way for you to print client and server-side information about your HTTP traffic.
type Logger struct {
	w          io.Writer
	filter     Filter
	skipHeader map[string]struct{}
	bodyFilter BodyFilter

	// Formatters for the request and response bodies.
	// No standard formatters are used. You need to add what you want to use explicitly.
	// We provide a JSONFormatter for convenience (add it manually).
	Formatters []Formatter

	// MaxRequestBody the logger can print.
	// If value is not set and Content-Length is not sent, 4096 bytes is considered.
	MaxRequestBody int64

	// MaxResponseBody the logger can print.
	// If value is not set and Content-Length is not sent, 4096 bytes is considered.
	MaxResponseBody int64

	flusher Flusher

	mu sync.Mutex // ensures atomic writes; protects the following fields
	// SkipRequestInfo avoids printing a line showing the request URI on all requests plus a line
	// containing the remote address on server-side requests.
	SkipRequestInfo bool

	// Time the request began and its duration.
	Time bool

	// TLS information, such as certificates and ciphers.
	// BUG(henvic): Currently, the TLS information prints after the response header, although it
	// should be printed before the request header.
	TLS bool

	// RequestHeader set by the client or received from the server.
	RequestHeader bool

	// RequestBody sent by the client or received by the server.
	RequestBody bool

	// ResponseHeader received by the client or set by the HTTP handlers.
	ResponseHeader bool

	// ResponseBody received by the client or set by the server.
	ResponseBody bool

	// SkipSanitize bypasses sanitizing headers containing credentials (such as Authorization).
	SkipSanitize bool

	// Colors set ANSI escape codes that terminals use to print text in different colors.
	Colors bool

	QPS float64
}

// Filter allows you to skip requests.
//
// If an error happens and you want to log it, you can pass a not-null error value.
type Filter func(req *http.Request) (skip bool, err error)

// BodyFilter allows you to skip printing a HTTP body based on its associated Header.
//
// It can be used for omitting HTTP Request and Response bodies.
// You can filter by checking properties such as Content-Type or Content-Length.
//
// On a HTTP server, this function is called even when no body is present due to
// http.Request always carrying a non-nil value.
type BodyFilter func(h http.Header) (skip bool, err error)

// Flusher defines how logger prints requests.
type Flusher int

// Logger can print without flushing, when they are available, or when the request is done.
const (
	// NoBuffer strategy prints anything immediately, without buffering.
	// It has the issue of mingling concurrent requests in unpredictable ways.
	NoBuffer Flusher = iota

	// OnReady buffers and prints each step of the request or response (header, body) whenever they are ready.
	// It reduces mingling caused by mingling but does not give any ordering guarantee, so responses can still be out of order.
	OnReady

	// OnEnd buffers the whole request and flushes it once, in the end.
	OnEnd
)

// SetFilter allows you to set a function to skip requests.
// Pass nil to remove the filter. This method is concurrency safe.
func (l *Logger) SetFilter(f Filter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.filter = f
}

// SkipHeader allows you to skip printing specific headers.
// This method is concurrency safe.
func (l *Logger) SkipHeader(headers []string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	m := map[string]struct{}{}
	for _, h := range headers {
		m[textproto.CanonicalMIMEHeaderKey(h)] = struct{}{}
	}
	l.skipHeader = m
}

// SetBodyFilter allows you to set a function to skip printing a body.
// Pass nil to remove the body filter. This method is concurrency safe.
func (l *Logger) SetBodyFilter(f BodyFilter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.bodyFilter = f
}

// SetOutput sets the output destination for the logger.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.w = w
}

// SetFlusher sets the flush strategy for the logger.
func (l *Logger) SetFlusher(f Flusher) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flusher = f
}

func (l *Logger) getWriter() io.Writer {
	if l.w == nil {
		return os.Stdout
	}

	return l.w
}

func (l *Logger) getFilter() Filter {
	l.mu.Lock()
	f := l.filter
	defer l.mu.Unlock()
	return f
}

func (l *Logger) getBodyFilter() BodyFilter {
	l.mu.Lock()
	f := l.bodyFilter
	defer l.mu.Unlock()
	return f
}

func (l *Logger) cloneSkipHeader() map[string]struct{} {
	l.mu.Lock()
	skipped := l.skipHeader
	l.mu.Unlock()

	m := map[string]struct{}{}
	for h := range skipped {
		m[h] = struct{}{}
	}

	return m
}

type contextHide struct{}

type roundTripper struct {
	logger     *Logger
	rt         http.RoundTripper
	printReqID bool
	qpsAllow   *qpsAllow
}

// RoundTripper returns a RoundTripper that uses the logger.
func (l *Logger) RoundTripper(rt http.RoundTripper, printReqID bool) http.RoundTripper {
	return roundTripper{
		logger:     l,
		rt:         rt,
		printReqID: printReqID,
		qpsAllow:   createQpsAlloer(l.QPS),
	}
}

// RoundTrip implements the http.RoundTrip interface.
func (r roundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	tripper := r.rt
	if tripper == nil {
		// BUG(henvic): net/http data race condition when the client
		// does concurrent requests using the very same HTTP transport.
		// See Go standard library issue https://golang.org/issue/30597
		tripper = http.RoundTripper(http.DefaultTransport)
	}

	if !r.qpsAllow.Allow() {
		return tripper.RoundTrip(req)
	}

	if hide := req.Context().Value(contextHide{}); hide != nil {
		return tripper.RoundTrip(req)
	}

	l := r.logger
	p := newPrinter(l, r.printReqID)
	defer p.flush()

	if p.checkFilter(req) {
		return tripper.RoundTrip(req)
	}

	if l.Time {
		defer p.printTimeRequest()()
	}
	if !l.SkipRequestInfo {
		p.printRequestInfo(req)
	}
	// Try to get some information from transport
	transport, ok := tripper.(*http.Transport)
	// If proxy is used, then print information about proxy server
	if ok && transport.Proxy != nil {
		proxyUrl, err := transport.Proxy(req)
		if proxyUrl != nil && err == nil {
			p.printf("* Using proxy: %s\n", p.format(color.FgBlue, proxyUrl.String()))
		}
	}

	var tlsClientConfig *tls.Config
	if ok && transport.TLSClientConfig != nil {
		tlsClientConfig = transport.TLSClientConfig
		if tlsClientConfig.InsecureSkipVerify {
			p.printf("* Skipping TLS verification: %s\n",
				p.format(color.FgRed, "connection is susceptible to man-in-the-middle attacks."))
		}
	}
	// Maybe print outgoing TLS information.
	if l.TLS && tlsClientConfig != nil {
		// please remember http.Request.TLS is ignored by the HTTP client.
		p.printOutgoingClientTLS(tlsClientConfig)
	}
	p.printRequest(req)
	defer func() {
		if err != nil {
			p.printf("* %s\n", p.format(color.FgRed, err.Error()))
			if resp == nil {
				return
			}
		}
		if l.TLS {
			p.printTLSInfo(resp.TLS, false)
			p.printTLSServer(req.Host, resp.TLS)
		}
		p.printResponse(resp)
	}()
	return tripper.RoundTrip(req)
}

type qpsAllow struct {
	tick *time.Ticker
}

func (q *qpsAllow) Close() error {
	if q.tick != nil {
		q.tick.Stop()
	}
	return nil
}
func (q *qpsAllow) Allow() bool {
	if q.tick == nil {
		return true
	}

	select {
	case _, ok := <-q.tick.C:
		return ok
	default:
		return false
	}
}

func createQpsAlloer(qps float64) *qpsAllow {
	if qps > 0 {
		return &qpsAllow{
			tick: time.NewTicker(time.Duration(1e6/(qps)) * time.Microsecond),
		}
	}

	return &qpsAllow{}
}

// Middleware for logging incoming requests to a HTTP server.
func (l *Logger) Middleware(next http.Handler, printReqID bool) http.Handler {
	return httpHandler{
		logger:     l,
		next:       next,
		printReqID: printReqID,
		qpsAllow:   createQpsAlloer(l.QPS),
	}
}

type httpHandler struct {
	logger     *Logger
	next       http.Handler
	printReqID bool

	qpsAllow *qpsAllow
}

// ServeHTTP is a middleware for logging incoming requests to a HTTP server.
func (h httpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !h.qpsAllow.Allow() {
		h.next.ServeHTTP(w, req)
		return
	}
	if hide := req.Context().Value(contextHide{}); hide != nil {
		h.next.ServeHTTP(w, req)
		return
	}

	l := h.logger
	p := newPrinter(l, h.printReqID)
	defer p.flush()

	if p.checkFilter(req) {
		h.next.ServeHTTP(w, req)
		return
	}
	if p.logger.Time {
		defer p.printTimeRequest()()
	}
	if !p.logger.SkipRequestInfo {
		p.printRequestInfo(req)
	}
	if p.logger.TLS {
		p.printTLSInfo(req.TLS, true)
		p.printIncomingClientTLS(req.TLS)
	}
	p.printRequest(req)
	rec := &responseRecorder{
		header:          w.Header(),
		statusCode:      http.StatusOK,
		maxReadableBody: l.MaxResponseBody,
		buf:             &bytes.Buffer{},
	}

	hooks := httpsnoop.Hooks{
		WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return func(code int) {
				next(code)
				rec.WriteHeader(code)
			}
		},
		Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
			return func(p []byte) (int, error) {
				n, err := next(p)
				rec.Write(p)
				return n, err
			}
		},
	}
	defer p.printServerResponse(req, rec)
	hw := httpsnoop.Wrap(w, hooks)
	h.next.ServeHTTP(hw, req)
}

// PrintRequest prints a request, even when WithHide is used to hide it.
//
// It doesn't log TLS connection details or request duration.
func (l *Logger) PrintRequest(req *http.Request) {
	p := printer{logger: l}
	if skip := p.checkFilter(req); skip {
		return
	}
	p.printRequest(req)
}

// PrintResponse prints a response.
func (l *Logger) PrintResponse(resp *http.Response) {
	p := printer{logger: l}
	p.printResponse(resp)
}

// JSONFormatter helps you read unreadable JSON documents.
//
// github.com/tidwall/pretty could be used to add colors to it.
// However, it would add an external dependency. If you want, you can define
// your own formatter using it or anything else. See Formatter.
type JSONFormatter struct{}

// jsonTypeRE can be used to identify JSON media types, such as
// application/json or application/vnd.api+json.
//
// Source: https://github.com/cli/cli/blob/63a4319f6caedccbadf1bf0317d70b6f0cb1b5b9/internal/authflow/flow.go#L27
var jsonTypeRE = regexp.MustCompile(`[/+]json($|;)`)

// Match JSON media type.
func (j *JSONFormatter) Match(mediatype string) bool {
	return jsonTypeRE.MatchString(mediatype)
}

// Format JSON content.
func (j *JSONFormatter) Format(w io.Writer, src []byte) error {
	if !json.Valid(src) {
		// We want to get the error of json.checkValid, not unmarshal it.
		// The happy path has been optimized, maybe prematurely.
		if err := json.Unmarshal(src, &json.RawMessage{}); err != nil {
			return err
		}
	}
	// avoiding allocation as we use *bytes.Buffer to store the formatted body before printing
	dst, ok := w.(*bytes.Buffer)
	if !ok {
		// mitigating panic to avoid upsetting anyone who uses this directly
		return errors.New("underlying writer for JSONFormatter must be *bytes.Buffer")
	}
	return json.Indent(dst, src, "", "    ")
}
