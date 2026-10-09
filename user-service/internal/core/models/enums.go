package models

// AccountStatus handles operational banking domain states
type AccountStatus string

const (
	AccountActive AccountStatus = "ACTIVE"
	AccountFrozen AccountStatus = "FROZEN"
	AccountClosed AccountStatus = "CLOSED"
)

// AccountType matches available portfolio instruments
type AccountType string

const (
	TypeSavings AccountType = "SAVINGS"
	TypeCredit  AccountType = "CREDIT"
	TypeCurrent AccountType = "CURRENT"
)

// Currency represents ISO currency denominations
type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyZAR Currency = "ZAR"
	CurrencyEUR Currency = "EUR"
	// Extend other variables (GBP, JPY etc.) as needed...
)
