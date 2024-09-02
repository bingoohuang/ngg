package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http/httptrace"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

// some code is copy from https://github.com/davecheney/httpstat.

func printf(format string, a ...any) {
	_, _ = fmt.Fprintf(color.Output, format, a...)
}

func grayscale(code color.Attribute) func(string, ...any) string {
	return color.New(code + 232).SprintfFunc()
}

type httpStat struct {
	t0, t1, t2, t3, t4, t5, t6 time.Time
	t7                         time.Time // after read body
	t31                        time.Time // WroteRequest
}

func createClientTrace(req *Request) *httptrace.ClientTrace {
	req.stat = &httpStat{}
	return &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { req.stat.t0 = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { req.stat.t1 = time.Now() },
		ConnectStart: func(_, _ string) {
			if req.stat.t1.IsZero() {
				req.stat.t1 = time.Now() // connecting to IP
			}
		},
		ConnectDone: func(net, addr string, err error) {
			if err != nil {
				log.Fatalf("unable to connect to host %v: %v", addr, err)
			}
			req.stat.t2 = time.Now()
		},
		GotConn: func(info httptrace.GotConnInfo) {
			req.stat.t3 = time.Now()
			req.ConnInfo = info
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			req.stat.t31 = time.Now()
		},
		GotFirstResponseByte: func() { req.stat.t4 = time.Now() },
		TLSHandshakeStart:    func() { req.stat.t5 = time.Now() },
		TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { req.stat.t6 = time.Now() },
	}
}

func (stat *httpStat) print(urlSchema string) {
	now := time.Now()
	stat.t7 = now
	if stat.t0.IsZero() { // we skipped DNS
		stat.t0 = stat.t1
	}
	fa := func(b, a time.Time) string {
		return color.CyanString("%7d ms", int(b.Sub(a)/time.Millisecond))
	}
	fb := func(b, a time.Time) string {
		return color.CyanString("%-9s", strconv.Itoa(int(b.Sub(a)/time.Millisecond))+" ms")
	}

	colorize := func(s string) string {
		v := strings.Split(s, "\n")
		v[0] = grayscale(16)(v[0])
		return strings.Join(v, "\n")
	}

	switch urlSchema {
	case "https":
		printf(colorize(httpsTemplate),
			fa(stat.t1, stat.t0),  // dns lookup
			fa(stat.t2, stat.t1),  // tcp connection
			fa(stat.t6, stat.t5),  // tls handshake
			fa(stat.t31, stat.t3), // Request transfer
			fa(stat.t4, stat.t31), // server processing
			fa(stat.t7, stat.t4),  // Response Transfer
			fb(stat.t1, stat.t0),  // namelookup
			fb(stat.t2, stat.t0),  // connect
			fb(stat.t3, stat.t0),  // pretransfer
			fb(stat.t31, stat.t0), // wrote request
			fb(stat.t4, stat.t0),  // starttransfer
			fb(stat.t7, stat.t0),  // total
		)
	case "http":
		printf(colorize(httpTemplate),
			fa(stat.t1, stat.t0),  // dns lookup
			fa(stat.t3, stat.t1),  // tcp connection
			fa(stat.t31, stat.t3), // Request transfer
			fa(stat.t4, stat.t31), // server processing
			fa(stat.t7, stat.t4),  // content transfer
			fb(stat.t1, stat.t0),  // namelookup
			fb(stat.t3, stat.t0),  // connect
			fb(stat.t31, stat.t0), // wrote request
			fb(stat.t4, stat.t0),  // starttransfer
			fb(stat.t7, stat.t0),  // total
		)
	}
}

func isRedirect(statusCode int) bool { return statusCode > 299 && statusCode < 400 }

const (
	httpsTemplate = `` +
		`  DNS Lookup   TCP Connection   TLS Handshake   Request Transfer   Server Processing   Response Transfer` + "\n" +
		`[%s  |    %s  |   %s  |     %s  |       %s  |      %s  ]` + "\n" +
		`             |                |               |                 |                   |                  |` + "\n" +
		` namelookup: %s        |               |                 |                   |                  |` + "\n" +
		`                     connect: %s       |                 |                   |                  |` + "\n" +
		`                                 pretransfer: %s         |                   |                  |` + "\n" +
		`                                                 wrote request: %s           |                  |` + "\n" +
		`                                                                     starttransfer: %s          |` + "\n" +
		`                                                                                                total: %s` + "\n"

	httpTemplate = `` +
		`   DNS Lookup   TCP Connection   Request Transfer   Server Processing   Response Transfer` + "\n" +
		`[ %s  |    %s  |      %s  |      %s  |       %s  ]` + "\n" +
		`              |                |                  |                  |                   |` + "\n" +
		`  namelookup: %s        |                  |                  |                   |` + "\n" +
		`                      connect: %s          |                  |                   |` + "\n" +
		`                                   wrote request: %s          |                   |` + "\n" +
		`                                                      starttransfer: %s           |` + "\n" +
		`                                                                                  total: %s` + "\n"
)
