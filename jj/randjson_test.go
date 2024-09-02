package jj_test

import (
	"fmt"
	"testing"

	"github.com/bingoohuang/ngg/jj"
)

func TestRandJSON(t *testing.T) {
	js := jj.Rand()
	fmt.Println(string(js))
}
