package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/bingoohuang/ngg/kt/pkg/kt"
	"github.com/spf13/cobra"
)

type produceCmd struct {
	kt.CommonArgs `squash:"1"`
	Timeout       time.Duration `default:"3s"`

	KeyDecoder   string `default:"string" enum:"hex,base64,string"`
	ValueDecoder string `default:"string" enum:"hex,base64,string"`
	Partition    int32  `help:"Partition to produce to (defaults to 0)"`
	Batch        int    `default:"1" help:"Max size of a batch before sending it off"`
	Json         bool   `help:"Interpret stdin line as JSON"`
	Stats        bool   `help:"Print only final stats info"`
	Compress     string `help:"Kafka message compress codec" enum:"gzip,snappy,lz4"`
	Partitioner  string `help:"Optional partitioner" default:"hash" enum:"hash,rand"`
	BufSize      int    `default:"16777216" help:"Buffer size for scanning stdin, default 16M"`

	keyDecoder, valDecoder kt.StringDecoder
	out                    chan kt.PrintContext

	leaders  map[int32]*sarama.Broker
	compress sarama.CompressionCodec
}

func (c *produceCmd) Run(*cobra.Command, []string) (err error) {
	if err = c.CommonArgs.Validate(); err != nil {
		return err
	}

	c.keyDecoder, err = kt.ParseStringDecoder(c.KeyDecoder)
	if err != nil {
		return err
	}
	c.valDecoder, err = kt.ParseStringDecoder(c.ValueDecoder)
	if err != nil {
		return err
	}

	c.compress = kafkaCompression(c.Compress)

	defer c.close()
	c.findLeaders()
	stdin := make(chan string)
	lines := make(chan string)
	messages := make(chan Message)
	batchedMessages := make(chan []Message)
	c.out = make(chan kt.PrintContext)
	q := make(chan struct{})

	go readStdinLines(c.BufSize, stdin)
	go listenForInterrupt(q)
	go c.readInput(q, stdin, lines)
	go c.deserializeLines(lines, messages, int32(len(c.leaders)))
	go c.batchRecords(messages, batchedMessages)
	go c.produce(batchedMessages)
	kt.PrintOutStats(c.out, c.Stats)

	return nil
}

type Message struct {
	Key       *string `json:"key,omitempty"`
	Value     *string `json:"value,omitempty"`
	K         *string `json:"k,omitempty"`
	V         *string `json:"v,omitempty"`
	Partition *int32  `json:"partition"`
	P         *int32  `json:"p"`
}

func kafkaCompression(codecName string) sarama.CompressionCodec {
	switch codecName {
	case "gzip":
		return sarama.CompressionGZIP
	case "snappy":
		return sarama.CompressionSnappy
	case "lz4":
		return sarama.CompressionLZ4
	case "none", "":
		return sarama.CompressionNone
	}

	failf("unsupported compress codec %#v - supported: gzip, snappy, lz4", codecName)
	panic("unreachable")
}

func (c *produceCmd) findLeaders() {
	var (
		err error
		res *sarama.MetadataResponse
	)

	req := sarama.MetadataRequest{Topics: []string{c.Topic}}
	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Version = c.KafkaVersion
	cfg.ClientID = "kt-produce-" + kt.CurrentUserName()
	if c.Verbose > 0 {
		log.Printf("sarama client configuration %#v\n", cfg)
	}

	if err := c.Validate(); err != nil {
		failf("configuration validate: %v", err)
	}

loop:
	for _, addr := range c.KafkaBrokers {
		broker := sarama.NewBroker(addr)
		if err = broker.Open(cfg); err != nil {
			log.Printf("Failed to open broker connection to %v. err=%s\n", addr, err)
			continue loop
		}
		if connected, err := broker.Connected(); !connected || err != nil {
			log.Printf("Failed to open broker connection to %v. err=%s\n", addr, err)
			continue loop
		}

		if res, err = broker.GetMetadata(&req); err != nil {
			log.Printf("Failed to get metadata from %#v. err=%v\n", addr, err)
			continue loop
		}

		brokers := map[int32]*sarama.Broker{}
		for _, b := range res.Brokers {
			brokers[b.ID()] = b
		}

		for _, tm := range res.Topics {
			if tm.Name == c.Topic {
				if !errors.Is(tm.Err, sarama.ErrNoError) {
					log.Printf("Failed to get metadata from %#v. err=%v\n", addr, tm.Err)
					continue loop
				}

				c.leaders = map[int32]*sarama.Broker{}
				for _, pm := range tm.Partitions {
					b, ok := brokers[pm.Leader]
					if !ok {
						failf("failed to find leader in broker response, giving up")
					}

					if err = b.Open(cfg); err != nil && errors.Is(err, sarama.ErrAlreadyConnected) {
						log.Printf("W! failed to open broker connection err=%s", err)
					}
					if connected, err := broker.Connected(); !connected && err != nil {
						failf("failed to wait for broker connection to open err=%s", err)
					}

					c.leaders[pm.ID] = b
				}
				return
			}
		}
	}

	failf("failed to find leader for given topic")
}

func (c *produceCmd) close() {
	for _, b := range c.leaders {
		var (
			connected bool
			err       error
		)

		if connected, err = b.Connected(); err != nil {
			log.Printf("Failed to check if broker is connected. err=%s\n", err)
			continue
		}

		if !connected {
			continue
		}

		if err = b.Close(); err != nil {
			log.Printf("Failed to close broker %v connection. err=%s\n", b, err)
		}
	}
}

func (c *produceCmd) deserializeLines(in chan string, out chan Message, partitionCount int32) {
	defer func() { close(out) }()
	for l := range in {
		msg := c.parseLine(l, partitionCount)
		out <- msg
	}
}

func (c *produceCmd) parseLine(l string, partitionCount int32) Message {
	v := &l
	if len(l) == 0 {
		v = nil
	}

	msg := Message{Partition: &c.Partition}
	if !c.Json {
		msg.Value = v
	} else if err := json.Unmarshal([]byte(l), &msg); err != nil {
		if c.Verbose > 0 {
			log.Printf("Failed to unmarshal input [%v], falling back to defaults. err=%v\n", l, err)
		}

		msg.Value = v
	}

	c.setPartition(&msg, partitionCount)
	return msg
}

func (c *produceCmd) setPartition(msg *Message, partitionCount int32) {
	if msg.Partition == nil && msg.P != nil {
		msg.Partition = msg.P
	}

	if msg.Partition != nil {
		return
	}

	part := int32(0)
	switch {
	case c.Partitioner == "rand":
		part = randPartition(partitionCount)
	case msg.Key != nil && c.Partitioner == "hash":
		part = hashCodePartition(*msg.Key, partitionCount)
	}

	msg.Partition = &part
}

func (c *produceCmd) batchRecords(in chan Message, out chan []Message) {
	defer func() { close(out) }()

	var messages []Message
	send := func() {
		out <- messages
		messages = messages[:0]
	}

	for {
		select {
		case m, ok := <-in:
			if !ok {
				send()
				return
			}

			messages = append(messages, m)
			if len(messages) > 0 && len(messages) >= c.Batch {
				send()
			}
		case <-time.After(c.Timeout):
			if len(messages) > 0 {
				send()
			}
		}
	}
}

type partitionProduceResult struct {
	start int64
	count int64
}

func (c *produceCmd) makeSaramaMessage(msg Message) (*sarama.Message, error) {
	var (
		err error
		sm  = &sarama.Message{Codec: c.compress}
	)

	if v := kt.FirstNotNil(msg.Key, msg.K); v != "" {
		if sm.Key, err = c.keyDecoder.Decode(v); err != nil {
			return sm, fmt.Errorf("failed to decode key as string, err=%v", err)
		}
	}

	if v := kt.FirstNotNil(msg.Value, msg.V); v != "" {
		if data, ok := readFile(v); ok {
			sm.Value = data
		} else if sm.Value, err = c.valDecoder.Decode(v); err != nil {
			return sm, fmt.Errorf("failed to decode value as string, err=%v", err)
		}
	}

	if c.KafkaVersion.IsAtLeast(sarama.V0_10_0_0) {
		sm.Version = 1
		sm.Timestamp = time.Now()
	}

	return sm, nil
}

func readFile(v string) ([]byte, bool) {
	if !strings.HasPrefix(v, "@") {
		return nil, false
	}

	data, err := os.ReadFile(v[1:])
	if err != nil {
		return nil, false
	}

	return data, true
}

func (c *produceCmd) produceBatch(leaders map[int32]*sarama.Broker, batch []Message) error {
	requests := map[*sarama.Broker]*sarama.ProduceRequest{}
	valueSize := 0
	for _, msg := range batch {
		broker, ok := leaders[*msg.Partition]
		if !ok {
			return fmt.Errorf("non-configured partition %v", *msg.Partition)
		}
		req, ok := requests[broker]
		if !ok {
			req = &sarama.ProduceRequest{RequiredAcks: sarama.WaitForAll, Timeout: 10000}
			requests[broker] = req
		}

		sm, err := c.makeSaramaMessage(msg)
		if err != nil {
			return err
		}
		valueSize += len(sm.Value)
		req.AddMessage(c.Topic, *msg.Partition, sm)
	}

	for broker, req := range requests {
		resp, err := broker.Produce(req)
		if err != nil {
			return fmt.Errorf("failed to send request to broker %#v. err=%s", broker, err)
		}

		offsets, err := readPartitionOffsetResults(resp)
		if err != nil {
			return fmt.Errorf("failed to read producer response err=%s", err)
		}

		for p, o := range offsets {
			result := map[string]any{"partition": p, "startOffset": o.start, "count": o.count}
			ctx := kt.PrintContext{Output: result, Done: make(chan struct{}), MessageNum: len(batch), ValueSize: valueSize}
			c.out <- ctx
			<-ctx.Done
		}
	}

	return nil
}

func readPartitionOffsetResults(resp *sarama.ProduceResponse) (map[int32]partitionProduceResult, error) {
	offsets := map[int32]partitionProduceResult{}
	for _, blocks := range resp.Blocks {
		for blackPartition, block := range blocks {
			if !errors.Is(block.Err, sarama.ErrNoError) {
				log.Printf("Failed to send message. err=%v\n", block.Err)
				return offsets, block.Err
			}

			if r, ok := offsets[blackPartition]; ok {
				offsets[blackPartition] = partitionProduceResult{start: block.Offset, count: r.count + 1}
			} else {
				offsets[blackPartition] = partitionProduceResult{start: block.Offset, count: 1}
			}
		}
	}
	return offsets, nil
}

func (c *produceCmd) produce(in chan []Message) {
	defer func() { close(c.out) }()

	for b := range in {
		if err := c.produceBatch(c.leaders, b); err != nil {
			log.Printf("produce batch error %v", err.Error())
			return
		}
	}
}

func (c *produceCmd) readInput(q chan struct{}, stdin chan string, out chan string) {
	defer func() { close(out) }()
	for {
		select {
		case l, ok := <-stdin:
			if !ok {
				return
			}
			out <- l
		case <-q:
			return
		}
	}
}

func (c *produceCmd) LongHelp() string {
	return `Input is read from stdin and separated by newlines.

To specify the key, value and partition individually pass it as a JSON object
like the following:
    {"key": "id-23", "value": "message content", "partition": 0}
    {"k": "id-23", "v": "message content", "p": 0}

In case the input line cannot be interpreted as a JSON object the key and value
both default to the input line and partition to 0.

Examples:
Send a single message with a specific key:
  $ echo '{"key": "id-23", "value": "ola", "partition": 0}' | kt produce --topic greetings --json
  Sent message to partition 0 at Offset 3.
  $ echo '{"k": "id-23", "v": "ola", "p": 0}' | kt produce --topic greetings --json
  Sent message to partition 0 at Offset 3.
  $ kt consume --topic greetings --timeout 1s --offsets 0:3-
  {"partition":0,"Offset":3,"key":"id-23","message":"ola"}

Keep reading input from stdin until interrupted (via ^C).
  $ kt produce --topic greetings
  hello.
  Sent message to partition 0 at Offset 4.
  bonjour.
  Sent message to partition 0 at Offset 5.
  $ kt consume --topic greetings --timeout 1s --offsets 0:4-
  {"partition":0,"Offset":4,"key":"hello.","message":"hello."}
  {"partition":0,"Offset":5,"key":"bonjour.","message":"bonjour."}
`
}
