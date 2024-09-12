package ss_test

import (
	"os"
	"testing"

	"github.com/bingoohuang/ngg/ss"
)

func TestClose(t *testing.T) {
	var closer *os.File = nil
	ss.Close(closer)
}
