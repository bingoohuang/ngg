package main

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/IBM/sarama"
	"github.com/bingoohuang/ngg/kt/pkg/kt"
	"github.com/bingoohuang/ngg/ss"
	"github.com/elliotchance/pie/v2"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type groupCmd struct {
	kt.ConsumerConfig `squash:"1"`

	FilterGroups string `help:"filter groups by regexp"`
	FilterTopics string `help:"filter topics by regexp"`
	Reset        string `help:"Target Offset to reset for consumer group (newest, oldest, or specific Offset)"`
	Partitions   string `default:"all" help:"comma separated list of partitions to limit offsets to, or all"`
	FetchOffsets bool   `default:"true" help:"Controls if offsets should be fetched"`
	Tags         string `help:"available tags: allOffSets"`

	client       sarama.Client
	filterGroups *regexp.Regexp
	filterTopics *regexp.Regexp
	out          chan kt.PrintContext
	partitions   []int32
	reset        int64
}

const (
	resetNotSpecified = -23
)

func (c *groupCmd) Run(*cobra.Command, []string) (err error) {
	if err := c.CommonArgs.Validate(); err != nil {
		return err
	}

	if err := c.parseArgs(); err != nil {
		return err
	}

	if c.client, err = sarama.NewClient(c.KafkaBrokers, c.saramaConfig()); err != nil {
		failf("E! create client: %v", err)
	}

	brokers := c.client.Brokers()
	log.Printf("found %v brokers: %s", len(brokers),
		kt.ColorJSON(pie.Map(brokers, func(b *sarama.Broker) string { return b.Addr() })))

	groups := []string{c.Group}
	if c.Group == "" {
		groups = pie.Of(c.findGroups(brokers)).Filter(c.filterGroups.MatchString).Result
		groups = lo.Uniq(groups)
		log.Printf("found %d groups: %s", len(groups), kt.ColorJSON(groups))
	}

	topics := []string{c.Topic}
	if c.filterTopics != nil {
		topics = pie.Of(c.fetchTopics()).Filter(c.filterTopics.MatchString).Result
		topics = lo.Uniq(topics)
		log.Printf("found %d topics: %s", len(topics), kt.ColorJSON(topics))
	}

	c.out = make(chan kt.PrintContext)
	go kt.PrintOut(c.out)

	if !c.FetchOffsets {
		for i, group := range groups {
			ctx := kt.PrintContext{Output: kt.GroupInfo{Group: group}, Done: make(chan struct{})}
			c.out <- ctx
			<-ctx.Done

			if c.Verbose > 0 {
				log.Printf("%v/%v\n", i+1, len(groups))
			}
		}
		return
	}

	topicPartitions := map[string][]int32{}
	for _, topic := range topics {
		parts := c.partitions
		if len(parts) == 0 {
			parts = c.fetchPartitions(topic)
			log.Printf("found %d partitions %s for topic %s", len(parts), kt.ColorJSON(parts), topic)
		}
		topicPartitions[topic] = parts
	}

	wg := &sync.WaitGroup{}
	for _, group := range groups {
		for topic, partitions := range topicPartitions {
			wg.Add(1)
			go func(group, topic string, partitions []int32) {
				defer wg.Done()
				c.printGroupTopicOffset(group, topic, partitions)
			}(group, topic, partitions)
		}
	}
	wg.Wait()
	return nil
}

func (c *groupCmd) printGroupTopicOffset(group, topic string, partitions []int32) {
	target := kt.GroupInfo{Group: group, Topic: topic, Offsets: make([]kt.GroupOffset, 0, len(partitions))}
	results := make(chan kt.GroupOffset)

	wg := &sync.WaitGroup{}
	for _, part := range partitions {
		wg.Add(1)
		go func(part int32) {
			defer wg.Done()
			c.fetchGroupOffset(group, topic, part, results)
		}(part)
	}
	go func() { wg.Wait(); close(results) }()

	for res := range results {
		target.Offsets = append(target.Offsets, res)
	}

	if len(target.Offsets) > 0 {
		sort.Slice(target.Offsets, func(i, j int) bool {
			return target.Offsets[j].Partition > target.Offsets[i].Partition
		})
		ctx := kt.PrintContext{Output: target, Done: make(chan struct{})}
		c.out <- ctx
		<-ctx.Done
	}
}

func (c *groupCmd) resolveOffset(top string, part int32, off int64) int64 {
	resolvedOff, err := c.client.GetOffset(top, part, off)
	if err != nil {
		failf("get Offset for partition=%d: %v", part, err)
	}

	if c.Verbose > 0 {
		log.Printf("resolved Offset %v for topic=%s partition=%d to %v", off, top, part, resolvedOff)
	}

	return resolvedOff
}

func (c *groupCmd) fetchGroupOffset(group, topic string, part int32, results chan<- kt.GroupOffset) {
	if c.Verbose > 0 {
		log.Printf("fetching Offset information for group=%v topic=%v partition=%v", group, topic, part)
	}

	offsetManager, err := sarama.NewOffsetManagerFromClient(group, c.client)
	if err != nil {
		failf("create client: %v", err)
	}
	defer kt.LogClose("Offset manager", offsetManager)

	pom, err := offsetManager.ManagePartition(topic, part)
	if err != nil {
		failf("manage partition group=%s topic=%s partition=%d: %v", group, topic, part, err)
	}
	defer kt.LogClose("partition Offset manager", pom)

	specialOffset := c.reset == sarama.OffsetNewest || c.reset == sarama.OffsetOldest

	groupOff, metadata := pom.NextOffset()
	if c.reset >= 0 || specialOffset {
		resolvedOff := c.reset
		if specialOffset {
			resolvedOff = c.resolveOffset(topic, part, c.reset)
		}
		if resolvedOff > groupOff {
			pom.MarkOffset(resolvedOff, "")
		} else {
			pom.ResetOffset(resolvedOff, "")
		}

		groupOff = resolvedOff
	}

	// we haven't reset it, and it wasn't set before - lag depends on client's config
	if specialOffset {
		results <- kt.GroupOffset{Partition: part}
		return
	}

	partOff := c.resolveOffset(topic, part, sarama.OffsetNewest)
	lag := partOff - groupOff

	if groupOff > 0 || ss.ContainsFold(c.Tags, "allOffSets") {
		results <- kt.GroupOffset{Partition: part, PartitionOffset: partOff, GroupOffset: groupOff, Lag: lag, Metadata: metadata}
	}
}

func (c *groupCmd) fetchTopics() []string {
	tps, err := c.client.Topics()
	if err != nil {
		failf("read topics: %v", err)
	}
	return tps
}

func (c *groupCmd) fetchPartitions(top string) []int32 {
	ps, err := c.client.Partitions(top)
	if err != nil {
		failf("read partitions for topic=%s: %v", top, err)
	}
	return ps
}

func (c *groupCmd) findGroups(brokers []*sarama.Broker) (groups []string) {
	var wg sync.WaitGroup

	groupCh := make(chan string)
	for _, broker := range brokers {
		wg.Add(1)
		go func(broker *sarama.Broker) {
			defer wg.Done()
			c.findGroupsOnBroker(broker, groupCh)
		}(broker)
	}

	go func() {
		wg.Wait()
		close(groupCh)
	}()

	for group := range groupCh {
		groups = append(groups, group)
	}

	return groups
}

func (c *groupCmd) findGroupsOnBroker(broker *sarama.Broker, groupCh chan string) {
	if err := c.connect(broker); err != nil {
		failf("failed to connect to broker %s err=%q", broker.Addr(), err)
	}

	resp, err := broker.ListGroups(&sarama.ListGroupsRequest{})
	if err != nil {
		failf("failed to list brokers on %s err=%q", broker.Addr(), err)
	}

	if !errors.Is(resp.Err, sarama.ErrNoError) {
		failf("failed to list brokers on %s err=%v", broker.Addr(), resp.Err)
	}

	for name := range resp.Groups {
		groupCh <- name
	}
}

func (c *groupCmd) connect(broker *sarama.Broker) error {
	if ok, _ := broker.Connected(); ok {
		return nil
	}

	if err := broker.Open(c.saramaConfig()); err != nil {
		return err
	}

	connected, err := broker.Connected()
	if err != nil {
		return err
	}

	if !connected {
		return fmt.Errorf("failed to connect broker %#v", broker.Addr())
	}

	return nil
}

func (c *groupCmd) saramaConfig() *sarama.Config {
	c.SetupClient("kt-group-" + kt.CurrentUserName())
	sc := sarama.NewConfig()
	sc.Version = c.KafkaVersion
	sc.ClientID = "kt-group-" + kt.CurrentUserName()
	if err := c.SetupAuth(sc); err != nil {
		failf("failed to setup auth: %v", err)
	}
	if err := c.Validate(); err != nil {
		failf("configuration validate: %v", err)
	}

	return sc
}

func (c *groupCmd) failStartup(msg string) {
	log.Print(msg)
	failf("use \"kt group -help\" for more information")
}

func (c *groupCmd) parseArgs() error {
	var err error

	switch c.Partitions {
	case "", "all":
		c.partitions = []int32{}
	default:
		pss := strings.Split(c.Partitions, ",")
		for _, ps := range pss {
			p, err := strconv.ParseInt(ps, 10, 32)
			if err != nil {
				failf("partition id invalid err=%v", err)
			}
			c.partitions = append(c.partitions, int32(p))
		}
	}

	if c.partitions == nil {
		return fmt.Errorf(`invalid %#v. Should be a comma separated list of partitions or "all"`, c.Partitions)
	}

	if c.filterGroups, err = regexp.Compile(c.FilterGroups); err != nil {
		return fmt.Errorf("groups filter regexp invalid: %w", err)
	}

	if c.filterTopics, err = regexp.Compile(c.FilterTopics); err != nil {
		return fmt.Errorf("topics filter regexp invalid: %w", err)
	}

	if c.Reset != "" && (c.Topic == "" || c.Group == "") {
		return fmt.Errorf("group and topic are required to reset offsets")
	}

	switch c.Reset {
	case "newest":
		c.reset = sarama.OffsetNewest
	case "oldest":
		c.reset = sarama.OffsetOldest
	case "":
		// optional flag
		c.reset = resetNotSpecified
	default:
		c.reset, err = strconv.ParseInt(c.Reset, 10, 64)
		if err != nil {
			if c.Verbose > 0 {
				log.Printf("failed to parse set %s: %v", c.Reset, err)
			}
			c.failStartup(fmt.Sprintf(`Invalid reset %q. either newest, oldest or specific Offset expected.`, c.Reset))
		}
	}

	return nil
}

func (c *groupCmd) LongHelp() string {
	return `
The group command can be used to list groups, their offsets and lag and to reset a group's Offset.

When an explicit Offset hasn't been set yet, kt prints out the respective sarama constants, cf. https://godoc.org/github.com/IBM/sarama#pkg-constants

1. To simply list all groups: kt group
2. This is faster when not fetching offsets: kt group -offsets=false
3. To filter by regex: kt group --filter specials
4. To filter by topic: kt group --topic fav-topic
5. To reset a consumer group's Offset: kt group --reset 23 --topic fav-topic --group specials --partitions 2
6. To reset a consumer group's Offset for all partitions: kt group --reset newest --topic fav-topic --group specials --partitions all
`
}
