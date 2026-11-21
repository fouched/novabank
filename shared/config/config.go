package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// Duration wraps time.Duration so TOML can unmarshal values like "15m" or "24h".
type Duration time.Duration

func (d *Duration) UnmarshalText(b []byte) error {
	parsed, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}

// Load reads configuration from dir using a three-tier precedence model:
//  1. Base file: config.<env>.toml (committed to Git with non-sensitive defaults).
//  2. Private override: config.<env>.private.toml (optional, git-ignored file for
//     local secrets like DB credentials and JWT keys—never committed to Git).
//  3. Runtime overrides: optional functions applied last (typically reading OS
//     environment variables in deployed environments).
//
// Defaults to "prod" when NOVABANK_ENV is unset so production deployments never
// accidentally load development settings. Set NOVABANK_ENV=dev locally.
func Load[T any](dir string, overrides ...func(*T)) (*T, error) {
	env := GetEnv("NOVABANK_ENV", "prod")
	baseFile := filepath.Join(dir, fmt.Sprintf("config.%s.toml", env))

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

	for _, apply := range overrides {
		if apply != nil {
			apply(&cfg)
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

// GetEnv returns the value of the environment variable named by key,
// or defaultValue if the variable is not present or empty.
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
