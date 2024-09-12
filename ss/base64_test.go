package ss_test

import (
	"testing"

	"github.com/bingoohuang/ngg/ss"
	"github.com/stretchr/testify/assert"
)

func TestBase64(t *testing.T) {
	str := "Money，就像内裤，你得有，但不必逢人就证明你有。"
	s1 := ss.Base64().Encode(str)
	assert.Nil(t, s1.V2)
	s2 := ss.Base64().Encode(str, ss.Url)
	assert.Nil(t, s2.V2)
	s3 := ss.Base64().Encode(str, ss.Url, ss.Raw)
	assert.Nil(t, s3.V2)
	s4 := ss.Base64().Encode(str, ss.Raw)
	assert.Nil(t, s4.V2)

	assert.Equal(t, str, ss.Base64().Decode(s1.V1.String()).V1.String())
	assert.Equal(t, str, ss.Base64().Decode(s2.V1.String()).V1.String())
	assert.Equal(t, str, ss.Base64().Decode(s3.V1.String()).V1.String())
	assert.Equal(t, str, ss.Base64().Decode(s4.V1.String()).V1.String())
}
