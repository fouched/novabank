package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/fouched/novabank/notifications/internal/core/services"
	"github.com/fouched/novabank/users/platform/config"
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

	// 2. Define connectivity configurations (Will shift to TOML structure mapping next)
	brokers := []string{"localhost:9092"}
	groupID := "notification-group"

	// 3. Instantiate the engine orchestration segments
	worker, err := services.NewNotificationWorker(brokers, groupID, logger)
	if err != nil {
		logger.Error("Critical initialization error blocking worker bootstrap execution", slog.Any("error", err))
		os.Exit(1)
	}

	// 4. Hook up elegant OS cancellation captures (Sigterm / Sigint) to handle smooth shutdowns
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// 5. Fire up the endless processing thread space
	go worker.Start(ctx)

	// Block main runtime path execution thread until OS termination signal triggers
	<-ctx.Done()
	logger.Info("Termination signal intercept confirmation triggered. Cleaning up runtime context allocations...")

	worker.Close()
	logger.Info("Worker daemon completely halted cleanly. Exiting system scope safely.")
}
