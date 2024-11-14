package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/samber/lo"

	"github.com/IBM/sarama"
	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/kt/pkg/kt"
	"github.com/bingoohuang/ngg/ss"
	"github.com/rcrowley/go-metrics"
	"github.com/spf13/cobra"
)

// from https://github.com/IBM/sarama/blob/main/tools/kafka-producer-performance/main.go
type perfProduceCmd struct {
	kt.CommonArgs `squash:"1"`
	Timeout       time.Duration `default:"3s"`

	Sync            bool    `help:"Use a synchronous producer"`
	Partitioner     string  `default:"roundrobin" help:"The partitioning scheme to use" enum:"hash,manual,random,roundrobin"`
	Compress        string  `help:"Kafka message compress codec" default:"none" enum:"none,gzip,snappy,lz4"`
	Seq             bool    `help:"Add seq header"`
	JsonTemplate    string  `help:"Use a json template for random message"`
	Binary          bool    `help:"Use a random binary message content other than ascii message"`
	Total           int     `short:"n" default:"50000" help:"The number of messages to produce"`
	Size            int     `default:"100" help:"The approximate size (in bytes) of each message to produce"`
	Partition       int     `short:"p" default:"-1" help:"The partition of -topic to run the performance test on"`
	Qps             float32 `help:"The maximum number of messages to send per second (0 for no limit)"`
	MaxOpenRequests int     `default:"5" help:"The maximum number of unacknowledged requests the client will send on a single connection before blocking"`
	MaxMessageBytes int     `default:"1000000" help:"The max permitted size of a message"`
	RequiredAcks    string  `default:"local" help:"The required number of acks needed from the broker" enum:"all,none,local"`

	FlushFrequency   time.Duration `help:"The best-effort frequency of flushes"`
	FlushBytes       int           `help:"The best-effort number of bytes needed to trigger a flush"`
	FlushMessages    int           `help:"The best-effort number of messages needed to trigger a flush"`
	FlushMaxMessages int           `help:"The maximum number of messages the producer will send in a single request"`

	ClientID          string `default:"sarama" help:"The client ID sent with every request to the brokers"`
	ChannelBufferSize int    `default:"256" help:"The number of events to buffer in internal and external channels"`
	Concurrency       int    `default:"1" help:"The number of cocurrent routines to send the messages from (-sync only)"`

	seq int32
}

func (p *perfProduceCmd) LongHelp() string {
	return `
--json-template example:
'{"id":"@objectId","sex":"@random(male,female)","image":"@base64(file=100.png)","a":"@姓名","b":"@汉字","c":"@性别","d":"@地址","e":"@手机","f":"@身份证","g":"@发证机关","h":"@邮箱","i":"@银行卡","j":"@name","k":"@ksuid","l":"@objectId","m":"@random(男,女)","n":"@random_int(20-60)","o":"@random_time(yyyy-MM-dd)", "p":"@random_bool","q":"@regex([a-z]{5}@xyz[.]cn)"}'
`
}

func (p *perfProduceCmd) Run(*cobra.Command, []string) error {
	if err := p.CommonArgs.Validate(); err != nil {
		return err
	}

	throttle, ticker := kt.CreateThrottle(context.Background(), p.Qps)
	if ticker != nil {
		defer ticker.Stop()
	}
	c := sarama.NewConfig()
	c.Net.MaxOpenRequests = p.MaxOpenRequests
	c.Producer.MaxMessageBytes = p.MaxMessageBytes
	c.Producer.RequiredAcks = kt.ParseRequiredAcks(p.RequiredAcks)
	c.Producer.Timeout = p.Timeout
	c.Producer.Partitioner = parsePartitioner(p.Partitioner, p.Partition)
	c.Producer.Compression = parseCompression(p.Compress)
	c.Producer.Flush.Frequency = p.FlushFrequency
	c.Producer.Flush.Bytes = p.FlushBytes
	c.Producer.Flush.Messages = p.FlushMessages
	c.Producer.Flush.MaxMessages = p.FlushMaxMessages
	c.Producer.Return.Successes = true
	c.ClientID = p.ClientID
	c.ChannelBufferSize = p.ChannelBufferSize
	c.Version = p.KafkaVersion
	if err := p.SetupAuth(c); err != nil {
		return err
	}
	if err := c.Validate(); err != nil {
		return fmt.Errorf("configuration validate: %w", err)
	}

	if err := lo.Ternary(p.Sync, p.runSyncProducer, p.runAsyncProducer)(c, throttle); err != nil {
		return err
	}

	// Print final metrics.
	p.printMetrics(os.Stdout, c.MetricRegistry)
	return nil
}

func (p *perfProduceCmd) runAsyncProducer(c *sarama.Config, throttle kt.ThrottleFn) error {
	producer, err := sarama.NewAsyncProducer(p.KafkaBrokers, c)
	if err != nil {
		return fmt.Errorf("create producer: %w", err)
	}
	defer producer.Close()

	messages := p.generateMessages(p.Total)

	go func() {
		for _, message := range messages {
			throttle()
			producer.Input() <- message
		}
	}()

	for i := 0; i < p.Total; i++ {
		select {
		case <-producer.Successes():
		case err := <-producer.Errors():
			printErrorAndExit("%s", err)
		}
	}

	return nil
}

func (p *perfProduceCmd) runSyncProducer(config *sarama.Config, throttle kt.ThrottleFn) error {
	producer, err := sarama.NewSyncProducer(p.KafkaBrokers, config)
	if err != nil {
		return fmt.Errorf("create producer: %s", err)
	}
	defer producer.Close()

	messages := make([][]*sarama.ProducerMessage, p.Concurrency)
	for i := 0; i < p.Concurrency; i++ {
		if i == p.Concurrency-1 {
			messages[i] = p.generateMessages(p.Total/p.Concurrency + p.Total%p.Concurrency)
		} else {
			messages[i] = p.generateMessages(p.Total / p.Concurrency)
		}
	}

	var wg sync.WaitGroup

	for _, m := range messages {
		wg.Add(1)
		go func(messages []*sarama.ProducerMessage) {
			defer wg.Done()

			for _, message := range messages {
				throttle()
				if _, _, err = producer.SendMessage(message); err != nil {
					printErrorAndExit("Failed to send message: %s", err)
				}
			}
		}(m)
	}

	wg.Wait()
	return nil
}

func (p *perfProduceCmd) printMetrics(w io.Writer, r metrics.Registry) {
	recordSendRateMetric := r.Get("record-send-rate")
	requestLatencyMetric := r.Get("request-latency-in-ms")
	outgoingByteRateMetric := r.Get("outgoing-byte-rate")
	requestsInFlightMetric := r.Get("requests-in-flight")

	if recordSendRateMetric == nil || requestLatencyMetric == nil || outgoingByteRateMetric == nil ||
		requestsInFlightMetric == nil {
		return
	}
	recordSendRate := recordSendRateMetric.(metrics.Meter).Snapshot()
	requestLatency := requestLatencyMetric.(metrics.Histogram).Snapshot()
	requestLatencyPercentiles := requestLatency.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
	outgoingByteRate := outgoingByteRateMetric.(metrics.Meter).Snapshot()
	requestsInFlight := requestsInFlightMetric.(metrics.Counter).Count()
	fmt.Fprintf(w, "%d records sent, %.1f records/sec (%.2f MiB/sec ingress, %.2f MiB/sec egress), "+
		"%.1f ms avg latency, %.1f ms stddev, %.1f ms 50th, %.1f ms 75th, "+
		"%.1f ms 95th, %.1f ms 99th, %.1f ms 99.9th, %d total req. in flight\n",
		recordSendRate.Count(),
		recordSendRate.RateMean(),
		recordSendRate.RateMean()*float64(p.Size)/1024/1024,
		outgoingByteRate.RateMean()/1024/1024,
		requestLatency.Mean(),
		requestLatency.StdDev(),
		requestLatencyPercentiles[0],
		requestLatencyPercentiles[1],
		requestLatencyPercentiles[2],
		requestLatencyPercentiles[3],
		requestLatencyPercentiles[4],
		requestsInFlight,
	)
}

func printUsageErrorAndExit(message string) {
	fmt.Fprintln(os.Stderr, "ERROR:", message)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Available command line options:")
	flag.PrintDefaults()
	os.Exit(64)
}

func printErrorAndExit(format string, values ...any) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", fmt.Sprintf(format, values...))
	os.Exit(1)
}

func parseCompression(scheme string) sarama.CompressionCodec {
	switch scheme {
	case "none":
		return sarama.CompressionNone
	case "gzip":
		return sarama.CompressionGZIP
	case "snappy":
		return sarama.CompressionSnappy
	case "lz4":
		return sarama.CompressionLZ4
	default:
		printUsageErrorAndExit(fmt.Sprintf("Unknown --compression: %s", scheme))
	}
	panic("should not happen")
}

func parsePartitioner(partitioner string, partition int) sarama.PartitionerConstructor {
	if partition < 0 && partitioner == "manual" {
		printUsageErrorAndExit("--partition must not be -1 for -partitioning=manual")
	}
	switch partitioner {
	case "manual":
		return sarama.NewManualPartitioner
	case "hash":
		return sarama.NewHashPartitioner
	case "random":
		return sarama.NewRandomPartitioner
	case "roundrobin":
		return sarama.NewRoundRobinPartitioner
	default:
		printUsageErrorAndExit(fmt.Sprintf("Unknown --partitioner: %s", partitioner))
	}
	panic("should not happen")
}

func (p *perfProduceCmd) generateMessages(total int) []*sarama.ProducerMessage {
	messages := make([]*sarama.ProducerMessage, total)
	gen := jj.NewGen()
	for i := 0; i < total; i++ {
		pm := &sarama.ProducerMessage{Topic: p.Topic, Partition: int32(p.Partition)}
		if p.Seq {
			pm.Headers = []sarama.RecordHeader{{
				Key:   []byte("seq"),
				Value: []byte(fmt.Sprintf("%d", atomic.AddInt32(&p.seq, 1))),
			}}
		}

		switch {

		case p.JsonTemplate != "":
			randJSON, _, _ := gen.Process(p.JsonTemplate)
			pm.Value = sarama.StringEncoder(randJSON.Out)
		case p.Binary:
			payload := make([]byte, p.Size)
			if _, err := rand.Read(payload); err != nil {
				printErrorAndExit("Failed to generate message payload: %s", err)
			}
			pm.Value = sarama.ByteEncoder(payload)
		default:
			pm.Value = sarama.StringEncoder(ss.Rand().String(p.Size))
		}

		messages[i] = pm
	}

	return messages
}
