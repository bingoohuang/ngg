package cmdtype_test

import (
	"testing"

	. "github.com/bingoohuang/ngg/gossh/pkg/cmdtype"
	"github.com/stretchr/testify/assert"
)

func TestParseResultVar(t *testing.T) {
	assert.Equal(t, []any{"date", "@abc"}, Slice2(ParseResultVar("date => @abc ")))
	assert.Equal(t, Slice2("date", ""), Slice2(ParseResultVar("date")))
}

func Slice2(a, b any) []any {
	return []any{a, b}
}