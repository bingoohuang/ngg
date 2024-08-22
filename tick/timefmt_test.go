package tick_test

import (
	"strings"
	"testing"
	"time"

	"github.com/bingoohuang/ngg/tick"
	"github.com/stretchr/testify/assert"
)

func TestFormat(t *testing.T) {
	s := tick.FormatTime(time.Now(), "06yyyy-MM-dd")
	assert.True(t, strings.HasPrefix(s, "06"))
}
