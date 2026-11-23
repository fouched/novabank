package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	// Create a root context that cancels on SIGINT (Ctrl+C) or SIGTERM (Docker/K8s)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
	permissionRegistry := security.NewPermissionRegistry(permissionRepo)
	if err := permissionRegistry.Reload(ctx); err != nil {
		log.Fatalf("Critical error loading permissions: %v", err)
	}

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return // Exits cleanly on shutdown signal
			case <-ticker.C:
				if err := permissionRegistry.Reload(ctx); err != nil {
					logger.Error("Failed to reload permission registry", "error", err)
				}
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

	// Use an explicit http.Server instead of http.ListenAndServe
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.Port),
		Handler: r,
	}

	go func() {
		logger.Info("Starting server", "address", srv.Addr, "env", cfg.App.Env)
		// ErrServerClosed is expected when Shutdown() is called
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed unexpectedly: %v", err)
		}
	}()

	// Block here until a termination signal is received
	<-ctx.Done()
	logger.Info("Shutdown signal received, draining in-flight requests...")

	// Give active HTTP/repository requests up to 10 seconds to finish
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Forced server shutdown", "error", err)
	}

	// When main() exits here, defer publisher.Close() and defer dbPool.Close()
	// will now safely run AFTER all in-flight repository calls have finished.
	logger.Info("Server exited cleanly")
}
