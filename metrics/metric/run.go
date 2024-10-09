package metric

import (
	"io"
	"log"
	"math"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/metrics/pkg/rotate"
	"github.com/bingoohuang/ngg/metrics/pkg/util"
)

// DefaultRunner is the default runner for metric recording.
var DefaultRunner = NewRunner(EnvOption())

func init() {
	// start the default runner at init.
	// so the application can have at least heartbeat metrics even if there is no explicit metrics api call.
	// according to lvyong's words.
	DefaultRunner.Start()
}

// Stop stops the default runner.
func Stop() {
	DefaultRunner.Stop()
}

// Runner is a runner for metric rotate writing.
type Runner struct {
	startTime time.Time

	MetricsLogfile io.Writer
	HBLogfile      io.Writer

	C    chan *Line
	stop chan bool

	cache   map[cacheKey]*Line
	option  *Option
	AppName string

	MetricsInterval time.Duration
	HBInterval      time.Duration

	autoDrop bool
}

type cacheKey struct {
	Key     string
	LogType LogType
}

func (l *Line) makeCacheKey() cacheKey {
	c := cacheKey{Key: l.Key, LogType: l.LogType}
	if l.Ks != nil {
		c.Key += "," + strings.Join(l.Ks.Keys[:], "#")
	}

	return c
}

// NewRunner creates a Runner.
func NewRunner(ofs ...OptionFn) *Runner {
	o := createOption(ofs...)

	r := &Runner{
		option:          o,
		AppName:         o.AppName,
		MetricsInterval: o.MetricsInterval,
		HBInterval:      o.HBInterval,
		C:               make(chan *Line, o.ChanCap),
		autoDrop:        o.AutoDrop,
		stop:            make(chan bool, 1),
		cache:           make(map[cacheKey]*Line),
	}

	runtime.SetFinalizer(r, func(r *Runner) { r.Stop() })

	return r
}

func createRotateFile(o *Option, prefix string) io.Writer {
	f := filepath.Join(o.LogPath, prefix+o.AppName+".log")
	lf, err := rotate.NewFile(f, o.MaxBackups)
	if err != nil {
		log.Printf("W! fail to new logMetrics file %s, error %v", f, err)
		return nil
	}

	return lf
}

// Start starts the runner.
func (r *Runner) Start() {
	o := r.option
	r.MetricsLogfile = createRotateFile(o, "metrics-key.")
	r.HBLogfile = createRotateFile(o, "metrics-hb.")

	go r.run()

	log.Printf("runner started")
}

// Stop stops the runner.
func (r *Runner) Stop() {
	select {
	case r.stop <- true:
	default:
	}
}

func (r *Runner) run() {
	r.startTime = time.Now()

	metricsTicker := time.NewTicker(r.MetricsInterval)
	defer metricsTicker.Stop()

	r.logHB()

	hbTicker := time.NewTicker(r.HBInterval)
	defer hbTicker.Stop()

	for {
		select {
		case l := <-r.C:
			r.mergeLog(l)
		case <-metricsTicker.C:
			r.logMetrics()
		case <-hbTicker.C:
			r.logHB()
		case <-r.stop:
			log.Printf("runner stopped")
			return
		}
	}
}

func (r *Runner) logMetrics() {
	if r.MetricsLogfile == nil {
		return
	}

	r.startTime = time.Now()

	for k, pv := range r.cache {
		v := *pv
		// 处理瞬间current > total的情况.
		if v.LogType.isPercent() && v.V1 > v.V2 {
			v.V1 = v.V2
		}

		if v.N == 0 {
			if v.hasExtraKeys() {
				delete(r.cache, k)
			}
			continue
		}

		r.writeLog(r.MetricsLogfile, v)

		if v.LogType.isSimple() {
			delete(r.cache, k)
		} else {
			pv.V1 -= v.V1
			pv.V2 -= v.V2
			pv.V3 -= v.V3
			pv.V4 -= v.V4
			pv.V5 -= v.V5
			pv.V6 -= v.V6
			pv.V7 -= v.V7
			pv.V8 -= v.V8
			pv.V9 -= v.V9

			// 重置 Min, Max
			pv.Min = 0
			pv.Max = 0
			pv.N = 0
		}
	}
}

func (r *Runner) writeLog(file io.Writer, v Line) {
	v.Time = time.Now().Format(TimeLayout)
	v.Hostname = util.Hostname
	v.fulfilKeys()

	if r.option.Debug {
		s, _ := v.ToLineProtocol()
		log.Printf("LineProtocol: %s", s)
	}
	if file == nil {
		return
	}

	var obj any = v
	if v.LogType != KeyRT { // 非 RT 排除 min/max 的 JSON 输出
		obj = struct {
			Line
			Min any `json:"min,omitempty"` // 每次采集区间（METRICS_INTERVAL）中 v1  最小/大值
			Max any `json:"max,omitempty"` // 只对 RT 生效
		}{
			Line: v,
		}
	}

	content := append(util.JSONCompact(obj), '\n')

	if _, err := file.Write(content); err != nil {
		log.Printf("W! fail to write log of metrics, error %+v", err)
	}
}

func (r *Runner) mergeLog(l *Line) {
	k := l.makeCacheKey()
	if c, ok := r.cache[k]; ok {
		if l.LogType.isSimple() { // 瞬值，直接更新日志
			c.V1 = l.V1
			c.V2 = l.V2
			c.V3 = l.V3
			c.V4 = l.V4
			c.V5 = l.V5
			c.V6 = l.V6
			c.V7 = l.V7
			c.V8 = l.V8
			c.V9 = l.V9
		} else {
			c.updateMinMax(l, r)
		}
	} else {
		if l.LogType == KeyRT {
			minVal, maxVal := l.Min, l.Max
			l.Min = Min(l.N, l.Min, l.V1)
			l.Max = Max(l.N, l.Max, l.V1)
			if r.option.Debug {
				log.Printf("[%s][%s] n: %d, lastMin: %g, lastMax: %g, l.V1: %g, min: %g, max: %g",
					l.LogType, l.Key, l.N, minVal, maxVal, l.V1, l.Min, l.Max)
			}
		}

		r.cache[k] = l
		l.N = 1
	}
}

func (r *Runner) logHB() {
	if r.HBLogfile == nil {
		return
	}

	r.writeLog(r.HBLogfile, Line{
		Key:     r.AppName + ".hb",
		LogType: HB,
		V1:      1,
	})
}

func (l *Line) updateMinMax(newLine *Line, r *Runner) {
	uv1, uv2 := l.V1+newLine.V1, l.V2+newLine.V2
	uv3 := l.V3 + newLine.V3
	uv4 := l.V4 + newLine.V4
	uv5 := l.V5 + newLine.V5
	uv6 := l.V6 + newLine.V6
	uv7 := l.V7 + newLine.V7
	uv8 := l.V8 + newLine.V8
	uv9 := l.V9 + newLine.V9

	var newMin, newMax float64
	if l.LogType == KeyRT {
		newMin = Min(l.N, l.Min, newLine.V1)
		newMax = Max(l.N, l.Max, newLine.V1)
		if r.option.Debug {
			log.Printf("[%s][%s] n: %d, lastMin: %g, lastMax: %g, n.V1: %g, min: %g, max: %g",
				l.LogType, l.Key, l.N, l.Min, l.Max, newLine.V1, newMin, newMax)
		}
	}

	l.update(uv1, uv2, uv3, uv4, uv5, uv6, uv7, uv8, uv9, newMin, newMax)
}

func (l *Line) update(v1, v2, v3, v4, v5, v6, v7, v8, v9, min, max float64) {
	l.V1 = v1
	l.V2 = v2
	l.V3 = v3
	l.V4 = v4
	l.V5 = v5
	l.V6 = v6
	l.V7 = v7
	l.V8 = v8
	l.V9 = v9
	l.Min = min
	l.Max = max
	l.N++
}

// Max returns the max of two number.
func Max(n int, max, v float64) float64 {
	if n <= 0 || v > max {
		return v
	}

	return max
}

// Min returns the min of two number.
func Min(n int, min, v float64) float64 {
	if n <= 0 || v < min {
		return v
	}

	return min
}

// EPSILON。 math.Nextafter(1, 2) 返回的结果是比 1 稍微大一点的浮点数，
// 它是浮点数 1 在计算机中表示的下一个可表示的值。将这个值减去 1，得到的是
// 浮点数 1 和它的下一个可表示值之间的差距，这通常被称为“机器精度”或 EPSILON。
var EPSILON = math.Nextafter(1, 2) - 1

func FloatEquals(a, b float64) bool {
	return a-b < EPSILON && b-a < EPSILON
}
