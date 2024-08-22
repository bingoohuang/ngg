package ss

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/bingoohuang/ngg/unit"
	"github.com/pbnjay/pixfont"
)

type RandomImageResult struct {
	Size     int
	Filename string
}

// Image creates a random image.
// Environment variables supported:
// GG_IMG_FAST=Y/N to enable fast mode or not
// GG_IMG_FORMAT=jpg/png to choose the format
// GG_IMG_FILE_SIZE=10M to set image file size
// GG_IMG_SIZE=640x320 to set the {width}x{height} of image
func (r Rand) Image(prefix string) (*RandomImageResult, error) {
	imgFormat := r.parseImageFormat("IMG_FMT")
	width, height := parseImageSize("IMG_SIZE")
	fn := fmt.Sprintf("%s_%dx%d%s", prefix, width, height, imgFormat)
	c := RandImgConfig{
		Width:      width,
		Height:     height,
		RandomText: prefix,
		FastMode:   Pick1(GetenvBool("IMG_FAST", false)),
	}
	size := c.GenFile(fn, int(Pick1(unit.GetEnvBytes("IMG_FILE_SIZE", 0))))
	return &RandomImageResult{Size: size, Filename: fn}, nil
}

func (r Rand) parseImageFormat(envName string) string {
	if v := os.Getenv(envName); v != "" {
		switch strings.ToLower(v) {
		case ".jpg", "jpg", ".jpeg", "jpeg":
			return ".jpg"
		case ".png", "png":
			return ".png"
		}
	}
	return If(r.Bool(), ".jpg", ".png")
}

func parseImageSize(envName string) (width, height int) {
	width, height = 640, 320
	if val := os.Getenv(envName); val != "" {
		val = strings.ToLower(val)
		parts := strings.SplitN(val, "x", 2)
		if len(parts) == 2 {
			if v, _ := Parse[int](parts[0]); v > 0 {
				width = v
			}
			if v, _ := Parse[int](parts[1]); v > 0 {
				height = v
			}
		}
	}
	return width, height
}

// GenerateRandomImageFile generate image file.
// If fastMode is true, a sparse file is filled with zero (ascii NUL) and doesn't actually take up the disk space
// until it is written to, but reads correctly.
// $ ls -lh 424661641.png
// -rw-------  1 bingoobjca  staff   488K Mar 15 12:19 424661641.png
// $ du -hs 424661641.png
// 8.0K    424661641.png
// If fastMode is false, an actually sized file will generated.
// $ ls -lh 1563611881.png
// -rw-------  1 bingoobjca  staff   488K Mar 15 12:28 1563611881.png
// $ du -hs 1563611881.png
// 492K    1563611881.png

type RandImgConfig struct {
	Width      int
	Height     int
	RandomText string
	FastMode   bool
	PixelSize  int
}

func (c *RandImgConfig) GenFile(filename string, fileSize int) int {
	f, _ := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0o600)
	defer f.Close()

	data, imgSize := c.Gen(filepath.Ext(filename))
	f.Write(data)
	if fileSize <= imgSize {
		return imgSize
	}

	if !c.FastMode {
		b := Rand{}.Bytes(fileSize - imgSize)
		f.Write(b)
		return fileSize
	}

	// refer to https://stackoverflow.com/questions/16797380/how-to-create-a-10mb-file-filled-with-000000-data-in-golang
	// use f.Truncate to change size of the file
	// If you are using unix, then you can create a sparse file very quickly.
	// A sparse file is filled with zero (ascii NUL) and doesn't actually take up the disk space
	// until it is written to, but reads correctly.
	f.Truncate(int64(fileSize))
	return fileSize
}

// Gen generate a random image with imageFormat (jpg/png) .
// refer: https://onlinejpgtools.com/generate-random-jpg
func (c *RandImgConfig) Gen(imageFormat string) ([]byte, int) {
	var img draw.Image

	format := strings.ToLower(imageFormat)
	switch format {
	case ".jpg", ".jpeg":
		img = image.NewNRGBA(image.Rect(0, 0, c.Width, c.Height))
	default: // png
		img = image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	}

	if c.PixelSize == 0 {
		c.PixelSize = 40
	}

	yp := c.Height / c.PixelSize
	xp := c.Width / c.PixelSize
	for yi := 0; yi < yp; yi++ {
		for xi := 0; xi < xp; xi++ {
			drawPixelWithColor(img, yi, xi, c.PixelSize, Rand{}.Color())
		}
	}

	if c.RandomText != "" {
		pixfont.DrawString(img, 10, 10, c.RandomText, color.Black)
	}

	var buf bytes.Buffer
	switch format {
	case ".jpg", ".jpeg":
		jpeg.Encode(&buf, img, &jpeg.Options{Quality: 100}) // 图像质量值为100，是最好的图像显示
	default: // png
		png.Encode(&buf, img)
	}

	return buf.Bytes(), buf.Len()
}

// drawPixelWithColor draw pixels on img from yi, xi and randomColor with size of pixelSize x pixelSize
func drawPixelWithColor(img draw.Image, yi, xi, pixelSize int, c color.Color) {
	ys := yi * pixelSize
	ym := ys + pixelSize
	xs := xi * pixelSize
	xm := xs + pixelSize

	for y := ys; y < ym; y++ {
		for x := xs; x < xm; x++ {
			img.Set(x, y, c)
		}
	}
}

// Color generate a random color
func (r Rand) Color() color.Color {
	return color.RGBA{
		R: uint8(r.Intn(255)),
		G: uint8(r.Intn(255)),
		B: uint8(r.Intn(255)),
		A: uint8(r.Intn(255)),
	}
}
