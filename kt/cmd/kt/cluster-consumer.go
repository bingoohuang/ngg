package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/bingoohuang/ngg/ggt/gterm"
	"github.com/bingoohuang/ngg/kt/pkg/kt"
	cluster "github.com/bingoohuang/ngg/kt/pkg/sarama-cluster"
	"github.com/bingoohuang/ngg/ss"
	"github.com/spf13/cobra"
)

type clusterConsumer struct {
	kt.CommonArgs `squash:"1"`
	Group         string `default:"kt" help:"Consumer group to use for marking offsets. kt will mark offsets if group is supplied"`
	Offset        string `default:"newest" enum:"oldest,newest" help:"The offset to start with"`
	Timeout       time.Duration
	N             int `help:"Max number of messages to consume"`
}

func (p *clusterConsumer) Run(*cobra.Command, []string) error {
	if err := p.CommonArgs.Validate(); err != nil {
		return err
	}

	c := cluster.NewConfig()
	if p.Verbose > 0 {
		sarama.Logger = log.New(os.Stderr, "", log.LstdFlags)
	} else {
		c.Consumer.Return.Errors = true
		c.Group.Return.Notifications = true
	}

	c.Version = p.KafkaVersion

	env := os.Getenv("KT_AUTH")
	if env != "" && p.SASL == "" {
		p.SASL = env
	}
	if p.SASL != "" {
		data, err := gterm.DecodeByTailTag(p.SASL)
		if err != nil {
			return fmt.Errorf("decode %q: %w", p.SASL, err)
		}
		user, pwd := ss.Split2(string(data), ":")
		c.Net.SASL.Enable = true
		c.Net.SASL.User = user
		c.Net.SASL.Password = pwd
		c.Net.SASL.Handshake = true
		c.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		c.Net.SASL.Version = sarama.SASLHandshakeV0
	}
	switch p.Offset {
	case "oldest":
		c.Consumer.Offsets.Initial = sarama.OffsetOldest
	case "newest":
		c.Consumer.Offsets.Initial = sarama.OffsetNewest
	}

	// Init consumer, consume errors & messages
	consumer, err := cluster.NewConsumer(p.KafkaBrokers, p.Group, []string{p.Topic}, c)
	if err != nil {
		return fmt.Errorf("start consumer: %w", err)
	}
	defer consumer.Close()

	// Create signal channel
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	// Consume all channels, wait for signal to exit
	for i := 1; p.N <= 0 || i <= p.N; {
		select {
		case msg, ok := <-consumer.Messages():
			if ok {
				log.Printf("#%d: Topic: %s Partition: %d Offset: %d\n     Vaule: %s\n",
					i, msg.Topic, msg.Partition, msg.Offset, msg.Value)
				i++
				consumer.MarkOffset(msg, "")
			}
		case ntf, ok := <-consumer.Notifications():
			if ok {
				log.Printf("Rebalanced: %+v\n", ntf)
			}
		case err, ok := <-consumer.Errors():
			if ok {
				log.Printf("Error: %s\n", err.Error())
			}
		case <-sigCh:
			return nil
		}
	}

	return nil
}
