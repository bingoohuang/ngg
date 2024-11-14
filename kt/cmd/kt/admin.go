package main

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/bingoohuang/ngg/kt/pkg/kt"
)

type adminCmd struct {
	kt.CommonArgs `squash:"1"`
	Timeout       time.Duration `default:"3s"`

	Config       string `help:"topic config"`
	CreateTopic  string `help:"create topic"`
	DeleteTopic  string `help:"delete topic"`
	ValidateOnly bool   `help:""`

	admin       sarama.ClusterAdmin
	topicDetail *sarama.TopicDetail
}

// TopicDetail is structure convenient of topic.config  in topic.create.
type TopicDetail struct {
	NumPartitions     *int32              `json:"numPartitions,omitempty"`
	ReplicationFactor *int16              `json:"replicationFactor,omitempty"`
	ReplicaAssignment *map[int32][]int32  `json:"replicaAssignment,omitempty"`
	ConfigEntries     *map[string]*string `json:"configEntries,omitempty"`
}

func (r *adminCmd) Run(*cobra.Command, []string) error {
	if err := r.CommonArgs.Validate(); err != nil {
		return err
	}

	if r.CreateTopic != "" {
		var (
			err error
			buf []byte
		)
		if strings.HasPrefix(r.Config, "@") {
			buf, err = os.ReadFile(r.Config[1:])
			if err != nil {
				failf("failed to read topic detail err=%v", err)
			}
		} else {
			buf = []byte(r.Config)
		}

		var detail TopicDetail
		if err = json.Unmarshal(buf, &detail); err != nil {
			failf("failed to unmarshal topic detail err=%v", err)
		}

		d := &sarama.TopicDetail{}
		d.NumPartitions = kt.FirstNotNil(detail.NumPartitions)
		d.ReplicationFactor = kt.FirstNotNil(detail.ReplicationFactor)
		d.ReplicaAssignment = kt.FirstNotNil(detail.ReplicaAssignment)
		d.ConfigEntries = kt.FirstNotNil(detail.ConfigEntries)
		r.topicDetail = d
	}

	var err error
	if r.admin, err = sarama.NewClusterAdmin(r.KafkaBrokers, r.saramaConfig()); err != nil {
		failf("failed to create cluster admin err=%v", err)
	}

	if r.CreateTopic != "" {
		if err := r.admin.CreateTopic(r.CreateTopic, r.topicDetail, r.ValidateOnly); err != nil {
			failf("failed to create topic err=%v", err)
		}
	}

	if r.DeleteTopic != "" {
		if err := r.admin.DeleteTopic(r.DeleteTopic); err != nil {
			failf("failed to delete topic err=%v", err)
		}
	}

	return nil
}

func (r *adminCmd) saramaConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Version = r.KafkaVersion
	cfg.ClientID = "kt-admin-" + kt.CurrentUserName()

	if r.Timeout > 0 {
		cfg.Admin.Timeout = r.Timeout
	}

	if err := r.SetupAuth(cfg); err != nil {
		failf("failed to setup auth: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		failf("configuration validate: %v", err)
	}

	return cfg
}

func (r *adminCmd) LongHelp() string {
	return `
The topic details should be passed via a JSON file that represents a sarama.TopicDetail struct.
cf https://godoc.org/github.com/IBM/sarama#TopicDetail

A simple way to pass a JSON file is to use a tool like https://github.com/fgeller/jsonify and shell's process substition:

kt admin --create-topic morenews --config '{"numPartitions": 1, "replicationFactor": 1 )`
}
