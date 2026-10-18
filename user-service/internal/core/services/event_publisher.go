package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/fouched/novabank/shared/messaging"
	"github.com/fouched/novabank/users/internal/core/models"
)

type AccountEventPublisher struct {
	pub    *messaging.KafkaPublisher
	logger *slog.Logger
}

func NewAccountEventPublisher(brokers []string, logger *slog.Logger) (*AccountEventPublisher, error) {
	topic := "user-registration-topic"

	// Explicitly catch and bubble up the initialization error from franz-go
	publisher, err := messaging.NewKafkaPublisher(brokers, topic)
	if err != nil {
		return nil, fmt.Errorf("failed to build internal kafka publisher: %w", err)
	}

	return &AccountEventPublisher{
		pub:    publisher,
		logger: logger,
	}, nil
}

func (p *AccountEventPublisher) PublishUserRegistrationEvent(ctx context.Context, event models.UserRegistrationEvent) {
	p.logger.Info("Publishing user registration event to Kafka topic", slog.String("topic", "user-registration-topic"), slog.String("key", event.Email))

	if err := p.pub.Publish(ctx, event.Email, event); err != nil {
		p.logger.Error("Failed to publish user registration event to cluster stream", slog.String("email", event.Email), slog.Any("error", err))
		return
	}

	p.logger.Info("User registration event dispatched onto async buffer segments successfully", slog.String("email", event.Email))
}

func (p *AccountEventPublisher) Close() error {
	return p.pub.Close()
}
