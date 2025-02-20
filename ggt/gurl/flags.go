package gurl

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/ss"
	"github.com/samber/lo"
	"github.com/spf13/pflag"
	"go.uber.org/atomic"
)

var (
	keepAlive, form, pretty                       bool
	ugly, raw, freeInnerJSON, gzipOn              bool
	countingItems, disableProxy, disableRedirect  bool
	auth, proxy, printV, body, think, method, dns string
	uploadFiles, urls                             []string
	printOption                                   uint32
	benchN, benchC, confirmNum                    int
	currentN                                      atomic.Int64
	timeout                                       time.Duration
	limitRate                                     = NewRateLimitFlag()
	download                                      = &ss.FlagStringBool{}

	jsonmap = map[string]any{}

	createDemoEnv bool
	unixSocket    string
)

func initFlags(p *pflag.FlagSet) {
	p.StringVarP(&unixSocket, "unix-socket", "s", "", "")
	flagEnvP(p, &urls, "url", "u", "", "", "URL")
	p.StringVarP(&method, "method", "m", "GET", "")

	p.BoolVar(&createDemoEnv, "demo.env", false, "")
	p.BoolVarP(&keepAlive, "keepalive", "k", true, "")
	p.StringVarP(&printV, "print", "p", "b", "")
	p.BoolVarP(&form, "form", "f", false, "")
	p.BoolVar(&gzipOn, "gzip", false, "")
	p.VarP(download, "download", "d", "")
	p.DurationVarP(&timeout, "timeout", "t", time.Minute, "")
	p.StringSliceVarP(&uploadFiles, "file", "F", nil, "")
	p.VarP(limitRate, "limit", "L", "")
	flagEnvVar(p, &think, "think", "0", "", "THINK")

	flagEnvVarP(p, &auth, "auth", "A", "", "", `AUTH`)
	flagEnvVarP(p, &proxy, "proxy", "P", "", "", `PROXY`)
	p.IntVarP(&benchN, "requests", "n", 1, "")
	p.IntVarP(&confirmNum, "confirm", "C", 0, "")
	p.IntVarP(&benchC, "concurrency", "c", 1, "")
	flagEnvVarP(p, &body, "body", "b", "", "", "BODY")
	flagEnvVar(p, &dns, "dns", "", "", "DNS")
}

const (
	printReqHeader uint32 = 1 << iota
	printReqURL
	printReqBody
	printRspOption
	printRspHeader
	printRspCode
	printRspBody
	printReqSession
	printVerbose
	printHTTPTrace
	printDebug
	printUgly
	printRaw
	printCountingItems
	quietFileUploadDownloadProgressing
	freeInnerJSONTag
	optionDisableProxy
	optionDisableRedirect
	optionDisableEval
)

func parsePrintOption(s string) {
	AdjustPrintOption(&s, 'A', printReqHeader|printReqBody|printRspHeader|printRspBody|printReqSession|printVerbose)
	AdjustPrintOption(&s, 'a', printReqHeader|printReqBody|printRspHeader|printRspBody|printReqSession|printVerbose)
	AdjustPrintOption(&s, 'H', printReqHeader)
	AdjustPrintOption(&s, 'B', printReqBody)
	AdjustPrintOption(&s, 'o', printRspOption)
	AdjustPrintOption(&s, 'h', printRspHeader)
	AdjustPrintOption(&s, 'b', printRspBody)
	AdjustPrintOption(&s, 's', printReqSession)
	AdjustPrintOption(&s, 'v', printVerbose)
	AdjustPrintOption(&s, 't', printHTTPTrace)
	AdjustPrintOption(&s, 'c', printRspCode)
	AdjustPrintOption(&s, 'u', printReqURL)
	AdjustPrintOption(&s, 'q', quietFileUploadDownloadProgressing)
	AdjustPrintOption(&s, 'f', freeInnerJSONTag)
	AdjustPrintOption(&s, 'd', printDebug)
	AdjustPrintOption(&s, 'U', printUgly)
	AdjustPrintOption(&s, 'r', printRaw)
	AdjustPrintOption(&s, 'C', printCountingItems)
	AdjustPrintOption(&s, 'N', optionDisableProxy)
	AdjustPrintOption(&s, 'E', optionDisableEval)
	AdjustPrintOption(&s, 'n', optionDisableRedirect)

	if s != "" {
		log.Fatalf("unknown print option: %s", s)
	}
}

func AdjustPrintOption(s *string, r rune, flags uint32) {
	if strings.ContainsRune(*s, r) {
		printOption |= flags
		*s = strings.ReplaceAll(*s, string(r), "")
	}
}

func HasAnyPrintOptions(flags ...uint32) bool {
	_, found := lo.Find(flags, HasPrintOption)
	return found
}

func HasPrintOption(flags uint32) bool {
	return printOption&flags == flags
}

const help = `gurl is a Go implemented cURL-like cli tool for humans.
Usage:
	gurl [flags] [METHOD] URL [URL] [ITEM [ITEM]]
flags:
  --unix-socket -s  Using unix socket file
  --url -u          HTTP request URL
  --method -m       HTTP method
  --keepalive -k    Enabled keepalive (default true)
  --version -v      Print Version Number
  -f                Submitting the data as a form
  --gzip            Gzip request body or not
  -d                Download the url content as file, yes/n
  -t                Timeout for read and write, default 1m
  --file -F file    Upload a file, e.g. gurl :2110 -F 1.png -F 2.png
  --limit -L limit  Limit rate /s, like 10K, append :req/:rsp to specific the limit direction
  --think            Think time, like 5s, 100ms, 100ms-5s, 100-200ms and etc.
  --auth USER[:PASS] HTTP authentication username:password, USER[:PASS]
  --proxy PROXY_URL Proxy host and port, PROXY_URL
  -n=1 -c=1         Number of requests and concurrency to run
  --confirm 0       Should confirm after number of requests \
  --body -b         Send RAW data as body 
				    @persons.tx to load body from the file's content
					@persons.txt:line 从文件按行读取请求体，发送多次请求
				    :rand.json to create rand JSON as body
  --print -p        String specifying what the output should contain, default will print only response body
                       H: request headers  B: request body,  u: request URL
                       h: response headers  b: response body, c: status code

                       s: http conn session v: Verbose t: HTTP trace
                       q: keep quiet for file uploading/downloading progress
                       f: expand inner JSON string as JSON object
                       d: print debugging info
                       U: print JSON In Ugly compact Format
                       r: print JSON Raw format other than pretty
                       C: print items counting in colored output
                       N: disable proxy
                       E: disable eval
                       n: disable redirects
                       o: print response option(like TLS)
                       a/A: HBhbsv
  --dns             Specified custom DNS resolver address, format: [DNS_SERVER]:[PORT]
  --version -v        Show Version Number
  --demo.env         Create a demo .env file
METHOD:
  gurl defaults to either GET (if there is no request data) or POST (with request data).
URL:
  The only one needed to perform a request is a URL. The default scheme is http://,
  which can be omitted from the argument; example.org works just fine.
ITEM:
  Can be any of: Query      : key=value  Header: key:value       Post data: key=value
                 Force query: key==value key==@/path/file
                 JSON data  : key:=value Upload: key@/path/file
                 File content as body: @/path/file
Example:
  gurl beego.me
  gurl :8080
  gurl -s $TMPDIR/a.sock http://unix/status -pa (服务端 httplive -p unix:$TMPDIR/a.sock, 代理端: tproxy -P unix://$TMPDIR/a.sock -p :3300)
Envs:
  1. URL:         URL
  2. PROXY:       Proxy host and port， like: http://proxy.cn, https://user:pass@proxy.cn
  3. AUTH:        HTTP authentication username:password, USER[:PASS]
  4. TLS_VERIFY:  Enable client verifies the server's certificate chain and host name.
  5. ROOT_CERTS:  Root CA.
     CERTS:       Client certs, e.g. client.crt,client.key, for TLCP: sign.cert,sign.key,enc.cert,enc.key
     TLS_SESSION_CACHE: 客户端会话缓存 LRUClientSessionCache 大小，默认值 32,  
  6. LOCAL_IP:    Specify the local IP address to connect to server.
  7. TLCP:        使用传输层密码协议(TLCP)，TLCP协议遵循《GB/T 38636-2020 信息安全技术 传输层密码协议》。
     TLCP_SESSION_CACHE: TLCP 会话缓存大小，默认值 32
  8. CHUNKED:     开启请求中的块传输
  9. INTERACTIVE=0  禁止交互模式，否则 请求参数值/地址中的注入 @age 将被解析成插值模式，会要求从命令行输入
more help information please refer to https://github.com/bingoohuang/ngg/ggt/gurl
`

//go:embed demo.env
var demoEnv []byte

func createDemoEnvFile() error {
	if !createDemoEnv {
		return nil
	}

	if ss.Pick1(ss.Exists(".env")) {
		return fmt.Errorf(".env file already exists")
	}

	if err := os.WriteFile(".env", demoEnv, 0o644); err != nil {
		return fmt.Errorf("write .env file: %w", err)
	}

	log.Printf("a demo .env file has been created!")

	return io.EOF
}

type RateLimitDirection int

const (
	RateLimitBoth RateLimitDirection = iota
	RateLimitRequest
	RateLimitResponse
)

func NewRateLimitFlag() *RateLimitFlag {
	return &RateLimitFlag{}
}

type RateLimitFlag struct {
	Val       *uint64
	Direction RateLimitDirection
}

func (i *RateLimitFlag) Enabled() bool { return i.Val != nil && *i.Val > 0 }

func (i *RateLimitFlag) String() string {
	if !i.Enabled() {
		return "0"
	}

	s := ss.Bytes(*i.Val)
	switch i.Direction {
	case RateLimitRequest:
		return s + ":req"
	case RateLimitResponse:
		return s + ":rsp"
	}

	return s
}

func (i *RateLimitFlag) Type() string {
	return "ratelimit"
}

func (i *RateLimitFlag) Set(value string) (err error) {
	dirPos := strings.IndexByte(value, ':')
	i.Direction = RateLimitBoth
	if dirPos > 0 {
		switch dir := value[dirPos+1:]; strings.ToLower(dir) {
		case "req":
			i.Direction = RateLimitRequest
		case "rsp":
			i.Direction = RateLimitResponse
		default:
			log.Fatalf("unknown rate limit %s", value)
		}
		value = value[:dirPos]
	}

	val, err := ss.ParseBytes(value)
	if err != nil {
		return err
	}

	i.Val = &val
	return nil
}

func (i *RateLimitFlag) IsForReq() bool {
	return i.Enabled() && (i.Direction == RateLimitRequest || i.Direction == RateLimitBoth)
}

func (i *RateLimitFlag) IsForRsp() bool {
	return i.Enabled() && (i.Direction == RateLimitResponse || i.Direction == RateLimitBoth)
}

func (i *RateLimitFlag) Float64() float64 {
	if i.Val == nil {
		return 0
	}
	return float64(*i.Val)
}
