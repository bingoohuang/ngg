package tick

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

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

// Jitter 在 interval 上增加最大 jitter 随机抖动时间
func Jitter(interval, jitter time.Duration) time.Duration {
	if jitter <= 0 {
		return interval
	}

	jitterNano := big.NewInt(jitter.Nanoseconds())
	if j, err := rand.Int(rand.Reader, jitterNano); err == nil {
		return interval + time.Duration(j.Int64())
	}

	return interval
}

// ParseTime 解析时间字符串
// 格式1(绝对时间): RFC3339 "2006-01-02T15:04:05Z07:00"
// 格式2(偏移间隔): -10d 10天前的此时
func ParseTime(tm string) (t time.Time, err error) {
	if t, err := time.Parse(time.RFC3339, tm); err == nil {
		return t, nil
	}

	if tt, _, err := Parse(tm); err == nil {
		return time.Now().Add(tt), nil
	}

	return time.Time{}, fmt.Errorf("invalid time %s", tm)
}

func ParseTimeMilli(tm string) (unixMilli int64, err error) {
	if t, err := time.Parse(time.RFC3339, tm); err == nil {
		return t.UnixMilli(), nil
	}

	if tt, _, err := Parse(tm); err == nil {
		return tt.Milliseconds() + time.Now().UnixMilli(), nil
	}

	return 0, fmt.Errorf("invalid time %s", tm)
}
