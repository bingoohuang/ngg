package jj

import "testing"

func TestGenCached(t *testing.T) {
	cached, _ := GenWithCache("@身份证_1 @身份证_1")
	t.Log(cached)
}
