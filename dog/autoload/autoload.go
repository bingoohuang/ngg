package autoload

import (
	"context"
	"log"
	"os"

	"github.com/bingoohuang/ngg/busy"
	"github.com/bingoohuang/ngg/dog"
	"github.com/bingoohuang/ngg/dur"
	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/unit"
	_ "github.com/joho/godotenv/autoload"
)

func init() {
	c := &dog.Config{
		Pid:                 os.Getpid(),
		Dir:                 os.Getenv("DOG_DIR"),
		Debug:               os.Getenv("DOG_DEBUG") == "1",
		RSSThreshold:        ss.Must(unit.GetEnvBytes("DOG_RSS", uint64(dog.DefaultRSSThreshold))),
		CPUPercentThreshold: ss.Must(ss.Getenv[uint64]("DOG_CPU", dog.DefaultCPUThreshold)),
		Interval:            ss.Must(dur.Getenv("DOG_INTERVAL", dog.DefaultInterval)),
		Jitter:              ss.Must(dur.Getenv("DOG_JITTER", dog.DefaultJitter)),
		Times:               ss.Must(ss.Getenv[int]("DOG_TIMES", dog.DefaultTimes)),
	}

	ctx := context.Background()
	dog := dog.New(dog.WithConfig(c))
	go func() {
		if err := dog.Watch(ctx); err != nil && c.Debug {
			log.Printf("watch error: %v", err)
		}
	}()

	bi := ss.Must(dur.Getenv("DOG_BUSY_INTERVAL", busy.DefaultCheckBusyInterval))
	go busy.Watch(ctx, c.Dir, c.Debug, bi)
}
