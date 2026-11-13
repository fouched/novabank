package main

import (
	"net/http"

	"github.com/fouched/novabank/users/internal/transport/http/handlers"
	"github.com/fouched/novabank/users/platform/config"
	"github.com/fouched/novabank/users/platform/security"
	"github.com/go-chi/chi/v5"
)

// RouterDependencies encapsulates the required handler adapters cleanly
type RouterDependencies struct {
	AuthHandler *handlers.AuthHandler
	UserHandler *handlers.UserHandler
}

func routes(app *config.Config, deps RouterDependencies) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", handlers.HealthHandler)

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", deps.AuthHandler.Register)
		r.Post("/login", deps.AuthHandler.Login)
	})

	r.Route("/api/users", func(r chi.Router) {
		// Protect all routes in this group with JWT validation
		r.Use(security.JWTMiddleware(app.Authority.JWTSecret, app.Authority.JWTIssuer))
		r.Get("/profile", deps.UserHandler.GetUserProfile)
	})

	return r
}
