package messaging

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaConsumer struct {
	client *kgo.Client
}

func NewKafkaConsumer(brokers []string, groupID string, topics []string) (*KafkaConsumer, error) {
	// Construct a modern, thread-safe consumer instance profile mapping group settings
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topics...),

		// Use the franz-go syntax to start consuming at the end of partitions
		kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize franz-go consumer client: %w", err)
	}

	return &KafkaConsumer{client: cl}, nil
}

// FetchRecords blocks execution threads until a batch of events strike the topic streams
func (c *KafkaConsumer) FetchRecords(ctx context.Context) ([]*kgo.Record, error) {
	fetches := c.client.PollRecords(ctx, 0) // 0 means poll asynchronously block until payload drops
	if err := fetches.Err(); err != nil {
		return nil, fmt.Errorf("failed polling records out of streaming topic segments: %w", err)
	}

	return fetches.Records(), nil
}

// CommitRecords signals back to Kafka that these records have been processed safely
func (c *KafkaConsumer) CommitRecords(ctx context.Context, records ...*kgo.Record) error {
	return c.client.CommitRecords(ctx, records...)
}

func (c *KafkaConsumer) Close() error {
	c.client.Close()
	return nil
}
