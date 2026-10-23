package models

import "time"

type Notification struct {
	ID                   int64              `json:"id"`
	Type                 NotificationType   `json:"type"`
	Status               NotificationStatus `json:"status"`
	Message              string             `json:"message"`
	RecipientEmail       string             `json:"recipient_email,omitempty"`
	RecipientPhone       string             `json:"recipient_phone,omitempty"`
	Subject              string             `json:"subject,omitempty"`
	TransactionReference string             `json:"transaction_reference,omitempty"`
	CreatedAt            time.Time          `json:"created_at"`
}
