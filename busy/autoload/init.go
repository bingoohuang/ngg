package autoload

import (
	"context"
	"os"

	"github.com/bingoohuang/ngg/busy"
	"github.com/bingoohuang/ngg/dur"
	"github.com/bingoohuang/ngg/ss"
)

func init() {
	dir := os.Getenv("DOG_DIR")
	debug := os.Getenv("DOG_DEBUG") == "1"

	ctx := context.TODO()

	bi := ss.Must(dur.Getenv("DOG_BUSY_INTERVAL", busy.DefaultCheckBusyInterval))
	go busy.Watch(ctx, dir, debug, bi)
}
