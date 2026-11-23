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
	DBPool             *pgxpool.Pool
	AuthHandler        *handlers.AuthHandler
	UserHandler        *handlers.UserHandler
	AccountHandler     *handlers.AccountHandler
	PermissionRegistry *security.PermissionRegistry
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

		r.Use(security.JWTMiddleware(app.Authority.JWTSecret, app.Authority.JWTIssuer))

		r.With(
			security.RequireAnyPermission(
				deps.PermissionRegistry,
				models.UserReadOwn,
				models.AccountReadOwn,
			),
		).Get(
			"/profile",
			deps.UserHandler.GetUserProfile,
		)

		r.With(
			security.RequireAnyPermission(
				deps.PermissionRegistry,
				models.UserSearch,
			),
		).Post(
			"/search",
			deps.UserHandler.SearchUsers,
		)

		r.With(
			security.RequireAnyPermission(
				deps.PermissionRegistry,
				models.UserReadAll,
			),
		).Post(
			"/all",
			deps.UserHandler.GetAllUsersByRole,
		)

		r.With(
			security.RequireAnyPermission(
				deps.PermissionRegistry,
				models.UserStats,
			),
		).Get(
			"/stats",
			deps.UserHandler.GetUserStatistics,
		)

		r.With(
			security.RequireAnyPermission(
				deps.PermissionRegistry,
				models.UserUpdate,
			),
		).Patch(
			"/toggle/{userId}",
			deps.UserHandler.ToggleUserStatus,
		)
	})

	r.Route("/api/accounts", func(r chi.Router) {

		r.Use(security.JWTMiddleware(app.Authority.JWTSecret, app.Authority.JWTIssuer))

		r.Get("/", deps.AccountHandler.GetAllUserAccounts)
		r.Get("/{account_number}", deps.AccountHandler.GetUserAccount)

		r.With(
			security.RequireAnyPermission(
				deps.PermissionRegistry,
				models.AccountSearch,
			),
		).Post(
			"/search",
			deps.AccountHandler.SearchAccounts,
		)

		r.With(
			security.RequireAnyPermission(
				deps.PermissionRegistry,
				models.AccountReadAll,
			),
		).Post(
			"/all",
			deps.AccountHandler.GetAllAccounts,
		)

		r.With(
			security.RequireAnyPermission(
				deps.PermissionRegistry,
				models.AccountUpdate,
			),
		).Post(
			"/change_status",
			deps.AccountHandler.ChangeAccountStatus,
		)
	})

	return r
}
