package gurl

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/ver"
)

var defaultSetting = Settings{
	DumpRequest:    true,
	UserAgent:      "curl/" + ver.AppVersion,
	ConnectTimeout: 60 * time.Second,
	DumpBody:       true,
}

var keyReg = regexp.MustCompile(`^([\w_.\-]*)(==|:=|=|:|@)(.*)`)

func getHTTP(method, url string, args []string, timeout time.Duration) (r *Request) {
	if confirmNum > 0 {
		timeout = 0
	}
	r = NewRequest(url, method)
	r.DisableKeepAlives = !keepAlive
	r.Setting = defaultSetting
	r.Setting.ConnectTimeout = timeout
	r.DryRequest = strings.HasPrefix(url, DryRequestURL)
	r.Timeout = timeout
	r.Header("Accept-Encoding", "gzip, deflate")
	if form || method == "GET" {
		r.Header("Accept", "*/*")
	} else {
		r.Header("Accept", "application/json")
	}
	r.Header("Gurl-Date", time.Now().UTC().Format(http.TimeFormat))
	// https://httpie.io/docs#request-items
	// Item Type	Description
	// HTTP Headers Name:Value	Arbitrary HTTP header, e.g. X-API-Token:123
	// URL parameters name==value	Appends the given name/value pair as a querystring parameter to the URL. The == separator is used.
	// Data Fields field=value, field=@file.txt	Request data fields to be serialized as a JSON object (default), to be form-encoded (with --form, -f), or to be serialized as multipart/form-data (with --multipart)
	// Raw JSON field:=json	Useful when sending JSON and one or more fields need to be a Boolean, Number, nested Object, or an Array, e.g., meals:='["ham","spam"]' or pies:=[1,2,3] (note the quotes)
	// File upload fields field@/dir/file, field@file;type=mime	Only available with --form, -f and --multipart. For example screenshot@~/Pictures/img.png, or 'cv@cv.txt;type=text/markdown'. With --form, the presence of a file field results in a --multipart request
	for i := range args {
		arg := args[i]
		subs := keyReg.FindStringSubmatch(arg)
		if len(subs) == 0 {
			continue
		}

		k, op, val := subs[1], subs[2], subs[3]
		if k == "" && op != "@" {
			log.Fatalf("Unsupported argument %s", arg)
		}

		switch op {
		case ":=": // Json raws
			if dat, fn, err := readFile(val); err != nil {
				log.Fatal("Read File", fn, err)
			} else if fn != "" {
				var j any
				if err := json.Unmarshal(dat, &j); err != nil {
					log.Fatal("Read from File", fn, "Unmarshal", err)
				}
				jsonmap[k] = j
			} else {
				jsonmap[k] = json.RawMessage(dat)
			}
		case "==": // Queries
			r.Query(k, tryReadFile(val))
		case "=": // Params
			if val = tryReadFile(val); form || method == "GET" {
				r.Param(k, val) // As Query parameter,
			} else {
				jsonmap[k] = val // body will be eval later
			}
		case ":": // Headers
			if k == "Host" {
				r.SetHost(val)
			} else {
				if strings.EqualFold(k, "Accept") && strings.EqualFold(val, "JSON") {
					r.Header("Accept", "application/json")
				} else {
					r.Header(k, val)
				}
			}
		case "@": // files
			if k != "" {
				r.PostFile(k, val)
			} else {
				r.Body(arg)
			}
		}
	}
	if !form && len(jsonmap) > 0 {
		r.Body(jsonmap)
	}
	return
}

func tryReadFile(s string) string {
	dat, _, err := readFile(s)
	if err != nil {
		log.Fatal("Read File", s, err)
	}

	return string(dat)
}

func readFile(s string) (data []byte, fn string, e error) {
	if !strings.HasPrefix(s, "@") {
		return []byte(s), "", nil
	}

	filename := s[1:]
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		eval, err := Eval(s)
		if err != nil {
			return []byte(s), "", nil
		}
		return []byte(eval), "", nil
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, filename, err
	}
	defer ss.Close(f)

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, filename, err
	}
	return content, filename, nil
}

const (
	MaxPayloadSize        = "MAX_PAYLOAD_SIZE"
	DefaultMaxPayloadSize = 1024 * 1024
)

func formatResponseBody(r *Request, pretty, ugly, freeInnerJSON, influxDB bool) string {
	dat, err := r.Bytes()
	if err != nil {
		log.Fatalln("can't get the url", err)
	}

	if influxTablePrint(ugly, influxDB, dat) {
		return ""
	}

	if saveTempFile(dat, MaxPayloadSize, ugly) {
		return ""
	}

	return formatBytes(dat, pretty, ugly, freeInnerJSON)
}

func saveTempFile(dat []byte, envName string, ugly bool) bool {
	if ugly {
		return false
	}

	if m, _ := ss.GetenvBytes(envName, DefaultMaxPayloadSize); m > 0 && len(dat) > int(m) {
		if name, err := ss.WriteTempFile("", "", dat, false); err == nil {
			log.Printf("body is too large, %d / %s > %d / %s (set $%s), write to file: %s",
				len(dat), ss.Bytes(uint64(len(dat))), m, ss.Bytes(uint64(m)), envName, name)
			return true
		}
	}

	return false
}

func formatBytes(body []byte, pretty, ugly, freeInnerJSON bool) string {
	body = bytes.TrimSpace(body)
	isJSON := jj.ParseBytes(body).IsJSON()

	if isJSON {
		if freeInnerJSON {
			body = jj.FreeInnerJSON(body)
		}

		if ugly {
			body = jj.Ugly(body)
		} else if pretty {
			body = jj.Pretty(body, jj.DefaultOptions)
		}
	}

	if hasStdoutDevice {
		return ColorfulResponse(string(body), isJSON)
	}

	return string(body)
}
