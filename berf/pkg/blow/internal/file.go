package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bingoohuang/ngg/berf/pkg/blow/internal/art"
	"github.com/bingoohuang/ngg/berf/pkg/util"
	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/tsid"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/gabriel-vasile/mimetype"
	"github.com/karrick/godirwalk"
	"github.com/mitchellh/go-homedir"
	"github.com/samber/lo"
)

var filePathCache sync.Map

type DataItem struct {
	Payload util.UploadPayload
}

func (d *DataItem) CreateFileField(fileFieldName string, uploadIndex bool) *util.Multipart {
	if uploadIndex {
		d.Payload.Name = insertIndexToFilename(d.Payload.Name)
	}

	return util.PrepareMultipartPayload(map[string]interface{}{
		fileFieldName: d.Payload,
	})
}

var uploadIndexVal uint64

func insertIndexToFilename(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return fmt.Sprintf("%s.%d", name, atomic.AddUint64(&uploadIndexVal, 1))
	}

	name = strings.TrimSuffix(name, ext)
	return fmt.Sprintf("%s.%d%s", name, atomic.AddUint64(&uploadIndexVal, 1), ext)
}

type UploadChanValueType int

const (
	NormalFile UploadChanValueType = iota
	DirectBytes
)

type UploadChanValue struct {
	Data        func() *DataItem
	Path        string
	ContentType string
	Type        UploadChanValueType
	UploadExit  bool
}

func (v UploadChanValue) GetCachePath() string {
	prefix := fmt.Sprintf("%d-%s-", v.Type, v.ContentType)
	if v.Type == NormalFile {
		return prefix + v.Path
	}
	return prefix + "DirectBytes"
}

type FileReader interface {
	Read(cache bool) *UploadChanValue
	Start(ctx context.Context)
}

type fileReaders struct {
	readers      []FileReader
	currentIndex int
}

func (f *fileReaders) Read(cache bool) *UploadChanValue {
	value := f.readers[f.currentIndex].Read(cache)
	if f.currentIndex++; f.currentIndex >= len(f.readers) {
		f.currentIndex = 0
	}

	return value
}

func (f fileReaders) Start(ctx context.Context) {
	for _, f := range f.readers {
		f.Start(ctx)
	}
}

func createDataItem(filePath string, isDiskFile bool, data []byte) (func() *DataItem, error) {
	var payload util.UploadPayload

	if isDiskFile {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("open file %s: %w", filePath, err)
		}
		defer ss.Close(file)

		if stat, err := file.Stat(); err != nil {
			return nil, fmt.Errorf("stat file %s: %w", filePath, err)
		} else if stat.Size() <= 10<<20 /* 10 M*/ {
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, file)

			payload = util.UploadPayload{
				Val:      buf.Bytes(),
				Original: filePath,
				Name:     changeUploadName(filePath),
				Size:     stat.Size(),
			}
		} else {
			payload = util.UploadPayload{
				DiskFile: true,
				Val:      []byte(filePath),
				Original: filePath,
				Name:     changeUploadName(filePath),
				Size:     stat.Size(),
			}
		}
	} else {
		payload = util.UploadPayload{
			Val:      data,
			Original: filePath,
			Name:     changeUploadName(filePath),
			Size:     int64(len(data)),
		}
	}

	return func() *DataItem {
		return &DataItem{Payload: payload}
	}, nil
}

func setUploadFileChanger(uploadIndex string) {
	uploadFileNameCreator = parseUploadFileChanger(uploadIndex)
}

func parseUploadFileChanger(uploadIndex string) func(filename string) string {
	if uploadIndex == "" {
		return func(filename string) string { return filename }
	}

	var idx atomic.Uint64

	return func(filename string) string {
		next := idx.Add(1)
		ext := filepath.Ext(filename)

		f, clear := FoldFindReplace(uploadIndex, "%clear", "")
		t := time.Now()
		f = FoldReplace(f, "%y", fmt.Sprintf("%04d", t.Year()))
		f = strings.ReplaceAll(f, "%M", fmt.Sprintf("%02d", t.Month()))
		f = strings.ReplaceAll(f, "%m", fmt.Sprintf("%02d", t.Minute()))
		f = FoldReplace(f, "%d", fmt.Sprintf("%02d", t.Day()))
		f = FoldReplace(f, "%H", fmt.Sprintf("%02d", t.Hour()))
		f = FoldReplace(f, "%s", fmt.Sprintf("%02d", t.Second()))

		f = FindReplaceFunc(f, `%\d*i`, func(repl string) string {
			format := repl[:len(repl)-1] + "d"
			return fmt.Sprintf(format, next)
		})
		f = FoldReplace(f, "%ext", ext)

		if clear {
			return f
		}

		dir := filepath.Dir(filename)
		base := filepath.Base(filename)
		base = base[:len(base)-len(ext)]
		return filepath.Join(dir, base+f)
	}
}

func changeUploadName(filePath string) string {
	uploadFileNameCreatorOnce.Do(func() {
		if uploadFileNameCreator == nil {
			setUploadFileChanger(os.Getenv("UPLOAD_INDEX"))
		}
	})

	return uploadFileNameCreator(filePath)
}

func FindReplaceFunc(subject, search string, repl func(string) string) string {
	searchRegex := regexp.MustCompile(search)
	if found := searchRegex.FindString(subject) != ""; found {
		return searchRegex.ReplaceAllStringFunc(subject, repl)
	}

	return subject
}

func FoldFindReplace(subject, search, replace string) (string, bool) {
	searchRegex := regexp.MustCompile("(?i)" + regexp.QuoteMeta(search))
	if found := searchRegex.FindString(subject) != ""; found {
		return searchRegex.ReplaceAllString(subject, replace), true
	}

	return subject, false
}

func FoldReplace(subject, search, replace string) string {
	searchRegex := regexp.MustCompile("(?i)" + regexp.QuoteMeta(search))
	return searchRegex.ReplaceAllString(subject, replace)
}

var (
	uploadFileNameCreator     func(filename string) string
	uploadFileNameCreatorOnce sync.Once
)

type artReader struct {
	uploadFileField string
	saveRandDir     string
}

func (r artReader) Read(cache bool) *UploadChanValue {
	uv := &UploadChanValue{
		Type:        DirectBytes,
		ContentType: "image/png",
	}

	cachePath := uv.GetCachePath()
	if cache {
		if load, ok := filePathCache.Load(cachePath); ok {
			return load.(*UploadChanValue)
		}
	}

	var err error
	data := art.Random(".png")
	uv.Path = tsid.Fast().ToString() + ".png"
	uv.Data, err = createDataItem(uv.Path, false, data)
	if err != nil {
		log.Fatal(err)
	}

	if r.saveRandDir != "" {
		util.LogErr1(os.WriteFile(filepath.Join(r.saveRandDir, uv.Path), data, os.ModePerm))
	}

	if cache {
		filePathCache.Store(cachePath, uv)
	}

	return uv
}

func (r artReader) Start(context.Context) {}

type randImgReader struct {
	uploadFileField string
	ContentType     string
	Extension       string
	saveRandDir     string
}

func (r randImgReader) Read(cache bool) *UploadChanValue {
	uv := &UploadChanValue{
		Type:        DirectBytes,
		ContentType: r.ContentType,
	}

	cachePath := uv.GetCachePath()
	if cache {
		if load, ok := filePathCache.Load(cachePath); ok {
			return load.(*UploadChanValue)
		}
	}

	c := ss.RandImgConfig{Width: 640, Height: 320, RandomText: tsid.Fast().ToString(), FastMode: false}
	data, _ := c.Gen(r.Extension)
	uv.Path = c.RandomText + r.Extension
	var err error
	uv.Data, err = createDataItem(uv.Path, false, data)
	if err != nil {
		log.Fatal(err)
	}

	if r.saveRandDir != "" {
		util.LogErr1(os.WriteFile(filepath.Join(r.saveRandDir, uv.Path), data, os.ModePerm))
	}

	if cache {
		filePathCache.Store(cachePath, uv)
	}

	return uv
}

func (r randImgReader) Start(context.Context) {}

type randJsonReader struct {
	uploadFileField string
	saveRandDir     string
}

func (r randJsonReader) Read(cache bool) *UploadChanValue {
	uv := &UploadChanValue{
		Type:        DirectBytes,
		ContentType: "application/json; charset=utf-8",
	}
	cachePath := uv.GetCachePath()
	if cache {
		if load, ok := filePathCache.Load(cachePath); ok {
			return load.(*UploadChanValue)
		}
	}

	data := jj.Rand()
	uv.Path = tsid.Fast().ToString() + ".json"
	var err error
	uv.Data, err = createDataItem(uv.Path, false, data)
	if err != nil {
		log.Fatal(err)
	}
	if r.saveRandDir != "" {
		util.LogErr1(os.WriteFile(filepath.Join(r.saveRandDir, uv.Path), data, os.ModePerm))
	}
	if cache {
		filePathCache.Store(cachePath, uv)
	}

	return uv
}

func (r randJsonReader) Start(context.Context) {}

type antReader struct {
	ch              chan string
	uploadFileField string
	pattern         string
	UploadExit      bool
}

func (f *antReader) Start(ctx context.Context) {
	basepath, pattern := doublestar.SplitPattern(f.pattern)
	f.ch = make(chan string, 1)
	errStop := fmt.Errorf("canceled")

	fn := func(p string, d fs.DirEntry) error {
		if d.IsDir() {
			return nil
		}

		select {
		case <-ctx.Done():
			return errStop
		default:
			if !strings.HasPrefix(filepath.Base(p), ".") {
				f.ch <- filepath.Join(basepath, p)
			}
		}

		return nil
	}
	go func() {
		defer close(f.ch)

		for {
			if err := doublestar.GlobWalk(
				os.DirFS(basepath),
				filepath.ToSlash(pattern),
				fn,
				doublestar.WithFailOnIOErrors(),
				doublestar.WithNoFollow(),
				doublestar.WithFilesOnly(),
			); err != nil {
				log.Printf("glob walk: %v", err)
				return
			}

			if f.UploadExit {
				break
			}
		}
	}()
}

func (f *antReader) Read(cache bool) *UploadChanValue {
	fr := &fileReader{
		File:            <-f.ch,
		uploadFileField: f.uploadFileField,
	}
	return fr.Read(cache)
}

type globReader struct {
	uploadFileField string
	matches         []string
	index           atomic.Uint64
	UploadExit      bool
}

func (g *globReader) Read(cache bool) *UploadChanValue {
	index := int(g.index.Load())
	file := g.matches[index%len(g.matches)]
	f := fileReader{
		File:            file,
		uploadFileField: g.uploadFileField,
	}
	uv := f.Read(cache)

	g.index.Add(1)

	if g.UploadExit && index == len(g.matches)-1 {
		uv.UploadExit = true
	}

	return uv
}

func (g *globReader) Start(context.Context) {}

type dirReader struct {
	Dir             string
	ch              chan string
	uploadFileField string
	UploadExit      bool
}

func (f *dirReader) Start(ctx context.Context) {
	f.ch = make(chan string, 1)
	errStop := fmt.Errorf("canceled")
	fn := func(osPathname string, dirEntry *godirwalk.Dirent) error {
		if v, e := dirEntry.IsDirOrSymlinkToDir(); v || e != nil {
			return e
		}

		select {
		case <-ctx.Done():
			return errStop
		default:
			if !strings.HasPrefix(filepath.Base(osPathname), ".") {
				f.ch <- osPathname
			}
		}

		return nil
	}

	options := godirwalk.Options{Unsorted: true, Callback: fn}
	go func() {
		defer close(f.ch)

		for {
			if err := godirwalk.Walk(f.Dir, &options); err != nil {
				log.Printf("walk dir: %s error: %v", f.Dir, err)
				return
			}

			if f.UploadExit {
				break
			}
		}
	}()
}

func (f *dirReader) Read(cache bool) *UploadChanValue {
	fr := &fileReader{
		File:            <-f.ch,
		uploadFileField: f.uploadFileField,
	}
	return fr.Read(cache)
}

type fileReader struct {
	File            string
	uploadFileField string
}

func (f fileReader) Start(context.Context) {}

func (f fileReader) Read(cache bool) *UploadChanValue {
	contentType := "application/octet-stream"
	fileMine, err := mimetype.DetectFile(f.File)
	if err == nil {
		contentType = fileMine.String()
	}
	uv := &UploadChanValue{
		Type:        NormalFile,
		ContentType: contentType,
		Path:        f.File,
	}
	uv.Data, err = createDataItem(f.File, true, nil)
	if err != nil {
		log.Fatal(err)
	}

	if !cache {
		return uv
	}

	cachePath := uv.GetCachePath()
	if load, ok := filePathCache.Load(cachePath); ok {
		return load.(*UploadChanValue)
	}

	filePathCache.Store(cachePath, uv)
	return uv
}

func CreateFileReader(uploadFileField, upload, saveRandDir string, ant bool) FileReader {
	var rr fileReaders

	if saveRandDir != "" {
		if saveDir, err := os.Stat(saveRandDir); err != nil {
			log.Printf("stat saveRandDir %s, failed: %v", saveDir, err)
			saveRandDir = ""
		} else if !saveDir.IsDir() {
			log.Printf("saveRandDir %s is not a directory", saveDir)
			saveRandDir = ""
		}
	}

	uploadExit := ss.Must(ss.GetenvBool("UPLOAD_EXIT", false))

	uploadFiles := ss.Split(upload, ",")
	for _, file := range uploadFiles {
		r := createUploadReader(file, uploadFileField, saveRandDir, uploadExit, ant)
		rr.readers = append(rr.readers, r)
	}

	return &rr
}

func createUploadReader(file, uploadFileField, saveRandDir string, uploadExit, ant bool) FileReader {
	switch file {
	case "rand.art":
		return &artReader{uploadFileField: uploadFileField, saveRandDir: saveRandDir}
	case "rand.png":
		return &randImgReader{uploadFileField: uploadFileField, ContentType: "image/png", Extension: ".png", saveRandDir: saveRandDir}
	case "rand.jpg":
		return &randImgReader{uploadFileField: uploadFileField, ContentType: "image/jpeg", Extension: ".jpeg", saveRandDir: saveRandDir}
	case "rand.json":
		return &randJsonReader{uploadFileField: uploadFileField, saveRandDir: saveRandDir}
	}

	file, _ = homedir.Expand(file)
	if stat, err := os.Stat(file); err == nil {
		if stat.IsDir() {
			return &dirReader{UploadExit: uploadExit, Dir: file, uploadFileField: uploadFileField}
		}

		return &fileReader{File: file, uploadFileField: uploadFileField}
	}

	if ant {
		if _, err := doublestar.Match(file, ""); err == nil {
			return &antReader{UploadExit: uploadExit, pattern: file, uploadFileField: uploadFileField}
		}
	}

	if matches, err := filepath.Glob(file); err == nil {
		matches = lo.Filter(matches, func(item string, index int) bool {
			base := filepath.Base(item)
			return !strings.HasPrefix(base, ".")
		})
		if ss.Must(ss.GetenvBool("UPLOAD_SHUFFLE", false)) {
			matches = lo.Shuffle(matches)
		}
		if len(matches) == 0 {
			log.Fatalf("no matched files found for: %s", file)
		}
		return &globReader{UploadExit: uploadExit, matches: matches, uploadFileField: uploadFileField}
	}

	log.Fatalf("upload %s pattern unknown", file)
	return nil
}
