package autoload

import (
	"context"
	"os"

	"github.com/bingoohuang/ngg/busy"
	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/tick"
)

func init() {
	dir := os.Getenv("DOG_DIR")
	debug := os.Getenv("DOG_DEBUG") == "1"

	ctx := context.TODO()

	bi := ss.Must(tick.Getenv("DOG_BUSY_INTERVAL", busy.DefaultCheckBusyInterval))
	go busy.Watch(ctx, dir, debug, bi)
}
