package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/fouched/novabank/users/internal/core/models"
	"github.com/fouched/novabank/users/internal/core/services"
)

type AuthHandler struct {
	service *services.AuthService
	logger  *slog.Logger
}

func NewAuthHandler(service *services.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		service: service,
		logger:  logger,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Malformed JSON request message body payload")
		return
	}

	registrationResponse, err := h.service.Register(r.Context(), req)
	if err != nil {
		h.logger.Error("error during registration", "error", err)
		MapErrorResponse(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, registrationResponse)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Malformed JSON request")
		return
	}

	loginResponse, err := h.service.Login(r.Context(), req)
	if err != nil {
		MapErrorResponse(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, loginResponse)
}
