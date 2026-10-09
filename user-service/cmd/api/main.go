package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

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
	logger := initLogging(cfg)

	// Connect to Postgres
	dbPool, err := postgres.NewConnectionPool(cfg.Postgres)
	if err != nil {
		log.Fatalf("Critical error connecting to database pool: %v", err)
	}
	defer dbPool.Close()
	logger.Info("Connected to database pool", "db", cfg.Postgres.DBName)

	// Initialize router
	r := routes(cfg)

	serverAddr := fmt.Sprintf(":%d", cfg.App.Port)
	logger.Info("Starting server", "address", serverAddr, "env", cfg.App.Env)
	if err := http.ListenAndServe(serverAddr, r); err != nil {
		log.Fatalf("Server shutdown abruptly: %v", err)
	}
}

func initLogging(cfg *config.Config) *slog.Logger {
	var encoder slog.Handler
	isDevelopment := cfg.App.Env == "dev"
	if isDevelopment {
		encoder = logging.NewPrettyDevHandler()
	} else {
		lvl, err := logging.ParseLevel(cfg.Logging.Level)
		if err != nil {
			// Fallback to INFO but log the issue
			fmt.Printf("Invalid log level %q, defaulting to INFO\n", cfg.Logging.Level)
			lvl = slog.LevelInfo
		}

		logging.ProdLevel.Set(lvl)
		jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logging.ProdLevel, // <-- dynamic!
		})
		encoder = &logging.ProdHandler{
			Handler: jsonHandler,
		}
	}

	// Create the core base logger
	baseLogger := slog.New(encoder)

	// Optional: Set as global just in case external libraries rely on it
	slog.SetDefault(baseLogger)
	return baseLogger
}
