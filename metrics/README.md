# gometrics

metrics golang client library.

## metrics

| \#  | TYPE         | 值类型 | Meaning            | v1           | v2         | v3-v9                                         | n          | min/max 计算基准 |
| --- | ------------ | ------ | ------------------ | ------------ | ---------- | --------------------------------------------- | ---------- | ---------------- |
| 1   | RT           | -      | 平均响应时间(毫秒) | 累计响应时间 | 累计次数   | v3: 300-400ms 总次数, ..., v9: >=900ms 总次数 | 累积打点数 | 单次打点值的 v1  |
| 2   | QPS          | -      | 业务量(次数)       | 累积次数     | 0          | 0                                             | 累积打点数 | NA               |
| 3   | SUCCESS_RATE | 百分比 | 成功率             | 累计成功数   | 累计调用数 | 0                                             | 累积打点数 | NA               |
| 4   | FAIL_RATE    | 百分比 | 失败率             | 累计失败数   | 累计调用数 | 0                                             | 累积打点数 | NA               |
| 5   | HIT_RATE     | 百分比 | 命中率             | 累计命中数   | 累计调用数 | 0                                             | 累积打点数 | NA               |
| 6   | CUR          | 离散值 | 瞬时值             | 累计瞬时值   | 0          | 0                                             | 累积打点数 | NA               |

注:

1. "记录周期" 指 `METRICS_INTERVAL` 定义的周期，一个“记录周期"内，可能累积多次打点，上面表格内的 `n` 即为记录周期内的多次打点数目
2. min/max 指一个“记录周期”内的 最小/最大

## HB

心跳

| \#  | TYPE | Meaning  | v1  | v2-v9 |
| --- | ---- | -------- | --- | ----- |
| 1   | HB   | 一次心跳 | 1   | 0     |

## Client Usage

### 准备参数

1. 通过.env环境文件设置，优先级最高。在当前目录下创建.env文件，设定一些参数， eg.

    ```properties
    # 应用名称，默认使用当前pid
    APP_NAME=bingoohuangapp
    # 写入指标日志的间隔时间，默认1s
    METRICS_INTERVAL=1s
    # 写入心跳日志的间隔时间，默认20s
    HB_INTERVAL=20s
    # Metrics对象的处理容量，默认1000，来不及处理时，超额扔弃处理
    CHAN_CAP=1000
    # 在指标来不及处理时，是否自动扔弃
    AUTO_DROP = false
    # 日志存放的目录，默认/tmp/log/metrics
    LOG_PATH=/var/log/footstone/metrics
    # 日志文件最大保留天数
    MAX_BACKUPS=7
    ```

2. 通过命令行环境变量设置

   eg. `APP_NAME=demo demoproc`

3. 通过命令行指定环境文件名

   eg. `ENV_FILE=testdata/golden.env demoproc`

### RT 平均响应时间

```go
package main

import (
	"github.com/bingoohuang/ngg/metrics/metric"
	"github.com/bingoohuang/ngg/metrics/pkg/ks"
)

func YourBusinessDemo1() {
	// 这里使用defer是为了在函数结束时，计算耗时
	defer metric.RT("key1", "key2", "key3").Ks(ks.K4("a").K8("8")).Record()

	// business logic
}

func YourBusinessDemo2() {
	rt := metric.RT("key1", "key2", "key3")

	// business logic
	start := time.Now()
	// ...
	rt.RecordSince(start)
}
```

### QPS 业务量(次数)

```go
package main

func YourBusinessDemoQPS() {
	metric.QPS("key1", "key2", "key3").Record(1 /* 业务量 */)
}
```

or in simplified way:

```go
func YourBusinessDemoQPS() {
	metric.QPS1("key1", "key2", "key3")
}
```

### SUCCESS_RATE 成功率

```go
package main

func YourBusinessDemoSuccessRate() {
	sr := metric.SuccessRate("key1", "key2", "key3")
	defer sr.IncrTotal()

	// business logic
	sr.IncrSuccess()
}
```

### FAIL_RATE 失败率

```go
package main

func YourBusinessDemoFailRate() {
	fr := metric.FailRate("key1", "key2", "key3")
	defer fr.IncrTotal()

	// business logic
	fr.IncrFail()
}
```

### HIT_RATE 命中率

```go
package main

func YourBusinessDemoHitRate() {
	fr := metric.HitRate("key1", "key2", "key3")
	defer fr.IncrTotal()

	// business logic
	fr.IncrHit()
}
```

### CUR 瞬时值

```go
package main

func YourBusinessDemoCur() {
	// business logic
	metric.Cur("key1", "key2", "key3").Record(100)
	// business logic
}
```

### Demo

1. build `cd cmd/metrics/; make -f ../../../ver/Makefile`
2. run `ENV_FILE=testdata/golden.env metrics`

```bash
$ tail -f /tmp/metricslog/metrics-hb.bingoohuangapp.log
{"time":"20220210151532000","key":"bingoohuangapp.hb","hostname":"bogon","logtype":"HB","v1":1,"v2":0,"v3":0,"v4":0,"v5":0,"v6":0,"v7":0,"v8":0,"v9":0}
{"time":"20220210151850000","key":"bingoohuangapp.hb","hostname":"bogon","logtype":"HB","v1":1,"v2":0,"v3":0,"v4":0,"v5":0,"v6":0,"v7":0,"v8":0,"v9":0}
{"time":"20220210151918000","key":"bingoohuangapp.hb","hostname":"bogon","logtype":"HB","v1":1,"v2":0,"v3":0,"v4":0,"v5":0,"v6":0,"v7":0,"v8":0,"v9":0}
{"time":"20220210151918000","key":"bingoohuangapp.hb","hostname":"bogon","logtype":"HB","v1":1,"v2":0,"v3":0,"v4":0,"v5":0,"v6":0,"v7":0,"v8":0,"v9":0}
{"time":"20220210151938000","key":"bingoohuangapp.hb","hostname":"bogon","logtype":"HB","v1":1,"v2":0,"v3":0,"v4":0,"v5":0,"v6":0,"v7":0,"v8":0,"v9":0}
{"time":"20220210151958000","key":"bingoohuangapp.hb","hostname":"bogon","logtype":"HB","v1":1,"v2":0,"v3":0,"v4":0,"v5":0,"v6":0,"v7":0,"v8":0,"v9":0}
{"time":"20220210152018000","key":"bingoohuangapp.hb","hostname":"bogon","logtype":"HB","v1":1,"v2":0,"v3":0,"v4":0,"v5":0,"v6":0,"v7":0,"v8":0,"v9":0}
{"time":"20220210152038000","key":"bingoohuangapp.hb","hostname":"bogon","logtype":"HB","v1":1,"v2":0,"v3":0,"v4":0,"v5":0,"v6":0,"v7":0,"v8":0,"v9":0}
```

```bash
$ tail -f /tmp/metricslog/metrics-key.bingoohuangapp.log
{"k1":"key1","k2":"key2","k3":"key3","k4":"k4","key":"key1#key2#key3","logtype":"RT","time":"20241009082454000","hostname":"bingoodeMacBook-Pro.local","min":299.582723,"max":705.085427,"v1":1347.0127029999999,"v2":3,"v3":1,"v4":0,"v5":0,"v6":0,"v7":1,"v8":0,"v9":0,"n":3}
{"k1":"key1","k2":"key2","k3":"key3","key":"key1#key2#key3","logtype":"QPS","time":"20241009082454000","hostname":"bingoodeMacBook-Pro.local","v1":3,"v2":0,"v3":0,"v4":0,"v5":0,"v6":0,"v7":0,"v8":0,"v9":0,"n":3}
{"k1":"key1","k2":"key2","k3":"key3","key":"key1#key2#key3","logtype":"SUCCESS_RATE","time":"20241009082454000","hostname":"bingoodeMacBook-Pro.local","v1":1,"v2":3,"v3":0,"v4":0,"v5":0,"v6":0,"v7":0,"v8":0,"v9":0,"n":4}
{"k1":"key1","k2":"key2","k3":"key3","key":"key1#key2#key3","logtype":"FAIL_RATE","time":"20241009082454000","hostname":"bingoodeMacBook-Pro.local","v1":0,"v2":3,"v3":0,"v4":0,"v5":0,"v6":0,"v7":0,"v8":0,"v9":0,"n":3}
{"k1":"key1","k2":"key2","k3":"key3","key":"key1#key2#key3","logtype":"HIT_RATE","time":"20241009082454000","hostname":"bingoodeMacBook-Pro.local","v1":0,"v2":2,"v3":0,"v4":0,"v5":0,"v6":0,"v7":0,"v8":0,"v9":0,"n":2}
{"k1":"key1","k2":"key2","k3":"key3","key":"key1#key2#key3","logtype":"CUR","time":"20241009082454000","hostname":"bingoodeMacBook-Pro.local","v1":100,"v2":0,"v3":0,"v4":0,"v5":0,"v6":0,"v7":0,"v8":0,"v9":0,"n":1}
```

## benchmark

```bash
$ go test -bench=.  ./...
WARN[0000] loading env file error open .env: no such file or directory
INFO[0000] log file /tmp/log/metrics/metrics-key.44739.log created
INFO[0000] log file /tmp/log/metrics/metrics-hb.44739.log created
/Users/bingoo/GitHub/gometrics/metric
goos: darwin
goarch: amd64
pkg: github.com/bingoohuang/ngg/metrics/metric
BenchmarkRT-12                   1803442               655 ns/op
BenchmarkQPS-12                  2232487               538 ns/op
BenchmarkSuccessRate-12          2175163               552 ns/op
BenchmarkFailRate-12             2246766               516 ns/op
BenchmarkHitRate-12              2110915               597 ns/op
BenchmarkCur-12                  3023659               388 ns/op
PASS
ok      github.com/bingoohuang/ngg/metrics/metric 11.385s
```

## cloc

```bash
$ go get -u github.com/hhatto/gocloc/cmd/gocloc
$ gocloc .
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              13            279             94           1033
XML                              5              0              0            225
Markdown                         3             62              0            210
Makefile                         1             15              7             46
-------------------------------------------------------------------------------
TOTAL                           22            356            101           1514
-------------------------------------------------------------------------------
$ date
Thu Feb 10 15:41:15 CST 2022
```
