package goup

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/bingoohuang/ngg/ss"
	"github.com/segmentio/ksuid"
)

// WriteJSON writes the JSON of v to http.ResponseWriter.
func WriteJSON(w http.ResponseWriter, v interface{}) error {
	js, err := json.Marshal(v)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)
	return err
}

// UploadResult is the structure of download result.
type UploadResult struct {
	Files         []string `json:"files,omitempty"`
	FileSizes     []string `json:"fileSizes,omitempty"`
	TotalSize     string   `json:"totalSize,omitempty"`
	Cost          string   `json:"cost,omitempty"`
	Start         string   `json:"start,omitempty"`
	End           string   `json:"end,omitempty"`
	MaxTempMemory string   `json:"maxTempMemory,omitempty"`
	LimitSize     string   `json:"limitSize,omitempty"`
}

// NetHTTPUpload upload
func NetHTTPUpload(w http.ResponseWriter, r *http.Request, rootDir string, limitSize uint64) (*UploadResult, error) {
	start := time.Now()
	maxMemory := 16 /*16 MiB */ << 20
	if err := r.ParseMultipartForm(int64(maxMemory)); err != nil {
		return nil, err
	}

	totalSize := int64(0)
	fileCount := len(r.MultipartForm.File)
	index := 0
	var files []string
	var fileSizes []string
	for k, v := range r.MultipartForm.File {
		index++
		file, n, err := saveFormFile(v[0], rootDir, r.URL.Path, index, fileCount)
		if err != nil {
			return nil, err
		}
		totalSize += n
		files = append(files, file)
		fileSizes = append(fileSizes, ss.Bytes(uint64(n)))
		log.Printf("recieved file %s: %s", k, file)
	}

	end := time.Now()
	return &UploadResult{
		Start:         start.UTC().Format(http.TimeFormat),
		End:           end.UTC().Format(http.TimeFormat),
		Files:         files,
		FileSizes:     fileSizes,
		MaxTempMemory: ss.Bytes(uint64(maxMemory)),
		LimitSize:     ss.Bytes(limitSize),
		TotalSize:     ss.Bytes(uint64(totalSize)),
		Cost:          end.Sub(start).String(),
	}, nil
}

// ParseFormFile returns the first file for the provided form key.
// FormFile calls ParseMultipartForm and ParseForm if necessary.
func ParseFormFile(m *multipart.Form) (*multipart.FileHeader, error) {
	if m != nil {
		if fhs := m.File["file"]; len(fhs) > 0 {
			return fhs[0], nil
		}

		for _, v := range m.File {
			return v[0], nil
		}
	}

	return nil, ErrMissingFile
}

// ErrMissingFile may be returned from FormFile when the is no uploaded file.
var ErrMissingFile = errors.New("there is no uploaded file")

// TrimExt trim ext from the right of filepath.
func TrimExt(filepath, ext string) string {
	return filepath[:len(filepath)-len(ext)]
}

func saveFormFile(fh *multipart.FileHeader, rootDir, urlPath string, fileIndex, fileCount int) (string, int64, error) {
	file, err := fh.Open()
	if err != nil {
		return "", 0, err
	}

	base := filepath.Base(urlPath)
	if base != "/" {
		if fileCount > 1 {
			ext := path.Ext(base)
			base = fmt.Sprintf("%s.%d%s", TrimExt(base, ext), fileIndex, ext)
		}
	}
	filename := firstFilename(base, filepath.Base(fh.Filename), ksuid.New().String())
	fullPath := filepath.Join(rootDir, filename)

	// use temporary file directly
	if f, ok := file.(*os.File); ok {
		n, err := file.Seek(0, io.SeekEnd)
		if err != nil {
			return "", n, err
		}
		if err := file.Close(); err != nil {
			return "", 0, err
		}
		if err := os.Rename(f.Name(), fullPath); err != nil {
			return "", 0, err
		}
		return fullPath, n, nil
	}

	n, err := writeChunk(fullPath, nil, file, nil)
	if err := file.Close(); err != nil {
		return "", 0, err
	}
	return fullPath, n, err
}

func firstFilename(s ...string) string {
	for _, i := range s {
		if i != "" && i != "/" {
			return i
		}
	}

	return ""
}
