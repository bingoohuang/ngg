package art

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/bingoohuang/ngg/ggtrand/pkg/img"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func Junas() string {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.Black)
	c.FillBackground()
	c.SetColorSchema(common.DarkRed)
	c.SetForeground(common.LightPink)
	c.Draw(arts.NewJanus(10, 0.2))
	// c.ToPNG("janus.png")
	return img.ToPng(c.Img(), false)
}

func RandomShapes() string {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.White)
	c.FillBackground()
	c.SetColorSchema([]color.RGBA{
		{0xCF, 0x2B, 0x34, 0xFF},
		{0xF0, 0x8F, 0x46, 0xFF},
		{0xF0, 0xC1, 0x29, 0xFF},
		{0x19, 0x6E, 0x94, 0xFF},
		{0x35, 0x3A, 0x57, 0xFF},
	})
	c.Draw(arts.NewRandomShape(150))
	return img.ToPng(c.Img(), false)
}

func ColorCircle2() string {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0x11, 0x60, 0xC6, 0xFF},
		{0xFD, 0xD9, 0x00, 0xFF},
		{0xF5, 0xB4, 0xF8, 0xFF},
		{0xEF, 0x13, 0x55, 0xFF},
		{0xF4, 0x9F, 0x0A, 0xFF},
	}
	c := generativeart.NewCanva(800, 800)
	c.SetBackground(common.White)
	c.FillBackground()
	c.SetColorSchema(colors)
	c.Draw(arts.NewColorCircle2(30))
	// c.ToPNG("colorcircle2.png")
	return img.ToPng(c.Img(), false)
}

func CircleGrid() string {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0xED, 0x34, 0x41, 0xFF},
		{0xFF, 0xD6, 0x30, 0xFF},
		{0x32, 0x9F, 0xE3, 0xFF},
		{0x15, 0x42, 0x96, 0xFF},
		{0x00, 0x00, 0x00, 0xFF},
		{0xFF, 0xFF, 0xFF, 0xFF},
	}
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(color.RGBA{R: 0xDF, G: 0xEB, B: 0xF5, A: 0xFF})
	c.FillBackground()
	c.SetColorSchema(colors)
	c.SetLineWidth(2.0)
	c.Draw(arts.NewCircleGrid(4, 6))
	// c.ToPNG("circlegrid.png")
	return img.ToPng(c.Img(), false)
}

func CircleComposesCircle() string {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0xF9, 0xC8, 0x0E, 0xFF},
		{0xF8, 0x66, 0x24, 0xFF},
		{0xEA, 0x35, 0x46, 0xFF},
		{0x66, 0x2E, 0x9B, 0xFF},
		{0x43, 0xBC, 0xCD, 0xFF},
	}
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(color.RGBA{R: 8, G: 10, B: 20, A: 255})
	c.FillBackground()
	c.SetColorSchema(colors)
	c.Draw(arts.NewCircleLoop2(7))
	// c.ToPNG("colorloop2.png")
	return img.ToPng(c.Img(), false)
}

func PixelHole() string {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0xF9, 0xC8, 0x0E, 0xFF},
		{0xF8, 0x66, 0x24, 0xFF},
		{0xEA, 0x35, 0x46, 0xFF},
		{0x66, 0x2E, 0x9B, 0xFF},
		{0x43, 0xBC, 0xCD, 0xFF},
	}
	c := generativeart.NewCanva(800, 800)
	c.SetBackground(common.Black)
	c.FillBackground()
	c.SetColorSchema(colors)
	c.SetIterations(1200)
	c.Draw(arts.NewPixelHole(60))
	// c.ToPNG("pixelhole.png")
	return img.ToPng(c.Img(), false)
}

func DotsWave() string {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0xFF, 0xBE, 0x0B, 0xFF},
		{0xFB, 0x56, 0x07, 0xFF},
		{0xFF, 0x00, 0x6E, 0xFF},
		{0x83, 0x38, 0xEC, 0xFF},
		{0x3A, 0x86, 0xFF, 0xFF},
	}
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.Black)
	c.FillBackground()
	c.SetColorSchema(colors)
	c.Draw(arts.NewDotsWave(300))
	// c.ToPNG("dotswave.png")
	return img.ToPng(c.Img(), false)
}

func ContourLine() string {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0x58, 0x18, 0x45, 0xFF},
		{0x90, 0x0C, 0x3F, 0xFF},
		{0xC7, 0x00, 0x39, 0xFF},
		{0xFF, 0x57, 0x33, 0xFF},
		{0xFF, 0xC3, 0x0F, 0xFF},
	}
	c := generativeart.NewCanva(1600, 1600)
	c.SetBackground(color.RGBA{R: 0x1a, G: 0x06, B: 0x33, A: 0xFF})
	c.FillBackground()
	c.SetColorSchema(colors)
	c.Draw(arts.NewContourLine(500))
	return img.ToPng(c.Img(), false)
}

func NoiseLine() string {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0x06, 0x7B, 0xC2, 0xFF},
		{0x84, 0xBC, 0xDA, 0xFF},
		{0xEC, 0xC3, 0x0B, 0xFF},
		{0xF3, 0x77, 0x48, 0xFF},
		{0xD5, 0x60, 0x62, 0xFF},
	}
	c := generativeart.NewCanva(1000, 1000)
	c.SetBackground(color.RGBA{R: 0xF0, G: 0xFE, B: 0xFF, A: 0xFF})
	c.FillBackground()
	c.SetColorSchema(colors)
	c.Draw(arts.NewNoiseLine(1000))
	return img.ToPng(c.Img(), false)
}

func DotLine() string {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(2080, 2080)
	c.SetBackground(color.RGBA{R: 230, G: 230, B: 230, A: 255})
	c.SetLineWidth(10)
	c.SetIterations(15000)
	c.SetColorSchema(common.DarkPink)
	c.FillBackground()
	c.Draw(arts.NewDotLine(100, 20, 50, false))
	// c.ToPNG("dotline.png")
	return img.ToPng(c.Img(), false)
}

func OceanFish() string {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0xCF, 0x2B, 0x34, 0xFF},
		{0xF0, 0x8F, 0x46, 0xFF},
		{0xF0, 0xC1, 0x29, 0xFF},
		{0x19, 0x6E, 0x94, 0xFF},
		{0x35, 0x3A, 0x57, 0xFF},
	}
	c := generativeart.NewCanva(500, 500)
	c.SetColorSchema(colors)
	c.Draw(arts.NewOceanFish(100, 8))
	return img.ToPng(c.Img(), false)
}

func CircleLoop() string {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.Black)
	c.SetLineWidth(1)
	c.SetLineColor(common.Orange)
	c.SetAlpha(30)
	c.SetIterations(1000)
	c.FillBackground()
	c.Draw(arts.NewCircleLoop(100))
	return img.ToPng(c.Img(), false)
}

func cmap(r, m1, m2 float64) color.RGBA {
	rgb := color.RGBA{
		R: uint8(common.Constrain(m1*200*r, 0, 255)),
		G: uint8(common.Constrain(r*200, 0, 255)),
		B: uint8(common.Constrain(m2*255*r, 70, 255)),
		A: 255,
	}
	return rgb
}

func DomainWarp() string {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.Black)
	c.FillBackground()
	c.Draw(arts.NewDomainWrap(0.01, 4, 4, 20, cmap))
	return img.ToPng(c.Img(), false)
}

func CircleNoise() string {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.White)
	c.SetAlpha(80)
	c.SetLineWidth(0.3)
	c.FillBackground()
	c.SetIterations(400)
	c.Draw(arts.NewCircleNoise(2000, 60, 80))
	return img.ToPng(c.Img(), false)
}

func PerlinPerls() string {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.White)
	c.SetAlpha(120)
	c.SetLineWidth(0.3)
	c.FillBackground()
	c.SetIterations(200)
	c.Draw(arts.NewPerlinPerls(10, 200, 40, 80))
	return img.ToPng(c.Img(), false)
}

func ColorCanve() string {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0xF9, 0xC8, 0x0E, 0xFF},
		{0xF8, 0x66, 0x24, 0xFF},
		{0xEA, 0x35, 0x46, 0xFF},
		{0x66, 0x2E, 0x9B, 0xFF},
		{0x43, 0xBC, 0xCD, 0xFF},
	}
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.Black)
	c.FillBackground()
	c.SetLineWidth(8)
	c.SetColorSchema(colors)
	c.Draw(arts.NewColorCanve(5))
	return img.ToPng(c.Img(), false)
}

func julia1(z complex128) complex128 {
	c := complex(-0.1, 0.651)

	z = z*z + c

	return z
}

func JuliaSet() string {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetIterations(800)
	c.SetColorSchema(common.Citrus)
	c.FillBackground()
	c.Draw(arts.NewJulia(julia1, 40, 1.5, 1.5))
	return img.ToPng(c.Img(), false)
}

func BlackHole() string {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(color.RGBA{R: 30, G: 30, B: 30, A: 255})
	c.FillBackground()
	c.SetLineWidth(1.0)
	c.SetLineColor(common.Tomato)
	c.Draw(arts.NewBlackHole(200, 400, 0.01))
	return img.ToPng(c.Img(), false)
}

func SilkSky() string {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetAlpha(10)
	c.Draw(arts.NewSilkSky(15, 5))
	return img.ToPng(c.Img(), false)
}

func CircleMove() string {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(1200, 500)
	c.SetBackground(common.White)
	c.FillBackground()
	c.Draw(arts.NewCircleMove(1000))
	return img.ToPng(c.Img(), false)
}

func RandomCircle() string {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.MistyRose)
	c.SetLineWidth(1.0)
	c.SetLineColor(color.RGBA{
		R: 122,
		G: 122,
		B: 122,
		A: 30,
	})
	c.SetColorSchema(common.Citrus)
	c.SetIterations(4)
	c.FillBackground()
	c.Draw(arts.NewRandCicle(30, 80, 0.2, 2, 10, 30, true))
	return img.ToPng(c.Img(), false)
}
