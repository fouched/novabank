package postgres

import (
	"context"
	"fmt"
	"time"

	sharedcfg "github.com/fouched/novabank/shared/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds the standard PostgreSQL connection and pool settings across NovaBank services.
type Config struct {
	Host            string             `toml:"host"`
	Port            int                `toml:"port"`
	User            string             `toml:"user"`
	Password        string             `toml:"password"`
	DBName          string             `toml:"dbname"`
	SSLMode         string             `toml:"sslmode"`
	MaxConns        int32              `toml:"max_conns"`
	MinConns        int32              `toml:"min_conns"`
	MaxConnLifetime sharedcfg.Duration `toml:"max_conn_lifetime"`
	MaxConnIdleTime sharedcfg.Duration `toml:"max_conn_idle_time"`
}

// NewConnectionPool instantiates a thread-safe connection pool for pgx/v5 and verifies connectivity.
func NewConnectionPool(cfg Config) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	poolCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = time.Duration(cfg.MaxConnLifetime)
	poolCfg.MaxConnIdleTime = time.Duration(cfg.MaxConnIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
