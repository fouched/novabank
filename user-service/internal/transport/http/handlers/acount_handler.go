package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/fouched/novabank/users/internal/core/models"
	"github.com/fouched/novabank/users/internal/core/services"
	"github.com/fouched/novabank/users/platform/security"
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

func (h *AccountHandler) GetUserAccount(w http.ResponseWriter, r *http.Request) {
	accNum := chi.URLParam(r, "account_number")
	if strings.TrimSpace(accNum) == "" {
		WriteError(w, http.StatusBadRequest, "Malformed JSON request")
		return
	}

	claims, ok := security.GetClaims(r.Context())
	if !ok {
		h.logger.Warn("missing jwt claims in request context")
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	acc, err := h.service.GetUserAccount(r.Context(), accNum, claims.Email)
	if err != nil {
		MapErrorResponse(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, acc)
}

func (h *AccountHandler) GetAllUserAccounts(w http.ResponseWriter, r *http.Request) {
	claims, ok := security.GetClaims(r.Context())
	if !ok {
		h.logger.Warn("missing jwt claims in request context")
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accounts, err := h.service.GetByEmail(r.Context(), claims.Email)
	if err != nil {
		MapErrorResponse(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, accounts)
}

func (h *AccountHandler) GetAccountsByEmail(w http.ResponseWriter, r *http.Request) {
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

func (h *AccountHandler) GetAllAccounts(w http.ResponseWriter, r *http.Request) {
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

func (h *AccountHandler) ChangeAccountStatus(w http.ResponseWriter, r *http.Request) {
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

func (h *AccountHandler) SearchAccounts(w http.ResponseWriter, r *http.Request) {
	var req models.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Malformed JSON request")
		return
	}

	accounts, err := h.service.Search(r.Context(), req.AccountNumber, req.Email)
	if err != nil {
		MapErrorResponse(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, accounts)
}
