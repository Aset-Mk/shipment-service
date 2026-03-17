package config

import (
	"fmt"
	"os"
)

// Config holds all runtime configuration for the service.
// Values are read from environment variables so the binary stays
// environment-agnostic and works equally well in Docker and locally.
type Config struct {
	GRPCAddr    string // e.g. ":50051"
	DatabaseURL string // postgres DSN
}

// Load reads configuration from environment variables.
// It returns an error if any required variable is missing.
func Load() (*Config, error) {
	cfg := &Config{
		GRPCAddr:    getEnv("GRPC_ADDR", ":50051"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
