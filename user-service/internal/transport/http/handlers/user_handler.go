package handlers

import (
	"log/slog"
	"net/http"

	"github.com/fouched/novabank/users/internal/core/services"
	"github.com/fouched/novabank/users/platform/security"
)

type UserHandler struct {
	userService *services.UserService
	authService *services.AuthService
	logger      *slog.Logger
}

func NewUserHandler(userService *services.UserService, authService *services.AuthService, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		authService: authService,
		logger:      logger,
	}
}

func (h *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := security.GetClaims(r.Context())
	if !ok {
		h.logger.Warn("missing jwt claims in request context")
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userService.GetUserAccountDetails(r.Context(), claims.Email)
	if err != nil {
		h.logger.Error("error getting user profile", "error", err)
		MapErrorResponse(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, user)
}
