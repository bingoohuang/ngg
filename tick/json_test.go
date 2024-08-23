package tick_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/bingoohuang/ngg/tick"
	"github.com/stretchr/testify/assert"
)

type Ex struct {
	S tick.Dur
	I tick.Dur
}

func TestJSON(t *testing.T) {
	var ex Ex
	in := strings.NewReader(`{"S": "15s350ms", "I": 400000}`)
	err := json.NewDecoder(in).Decode(&ex)
	assert.Nil(t, err)

	assert.Equal(t, Ex{
		S: tick.MakeDur(15*time.Second + 350*time.Millisecond),
		I: tick.MakeDur(400000),
	}, ex)

	out, err := json.Marshal(ex)
	assert.Nil(t, err)
	assert.Equal(t, `{"S":"15.35s","I":"400µs"}`, string(out))
}
