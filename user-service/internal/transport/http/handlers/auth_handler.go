package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fouched/novabank/users/internal/core/faults"
	"github.com/fouched/novabank/users/internal/core/models"
	"github.com/fouched/novabank/users/internal/core/services"
)

type AuthHandler struct {
	service *services.AuthService
}

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Malformed JSON request message body payload")
		return
	}

	registrationResponse, err := h.service.Register(r.Context(), req)
	if err != nil {
		h.mapErrorResponse(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, registrationResponse)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Malformed JSON request message body payload")
		return
	}

	loginResponse, err := h.service.Login(r.Context(), req)
	if err != nil {
		h.mapErrorResponse(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, loginResponse)
}

func (h *AuthHandler) mapErrorResponse(w http.ResponseWriter, err error) {
	var appErr *faults.ApplicationError
	if errors.As(err, &appErr) {
		if errors.Is(appErr.Unwrap(), faults.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, appErr.Message)
			return
		}
		if errors.Is(appErr.Unwrap(), faults.ErrUnauthorized) {
			h.writeError(w, http.StatusUnauthorized, appErr.Message)
			return
		}
		h.writeError(w, http.StatusBadRequest, appErr.Message)
		return
	}
	h.writeError(w, http.StatusInternalServerError, "An unexpected distributed internal banking system exception occurred")
}

func (h *AuthHandler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *AuthHandler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, ErrorResponse{Error: msg})
}
