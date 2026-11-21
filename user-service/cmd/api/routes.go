package main

import (
	"net/http"

	"github.com/fouched/novabank/shared/health"
	"github.com/fouched/novabank/users/internal/core/models"
	"github.com/fouched/novabank/users/internal/transport/http/handlers"
	"github.com/fouched/novabank/users/platform/config"
	"github.com/fouched/novabank/users/platform/security"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RouterDependencies encapsulates the required handler adapters cleanly
type RouterDependencies struct {
	DBPool         *pgxpool.Pool
	AuthHandler    *handlers.AuthHandler
	UserHandler    *handlers.UserHandler
	AccountHandler *handlers.AccountHandler
}

func routes(app *config.Config, deps RouterDependencies) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", health.Handler(map[string]health.CheckFunc{
		"database": deps.DBPool.Ping,
	}))

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", deps.AuthHandler.Register)
		r.Post("/login", deps.AuthHandler.Login)
	})

	r.Route("/api/users", func(r chi.Router) {
		// Protect all routes in this group with JWT validation
		r.Use(security.JWTMiddleware(app.Authority.JWTSecret, app.Authority.JWTIssuer))

		r.Get("/profile", deps.UserHandler.GetUserProfile)

		// Admin sub-router
		r.Route("/admin", func(r chi.Router) {
			r.Use(security.RequireAnyRole(models.RoleAdmin))

			r.Post("/search", deps.UserHandler.SearchUsers)
			r.Post("/all", deps.UserHandler.GetAllUsersByRole)
			r.Get("/stats", deps.UserHandler.GetUserStatistics)
			r.Patch("/toggle/{userId}", deps.UserHandler.ToggleUserStatus)
		})

	})

	r.Route("/api/accounts", func(r chi.Router) {
		// Protect all routes in this group with JWT validation
		r.Use(security.JWTMiddleware(app.Authority.JWTSecret, app.Authority.JWTIssuer))

		r.Get("/{account_number}", deps.AccountHandler.GetByAccountNumber)
	})

	return r
}
