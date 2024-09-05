package metric

import (
	"time"

	"github.com/bingoohuang/ngg/metrics/pkg/ks"
)

// Recorder record rate.
type Recorder struct {
	Runner  *Runner
	LogType LogType
	Key
}

// MakeRecorder creates a Recorder.
func MakeRecorder(logType LogType, keys []string) Recorder {
	return DefaultRunner.MakeRecorder(logType, keys)
}

// MakeRecorder creates a Recorder.
func (r *Runner) MakeRecorder(logType LogType, keys []string) Recorder {
	return Recorder{Runner: r, LogType: logType, Key: NewKey(keys)}
}

// PutRecord put a metric record to channel.
func (c Recorder) PutRecord(v1, v2 float64, vx ...float64) {
	if c.Checked {
		c.Runner.AsyncPut(c.Key, c.LogType, v1, v2, vx...)
	}
}

// RTRecorder is a Round-Time recorder 平均响应时间.
type RTRecorder struct {
	Start time.Time
	Recorder
}

// RT makes an RT Recorder.
func RT(keys ...string) RTRecorder { return DefaultRunner.RT(keys...) }

// RT makes an RT Recorder.
func (r *Runner) RT(keys ...string) RTRecorder {
	return RTRecorder{Recorder: r.MakeRecorder(KeyRT, keys), Start: time.Now()}
}

// Ks add extra keys.
func (r RTRecorder) Ks(k *ks.Ks) RTRecorder {
	r.ks = k
	return r
}

// Record records a round-time.
func (r RTRecorder) Record() { r.RecordSince(r.Start) }

// RecordSince records a round-time since start.
func (r RTRecorder) RecordSince(start time.Time) {
	v1 := float64(time.Since(start)) / 1e6
	vx := make([]float64, 9-2)
	switch {
	case v1 >= 900:
		vx[9-3] = 1
	case v1 >= 800:
		vx[8-3] = 1
	case v1 >= 700:
		vx[7-3] = 1
	case v1 >= 600:
		vx[6-3] = 1
	case v1 >= 500:
		vx[5-3] = 1
	case v1 >= 400:
		vx[4-3] = 1
	case v1 >= 300:
		vx[3-3] = 1
	}
	r.PutRecord(v1, 1, vx...)
}

// QPSRecorder is a QPS recorder.
type QPSRecorder struct{ Recorder }

// QPS makes a QPS Recorder.
func QPS(keys ...string) QPSRecorder { return DefaultRunner.QPS(keys...) }

// QPS makes a QPS Recorder.
func (r *Runner) QPS(keys ...string) QPSRecorder { return QPSRecorder{r.MakeRecorder(KeyQPS, keys)} }

// Ks add extra keys.
func (q QPSRecorder) Ks(k *ks.Ks) QPSRecorder {
	q.ks = k
	return q
}

// Record records a request.
func (q QPSRecorder) Record(times float64) {
	if times > 0 {
		q.PutRecord(times, 0)
	}
}

// QPS1 makes a QPS Recorder and then record a request with times 1.
func QPS1(keys ...string) { QPS(keys...).Record(1) }

// QPS1 makes a QPS Recorder and then record a request with times 1.
func (r *Runner) QPS1(keys ...string) { r.QPS(keys...).Record(1) }

// SuccessRateRecorder record success rate.
type SuccessRateRecorder struct{ Recorder }

// SuccessRate makes a SuccessRateRecorder.
func SuccessRate(keys ...string) SuccessRateRecorder { return DefaultRunner.SuccessRate(keys...) }

// SuccessRate makes a SuccessRateRecorder.
func (r *Runner) SuccessRate(keys ...string) SuccessRateRecorder {
	return SuccessRateRecorder{r.MakeRecorder(KeySuccessRate, keys)}
}

// Ks add extra keys.
func (c SuccessRateRecorder) Ks(k *ks.Ks) SuccessRateRecorder {
	c.ks = k
	return c
}

// IncrSuccess increment success count.
func (c SuccessRateRecorder) IncrSuccess() { c.PutRecord(1, 0) }

// IncrTotal increment total.
func (c SuccessRateRecorder) IncrTotal() { c.PutRecord(0, 1) }

// FailRateRecorder record success rate.
type FailRateRecorder struct{ Recorder }

// FailRate creates a FailRateRecorder.
func FailRate(keys ...string) FailRateRecorder { return DefaultRunner.FailRate(keys...) }

// FailRate creates a FailRateRecorder.
func (r *Runner) FailRate(keys ...string) FailRateRecorder {
	return FailRateRecorder{r.MakeRecorder(KeyFailRate, keys)}
}

// Ks add extra keys.
func (c FailRateRecorder) Ks(k *ks.Ks) FailRateRecorder {
	c.ks = k
	return c
}

// IncrFail increment success count.
func (c FailRateRecorder) IncrFail() { c.PutRecord(1, 0) }

// IncrTotal increment total.
func (c FailRateRecorder) IncrTotal() { c.PutRecord(0, 1) }

// HitRateRecorder record hit rate.
type HitRateRecorder struct{ Recorder }

// HitRate makes a HitRateRecorder.
func HitRate(keys ...string) HitRateRecorder { return DefaultRunner.HitRate(keys...) }

// HitRate makes a HitRateRecorder.
func (r *Runner) HitRate(keys ...string) HitRateRecorder {
	return HitRateRecorder{r.MakeRecorder(KeyHitRate, keys)}
}

// Ks add extra keys.
func (c HitRateRecorder) Ks(k *ks.Ks) HitRateRecorder {
	c.ks = k
	return c
}

// IncrHit increment success count.
func (c HitRateRecorder) IncrHit() { c.PutRecord(1, 0) }

// IncrTotal increment total.
func (c HitRateRecorder) IncrTotal() { c.PutRecord(0, 1) }

// CurRecorder record 瞬时值(Gauge).
type CurRecorder struct{ Recorder }

// Cur makes a Cur Recorder.
func Cur(keys ...string) CurRecorder { return DefaultRunner.Cur(keys...) }

// Cur makes a Cur Recorder.
func (r *Runner) Cur(keys ...string) CurRecorder {
	return CurRecorder{r.MakeRecorder(KeyCUR, keys)}
}

// Ks add extra keys.
func (c CurRecorder) Ks(k *ks.Ks) CurRecorder {
	c.ks = k
	return c
}

// Record records v1.
func (c CurRecorder) Record(v1 float64) { c.PutRecord(v1, 0) }
