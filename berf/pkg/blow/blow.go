package blow

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/berf"
	"github.com/bingoohuang/ngg/berf/pkg/blow/internal"
	"github.com/bingoohuang/ngg/berf/pkg/util"
	"github.com/bingoohuang/ngg/gnet"
	"github.com/bingoohuang/ngg/ss"
	"github.com/spf13/pflag"
	"github.com/valyala/fasthttp"
)

var (
	pUrls = pflag.StringSlice("url", nil, "URL")
	pBody = pflag.StringP("body", "b", "",
		"HTTP request body, or @file to read from, "+
			"or @file:stream for chunked encoding of file, "+
			"or @file:line to read line by line")
	pUpload = pflag.StringP("upload", "u", "",
		"HTTP upload multipart form file or directory or glob pattern like ./*.jpg, \n"+
			"      prefix file: to set form field name\n"+
			"      extension: rand.png,rand.art,rand.jpg,rand.json\n"+
			"      env export UPLOAD_INDEX=%clear%y%M%d.%H%m%s.%i%ext to append index to the file base name, \n"+
			"                                %clear: 清除原始文件名\n"+
			"                                %y: 4位年 %M: 2位月 %d: 2位日 %H: 2位时 %m: 2位分 %s: 2位秒\n"+
			"                                %i: 自增长序号, %05i： 补齐5位的自增长序号（前缀补0)\n"+
			"      env export UPLOAD_EXIT=1 to exit when all files are uploaded\n"+
			"      env export UPLOAD_SHUFFLE=1 to shuffle the upload files (only for glob pattern)")
	pMethod  = pflag.StringP("method", "m", "", "HTTP method")
	pNetwork = pflag.String("network", "",
		"Network simulation, local: simulates local network, lan: local, wan: wide, bad: bad network, "+
			"or BPS:latency like 20M:20ms")
	pHeaders = pflag.StringSlice("header,H", nil,
		"Custom HTTP headers, K:V, e.g. Content-Type")
	pProfiles = pflag.StringSliceP("profile", "P", nil,
		"Profile file, append :new to create a demo profile, or :tag to run only specified profile, "+
			"or range :tag1,tag3-tag5")
	pEnv  = pflag.String("env", "", "Profile env name selected")
	pOpts = pflag.StringSlice("opt", nil, "options, multiple by comma: \n"+
		"      gzip:               enabled content gzip  \n"+
		"      tlsVerify:          verify the server's cert chain and host name \n"+
		"      no-keepalive/no-ka: disable keepalive \n"+
		"      form:               use form instead of json \n"+
		"      pretty:             pretty JSON \n"+
		"      uploadIndex:        insert upload index to filename \n"+
		"      saveRandDir:        save rand generated request files to dir \n"+
		"      json:               set Content-Type=application/json; charset=utf-8 \n"+
		"      eval:               evaluate url and body's variables \n"+
		"      log:                log sampling rate, e.g. 0.5 \n"+
		"      notty:              no tty color \n")
	pAuth = pflag.String("auth", "",
		"basic auth, eg. scott:tiger or direct base64 encoded like c2NvdHQ6dGlnZXI")
	pDir     = pflag.String("dir", "", "download dir, use :temp for temp dir")
	pCertKey = pflag.String("cert", "",
		"Path to the client's TLS Cert and private key file, eg. ca.pem,ca.key")
	pRootCert = pflag.String("root-ca", "",
		"Ca root certificate file to verify TLS")
	pTimeout = pflag.String("timeout", "",
		"Timeout for each http request, e.g. 5s for do:5s,dial:5s,write:5s,read:5s")
	pPrint = pflag.StringP("print", "p", "",
		"a: all, R: req all, H: req headers, B: req body, r: resp all, h: resp headers b: resp body c: status code")
	pStatusName = pflag.String("status", "", "Status name in json, like resultCode")

	pCreateEnvFile = pflag.Bool("demo.env", false, fmt.Sprintf("create a demo .env in current dir.\n"+
		"       env HOSTS          以,分隔指定HTTP多个请求 HOST 头, e.g. HOSTS=a.cn,b.cn berf ...\n"+
		"       env LOCAL_IP       指定网卡IP,多个IP以,分割, e.g. LOCAL_IP=192.168.1.2,192.168.1.3 berf ...\n"+
		"       env DEBUG          激活全局DEBUG模式，打印更多日志, e.g. DEBUG=1 berf ...\n"+
		"       env TLCP           使用传输层密码协议(TLCP)，遵循《GB/T 38636-2020 信息安全技术 传输层密码协议》，默认值0, e.g. TLCP=1 berf ...\n"+
		"       env TLCP_CERTS     TLCP客户端证书(ECC系列单证书/ECDHE系列套件双证书)，即认证密钥,认证证书[,加密密钥,加密证书], 示例: sign.cert.pem,sign.key.pem[,enc.cert.pem,enc.key.pem] \n"+
		"       env TLS_SESSION_CACHE         客户端会话缓存，默认值 32, e.g. TLS_SESSION_CACHE=32 berf ...\n"+
		"       env MAX_GREEDY_CONNS_PER_HOST 在从连接池获取连接时，总是优先创建新链接，直到 MAX_GREEDY_CONNS_PER_HOST 为止，默认值0, e.g. MAX_GREEDY_CONNS_PER_HOST=100 berf ...\n"+
		"       env MAX_CONNS_PER_HOST        单台主机最大连接数，默认%d, e.g. MAX_CONNS_PER_HOST=1000 berf ...\n"+
		"       env MAX_IDLE_CONN_DURATION    连接最大空闲时间，默认%s, e.g. MAX_IDLE_CONN_DURATION=10s berf ...\n",
		fasthttp.DefaultMaxConnsPerHost, fasthttp.DefaultMaxIdleConnDuration))
)

const (
	printReqHeader uint8 = 1 << iota
	printReqBody
	printRespOption
	printRespHeader
	printRespBody
	printRespStatusCode
	printDebug
	printAll
)

func parsePrintOption(s string) (printOption uint8) {
	for r, v := range map[string]uint8{
		"A": printAll,
		"a": printAll,
		"R": printReqHeader | printReqBody,
		"r": printRespHeader | printRespBody,
		"H": printReqHeader,
		"B": printReqBody,
		"o": printRespOption,
		"h": printRespHeader,
		"b": printRespBody,
		"c": printRespStatusCode,
		"d": printDebug,
	} {
		if strings.Contains(s, r) {
			printOption |= v
			s = strings.ReplaceAll(s, r, "")
		}
	}

	if s = strings.TrimSpace(s); s != "" {
		log.Printf("unknown print options, %s", s)
		os.Exit(1)
	}

	if printOption&printRespHeader == printRespHeader {
		printOption &^= printRespStatusCode
	}

	return printOption
}

type Bench struct {
	invoker *Invoker
}

func (b *Bench) Name(context.Context, *berf.Config) string {
	opt := b.invoker.opt
	if v := opt.urls; len(v) > 0 {
		return strings.Join(v, ",")
	}

	return "profiles " + strings.Join(*pProfiles, ",")
}

func (b *Bench) Final(_ context.Context, conf *berf.Config) error {
	b.invoker.waitLog()
	return nil
}

//go:embed .env
var envFileDemo []byte

func (b *Bench) Init(ctx context.Context, conf *berf.Config) (*berf.BenchOption, error) {
	if *pCreateEnvFile {
		return nil, b.createEnvFileDemo()
	}

	b.invoker = Blow(ctx, conf)
	b.invoker.Run(ctx, conf, true)
	return &berf.BenchOption{
		NoReport: b.invoker.opt.printOption > 0,
	}, nil
}

func (b *Bench) createEnvFileDemo() error {
	if ok, _ := ss.Exists(".env"); ok {
		return fmt.Errorf(".env file already exists, please remove or rename it first")
	}

	if err := os.WriteFile(".env", envFileDemo, 0o644); err != nil {
		return fmt.Errorf("create .env file: %w", err)
	}

	log.Printf(".env file created")

	return io.EOF
}

func (b *Bench) Invoke(ctx context.Context, conf *berf.Config) (*berf.Result, error) {
	return b.invoker.Run(ctx, conf, false)
}

type Opt struct {
	berfConfig    *berf.Config
	bodyLinesChan chan string
	urls          []string
	parsedUrls    []*url.URL
	upload        string

	rootCert string
	certPath string
	keyPath  string

	downloadDir string

	method  string
	network string

	auth        string
	saveRandDir string
	statusName  string

	bodyStreamFile string

	bodyBytes []byte

	profiles []*internal.Profile

	headers []string

	doTimeout    time.Duration
	verbose      int
	readTimeout  time.Duration
	writeTimeout time.Duration
	dialTimeout  time.Duration

	maxConns    int
	jsonBody    bool
	eval        bool
	uploadIndex bool
	enableGzip  bool
	ant         bool

	printOption uint8
	form        bool
	noKeepalive bool

	tlsVerify          bool
	pretty             bool
	noTLSessionTickets bool
	logSamplingRate    float64
}

func (o *Opt) HasPrintOption(feature uint8) bool {
	return o.printOption&feature == feature || o.printOption&printAll == printAll
}

func (o *Opt) MaybePost() bool {
	return o.upload != "" || len(o.bodyBytes) > 0 || o.bodyStreamFile != "" || o.bodyLinesChan != nil
}

func TryStartAsBlow() bool {
	if !IsBlowEnv() {
		return false
	}

	StartBlow()
	return true
}

func StartBlow() {
	berf.StartBench(context.Background(),
		&Bench{},
		berf.WithOkStatus(ss.Or(*pStatusName, "200")),
		berf.WithCounting("连接数"))
}

func IsBlowEnv() bool {
	if len(*pUrls) > 0 {
		return true
	}

	if isBlow := len(*pProfiles) > 0; isBlow {
		return true
	}

	return parseUrlFromArgs() != ""
}

func parseUrlFromArgs() string {
	if args := excludeHttpieLikeArgs(pflag.Args()); len(args) > 0 {
		urlAddr, err := gnet.FixURI{}.Fix(args[0])
		if err == nil {
			return urlAddr.String()
		}
	}

	return ""
}

func Blow(ctx context.Context, conf *berf.Config) *Invoker {
	var urlAddrs []string
	urlAddrs = append(urlAddrs, *pUrls...)
	if len(urlAddrs) == 0 {
		if urlAddr := parseUrlFromArgs(); urlAddr != "" {
			urlAddrs = append(urlAddrs, urlAddr)
		}
	}

	stream := util.SplitTail(pBody, ":stream")
	lineMode := util.SplitTail(pBody, ":line")

	bodyStreamFile, bodyBytes, linesChan := internal.ParseBodyArg(*pBody, stream, lineMode)
	cert, key := ss.Split2(*pCertKey, ",")

	opts := util.NewFeatures(*pOpts...)

	timeout, err := parseDurations(*pTimeout)
	if err != nil {
		log.Fatal(err.Error())
	}

	opt := &Opt{
		urls:           urlAddrs,
		method:         *pMethod,
		headers:        *pHeaders,
		bodyLinesChan:  linesChan,
		bodyBytes:      bodyBytes,
		bodyStreamFile: bodyStreamFile,
		upload:         *pUpload,

		rootCert:    *pRootCert,
		certPath:    cert,
		keyPath:     key,
		tlsVerify:   opts.HasAny("tlsVerify"),
		downloadDir: *pDir,

		doTimeout:    timeout.Get("do"),
		readTimeout:  timeout.Get("read", "r"),
		writeTimeout: timeout.Get("write", "w"),
		dialTimeout:  timeout.Get("dial", "d"),

		network:  *pNetwork,
		auth:     *pAuth,
		maxConns: conf.Goroutines,

		enableGzip:         opts.HasAny("g", "gzip"),
		uploadIndex:        opts.HasAny("uploadIndex", "ui"),
		noKeepalive:        opts.HasAny("no-keepalive", "no-ka"),
		noTLSessionTickets: opts.HasAny("no-tls-session-tickets", "no-tst"),
		form:               opts.HasAny("form"),
		pretty:             opts.HasAny("pretty"),
		eval:               opts.HasAny("eval"),
		jsonBody:           opts.HasAny("json"),
		logSamplingRate:    opts.GetFloat("log", -1),
		ant:                opts.HasAny("ant"),
		saveRandDir:        opts.Get("saveRandDir"),
		verbose:            conf.Verbose,
		statusName:         *pStatusName,
		printOption:        parsePrintOption(*pPrint),
		berfConfig:         conf,
	}

	if opt.downloadDir == ":temp" {
		opt.downloadDir = os.TempDir()
	}

	if opts.HasAny("notty") {
		hasStdoutDevice = false
	}

	opt.profiles = internal.ParseProfileArg(*pProfiles, *pEnv)
	invoker, err := NewInvoker(ctx, opt)
	ss.ExitIfErr(err)
	return invoker
}

type Durations struct {
	Map     map[string]time.Duration
	Default time.Duration
}

func (d *Durations) Get(keys ...string) time.Duration {
	for _, key := range keys {
		if v, ok := d.Map[strings.ToLower(key)]; ok {
			return v
		}
	}
	return d.Default
}

// parseDurations parses expression like do:5s,dial:5s,write:5s,read:5s to Durations struct.
func parseDurations(s string) (*Durations, error) {
	d := &Durations{Map: make(map[string]time.Duration)}
	var err error
	for _, one := range ss.Split(s, ",") {
		if p := strings.IndexRune(one, ':'); p > 0 {
			k, v := strings.TrimSpace(one[:p]), strings.TrimSpace(one[p+1:])
			d.Map[strings.ToLower(k)], err = time.ParseDuration(v)
			if err != nil {
				return nil, fmt.Errorf("failed to parse expressions %s, err: %w", s, err)
			}
		} else {
			if d.Default, err = time.ParseDuration(one); err != nil {
				return nil, fmt.Errorf("failed to parse expressions %s, err: %w", s, err)
			}
		}
	}

	return d, nil
}
