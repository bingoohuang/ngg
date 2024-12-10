package poetrytang

import (
	_ "embed"

	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/jj/randpoem"
)

func init() {
	jj.RegisterSubstituteFn("唐诗", func(options string) any {
		val := RandPoetryTang()
		return randpoem.AdaptEncoding(val, options)
	})
}

func RandPoetryTang() string { return randpoem.SliceRandItem(PoetryTangsLines) }

var (
	//go:embed poetryTang.txt.gz
	PoetryTangTxtGz []byte

	PoetryTangsLines = randpoem.UnGzipLines(PoetryTangTxtGz)
)
