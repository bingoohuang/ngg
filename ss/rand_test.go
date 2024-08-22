package ss_test

import (
	"testing"

	"github.com/bingoohuang/ngg/ss"
	"github.com/stretchr/testify/assert"
)

func TestRand(t *testing.T) {
	chineseName := ss.Rand().ChineseName()
	assert.NotZero(t, chineseName)
}
