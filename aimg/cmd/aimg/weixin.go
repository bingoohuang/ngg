package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/tick"
	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
	"github.com/segmentio/ksuid"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
)

func (c *Config) collectWeixinImages(baseKsuid ksuid.KSUID, pageLink string, minSize int) {
	it := StatPage{
		Link: pageLink,
	}
	defer func() {
		c.StatChan <- it
	}()

	for ; it.Retries < 3; it.Retries++ {
		it.PageError = ""
		it.OkCount, it.FailCount, it.DuplicateCount = 0, 0, 0
		func() {
			imageLinks, err := CollectWebImageLinks(c.pageID, pageLink)
			if err != nil {
				it.PageError = err.Error()
				return
			}

			it.Title = imageLinks.Title

			imgs := DownloadImages(baseKsuid, imageLinks, c.timeout, minSize)
			if err := SaveDB(c.db, imgs, &it.OkCount, &it.FailCount, &it.DuplicateCount); err != nil {
				it.PageError = err.Error()
				log.Printf("saveDB error: %v", err)
			}
		}()

		if it.OkCount > 0 || it.FailCount > 1 /* 当前不是图片连接错误 */ || it.DuplicateCount > 0 {
			break
		}
		time.Sleep(3 * time.Second)
	}
}

func downloadURLImage(baseKsuid ksuid.KSUID, pageID, db, link, title string, timeout time.Duration, okCount, errCount, duplicateCount *int, minSize int) {
	imageLinks := &CollectedImages{
		ImageLinks: []string{link},
		PageID:     pageID,
		Link:       link,
		Title:      title,
	}
	imgs := DownloadImages(baseKsuid, imageLinks, timeout, minSize)
	if err := SaveDB(db, imgs, okCount, errCount, duplicateCount); err != nil {
		log.Printf("saveDB error: %v", err)
	}
}

func DownloadImages(baseKsuid ksuid.KSUID, images *CollectedImages, timeout time.Duration, minSize int) []Img {
	imgs := make([]Img, 0, len(images.ImageLinks))
	createTime := time.Now().Format(`20060102150405`)

	for _, imageLink := range images.ImageLinks {
		img := downloadImage(baseKsuid.String(), imageLink, images, createTime, timeout, minSize)
		baseKsuid = baseKsuid.Next()
		imgs = append(imgs, img)
	}

	return imgs
}

var client = req.C()

func downloadImage(id, addr string, images *CollectedImages, createTime string, timeout time.Duration, minSize int) Img {
	log.Printf("Downloaded: %s", addr)

	img := Img{
		ID:        id,
		CreatedAt: createTime,
		Title:     images.Title,
		Addr:      addr,
		PageID:    images.PageID,
		PageLink:  images.Link,
	}
	for i := 0; i < 3; i++ {
		rsp, err := client.SetTimeout(timeout).R().Get(addr)
		if err != nil {
			if errors.Is(err, io.EOF) {
				time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
				continue
			}
			log.Printf("Download %s error: %v", addr, err)
			img.err = err
			return img
		}

		contentType := rsp.GetHeader("Content-Type")
		if !strings.Contains(contentType, "image/") {
			img.err = fmt.Errorf("%s is not an image", contentType)
			return img
		}

		img.Body = rsp.Bytes()
		img.Size = len(img.Body)
		img.ContentType = contentType
		img.Xxhash = XxHash(img.Body)

		reader := bytes.NewReader(img.Body)
		widthHeight := false
		if c, format, err := image.DecodeConfig(reader); err != nil {
			log.Printf("解码图像配置信息时出错: %v", err)
		} else {
			img.Format = format
			img.Width = c.Width
			img.Height = c.Height
			widthHeight = true
		}

		if img.Size < minSize || (widthHeight && (img.Width < 500 || img.Height < 500)) {
			img.err = fmt.Errorf("image is small than %d, %dx%d", minSize, img.Width, img.Height)
			return img
		}
		break
	}

	if images.OfficialAccount != "" {
		img.Title = images.OfficialAccount + "::" + img.Title
	}

	return img
}

var linksRegex = regexp.MustCompile(`https?://mmbiz\.qpic\.cn.*?['"]`)

func CollectWebImageLinks(pageID, pageLink string) (*CollectedImages, error) {
	result := &CollectedImages{
		Link:   pageLink,
		PageID: pageID,
	}

	c := colly.NewCollector()
	// 选择器： https://www.w3.org/TR/selectors-3/#selectors
	c.OnHTML(`a.weui-wa-hotarea`, func(e *colly.HTMLElement) {
		result.OfficialAccount = strings.TrimSpace(e.Text) // 公众号
	})
	c.OnHTML(`div.wx_follow_nickname`, func(e *colly.HTMLElement) {
		result.OfficialAccount = strings.TrimSpace(e.Text) // 公众号
	})
	c.OnHTML(`meta[property]`, func(e *colly.HTMLElement) {
		if result.Title != "" {
			return
		}
		switch property := e.Attr("property"); property {
		case "og:title", "twitter:title":
			result.Title = e.Attr("content")
		}
	})
	c.OnHTML("script", func(h *colly.HTMLElement) {
		links := linksRegex.FindAllString(h.Text, -1)
		for _, link := range links {
			result.ImageLinks = append(result.ImageLinks, link[:len(link)-1])
		}
	})
	c.OnHTML("img[data-src]", func(e *colly.HTMLElement) {
		dataSrc := e.Attr("data-src")
		result.ImageLinks = append(result.ImageLinks, dataSrc)
	})
	c.OnHTML("li[data-link]", func(e *colly.HTMLElement) {
		dataLink := e.Attr("data-link")
		if err := e.Request.Visit(dataLink); err != nil {
			log.Printf("visit %s error: %v", dataLink, err)
		}
	})
	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		log.Printf("Visiting %s", r.URL)
	})

	if err := Retry(func(context.Context) error { return c.Visit(pageLink) },
		WithRetryIf(func(err error) bool {
			log.Printf("error: %v", err)
			// protocol error: received DATA on a HEAD request
			return strings.Contains(err.Error(), "protocol error")
		}),
	); err != nil {
		return nil, err
	}

	if len(result.ImageLinks) == 0 { // maybe a direct image URL
		result.ImageLinks = append(result.ImageLinks, pageLink)
	}

	return result, nil
}

type RetryOption struct {
	MaxRetryTimes int
	context.Context
	Jitter  time.Duration
	Delay   time.Duration
	RetryIf func(error) bool
}

type RetryOptionFn func(*RetryOption)

func WithMaxRetryTimes(times int) RetryOptionFn {
	return func(ro *RetryOption) { ro.MaxRetryTimes = times }
}

func WithContext(ctx context.Context) RetryOptionFn {
	return func(ro *RetryOption) { ro.Context = ctx }
}

func WithDelay(delay, jitter time.Duration) RetryOptionFn {
	return func(ro *RetryOption) {
		ro.Delay = delay
		ro.Jitter = jitter
	}
}

func WithRetryIf(f func(error) bool) RetryOptionFn {
	return func(ro *RetryOption) {
		ro.RetryIf = f
	}
}

func Retry(f func(context.Context) error, fns ...RetryOptionFn) error {
	option := RetryOption{
		Delay:  time.Second,
		Jitter: time.Second,
	}
	for _, fn := range fns {
		fn(&option)
	}

	if option.Context == nil {
		option.Context = context.Background()
	}

	var err error
	for i := 0; option.MaxRetryTimes <= 0 || i < option.MaxRetryTimes; i++ {
		select {
		case <-option.Context.Done():
			return option.Context.Err()
		default:
		}
		if err := f(option.Context); err == nil {
			return nil
		} else if option.RetryIf != nil && !option.RetryIf(err) {
			return err
		}

		if option.Delay > 0 {
			tick.SleepRandom(option.Context, tick.Jitter(option.Delay, option.Jitter))
		}

		log.Printf("retrying...")
	}
	return err
}
