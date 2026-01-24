// Package config provides application configuration management,
// loading settings from environment variables and .env files.
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Service configuration
	ServiceName string
	Environment string
	GRPCPort    string
	HTTPPort    string

	// Database configuration
	PostgresHost     string
	PostgresPort     string
	PostgresDB       string
	PostgresUser     string
	PostgresPassword string

	// Home location coordinates
	HomeLatitude  float64
	HomeLongitude float64

	// Distance thresholds
	AwayThresholdKM float64

	// CSV output path
	CSVOutputPath string

	// OpenTelemetry configuration
	OTELEndpoint string

	// Logging
	LogLevel string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (ignore error if it doesn't exist)
	_ = godotenv.Load()

	cfg := &Config{
		ServiceName: getEnv("SERVICE_NAME", "otel-worker"),
		Environment: getEnv("ENVIRONMENT", "development"),
		GRPCPort:    getEnv("GRPC_PORT", "50051"),
		HTTPPort:    getEnv("HTTP_PORT", "8080"),

		PostgresHost:     getEnv("POSTGRES_HOST", "192.168.1.175"),
		PostgresPort:     getEnv("POSTGRES_PORT", "6432"),
		PostgresDB:       getEnv("POSTGRES_DB", "owntracks"),
		PostgresUser:     getEnv("POSTGRES_USER", "development"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "development"),

		CSVOutputPath: getEnv("CSV_OUTPUT_PATH", "/data/csv"),
		OTELEndpoint:  getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
	}

	// Parse float values
	var err error
	cfg.HomeLatitude, err = parseFloat("HOME_LATITUDE", "40.736097")
	if err != nil {
		return nil, fmt.Errorf("invalid HOME_LATITUDE: %w", err)
	}

	cfg.HomeLongitude, err = parseFloat("HOME_LONGITUDE", "-74.039373")
	if err != nil {
		return nil, fmt.Errorf("invalid HOME_LONGITUDE: %w", err)
	}

	cfg.AwayThresholdKM, err = parseFloat("AWAY_THRESHOLD_KM", "0.5")
	if err != nil {
		return nil, fmt.Errorf("invalid AWAY_THRESHOLD_KM: %w", err)
	}

	return cfg, nil
}

// DatabaseDSN returns the PostgreSQL connection string
func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresDB,
		c.PostgresUser,
		c.PostgresPassword,
	)
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseFloat parses a float64 from an environment variable or default value
func parseFloat(key, defaultValue string) (float64, error) {
	value := getEnv(key, defaultValue)
	return strconv.ParseFloat(value, 64)
}
