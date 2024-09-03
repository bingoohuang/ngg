package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

type Pixabay struct {
	okCount, errCount, duplicateCount int

	db        string
	pageID    string
	urlAddr   string
	startPage int
	endPage   int
	dryRun    bool
}

// Pixabay 主要参考 https://github.com/labmem0zero/portfolio_scrape_pixabay/blob/master/main.go
func (p *Pixabay) Pixabay() error {
	cookieDir, err := os.MkdirTemp("", "chromedp-cookie")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(cookieDir); err != nil {
			log.Printf("remove %s error: %v", cookieDir, err)
		}
	}()

	options := []chromedp.ExecAllocatorOption{
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", os.Getenv("HEAD_ON") == "0"),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("user-data-dir", cookieDir),
	}
	options = append(chromedp.DefaultExecAllocatorOptions[:], options...)
	c, _ := chromedp.NewExecAllocator(context.Background(), options...)
	ctx, cancel := chromedp.NewContext(c, chromedp.WithLogf(log.Printf))
	if err := chromedp.Run(ctx, make([]chromedp.Action, 0, 1)...); err != nil {
		return err
	}
	ctx, cancel2 := context.WithCancel(ctx)
	defer cancel2()
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		_ = <-sigChan
		cancel2()
		cancel()
	}()

	return p.keywordScrapping(ctx)
}

func AppendQuery(originalURL string, queries map[string]string) (string, error) {
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return "", err
	}

	// 添加查询参数
	queryParams := parsedURL.Query()
	for k, v := range queries {
		queryParams.Add(k, v)
	}

	parsedURL.RawQuery = queryParams.Encode()
	return parsedURL.String(), nil
}

func (p *Pixabay) keywordScrapping(ctx context.Context) error {
	done := make(chan browser.DownloadProgressState)
	chromedp.ListenTarget(ctx, func(v any) {
		if ev, ok := v.(*browser.EventDownloadProgress); ok {
			// completed := "(unknown)"
			// if ev.TotalBytes > 0 {
			// 	completed = fmt.Sprintf(", completed: %0.2f%%", ev.ReceivedBytes/ev.TotalBytes*100.0)
			// }
			// log.Printf("state: %s%s\n", ev.State.String(), completed)
			switch ev.State {
			case browser.DownloadProgressStateCompleted, browser.DownloadProgressStateCanceled:
				done <- ev.State
			}
		}
	})

	baseDir, err := os.MkdirTemp("", "chromedp-pixabay")
	if err != nil {
		return err
	}

	process := &EachProcess{
		ctx:     ctx,
		baseDir: baseDir,
		done:    done,
		Pixabay: p,
	}

	for i := p.startPage; i <= p.endPage; i++ {
		pageURL, err := AppendQuery(p.urlAddr, map[string]string{"pagi": strconv.Itoa(i)})
		if err != nil {
			return err
		}

		res, err := ScollHtmlPage(ctx, pageURL)
		if err != nil {
			return err
		}
		if err := process.ScrapeImages(res); err != nil {
			return err
		}

		if i < p.endPage {
			time.Sleep(1 * time.Second)
		}
	}
	return nil
}

func ScollHtmlPage(ctx context.Context, url string) (htmlContent string, err error) {
	err = chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if _, exp, err := runtime.Evaluate(`window.scrollTo(0,document.body.scrollHeight);`).Do(ctx); err != nil {
				return err
			} else if exp != nil {
				return exp
			}

			time.Sleep(1000 * time.Millisecond)
			return nil
		}),
		chromedp.InnerHTML("*", &htmlContent, chromedp.ByQuery),
	)

	return
}

func (p *EachProcess) ScrapeImages(html string) error {
	tempReader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(tempReader)
	if err != nil {
		return err
	}

	doc.Find(`div[class*='container--'] a[class*='link--']`).EachWithBreak(p.DoEach)

	return nil
}

type EachProcess struct {
	ctx     context.Context
	Pixabay *Pixabay
	done    chan browser.DownloadProgressState
	baseDir string
	max     int
	count   int
}

func (p *EachProcess) DoEach(_ int, img *goquery.Selection) bool {
	if p.max > 0 && p.count >= p.max {
		return false
	}

	tmpUrl, exists := img.Attr("href")
	if !exists {
		return true
	}
	link, err := url.JoinPath("https://pixabay.com/", tmpUrl)
	if err != nil {
		log.Print(err)
		return true
	}

	log.Printf("visiting %s", link)

	var descriptionH1 string
	err = chromedp.Run(p.ctx,
		chromedp.Navigate(link),
		browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllow).
			WithDownloadPath(p.baseDir).WithEventsEnabled(true),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Click(`button[class*='fullWidthTrigger--']`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Click(`div[class*='radioGroupContainer--'] > label[class*='input--']:nth-of-type(3)`, chromedp.ByQuery),
		chromedp.Click(`div[class*='buttons--'] > a[class*='buttonBase--']`, chromedp.ByQuery),
		chromedp.InnerHTML("div[class*='descriptionSection--'] h1", &descriptionH1, chromedp.ByQuery),
	)
	if err != nil {
		log.Print(err)
		return true
	}

	state := <-p.done
	if state == browser.DownloadProgressStateCanceled {
		return true
	}

	var imgPath string

	if err := filepath.WalkDir(p.baseDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			relativeDir := strings.TrimPrefix(path, p.baseDir)
			if depth := strings.Count(relativeDir, string(os.PathSeparator)); depth > 1 {
				return fs.SkipDir
			}
			return nil
		}

		imgPath = path
		return nil
	}); err != nil {
		log.Print(err)
	}

	if imgPath != "" {
		p.count++
		log.Printf("image: %s, description: %s", imgPath, descriptionH1)

		pp := p.Pixabay
		if err := importDirImages(pp.db, imgPath, pp.pageID, descriptionH1, link, pp.dryRun, true, 0); err != nil {
			log.Printf("import dir error: %v", err)
		}
	}

	return true
}

func UrlExistence(ctx context.Context, urlAddr string) bool {
	resp, err := chromedp.RunResponse(ctx,
		chromedp.Navigate(urlAddr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		}),
	)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return resp.URL == urlAddr
}
