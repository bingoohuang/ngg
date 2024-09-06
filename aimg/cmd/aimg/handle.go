package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bingoohuang/ngg/httpretty"
	"github.com/bingoohuang/ngg/rotatefile"
	"github.com/bingoohuang/ngg/rotatefile/stdlog"
	"github.com/bingoohuang/ngg/ss"
)

var pretty = func() *httpretty.Logger {
	p := &httpretty.Logger{
		SkipSanitize:   true,
		Time:           true,
		TLS:            true,
		RequestHeader:  true,
		RequestBody:    true,
		ResponseHeader: true,
		ResponseBody:   false,
		Colors:         rotatefile.IsTerminal,
		Formatters:     []httpretty.Formatter{&httpretty.JSONFormatter{}},
	}
	p.SetOutput(stdlog.LevelLog)
	return p
}()

func handle(baseURL, pageID, dbName string, httprettyLog bool, w http.ResponseWriter, r *http.Request) error {
	if !strings.HasPrefix(r.URL.Path, "/") {
		r.URL.Path = "/" + r.URL.Path
	}

	if strings.HasPrefix(r.URL.Path, "/static/") {
		http.FileServer(http.FS(staticFS)).ServeHTTP(w, r)
		return nil
	}

	h := &handler{dbName: dbName, baseURL: baseURL, pageID: pageID}
	var handler http.Handler = h
	if httprettyLog {
		handler = pretty.Middleware(handler, true)
	}

	handler.ServeHTTP(w, r)

	return h.error
}

type handler struct {
	error   error
	dbName  string
	baseURL string
	pageID  string
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.error = h.serveHTTP(w, r)
}

func (h *handler) serveHTTP(w http.ResponseWriter, r *http.Request) error {
	db, err := UsingDB(h.dbName)
	if err != nil {
		return err
	}

	urlPath := r.URL.Path
	if r.Method == http.MethodPost {
		if prefix, sub := TrimPathAnyPrefix(urlPath, "/delete/x/"); prefix != "" {
			if db1 := db.DB.Delete(&Img{}, "xxhash=?", sub); db1.Error != nil {
				return db1.Error
			}
			return nil
		}
		if prefix, sub := TrimPathAnyPrefix(urlPath, "/favorite/x/"); prefix != "" {
			if xh, rating := ss.Split2(sub, "/"); rating != "" {
				if db1 := db.DB.Model(&Img{}).
					Where("xxhash=?", xh).
					Update("favorite", ss.Pick1(ss.Parse[int](rating))); db1.Error != nil {
					return db1.Error
				}
				return nil
			}
		}
		return fmt.Errorf("bad request")
	}

	limitN, limit := parseLimit(r)

	if prefix, sub := TrimPathAnyPrefix(urlPath, "/size/", "/s/"); prefix != "" {
		size1, size2, err := parseSizeExpr(sub)
		if err != nil {
			return err
		}
		return QuerySize(h.baseURL, w, r, db.DB, size1, size2, limit, limitN)
	}
	if prefix, path := TrimPathAnyPrefix(urlPath, "/page/", "/p/", "/P/"); prefix != "" {
		return QueryPage(h.baseURL, w, r, db.DB, path, limit, limitN, prefix == "/P/")
	}
	if prefix, path := TrimPathAnyPrefix(urlPath, "/xxhash/", "/x/"); prefix != "" {
		return QueryXxHash(h.baseURL, w, r, db.DB, path, limitN)
	}

	if urlPath == "/today" {
		pageID := getTodayPageID(db.DB, h.pageID)
		return QueryPage(h.baseURL, w, r, db.DB, pageID, limit, limitN, true)
	}
	if urlPath != "/" {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	return QueryRandImage(h.baseURL, w, r, db, limit, limitN)
}

func AnyPrefix(src string, prefixes ...string) (prefix string) {
	for _, p := range prefixes {
		if strings.HasPrefix(src, p) {
			return p
		}
	}
	return ""
}

func TrimPathAnyPrefix(src string, prefixes ...string) (prefix, trimmed string) {
	for _, p := range prefixes {
		if strings.HasPrefix(src, p) {
			return p, strings.TrimSuffix(strings.TrimPrefix(src, p), "/")
		}
	}

	return "", src
}

func parseSizeExpr(expr string) (size1, size2 uint64, err error) {
	sep := strings.IndexByte(expr, '/')
	if sep < 0 {
		size1, err = ss.ParseBytes(expr)
		return
	}

	size1, err = ss.ParseBytes(expr[:sep])
	if err != nil {
		return 0, 0, fmt.Errorf("parse bytes %s: %w", expr[:sep], err)
	}

	size2, err = ss.ParseBytes(expr[sep+1:])
	if err != nil {
		return 0, 0, fmt.Errorf("parse bytes %s: %w", expr[sep+1:], err)
	}

	return size1, size2, nil
}
