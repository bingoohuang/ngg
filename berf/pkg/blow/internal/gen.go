package internal

import (
	"fmt"

	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/ss"
)

var Valuer = jj.NewCachingSubstituter()

type StringMode int

const (
	IgnoreJSON StringMode = iota
	MayJSON
	SureJSON
)

var gen = jj.NewGenContext(Valuer)

func Gen(s string, mode StringMode) string {
	if mode == SureJSON || mode == MayJSON && jj.Valid(s) {
		gs, _ := gen.Gen(s)
		return gs
	}

	eval, _ := ss.ParseExpr(s).Eval(Valuer)
	return fmt.Sprintf("%v", eval)
}
