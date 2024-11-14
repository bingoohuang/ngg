package kt

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/IBM/sarama"
	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/ss"
)

type Consumer struct {
	sarama.Consumer
	sarama.OffsetManager

	MessageConsumer

	Client *Client

	Poms    map[int32]sarama.PartitionOffsetManager
	Offsets map[int32]OffsetInterval

	out chan ConsumedContext

	Topic   string
	Group   string
	Timeout time.Duration
	sync.Mutex
}

type CommonArgs struct {
	AuthConfig `squash:"1"`

	Brokers string `short:"s" env:"KT_BROKERS" help:"Kafka brokers" default:"localhost:9092" persistent:"1"`
	Topic   string `short:"t" env:"KT_TOPIC" help:"On which topic to consume/produce" persistent:"1"`
	Version string `short:"v" env:"KT_VERSION" help:"Kafka protocol version" default:"0.10.0.0" persistent:"1"`
	Verbose int    `count:"1" persistent:"1"`

	KafkaVersion sarama.KafkaVersion `kong:"-"`
	KafkaBrokers []string            `kong:"-"`
}

func (c *CommonArgs) Validate() (err error) {
	c.KafkaBrokers = ParseBrokers(c.Brokers)
	c.KafkaVersion, err = ParseKafkaVersion(c.Version)
	if err != nil {
		return err
	}

	return nil
}

type ConsumerConfig struct {
	MessageConsumer
	CommonArgs `squash:"1"`
	Group      string `help:"Consumer group to use for marking offsets. kt will mark offsets if group is supplied"`
	Offsets    string `default:"newest" help:"Specifies what messages to read by partition and offset range (defaults to newest)"`
	Timeout    time.Duration
}

func StartConsume(conf ConsumerConfig) (c *Consumer, err error) {
	c = &Consumer{
		MessageConsumer: conf.MessageConsumer,
		Topic:           conf.Topic,
		Group:           conf.Group,
		Timeout:         conf.Timeout,
	}
	c.Client, err = conf.SetupClient()
	if err != nil {
		return nil, err
	}

	if err = c.setupOffsetManager(); err != nil {
		return nil, err
	}

	if c.Consumer, err = sarama.NewConsumerFromClient(c.Client.SaramaClient); err != nil {
		return nil, err
	}

	defer LogClose("consumer", c.Consumer)

	if c.Offsets, err = ParseOffsets(conf.Offsets); err != nil {
		return nil, err
	}

	partitions, err := c.findPartitions()
	if err != nil {
		return nil, err
	}

	if len(partitions) == 0 {
		return nil, fmt.Errorf("no partitions to consume")
	}
	defer c.Close()

	c.consume(partitions)
	return c, nil
}

func (c *Consumer) setupOffsetManager() (err error) {
	if c.Group == "" {
		return nil
	}

	c.OffsetManager, err = c.Client.NewOffsetManager(c.Group)
	return err
}

func (c *Consumer) consume(partitions []int32) {
	c.out = make(chan ConsumedContext)

	go c.consumeMsg()
	c.consumePartitions(partitions)
}

func (c *Consumer) consumeMsg() {
	for {
		ctx := <-c.out
		if c.MessageConsumer != nil {
			c.MessageConsumer.Consume(ctx.Message)
		}
		close(ctx.Done)
	}
}

func (c *Consumer) consumePartitions(partitions []int32) {
	var wg sync.WaitGroup
	wg.Add(len(partitions))
	for _, p := range partitions {
		go func(p int32) {
			defer wg.Done()
			if err := c.consumePartition(p); err != nil {
				log.Printf("E! consume partition %d error %v", p, err)
			}
		}(p)
	}
	wg.Wait()
}

func (c *Consumer) consumePartition(partition int32) error {
	offset, ok := c.Offsets[partition]
	if !ok {
		offset = c.Offsets[-1]
	}

	var err error
	var start, end int64
	if start, err = c.resolveOffset(offset.Start, partition); err != nil {
		log.Printf("Failed to read Start Offset for partition %v err=%v\n", partition, err)
		return nil
	}

	if end, err = c.resolveOffset(offset.End, partition); err != nil {
		log.Printf("Failed to read End Offset for partition %v err=%v\n", partition, err)
		return nil
	}

	log.Printf("start to consume partition %d in [%d, %d] / [%s,%s]",
		partition, start, end, offset.Start, offset.End)

	var pc sarama.PartitionConsumer
	if pc, err = c.Consumer.ConsumePartition(c.Topic, partition, start); err != nil {
		log.Printf("Failed to consume partition %v err=%v\n", partition, err)
		return nil
	}

	return c.partitionLoop(pc, partition, end)
}

func (c *Consumer) resolveOffset(o Offset, partition int32) (int64, error) {
	if !o.Relative {
		return o.Start, nil
	}

	switch o.Start {
	case sarama.OffsetNewest, sarama.OffsetOldest:
		res, err := c.Client.SaramaClient.GetOffset(c.Topic, partition, o.Start)
		if err != nil {
			return 0, err
		}
		if o.Start == sarama.OffsetNewest {
			res--
		}

		return res + o.Diff, nil
	case OffsetResume:
		if c.Group == "" {
			return 0, fmt.Errorf("cannot resume without -group argument")
		}
		pom, _ := c.getPOM(partition)
		next, _ := pom.NextOffset()
		return next, nil
	}

	return o.Start + o.Diff, nil
}

func (c *Consumer) Close() {
	c.Lock()
	defer c.Unlock()

	for p, pom := range c.Poms {
		LogClose(fmt.Sprintf("partition Offset manager for partition %v", p), pom)
	}
}

func (c *Consumer) getPOM(p int32) (sarama.PartitionOffsetManager, error) {
	c.Lock()
	defer c.Unlock()

	if c.Poms == nil {
		c.Poms = map[int32]sarama.PartitionOffsetManager{}
	}

	pom, ok := c.Poms[p]
	if ok {
		return pom, nil
	}

	pom, err := c.OffsetManager.ManagePartition(c.Topic, p)
	if err != nil {
		return nil, fmt.Errorf("failed to create partition Offset manager err:%q", err)
	}
	c.Poms[p] = pom

	return pom, nil
}

func (c *Consumer) partitionLoop(pc sarama.PartitionConsumer, p int32, end int64) (err error) {
	defer LogClose(fmt.Sprintf("partition consumer %v", p), pc)

	var timer *time.Timer
	timeout := make(<-chan time.Time)

	var pom sarama.PartitionOffsetManager
	if c.Group != "" {
		if pom, err = c.getPOM(p); err != nil {
			return err
		}
	}

	for {
		if c.Timeout > 0 {
			if timer != nil {
				timer.Stop()
			}
			timer = time.NewTimer(c.Timeout)
			timeout = timer.C
		}

		select {
		case <-timeout:
			log.Printf("consuming from partition %v timed out after %s\n", p, c.Timeout)
			return
		case err = <-pc.Errors():
			log.Printf("partition %v consumer encountered err %s", p, err)
			return
		case msg, ok := <-pc.Messages():
			if !ok {
				log.Printf("unexpected closed messages chan")
				return
			}

			if pom != nil {
				pom.MarkOffset(msg.Offset+1, "")
			}

			ctx := ConsumedContext{Message: msg, Done: make(chan struct{})}
			c.out <- ctx
			<-ctx.Done

			if end > 0 && msg.Offset >= end {
				return
			}
		}
	}
}

func (c *Consumer) findPartitions() ([]int32, error) {
	all, err := c.Consumer.Partitions(c.Topic)
	if err != nil {
		return nil, fmt.Errorf("failed to read partitions for topic %v err %q", c.Topic, err)
	}

	if _, hasDefault := c.Offsets[-1]; hasDefault {
		return all, nil
	}

	var res []int32
	for _, p := range all {
		if _, ok := c.Offsets[p]; ok {
			res = append(res, p)
		}
	}

	return res, nil
}

type MessageConsumer interface {
	Consume(*sarama.ConsumerMessage)
}

type ConsumedContext struct {
	Message *sarama.ConsumerMessage
	Done    chan struct{}
}

type ConsumedMessage struct {
	Timestamp *time.Time `json:"timestamp,omitempty"`
	Key       string     `json:"key"`
	Value     string     `json:"value"`
	Offset    int64      `json:"Offset"`
	Partition int32      `json:"partition"`
}

type PrintMessageConsumer struct {
	Marshal                func(v any) ([]byte, error)
	ValEncoder, KeyEncoder BytesEncoder
	sseSender              *SSESender
	Grep                   *regexp.Regexp

	N, n int64
}

func NewPrintMessageConsumer(keyEncoder, valEncoder BytesEncoder, sseSender *SSESender, grep *regexp.Regexp, n int64) *PrintMessageConsumer {

	return &PrintMessageConsumer{
		KeyEncoder: keyEncoder,
		ValEncoder: valEncoder,
		sseSender:  sseSender,
		Grep:       grep,
		N:          n,
	}
}

func (p *PrintMessageConsumer) Consume(m *sarama.ConsumerMessage) {
	if p.Grep != nil && !p.Grep.Match(m.Value) {
		return
	}

	n := atomic.AddInt64(&p.n, 1)
	if p.N > 0 && n >= p.N {
		defer os.Exit(1)
	}

	val := m.Value
	if !RawMessageFlag && jj.ValidBytes(val) {
		val = jj.FreeInnerJSON(val)
	}

	fmt.Printf("#%03d topic: %s offset: %d partition: %d timestamp: %s valueSize: %s key: [[%s]] value: [[%s]]\n",
		n, m.Topic, m.Offset, m.Partition,
		m.Timestamp.Format("2006-01-02 15:04:05.000"),
		ss.Bytes(uint64(len(m.Value))),
		m.Key, val,
	)

	if p.sseSender != nil {
		b := sseBean{
			Topic:     m.Topic,
			Offset:    strconv.FormatInt(m.Offset, 10),
			Partition: strconv.FormatInt(int64(m.Partition), 10),
			Key:       string(m.Key),
			Timestamp: m.Timestamp.Format("2006-01-02 15:04:05.000"),
			Message:   string(val),
		}
		b.MessageSize = ss.Bytes(uint64(len(b.Message)))
		e, _ := json.Marshal(b)
		p.sseSender.Send(string(e))
	}
}

var RawMessageFlag, _ = ss.GetenvBool("RAW_MESSAGE", false)

type sseBean struct {
	Topic       string
	Offset      string
	Partition   string
	Key         string `json:",omitempty"`
	Timestamp   string
	Message     string
	MessageSize string
}
