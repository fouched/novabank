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
type Duration time.Duration

func (d *Duration) UnmarshalText(b []byte) error {
	parsed, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}

// Load reads config.<env>.toml from dir, merges an optional config.<env>.private.toml
// on top of it, and applies any optional override functions (e.g., environment variables).
func Load[T any](dir string, overrides ...func(*T)) (*T, error) {
	env := GetEnv("NOVABANK_ENV", "dev")
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
