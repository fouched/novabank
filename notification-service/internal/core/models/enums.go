package models

// NotificationType represents available notification types
type NotificationType string

const (
	TypeSavings NotificationType = "EMAIL"
	TypeCredit  NotificationType = "SMS"
)

// NotificationStatus represents available notification types
type NotificationStatus string

const (
	StatusSent   NotificationStatus = "SENT"
	StatusFailed NotificationStatus = "FAILED"
)

// Currency represents ISO currency denominations
type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyZAR Currency = "ZAR"
	CurrencyEUR Currency = "EUR"
	// Extend other variables (GBP, JPY etc.) as needed...
)
