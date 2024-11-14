package kt

import (
	"fmt"
	"github.com/IBM/sarama"
)

type Client struct {
	SaramaClient sarama.Client
}

func (c *Client) NewOffsetManager(group string) (sarama.OffsetManager, error) {
	if group == "" {
		return nil, nil
	}

	return sarama.NewOffsetManagerFromClient(group, c.SaramaClient)
}

func (c ConsumerConfig) SetupClient() (*Client, error) {
	sc := sarama.NewConfig()
	sc.Version = c.KafkaVersion
	sc.ClientID = "kt-consume-" + CurrentUserName()

	if err := c.SetupAuth(sc); err != nil {
		return nil, err
	}
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validate: %w", err)
	}

	client, err := sarama.NewClient(c.KafkaBrokers, sc)
	if err != nil {
		return nil, err
	}

	return &Client{SaramaClient: client}, err
}
