package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/fouched/novabank/users/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewConnectionPool instantiates a thread-safe connection pool for pgx/v5
func NewConnectionPool(cfg config.PostgresConfig) (*pgxpool.Pool, error) {
	// Construct standard connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	// Parse configuration string into pgx pool config structure
	poolCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	// Apply robust scaling constraints
	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = time.Duration(cfg.MaxConnLifetime)
	poolCfg.MaxConnIdleTime = time.Duration(cfg.MaxConnIdleTime)

	// Establish connection timeout threshold
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	// Ping connection pool to guarantee communication channel is open
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
