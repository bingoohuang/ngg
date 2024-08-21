package tick

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"
)

func Tick(ctx context.Context, interval, jitter time.Duration, f func() error) error {
	timer := time.NewTimer(interval)
	defer timer.Stop()

	for ctx.Err() == nil {
		if jitter > 0 {
			SleepRandom(ctx, jitter)
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if err := f(); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			timer.Reset(interval)
		}
	}

	return ctx.Err()
}

// SleepRandom will sleep for a random amount of time up to max.
// If the shutdown channel is closed, it will return before it has finished
// sleeping.
func SleepRandom(ctx context.Context, max time.Duration) {
	var ns time.Duration
	maxSleep := big.NewInt(max.Nanoseconds())
	if j, err := rand.Int(rand.Reader, maxSleep); err == nil {
		ns = time.Duration(j.Int64())
	}

	select {
	case <-ctx.Done():
	case <-time.After(ns):
	}
}

func Sleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}
