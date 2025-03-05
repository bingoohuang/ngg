package jj

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bingoohuang/ngg/ss"
)

func GenWithCache(s string) (string, error) {
	ret, err := ss.ParseExpr(s).Eval(NewCachingSubstituter())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", ret), nil
}

func NewCachingSubstituter() Substitute {
	internal := NewSubstituter(DefaultSubstituteFns)
	return &cacheValuer{MapCache: make(map[string]any), internal: internal}
}

type cacheValuer struct {
	MapCache map[string]any
	internal *Substituter
}

func (v *cacheValuer) UsageDemos() []string {
	return v.internal.UsageDemos()
}

func (v *cacheValuer) Register(fn string, f SubstituteFn) {
	v.internal.Register(fn, f)
}

var cacheSuffix = regexp.MustCompile(`^(.+)_\d+`)

func (v *cacheValuer) Value(name, params, expr string) (any, error) {
	wrapper := ""
	if p := strings.LastIndex(name, ".."); p > 0 {
		wrapper = name[p:]
		name = name[:p]
	}
	pureName := name

	subs := cacheSuffix.FindStringSubmatch(name)
	hasCachingResultTip := len(subs) > 0
	if hasCachingResultTip { // CachingSubstituter tips found
		pureName = subs[1]
		x, ok := v.MapCache[name]
		if ok {
			return x, nil
		}
	}

	x, err := v.internal.Value(pureName+wrapper, params, expr)
	if err != nil {
		return nil, err
	}

	if hasCachingResultTip {
		v.MapCache[name] = x
	}
	return x, nil
}
