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
	DBPool              *pgxpool.Pool
	AuthHandler         *handlers.AuthHandler
	UserHandler         *handlers.UserHandler
	UserAdminHandler    *handlers.UserAdminHandler
	AccountHandler      *handlers.AccountHandler
	AccountAdminHandler *handlers.AccountAdminHandler
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

			r.Post("/search", deps.UserAdminHandler.SearchUsers)
			r.Post("/all", deps.UserAdminHandler.GetAllUsersByRole)
			r.Get("/stats", deps.UserAdminHandler.GetUserStatistics)
			r.Patch("/toggle/{userId}", deps.UserAdminHandler.ToggleUserStatus)
		})

	})

	r.Route("/api/accounts", func(r chi.Router) {
		// Protect all routes in this group with JWT validation
		r.Use(security.JWTMiddleware(app.Authority.JWTSecret, app.Authority.JWTIssuer))

		r.Get("/", deps.AccountHandler.GetUserAccounts)
		r.Get("/{account_number}", deps.AccountHandler.GetUserAccountByNumber)

		r.Route("/admin", func(r chi.Router) {
			r.Use(security.RequireAnyRole(models.RoleAdmin))

			r.Get("/{account_number}", deps.AccountAdminHandler.GetAccountByNumber)
			r.Post("/email", deps.AccountAdminHandler.GetAccountsByEmail)
			r.Post("/all", deps.AccountAdminHandler.GetAllAccounts)
			r.Patch("/change_status", deps.AccountAdminHandler.ChangeAccountStatus)

		})
	})

	return r
}
