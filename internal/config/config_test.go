package config

import (
	"os"
	"testing"
)

// nolint:gocyclo // Test function complexity from multiple subtests and assertions
func TestLoad(t *testing.T) {
	// Save original env vars
	originalEnv := make(map[string]string)
	envVars := []string{
		"SERVICE_NAME", "ENVIRONMENT", "GRPC_PORT", "HTTP_PORT",
		"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_DB",
		"HOME_LATITUDE", "HOME_LONGITUDE", "AWAY_THRESHOLD_KM",
	}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
	}

	// Clean up after test
	defer func() {
		for key, val := range originalEnv {
			if val != "" {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	t.Run("loads default values", func(t *testing.T) {
		// Clear env vars
		for _, key := range envVars {
			os.Unsetenv(key)
		}

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if cfg.ServiceName != "otel-worker" {
			t.Errorf("expected ServiceName 'otel-worker', got '%s'", cfg.ServiceName)
		}
		if cfg.GRPCPort != "50051" {
			t.Errorf("expected GRPCPort '50051', got '%s'", cfg.GRPCPort)
		}
		if cfg.PostgresPort != "6432" {
			t.Errorf("expected PostgresPort '6432' (PgBouncer), got '%s'", cfg.PostgresPort)
		}
		if cfg.HomeLatitude != 40.736097 {
			t.Errorf("expected HomeLatitude 40.736097, got %f", cfg.HomeLatitude)
		}
		if cfg.HomeLongitude != -74.039373 {
			t.Errorf("expected HomeLongitude -74.039373, got %f", cfg.HomeLongitude)
		}
	})

	t.Run("loads custom values from environment", func(t *testing.T) {
		os.Setenv("SERVICE_NAME", "test-service")
		os.Setenv("GRPC_PORT", "9999")
		os.Setenv("POSTGRES_PORT", "5432")
		os.Setenv("HOME_LATITUDE", "42.0")
		os.Setenv("HOME_LONGITUDE", "-73.0")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if cfg.ServiceName != "test-service" {
			t.Errorf("expected ServiceName 'test-service', got '%s'", cfg.ServiceName)
		}
		if cfg.GRPCPort != "9999" {
			t.Errorf("expected GRPCPort '9999', got '%s'", cfg.GRPCPort)
		}
		if cfg.PostgresPort != "5432" {
			t.Errorf("expected PostgresPort '5432', got '%s'", cfg.PostgresPort)
		}
		if cfg.HomeLatitude != 42.0 {
			t.Errorf("expected HomeLatitude 42.0, got %f", cfg.HomeLatitude)
		}
		if cfg.HomeLongitude != -73.0 {
			t.Errorf("expected HomeLongitude -73.0, got %f", cfg.HomeLongitude)
		}
	})

	t.Run("returns error for invalid float", func(t *testing.T) {
		os.Setenv("HOME_LATITUDE", "invalid")

		_, err := Load()
		if err == nil {
			t.Error("expected error for invalid HOME_LATITUDE, got nil")
		}
	})
}

func TestDatabaseDSN(t *testing.T) {
	cfg := &Config{
		PostgresHost:     "192.168.1.175",
		PostgresPort:     "6432",
		PostgresDB:       "owntracks",
		PostgresUser:     "testuser",
		PostgresPassword: "testpass",
	}

	expected := "host=192.168.1.175 port=6432 dbname=owntracks user=testuser password=testpass sslmode=disable"
	if dsn := cfg.DatabaseDSN(); dsn != expected {
		t.Errorf("expected DSN '%s', got '%s'", expected, dsn)
	}
}
