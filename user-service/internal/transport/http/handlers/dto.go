package handlers

// ErrorResponse represents the clean, unified failure layout for transport errors
type ErrorResponse struct {
	Error string `json:"error"`
}
