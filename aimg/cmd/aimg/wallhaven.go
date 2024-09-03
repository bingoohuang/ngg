package main

import (
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bingoohuang/ngg/ss"
)

// 参考实现：https://github.com/jerryshell/wallhaven-spider

func getDocByUrl(url string) (*goquery.Document, error) {
	for i := 0; i < 3; i++ {
		log.Printf("start url: %s", url)

		response, err := client.R().Get(url)
		if err != nil {
			return nil, err
		}

		responseStatusCode := response.StatusCode
		log.Printf("end url: %s responseStatusCode: %d", url, responseStatusCode)
		if responseStatusCode != 200 {
			time.Sleep(time.Second * 2)
			continue
		}

		doc, err := goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			return nil, err
		}

		_ = response.Body.Close()

		return doc, nil
	}

	return nil, ErrRetryOut
}

var ErrRetryOut = errors.New("retry out")

func getWallhavenPageUrls(wallhavenURL string, pageNo int) ([]string, error) {
	target, err := AppendQuery(wallhavenURL, map[string]string{
		"seed": ss.Rand().String(6),
		"page": strconv.Itoa(pageNo),
	})
	if err != nil {
		return nil, err
	}
	doc, err := getDocByUrl(target)
	if err != nil {
		return nil, err
	}
	var pageUrls []string
	doc.Find("a.preview").Each(func(i int, s *goquery.Selection) {
		pageUrl, exists := s.Attr("href")
		if !exists {
			return
		}
		pageUrls = append(pageUrls, pageUrl)
	})

	return pageUrls, nil
}

func getWallhavenImageSrc(pageUrl string) string {
	doc, _ := getDocByUrl(pageUrl)
	element := doc.Find("#wallpaper")
	imageUrl, _ := element.Attr("src")
	return imageUrl
}

func GetWallhavenImageSrcs(wallhavenURL string, pageNo int) ([]string, error) {
	pageUrls, err := getWallhavenPageUrls(wallhavenURL, pageNo)
	if err != nil {
		return nil, err
	}
	var imageUrls []string
	for _, pageUrl := range pageUrls {
		imageURL := getWallhavenImageSrc(pageUrl)
		if imageURL != "" {
			imageUrls = append(imageUrls, imageURL)
		}
	}
	return imageUrls, nil
}
