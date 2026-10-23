package repositories

import (
	"context"

	"github.com/fouched/novabank/notifications/internal/core/models"
)

type NotificationRepository interface {
	Save(ctx context.Context, n *models.Notification) error
}
