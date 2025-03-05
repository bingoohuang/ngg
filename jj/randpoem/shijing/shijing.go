package shijing

import (
	_ "embed"

	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/jj/randpoem"
)

func init() {
	jj.RegisterSubstituteFn("诗经", jj.SubstituteFn{
		Fn: func(options string) any {
			val := RandShijing()
			return randpoem.AdaptEncoding(val, options)
		},
		Demo: "诗经: e.g. @诗经 @诗经(base64) @诗经(base64 url raw) @诗经(hex)",
	})
}

func RandShijing() string { return randpoem.SliceRandItem(ShijingLines) }

var (
	//go:embed shijing.txt.gz
	ShijingTxtGz []byte

	ShijingLines = randpoem.UnGzipLines(ShijingTxtGz)
)
