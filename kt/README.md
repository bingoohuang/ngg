# kt - a Kafka tool that likes JSON

## kt 使用简介

1. 通用设置 brokers 和 topic
    1. 环境变量 `export BROKERS=192.168.1.1:9092,192.168.1.2:9092,192.168.1.3:9092 AUTH=user:123456 TOPIC=elastic.backup`
    2. 命令参数 `kt consume -b 192.168.1.1:9092,192.168.1.2:9092,192.168.1.3:9092 -t elastic.backup`
       ，不方便的是，导致命令过长，每次执行，都得带上这两个参数
2. 消费最新消息 `kt consume`
3. 生产消息
    1. 直接消息：`echo '你要发送的消息载荷' | kt produce`
    2. 指定 key 和
       partition ，`echo '{"key": "id-23", "value": "消息载荷", "partition": 0}' | kt produce --json -t greetings`
    3. 使用 JJ
       命令生成随机消息：`JJ_N=3 jj -gu a=@姓名 b=@汉字 c=@性别 d=@地址 e=@手机 f=@身份证 g=@发证机关 h=@邮箱 i=@银行卡 j=@name k=@ksuid l=@objectId m='@random(男,女)' n='@random_int(20-60)' o='@random_time(yyyy-MM-dd)' p=@random_bool q='@regex([a-z]{5}@xyz[.]cn)' | kt produce`
    4. 从文件中读取,每一行作为一个消息： `cat p20w.txt | kt produce`
4. 生产消息性能压测
    1. 随机字符串写入压测 `kt perf`
    2. 使用 JSON 模板生成写入压测： `kt perf --json_template '{"id":"@objectId","sex":"@random(male,female)"}'`
5. 其它，看帮助
    1. 子命令列表：`kt -h`
    2. 子命令帮助：`kt consume -h`

## SASL 示例

1. 编译 [kafka-proxy](https://github.com/grepplabs/kafka-proxy)
   - 编译机: `export GOOS=linux GOARCH=arm64`
   - 编译机: `make` 生成 `build/kafka-proxy`, `make plugin.auth-user` 生成 sasl 插件，二者上传到服务器
   - 服务器: `kafka-proxy server --bootstrap-server-mapping "192.168.126.18:9092,0.0.0.0:19002" --bootstrap-server-mapping "192.168.126.18:9091,0.0.0.0:19001" --bootstrap-server-mapping "192.168.126.18:9093,0.0.0.0:19003" --auth-local-enable --auth-local-command ./auth-user --auth-local-param "--username=my-test-user" --auth-local-param "--password=my-test-password"`
   - 或者: `BOOTSTRAP_SERVER_MAPPING="192.168.126.18:9092,0.0.0.0:19002 192.168.126.18:9091,0.0.0.0:19001 192.168.126.18:9093,0.0.0.0:19003" kafka-proxy server --auth-local-enable --auth-local-command ./auth-user --auth-local-param "--username=my-test-user" --auth-local-param "--password=my-test-password"`
   - 客户端: `BROKERS=127.0.0.1:19001,127.0.0.1:19002,127.0.0.1:19003 VERSION=0.10.0.0 TOPIC=fluent-bit-test KT_AUTH=my-test-user:my-test-password kt topic`
   - 客户端（复杂密码 base64）: `kt -b 127.0.0.1:19001,127.0.0.1:19002,127.0.0.1:19003 -v 0.10.0.0 -t fluent-bit-test --sasl base64://bXktdGVzdC11c2VyOm15LXRlc3QtcGFzc3dvcmQ topic`

## 示例日志

```sh
# kt consume -b 192.168.1.1:9092,192.168.1.2:9092,192.168.1.3:9092  -t elastic.backup
topic: elastic.backup offset: 42840172 partition: 1 key:  timestamp: 2022-07-06 09:16:29.011 valueSize: 100B msg: {"partition":1,"offset":42840172,"value":"AHn3XiZADEPb1UG36b3Eh3yEM84csGvMgJ77A8cJyRiue5FeQQwBH9PeZILJT2MIWZlgTUllCiYFT2Xdi1n4mJsbKtdz5hoqkenj","timestamp":"2022-07-06T09:16:29.011+08:00"}
topic: elastic.backup offset: 43249889 partition: 0 key:  timestamp: 2022-07-06 09:16:29.011 valueSize: 100B msg: {"partition":0,"offset":43249889,"value":"ufLYBbGHJ6okJoziJOcTtKwNQECXdAwczyoSGSYl3prCHpKQJdGlW6p3l3d7S6pYe9clGkt0zoJ2fBnYdNPhjPPgC7JBwA1rCt2V","timestamp":"2022-07-06T09:16:29.011+08:00"}
topic: elastic.backup offset: 42835575 partition: 2 key:  timestamp: 2022-07-06 09:16:29.011 valueSize: 100B msg: {"partition":2,"offset":42835575,"value":"oubuyjAFVdCoN0aB4lJHgYnagkOg3Ivf8zT0Ui5SEotX9SsAqv4VTbQtcSvC2AKIms50VioUa7DpJJBDQOIOjCHjjmcCB4SvOMBU","timestamp":"2022-07-06T09:16:29.011+08:00"}
^C
# export BROKERS 192.168.1.1:9092,192.168.1.2:9092,192.168.1.3:9092 TOPIC=elastic.backup
# kt tail
topic: elastic.backup offset: 43249889 partition: 0 key:  timestamp: 2022-07-06 09:16:29.011 valueSize: 100B msg: {"partition":0,"offset":43249889,"value":"ufLYBbGHJ6okJoziJOcTtKwNQECXdAwczyoSGSYl3prCHpKQJdGlW6p3l3d7S6pYe9clGkt0zoJ2fBnYdNPhjPPgC7JBwA1rCt2V","timestamp":"2022-07-06T09:16:29.011+08:00"}
topic: elastic.backup offset: 42840172 partition: 1 key:  timestamp: 2022-07-06 09:16:29.011 valueSize: 100B msg: {"partition":1,"offset":42840172,"value":"AHn3XiZADEPb1UG36b3Eh3yEM84csGvMgJ77A8cJyRiue5FeQQwBH9PeZILJT2MIWZlgTUllCiYFT2Xdi1n4mJsbKtdz5hoqkenj","timestamp":"2022-07-06T09:16:29.011+08:00"}
topic: elastic.backup offset: 42835575 partition: 2 key:  timestamp: 2022-07-06 09:16:29.011 valueSize: 100B msg: {"partition":2,"offset":42835575,"value":"oubuyjAFVdCoN0aB4lJHgYnagkOg3Ivf8zT0Ui5SEotX9SsAqv4VTbQtcSvC2AKIms50VioUa7DpJJBDQOIOjCHjjmcCB4SvOMBU","timestamp":"2022-07-06T09:16:29.011+08:00"}
```

```sh
$ kt perf
50000 records sent, 98584.6 records/sec (9.40 MiB/sec ingress, 4.93 MiB/sec egress), 209.7 ms avg latency, 161.2 ms stddev, 191.0 ms 50th, 369.5 ms 75th, 429.0 ms 95th, 429.0 ms 99th, 429.0 ms 99.9th, 0 total req. in flight

$ kt perf --json_template '{"id":"@objectId","sex":"@random(male,female)"}'
50000 records sent, 608952.2 records/sec (58.07 MiB/sec ingress, 5.23 MiB/sec egress), 164.1 ms avg latency, 170.8 ms stddev, 119.0 ms 50th, 405.8 ms 75th, 420.0 ms 95th, 420.0 ms 99th, 420.0 ms 99.9th, 0 total req. in flight
```

Some reasons why you might be interested:

* Consume messages on specific partitions between specific offsets.
* Display topic information (e.g., with partition offset and leader info).
* Modify consumer group offsets (e.g., resetting or manually setting offsets per topic and per partition).
* JSON output for easy consumption with tools like [kp](https://github.com/echojc/kp)
  or [jq](https://stedolan.github.io/jq/).
* JSON input to facilitate automation via tools like [jsonify](https://github.com/fgeller/jsonify).
* Configure brokers, topic and authentication via environment variables `KT_BROKERS`, `KT_TOPIC` and `KT_AUTH`.
* Fast start up time.
* No buffering of output.
* Binary keys and payloads can be passed and presented in base64 or hex encoding.
* Support for TLS authentication.
* Basic cluster admin functions: Create & delete topics.

I'm not using kt actively myself anymore, so if you think it's lacking some feature - please let me know by creating an
issue!

## Examples

<details><summary>Read details about topics that match a regex</summary>

```sh
$ kt topic --filter news --partitions
{
  "name": "actor-news",
  "partitions": [
    {
      "id": 0,
      "oldest": 0,
      "newest": 0
    }
  ]
}
```

</details>

<details><summary>Produce messages</summary>

```sh
$ echo 'Alice wins Oscar' | kt produce -t actor-news
{
  "count": 1,
  "partition": 0,
  "startOffset": 0
}
$ echo 'Bob wins Oscar' | kt produce  -t actor-news
{
  "count": 1,
  "partition": 0,
  "startOffset": 0
}
$ for i in {6..9} ; do echo Bourne sequel $i in production. | kt produce  -t actor-news  ;done
{
  "count": 1,
  "partition": 0,
  "startOffset": 1
}
{
  "count": 1,
  "partition": 0,
  "startOffset": 2
}
{
  "count": 1,
  "partition": 0,
  "startOffset": 3
}
{
  "count": 1,
  "partition": 0,
  "startOffset": 4
}
```

</details>

<details><summary>Or pass in JSON object to control key, value and partition</summary>

```sh
$ echo '{"value": "Terminator terminated", "key": "Arni", "partition": 0}' | kt produce  -t actor-news
{
  "count": 1,
  "partition": 0,
  "startOffset": 5
}
```

</details>

<details><summary>Read messages at specific offsets on specific partitions</summary>

```sh
$ kt consume  -t actor-news --offsets 0=1:2
{
  "partition": 0,
  "offset": 1,
  "key": "",
  "value": "Bourne sequel 6 in production.",
  "timestamp": "1970-01-01T00:59:59.999+01:00"
}
{
  "partition": 0,
  "offset": 2,
  "key": "",
  "value": "Bourne sequel 7 in production.",
  "timestamp": "1970-01-01T00:59:59.999+01:00"
}
```

</details>

<details><summary>Follow a topic, starting relative to newest offset</summary>

```sh
$ kt consume  -t actor-news --offsets all=newest-1:
{
  "partition": 0,
  "offset": 4,
  "key": "",
  "value": "Bourne sequel 9 in production.",
  "timestamp": "1970-01-01T00:59:59.999+01:00"
}
{
  "partition": 0,
  "offset": 5,
  "key": "Arni",
  "value": "Terminator terminated",
  "timestamp": "1970-01-01T00:59:59.999+01:00"
}
^Creceived interrupt - shutting down
shutting down partition consumer for partition 0
```

</details>

<details><summary>View offsets for a given consumer group</summary>

```sh
$ kt group --group enews  -t actor-news --partitions 0
found 1 groups
found 1 topics
{
  "name": "enews",
  "topic": "actor-news",
  "offsets": [
    {
      "partition": 0,
      "offset": 6,
      "lag": 0
    }
  ]
}
```

</details>

<details><summary>Change consumer group offset</summary>

```sh
$ kt group --group enews  -t actor-news --partitions 0 --reset 1
found 1 groups
found 1 topics
{
  "name": "enews",
  "topic": "actor-news",
  "offsets": [
    {
      "partition": 0,
      "offset": 1,
      "lag": 5
    }
  ]
}
$ kt group --group enews  -t actor-news --partitions 0
found 1 groups
found 1 topics
{
  "name": "enews",
  "topic": "actor-news",
  "offsets": [
    {
      "partition": 0,
      "offset": 1,
      "lag": 5
    }
  ]
}
```

</details>

<details><summary>Create and delete a topic</summary>

```sh
$ kt admin  --create_topic morenews  --config $(jsonify =NumPartitions 1 =ReplicationFactor 1)
$ kt topic -filter news
{
  "name": "morenews"
}
$ kt admin  -t.delete morenews
$ kt topic -filter news
```

</details>

<details><summary>Change broker address via environment variable</summary>

```sh
$ export KT_BROKERS=brokers.kafka:9092
$ kt <command> <option>
```

</details>


## Usage:

    $ kt -h

## Authentication / Encryption

Authentication configuration is possibly via a JSON file. You indicate the mode
of authentication you need and provide additional information as required for
your mode. You pass the path to your configuration file via the `-auth` flag to
each command individually, or set it via the environment variable `KT_AUTH`.

### TLS

Required fields:

- `client-cert`: Path to your certificate
- `client-key`: Path to your certificate key
- `ca`: Path to your CA certificate

Example for an authorization configuration that is used for the system tests:

    {
        "client-cert": "testdata/test-secrets/kt-test.crt",
        "client-key": "testdata/test-secrets/kt-test.key",
        "ca": "testdata/test-secrets/snakeoil-ca-1.crt"
    }


## handy scripts

1. `kt consume  -b 192.168.18.14:9092 -t metrics -v 0.10.0.0`

## relative resources

1. [segmentio/kafka-go](https://github.com/segmentio/kafka-go) It provides both low and high level APIs for interacting
   with Kafka, mirroring concepts and implementing interfaces of the Go standard library to make it easy to use and
   integrate with existing software.
2. [Go go-queue 库实现 kafka 的发布/订阅](https://mp.weixin.qq.com/s/x1KIbn9NeLyKTISzWCPIdA), go-queue
   库是由 [go-zero](https://github.com/zeromicro/go-zero) 团队针对于消息队列的封装，目前支持
   kafka、rabbitmq、stan、beanstalkd等。
3. [Watermill Kafka Pub/Sub](https://github.com/ThreeDotsLabs/watermill-kafka) Kafka Pub/Sub for the Watermill project.
4. sarama 库的问题：阿里云官方文档不推荐使用 sarama 库，[为什么不推荐使用 Sarama Go 客户端收发消息](https://help.aliyun.com/document_detail/266782.html)，这里简单列举下原文，其中解决方案对项目实践还是有些指导意义。所有Sarama Go版本客户端存在以下已知问题：
    - 当Topic新增分区时，Sarama Go客户端无法感知并消费新增分区，需要客户端重启后，才能消费到新增分区。
    - 当Sarama Go客户端同时订阅两个以上的Topic时，有可能会导致部分分区无法正常消费消息。
    - 当Sarama Go客户端的消费位点重置策略设置为Oldest(earliest)时，如果客户端宕机或服务端版本升级，由于Sarama Go客户端自行实现OutOfRange机制，有可能会导致客户端从最小位点开始重新消费所有消息。
    - 解决方案 建议尽早将Sarama Go客户端替换为Confluent Go客户端。 Confluent Go客户端的Demo地址，请访问 [kafka-confluent-go-demo](https://github.com/AliwareMQ/aliware-kafka-demos/tree/master/kafka-confluent-go-demo)。
5. [一些关于 kafka 客户端库实践经验汇总](https://pandaychen.github.io/2022/02/08/A-KAFKA-USAGE-SUMUP-3/)
6. Modern CLI for Apache Kafka, written in Go. [birdayz/kaf](https://github.com/birdayz/kaf)
7. Open-Source Web UI for Apache Kafka Management [provectus/kafka-ui](https://github.com/provectus/kafka-ui)
8. [franz-go](https://github.com/twmb/franz-go/) 包含一个功能完整的纯 Go 库，用于与 Kafka 0.8.0 到 3.8+ 进行交互。生产、消费、交易、管理等
9. [Burrow](https://github.com/linkedin/Burrow) 是 Apache Kafka 的监控伴侣，它以服务形式提供使用者滞后检查，而无需指定阈值。它监控所有使用者提交的偏移量，并按需计算这些使用者的状态。提供 HTTP 终端节点以按需请求状态，以及提供其他 Kafka 集群信息。还有一些可配置的通知程序，可以通过电子邮件或 HTTP 调用将状态发送到其他服务。

## kafka-proxy 使用

1. 编译: [kafka-proxy](https://github.com/grepplabs/kafka-proxy), `make -f /Volumes/e2t/Github/ngg/ver/Makefile`, [我的fork版本](https://github.com/goldstd/kafka-proxy)
2. 启动: `BOOTSTRAP_SERVER_MAPPING="192.168.126.18:9092,0.0.0.0:19002 192.168.126.18:9091,0.0.0.0:19001 192.168.126.18:9093,0.0.0.0:19003" kafka-proxy server`

## examples

### 消费 __consumer_offsets 的 9 号分区
```sh
[5.046s][130][~]$ kt -b 127.0.0.1:39092 --sasl 'admin:12346!' --topic __consumer_offsets consume --offsets 9=-1
2024/12/20 16:51:05 start to consume partition 9 in [59827, 9223372036854775807] / [Newest-1,1<<63-1]
#001 {"Group":"err","Topic":"errorlog","Partition":2,"Offset":1248943,"LeaderEpoch":0,"Metadata":"","CommitTimestamp":1734684662346,"ExpireTimestamp":1734771062346}
#002 {"Group":"err","Topic":"errorlog","Partition":1,"Offset":1248943,"LeaderEpoch":0,"Metadata":"","CommitTimestamp":1734684663344,"ExpireTimestamp":1734771063344}
#003 {"Group":"err","Topic":"errorlog","Partition":0,"Offset":1248947,"LeaderEpoch":0,"Metadata":"","CommitTimestamp":1734684665347,"ExpireTimestamp":1734771065347}
```