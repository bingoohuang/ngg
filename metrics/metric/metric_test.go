package metric_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/bingoohuang/ngg/metrics/metric"
	"github.com/bingoohuang/ngg/metrics/pkg/ks"
)

func TestRT(t *testing.T) {
	rt1 := metric.RT("key1")
	rt2 := metric.RT("key1", "key2")
	rt3 := metric.RT("key1", "key2", "key3")

	f := func() {
		time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)
	}

	c := make(chan bool)

	go func() {
		f()
		rt1.Record()
		c <- true
	}()

	go func() {
		f()
		rt2.Record()
		c <- true
	}()

	go func() {
		f()
		rt3.Record()
		c <- true
	}()

	<-c
	<-c
	<-c
}

func TestRT2(t *testing.T) {
	metric.RT("key1", "key2", "key3").Ks(ks.K4("a").K8("8")).Record()
	//select {}
}

func BenchmarkRT(b *testing.B) {
	for i := 0; i < b.N; i++ {
		metric.RT("key1", "key2", "key3").Ks(ks.K4("a").K8("8")).Record()
	}
}

func TestQPS(t *testing.T) {
	metric.QPS("key1").Ks(ks.K4("a").K8("8")).Record(1)
	metric.QPS("key1", "key2").Record(1)
	metric.QPS("key1", "key2", "key3").Record(1)
}

func TestQPS1(t *testing.T) {
	metric.QPS1("key1")
	metric.QPS1("key1", "key2")
	metric.QPS1("key1", "key2", "key3")
}

func BenchmarkQPS(b *testing.B) {
	for i := 0; i < b.N; i++ {
		metric.QPS("key1", "key2", "key3").Record(1)
	}
}

func BenchmarkQPS1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		metric.QPS1("key1", "key2", "key3")
	}
}

func TestSuccessRate1(t *testing.T) {
	//cpuProfileFile, _ := os.Create("cpu.profile")
	//if err := pprof.StartCPUProfile(cpuProfileFile); err != nil {
	//	log.Fatalf("pprof.StartCPUProfile failed: %v", err)
	//}

	start := time.Now()
	for i := 0; i < 10000; i++ {
		metric.QPS1("key1", "key2", "key3")
		//
		//sr := metric.SuccessRate("key1", "key2", "key3")
		//sr.IncrSuccess()
		//sr.IncrTotal()
	}

	//pprof.StopCPUProfile()
	t.Logf("cost %s", time.Since(start))

	//time.Sleep(1 * time.Minute)
}

func TestSuccessRate(t *testing.T) {
	sr := metric.SuccessRate("key1", "key2", "key3")
	sr.Ks(ks.K4("a").K8("8")).IncrSuccess()
	sr.IncrTotal()
}

func BenchmarkSuccessRate(b *testing.B) {
	sr := metric.SuccessRate("key1", "key2", "key3")

	for i := 0; i < b.N; i++ {
		sr.IncrSuccess()
		sr.IncrTotal()
	}
}

func TestFailRate(t *testing.T) {
	fr := metric.FailRate("key1", "key2", "key3")
	fr.IncrFail()
	fr.IncrTotal()
}

func BenchmarkFailRate(b *testing.B) {
	fr := metric.FailRate("key1", "key2", "key3")

	for i := 0; i < b.N; i++ {
		fr.IncrFail()
		fr.IncrTotal()
	}
}

func TestHitRate(t *testing.T) {
	fr := metric.HitRate("key1", "key2", "key3")
	fr.IncrHit()
	fr.IncrTotal()
}

func BenchmarkHitRate(b *testing.B) {
	fr := metric.HitRate("key1", "key2", "key3")

	for i := 0; i < b.N; i++ {
		fr.IncrHit()
		fr.IncrTotal()
	}
}

func TestCur(t *testing.T) {
	c1 := metric.Cur("key1")
	c2 := metric.Cur("key1", "key2")
	c3 := metric.Cur("key1", "key2", "key3")

	c1.Record(1)
	c2.Record(2)
	c3.Record(3)

	c1.Record(4)
	c2.Record(5)
	c3.Record(6)
}

func BenchmarkCur(b *testing.B) {
	c := metric.Cur("key1", "key2", "key3")

	for i := 0; i < b.N; i++ {
		c.Record(float64(rand.Int63n(10)))
	}
}

func TestFloat64(t *testing.T) {
	a := 0.15 + float64(0.15)
	b := 0.1 + float64(0.2)
	fmt.Println(a == b)
	fmt.Println(metric.FloatEquals(a, b))

	a = float64(0.15) - float64(0.15)
	b = float64(0.1) - float64(0.1)
	fmt.Println(a == b)
	fmt.Println(a == float64(0.))

	fmt.Println(metric.FloatEquals(a, b))
}
