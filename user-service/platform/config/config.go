package config

import (
	sharedcfg "github.com/fouched/novabank/shared/config"
)

// Duration wraps time.Duration so TOML can unmarshal values like "15m" or "24h".
// Note: this is a distinct type, so convert explicitly with time.Duration(d) when using it.
type Duration = sharedcfg.Duration

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

func Load(dir string) (*Config, error) {
	return sharedcfg.Load[Config](dir, func(c *Config) {
		c.Postgres.User = sharedcfg.GetEnv("NOVABANK_USER_DB_USER", c.Postgres.User)
		c.Postgres.Password = sharedcfg.GetEnv("NOVABANK_USER_DB_PASSWORD", c.Postgres.Password)
		c.Authority.JWTSecret = sharedcfg.GetEnv("NOVABANK_USER_JWT_SECRET", c.Authority.JWTSecret)
	})
}
