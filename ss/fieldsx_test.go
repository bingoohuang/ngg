package ss

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldsX(t *testing.T) {
	assert.Nil(t, FieldsX("(a b) c", "(", ")", 0))
	assert.Equal(t, []string{"(a b)  c"}, FieldsX("(a b)  c ", "(", ")", 1))
	assert.Equal(t, []string{"(a b)", "c"}, FieldsX("(a b)  c", "(", ")", 2))
	assert.Equal(t, []string{"(a b)", "c  d e"}, FieldsX("(a b)  c  d e ", "(", ")", 2))
	assert.Equal(t, []string{"(a b)", "c"}, FieldsX("(a b) c", "(", ")", -1))
	assert.Equal(t, []string{"(a b)", "(c d)"}, FieldsX(" (a b) (c d) ", "(", ")", -1))
	assert.Equal(t, []string{"(中 华) (人 民)"}, FieldsX("(中 华) (人 民)  ", "(", ")", 1))
	assert.Equal(t, []string{"(中 华)", "(人 民)"}, FieldsX(" (中 华) (人 民)  ", "(", ")", -1))
	assert.Equal(t, []string{"(中 华)", "(人 民)  共和国"}, FieldsX(" (中 华) (人 民)  共和国", "(", ")", 2))
}

func TestFields(t *testing.T) {
	assert.Nil(t, Fields("a b c", 0), nil)
	assert.Equal(t, []string{"a b c"}, Fields(" a b c ", 1))
	assert.Equal(t, Fields(" a b c", 2), []string{"a", "b c"})
	assert.Equal(t, Fields("a   b c", 3), []string{"a", "b", "c"})
	assert.Equal(t, Fields("a b c", 4), []string{"a", "b", "c"})
	assert.Equal(t, Fields("a b c", -1), []string{"a", "b", "c"})
	assert.Equal(t, []string{"中国", "c"}, Fields("中国 c", -1))
	assert.Equal(t, []string{"中国 c"}, Fields("中国 c", 1))
	assert.Equal(t, []string{"中国", "人民  共和国"}, Fields("   中国 人民  共和国   ", 2))
	assert.Equal(t, []string{"中国", "人民共和国"}, Fields("   中国  人民共和国  ", 2))
	assert.Equal(t, []string{"中国", "人民共和国"}, Fields("  中国  人民共和国  ", 3))
}
