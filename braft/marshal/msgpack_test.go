package marshal_test

import (
	"testing"

	"github.com/bingoohuang/ngg/braft/marshal"
	"github.com/stretchr/testify/assert"
)

func TestMsgPack(t *testing.T) {
	type MyStruct struct {
		Name string
	}
	ms := MyStruct{Name: "bingoohuang"}

	tr := marshal.NewTypeRegister(marshal.NewMsgPacker())
	data, err := tr.Marshal(ms)
	assert.Nil(t, err)

	m1, err := tr.Unmarshal(data)
	assert.Nil(t, err)
	assert.Equal(t, ms, m1)

	_, err = tr.Marshal(nil)
	assert.Nil(t, err)
}
