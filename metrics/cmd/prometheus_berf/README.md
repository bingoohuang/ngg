# berf result

host info

```json
{
  "OS": "darwin",
  "HostInfo": {
    "MemAvailable": "5.893GiB/16GiB, 00.37%",
    "Platform": "darwin",
    "CpuModel": "Intel(R) Core(TM) i7-8850H CPU @ 2.60GHz",
    "UptimeHuman": "12 days",
    "PlatformVersion": "13.2.1",
    "OS": "darwin",
    "OsRelease": "",
    "KernelVersion": "22.3.0",
    "KernelArch": "x86_64",
    "Ips": [
      "192.168.227.46",
      "172.16.28.1",
      "172.16.177.1"
    ],
    "Procs": 741,
    "NumCPU": 12,
    "CpuMhz": 2600,
    "Uptime": 1043459
  }
}
```

```shell
$ berf :22345/none -d1m
Berf benchmarking http://127.0.0.1:22345/none for 1m0s using 100 goroutine(s), 12 GoMaxProcs.

Summary:
  Elapsed               1m0.006s
  Count/RPS    4816236 80261.607
    200        4816236 80261.607
  ReadWrite  86.683 104.019 Mbps

Statistics    Min       Mean     StdDev      Max
  Latency     39µs    1.239ms    1.61ms   119.862ms
  RPS       48586.56  80247.41  12306.91  100712.9

Latency Percentile:
  P50      P75      P90      P95      P99     P99.9     P99.99
  869µs  1.201ms  2.668ms  3.905ms  7.155ms  16.761ms  36.423ms
```

```sh
$ berf :22345/qps -d1m
Berf benchmarking http://127.0.0.1:22345/qps for 1m0s using 100 goroutine(s), 12 GoMaxProcs.

Summary:
  Elapsed              1m0.009s
  Count/RPS   3369357 56146.807
    200       3369357 56146.807
  ReadWrite  60.639 72.317 Mbps

Statistics    Min       Mean     StdDev     Max
  Latency     42µs    1.773ms   2.666ms   88.473ms
  RPS       38354.35  56115.33  11916.37  87292.64

Latency Percentile:
  P50      P75      P90      P95      P99      P99.9    P99.99
  964µs  1.704ms  4.152ms  6.455ms  12.743ms  27.563ms  48.98ms
```

```sh
$ berf :22345/qps_succ -d1m
Berf benchmarking http://127.0.0.1:22345/qps_succ for 1m0s using 100 goroutine(s), 12 GoMaxProcs.

Summary:
  Elapsed              1m0.004s
  Count/RPS   3556126 59264.707
    200       3556126 59264.707
  ReadWrite  64.006 78.704 Mbps

Statistics    Min       Mean    StdDev      Max
  Latency     51µs    1.679ms   2.447ms  127.038ms
  RPS       43929.77  59256.67  9677.26  81950.18

Latency Percentile:
  P50        P75      P90      P95      P99      P99.9     P99.99
  1.052ms  1.586ms  3.729ms  5.554ms  10.974ms  27.388ms  58.566ms
```