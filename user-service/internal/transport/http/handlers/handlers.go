package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fouched/novabank/users/internal/core/errdefs"
)

func MapErrorResponse(w http.ResponseWriter, err error) {
	var appErr *errdefs.ApplicationError
	if errors.As(err, &appErr) {
		if errors.Is(appErr.Unwrap(), errdefs.ErrNotFound) {
			WriteError(w, http.StatusNotFound, appErr.Message)
			return
		}
		if errors.Is(appErr.Unwrap(), errdefs.ErrUnauthorized) {
			WriteError(w, http.StatusUnauthorized, appErr.Message)
			return
		}
		WriteError(w, http.StatusBadRequest, appErr.Message)
		return
	}
	WriteError(w, http.StatusInternalServerError, "An unexpected distributed internal banking system exception occurred")
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, ErrorResponse{Error: msg})
}
