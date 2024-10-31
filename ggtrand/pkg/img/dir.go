package img

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/tsid"
)

var Dir string

func init() {
	d, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	Dir = d
}

func ToPngBytes(img image.Image) []byte {
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		log.Panicf("failed to png.Encode, error: %v", err)
	}

	return b.Bytes()
}

// ToPng saves the image to local with PNG format.
func ToPng(img image.Image, appendBase64 bool) string {
	file := filepath.Join(Dir, tsid.Fast().ToString()+".png")
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0o755)
	if err != nil {
		log.Printf("failed to open file %s, error: %v", file, err)
		return ""
	}
	defer ss.Close(f)

	if err := png.Encode(f, img); err != nil {
		log.Printf("failed to png.Encode, error: %v", err)
		return ""
	}

	result := f.Name()

	if appendBase64 {
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return ""
		}
		data := buf.Bytes()
		result += " base64: " + base64.StdEncoding.EncodeToString(data)
	}

	return result
}

var randInt = ss.Rand().Int()

// RandomImage creates a random image.
// Environment variables supported:
// GG_IMG_FAST=Y/N to enable fast mode or not
// GG_IMG_FORMAT=jpg/png to choose the format
// GG_IMG_FILE_SIZE=10M to set image file size
// GG_IMG_SIZE=640x320 to set the {width}x{height} of image
func RandomImage(i int) interface{} {
	prefix := fmt.Sprintf("%d", randInt+i)
	img, _ := ss.Rand().Image(filepath.Join(Dir, prefix))
	return ss.IBytes(uint64(img.Size)) + " " + img.Filename
}
