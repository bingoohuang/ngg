package songci

import (
	_ "embed"

	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/jj/randpoem"
)

func init() {
	jj.RegisterSubstituteFn("宋词", func(options string) any {
		val := RandSongci()
		return randpoem.AdaptEncoding(val, options)
	})
}

func RandSongci() string { return randpoem.SliceRandItem(SongciLines) }

var (

	//go:embed songci.txt.gz
	SongciTxtGz []byte

	SongciLines = randpoem.UnGzipLines(SongciTxtGz)
)
