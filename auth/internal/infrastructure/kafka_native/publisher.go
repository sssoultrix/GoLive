package kafka_native

import (
	"context"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
)

type Publisher struct {
	Client *kgo.Client
	Topic  string
}

func New(brokers []string, topic string) (*Publisher, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		// keep it simple: no TLS/SASL here; local Redpanda is plaintext
	)
	if err != nil {
		return nil, err
	}
	return &Publisher{Client: cl, Topic: topic}, nil
}

func (p *Publisher) Close() {
	if p != nil && p.Client != nil {
		p.Client.Close()
	}
}

func (p *Publisher) Publish(ctx context.Context, key string, value []byte) error {
	if p == nil || p.Client == nil {
		return fmt.Errorf("kafka publisher is nil")
	}
	if p.Topic == "" {
		return fmt.Errorf("kafka topic is empty")
	}

	rec := &kgo.Record{
		Topic: p.Topic,
		Key:   []byte(key),
		Value: value,
	}
	res := p.Client.ProduceSync(ctx, rec)
	return res.FirstErr()
}

// Ensure franz-go types are linked (sometimes helpful for tooling).
var _ = kmsg.RecordBatch{}
var _ = time.Second

