package goup

import (
	"compress/gzip"
	_ "embed" // embed
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bingoohuang/ngg/ggt/goup/codec"
	"github.com/bingoohuang/ngg/ggt/goup/shapeio"
	"github.com/bingoohuang/ngg/ss"
	"github.com/minio/sio"
	"github.com/schollz/pake/v3"
)

// for Drag and Drop File Uploading, https://css-tricks.com/drag-and-drop-file-uploading/
//
//go:embed index.html
var indexPage []byte

// InitServer initializes the server.
func InitServer() error {
	return ensureDir(RootDir)
}

type limitResponseWriter struct {
	http.ResponseWriter
	*shapeio.RateLimiter
}

// Write writes bytes from p.
func (s *limitResponseWriter) Write(p []byte) (int, error) {
	n, err := s.ResponseWriter.Write(p)
	if err != nil || s.Limiter == nil {
		return n, err
	}

	err = s.WaitN(s.Context, n)
	return n, err
}

// ServerHandle is main request/response handler for HTTP server.
func ServerHandle(code, cipher string, chunkSize, limitRate uint64, paths []string) http.HandlerFunc {
	f := func(w http.ResponseWriter, r *http.Request) error {
		h := ParseHeader(r.Header.Get("Content-Gulp"))
		if chunkSize > 0 {
			r.Body = http.MaxBytesReader(w, r.Body, int64(chunkSize*2)) // with extra 1 MiB, for padding compatible like encryption
		}
		if limitRate > 0 {
			l := shapeio.WithRateLimit(float64(limitRate))
			r.Body = shapeio.NewReader(r.Body, l)
			w = &limitResponseWriter{ResponseWriter: w, RateLimiter: shapeio.NewRateLimiter(l)}
		}
		defer func() {
			io.Copy(io.Discard, r.Body)
			ss.Close(r.Body)
		}()

		switch {
		case h.Filename != "" && r.Method == http.MethodPost:
			log.Printf("明文上传（文件作为 Body) %s", h.Filename)
			return serveBodyAsFile(r.Body, h.Filename)
		case h.Session != "" && h.Curve != "" && r.Method == http.MethodPost:
			log.Printf("PAKE 生成会话秘钥 %s", h.Session)
			return servePake(w, h.Session, code, h.Curve)
		case h.Session != "" && r.URL.Path == "/" && h.Range != "" && ss.AnyOf(r.Method, http.MethodPost, http.MethodGet):
			// 校验分块 checksum，返回 304 或 其它
			log.Printf("分块加密上传（加密分块作为 Body) %s", h.Range)
			return serveUpload(w, r, h.Range, h.Session, cipher, h.Checksum, h.Salt)
		case r.URL.Path == "/" && r.Method == http.MethodGet:
			// HTML JS 上传页面 / 服务端文件列表（Accept: application/json 时）
			if r.Header.Get("Accept") == "application/json" {
				return servList(w)
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, err := w.Write(indexPage)
			return err
		case r.URL.Path != "/" && (r.Method == http.MethodGet || r.Method == http.MethodHead): // may be downloads
			log.Printf("明文下载 %s", r.URL.Path)
			if status := serveDownload(w, r, h.Session, cipher, h.Range, h.Checksum, chunkSize, paths); status > 0 {
				w.WriteHeader(status)
			}
		case r.Method == http.MethodPost:
			log.Printf("明文上传（multipart-form)")
			r, err := NetHTTPUpload(w, r, RootDir, chunkSize)
			if err != nil {
				return err
			} else {
				return WriteJSON(w, r)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
		return nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w1 := newStatWriter(w)
		start := time.Now()

		if err := f(w1, r); err != nil {
			log.Printf("E! failed: %v", err)
			http.Error(w1, err.Error(), http.StatusInternalServerError)
		}
		log.Printf("%s %s %s [%d] %d %s %s (%s)", r.RemoteAddr, r.Method, r.URL.Path, w1.StatusCode,
			w1.Count, r.Header["Referer"], r.Header["User-Agent"], time.Since(start))
	}
}

func newStatWriter(w http.ResponseWriter) *statWriter {
	return &statWriter{ResponseWriter: w, StatusCode: http.StatusOK}
}

type statWriter struct {
	http.ResponseWriter
	Count      int
	StatusCode int
}

func (s *statWriter) Write(i []byte) (int, error) {
	n, err := s.ResponseWriter.Write(i)
	s.Count += n
	return n, err
}

func (s *statWriter) WriteHeader(statusCode int) {
	s.ResponseWriter.WriteHeader(statusCode)
	s.StatusCode = statusCode
}

var _ http.ResponseWriter = (*statWriter)(nil)

var pakeCache = sync.Map{}

func setSessionKey(sessionID string, sessionKey []byte) {
	pakeCache.Store(sessionID, sessionKey)
}

func getSessionKey(sessionID string) []byte {
	sessionKey, ok := pakeCache.Load(sessionID)
	if !ok {
		return nil
	}

	if d, ok := sessionKey.([]byte); ok {
		return d
	}

	return nil
}

func servePake(w http.ResponseWriter, sessionID, code, contentCurve string) error {
	a := ss.Base64().Decode(contentCurve)
	if a.V2 != nil {
		return fmt.Errorf("base64 decode error: %w", a.V2)
	}

	b, err := pake.InitCurve([]byte(code), 1, "siec")
	if err != nil {
		return fmt.Errorf("init curve error: %w", err)
	}

	if err := b.Update(a.V1.Bytes()); err != nil {
		return fmt.Errorf("update b error: %w", err)
	}

	bb := b.Bytes()
	bk, err := b.SessionKey()
	if err != nil {
		return err
	}

	setSessionKey(sessionID, bk)
	w.Header().Set("Content-Gulp", "Curve="+ss.Base64().EncodeBytes(bb, ss.Raw, ss.Url).V1.String())
	return nil
}

// Entry is the file item for list.
type Entry struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

func servList(w http.ResponseWriter) error {
	var entries []Entry
	if err := filepath.WalkDir(RootDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		stat, err := d.Info()
		if err != nil {
			return err
		}
		entries = append(entries, Entry{
			Name: d.Name(),
			Size: stat.Size(),
		})
		return nil
	}); err != nil {
		return fmt.Errorf("walk dir %s: %w", RootDir, err)
	}
	return json.NewEncoder(w).Encode(entries)
}

func serveDownload(w http.ResponseWriter, r *http.Request, sessionID, cipher, contentRange, checksum string, chunkSize uint64, paths []string) int {
	urlPath := r.URL.Path
	for _, p := range paths {
		if strings.HasPrefix(p, urlPath) {
			p = p[len(urlPath):]
			switch p[0] {
			case ':', '=':
				urlPath = p[1:]
			}
		}
	}
	fullPath := filepath.Join(RootDir, "."+urlPath)
	stat, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return http.StatusNotFound
	}

	filename := filepath.Base(fullPath)

	if r.Method == http.MethodHead {
		w.Header().Set(ContentDisposition, mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
		return 0
	}

	if sessionID == "" {
		if err := serveMultipartDownload(w, r, fullPath, filename); err != nil {
			log.Printf("E! serveMultipartDownload failed: %v", err)
		}
		return 0
	}

	if contentRange == "" {
		totalSize := uint64(stat.Size())
		partSize := GetPartSize(totalSize, chunkSize, 0)
		cr := newChunkRange(0, chunkSize, partSize, totalSize)
		w.Header().Set("Content-Gulp", "Range="+cr.createContentRange())
		w.Header().Set(ContentType, "application/octet-stream")
		w.Header().Set(ContentDisposition, mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
		return 0
	}

	cr, err := parseContentRange(contentRange)
	if err != nil {
		log.Printf("E! parse contentRange %s failed: %v", contentRange, err)
		return http.StatusInternalServerError
	}

	if checksum != "" {
		if old, _ := readChunkChecksum(fullPath, cr.From, cr.To); old == checksum {
			log.Printf("304 file %s with session %s, range %s", filename, sessionID, contentRange)
			return http.StatusNotModified
		}
	}

	chunkReader, err := CreateChunkReader(fullPath, cr.From, cr.To, 0)
	if err != nil {
		log.Printf("E! CreateChunkReader %s failed: %v", fullPath, err)
		return http.StatusInternalServerError
	}
	defer ss.Close(chunkReader)

	salt := codec.GenSalt(8)
	key, _, err := codec.Scrypt(getSessionKey(sessionID), salt)
	if err != nil {
		log.Printf("E! new key failed: %v", err)
		return http.StatusInternalServerError
	}

	w.Header().Set(ContentType, "application/octet-stream")
	w.Header().Set(ContentDisposition, mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
	w.Header().Set("Content-Gulp", "Rang="+contentRange+"; Salt="+ss.Base64().EncodeBytes(salt, ss.Raw, ss.Url).V1.String())

	_, cipherSuites := parseCipherSuites(cipher)
	cfg := sio.Config{Key: key, CipherSuites: cipherSuites}
	if n, err := sio.Encrypt(w, chunkReader, cfg); err != nil {
		log.Printf("E! encrypt %s bytes: %d, failed: %v", fullPath, n, err)
		return http.StatusInternalServerError
	}

	log.Printf("send file %s with session %s, range %s", filename, sessionID, contentRange)
	return 0
}

func serveMultipartDownload(w http.ResponseWriter, r *http.Request, fullPath, filename string) error {
	partFrom, partTo := uint64(0), uint64(0)
	if v := r.Header.Get("Range"); v != "" {
		if cr, _ := parseRange(v); cr != nil {
			partFrom = cr.startByte
			partTo = cr.endByte
		}
	}
	chunkReader, err := CreateChunkReader(fullPath, partFrom, partTo, 0)
	if err != nil {
		return err
	}
	defer ss.Close(chunkReader)

	var dst io.Writer = w
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && !ss.HasSuffix(filename, ".gz", ".zip") {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(dst)
		defer ss.Close(gz)
		dst = gz
	} else {
		w.Header().Set(ContentLength, fmt.Sprintf("%d", chunkReader.(PayloadFileReader).FileSize()))
	}
	w.Header().Set(ContentType, "application/octet-stream")
	w.Header().Set(ContentDisposition, mime.FormatMediaType("attachment", map[string]string{"filename": filename}))

	n, err := io.Copy(dst, chunkReader)
	log.Printf("E! send file %s bytes: %d, failed: %v", fullPath, n, err)
	return nil
}

func serveBodyAsFile(src io.Reader, contentFilename string) error {
	fullPath := filepath.Join(RootDir, contentFilename)
	f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY, 0o755)
	if err != nil {
		return fmt.Errorf("open file %s error: %w", fullPath, err)
	}
	defer ss.Close(f)

	if _, err := io.Copy(f, src); err != nil {
		return fmt.Errorf("write file %s error: %w", fullPath, err)
	}

	log.Printf("file pushed %s", fullPath)
	return nil
}

type countReadCloser struct {
	n          int
	ReadCloser io.ReadCloser
}

func (c *countReadCloser) Close() error { return c.ReadCloser.Close() }

func (c *countReadCloser) Read(p []byte) (n int, err error) {
	n, err = c.ReadCloser.Read(p)
	c.n += n
	return
}

func serveUpload(w http.ResponseWriter, r *http.Request, contentRange, sessionID, cipher, contentChecksum, headerSalt string) error {
	cr, err := parseContentRange(contentRange)
	if err != nil {
		return fmt.Errorf("parse contentRange %s error: %w", contentRange, err)
	}
	_, params, err := mime.ParseMediaType(r.Header.Get(ContentDisposition))
	if err != nil {
		return fmt.Errorf("parse Content-Disposition error: %w", err)
	}

	filename := params["filename"]
	fullPath := filepath.Join(RootDir, filename)

	if r.Method == http.MethodGet {
		if contentChecksum != "" {
			if old, _ := readChunkChecksum(fullPath, cr.From, cr.To); old == contentChecksum {
				w.WriteHeader(http.StatusNotModified)
			}
		}

		return nil
	}

	salt := ss.Base64().Decode(headerSalt)
	if salt.V2 != nil {
		return salt.V2
	}
	key, _, err := codec.Scrypt(getSessionKey(sessionID), salt.V1.Bytes())
	if err != nil {
		return fmt.Errorf("scrypt error: %w", err)
	}

	f, err := openChunk(fullPath, cr)
	if err != nil {
		return err
	}
	defer ss.Close(f)

	_, cipherSuites := parseCipherSuites(cipher)

	body := &countReadCloser{ReadCloser: r.Body}
	n, err := sio.Decrypt(f, body, sio.Config{Key: key, CipherSuites: cipherSuites})
	if err != nil {
		return fmt.Errorf("decrypt %s bytes: %d, error: %w", fullPath, n, err)
	}
	if _, err := w.Write([]byte(contentRange)); err != nil {
		return fmt.Errorf("write file %s error: %w", fullPath, err)
	}

	log.Printf("recv file %s with session %s, range %s, bytes: %d, original bytes: %d",
		filename, sessionID, contentRange, n, body.n)
	return nil
}

func parseCipherSuites(cipher string) (string, []byte) {
	switch cipher {
	case "AES256":
		return "sio.AES_256_GCM", []byte{sio.AES_256_GCM}
	default: // C20P1305
		return "sio.CHACHA20_POLY1305", []byte{sio.CHACHA20_POLY1305}
	}
}
