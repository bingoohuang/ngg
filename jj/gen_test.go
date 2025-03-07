package jj_test

import (
	"fmt"
	"testing"

	"github.com/bingoohuang/ngg/jj"
	"github.com/stretchr/testify/assert"
)

func TestGenKeyHitRepeat(t *testing.T) {
	ret, _ := jj.NewGen().Gen(`{"id|2": "1" }`)
	assert.Equal(t, `{"id":"11"}`, ret)
}

func TestGenKeyHitRepeatObjectId(t *testing.T) {
	gen := jj.NewGen()
	gen.RegisterFn("对象ID", jj.SubstituteFn{Fn: func(args string) any { return 456 }, Demo: "对象ID, e.g. @对象ID"})
	ret, _ := gen.Gen(`{"id|2": "@对象ID" }`)
	assert.Equal(t, `{"id":"456456"}`, ret)
}

func TestGenRepeatObject(t *testing.T) {
	ret, _ := jj.NewGen().Gen(`["|2", { "id": 123 }]`)
	assert.Equal(t, `[{"id":123},{"id":123}]`, ret)
}

func TestGenRepeatString(t *testing.T) {
	ret, _ := jj.NewGen().Gen(`["|2", "123"]`)
	assert.Equal(t, `["123","123"]`, ret)
}

func TestGenRepeatInt(t *testing.T) {
	ret, _ := jj.NewGen().Gen(`["|2", 123]`)
	assert.Equal(t, `[123,123]`, ret)
}

func TestGenRepeatInt2(t *testing.T) {
	ret, _ := jj.NewGen().Gen(`["|2", 123, 456]`)
	assert.Equal(t, `[123,123,456]`, ret)
}

func TestGenRepeatObjectId(t *testing.T) {
	gen := jj.NewGen()
	gen.MockTimes = 2
	gen.RegisterFn("objectId", jj.SubstituteFn{Fn: func(args string) any { return 456 }, Demo: "对象ID, e.g. @objectId"})
	ret, _ := gen.Gen(`["|2-7", { "id": "@objectId" }]`)
	assert.Equal(t, `[{"id":456},{"id":456}]`, ret)
}

func TestGenRepeat2(t *testing.T) {
	gen := jj.NewGen()
	gen.MockTimes = 2
	gen.RegisterFn("objectId", jj.SubstituteFn{Fn: func(args string) any { return 456 }, Demo: "对象ID, e.g. @objectId"})
	gen.RegisterFn("random", jj.SubstituteFn{Fn: func(args string) any { return 1010 }, Demo: "随机数, e.g. @random(10)"})
	out, _, _ := gen.Process(`["|2-7", { "id": "@objectId",  "tags": ["|3", "@random(10)"] }]`)
	assert.Equal(t, `[{"id":456,"tags":[1010,1010]},{"id":456,"tags":[1010,1010]}]`, out.Out)
}

func TestGenObjectId(t *testing.T) {
	gen := jj.NewGen()
	gen.RegisterFn("objectId", jj.SubstituteFn{Fn: func(args string) any { return "123" }, Demo: "对象ID, e.g. @objectId"})
	ret, _ := gen.Gen(` {"id": "@objectId"} `)
	assert.Equal(t, `{"id":"123"}`, ret)
}

func TestGenExample(t *testing.T) {
	fmt.Println(jj.Gen(`["|1-3", { "id": "@objectId",  "tags": ["|1-2", "@random(5-10)"] }]`))
	// [{"id":"60bcba88ac8b71e848c7d0a7","tags":["qxr_yv"]},{"id":"60bcba88ac8b71e848c7d0a8","tags":["v4G9Xnd","xCsWH4"]}]
	fmt.Println(jj.Gen(`{"id": "@objectId"}`))                                             // {"id":"60bcba88ac8b71e848c7d0a6"}
	fmt.Println(jj.Gen(`{"id": "@random(red,green,blue)"}`))                               // {"id":"red"}
	fmt.Println(jj.Gen(`{"id": "@random(1,2,3)"}`))                                        // {"id":"3"}
	fmt.Println(jj.Gen(`{"id": "@regex([abc]{10})"}`))                                     // {"id":"ccbbbaaccc"}
	fmt.Println(jj.Gen(`{"id|2-5": "1" }`))                                                // {"id":"11"}
	fmt.Println(jj.Gen(`{"id": "@random_int"}`))                                           // {"id":1991593051}
	fmt.Println(jj.Gen(`{"id": "@random_int(100-999)"}`))                                  // {"id":330}
	fmt.Println(jj.Gen(`{"id": "Hello@random_int(100-999)"}`))                             // {"id":"Hello846"}
	fmt.Println(jj.Gen(`{"ok": "@random_bool"}`))                                          // {"ok":true}
	fmt.Println(jj.Gen(`{"day": "@random_time"}`))                                         // {"day":"2021-06-06T20:07:36.15813+08:00"}
	fmt.Println(jj.Gen(`{"day": "@random_time(yyyy-MM-dd)"}`))                             // {"day":"2021-06-06"}
	fmt.Println(jj.Gen(`{"day": "@random_time(now, yyyy-MM-dd)"}`))                        // {"day":"2021-06-06"}
	fmt.Println(jj.Gen(`{"day": "@random_time(now, yyyy-MM-dd)"}`))                        // {"day":"2021-06-06"}
	fmt.Println(jj.Gen(`{"day": "@random_time(now, yyyy-MM-ddTHH:mm:ss)"}`))               // {"day":"2021-06-06T20:07:36"}
	fmt.Println(jj.Gen(`{"day": "@random_time(yyyy-MM-dd,1990-01-01,2021-06-06)"}`))       // {"day":"1996-06-04"}
	fmt.Println(jj.Gen(`{"day": "@random_time(sep=# yyyy-MM-dd#1990-01-01#2021-06-06)"}`)) // {"day":"1995-08-23"}
	fmt.Println(jj.Gen(`{"uid": "@uuid"}`))                                                // {"uid":"619f3117-3c76-4b3f-941c-7df2a109b625"}
}

func TestParseArguments(t *testing.T) {
	assert.Equal(t, map[string][]string{
		"size": {"10"},
		"std":  {""},
		"url":  {""},
		"raw":  {""},
	}, jj.ParseArguments("size=10 std url raw"))

	arg := struct {
		Size int
		Std  bool
		Url  bool
		Raw  bool
	}{}
	jj.ParseConf("size=10 std url", &arg)
	assert.Equal(t, struct {
		Size int
		Std  bool
		Url  bool
		Raw  bool
	}{
		Size: 10,
		Std:  true,
		Url:  true,
		Raw:  false,
	}, arg)

	type mArg struct {
		Map map[string]string `prefix:"result."`
	}

	var m mArg

	jj.ParseConf("result.AccessToken=a.b.c", &m)
	assert.Equal(t, mArg{
		Map: map[string]string{"AccessToken": "a.b.c"},
	}, m)
}
