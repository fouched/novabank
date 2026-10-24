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
