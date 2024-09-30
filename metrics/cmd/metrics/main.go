package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/bingoohuang/ngg/metrics/metric"
	"github.com/bingoohuang/ngg/metrics/pkg/ks"
)

func main() {
	port := flag.Int("port", 0, "http port")

	flag.Parse()
	if *port > 0 {
		http.HandleFunc("/none", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf8")
			w.Write([]byte(`{"State":200}`))
		})
		http.HandleFunc("/qps", func(w http.ResponseWriter, r *http.Request) {
			metric.QPS1("key1", "key2", "key3")
			w.Header().Set("Content-Type", "application/json; charset=utf8")
			w.Write([]byte(`{"State":200}`))
		})
		http.HandleFunc("/qps_succ", func(w http.ResponseWriter, r *http.Request) {
			metric.QPS1("key1", "key2", "key3")
			sr := metric.SuccessRate("key1", "key2", "key3")
			defer sr.IncrTotal()
			sr.IncrSuccess()
			w.Header().Set("Content-Type", "application/json; charset=utf8")
			w.Write([]byte(`{"State":200}`))
		})
		http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
		return
	}

	f := func() {
		time.Sleep(100 + time.Duration(rand.Int31n(900))*time.Millisecond)
	}

	for i := 0; ; i++ {
		f()

		m := i % 6
		switch m {
		case 0:
			func() {
				defer metric.RT("key1", "key2", "key3").Ks(ks.K4("k4")).Record()
				f()
			}()
		case 1:
			func() {
				metric.QPS("key1", "key2", "key3").Record(1)
			}()
		case 2:
			func() {
				sr := metric.SuccessRate("key1", "key2", "key3")
				defer sr.IncrTotal()

				if rand.Intn(3) == 0 {
					sr.IncrSuccess()
				}
			}()
		case 3:
			func() {
				fr := metric.FailRate("key1", "key2", "key3")
				defer fr.IncrTotal()

				if rand.Intn(10) == 0 {
					fr.IncrFail()
				}
			}()
		case 4:
			func() {
				fr := metric.HitRate("key1", "key2", "key3")
				defer fr.IncrTotal()

				if rand.Intn(5) == 0 {
					fr.IncrHit()
				}
			}()
		case 5:
			func() {
				metric.Cur("key1", "key2", "key3").Record(100)
			}()
		}
	}
}
