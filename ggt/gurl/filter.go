package gurl

import (
	"net"
	"os"
	"strings"

	"github.com/bingoohuang/ngg/gnet"
	"github.com/bingoohuang/ngg/ss"
)

var (
	methodList            = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	methodSpecifiedInArgs bool
)

var caFile = os.Getenv("CERT")

func filter(args []string) []string {
	var filteredArgs []string
	defaultSchema := ss.If(caFile != "", "https", "http")
	fixURI := gnet.FixURI{DefaultScheme: defaultSchema}

	for _, arg := range args {
		if ss.HasPrefix(arg, "http://", "https://") {
			urls = append(urls, arg)
			continue
		}

		if inSlice(strings.ToUpper(arg), methodList) {
			method = strings.ToUpper(arg)
			methodSpecifiedInArgs = true
			continue
		}

		if subs := keyReg.FindStringSubmatch(arg); len(subs) > 0 && subs[1] != "" {
			k := subs[1]
			if ip := net.ParseIP(k); ip != nil { // 127.0.0.1:5003
				if u, err := fixURI.Fix(arg); err == nil && strings.ContainsAny(arg, ":/") {
					urls = append(urls, u.String())
					continue
				}
			} else if strings.Contains(subs[1], ".") && subs[2] == ":" { // a.b.c:5003
				if u, err := fixURI.Fix(arg); err == nil && strings.ContainsAny(arg, ":/") {
					urls = append(urls, u.String())
					continue
				}
			}

			filteredArgs = append(filteredArgs, arg)
			continue
		}

		if u, err := fixURI.Fix(arg); err == nil && strings.ContainsAny(arg, ":/") {
			urls = append(urls, u.String())
		} else if subs := keyReg.FindStringSubmatch(arg); len(subs) == 0 {
			urls = append(urls, arg)
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	args = filteredArgs

	if isMethodDefaultGet() {
		if len(uploadFiles) > 0 {
			method = "POST"
		} else {
			for _, v := range args {
				subs := keyReg.FindStringSubmatch(v)
				if len(subs) == 0 {
					continue
				}

				// defaults to either GET (with no request data) or POST (with request data).
				switch _, op, _ := subs[1], subs[2], subs[3]; op {
				case ":=": // Json raws
					method = "POST"
				case "==": // Queries
				case "=": // Params
					method = "POST"
				case ":": // Headers
				case "@": // files
					method = "POST"
				}
				if method == "POST" {
					break
				}
			}
		}
	}

	if isMethodDefaultGet() && body != "" {
		method = "POST"
	}

	return args
}

func isMethodDefaultGet() bool {
	return !methodSpecifiedInArgs && method == "GET"
}
