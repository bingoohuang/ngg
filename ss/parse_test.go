package ss_test

import (
	"testing"

	"github.com/bingoohuang/ngg/ss"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	i, err := ss.Parse[int]("100")
	assert.Nil(t, err)
	assert.Equal(t, 100, i)

	b, err := ss.Parse[bool]("true")
	assert.Nil(t, err)
	assert.True(t, b)

	f, err := ss.Parse[float32]("1.32")
	assert.Nil(t, err)
	assert.Equal(t, float32(1.32), f)
}
