package stdlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomLevel(t *testing.T) {
	RegisterCustomLevel("[DEBUG]", DebugLevel)
	RegisterCustomLevel("[INFO]", InfoLevel)
	RegisterCustomLevel("[WARN]", WarnLevel)
	RegisterCustomLevel("ERROR", ErrorLevel)
	RegisterCustomLevel("[FATAL]", FatalLevel)
	RegisterCustomLevel("[PANIC]", PanicLevel)

	level, msg, ok := parseLevelFromMsg([]byte("[DEBUG] hello"))
	assert.Equal(t, DebugLevel, level)
	assert.Equal(t, []byte("hello"), msg)
	assert.True(t, ok)
}
