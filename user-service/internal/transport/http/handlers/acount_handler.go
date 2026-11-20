package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/fouched/novabank/users/internal/core/services"
	"github.com/go-chi/chi/v5"
)

type AccountHandler struct {
	service *services.AccountService
	logger  *slog.Logger
}

func NewAccountHandler(service *services.AccountService, logger *slog.Logger) *AccountHandler {
	return &AccountHandler{
		service: service,
		logger:  logger,
	}
}

func (h *AccountHandler) GetByAccountNumber(w http.ResponseWriter, r *http.Request) {
	accNum := chi.URLParam(r, "account_number")
	if strings.TrimSpace(accNum) == "" {
		WriteError(w, http.StatusBadRequest, "Malformed JSON request")
		return
	}

	accRes, err := h.service.GetByAccountNumber(r.Context(), accNum)
	if err != nil {
		MapErrorResponse(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, accRes)
}
