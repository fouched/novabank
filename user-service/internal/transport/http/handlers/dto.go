package handlers

import "github.com/fouched/novabank/users/internal/core/models"

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegistrationRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

// --- Response DTOs ---

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

type UserStatisticsResponse struct {
	TotalUsers             int64   `json:"total_users"`
	ActiveUsers            int64   `json:"active_users"`
	InactiveUsers          int64   `json:"inactive_users"`
	TotalAccounts          int64   `json:"total_accounts"`
	AverageAccountsPerUser float64 `json:"average_accounts_per_user"`
	CustomersCount         int64   `json:"customers_count"`
	AdminsCount            int64   `json:"admins_count"`
}

// ErrorResponse represents the clean, unified failure layout for transport errors
type ErrorResponse struct {
	Error string `json:"error"`
}
