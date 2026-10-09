package models

import (
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Hidden from outbound json transmission structures
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Enabled   bool      `json:"enabled"`
	Roles     []Role    `json:"roles,omitempty"`
	Accounts  []Account `json:"accounts,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Role struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Account struct {
	ID            int64         `json:"id"`
	AccountNumber string        `json:"account_number"`
	Balance       float64       `json:"balance"` // Will map from pgx pool numeric data fields smoothly
	Currency      Currency      `json:"currency"`
	AccountType   AccountType   `json:"account_type"`
	AccountStatus AccountStatus `json:"account_status"`
	UserID        int64         `json:"user_id"`
	CreatedAt     time.Time     `json:"created_at"`
}
