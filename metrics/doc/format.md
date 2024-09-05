# Metrics 文件格式

## 日志文件

`/var/log/metrics/metrics-key.{{appName}}.log`

examples:

```bash
/var/log/metrics/metrics-key.test100.log
/var/log/metrics/metrics-key.test100.log.2020-02-02
/var/log/metrics/metrics-key.test100.log.2020-02-01
```

## 内容格式

### 每行JSON

```json
{"time":"20191121145846358","key":"{{k1}}#{{k2}}#{{k3}}","hostname":"bogon","logtype":"QPS","v1":1116,"v2":0,"min":-1,"max":-1}
```

### 字段含义

名称     | 描述                                                          
---      | ---
time     | 记录埋点时间格式(yyyyMMddHHmmssSSS)                               
key      | 多级埋点，使用#分割。要求：1. 最大100字符 2. 屏蔽特殊字符（, \| # \\r \\n \\t）
hostname | 主机hostname
logtype  | QPS:请求量 RT:平均响应时间 SUCCESS_RATE:成功率 FAIL_RATE:失败率 HIT_RATE:失败率 CUR:瞬时值 DEFAULT:默认
v1       | QPS/v1:当前值 RT/v1:总响应时间 SUCCESS_RATE/v1:成功次数
v2       | QPS/v2:0 RT/v2:总访问次数 SUCCESS_RATE/v2:总次数
max      | RT/max:最大响应时间
min      | RT/min:最小响应时间
