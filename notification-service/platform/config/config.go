package config

import (
	"fmt"

	sharedcfg "github.com/fouched/novabank/shared/config"
)

// Duration wraps time.Duration so TOML can unmarshal values like "15m" or "24h".
// Note: this is a distinct type, so convert explicitly with time.Duration(d) when using it.
type Duration = sharedcfg.Duration

type Config struct {
	App       AppConfig       `toml:"app"`
	Postgres  PostgresConfig  `toml:"postgres"`
	Kafka     KafkaConfig     `toml:"kafka"`
	Templates TemplatesConfig `toml:"templates"`
	SMTP      SMTPConfig      `toml:"smtp"`
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

func (p PostgresConfig) BuildDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s pool_max_conns=%d pool_min_conns=%d",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SSLMode, p.MaxConns, p.MinConns)
}

type KafkaConfig struct {
	Brokers []string `toml:"brokers"`
	GroupID string   `toml:"group_id"`
}

type TemplatesConfig struct {
	Directory    string `toml:"directory"`
	CacheEnabled bool   `toml:"cache_enabled"`
}

type SMTPConfig struct {
	Host       string `toml:"host"`
	Port       int    `toml:"port"`
	Username   string `toml:"username"`
	Password   string `toml:"password"`
	Encryption string `toml:"encryption"`
	FromEmail  string `toml:"from_email"`
}

type LoggingConfig struct {
	Level string `toml:"level"`
}

func Load(dir string) (*Config, error) {
	return sharedcfg.Load[Config](dir, func(c *Config) {
		c.Postgres.User = sharedcfg.GetEnv("NOVABANK_NOTIFICATION_DB_USER", c.Postgres.User)
		c.Postgres.Password = sharedcfg.GetEnv("NOVABANK_NOTIFICATION_DB_PASSWORD", c.Postgres.Password)
	})
}
