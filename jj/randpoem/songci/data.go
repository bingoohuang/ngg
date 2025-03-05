package songci

import (
	_ "embed"

	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/jj/randpoem"
)

func init() {
	jj.RegisterSubstituteFn("宋词", jj.SubstituteFn{
		Fn: func(options string) any {
			val := RandSongci()
			return randpoem.AdaptEncoding(val, options)
		},
		Demo: "宋词: e.g. @宋词 @宋词(base64) @宋词(base64 url raw) @宋词(hex)",
	})
}

func RandSongci() string { return randpoem.SliceRandItem(SongciLines) }

var (

	//go:embed songci.txt.gz
	SongciTxtGz []byte

	SongciLines = randpoem.UnGzipLines(SongciTxtGz)
)
