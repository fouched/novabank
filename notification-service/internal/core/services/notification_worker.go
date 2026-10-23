package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/fouched/novabank/notifications/internal/core/models"
	"github.com/fouched/novabank/shared/messaging"
	"github.com/twmb/franz-go/pkg/kgo"
)

type NotificationWorker struct {
	consumer  *messaging.KafkaConsumer
	publisher *messaging.KafkaPublisher // Used exclusively to forward failed records to the DLT
	logger    *slog.Logger
}

func NewNotificationWorker(brokers []string, groupID string, logger *slog.Logger) (*NotificationWorker, error) {
	topics := []string{"user-registration-topic"}

	cons, err := messaging.NewKafkaConsumer(brokers, groupID, topics)
	if err != nil {
		return nil, fmt.Errorf("failed to init consumer: %w", err)
	}

	// Instantiate a fallback publisher pointing to the dead-letter topology destination suffix
	pub, err := messaging.NewKafkaPublisher(brokers, "user-registration-topic-dlt")
	if err != nil {
		cons.Close()
		return nil, fmt.Errorf("failed to init dlt publisher: %w", err)
	}

	return &NotificationWorker{
		consumer:  cons,
		publisher: pub,
		logger:    logger,
	}, nil
}

func (w *NotificationWorker) Start(ctx context.Context) {
	w.logger.Info("Notification service background engine active, polling stream chunks...")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Stopping consumer background loops safely...")
			return
		default:
			// Fetch fresh record matrix sets from the active partitions
			records, err := w.consumer.FetchRecords(ctx)
			if err != nil {
				w.logger.Error("Stream polling error encountered", slog.Any("error", err))
				time.Sleep(1 * time.Second) // Protect system resources from rapid failure spin cycles
				continue
			}

			for _, record := range records {
				w.processWithRetryPolicy(ctx, record)
			}
		}
	}
}

func (w *NotificationWorker) processWithRetryPolicy(ctx context.Context, record *kgo.Record) {
	maxRetries := 2
	backoff := 1 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := w.executeBusinessLogic(ctx, record.Value)
		if err == nil {
			// Success -> Commit tracking metrics straight up to the cluster coordinators
			if commitErr := w.consumer.CommitRecords(ctx, record); commitErr != nil {
				w.logger.Error("Failed committing offset verification mark", slog.Any("error", commitErr))
			}
			return
		}

		w.logger.Warn("Event execution pass failed",
			slog.Int("attempt", attempt+1),
			slog.Int("max_retries", maxRetries+1),
			slog.Any("error", err),
		)

		if attempt < maxRetries {
			time.Sleep(backoff)
			continue
		}
	}

	// Retries exhausted -> Route raw message payload safely to the Dead Letter Topic (-dlt)
	w.logger.Error("Processing constraints exhausted. Route event payload directly to DLT", slog.String("key", string(record.Key)))

	dltCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if dltErr := w.publisher.Publish(dltCtx, string(record.Key), record.Value); dltErr != nil {
		w.logger.Error("CRITICAL: Failed forwarding payload onto dead-letter storage segments", slog.Any("error", dltErr))
		return
	}

	// Commit the poisoned record offset anyway so the consumer partition doesn't freeze permanently
	_ = w.consumer.CommitRecords(ctx, record)
}

func (w *NotificationWorker) executeBusinessLogic(ctx context.Context, value []byte) error {
	var event models.UserRegistrationEvent
	if err := json.Unmarshal(value, &event); err != nil {
		// Deserialization errors are non-recoverable (Spring parity logic rule match)
		return fmt.Errorf("malformed structure payload unmarshal match failed: %w", err)
	}

	// Placeholder verification context line mapping
	w.logger.Info("Processing record tracking frame context", slog.String("email", event.Email))

	// TODO: Wire repository.Save() and SMTP transport hooks when Dennis defines the service tier
	return nil
}

func (w *NotificationWorker) Close() {
	if err := w.consumer.Close(); err != nil {
		w.logger.Error("Error closing franz-go consumer stream", slog.Any("error", err))
	}
	if err := w.publisher.Close(); err != nil {
		w.logger.Error("Error terminating dead-letter publisher channel", slog.Any("error", err))
	}
}
