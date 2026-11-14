package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fouched/novabank/users/internal/core/models"
	"github.com/go-chi/chi/v5"
)

func (h *UserHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	var req models.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Malformed JSON request")
		return
	}

	user, err := h.userService.Search(r.Context(), req.Email, req.AccountNumber)
	if err != nil {
		MapErrorResponse(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, user)
}

func (h *UserHandler) GetAllUsersByRole(w http.ResponseWriter, r *http.Request) {
	var req models.UsersByRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Malformed JSON request")
		return
	}

	users, err := h.userService.GetAllUsersByRole(r.Context(), req.RoleName, req.PageSize, req.Page)
	if err != nil {
		MapErrorResponse(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, users)
}

func (h *UserHandler) GetUserStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.userService.GetUserStats(r.Context())
	if err != nil {
		MapErrorResponse(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, stats)
}

func (h *UserHandler) ToggleUserStatus(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		MapErrorResponse(w, err)
		return
	}

	enabled, err := h.userService.ToggleUserStatus(r.Context(), userId)
	if err != nil {
		MapErrorResponse(w, err)
		return
	}

	toggleResponse := models.ToggleResponse{
		UserID:  userId,
		Enabled: enabled,
	}

	WriteJSON(w, http.StatusOK, toggleResponse)

}
