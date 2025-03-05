package poetrytang

import (
	_ "embed"

	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/jj/randpoem"
)

func init() {
	jj.RegisterSubstituteFn("唐诗", jj.SubstituteFn{
		Fn: func(options string) any {
			val := RandPoetryTang()
			return randpoem.AdaptEncoding(val, options)
		},
		Demo: "唐诗: e.g. @唐诗 @唐诗(base64) @唐诗(base64 url raw) @唐诗(hex)",
	})
}

func RandPoetryTang() string { return randpoem.SliceRandItem(PoetryTangsLines) }

var (
	//go:embed poetryTang.txt.gz
	PoetryTangTxtGz []byte

	PoetryTangsLines = randpoem.UnGzipLines(PoetryTangTxtGz)
)
