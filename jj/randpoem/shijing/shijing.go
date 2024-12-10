package shijing

import (
	_ "embed"

	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/jj/randpoem"
)

func init() {
	jj.RegisterSubstituteFn("诗经", func(options string) any {
		val := RandShijing()
		return randpoem.AdaptEncoding(val, options)
	})
}

func RandShijing() string { return randpoem.SliceRandItem(ShijingLines) }

var (
	//go:embed shijing.txt.gz
	ShijingTxtGz []byte

	ShijingLines = randpoem.UnGzipLines(ShijingTxtGz)
)
