package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fouched/novabank/notifications/internal/core/services"
	"github.com/fouched/novabank/notifications/platform/config"
	"github.com/fouched/novabank/notifications/platform/mail"
	"github.com/fouched/novabank/notifications/platform/postgres"
	"github.com/fouched/novabank/notifications/platform/templates"
	"github.com/fouched/novabank/shared/health"
	sharedpg "github.com/fouched/novabank/shared/postgres"
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
	logger.Info("Starting NovaBank notification-service background daemon processes...")

	// Connect to Postgres
	dbPool, err := sharedpg.NewConnectionPool(cfg.Postgres)
	if err != nil {
		log.Fatalf("Critical error connecting to database pool: %v", err)
	}
	defer dbPool.Close()
	logger.Info("Connected to database pool", "db", cfg.Postgres.DBName)

	repo := postgres.NewNotificationRepository(dbPool)

	// template configuration
	tmplEngine, err := templates.NewTemplateEngine(cfg.Templates)
	if err != nil {
		logger.Error("Template engine cache pre-compilation routine failed", slog.Any("error", err))
		os.Exit(1)
	}

	// mail configuration
	mailer, err := mail.NewSMTPClient(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.Encryption,
		cfg.SMTP.FromEmail,
	)
	if err != nil {
		logger.Error("Failed to bind SMTP infrastructure client components", slog.Any("error", err))
		os.Exit(1)
	}

	// Define connectivity configurations
	brokers := cfg.Kafka.Brokers
	groupID := cfg.Kafka.GroupID

	// Instantiate the engine orchestration segments
	worker, err := services.NewNotificationWorker(brokers, groupID, repo, tmplEngine, mailer, logger)
	if err != nil {
		logger.Error("Critical initialization error blocking worker bootstrap execution", slog.Any("error", err))
		os.Exit(1)
	}

	// Hook up elegant OS cancellation captures (Sigterm / Sigint) to handle smooth shutdowns
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// health endpoint
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health.Handler(map[string]health.CheckFunc{
		"database": dbPool.Ping,
		"kafka":    worker.PingKafka,
	}))

	healthServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("Starting health check server", slog.String("addr", healthServer.Addr))
		if err := healthServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Health check server error", slog.Any("error", err))
		}
	}()

	// Fire up the endless processing thread space
	go worker.Start(ctx)

	// Block main runtime path execution thread until OS termination signal triggers
	<-ctx.Done()
	logger.Info("Termination signal intercept confirmation triggered. Cleaning up runtime context allocations...")

	worker.Close()
	logger.Info("Worker daemon completely halted cleanly. Exiting system scope safely.")
}
