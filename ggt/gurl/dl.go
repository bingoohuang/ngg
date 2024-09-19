package gurl

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/bingoohuang/ngg/ss"
)

var (
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Range
	// Content-Range: <unit> <range-start>-<range-end>/<size>
	// Content-Range: <unit> <range-start>-<range-end>/*
	contentRangeRegexp1 = regexp.MustCompile(`bytes ([0-9]+)-([0-9]+)/([0-9]+|\\*)`)
	// Content-Range: <unit> */<size>
	contentRangeRegexp2 = regexp.MustCompile(`bytes \\*/([0-9]+)`)

	// ErrWrongCodeForByteRange is returned if the client sends a request
	// with a Range header but the server returns a 2xx or 3xx code other
	// than 206 Partial Content.
	ErrWrongCodeForByteRange = errors.New("expected HTTP 206 from byte range request")
)

type contentRange struct {
	startByte   uint64
	endByte     uint64
	contentSize int64
}

func parseContentRange(contentRangeHead string) (*contentRange, error) {
	if contentRangeHead == "" {
		return nil, errors.New("no Content-Range header found in HTTP response")
	}

	subs := contentRangeRegexp1.FindStringSubmatch(contentRangeHead)
	if len(subs) < 4 {
		subs = contentRangeRegexp2.FindStringSubmatch(contentRangeHead)
		if len(subs) < 2 {
			return nil, fmt.Errorf("parse Content-Range: %s", contentRangeHead)
		}

		totalSize, err := strconv.ParseUint(subs[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse Content-Range: %s", contentRangeHead)
		}
		return &contentRange{
			contentSize: int64(totalSize),
		}, nil
	}

	startByte, err := strconv.ParseUint(subs[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse Content-Range: %s", contentRangeHead)
	}

	endByte, err := strconv.ParseUint(subs[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse Content-Range: %s", contentRangeHead)
	}

	contentSize := int64(0)

	if subs[3] == "*" {
		contentSize = -1
	} else {
		size, err := strconv.ParseUint(subs[3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse Content-Range: %s", contentRangeHead)
		}

		if endByte+1 != size {
			return nil, fmt.Errorf("range in Content-Range stops before the end of the content: %s", contentRangeHead)
		}

		contentSize = int64(size)
	}

	return &contentRange{
		startByte:   startByte,
		endByte:     endByte,
		contentSize: contentSize,
	}, nil
}

func downloadFile(req *Request, res *http.Response, filename string) {
	if ext := filepath.Ext(filename); ext == "" {
		contentType := res.Header.Get("Content-Type")
		if exts, _ := mime.ExtensionsByType(contentType); len(exts) > 0 {
			filename += exts[0]
		}
	}

	fd, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0o666)
	if err != nil {
		log.Fatalf("create download file %q failed: %v", filename, err)
	}

	var cr *contentRange
	contentRangeHead := res.Header.Get("Content-Range")
	if contentRangeHead != "" {
		cr, err = parseContentRange(contentRangeHead)
		if err != nil {
			log.Fatalf("parse Content-Range header failed: %v", err)
		}
	}

	if cr != nil && cr.startByte > 0 {
		if _, err := fd.Seek(int64(cr.startByte), io.SeekStart); err != nil {
			log.Fatalf("seek failed: %v", err)
		}
	}

	printRequestResponseForNonWindows(req, res, true)

	total, _ := strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)
	if total == 0 && !chunked(res.TransferEncoding) {
		return
	}

	if !HasPrintOption(quietFileUploadDownloadProgressing) {
		fmt.Printf("Downloading to %q\n", filename)
	}

	pb := NewProgressBar(total).Start()
	br := newProgressBarReader(res.Body, pb)

	if res.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(br)
		if err != nil {
			log.Fatalf("create download file %q failed: %v", filename, err)
		}
		br = reader
	}

	// disable timeout for downloading.
	if req.cancelTimeout != nil {
		req.cancelTimeout()
		req.cancelTimeout = nil
	}

	if conn := req.ConnInfo.Conn; conn != nil {
		// A zero value for t means I/O operations will not time out.
		if err := conn.SetDeadline(time.Time{}); err != nil {
			log.Printf("failed to set deadline: %v", err)
		}
	}

	if _, err := io.Copy(fd, br); err != nil {
		// A successful Copy returns err == nil, not err == EOF.
		log.Fatalf("download file %q failed: %v", filename, err)
	}
	if pb != nil {
		pb.Finish()
	}
	ss.Close[io.Closer](fd, br)
	fmt.Println()
}
