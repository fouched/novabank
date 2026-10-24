package postgres

import (
	"context"
	"fmt"

	"github.com/fouched/novabank/notifications/internal/core/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository struct {
	db *pgxpool.Pool
}

// NewNotificationRepository instantiates the explicit pgx/v5 implementation layer
func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Save inserts a new notification record into fintech_notification_db and populates generated fields
func (r *NotificationRepository) Save(ctx context.Context, n *models.Notification) error {
	query := `
		INSERT INTO notifications (
			type, 
			status, 
			message, 
			recipient_email, 
			recipient_phone, 
			subject, 
			transaction_reference
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id, created_at`

	// Using QueryRow with Scan mirrors the pgx/v5 pattern we established in the user-service
	err := r.db.QueryRow(
		ctx,
		query,
		n.Type,
		n.Status,
		n.Message,
		// Go handles empty string fields mapping to nullable database text rows cleanly.
		// If explicit database NULL types are needed later, we can map to *string pointers.
		n.RecipientEmail,
		n.RecipientPhone,
		n.Subject,
		n.TransactionReference,
	).Scan(&n.ID, &n.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to persist notification tracking record: %w", err)
	}

	return nil
}
