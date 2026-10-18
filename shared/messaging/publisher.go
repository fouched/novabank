package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaPublisher struct {
	client *kgo.Client
	topic  string
}

func NewKafkaPublisher(brokers []string, topic string) (*KafkaPublisher, error) {
	// Franz-go configures a single highly-optimized client block configuration
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.DefaultProduceTopic(topic),
		kgo.RecordPartitioner(kgo.StickyKeyPartitioner(nil)), // Automatic, ultra-fast key-hashing partition strategy
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize franz-go client: %w", err)
	}

	return &KafkaPublisher{
		client: cl,
		topic:  topic,
	}, nil
}

func (p *KafkaPublisher) Publish(ctx context.Context, key string, payload any) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed serializing data structure payload: %w", err)
	}

	// Franz-go utilizes an explicit block pointer record allocation setup
	record := &kgo.Record{
		Topic: p.topic,
		Key:   []byte(key),
		Value: bytes,
	}

	// Dispatches non-blocking async requests. It fires an internal callback hook once delivered
	p.client.Produce(ctx, record, func(r *kgo.Record, err error) {
		if err != nil {
			// In production, route this error to your high-level slog diagnostics engine
			fmt.Printf("Asynchronous delivery audit track failure observed: %v\n", err)
		}
	})

	return nil
}

func (p *KafkaPublisher) Close() error {
	p.client.Close()
	return nil
}
