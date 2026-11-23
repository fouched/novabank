package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	sharedpg "github.com/fouched/novabank/shared/postgres"
	"github.com/fouched/novabank/users/internal/core/services"
	"github.com/fouched/novabank/users/internal/transport/http/handlers"
	"github.com/fouched/novabank/users/platform/config"
	"github.com/fouched/novabank/users/platform/postgres"
	"github.com/fouched/novabank/users/platform/security"
	"github.com/fouched/toolkit/v2/logging"
)

func main() {

	// Load config
	cfg, err := config.Load(".")
	if err != nil {
		log.Fatalf("Critical error initializing config: %v", err)
	}

	// Initialise logging
	logger := logging.InitLogger(cfg.App.Env, cfg.Logging.Level)

	// Connect to Postgres
	dbPool, err := sharedpg.NewConnectionPool(cfg.Postgres)
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

	// Repositories
	userRepo := postgres.NewUserRepository(dbPool)
	accountRepo := postgres.NewAccountRepository(dbPool)
	roleRepo := postgres.NewRoleRepository(dbPool)
	permissionRepo := postgres.NewPermissionRepository(dbPool)

	// Wire repositories -> service -> handler
	authService := services.NewAuthService(
		userRepo,
		accountRepo,
		roleRepo,
		publisher,
		cfg,
		logger,
	)
	userService := services.NewUserService(
		userRepo,
		accountRepo,
		logger,
	)
	accountService := services.NewAccountService(
		accountRepo,
		logger,
	)

	// Wire permission registry
	permissionRegistry := security.NewPermissionRegistry(
		permissionRepo,
	)
	if err := permissionRegistry.Reload(context.Background()); err != nil {
		log.Fatalf(
			"Critical error loading permissions: %v",
			err,
		)
	}

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if err := permissionRegistry.Reload(context.Background()); err != nil {
				logger.Error(
					"Failed to reload permission registry",
					"error",
					err,
				)
			}
		}
	}()

	// Initialise router
	r := routes(cfg, RouterDependencies{
		DBPool:             dbPool,
		PermissionRegistry: permissionRegistry,
		AuthHandler:        handlers.NewAuthHandler(authService, logger),
		UserHandler:        handlers.NewUserHandler(userService, logger),
		AccountHandler:     handlers.NewAccountHandler(accountService, logger),
	})

	serverAddr := fmt.Sprintf(":%d", cfg.App.Port)
	logger.Info("Starting server", "address", serverAddr, "env", cfg.App.Env)
	if err := http.ListenAndServe(serverAddr, r); err != nil {
		log.Fatalf("Server shutdown abruptly: %v", err)
	}
}
