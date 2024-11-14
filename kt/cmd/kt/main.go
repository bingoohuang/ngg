package main

import (
	"fmt"
	"os"

	_ "github.com/bingoohuang/ngg/daemon/autoload"
	"github.com/bingoohuang/ngg/ggt/root"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "kt",
		Short: "kafka cli tools",
		Long:  ktHelpLong,
	}
	root.CreateCmd(rootCmd, "cluster", "cluster consume for 0.10.0.0", &clusterConsumer{})
	root.CreateCmd(rootCmd, "admin", "asic cluster administration", &adminCmd{})
	root.CreateCmd(rootCmd, "perf", "produce messages for performance test", &perfProduceCmd{})
	root.CreateCmd(rootCmd, "produce", "produce messages", &produceCmd{})
	root.CreateCmd(rootCmd, "consume", "consume messages", &consumeCmd{})
	root.CreateCmd(rootCmd, "topic", "topic info", &topicCmd{})
	root.CreateCmd(rootCmd, "group", "consumer group information and modification", &groupCmd{})

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}

}

const ktHelpLong = `
用法示例:
1. 通用设置 brokers 和 topic
    1. 通过环境变量: export KT_BROKERS=192.168.1.1:9092,192.168.1.2:9092,192.168.1.3:9092 KT_TOPIC=topic.test KT_VERSION=0.10.0.0 KT_AUTH=usr:user,pwd:123123
    2. 通过命令参数: kt tail -brokers=192.168.1.1:9092,192.168.1.2:9092,192.168.1.3:9092 -topic elastic.backup -version 0.10.0.0，不方便的是，导致命令过长，每次执行，都得带上这两个参数
2. 消费最新消息: kt tail (环境变量 CLEAN_MSG=1 清除转移符号）
3. 生产消息
    1. 直接消息：echo '你要发送的消息载荷' |  kt produce -literal -topic greetings
    2. 指定 key 和 partition : echo '{"key": "id-23", "value": "消息载荷", "partition": 0}' | kt produce -topic greetings
    3. 使用 JJ 命令生成随机消息：N=3 jj -gu a=@姓名 b=@汉字 c=@性别 d=@地址 e=@手机 f=@身份证 g=@发证机关 h=@邮箱 i=@银行卡 j=@name k=@ksuid l=@objectId m='@random(男,女)' n='@random_int(20-60)' o='@random_time(yyyy-MM-dd)' p=@random_bool q='@regex([a-z]{5}@xyz[.]cn)' |  kt produce -literal -topic greetings
    4. 从文件中读取,每一行作为一个消息： cat p20w.txt | kt produce -literal -stats
4. 生产消息性能压测
    1. 随机字符串写入压测: kt perf
    2. 使用 JSON 模板生成写入压测： kt perf -json-template '{"id":"@objectId","sex":"@random(male,female)"}'
5. 其它，看帮助
    1. 子命令列表：kt help
    2. 子命令帮助，例如：kt tail -help
`
