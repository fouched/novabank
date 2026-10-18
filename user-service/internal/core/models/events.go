package models

type UserRegistrationEvent struct {
	Email         string `json:"email"`
	FirstName     string `json:"firstName"` // Lower CamelCase guarantees compatibility with Java consumer naming rules
	LastName      string `json:"lastName"`
	AccountNumber string `json:"accountNumber"`
	BankName      string `json:"bankName"`
}
