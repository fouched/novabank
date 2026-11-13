package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/fouched/novabank/users/internal/core/services"
	"github.com/fouched/novabank/users/internal/transport/http/handlers"
	"github.com/fouched/novabank/users/platform/config"
	"github.com/fouched/novabank/users/platform/postgres"
	"github.com/fouched/toolkit/v2/logging"
)

func main() {

	// Load config
	cfg, err := config.Load(".")
	if err != nil {
		log.Fatalf("Critical error initializing config: %v", err)
	}

	// Initialize logging
	logger := logging.InitLogger(cfg.App.Env, cfg.Logging.Level)

	// Connect to Postgres
	dbPool, err := postgres.NewConnectionPool(cfg.Postgres)
	if err != nil {
		log.Fatalf("Critical error connecting to database pool: %v", err)
	}
	defer dbPool.Close()
	logger.Info("Connected to database pool", "db", cfg.Postgres.DBName)

	// Kafka publisher for user registration events
	publisher, err := services.NewAccountEventPublisher(cfg.Kafka.Brokers, logger)
	if err != nil {
		log.Fatalf("Critical error initializing Kafka publisher: %v", err)
	}
	defer publisher.Close()

	// Wire repositories -> service -> handler
	authService := services.NewAuthService(
		postgres.NewUserRepository(dbPool),
		postgres.NewAccountRepository(dbPool),
		postgres.NewRoleRepository(dbPool),
		publisher,
		cfg,
		logger,
	)
	userService := services.NewUserService(
		postgres.NewUserRepository(dbPool),
		postgres.NewAccountRepository(dbPool),
		logger,
	)

	// Initialize router
	r := routes(cfg, RouterDependencies{
		AuthHandler: handlers.NewAuthHandler(authService, logger),
		UserHandler: handlers.NewUserHandler(userService, authService, logger),
	})

	serverAddr := fmt.Sprintf(":%d", cfg.App.Port)
	logger.Info("Starting server", "address", serverAddr, "env", cfg.App.Env)
	if err := http.ListenAndServe(serverAddr, r); err != nil {
		log.Fatalf("Server shutdown abruptly: %v", err)
	}
}
