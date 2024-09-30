# berf bench result

```sh
$ berf :12345/none -d1m
Berf benchmarking http://127.0.0.1:12345/none for 1m0s using 100 goroutine(s), 12 GoMaxProcs.

Summary:
  Elapsed               1m0.001s
  Count/RPS    5126793 85444.949
    200        5126793 85444.949
  ReadWrite  92.281 110.737 Mbps

Statistics    Min      Mean    StdDev      Max
  Latency    38µs    1.164ms   1.084ms  56.159ms
  RPS       64380.8  85429.57  9152.6   107763.81

Latency Percentile:
  P50      P75      P90      P95      P99     P99.9     P99.99
  937µs  1.215ms  2.279ms  3.352ms  5.485ms  10.456ms  16.584ms
```

```sh
$ berf :12345/qps -d1m
Berf benchmarking http://127.0.0.1:12345/qps for 1m0s using 100 goroutine(s), 12 GoMaxProcs.

Summary:
  Elapsed              1m0.008s
  Count/RPS   4480010 74656.555
    200       4480010 74656.555
  ReadWrite  80.629 96.158 Mbps

Statistics    Min       Mean     StdDev     Max
  Latency     43µs    1.333ms   1.536ms   95.769ms
  RPS       48337.38  74650.07  14055.77  99805.59

Latency Percentile:
  P50      P75      P90      P95      P99     P99.9     P99.99
  862µs  1.404ms  3.074ms  4.312ms  7.536ms  13.467ms  21.366ms
```

```sh
$ berf :12345/qps_succ -d1m
Berf benchmarking http://127.0.0.1:12345/qps_succ for 1m0s using 100 goroutine(s), 12 GoMaxProcs.

Summary:
  Elapsed              1m0.002s
  Count/RPS   4027681 67124.774
    200       4027681 67124.774
  ReadWrite  72.495 89.142 Mbps

Statistics    Min       Mean     StdDev     Max
  Latency     40µs    1.484ms   2.085ms    75.3ms
  RPS       44097.97  67098.95  16866.92  95573.96

Latency Percentile:
  P50      P75      P90      P95      P99     P99.9     P99.99
  727µs  1.582ms  3.693ms  5.588ms  10.65ms  17.854ms  28.199ms
```
