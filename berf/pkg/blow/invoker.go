package blow

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/bingoohuang/ngg/berf"
	"github.com/bingoohuang/ngg/berf/pkg/blow/internal"
	"github.com/bingoohuang/ngg/gnet"
	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/tick"
	"github.com/bingoohuang/ngg/tsid"
	"github.com/spf13/pflag"
	"github.com/thoas/go-funk"
	"github.com/valyala/fasthttp"
)

type Invoker struct {
	pieArg     HttpieArg
	printLock  sync.Locker
	httpHeader *fasthttp.RequestHeader
	pieBody    *HttpieArgBody
	opt        *Opt
	uploadChan chan *internal.UploadChanValue

	httpInvoke      func(*fasthttp.Request, *fasthttp.Response) error
	uploadFileField string
	upload          string
	requestUriExpr  ss.Subs
	writeBytes      int64
	readBytes       int64
	isTLS           bool
	uploadCache     bool

	logChan         chan LogEntry
	logDealerDone   chan struct{}
	logSamplingFunc func() bool
}

type LogEntry struct {
	Log string
}

func NewInvoker(ctx context.Context, opt *Opt) (*Invoker, error) {
	r := &Invoker{
		opt:           opt,
		logDealerDone: make(chan struct{}),
	}
	if opt.logSamplingRate > 0 {
		logSamplingRate := opt.logSamplingRate
		r.logSamplingFunc = func() func() bool {
			return func() bool {
				return rand.Float64() <= logSamplingRate
			}
		}()

		r.logChan = make(chan LogEntry, 1000)
		go r.dealLog()
	}
	r.printLock = NewConditionalLock(r.opt.printOption > 0)

	header, err := r.buildRequestClient(ctx, opt)
	if err != nil {
		return nil, err
	}

	if header != nil {
		requestURI := string(header.RequestURI())
		if opt.eval {
			r.requestUriExpr = ss.ParseExpr(requestURI)
		}
		r.httpHeader = header
	}

	if opt.upload != "" {
		const cacheTag = ":cache"
		if strings.HasSuffix(opt.upload, cacheTag) {
			r.uploadCache = true
			opt.upload = strings.TrimSuffix(opt.upload, cacheTag)
		}

		if pos := strings.IndexRune(opt.upload, ':'); pos > 0 {
			r.uploadFileField = opt.upload[:pos]
			r.upload = opt.upload[pos+1:]
		} else {
			r.uploadFileField = "file"
			r.upload = opt.upload
		}
	}

	if r.upload != "" {
		uploadReader := internal.CreateFileReader(r.uploadFileField, r.upload, r.opt.saveRandDir, r.opt.ant)
		r.uploadChan = make(chan *internal.UploadChanValue)
		go internal.DealUploadFilePath(ctx, uploadReader, r.uploadChan, r.uploadCache)
	}

	return r, nil
}

func (r *Invoker) buildRequestClient(ctx context.Context, opt *Opt) (*fasthttp.RequestHeader, error) {
	var u *url.URL
	var err error

	switch {
	case len(opt.urls) > 0:
		r.opt.parsedUrls = make([]*url.URL, len(opt.urls))
		for i, optURL := range opt.urls {
			rr, err := gnet.FixURI{}.Fix(optURL)
			if err == nil {
				r.opt.parsedUrls[i] = rr
			} else {
				return nil, err
			}
		}

		u = r.opt.parsedUrls[0]
	case len(opt.profiles) > 0:
		u, err = url.Parse(opt.profiles[0].URL)
	default:
		if opt.berfConfig.Features.IsNop() {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to parse url")
	}

	if err != nil {
		return nil, err
	}

	usingTLCP := u.Scheme == "https" && ss.Must(ss.GetenvBool("TLCP", false))
	if usingTLCP {
		u.Scheme = "http"
		if u.Port() == "" {
			u.Host += ":443"
		}
	}

	r.isTLS = u.Scheme == "https"

	cli := &fasthttp.Client{
		Name:         "blow",
		ReadTimeout:  opt.readTimeout,
		WriteTimeout: opt.writeTimeout,
		// 在从主机连接池获取连接时，总是优先创建新的链接，直到 MaxGreedyConnsPerHost 为止
		MaxGreedyConnsPerHost: ss.Must(ss.Getenv[int]("MAX_GREEDY_CONNS_PER_HOST", 0)),
		// 主机最大连接池大小
		MaxConnsPerHost: ss.Must(ss.Getenv[int]("MAX_CONNS_PER_HOST", fasthttp.DefaultMaxConnsPerHost)),
		// 最大连接空闲时间（超时会被回收）
		MaxIdleConnDuration: ss.Must(tick.Getenv("MAX_IDLE_CONN_DURATION", fasthttp.DefaultMaxIdleConnDuration)),
	}

	if cli.MaxConnsPerHost < cli.MaxGreedyConnsPerHost {
		cli.MaxConnsPerHost = cli.MaxGreedyConnsPerHost
	}

	cli.Dial = ProxyHTTPDialerTimeout(opt.dialTimeout, dialer, r.isTLS)

	wrap := internal.NetworkWrap(opt.network)
	cli.Dial = internal.ThroughputStatDial(wrap, cli.Dial, &r.readBytes, &r.writeBytes)
	if usingTLCP {
		cli.Dial = createTlcpDialer(ctx, cli.Dial, r.opt.certPath, r.opt.HasPrintOption, r.opt.tlsVerify)
	}

	if cli.TLSConfig, err = opt.buildTLSConfig(); err != nil {
		return nil, err
	}

	r.pieArg = parseHttpieLikeArgs(pflag.Args())

	var h fasthttp.RequestHeader

	host := ""
	contentType := ""
	for _, hdr := range opt.headers {
		k, v := ss.Split2(hdr, ":")
		if strings.EqualFold(k, "Host") {
			host = v
		} else if strings.EqualFold(k, "Content-Type") {
			contentType = v
		}
	}

	h.SetContentType(adjustContentType(opt, contentType))
	if host != "" {
		h.SetHost(host)
	}

	opt.headers = funk.FilterString(opt.headers, func(hdr string) bool {
		k, _ := ss.Split2(hdr, ":")
		return !strings.EqualFold(k, "Host") && !strings.EqualFold(k, "Content-Type")
	})

	method := detectMethod(opt, r.pieArg)
	h.SetMethod(method)
	r.pieBody = r.pieArg.Build(method, opt.form)

	query := u.Query()
	for _, v := range r.pieArg.query {
		query.Add(v.V1, v.V2)
	}

	if method == "GET" {
		for k, v := range r.pieArg.param {
			query.Add(k, v)
		}
	}

	u.RawQuery = query.Encode()
	h.SetRequestURI(u.RequestURI())

	h.Set("Accept", "application/json")
	for k, v := range r.pieArg.header {
		h.Set(k, v)
	}
	for _, hdr := range opt.headers {
		h.Set(ss.Split2(hdr, ":"))
	}

	if opt.auth != "" {
		b := opt.auth
		if c := ss.Base64().Decode(b); c.V2 != nil { // check if it is already set by base64 encoded
			b = ss.Base64().Encode(b).V1.String()
		} else {
			b = ss.Base64().Encode(c.V1.String()).V1.String()
		}

		h.Set("Authorization", "Basic "+b)
	}
	if r.opt.enableGzip {
		h.Set("Accept-Encoding", "gzip")
	}
	if r.opt.noKeepalive {
		h.Set("Connection", "close")
	}

	if r.opt.doTimeout == 0 {
		r.httpInvoke = cli.Do
	} else {
		r.httpInvoke = func(req *fasthttp.Request, rsp *fasthttp.Response) error {
			return cli.DoTimeout(req, rsp, r.opt.doTimeout)
		}
	}

	// 内部状态 http 监听地址, e.g. :2888
	if stateAddr := os.Getenv("STATE_ADDR"); stateAddr != "" {
		mux := http.NewServeMux()
		mux.HandleFunc("/stat", func(w http.ResponseWriter, r *http.Request) {
			stat := cli.Stat()
			json.NewEncoder(w).Encode(stat)
		})

		go func() {
			// 启动 HTTP 服务器
			log.Printf("Starting server on %s\n", stateAddr)
			if err := http.ListenAndServe(stateAddr, mux); err != nil {
				log.Fatal("ListenAndServe error: ", err)
			}
		}()
	}

	return &h, nil
}

var envHosts = func() []string {
	hosts := os.Getenv("HOSTS")
	return ss.Split(hosts, ",")
}()

var envHostsIndex uint32

func (r *Invoker) Run(ctx context.Context, bc *berf.Config, initial bool) (*berf.Result, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	if bc.N == 1 {
		r.opt.printOption = printAll
		req.ConnAcquiredCallback = func(conn net.Conn) {
			printConnectState(conn)
		}
	} else {
		req.ConnAcquiredCallback = nil
	}

	if len(r.opt.profiles) > 0 {
		return r.runProfiles(req, resp, initial)
	}

	if initial {
		return nil, nil
	}

	r.setReq(req)

	return r.runOne(req, resp)
}

func (r *Invoker) setReq(req *fasthttp.Request) {
	r.httpHeader.CopyTo(&req.Header)

	if len(r.opt.parsedUrls) > 0 {
		atomicIdx := atomic.AddUint32(&envHostsIndex, 1)
		idx := (int(atomicIdx) - 1) % len(r.opt.parsedUrls) // 从 0 开始
		u := r.opt.parsedUrls[idx]
		req.Header.SetHost(u.Host)
		req.URI().SetHostBytes(req.Header.Host())
		req.Header.SetRequestURI(u.RequestURI())
	} else if len(envHosts) > 0 {
		atomicIdx := atomic.AddUint32(&envHostsIndex, 1)
		idx := (int(atomicIdx) - 1) % len(envHosts) // 从 0 开始
		req.URI().SetHost(envHosts[idx])
		req.Header.SetHost(envHosts[idx])
	}

	if len(r.requestUriExpr) > 0 && r.requestUriExpr.CountVars() > 0 {
		result, _ := r.requestUriExpr.Eval(internal.Valuer)
		if v, ok := result.(string); ok {
			req.SetRequestURI(v)
		}
	}

	if r.isTLS {
		req.URI().SetScheme("https")
	}

	if host := r.httpHeader.Host(); len(host) > 0 {
		req.UseHostHeader = true
		req.Header.SetHostBytes(host)
	}

	if proxyURL, _ := proxyFunc(string(req.URI().Host()), r.isTLS); proxyURL != nil {
		req.UsingProxy = true
	}
}

func (r *Invoker) runOne(req *fasthttp.Request, resp *fasthttp.Response) (*berf.Result, error) {
	closers, err := r.setBody(req)
	if err != nil {
		return nil, err
	}

	defer ss.Close(closers)

	rr := &berf.Result{}
	err = r.doRequest(req, resp, rr)
	r.updateThroughput(rr)

	return rr, err
}

func (r *Invoker) logSample(line string, samplingYes bool) {
	if samplingYes {
		r.logChan <- LogEntry{Log: line}
	}
}

func (r *Invoker) log(line string) bool {
	if r.logSamplingFunc != nil && r.logSamplingFunc() {
		r.logChan <- LogEntry{Log: line}
		return true
	}
	return false
}
func (r *Invoker) updateThroughput(rr *berf.Result) {
	rr.ReadBytes = atomic.SwapInt64(&r.readBytes, 0)
	rr.WriteBytes = atomic.SwapInt64(&r.writeBytes, 0)
}

func (r *Invoker) doRequest(req *fasthttp.Request, rsp *fasthttp.Response, rr *berf.Result) (err error) {
	t1 := time.Now()
	err = r.httpInvoke(req, rsp)
	rr.Cost = time.Since(t1)
	if err != nil {
		return err
	}

	err = r.processRsp(req, rsp, rr, nil, "", nil, nil)
	return err
}

var seqFn = func() func() int64 {
	var seq atomic.Int64
	return func() int64 {
		return seq.Add(1)
	}
}()

func (r *Invoker) processRsp(req *fasthttp.Request, rsp *fasthttp.Response, rr *berf.Result,
	p *internal.Profile, tracerID string,
	resultMap *map[string]string, asserts map[string]string) error {

	status := parseStatus(rsp, r.opt.statusName)
	rr.Status = append(rr.Status, status)
	if r.opt.verbose >= 1 {
		rr.Counting = append(rr.Counting, rsp.LocalAddr().String()+"->"+rsp.RemoteAddr().String())
	}

	if r.opt.printOption == 0 && r.logSamplingFunc == nil && p == nil {
		return rsp.BodyWriteTo(io.Discard)
	}

	b1 := &bytes.Buffer{}

	conn := rsp.LocalAddr().String() + "->" + rsp.RemoteAddr().String()
	traceInfo := ""
	if tracerID != "" {
		traceInfo = "[RunID: " + tracerID + "]\n"
	}

	summary := fmt.Sprintf("### #%d %s 时间: %s 耗时: %s  读/写: %d/%d 字节\n%s",
		seqFn(), conn, time.Now().Format(time.RFC3339Nano), rr.Cost, r.readBytes, r.writeBytes, traceInfo)

	_, _ = b1.WriteString(summary)

	bw := bufio.NewWriter(b1)
	_ = req.Header.Write(bw)
	_ = req.BodyWriteTo(bw)
	_ = bw.Flush()

	if p != nil {
		var reqBody bytes.Buffer
		req.BodyWriteTo(&reqBody)
		if reqBody.Len() > 0 {
			p.SourceJSONValuer(reqBody.Bytes(), resultMap)
		}
	}

	r.printLock.Lock()
	defer r.printLock.Unlock()

	h := &req.Header
	ignoreBody := h.IsGet() || h.IsHead()
	statusCode := rsp.StatusCode()

	samplingYes := r.log(b1.String())
	r.printReq(b1, io.Discard, ignoreBody, statusCode, status, samplingYes)
	b1.Reset()

	header := rsp.Header.Header()
	_, _ = b1.Write(header)
	bb1 := b1

	var body *bytes.Buffer

	contentType := rsp.Header.Peek("Content-Type")
	if p != nil && bytes.HasPrefix(contentType, []byte("application/json")) {
		body = &bytes.Buffer{}
		bb1 = body
	}

	d, err := rsp.BodyUncompressed()
	if err != nil {
		return err
	}
	bb1.Write(d)

	if p != nil && body != nil {
		i := body.Bytes()
		p.ResultJSONValuer(i, resultMap)
		b1.Write(i)
	}

	var assertResult []string
	for k1, k2 := range asserts {
		switch {
		case strings.HasPrefix(k1, "eq."):
			k1 = k1[3:]
			v1 := (*resultMap)[k1]
			v2 := (*resultMap)[k2]
			if v1 == v2 {
				assertResult = append(assertResult, "ASSERT OK: $"+k1+" == $"+k2)
			} else {
				assertResult = append(assertResult, "ASSERT FAIL: $"+k1+" != $"+k2)
				rr.AssertFail++
			}
		case strings.HasPrefix(k1, "ne."):
			k1 = k1[3:]
			v1 := (*resultMap)[k1]
			v2 := (*resultMap)[k2]
			if v1 != v2 {
				assertResult = append(assertResult, "ASSERT OK: $"+k1+" != $"+k2)
			} else {
				assertResult = append(assertResult, "ASSERT FAIL: $"+k1+" == $"+k2)
				rr.AssertFail++
			}
		}
	}

	if len(assertResult) > 0 {
		traceInfo += "[" + strings.Join(assertResult, "]\n[") + "]"
	}

	if traceInfo != "" {
		r.logSample(traceInfo+"\n"+b1.String(), samplingYes)
	} else {
		r.logSample(b1.String(), samplingYes)
	}

	r.printResp(b1, io.Discard, rsp, statusCode, status, samplingYes, traceInfo)

	return nil
}

func (r *Invoker) printReq(b *bytes.Buffer, bx io.Writer, ignoreBody bool, statusCode int, status string, samplingYes bool) {
	if !logStatus(r.opt.berfConfig.N, statusCode, status, r.opt.printOption) {
		bx = nil
	}

	if bx == nil && (r.opt.printOption == 0 || !samplingYes) {
		return
	}

	dumpHeader, dumpBody := r.dump(b, bx, ignoreBody, nil)
	if bx != nil {
		_, _ = bx.Write([]byte("\n"))
	}

	printNum := 0
	if r.opt.HasPrintOption(printReqHeader) {
		fmt.Println(ColorfulHeader(string(dumpHeader)))
		printNum++
	}
	if r.opt.HasPrintOption(printReqBody) {
		if strings.TrimSpace(string(dumpBody)) != "" {
			printBody(dumpBody, printNum, r.opt.pretty)
			printNum++
		}
	}

	if printNum > 0 {
		fmt.Println()
	}
}

var logStatus = func() func(n, code int, customizedStatus string, printOption uint8) bool {
	if env := os.Getenv("BLOW_STATUS"); env != "" {
		excluded := ss.HasPrefix(env, "-")
		if excluded {
			env = env[1:]
		}
		return func(n, code int, customizedStatus string, printOption uint8) bool {
			if n == 1 || printOption > 0 {
				return true
			}
			if excluded {
				return customizedStatus != env
			}
			return customizedStatus == env
		}
	}

	return func(n, code int, status string, printOption uint8) bool {
		if n == 1 || printOption > 0 {
			return true
		}
		return code < 200 || code >= 300
	}
}()

func (r *Invoker) printResp(b *bytes.Buffer, bx io.Writer, rsp *fasthttp.Response, statusCode int, status string, samplingYes bool, traceInfo string) {
	if !logStatus(r.opt.berfConfig.N, statusCode, status, r.opt.printOption) {
		bx = nil
	}

	if bx == nil && (r.opt.printOption == 0 || !samplingYes) {
		return
	}

	dumpHeader, dumpBody := r.dump(b, bx, false, rsp)

	printNum := 0
	if traceInfo != "" {
		fmt.Println(traceInfo)
		printNum++
	}
	if r.opt.HasPrintOption(printRespStatusCode) && !r.opt.HasPrintOption(printRespHeader) {
		fmt.Println(Color(strconv.Itoa(rsp.StatusCode()), Magenta))
		printNum++
	}
	if r.opt.HasPrintOption(printRespHeader) {
		fmt.Println(ColorfulHeader(string(dumpHeader)))
		printNum++
	}
	if r.opt.HasPrintOption(printRespBody) {
		if string(dumpBody) != "\r\n" {
			printBody(dumpBody, printNum, r.opt.pretty)
			printNum++
		}
	}
	if printNum > 0 && !r.opt.HasPrintOption(printRespStatusCode) {
		fmt.Println()
	}
}

var cLengthReg = regexp.MustCompile(`Content-Length: (\d+)`)

func (r *Invoker) dump(b *bytes.Buffer, bx io.Writer, ignoreBody bool, rsp *fasthttp.Response) (dumpHeader, dumpBody []byte) {
	dump := b.String()
	dps := strings.Split(dump, "\n")
	for i, line := range dps {
		if len(strings.Trim(line, "\r\n ")) == 0 {
			dumpHeader = []byte(strings.Join(dps[:i], "\n"))
			dumpBody = []byte("\n" + strings.Join(dps[i:], "\n"))
			break
		}
	}

	if bx != nil {
		_, _ = bx.Write(dumpHeader)
	}
	cl := -1
	if subs := cLengthReg.FindStringSubmatch(string(dumpHeader)); len(subs) > 0 {
		cl, _ = ss.Parse[int](subs[1])
	}

	var filename string
	if rsp != nil && r.opt.downloadDir != "" {
		// Content-Disposition: inline; filename=20230529.00475.jpg
		if d := string(rsp.Header.Peek("Content-Disposition")); d != "" {
			_, params, err := mime.ParseMediaType(d)
			if err != nil {
				log.Printf("parse media type %s: %v", d, err)
			} else {
				filename = params["filename"]
			}
		}
		if filename != "" {
			filename = filepath.Join(r.opt.downloadDir, filename)
		}
	}

	if filename != "" {
		if err := os.WriteFile(filename, rsp.Body(), os.ModePerm); err != nil {
			log.Printf("write file %s: %v", filename, err)
		}
		dumpBody = []byte(fmt.Sprintf("\n\n--- %s downloaded ---\n", filename))
		if bx != nil {
			_, _ = bx.Write(dumpBody)
		}
		return dumpHeader, dumpBody
	}

	maxBody := 4096
	if envValue := os.Getenv("MAX_BODY"); envValue != "" {
		if bytesValue, err := ss.ParseBytes(envValue); err == nil {
			maxBody = int(bytesValue)
		} else {
			log.Printf("bad environment value format: %s", envValue)
		}
	}
	if !ignoreBody && (cl == 0 || (maxBody > 0 && cl > maxBody)) {
		dumpBody = []byte("\n\n--- streamed or too long, ignored ---\n")
	}

	if bx != nil {
		_, _ = bx.Write(dumpBody)
	}
	return dumpHeader, dumpBody
}

func printBody(dumpBody []byte, printNum int, pretty bool) {
	if printNum > 0 {
		fmt.Println()
	}

	body := formatResponseBody(dumpBody, pretty, berf.IsStdoutTerminal)
	body = strings.TrimFunc(body, unicode.IsSpace)
	fmt.Println(body)
}

func (r *Invoker) runProfiles(req *fasthttp.Request, rsp *fasthttp.Response, initial bool) (*berf.Result, error) {
	rr := &berf.Result{}
	defer r.updateThroughput(rr)

	profiles := r.opt.profiles
	if initial {
		initProfiles := make([]*internal.Profile, 0, len(profiles))
		nonInitial := make([]*internal.Profile, 0, len(profiles))
		for _, p := range r.opt.profiles {
			if p.Init {
				initProfiles = append(initProfiles, p)
			} else {
				nonInitial = append(nonInitial, p)
			}
		}
		profiles = initProfiles
		r.opt.profiles = nonInitial
	}

	resultMap := map[string]string{}
	traceId := tsid.Fast().ToString()
	for _, p := range profiles {
		if err := r.runOneProfile(p, &resultMap, req, rsp, rr, traceId); err != nil {
			return rr, err
		}

		if code := rsp.StatusCode(); code < 200 || code > 300 {
			break
		}

		req.Reset()
		rsp.Reset()
	}

	return rr, nil
}

func (r *Invoker) runOneProfile(p *internal.Profile, resultMap *map[string]string,
	req *fasthttp.Request, rsp *fasthttp.Response, rr *berf.Result, tracerID string) error {
	closers, err := p.CreateReq(r.isTLS, req, r.opt.enableGzip, r.opt.uploadIndex, *resultMap)
	defer ss.Close(closers)

	if err != nil {
		return err
	}

	t1 := time.Now()
	err = r.httpInvoke(req, rsp)
	rr.Cost += time.Since(t1)
	if err != nil {
		return err
	}

	return r.processRsp(req, rsp, rr, p, tracerID, resultMap, p.Asserts)
}

func parseStatus(rsp *fasthttp.Response, statusName string) string {
	if statusName != "" {
		if d, err := rsp.BodyUncompressed(); err == nil {
			if status := jj.GetBytes(d, statusName).String(); status != "" {
				return statusName + " " + status
			}
		}
	}

	return "HTTP " + strconv.Itoa(rsp.StatusCode())
}

func (r *Invoker) setBody(req *fasthttp.Request) (internal.Closers, error) {
	if r.opt.bodyStreamFile != "" {
		file, err := os.Open(r.opt.bodyStreamFile)
		if err != nil {
			return nil, err
		}
		req.SetBodyStream(file, -1)
		return []io.Closer{file}, nil
	}

	if r.upload != "" {
		uv, ok := <-r.uploadChan
		if !ok {
			return nil, io.EOF
		}
		data := uv.Data()
		multi := data.CreateFileField(r.uploadFileField, r.opt.uploadIndex)
		for k, v := range multi.Headers {
			internal.SetHeader(req, k, v)
		}
		req.Header.Set("Beefs-Original", data.Payload.Original)
		req.SetBodyStream(multi.NewReader(), int(multi.Size))
		return nil, nil
	}

	bodyBytes := r.opt.bodyBytes
	if len(bodyBytes) == 0 && r.pieBody.BodyString != "" {
		internal.SetHeader(req, "Content-Type", r.pieBody.ContentType)
		bodyBytes = []byte(r.pieBody.BodyString)
	}

	if len(bodyBytes) == 0 && r.opt.bodyLinesChan != nil {
		line, ok := <-r.opt.bodyLinesChan
		if !ok { // lines is read over
			return nil, io.EOF
		}
		bodyBytes = []byte(line)
	}

	if len(bodyBytes) > 0 && r.opt.eval {
		bodyBytes = []byte(internal.Gen(string(bodyBytes), internal.MayJSON))
	}

	if len(bodyBytes) > 0 {
		if r.opt.enableGzip {
			if d, err := ss.Gzip(bodyBytes); err == nil && len(d) < len(bodyBytes) {
				bodyBytes = d
				req.Header.Set("Content-Encoding", "gzip")
			}
		}
	} else if r.pieBody.Body != nil {
		internal.SetHeader(req, "Content-Type", r.pieBody.ContentType)
		req.SetBodyStream(r.pieBody.Body, -1)
		return nil, nil
	}

	req.SetBodyRaw(bodyBytes)
	return nil, nil
}

func (r *Invoker) waitLog() {
	if r.logSamplingFunc != nil {
		close(r.logChan)
		<-r.logDealerDone
	}
}
func (r *Invoker) dealLog() {
	f, err := os.CreateTemp("", "blow_"+time.Now().Format(`20060102150405`)+"_"+"*.log")
	ss.ExitIfErr(err)

	fmt.Printf("Log details to: %s\n", f.Name())

	if err != nil {
		log.Fatalf("create debug file error: %v", err)
	}
	defer f.Close()

	for logEntry := range r.logChan {
		line := logEntry.Log + "\n\n"
		f.Write([]byte(line))
	}

	r.logDealerDone <- struct{}{}
}

func adjustContentType(opt *Opt, contentType string) string {
	if contentType != "" {
		return contentType
	}

	if opt.jsonBody || json.Valid(opt.bodyBytes) {
		return `application/json; charset=utf-8`
	}

	return `plain/text; charset=utf-8`
}

func detectMethod(opt *Opt, arg HttpieArg) string {
	if opt.method != "" {
		return opt.method
	}

	if opt.MaybePost() || arg.MaybePost() {
		return "POST"
	}

	return "GET"
}

type noSessionCache struct{}

func (noSessionCache) Get(string) (*tls.ClientSessionState, bool) { return nil, false }
func (noSessionCache) Put(string, *tls.ClientSessionState)        { /* no-op */ }

func (o *Opt) buildTLSConfig() (*tls.Config, error) {
	var certs []tls.Certificate
	if o.certPath != "" && o.keyPath != "" {
		c, err := tls.LoadX509KeyPair(o.certPath, o.keyPath)
		if err != nil {
			return nil, err
		}
		certs = append(certs, c)
	}

	t := &tls.Config{
		InsecureSkipVerify: !o.tlsVerify,
		Certificates:       certs,
		// 关闭 HTTP 客户端的会话缓存
		SessionTicketsDisabled: o.noTLSessionTickets,
	}

	if cacheSize := ss.Must(ss.Getenv[int](`TLS_SESSION_CACHE`, 32)); cacheSize > 0 {
		t.ClientSessionCache = tls.NewLRUClientSessionCache(cacheSize)
	} else {
		t.ClientSessionCache = &noSessionCache{}
	}

	if o.rootCert != "" {
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(ss.Must(ss.ReadFile(o.rootCert)))
		t.RootCAs = pool
	}

	return t, nil
}
