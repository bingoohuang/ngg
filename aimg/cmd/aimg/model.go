package main

import (
	"bytes"
	"image"
	"mime"
)

type CollectedImages struct {
	Title           string
	Link            string
	PageID          string
	OfficialAccount string // 公众号名称
	ImageLinks      []string
}

// Img 表示一张图片
type Img struct {
	err error

	ID        string `json:",omitempty" gorm:"primarykey"`
	CreatedAt string `json:",omitempty"`

	Xxhash string `gorm:"unique"` // xxhash base64，唯一索引
	// PerceptionHash  // 感知哈希算法 (perceptual hash algorithm)，用于相似图片搜索
	PerceptionHash        string `gorm:"index"`
	PerceptionHashGroupId uint   `gorm:"index"`

	ContentType string `json:",omitempty"` // e.g. image/png
	PageLink    string `json:",omitempty"`
	Title       string `json:",omitempty"`
	PageID      string `json:",omitempty" gorm:"index"` // 图片所属页面链接 ID，可用于查找同一批图片

	Addr     string `json:",omitempty" gorm:"-"`
	FileName string `json:",omitempty" gorm:"-"`

	Body         []byte `json:",omitempty"`              // 图片内容
	Size         int    `json:",omitempty" gorm:"index"` // 图片大小，索引
	Favorite     int    `json:",omitempty"`              // 喜欢程度
	SecretKeyID  int    `json:",omitempty"`              // 加密 KeyID
	Format       string `json:",omitempty"`
	Width        int    `json:",omitempty"`
	Height       int    `json:",omitempty"`
	HumanizeSize string `json:",omitempty"`
}

func (i *Img) ExportFileName() string {
	if i.Format != "" {
		return i.Xxhash + "." + i.Format
	}

	// 使用 mime 包的 ExtensionsByType 函数获取可能的文件扩展名列表
	extensions, _ := mime.ExtensionsByType(i.ContentType)
	if len(extensions) > 0 {
		return i.Xxhash + extensions[len(extensions)-1]
	}

	reader := bytes.NewReader(i.Body)
	if _, format, _ := image.DecodeConfig(reader); format != "" {
		return i.Xxhash + "." + i.Format
	}

	return i.Xxhash
}
