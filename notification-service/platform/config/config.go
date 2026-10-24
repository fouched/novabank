package config

import (
	"fmt"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// Duration wraps time.Duration so TOML can unmarshal values like "15m" or "24h".
// Note: this is a distinct type, so convert explicitly with time.Duration(d) when using it.
type Duration time.Duration

type Config struct {
	App      AppConfig      `toml:"app"`
	Postgres PostgresConfig `toml:"postgres"`
	Kafka    KafkaConfig    `toml:"kafka"`
	Logging  LoggingConfig  `toml:"logging"`
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

type KafkaConfig struct {
	Brokers []string `toml:"brokers"`
	GroupID string   `toml:"group_id"`
}

type LoggingConfig struct {
	Level string `toml:"level"`
}

func Load(path string) (*Config, error) {
	env := getEnv("NOVABANK_ENV", "dev")
	filename := fmt.Sprintf("config.%s.toml", env)

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		filename = "config.prod.toml"
	}

	data, err := os.ReadFile(fmt.Sprintf("%s/%s", path, filename))
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse toml: %w", err)
	}

	// Go standard practices dictate mapping environment variable overrides
	// so the configuration remains secure in production deployment environments
	if envPass := getEnv("NOVABANK_DB_PASSWORD", ""); envPass != "" {
		cfg.Postgres.Password = envPass
	}

	return &cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
