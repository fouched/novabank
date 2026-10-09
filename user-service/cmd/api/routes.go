package main

import (
	"net/http"

	"github.com/fouched/novabank/users/internal/transport/http/handlers"
	"github.com/fouched/novabank/users/platform/config"
	"github.com/go-chi/chi/v5"
)

func routes(app *config.Config) http.Handler {
	r := chi.NewRouter()

	// Health endpoint mirroring spring boot actuator
	r.Get("/health", handlers.HealthHandler)

	return r
}
