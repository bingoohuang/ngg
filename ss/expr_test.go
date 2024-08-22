package ss_test

import (
	"testing"

	"github.com/bingoohuang/ngg/ss"
	"github.com/stretchr/testify/assert"
)

func subTxt(n string) *ss.SubTxt         { return &ss.SubTxt{Val: n} }
func subVar(n, p, exp string) *ss.SubVar { return &ss.SubVar{Name: n, Params: p, Expr: exp} }

func TestParseExpr(t *testing.T) {
	assert.Equal(t, ss.Subs{
		subTxt("values('"),
		subVar("random_int", "15-95", "@random_int(15-95)"),
		subTxt("','"),
		subVar("身份证", "", "@身份证"),
		subTxt("')")},
		ss.ParseExpr("values('@random_int(15-95)','@身份证')"))
	assert.Equal(t, ss.Subs{subVar("中文", "", "@中文")}, ss.ParseExpr("@中文"))
	assert.Equal(t, ss.Subs{subVar("fn", "", "@fn")}, ss.ParseExpr("@fn"))
	assert.Equal(t, ss.Subs{subVar("fn.1", "", "@fn.1")}, ss.ParseExpr("@fn.1"))
	assert.Equal(t, ss.Subs{subVar("fn-1", "", "@fn-1")}, ss.ParseExpr("@fn-1"))
	assert.Equal(t, ss.Subs{subVar("fn_1", "", "@fn_1")}, ss.ParseExpr("@fn_1"))
	assert.Equal(t, ss.Subs{subVar("fn", "", "@fn"), subTxt("@")}, ss.ParseExpr("@fn@"))
	assert.Equal(t, ss.Subs{subTxt("abc"), subVar("fn", "", "@{fn}")}, ss.ParseExpr("abc@{fn}"))
	assert.Equal(t, ss.Subs{subTxt("abc"), subVar("fn", "", "@<fn>")}, ss.ParseExpr("abc@<fn>"))
	assert.Equal(t, ss.Subs{subVar("fn", "", "@fn"), subVar("fn", "", "@fn")}, ss.ParseExpr("@fn@fn"))
	assert.Equal(t, ss.Subs{
		subTxt("abc"),
		subVar("fn", "", "@fn"),
		subVar("fn", "", "@{fn}"),
		subTxt("efg")},
		ss.ParseExpr("abc@fn@{fn}efg"))
	assert.Equal(t, ss.Subs{
		subTxt("abc"),
		subVar("fn", "", "@fn"),
		subVar("fn", "1", "@{fn(1)}"),
		subTxt("efg")},
		ss.ParseExpr("abc@fn@{fn(1)}efg"))
	assert.Equal(t, ss.Subs{
		subTxt("abc"),
		subVar("fn", "", "@fn"),
		subVar("中文", "1", "@{中文(1)}"),
		subTxt("efg")},
		ss.ParseExpr("abc@fn@{中文(1)}efg"))
	assert.Equal(t, ss.Subs{subVar("fn", "100", "@fn(100)")}, ss.ParseExpr("@fn(100)"))
	assert.Equal(t, ss.Subs{subTxt("@")}, ss.ParseExpr("@"))
	assert.Equal(t, ss.Subs{subTxt("@@")}, ss.ParseExpr("@@"))
}

func TestVars(t *testing.T) {
	m := map[string]ss.GenFnFn{
		"name": func(params string) ss.GenFn { return func() any { return "bingoo" } },
	}
	mv := ss.NewMapGenValue(m)
	s := ss.Eval("hello {name}", mv)
	assert.Equal(t, "hello bingoo", s)
	assert.Equal(t, map[string]any{"name": "bingoo"}, mv.Vars)

	s = ss.Eval("hello {{name}}", mv)
	assert.Equal(t, "hello bingoo", s)
	assert.Equal(t, map[string]any{"name": "bingoo"}, mv.Vars)

	s = ss.Eval("hello ${name}", mv)
	assert.Equal(t, "hello bingoo", s)
	assert.Equal(t, map[string]any{"name": "bingoo"}, mv.Vars)

	mv = ss.NewMapGenValue(map[string]ss.GenFnFn{})
	s = ss.Eval("hello ${name}", mv)
	assert.Equal(t, "hello name", s)
	assert.Equal(t, map[string]bool{"name": true}, mv.MissedVars)
}
