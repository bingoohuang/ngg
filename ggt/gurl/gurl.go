package gurl

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/ggt/goup"
	"github.com/bingoohuang/ngg/ggt/gurl/certinfo"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/gnet"
	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/tick"
	"github.com/deatil/go-cryptobin/hash/sm3"
	"github.com/spf13/cobra"
	"github.com/zeebo/blake3"
)

func init() {
	fc := &subCmd{}
	c := &cobra.Command{
		Use:  "gurl",
		Long: "curl in golang",
		RunE: fc.run,
	}
	initFlags(c.Flags())
	c.SetHelpFunc(func(command *cobra.Command, i []string) {
		out := command.OutOrStdout()
		_, _ = out.Write([]byte(help))
	})
	root.AddCommand(c, fc)
}

type subCmd struct{}

const DryRequestURL = `http://dry.run.url`

func (f *subCmd) run(cmd *cobra.Command, args []string) error {
	nonFlagArgs := filter(args)
	if err := createDemoEnvFile(); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}

		return fmt.Errorf("create demo env file: %w", err)
	}

	if RemoveChars(printV, "svtqfdUrCN") == "" {
		printV += "b"
	}

	parsePrintOption(printV)
	freeInnerJSON = HasPrintOption(freeInnerJSONTag)
	ugly = HasPrintOption(printUgly)
	raw = HasPrintOption(printRaw)
	countingItems = HasPrintOption(printCountingItems)
	disableProxy = HasPrintOption(optionDisableProxy)

	pretty = !raw

	if !HasPrintOption(printReqBody) {
		defaultSetting.DumpBody = false
	}

	if len(urls) == 0 {
		urls = []string{DryRequestURL}
	}

	stdin := parseStdin()

	start := time.Now()
	for _, urlAddr := range urls {
		run(len(urls), urlAddr, nonFlagArgs, stdin)
	}

	if HasPrintOption(printVerbose) {
		log.Printf("complete, total cost: %s", time.Since(start))
	}

	return nil
}

func parseStdin() io.Reader {
	if isWindows() {
		return nil
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return os.Stdin
	}

	return nil
}

var uploadFilePb *ProgressBar

func run(totalUrls int, urlAddr string, nonFlagArgs []string, reader io.Reader) {
	if reader != nil && isMethodDefaultGet() {
		method = http.MethodPost
	}

	urlAddr2, _ := Eval(urlAddr)
	u := ss.Must(gnet.FixURI{
		DefaultScheme: ss.If(caFile != "", "https", "http"),
		DefaultPort:   8080,
	}.Fix(urlAddr2))

	addrGen := func() *url.URL { return u }
	if urlAddr2 != urlAddr {
		cnt := 0
		addrGen = func() *url.URL {
			cnt++
			if cnt == 1 {
				return u
			}

			eval, err := Eval(urlAddr)
			if err != nil {
				log.Fatalf("eval %v", err)
			}

			return ss.Must(gnet.FixURI{DefaultScheme: ss.If(caFile != "", "https", "http")}.Fix(eval))
		}
	}
	realURL := u.String()
	req := getHTTP(method, realURL, nonFlagArgs, timeout)

	if auth != "" {
		// check if it is already set by base64 encoded
		if c := ss.Base64().Decode(auth); c.V2 != nil {
			auth = ss.Base64().Encode(auth).V1.String()
		} else {
			auth = ss.Base64().Encode(c.V1.String()).V1.String()
		}

		req.Req.Header.Set("Authorization", "Basic "+auth)
	}

	req.Req = req.Req.WithContext(httptrace.WithClientTrace(req.Req.Context(), createClientTrace(req)))
	setTimeoutRequest(req)

	req.SetTLSClientConfig(createTLSConfig(strings.HasPrefix(realURL, "https://")))
	if proxyURL := parseProxyURL(req.Req); proxyURL != nil {
		if HasPrintOption(printVerbose) {
			log.Printf("Proxy URL: %s", proxyURL)
		}
		req.SetProxy(http.ProxyURL(proxyURL))
	}

	if reader != nil {
		req.BodyCh(readStdin(reader))
	}

	req.BodyFileLines(body)

	thinkerFn := func() {}
	if thinker, _ := tick.ParseThinkTime(think); thinker != nil {
		thinkerFn = func() {
			thinker.Think(true)
		}
	}

	req.SetupTransport()
	req.BuildURL()

	if benchC > 1 { // AB bench
		req.DumpRequest(false)
		RunBench(req, thinkerFn)
		return
	}

	req.DumpRequest(HasAnyPrintOptions(printReqHeader, printReqBody))

	for i := 0; benchN == 0 || i < benchN; i++ {
		if i > 0 {
			req.Reset()

			if confirmNum > 0 && (i+1)%confirmNum == 0 {
				surveyConfirm()
			}

			if benchN == 0 || i < benchN-1 {
				thinkerFn()
			}
		}

		start := time.Now()
		if HasPrintOption(printVerbose) && benchN == 0 {
			log.Printf("N: %d", i+1)
		}
		err := doRequest(req, addrGen)
		if HasPrintOption(printVerbose) && totalUrls > 1 {
			log.Printf("current request cost: %s", time.Since(start))
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

func setTimeoutRequest(req *Request) {
	if req.Timeout > 0 {
		var cancelCtx context.Context
		cancelCtx, req.cancelTimeout = context.WithCancel(context.Background())
		ctx, cancel := context.WithCancel(req.Req.Context())
		req.timeResetCh = make(chan struct{})
		go func() {
			t := time.NewTicker(req.Timeout)
			defer t.Stop()

			for {
				select {
				case <-t.C:
					cancel()
				case <-cancelCtx.Done():
					return
				case <-req.timeResetCh:
					t.Reset(req.Timeout)
				}
			}
		}()
		req.Req = req.Req.WithContext(ctx)
	}
}

func setBody(req *Request) {
	switch {
	case len(uploadFiles) > 0:
		var hasher hash.Hash
		hashAlgo := strings.ToLower(os.Getenv("BEEFS_HASH"))
		switch hashAlgo {
		case "blake3":
			hasher = blake3.New()
		case "sm3":
			hasher = sm3.New()
		case "sha256":
			hasher = sha256.New()
		case "sha512":
			hasher = sha512.New()
		case "md5":
			hasher = md5.New()
		case "":
		default:
			fmt.Printf("unsupported hash algo %s\n", hashAlgo)
		}

		if hasher != nil {
			hashValue, err := HashFile(uploadFiles[0], hasher)
			if err != nil {
				log.Fatal(err)
			}
			req.Header("Beefs-Hash", hashAlgo+":"+base64.RawURLEncoding.EncodeToString(hashValue))

		}
		var fileReaders []io.ReadCloser
		for _, uploadFile := range uploadFiles {
			fileReader, err := goup.CreateChunkReader(uploadFile, 0, 0, 0)
			if err != nil {
				log.Fatal(err)
			}
			fileReaders = append(fileReaders, fileReader)
		}

		uploadFilePb = NewProgressBar(0)
		fields := map[string]any{}

		if len(fileReaders) == 1 {
			fields["file"] = fileReaders[0]
		} else {
			for i, r := range fileReaders {
				name := fmt.Sprintf("file-%d", i+1)
				fields[name] = r
			}
		}

		up := goup.PrepareMultipartPayload(fields)
		pb := &goup.PbReader{Reader: up.Body}
		if uploadFilePb != nil {
			uploadFilePb.SetTotal(up.Size)
			pb.Adder = goup.AdderFn(func(value uint64) {
				uploadFilePb.Add64(int64(value))
			})
		}

		req.BodyAndSize(io.NopCloser(pb), up.Size)
		req.Setting.DumpBody = false

		for hk, hv := range up.Headers {
			req.Header(hk, hv)
		}
	case body != "":
		req.Body(body)
	default:
		req.RefreshBody()
	}
}

func readStdin(stdin io.Reader) func() (string, error) {
	d := json.NewDecoder(stdin)
	d.UseNumber()

	return func() (string, error) {
		var j json.RawMessage
		if err := d.Decode(&j); err != nil {
			return "", err
		}

		return string(j), nil
	}
}

// Proxy Support
func parseProxyURL(req *http.Request) *url.URL {
	if proxy != "" {
		return ss.Must(gnet.FixURI{}.Fix(proxy))
	}

	if disableProxy {
		return nil
	}

	p, err := http.ProxyFromEnvironment(req)
	if err != nil {
		log.Fatal("Environment Proxy Url parse err", err)
	}
	return p
}

var clientSessionCache tls.ClientSessionCache

func init() {
	if cacheSize, _ := ss.Getenv[int](`TLS_SESSION_CACHE`, 32); cacheSize > 0 {
		clientSessionCache = tls.NewLRUClientSessionCache(cacheSize)
	}
}

func createTLSConfig(isHTTPS bool) (tlsConfig *tls.Config) {
	if !isHTTPS {
		return nil
	}

	tlsConfig = &tls.Config{
		InsecureSkipVerify: !ss.Pick1(ss.GetenvBool(`TLS_VERIFY`, false)),
		ClientSessionCache: clientSessionCache,
	}

	if caFile != "" {
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(ss.Must(ss.ReadFile(caFile)))
		tlsConfig.RootCAs = pool
	}

	return tlsConfig
}

func doRequest(req *Request, addrGen func() *url.URL) error {
	if req.bodyCh != nil {
		if err := req.NextBody(); err != nil {
			return err
		}
	} else {
		setBody(req)
	}

	u := addrGen()
	req.url = u.String()

	doRequestInternal(req, u)
	return nil
}

func Stat(name string) (int64, bool, error) {
	if s, err := os.Stat(name); err == nil {
		return s.Size(), true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return 0, false, nil
	} else {
		// file may or may not exist. See err for details.
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		return 0, false, err
	}
}

func parseFileNameFromContentDisposition(header http.Header) (filename string) {
	if d := header.Get("Content-Disposition"); d != "" {
		if _, params, _ := mime.ParseMediaType(d); params != nil {
			return params["filename"]
		}
	}

	return ""
}

func doRequestInternal(req *Request, u *url.URL) {
	if benchN == 0 || benchN > 1 {
		req.Header("Gurl-N", fmt.Sprintf("%d", currentN.Inc()))
	}

	_, pathFile := path.Split(u.Path)
	pathFileSize, pathFileExists, _ := Stat(pathFile)
	if pathFileExists && pathFileSize > 0 {
		req.Header("Range", fmt.Sprintf("bytes=%d-", pathFileSize))
	}

	dl := strings.ToLower(download.String())
	if download.Exists && dl == "" {
		dl = "yes"
	}

	fn := ""

	// 如果URL显示的文件不存在并且携带显式下载命令行参数，则尝试先发送 Head 请求，尝试从中获取文件名，并且尝试断点续传
	if !pathFileExists && (dl == "yes" || dl == "y") {
		originalMethod := req.Req.Method
		req.Req.Method = http.MethodHead
		if res, err := req.Response(); err == nil {
			if fn = parseFileNameFromContentDisposition(res.Header); fn != "" {
				if fileSize, fileExists, _ := Stat(fn); fileExists && fileSize > 0 {
					req.Header("Range", fmt.Sprintf("bytes=%d-", fileSize))
				}
			}
		}
		req.Req.Method = originalMethod
		req.Reset()
	}

	if uploadFilePb != nil {
		fmt.Printf("Uploading %q\n", strings.Join(uploadFiles, "; "))
		uploadFilePb.Set(0)
		uploadFilePb.Start()
	}

	res, err := req.Response()
	if uploadFilePb != nil {
		uploadFilePb.Finish()
		fmt.Println()
	}
	if err != nil {
		log.Fatalf("execute error: %+v", err)
	}

	if processDownload(req, res, pathFileExists, dl, fn, pathFile) {
		return
	}

	// 保证 response body 被 读取并且关闭
	_, _ = req.Bytes()

	if isWindows() {
		printRequestResponseForWindows(req, res)
	} else {
		printRequestResponseForNonWindows(req, res, false)
	}

	if HasPrintOption(printHTTPTrace) {
		req.stat.print(u.Scheme)
	}
}

func processDownload(req *Request, res *http.Response, pathFileExists bool, dl, fn, pathFile string) bool {
	if method == "HEAD" || dl == "no" || dl == "n" {
		return false
	}

	if fn == "" {
		fn = parseFileNameFromContentDisposition(res.Header)
	}

	fnFromContentDisposition := fn != ""

	clh := res.Header.Get("Content-Length")
	cl, _ := strconv.ParseInt(clh, 10, 64)

	if clh != "" && cl == 0 {
		return false
	}

	ct := res.Header.Get("Content-Type")

	if dl == "" && pathFileExists {
		dl = "yes"
	}

	if (dl == "yes" || dl == "y") ||
		(cl > 2048 || fn != "" || !ss.ContainsFold(ct, "json", "text", "xml")) {
		if fn == "" {
			fn = pathFile
		}

		if !fnFromContentDisposition {
			switch {
			case ss.ContainsFold(ct, "json") && !ss.HasSuffix(fn, ".json"):
				fn = ss.If(ugly, "", fn+".json")
			case ss.ContainsFold(ct, "text") && !ss.HasSuffix(fn, ".txt"):
				fn = ss.If(ugly, "", fn+".txt")
			case ss.ContainsFold(ct, "xml") && !ss.HasSuffix(fn, ".xml"):
				fn = ss.If(ugly, "", fn+".xml")
			}
		}
		if fn != "" {
			downloadFile(req, res, fn)
			return true
		}
	}

	return false
}

var hasStdoutDevice = func() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeDevice == os.ModeDevice
}()

func printRequestResponseForNonWindows(req *Request, res *http.Response, download bool) {
	var dumpHeader, dumpBody []byte
	dps := strings.Split(string(req.reqDump), "\n")
	for i, line := range dps {
		if len(strings.Trim(line, "\r\n ")) == 0 {
			dumpHeader = []byte(strings.Join(dps[:i], "\n"))
			dumpBody = []byte(strings.Join(dps[i:], "\n"))
			break
		}
	}

	if HasPrintOption(printReqSession) && req.ConnInfo.Conn != nil {
		i := req.ConnInfo
		connSession := fmt.Sprintf("%s->%s (reused: %t, wasIdle: %t, idle: %s)",
			i.Conn.LocalAddr(), i.Conn.RemoteAddr(), i.Reused, i.WasIdle, i.IdleTime)
		fmt.Println(Color("Conn-Session:", Magenta), Color(connSession, Yellow))
	}
	if HasPrintOption(printReqHeader) {
		fmt.Println(ColorfulRequest(string(dumpHeader)))
		fmt.Println()
	} else if HasPrintOption(printReqURL) {
		fmt.Println(Color(req.Req.URL.String(), Green))
	}

	if HasPrintOption(printReqBody) {
		if !saveTempFile(dumpBody, MaxPayloadSize, ugly) {
			fmt.Println(formatBytes(dumpBody, pretty, ugly, freeInnerJSON))
		}
	}

	if !req.DryRequest {
		influxDB := false
		for k := range res.Header {
			if strings.Contains(k, "X-Influxdb-") {
				influxDB = true
				break
			}
		}

		if HasPrintOption(printRspHeader) {
			fmt.Println(Color(res.Proto, Magenta), Color(res.Status, Green))
			for k, val := range res.Header {
				fmt.Printf("%s: %s\n", Color(k, Gray), Color(strings.Join(val, " "), Cyan))
			}

			// Checks whether chunked is part of the encodings stack
			if chunked(res.TransferEncoding) {
				fmt.Printf("%s: %s\n", Color("Transfer-Encoding", Gray), Color("chunked", Cyan))
			}
			if res.Close {
				fmt.Printf("%s: %s\n", Color("Connection", Gray), Color("Close", Cyan))
			}

			fmt.Println()
		} else if HasPrintOption(printRspCode) {
			fmt.Println(Color(res.Status, Green))
		}

		if !download && HasPrintOption(printRspBody) {
			fmt.Println(formatResponseBody(req, pretty, ugly, freeInnerJSON, influxDB))
		}
	}
}

func printTLSConnectState(conn net.Conn, state tls.ConnectionState) {
	if !HasPrintOption(printRspOption) {
		return
	}

	fmt.Printf("option Conn Type: %T\n", conn)
	fmt.Printf("option TLS.Version: %s\n", func(version uint16) string {
		switch version {
		case tls.VersionTLS10:
			return "TLSv10"
		case tls.VersionTLS11:
			return "TLSv11"
		case tls.VersionTLS12:
			return "TLSv12"
		case tls.VersionTLS13:
			return "TLSv13"
		default:
			return "Unknown"
		}
	}(state.Version))
	for i, cert := range state.PeerCertificates {
		text, _ := certinfo.CertificateText(cert)
		fmt.Printf("option Cert[%d]: %s\n", i, text)
	}
	fmt.Printf("option TLS.HandshakeComplete: %t\n", state.HandshakeComplete)
	fmt.Printf("option TLS.DidResume: %t\n", state.DidResume)
	for _, suit := range tls.CipherSuites() {
		if suit.ID == state.CipherSuite {
			fmt.Printf("option TLS.CipherSuite: %+v", suit)
			break
		}
	}
	fmt.Println()
}

func chunked(te []string) bool { return len(te) > 0 && te[0] == "chunked" }

func printRequestResponseForWindows(req *Request, res *http.Response) {
	var dumpHeader, dumpBody []byte
	dps := strings.Split(string(req.reqDump), "\n")
	for i, line := range dps {
		if len(strings.Trim(line, "\r\n ")) == 0 {
			dumpHeader = []byte(strings.Join(dps[:i], "\n"))
			dumpBody = []byte(strings.Join(dps[i:], "\n"))
			break
		}
	}

	if HasPrintOption(printReqHeader) {
		fmt.Println(string(dumpHeader))
		fmt.Println()
	}
	if HasPrintOption(printReqBody) {
		fmt.Println(string(dumpBody))
		fmt.Println()
	}

	if !req.DryRequest && HasPrintOption(printRspOption) {
		if res.TLS != nil {
			fmt.Printf("option TLS.DidResume: %t\n", res.TLS.DidResume)
			fmt.Println()
		}
	}

	if !req.DryRequest && HasPrintOption(printRspHeader) {
		fmt.Println(res.Proto, res.Status)
		for k, val := range res.Header {
			fmt.Println(k, ":", strings.Join(val, " "))
		}
		fmt.Println()
	}
	if !req.DryRequest && HasPrintOption(printRspBody) {
		fmt.Println(formatResponseBody(req, pretty, ugly, freeInnerJSON, false))
	}
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}
