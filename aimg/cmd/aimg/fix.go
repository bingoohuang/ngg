package main

import (
	"bytes"
	"image"
	"log"
	"time"

	"github.com/corona10/goimagehash"
	"gorm.io/gorm"
)

func fixDB(c *Config) {

	fixRowsNum := 0
	start := time.Now()

	for {
		db, err := UsingDB(c.db)
		if err != nil {
			log.Panicf("using db: %v", err)
		}
		rows := scanRows(db.DB)
		if len(rows) == 0 {
			break
		}

		for _, img := range rows {
			c, format, err := image.DecodeConfig(bytes.NewReader(img.Body))
			if err != nil {
				log.Printf("解码图像配置信息时出错: %v", err)
				continue
			}

			if decoded, _, _ := image.Decode(bytes.NewReader(img.Body)); decoded != nil {
				if hash, _ := goimagehash.PerceptionHash(decoded); hash != nil {
					img.PerceptionHash = hash.ToString()
					img.PerceptionHashGroupId, _ = AssignGroupID(db.DB, img.PerceptionHash)
				}
			}

			img.Format = format
			img.Width = c.Width
			img.Height = c.Height
			if err := db.DB.Save(&img).Error; err != nil {
				log.Panicf("save image error: %v", err)
			}

			fixRowsNum++
			if time.Since(start) > 30*time.Second {
				start = time.Now()
				log.Printf("fixRowsNum: %d", fixRowsNum)
			}
		}
	}

	log.Printf("fixRowsNum: %d", fixRowsNum)
}

func scanRows(db *gorm.DB) (imgs []Img) {
	rows, err := db.Model(&Img{}).Where("perception_hash is null or perception_hash_group_id is null").Limit(100).Rows()
	if err != nil {
		log.Panicf("query error: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var img Img
		if err := db.ScanRows(rows, &img); err != nil {
			log.Panicf("scan rows error: %v", err)
		}

		imgs = append(imgs, img)
	}

	return imgs
}
