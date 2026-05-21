package config

import (
	"fmt"
	"log/slog"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config holds all configuration for the application.
type Config struct {
	App struct {
		Host string `yaml:"host" env:"APP_HOST" env-default:"0.0.0.0"`
		Port int    `yaml:"port" env:"APP_PORT" env-default:"8080"`
	} `yaml:"app"`
	DB struct {
		Host     string `yaml:"host"     env:"DB_HOST"     env-default:"localhost"`
		Port     int    `yaml:"port"     env:"DB_PORT"     env-default:"5432"`
		User     string `yaml:"user"     env:"DB_USER"     env-default:"postgres"`
		Password string `yaml:"password" env:"DB_PASSWORD" env-default:"postgres"`
		Name     string `yaml:"name"     env:"DB_NAME"     env-default:"subscriptions"`
		SSLMode  string `yaml:"sslmode"  env:"DB_SSLMODE"  env-default:"disable"`
	} `yaml:"db"`
}

// DSN returns the PostgreSQL connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DB.User, c.DB.Password, c.DB.Host, c.DB.Port, c.DB.Name, c.DB.SSLMode,
	)
}

// Load reads configuration from .env file (if present) and environment variables.
func Load() (*Config, error) {
	var cfg Config

	// Try to read .env file first; if it doesn't exist, cleanenv will fall back to env vars.
	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		// .env file might not exist; try reading env vars only.
		slog.Warn("could not read .env file, falling back to environment variables", "error", err)
		if envErr := cleanenv.ReadEnv(&cfg); envErr != nil {
			return nil, fmt.Errorf("failed to load config: %w", envErr)
		}
	}

	return &cfg, nil
}
