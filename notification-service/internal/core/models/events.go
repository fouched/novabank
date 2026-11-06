package models

import "time"

type UserRegistrationEvent struct {
	Email         string `json:"email"`
	FirstName     string `json:"firstName"` // Lower CamelCase guarantees compatibility with Java consumer naming rules
	LastName      string `json:"lastName"`
	AccountNumber string `json:"accountNumber"`
	BankName      string `json:"bankName"`
}

type TransactionEvent struct {
	Email         string    `json:"email"`
	FirstName     string    `json:"firstName"`
	LastName      string    `json:"lastName"`
	AccountNumber string    `json:"accountNumber"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Reference     string    `json:"reference"`
	Description   string    `json:"description"`
	Balance       float64   `json:"balance"`
	Date          time.Time `json:"date"`
}
