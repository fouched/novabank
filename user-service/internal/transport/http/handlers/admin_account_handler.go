package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/fouched/novabank/users/internal/core/models"
	"github.com/fouched/novabank/users/internal/core/services"
	"github.com/go-chi/chi/v5"
)

type AccountAdminHandler struct {
	service *services.AccountService
	logger  *slog.Logger
}

func NewAdminAccountHandler(service *services.AccountService, logger *slog.Logger) *AccountAdminHandler {
	return &AccountAdminHandler{
		service: service,
		logger:  logger,
	}
}

func (h *AccountAdminHandler) GetAccountByNumber(w http.ResponseWriter, r *http.Request) {
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

func (h *AccountAdminHandler) GetAccountsByEmail(w http.ResponseWriter, r *http.Request) {
	var req models.AccountsByEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Malformed JSON request")
		return
	}

	accounts, err := h.service.GetByEmail(r.Context(), req.Email)
	if err != nil {
		MapErrorResponse(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, accounts)
}

func (h *AccountAdminHandler) GetAllAccounts(w http.ResponseWriter, r *http.Request) {
	var req models.AccountsAllRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Malformed JSON request")
		return
	}

	accounts, err := h.service.GetAll(r.Context(), req.PageSize, req.Page)
	if err != nil {
		MapErrorResponse(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, accounts)
}

func (h *AccountAdminHandler) ChangeAccountStatus(w http.ResponseWriter, r *http.Request) {
	var req models.AccountChangeStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Malformed JSON request")
		return
	}

	success, err := h.service.ChangeAccountStatus(r.Context(), req.AccountNumber, req.AccountStatus)
	if err != nil {
		MapErrorResponse(w, err)
		return
	}

	res := models.AccountChangeStatusResponse{
		AccountNumber: req.AccountNumber,
		AccountStatus: req.AccountStatus,
		Result:        false,
	}
	if success {
		res.Result = true
	}

	WriteJSON(w, http.StatusOK, res)
}
