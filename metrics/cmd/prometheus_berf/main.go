package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func main() {
	port := flag.Int("port", 0, "http port")
	flag.Parse()

	// Define a new total metric
	total := prometheus.NewCounter(prometheus.CounterOpts{Name: "qps", Help: "qps"})
	succ2 := prometheus.NewCounter(prometheus.CounterOpts{Name: "succ2", Help: "succ2"})
	total2 := prometheus.NewCounter(prometheus.CounterOpts{Name: "total2", Help: "total2"})

	// Register the metric with the prometheus registry
	prometheus.MustRegister(total, succ2, total2)

	http.HandleFunc("/none", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf8")
		w.Write([]byte(`{"State":200}`))
	})
	http.HandleFunc("/qps", func(w http.ResponseWriter, r *http.Request) {
		total.Inc()
		w.Header().Set("Content-Type", "application/json; charset=utf8")
		w.Write([]byte(`{"State":200}`))
	})
	http.HandleFunc("/qps_succ", func(w http.ResponseWriter, r *http.Request) {
		succ2.Inc()
		total2.Inc()
		w.Header().Set("Content-Type", "application/json; charset=utf8")
		w.Write([]byte(`{"State":200}`))
	})

	// Expose the metrics via an HTTP endpoint
	http.Handle("/metrics", promhttp.Handler())

	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}
