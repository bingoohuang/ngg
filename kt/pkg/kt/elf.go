package kt

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/bingoohuang/ngg/jj"
)

func ColorJSON(data any) []byte {
	jsonData, _ := json.Marshal(data)
	return jj.Color(jsonData, nil, nil)
}

func ParseRequiredAcks(acks string) sarama.RequiredAcks {
	acks = strings.ToLower(acks)
	switch acks {
	case "waitforlocal", "local":
		return sarama.WaitForLocal
	case "noresponse", "none":
		return sarama.NoResponse
	case "waitforall", "all":
		return sarama.WaitForAll
	default:
		return sarama.WaitForLocal
	}
}

func CreateCancelContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	// `signal.Notify` registers the given channel to
	// receive notifications of the specified signals.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		log.Println(<-sigs)
		cancel()
	}()
	return ctx, cancel
}

type ThrottleFn func() bool

func CreateThrottle(ctx context.Context, qps float32) (ThrottleFn, *time.Ticker) {
	if qps <= 0 {
		return func() bool { return ctx.Err() == nil }, nil
	}

	t := time.NewTicker(time.Duration(1e6/(qps)) * time.Microsecond)
	return func() bool {
		select {
		case <-t.C:
			return true
		case <-ctx.Done():
			return false
		}
	}, t
}

func Getenv(keys ...string) string {
	for _, key := range keys {
		if env := os.Getenv(key); env != "" {
			return env
		}
	}

	return ""
}
