package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/fouched/novabank/users/internal/core/faults"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteError acts like the GlobalExceptionHandler block in Spring Boot
func WriteError(w http.ResponseWriter, logger *slog.Logger, err error) {
	statusCode := http.StatusInternalServerError
	message := "An internal server error occurred"

	// Unwrap and identify the application domain fault
	var appErr *faults.ApplicationError
	if errors.As(err, &appErr) {
		message = appErr.Message

		if errors.Is(appErr.Unwrap(), faults.ErrNotFound) {
			statusCode = http.StatusNotFound
		} else if errors.Is(appErr.Unwrap(), faults.ErrBadRequest) {
			statusCode = http.StatusBadRequest
		}
	} else if errors.Is(err, faults.ErrNotFound) {
		statusCode = http.StatusNotFound
		message = err.Error()
	} else if errors.Is(err, faults.ErrBadRequest) {
		statusCode = http.StatusBadRequest
		message = err.Error()
	}

	// Log with high-level context using your slog toolkit wrapper
	if statusCode == http.StatusInternalServerError {
		logger.Error("Generic execution failure observed", slog.Any("error", err))
	} else {
		logger.Warn("Application operational error handled",
			slog.Int("status_code", statusCode),
			slog.String("detail", message),
		)
	}

	// Write clean, flattened JSON without redundant ApiResponse wrappers
	// using standard encoder utility: json.NewEncoder(w).Encode(...)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(`{"error":"` + message + `"}`))
}
