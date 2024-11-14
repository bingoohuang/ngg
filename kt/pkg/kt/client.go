package kt

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

type Client struct {
	SaramaClient sarama.Client
	SaramaConfig *sarama.Config
}

func (c *Client) NewOffsetManager(group string) (sarama.OffsetManager, error) {
	if group == "" {
		return nil, nil
	}

	return sarama.NewOffsetManagerFromClient(group, c.SaramaClient)
}

func (c CommonArgs) SetupClient(clientID string) (*Client, error) {
	sc := sarama.NewConfig()
	sc.Version = c.KafkaVersion
	sc.ClientID = clientID

	if err := c.SetupAuth(sc); err != nil {
		return nil, err
	}
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validate: %w", err)
	}

	if c.Verbose > 0 {
		log.Printf("sarama client configuration %#v\n", sc)
	}

	client, err := sarama.NewClient(c.KafkaBrokers, sc)
	if err != nil {
		return nil, err
	}

	return &Client{SaramaClient: client, SaramaConfig: sc}, err
}
