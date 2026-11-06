package main

import (
	"net/http"

	"github.com/fouched/novabank/users/internal/transport/http/handlers"
	"github.com/fouched/novabank/users/platform/config"
	"github.com/go-chi/chi/v5"
)

// RouterDependencies encapsulates the required handler adapters cleanly
type RouterDependencies struct {
	AuthHandler *handlers.AuthHandler
}

func routes(app *config.Config, deps RouterDependencies) http.Handler {
	r := chi.NewRouter()

	// Health endpoint mirroring spring boot actuator
	r.Get("/health", handlers.HealthHandler)

	// Mirrors @RequestMapping("/api/auth") on the Spring AuthController
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", deps.AuthHandler.Register)
		r.Post("/login", deps.AuthHandler.Login)
	})

	return r
}
