package internal

import (
	"io"
	"log"
	"os"
	"strings"

	"github.com/bingoohuang/ngg/ss"
	"github.com/valyala/fasthttp"
	"go.uber.org/multierr"
)

func If[T any](c bool, a, b T) T {
	if c {
		return a
	}

	return b
}

func Quoted(s, open, close string) (string, bool) {
	p1 := strings.Index(s, open)
	if p1 != 0 {
		return "", false
	}

	s = s[len(open):]
	if !strings.HasSuffix(s, close) {
		return "", false
	}

	return strings.TrimSuffix(s, close), true
}

type Closers []io.Closer

func (closers Closers) Close() (err error) {
	for _, c := range closers {
		err = multierr.Append(err, c.Close())
	}

	return
}

// SetHeader set request header if value is not empty.
func SetHeader(r *fasthttp.Request, header, value string) {
	if value != "" {
		r.Header.Set(header, value)
	}
}

func ParseBodyArg(body string, stream, lineMode bool) (streamFileName string, bodyBytes []byte, lines chan string) {
	filename := body
	if strings.HasPrefix(body, "@") {
		filename = (body)[1:]
		if ok, _ := ss.Exists(filename); !ok {
			return "", []byte(body), nil
		}
	}

	fileExists, _ := ss.Exists(filename)
	if lineMode && fileExists {
		var err error
		lines, err = ss.LinesChan(filename, 1000)
		if err != nil {
			log.Fatalf("E! create line chan for %s, failed: %v", filename, err)
		}
		return "", nil, lines
	}

	if fileExists {
		if f, err := ss.ReadFile(filename); err == nil {
			body = string(f)
		}
	}

	streamFileName, bodyBytes = ParseFileArg(body)
	return ss.If(stream, streamFileName, ""), bodyBytes, nil
}

// ParseFileArg parse an argument which represents a string content,
// or @file to represents the file's content.
func ParseFileArg(arg string) (file string, data []byte) {
	if strings.HasPrefix(arg, "@") {
		f := (arg)[1:]
		if v, err := os.ReadFile(f); err != nil {
			log.Fatalf("failed to read file %s, error: %v", f, err)
			return f, nil
		} else {
			return f, v
		}
	}

	return "", []byte(arg)
}
