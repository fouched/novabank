package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// Duration wraps time.Duration so TOML can unmarshal values like "15m" or "24h".
// Note: this is a distinct type, so convert explicitly with time.Duration(d) when using it.
type Duration time.Duration

type Config struct {
	App       AppConfig       `toml:"app"`
	Postgres  PostgresConfig  `toml:"postgres"`
	Consul    ConsulConfig    `toml:"consul"`
	Kafka     KafkaConfig     `toml:"kafka"`
	Authority AuthorityConfig `toml:"authority"`
	Logging   LoggingConfig   `toml:"logging"`
}

type AppConfig struct {
	Name string `toml:"name"`
	Env  string `toml:"env"`
	Port int    `toml:"port"`
}

type PostgresConfig struct {
	Host            string   `toml:"host"`
	Port            int      `toml:"port"`
	User            string   `toml:"user"`
	Password        string   `toml:"password"`
	DBName          string   `toml:"dbname"`
	SSLMode         string   `toml:"sslmode"`
	MaxConns        int32    `toml:"max_conns"`
	MinConns        int32    `toml:"min_conns"`
	MaxConnLifetime Duration `toml:"max_conn_lifetime"`
	MaxConnIdleTime Duration `toml:"max_conn_idle_time"`
}

type ConsulConfig struct {
	Address     string `toml:"address"`
	ServiceName string `toml:"service_name"`
}

type KafkaConfig struct {
	Brokers []string `toml:"brokers"`
	GroupID string   `toml:"group_id"`
}

type AuthorityConfig struct {
	JWTIssuer     string   `toml:"jwt_issuer"`
	JWTAudience   string   `toml:"jwt_audience"`
	JWTSecret     string   `toml:"jwt_secret"`
	TokenExpiry   Duration `toml:"token_expiry"`
	RefreshExpiry Duration `toml:"refresh_expiry"`
}

type LoggingConfig struct {
	Level string `toml:"level"`
}

func (d *Duration) UnmarshalText(b []byte) error {
	parsed, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}

// Load reads config.<env>.toml from dir, merges an optional config.<env>.private.toml
// on top of it, and returns the populated config struct T.
func Load[T any](dir string) (*T, error) {
	env := getEnv("NOVABANK_ENV", "dev")
	baseFile := filepath.Join(dir, fmt.Sprintf("config.%s.toml", env))

	if _, err := os.Stat(baseFile); errors.Is(err, os.ErrNotExist) {
		baseFile = filepath.Join(dir, "config.prod.toml")
	}

	var cfg T
	if err := unmarshalFile(baseFile, &cfg); err != nil {
		return nil, fmt.Errorf("load base config: %w", err)
	}

	// Merge optional private file (e.g., config.dev.private.toml) if it exists
	privateFile := filepath.Join(dir, fmt.Sprintf("config.%s.private.toml", env))
	if _, err := os.Stat(privateFile); err == nil {
		if err := unmarshalFile(privateFile, &cfg); err != nil {
			return nil, fmt.Errorf("load private config: %w", err)
		}
	}

	return &cfg, nil
}

func unmarshalFile(filePath string, target any) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}
	if err := toml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("parse %s: %w", filePath, err)
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
