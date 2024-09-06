package main

import (
	"cmp"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/alitto/pond"
	"github.com/bingoohuang/aimg/tmpjson"
	"github.com/bingoohuang/ngg/daemon"
	"github.com/bingoohuang/ngg/jj"
	_ "github.com/bingoohuang/ngg/rotatefile/stdlog/autoload"
	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/ver"
	"github.com/golang-module/carbon/v2"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/segmentio/ksuid"
)

const defaultListen = ":1100"

type Config struct {
	db         string
	pageID     string
	baseURL    string
	export     string
	timeout    time.Duration
	dryRun     bool
	foreground bool
	version    bool
	debug      bool

	page    int
	workers int
	minSize int

	StatChan chan StatPage
}

type StatPage struct {
	Link      string
	Title     string
	PageError string

	Retries        int
	OkCount        int
	FailCount      int
	DuplicateCount int
}

type AimgState struct {
	PageID string `json:"pageID"`
}

// ArgSize 是一个自定义类型，用于存储解析后的文件大小
type ArgSize struct {
	Value  uint64
	HasSet bool
}

// String 方法返回文件大小的字符串表示形式
func (a *ArgSize) String() string {
	return ss.Bytes(a.Value)
}

// Set 方法解析命令行参数并将其转换为 int64
func (a *ArgSize) Set(value string) error {
	size, err := ss.ParseBytes(value)
	if err != nil {
		return err
	}
	a.HasSet = true
	a.Value = size
	return nil
}

func parseFlags() (c *Config) {
	c = &Config{
		StatChan: make(chan StatPage),
	}
	flag.DurationVar(&c.timeout, "timeout", 10*time.Minute, "timeout")
	flag.StringVar(&c.db, "db", "~/.aimg.db", "db path(can be link like: ln -s /other/path/.aimg.db ~/.aimg.db)")
	flag.StringVar(&c.pageID, "pageID", os.Getenv("PAGE_ID"), "specified pageID")
	flag.StringVar(&c.baseURL, "baseurl", os.Getenv("BASE_URL"), "base url")
	flag.StringVar(&c.export, "export", "", "export images to dir. e.g. n=10,size=1M")
	flag.BoolVar(&c.foreground, "fg", false, "run in foreground")
	flag.BoolVar(&c.dryRun, "dry", false, "dry run in image importing (for debug)")
	flag.BoolVar(&c.version, "version", false, "print version")
	flag.BoolVar(&c.debug, "debug", false, "enable debug moed")
	flag.IntVar(&c.page, "page", 1, "page no (for pixabay/wallhaven)")
	flag.IntVar(&c.workers, "workers", 10, "goroutine workers")

	sizeArg := ArgSize{
		Value: 20 * 1024,
	}
	flag.Var(&sizeArg, "minSize", "min image size with units, e.g. 1K, 1M, etc. (default 20K)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: aimg [options] [url1] [url2] [imageURL] [importDbFilePath] [/dir/img.png] [/dir/imgdir:X] [:port] \n\n")
		fmt.Fprintf(os.Stderr, "Args:\n\n1. [/dir/imgdir:X] delete image files after import\n")
		fmt.Fprintf(os.Stderr, "2. [url] collection images from url address\n")
		fmt.Fprintf(os.Stderr, "3. [port] listen port, default "+defaultListen+"\n\nOptions:\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if c.version {
		fmt.Printf("%s\n", ver.Version())
		os.Exit(0)
	}

	c.minSize = int(sizeArg.Value)
	c.reuseTodayPageID()
	c.expandDBFile()

	if c.export != "" {
		exportImages(c)
		os.Exit(0)
	}

	return c
}

func exportImages(c *Config) {
	db, err := UsingDB(c.db)
	if err != nil {
		log.Panicf("using db: %v", err)
	}
	value, err := ksuidCache.Value(db.Name)
	if err != nil {
		log.Panicf("ksuidCache: %v", err)
	}

	imgs := value.Data().(*CountCacheValue).Imgs
	offset := ss.Rand().Intn(len(imgs))
	randomID := imgs[offset].ID

	limit, conds := parseExport(randomID, c)

	log.Printf("start to query limit %d, ID >= %s", limit, randomID)
	var found []Img
	db1 := db.DB.Select("*").
		Limit(limit).Find(&found, conds...)
	log.Printf("end to query limit %d, ID >= %s", limit, randomID)
	if db1.Error != nil {
		log.Panicf("select: %v", db1.Error)
	}

	for _, img := range found {
		os.WriteFile(img.ExportFileName(), img.Body, os.ModePerm)
		img.Body = nil // 下面的 JSON 不用序列化该字段
		imgJSON := jj.FormatQuoteNameLeniently(ss.Json(img), jj.WithLenientValue())
		log.Printf("image exported: %s", imgJSON)
	}
}

// parseExport 解析 形如 -export 'n=10&size=1M/2M' 的参数
func parseExport(randomID string, c *Config) (limit int, conds []any) {
	condition := "id >= ?"
	vars := []any{randomID}

	limit = 10
	exportParams, _ := url.ParseQuery(c.export)
	if exportParams.Has("n") {
		if n, _ := strconv.Atoi(exportParams.Get("n")); n > 0 {
			limit = n
		}
	}
	if exportParams.Has("size") {
		if fromSize, toSize, err := parseSizeExpr(exportParams.Get("size")); err == nil {
			if fromSize > 0 {
				condition += " and size > ?"
				vars = append(vars, fromSize)
			}
			if toSize > 0 {
				condition += " and size < ?"
				vars = append(vars, toSize)
			}
		}
	}
	conds = append(conds, condition)
	conds = append(conds, vars...)
	return limit, conds
}

func (c *Config) expandDBFile() {
	var err error
	if c.db, err = ss.ExpandFilename(c.db); err != nil {
		if !os.IsNotExist(err) {
			log.Panicf("Expand %s error: %v", c.db, err)
		}
	} else {
		log.Printf("db file: %s", c.db)
	}
}

func (c *Config) reuseTodayPageID() {
	if c.pageID != "" {
		return
	}

	if state, err := tmpjson.Read("aimg.json", &AimgState{}); err == nil {
		if k, err := ksuid.Parse(state.PageID); err == nil {
			if carbon.CreateFromStdTime(k.Time()).IsToday() {
				log.Printf("reuse today's pageID: %s", state.PageID)
				c.pageID = state.PageID
			}
		}
	}

	if c.pageID == "" {
		c.pageID = ksuid.New().String()
		_ = tmpjson.Write("aimg.json", AimgState{PageID: c.pageID})
	}
}

func main() {
	c := parseFlags()

	pool := pond.New(c.workers, c.workers)
	listen := c.handleArgs(pool)

	go func() {
		pool.StopAndWait()
		close(c.StatChan)
	}()

	c.summary(listen)
	dealListen(c, listen, c.db)
	log.Printf("Exited!")
}

func (c *Config) dealPixabay(_ ksuid.KSUID, pixabayURL string) {
	p := &Pixabay{
		db:        c.db,
		pageID:    c.pageID,
		urlAddr:   pixabayURL,
		startPage: c.page,
		endPage:   c.page,
		dryRun:    c.dryRun,
	}
	if err := p.Pixabay(); err != nil {
		log.Printf("pixabay: %v", err)
	}
}

func dealListen(c *Config, listen, db string) {
	if listen == "" && len(flag.Args()) == 0 {
		listen = defaultListen
	}

	if listen == "" {
		return
	}

	if !c.foreground {
		daemon.Option{}.Daemonize()
	}

	log.Printf("listen on %s", listen)

	f := func(w http.ResponseWriter, r *http.Request) {
		if err := handle(c.baseURL, c.pageID, db, c.debug, w, r); err != nil {
			log.Printf("handle error: %v", err)
		}
	}
	baseURL := path.Clean(path.Join("/", c.baseURL))
	http.Handle("/", http.StripPrefix(baseURL, http.HandlerFunc(f)))
	if err := http.ListenAndServe(listen, nil); err != nil {
		log.Printf("ListenAndServe error: %v", err)
	}
}

func (c *Config) dealWallHaven(baseKsuid ksuid.KSUID, wallhavenURL string) {
	if srcs, err := GetWallhavenImageSrcs(wallhavenURL, c.page); err != nil {
		log.Printf("GetWallhavenImageSrcs error: %v", err)
	} else {
		var okCount, errCount, duplicateCount int
		for _, src := range srcs {
			baseKsuid := baseKsuid.Next()
			downloadURLImage(baseKsuid, c.pageID, c.db, src, "wallhaven.cc",
				c.timeout, &okCount, &errCount, &duplicateCount, c.minSize)
		}
	}
}

func (c *Config) handleArgs(pool *pond.WorkerPool) string {
	listen := ""
	for _, arg := range flag.Args() {
		arg := arg
		baseKsuid := ksuid.New()

		if AnyPrefix(arg, "https://", "http://") != "" {
			var f func()
			switch {
			// e.g. https://wallhaven.cc/random
			case AnyPrefix(arg, "https://wallhaven.cc/", "http://wallhaven.cc/") != "":
				f = func() { c.dealWallHaven(baseKsuid, arg) }
			// e.g. https://pixabay.com/users/elf-moondance-19728901/
			case AnyPrefix(arg, "https://pixabay.co", "http://pixabay.co") != "":
				f = func() { c.dealPixabay(baseKsuid, arg) }
			case AnyPrefix(arg, "https://mp.weixin.qq.com/", "http://mp.weixin.qq.com/") != "":
				f = func() { c.collectWeixinImages(baseKsuid, arg, c.minSize) }
			default:
				f = func() { c.collectWeixinImages(baseKsuid, arg, c.minSize) }
			}

			pool.Submit(f)
			continue
		}

		argx := strings.TrimSuffix(arg, ":X")
		argx, _ = ss.ExpandAtFile(argx)
		if _, err := os.Stat(argx); err == nil {
			pool.Submit(func() {
				if strings.HasSuffix(argx, ".db") {
					if err != nil {
						log.Printf("ExpandFile %s error: %v", argx, err)
					} else if err := CopyDB(argx, c.db); err != nil {
						log.Printf("copyDB error: %v", err)
					}
				} else {
					del := strings.HasSuffix(arg, ":X")
					if err := importDirImages(c.db, argx, c.pageID, "", "", c.dryRun, del, 0); err != nil {
						log.Printf("import dir error: %v", err)
					}
				}
			})
			continue
		}

		if listen == "" {
			listen = arg
		}
	}

	return listen
}

func (c *Config) summary(listen string) {
	t := table.NewWriter()

	i := 0

	var okCount, failCount, duplicateCount int
	for stat := range c.StatChan {
		i++
		okCount += stat.OkCount
		failCount += stat.FailCount
		duplicateCount += stat.DuplicateCount
		t.AppendRow([]any{i, stat.Title, stat.OkCount, stat.FailCount, stat.DuplicateCount, stat.PageError, stat.Retries, stat.Link})
	}

	if i > 0 {
		t.SetOutputMirror(os.Stdout)
		style := table.StyleDefault
		style.Format.Header = text.FormatDefault
		t.SetStyle(style)
		t.AppendHeader([]any{"#", "Title", "OK", "Fail", "Duplicate", "Error", "Retries", "Link"})
		t.Render()

		log.Printf("Total %d image saved, %d failed, %d duplicated!", okCount, failCount, duplicateCount)
		go func() {
			browserURL, _ := ss.OpenInBrowser(cmp.Or(listen, defaultListen), "/P/"+c.pageID)
			log.Printf("Link: %s", browserURL)
		}()
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}
}
