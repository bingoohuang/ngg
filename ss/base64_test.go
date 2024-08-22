package ss_test

import (
	"testing"

	"github.com/bingoohuang/ngg/ss"
	"github.com/stretchr/testify/assert"
)

func TestBase64(t *testing.T) {
	str := "Money，就像内裤，你得有，但不必逢人就证明你有。"
	s1, err := ss.Base64().Encode(str)
	assert.Nil(t, err)
	s2, err := ss.Base64().Encode(str, ss.Url)
	assert.Nil(t, err)
	s3, err := ss.Base64().Encode(str, ss.Url, ss.Raw)
	assert.Nil(t, err)
	s4, err := ss.Base64().Encode(str, ss.Raw)
	assert.Nil(t, err)

	assert.Equal(t, str, ss.Pick1(ss.Base64().Decode(s1.String())).String())
	assert.Equal(t, str, ss.Pick1(ss.Base64().Decode(s2.String())).String())
	assert.Equal(t, str, ss.Pick1(ss.Base64().Decode(s3.String())).String())
	assert.Equal(t, str, ss.Pick1(ss.Base64().Decode(s4.String())).String())
}
