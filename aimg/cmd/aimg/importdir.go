package main

import (
	"archive/zip"
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/ss"
	"github.com/gabriel-vasile/mimetype"
	"github.com/karrick/godirwalk"
	"github.com/mholt/archiver/v4"
	"github.com/mitchellh/go-homedir"
	"github.com/segmentio/ksuid"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"gorm.io/gorm"
)

func importDirImages(db, dir, pageID, title, pageLink string, dryRun, del bool, maxFiles int) error {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("recover: %+v, stack: %s", err, debug.Stack())
		}
	}()
	file, err := homedir.Expand(dir)
	if err != nil {
		return fmt.Errorf("expand dir %s: %w", dir, err)
	}

	stat, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("stat file %s: %w", file, err)
	}

	createTime := time.Now().Format(`20060102150405`)

	walker := &DirWalker{
		maxFiles:   maxFiles,
		del:        del,
		db:         db,
		pageID:     pageID,
		createTime: createTime,
		title:      cmp.Or(title, file),
		dryRun:     dryRun,
		pageLink:   pageLink,
	}
	if err := walker.Init(); err != nil {
		return err
	}

	if !stat.IsDir() {
		if err := walker.Walk(file, nil); err != nil {
			return fmt.Errorf("walk dir: %s: %w", file, err)
		}
	} else {
		options := godirwalk.Options{Unsorted: true, Callback: walker.Walk}
		for {
			if err := godirwalk.Walk(file, &options); err != nil {
				return fmt.Errorf("walk dir: %s: %w", file, err)
			}

			if !del || walker.okCount+walker.errCount == 0 || maxFiles > 0 && walker.okCount >= maxFiles {
				break
			}
		}
	}

	err = walker.saveImages()
	log.Printf("walked %d files", walker.walkedFiles)

	return err
}

var ErrMaxFiles = errors.New("max files processed")

const batchSize = 30

type DirWalker struct {
	gormDB     *gorm.DB
	db         string
	pageID     string
	createTime string
	title      string

	pageLink string

	imgs     []Img
	batchNo  int
	maxFiles int

	okCount, errCount, duplicateCount int

	walkedFiles int
	del         bool
	dryRun      bool
}

func (w *DirWalker) Init() error {
	db, err := DirectDB(w.db)
	if err != nil {
		return err
	}

	w.gormDB = db.DB
	return nil
}

func (w *DirWalker) Walk(osPathname string, dirEntry *godirwalk.Dirent) error {
	if dirEntry != nil && dirEntry.IsDir() {
		return nil
	}

	if err := w.processSingleFile(osPathname); err != nil {
		return err
	}

	if w.maxFiles > 0 && w.okCount >= w.maxFiles {
		return ErrMaxFiles
	}

	return nil
}

func (w *DirWalker) processSingleFile(osPathname string) error {
	ext := filepath.Ext(osPathname)
	switch strings.ToLower(ext) {
	case ".zip":
		if err := w.Zip(osPathname); err != nil {
			return err
		}
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
		body, err := os.ReadFile(osPathname)
		if err != nil {
			return fmt.Errorf("read file %s: %w", osPathname, err)
		}
		if err := w.processImage(osPathname, body, osPathname); err != nil {
			return err
		}
		w.walkedFiles++
		if w.walkedFiles > 0 && w.walkedFiles%100 == 0 {
			log.Printf("walked %d files", w.walkedFiles)
		}

	default:
		fsys, err := archiver.FileSystem(context.Background(), osPathname)
		if err != nil {
			log.Printf("archiver.FileSystem error: %v", err)
			return err
		}

		firstIndex := 0
		return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
			switch {
			case err != nil:
				return err
			case path == ".git":
				return fs.SkipDir
			case d.IsDir():
				return nil
			}

			switch ext := filepath.Ext(path); strings.ToLower(ext) {
			case ".jpg", ".jpeg", ".png", ".webp", ".gif":
				w.walkedFiles++
				if w.walkedFiles > 0 && w.walkedFiles%100 == 0 {
					log.Printf("walked %d files", w.walkedFiles)
				}

				info, err := d.Info()
				if err != nil {
					return err
				}
				if size := info.Size(); uint64(size) >= maxFileSize || uint64(size) <= minFileSize {
					log.Printf("ignore file %s size: %d/%s", path, size, ss.Bytes(uint64(size)))
					return nil
				}

				var body []byte
				if !w.dryRun {
					if body, err = ReadFsFile(fsys, path); err != nil {
						return err
					}
				} else {
					return nil
				}

				firstIndex++
				return w.processImage(ss.If(firstIndex == 1, osPathname, ""), body, path)
			}

			return nil
		})

	}
	return nil
}

var (
	maxFileSize, _ = ss.GetenvBytes("MAX_FILE", 100*ss.MiByte)
	minFileSize, _ = ss.GetenvBytes("MIN_FILE", 0)
)

func (w *DirWalker) processImage(addr string, body []byte, fileName string) error {
	xh := XxHash(body)
	var found Img
	if db1 := w.gormDB.Select("id", "page_id").Find(&found, "xxhash=?", xh); db1.Error != nil {
		return db1.Error
	}
	if found.ID != "" {
		log.Printf("Image Found, id: %s, xxhash: %s, size: %d/%s PageID: %s, file: %s",
			found.ID, xh, len(body), ss.Bytes(uint64(len(body))), found.PageID, fileName)
		if w.del {
			if err := os.Remove(addr); err != nil {
				log.Printf("remove %s error: %v", addr, err)
			}
		}
		return nil
	}

	w.imgs = append(w.imgs, Img{
		ID:          ksuid.New().String(),
		CreatedAt:   w.createTime,
		Xxhash:      xh,
		Body:        body,
		Size:        len(body),
		ContentType: mimetype.Detect(body).String(),
		PageLink:    w.pageLink,
		Title:       w.title,
		PageID:      w.pageID,
		Addr:        addr,
		FileName:    fileName,
	})
	w.batchNo++
	if w.batchNo > batchSize || (w.maxFiles > 0 && w.okCount >= w.maxFiles) {
		w.batchNo = 0
		w.pageID = ksuid.New().String()
	}
	if w.batchNo == 0 {
		log.Printf("pageID: %s, batchSize: %d", w.pageID, batchSize)
		return w.saveImages()
	}

	return nil
}

func (w *DirWalker) saveImages() error {
	if len(w.imgs) == 0 {
		return nil
	}

	if w.dryRun {
		w.imgs = w.imgs[:0]
		return nil
	}

	SaveImages(w.gormDB, w.imgs, &w.okCount, &w.errCount, &w.duplicateCount)

	if w.del {
		for _, img := range w.imgs {
			if img.Addr == "" {
				continue
			}
			if err := os.Remove(img.Addr); err != nil {
				log.Printf("remove %s error: %v", img.Addr, err)
			}
		}
	}

	w.imgs = w.imgs[:0]
	return nil
}

func (w *DirWalker) Zip(zipFileName string) error {
	// 打开zip文件
	zipFile, err := zip.OpenReader(zipFileName)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	firstIndex := 0
	// 遍历zip文件中的每个文件
	for _, file := range zipFile.File {
		decodeName := file.Name
		if file.Flags == 0 {
			// 如果标致位是0  则是默认的本地编码
			// https://www.yisu.com/zixun/126692.html
			// winrar压缩时，默认采用本地编码方式来进行压缩。在中国，本地编码方式一般是GBK。
			// 而我们知道go语言字符串都是utf-8格式的，所以有可能出现乱码的情况。
			i := bytes.NewReader([]byte(file.Name))
			decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
			content, _ := io.ReadAll(decoder)
			decodeName = string(content)
		}

		switch ext := filepath.Ext(decodeName); strings.ToLower(ext) {
		case ".jpg", ".jpeg", ".png", ".webp", ".gif":
			body, err := ReadZipFile(file)
			if err != nil {
				return err
			}
			firstIndex++
			if err := w.processImage(ss.If(firstIndex == 1, zipFileName, ""), body, decodeName); err != nil {
				return err
			}
		}
	}

	return nil
}

func ReadZipFile(file *zip.File) ([]byte, error) {
	r, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", file.Name, err)
	}
	defer r.Close()

	// 读取文件内容
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", file.Name, err)
	}
	return body, nil
}
