package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fouched/novabank/users/internal/core/errdefs"
)

func MapErrorResponse(w http.ResponseWriter, err error) {
	// 1. Check if a custom ApplicationError with a specific user-facing message was returned
	var appErr *errdefs.ApplicationError
	if errors.As(err, &appErr) {
		switch {
		case errors.Is(appErr, errdefs.ErrNotFound):
			WriteError(w, http.StatusNotFound, appErr.Message)
		case errors.Is(appErr, errdefs.ErrUnauthorized):
			WriteError(w, http.StatusUnauthorized, appErr.Message)
		default:
			WriteError(w, http.StatusBadRequest, appErr.Message)
		}
		return
	}

	// 2. Fallback: check for bare sentinel errors returned directly (e.g., from repositories)
	switch {
	case errors.Is(err, errdefs.ErrNotFound):
		WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, errdefs.ErrUnauthorized):
		WriteError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, errdefs.ErrBadRequest):
		WriteError(w, http.StatusBadRequest, err.Error())
	default:
		WriteError(w, http.StatusInternalServerError, "An unexpected internal banking system exception occurred")
	}
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, ErrorResponse{Error: msg})
}
