package berf

import (
	"encoding/json"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/axiomhq/hyperloglog"
	"github.com/beorn7/perks/histogram"
	"github.com/beorn7/perks/quantile"
	"github.com/bingoohuang/ngg/berf/pkg/util"
)

type ReportRecord struct {
	error      string
	code       []string
	counting   []string
	cost       time.Duration
	readBytes  int64
	writeBytes int64
	assertFail int64
}

func (r *ReportRecord) Reset() {
	r.cost = 0
	r.code = nil
	r.counting = nil
	r.error = ""
	r.readBytes = 0
	r.writeBytes = 0
}

var (
	startTime = time.Now()

	recordPool      = sync.Pool{New: func() interface{} { return new(ReportRecord) }}
	quantiles       = []float64{0.50, 0.75, 0.90, 0.95, 0.99, 0.999, 0.9999}
	quantilesTarget = map[float64]float64{0.50: 0.01, 0.75: 0.01, 0.90: 0.001, 0.95: 0.001, 0.99: 0.001, 0.999: 0.0001, 0.9999: 0.00001}
)

type Stats struct {
	count                int64
	sum, sumSq, min, max float64
}

func (s *Stats) Update(v float64) {
	s.count++
	s.sum += v
	s.sumSq += v * v
	if v < s.min || s.count == 1 {
		s.min = v
	}
	if v > s.max || s.count == 1 {
		s.max = v
	}
}

func (s *Stats) Stddev() float64 {
	div := float64(s.count * (s.count - 1))
	if div == 0 {
		return 0
	}
	num := (float64(s.count) * s.sumSq) - math.Pow(s.sum, 2)
	return math.Sqrt(num / div)
}

func (s *Stats) Mean() float64 {
	if s.count == 0 {
		return 0
	}
	return s.sum / float64(s.count)
}

func (s *Stats) Reset() {
	s.count = 0
	s.sum = 0
	s.sumSq = 0
	s.min = 0
	s.max = 0
}

type StreamReport struct {
	errors map[string]int64

	latencyWithinSec *Stats
	rpsStats         *Stats
	latencyQuantile  *quantile.Stream
	latencyHistogram *histogram.Histogram
	codes            map[string]int64

	latencyStats *Stats
	counts       *hyperloglog.Sketch

	requester *Requester

	doneChan chan struct{}

	readBytes  int64
	writeBytes int64

	rpsWithinSec float64
	lock         sync.Mutex

	noDateWithinSec bool
	assertFail      int64
}

func NewStreamReport(requester *Requester) *StreamReport {
	return &StreamReport{
		latencyQuantile:  quantile.NewTargeted(quantilesTarget),
		latencyHistogram: histogram.New(8),
		codes:            make(map[string]int64, 1),
		errors:           make(map[string]int64, 1),
		doneChan:         make(chan struct{}, 1),
		counts:           hyperloglog.New16(),
		latencyStats:     &Stats{},
		rpsStats:         &Stats{},
		latencyWithinSec: &Stats{},
		requester:        requester,
	}
}

func (s *StreamReport) insert(v float64) {
	s.latencyQuantile.Insert(v)
	s.latencyHistogram.Insert(v)
	s.latencyStats.Update(v)
}

func (s *StreamReport) Collect(recordChan <-chan *ReportRecord) {
	latencyWithinSecTemp := &Stats{}
	go func() {
		ticker := time.NewTicker(time.Second)
		lastCount := int64(0)
		lastTime := startTime
		for {
			select {
			case <-ticker.C:
				s.lock.Lock()
				if diff := s.latencyStats.count - lastCount; diff > 0 {
					rps := float64(diff) / time.Since(lastTime).Seconds()
					s.rpsStats.Update(rps)
					lastCount = s.latencyStats.count
					lastTime = time.Now()

					*s.latencyWithinSec = *latencyWithinSecTemp
					s.rpsWithinSec = rps
					latencyWithinSecTemp.Reset()
					s.noDateWithinSec = false
				} else {
					s.noDateWithinSec = true
				}
				s.lock.Unlock()
			case <-s.doneChan:
				return
			}
		}
	}()

	for {
		r, ok := <-recordChan
		if !ok {
			close(s.doneChan)
			return
		}
		s.lock.Lock()
		latencyWithinSecTemp.Update(float64(r.cost))
		s.insert(float64(r.cost))
		if len(r.code) > 0 {
			codes := util.MergeCodes(r.code)
			r.code = nil
			s.codes[codes]++
		}
		if r.error != "" {
			s.errors[r.error]++
		}
		for _, counting := range r.counting {
			s.counts.Insert([]byte(counting))
		}
		r.counting = nil
		s.readBytes += r.readBytes
		s.writeBytes += r.writeBytes
		s.assertFail += r.assertFail
		s.lock.Unlock()
		recordPool.Put(r)
	}
}

type SnapshotPercentile struct {
	Percentile float64
	Latency    time.Duration
}

type SnapshotRpsStats struct {
	Min, Mean, StdDev, Max float64
}

type SnapshotStats struct {
	Min, Mean, StdDev, Max time.Duration
}

type SnapshotHistogram struct {
	Mean  time.Duration
	Count int
}

type SnapshotReport struct {
	Stats            *SnapshotStats
	Codes, Errors    map[string]int64
	RpsStats         *SnapshotRpsStats
	Histograms       []*SnapshotHistogram
	Percentiles      []*SnapshotPercentile
	RPS, ElapseInSec float64
	Count, Counting  int64

	ReadBytes, WriteBytes int64
	AssertFail            int64
	Elapsed               time.Duration
}

func (s *StreamReport) Snapshot() *SnapshotReport {
	s.lock.Lock()
	defer s.lock.Unlock()

	rs := &SnapshotReport{
		Elapsed: time.Since(startTime),
		Count:   s.latencyStats.count,
		Stats: &SnapshotStats{
			Min: time.Duration(s.latencyStats.min), Mean: time.Duration(s.latencyStats.Mean()),
			StdDev: time.Duration(s.latencyStats.Stddev()), Max: time.Duration(s.latencyStats.max),
		},
	}
	if s.rpsStats.count > 0 {
		rs.RpsStats = &SnapshotRpsStats{
			Min: s.rpsStats.min, Mean: s.rpsStats.Mean(),
			StdDev: s.rpsStats.Stddev(), Max: s.rpsStats.max,
		}
	}

	elapseInSec := rs.Elapsed.Seconds()
	rs.RPS = float64(rs.Count) / elapseInSec
	rs.ReadBytes = s.readBytes
	rs.WriteBytes = s.writeBytes
	rs.AssertFail = s.assertFail
	rs.Counting = int64(s.counts.Estimate())
	rs.ElapseInSec = elapseInSec

	rs.Codes = make(map[string]int64, len(s.codes))
	for k, v := range s.codes {
		rs.Codes[k] = v
	}
	rs.Errors = make(map[string]int64, len(s.errors))
	for k, v := range s.errors {
		rs.Errors[k] = v
	}

	rs.Percentiles = make([]*SnapshotPercentile, len(quantiles))
	for i, p := range quantiles {
		rs.Percentiles[i] = &SnapshotPercentile{Percentile: p, Latency: time.Duration(s.latencyQuantile.Query(p))}
	}

	hisBins := s.latencyHistogram.Bins()
	rs.Histograms = make([]*SnapshotHistogram, len(hisBins))
	for i, b := range hisBins {
		rs.Histograms[i] = &SnapshotHistogram{Mean: time.Duration(b.Mean()), Count: b.Count}
	}

	return rs
}

func (s *StreamReport) Done() <-chan struct{} { return s.doneChan }

type ChartsReport struct {
	Latency            []util.Float64
	LatencyPercentiles []util.Float64
	RPS                util.Float64
	Concurrent         int64
}

func (s *StreamReport) Charts() *ChartsReport {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.noDateWithinSec {
		return nil
	}

	percentiles := make([]util.Float64, len(quantiles))
	for i, p := range quantiles {
		percentiles[i] = util.Float64(s.latencyQuantile.Query(p) / 1e6)
	}

	l := s.latencyWithinSec
	return &ChartsReport{
		RPS:                util.Float64(s.rpsWithinSec),
		Latency:            []util.Float64{util.Float64(l.min / 1e6), util.Float64(l.Mean() / 1e6), util.Float64(l.Stddev() / 1e6), util.Float64(l.max / 1e6)},
		LatencyPercentiles: percentiles,
		Concurrent:         atomic.LoadInt64(&s.requester.concurrent),
	}
}

func createMetrics(rd *ChartsReport, noop bool) []byte {
	m := map[string]interface{}{}
	if rd != nil && !noop {
		m["latencyPercentile"] = rd.LatencyPercentiles
		m["latency"] = rd.Latency
		m["concurrent"] = []interface{}{rd.Concurrent}
		m["tps"] = []interface{}{rd.RPS, 0}
	}

	md := Metrics{Time: time.Now().Format("2006-01-02 15:04:05"), Values: m}
	data, _ := json.Marshal(md)
	return data
}

type Metrics struct {
	Values map[string]interface{} `json:"values"`
	Time   string                 `json:"time"`
}
