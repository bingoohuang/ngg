package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/ss"
	"github.com/segmentio/ksuid"
	"gorm.io/gorm"
)

func QueryRandImage(baseURL string, w http.ResponseWriter, r *http.Request, db *DbCacheValue, limit int, limitN string) error {
	value, err := ksuidCache.Value(db.Name)
	if err != nil {
		return err
	}

	imgs := value.Data().(*CountCacheValue).Imgs
	offset := ss.Rand().Intn(len(imgs))
	randomID := imgs[offset].ID

	var found []Img

	if limit <= 0 {
		limit = 10
	}
	log.Printf("start to query limit %d, ID >= %s", limit, randomID)
	db1 := db.DB.Select("id", selectItems...).
		Limit(limit).Find(&found, "id >= ?", randomID)
	log.Printf("end to query limit %d, ID >= %s", limit, randomID)

	if db1.Error != nil {
		return db1.Error
	}

	return servePage(baseURL, w, r, found, limitN == "1")
}

var selectItems = []any{"xxhash", "content_type", "page_id", "size", "title", "page_link", "favorite", "height", "width", "format"}

func QueryPage(baseURL string, w http.ResponseWriter, r *http.Request, db *gorm.DB, pageID string, limit int, limitN string, desc bool) error {
	var imgs []Img
	if limit > 0 {
		db = db.Limit(limit)
	}
	if db1 := db.Select("id", selectItems...).
		Order("id"+ss.If(desc, " desc", "")).
		Find(&imgs, "page_id=?", pageID); db1.Error != nil {
		return db1.Error
	}

	return servePage(baseURL, w, r, imgs, limitN == "1")
}

func parseLimit(r *http.Request) (string, int) {
	limitN := r.URL.Query().Get("n")
	if limitN != "" {
		if ln, err := strconv.Atoi(limitN); err != nil {
			log.Printf("parse n %s error: %v", limitN, err)
		} else if ln > 0 {
			return limitN, ln
		}
	}
	return limitN, 0
}

func SaveDB(dbName string, imgs []Img, okCount, errCount, duplicateCount *int) (err error) {
	db, err := UsingDB(dbName)
	if err != nil {
		return err
	}
	SaveImages(db.DB, imgs, okCount, errCount, duplicateCount)
	return nil
}

func SaveImages(db *gorm.DB, imgs []Img, okCount, errCount, duplicateCount *int) {
	for _, img := range imgs {
		SaveImageToDB(db, img, okCount, errCount, duplicateCount)
	}
}

func SaveImageToDB(db *gorm.DB, img Img, okCount, errCount, duplicateCount *int) {
	img.HumanizeSize = strings.ReplaceAll(ss.Bytes(uint64(img.Size)), " ", "")

	if img.err != nil {
		img.Body = nil // 下面的 JSON 不用序列化该字段
		imgJSON := jj.FormatQuoteNameLeniently(ss.Json(img), jj.WithLenientValue())

		log.Printf("Get Error! image: %s, error: %v", imgJSON, img.err)
		*errCount++
		return
	}

	if img.PerceptionHash != "" {
		img.PerceptionHashGroupId, _ = AssignGroupID(db, img.PerceptionHash)
	}

	db1 := db.Create(img)
	img.Body = nil // 下面的 JSON 不用序列化该字段
	imgJSON := jj.FormatQuoteNameLeniently(ss.Json(img), jj.WithLenientValue())

	if db1.Error != nil {
		if strings.Contains(db1.Error.Error(), "UNIQUE") {
			*duplicateCount++
		} else {
			*errCount++
		}

		log.Printf("Save error! image: %s, error: %v", imgJSON, db1.Error)
		return
	}

	*okCount++
	log.Printf("OK! image: %s", imgJSON)
}

func CopyDB(fromDB, toDB string) error {
	db1, err := UsingDB(fromDB)
	if err != nil {
		return fmt.Errorf("using db %s: %w", fromDB, err)
	}
	defer func() {
		dbCache.Delete(fromDB)
	}()

	db2, err := UsingDB(toDB)
	if err != nil {
		return fmt.Errorf("using db %s: %w", toDB, err)
	}

	rows, err := db1.DB.Model(&Img{}).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	var okCount, errCount, duplicateCount int

	for rows.Next() {
		var img Img
		if err := db1.DB.ScanRows(rows, &img); err != nil {
			log.Printf("scan rows error: %v", err)
			continue
		}
		SaveImageToDB(db2.DB, img, &okCount, &errCount, &duplicateCount)
	}

	log.Printf("Copy %d rows from %s to %s successfully, %d duplicated, %d failed", okCount, fromDB, toDB, duplicateCount, errCount)
	return nil
}

func DirectDB(dbName string) (*DbCacheValue, error) {
	value := dbLoader(dbName)
	cacheValue := value.Data().(*DbCacheValue)
	return cacheValue, cacheValue.Error
}

func UsingDB(dbName string) (*DbCacheValue, error) {
	value, err := dbCache.Value(dbName)
	if err != nil {
		return nil, fmt.Errorf("get cache db: %w", err)
	}

	cacheValue := value.Data().(*DbCacheValue)
	return cacheValue, cacheValue.Error
}

func QuerySize(baseURL string, w http.ResponseWriter, r *http.Request, db *gorm.DB, size1, size2 uint64, limit int, limitN string) error {
	if limit <= 0 {
		limit = 10
	}

	limit1 := limit / 2
	limit2 := limit - limit1

	if size2 == 0 {
		var imgs1 []Img
		if limit1 > 0 {
			if db1 := db.Limit(limit1).Order("size").Find(&imgs1, "size>=?", size1); db1.Error != nil {
				return db1.Error
			}
		}

		var imgs2 []Img
		if db1 := db.Limit(limit2).Order("size desc").Find(&imgs2, "size<?", size1); db1.Error != nil {
			return db1.Error
		}

		var imgs []Img
		for i := len(imgs2) - 1; i >= 0; i-- {
			imgs = append(imgs, imgs2[i])
		}
		imgs = append(imgs, imgs1...)
		return servePage(baseURL, w, r, imgs, limitN == "1")
	}

	var imgs []Img
	if db1 := db.Limit(limit).Order("size").Find(&imgs, "size>=?", size1, "size <= ?", size2); db1.Error != nil {
		return db1.Error
	}

	return servePage(baseURL, w, r, imgs, limitN == "1")
}

func getTodayPageID(db *gorm.DB, defaultPageID string) string {
	now := time.Now()
	year, month, day := now.Date()
	midnight := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
	midnightKsuid, _ := ksuid.NewRandomWithTime(midnight)
	zeroPageID := midnightKsuid.String()
	var imgs []Img
	if db1 := db.Select("page_id").Limit(1).
		Find(&imgs, "page_id>?", zeroPageID); len(imgs) > 0 {
		log.Printf("found today's page id: %s", imgs[0].PageID)
		return imgs[0].PageID
	} else if db1.Error == nil {
		log.Printf("find today's page error: %v", db1.Error)
	}

	return defaultPageID
}

func processBasicURL(baseURL string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	switch baseURL {
	case "":
		return ""
	default:
		return path.Join("/", baseURL)
	}
}

func servePage(baseURL string, w http.ResponseWriter, r *http.Request, imgs []Img, forceDirectImage bool) error {
	if len(imgs) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	queryStyle := r.URL.Query().Get("style")
	style := strings.ReplaceAll(queryStyle, ",", ";")
	page := ListPage{
		ImageStyle: style, // max-width:600px;
	}
	page.Title = imgs[0].Title + fmt.Sprintf("(%d images)", len(imgs))
	page.BasicURL = processBasicURL(baseURL)

	if forceDirectImage {
		return serveDirectImg(w, imgs[0])
	}

	for i, img := range imgs {
		page.Images = append(page.Images, ListPageImage{
			Seq:            i + 1,
			Total:          len(imgs),
			XxHash:         img.Xxhash,
			PerceptionHash: img.PerceptionHash,
			PageID:         img.PageID,
			Size:           img.Size,
			Type:           img.ContentType,
			Title:          img.Title,
			PageLink:       img.PageLink,
			Favorite:       img.Favorite,
			CreatedTime: func() string {
				if k, err := ksuid.Parse(img.ID); err == nil {
					return k.Time().Format(time.RFC3339)
				}
				return ""
			}(),
			HumanizeSize: strings.ReplaceAll(ss.Bytes(uint64(img.Size)), " ", "") +
				fmt.Sprintf("(%dx%d)", img.Width, img.Height),
		})
	}

	return listTmpl.Execute(w, page)
}

func QueryXxHash(baseURL string, w http.ResponseWriter, r *http.Request, db *gorm.DB, xh string, limitN string) error {
	var img Img
	if db1 := db.Find(&img, "xxhash=?", xh); db1.Error != nil {
		return db1.Error
	}
	img.Fix()
	if r.Header.Get("Accept") == "application/json" || r.URL.Query().Get("format") == "json" {
		json.NewEncoder(w).Encode(img)
		return nil
	}

	return servePage(baseURL, w, r, ss.If(img.ID != "", []Img{img}, nil), limitN == "")
}

func serveDirectImg(w http.ResponseWriter, img Img) error {
	w.Header().Set("Content-Type", img.ContentType)
	_, err := w.Write(img.Body)
	return err
}

//go:embed static
var staticFS embed.FS

//go:embed list.html
var listHtml string

type ListPage struct {
	Title      string
	BasicURL   string
	Images     []ListPageImage
	ImageStyle string // max-width: 600px;
}

type ListPageImage struct {
	XxHash         string
	PerceptionHash string
	PageID         string
	Type           string
	HumanizeSize   string
	Title          string
	PageLink       string
	CreatedTime    string
	Size           int
	Seq            int
	Total          int
	Favorite       int
}

// 创建模板并解析HTML文件
var listTmpl, _ = template.New("").Parse(listHtml)
