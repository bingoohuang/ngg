package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/IBM/sarama"
	"github.com/bingoohuang/ngg/kt/pkg/kt"
	"github.com/elliotchance/pie/v2"
	"github.com/spf13/cobra"
)

type topicCmd struct {
	kt.CommonArgs `squash:"1"`

	Grep   string `short:"g" help:"topic grep string"`
	Detail bool   `short:"d" help:"Include details like partitions/replica/leader"`
	Config bool   `short:"c" help:"Include topic configuration"`
	All    bool   `short:"a" help:"Show all topics like __consumer_offsets, etc."`
	admin  sarama.ClusterAdmin

	client   sarama.Client
	grepExpr *regexp.Regexp
	out      chan kt.PrintContext
}

func (c *topicCmd) Run(*cobra.Command, []string) (err error) {
	if err := c.CommonArgs.Validate(); err != nil {
		return err
	}

	if c.Grep != "" {
		if c.grepExpr, err = regexp.Compile(c.Grep); err != nil {
			return fmt.Errorf("regex %q is invalid: %v", c.Grep, err)
		}
	}

	if err := c.connect(); err != nil {
		return err
	}
	defer c.client.Close()

	var topics []string
	if c.Topic != "" {
		topics = []string{c.Topic}
	} else {
		if topics, err = c.client.Topics(); err != nil {
			return err
		}
	}

	if c.grepExpr != nil {
		topics = pie.Of(topics).Filter(c.grepExpr.MatchString).Result
	}

	c.out = make(chan kt.PrintContext)
	go kt.PrintOut(c.out)

	var wg sync.WaitGroup
	for _, tn := range topics {
		if !c.All && strings.HasPrefix(tn, "__") {
			continue
		}

		wg.Add(1)
		go func(topic string) {
			c.print(topic)
			wg.Done()
		}(tn)
	}
	wg.Wait()
	return nil
}

type topicInfo struct {
	Config     map[string]string `json:"config,omitempty"`
	Name       string            `json:"name"`
	Partitions []partition       `json:"partitions,omitempty"`
}

type partition struct {
	Leader       string  `json:"leader,omitempty"`
	Replicas     []int32 `json:"replicas,omitempty"`
	ISRs         []int32 `json:"isrs,omitempty"`
	OldestOffset int64   `json:"oldestOffset"`
	NewestOffset int64   `json:"newestOffset"`
	ID           int32   `json:"id"`
}

func (c *topicCmd) connect() error {
	client, err := c.SetupClient("kt-topic-" + kt.CurrentUserName())
	if err != nil {
		return err
	}
	c.client = client.SaramaClient

	if c.admin, err = sarama.NewClusterAdminFromClient(c.client); err != nil {
		return err
	}
	return nil
}

func (c *topicCmd) print(name string) {
	top, err := c.readTopic(name)
	if err != nil {
		log.Printf("E! read info for topic %s: %v", name, err)
		return
	}

	ctx := kt.PrintContext{Output: top, Done: make(chan struct{})}
	c.out <- ctx
	<-ctx.Done
}

func (c *topicCmd) readTopic(name string) (topicInfo, error) {
	top := topicInfo{Name: name}

	if c.Config {
		resource := sarama.ConfigResource{Name: name, Type: sarama.TopicResource}
		configEntries, err := c.admin.DescribeConfig(resource)
		if err != nil {
			return top, err
		}

		top.Config = make(map[string]string)
		for _, entry := range configEntries {
			top.Config[entry.Name] = entry.Value
		}
	}

	if !c.Detail {
		return top, nil
	}

	var err error
	var ps []int32
	if ps, err = c.client.Partitions(name); err != nil {
		return top, err
	}

	for _, p := range ps {
		np := partition{ID: p}

		if np.OldestOffset, err = c.client.GetOffset(name, p, sarama.OffsetOldest); err != nil {
			return top, err
		}

		if np.NewestOffset, err = c.client.GetOffset(name, p, sarama.OffsetNewest); err != nil {
			return top, err
		}

		led, err := c.client.Leader(name, p)
		if err != nil {
			return top, err
		}

		np.Leader = led.Addr()

		if np.Replicas, err = c.client.Replicas(name, p); err != nil {
			return top, err
		}

		if np.ISRs, err = c.client.InSyncReplicas(name, p); err != nil {
			return top, err
		}

		top.Partitions = append(top.Partitions, np)
	}

	return top, nil
}
