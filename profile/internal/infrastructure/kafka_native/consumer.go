package kafka_native

import (
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Consumer struct {
	Client *kgo.Client
	Topic  string
}

func NewConsumer(brokers []string, group string, topic string) (*Consumer, error) {
	if topic == "" {
		return nil, fmt.Errorf("topic is empty")
	}
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
	)
	if err != nil {
		return nil, err
	}
	return &Consumer{Client: cl, Topic: topic}, nil
}

func (c *Consumer) Close() {
	if c != nil && c.Client != nil {
		c.Client.Close()
	}
}

