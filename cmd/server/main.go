// Package main is the entry point for the otel-worker gRPC server,
// providing distance calculation services with PostgreSQL integration.
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/stuartshay/otel-worker/internal/config"
	"github.com/stuartshay/otel-worker/internal/database"
	grpcserver "github.com/stuartshay/otel-worker/internal/grpc"
	"github.com/stuartshay/otel-worker/internal/tracing"
	distancev1 "github.com/stuartshay/otel-worker/proto/distance/v1"
)

func main() { // nolint:gocyclo // Main function complexity is acceptable for server initialization
	// Initialize structured logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	log.Info().Msg("Starting otel-worker service")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Set log level
	setLogLevel(cfg.LogLevel)

	log.Info().
		Str("service_name", cfg.ServiceName).
		Str("environment", cfg.Environment).
		Str("grpc_port", cfg.GRPCPort).
		Str("db_host", cfg.PostgresHost).
		Str("db_port", cfg.PostgresPort).
		Float64("home_lat", cfg.HomeLatitude).
		Float64("home_lon", cfg.HomeLongitude).
		Msg("Configuration loaded")

	// Initialize OpenTelemetry tracing
	var shutdownTracer func(context.Context) error
	if cfg.OTELEnabled {
		st, err := tracing.InitTracer(tracing.Config{
			ServiceName:      cfg.ServiceName,
			ServiceNamespace: cfg.ServiceNamespace,
			ServiceVersion:   cfg.ServiceVersion,
			Environment:      cfg.Environment,
			OTLPEndpoint:     cfg.OTELEndpoint,
			Enabled:          cfg.OTELEnabled,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to initialize OpenTelemetry tracing")
		} else {
			shutdownTracer = st
			log.Info().
				Str("endpoint", cfg.OTELEndpoint).
				Msg("OpenTelemetry tracing initialized")
		}
	} else {
		log.Info().Msg("OpenTelemetry tracing disabled")
	}

	// Defer tracer shutdown if initialized
	defer func() {
		if shutdownTracer != nil {
			if shutdownErr := shutdownTracer(context.Background()); shutdownErr != nil {
				log.Error().Err(shutdownErr).Msg("Failed to shutdown OpenTelemetry tracer")
			}
		}
	}()

	// Initialize database client
	dbClient, err := database.NewClient(cfg.DatabaseDSN())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database client") // nolint:gocritic // Fatal error before service starts is acceptable
	}
	defer func() {
		if closeErr := dbClient.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("Failed to close database connection")
		}
	}()

	log.Info().Msg("Database connection established")

	// Verify database connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	if healthErr := dbClient.HealthCheck(ctx); healthErr != nil {
		cancel()
		log.Error().Err(healthErr).Msg("Database health check failed")
		os.Exit(1) // nolint:gocritic // Immediate exit on health check failure is intentional
	}
	cancel()

	log.Info().Msg("Database health check passed")

	// Initialize gRPC server with OpenTelemetry interceptors
	var serverOpts []grpc.ServerOption
	if cfg.OTELEnabled {
		serverOpts = append(serverOpts,
			grpc.StatsHandler(otelgrpc.NewServerHandler()),
		)
	}
	grpcServer := grpc.NewServer(serverOpts...)

	// Register distance service
	distanceServer := grpcserver.NewServer(cfg, dbClient)
	distancev1.RegisterDistanceServiceServer(grpcServer, distanceServer)

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable server reflection for debugging
	reflection.Register(grpcServer)

	// Start gRPC server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create TCP listener")
	}

	go func() {
		log.Info().Str("port", cfg.GRPCPort).Msg("gRPC server listening")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal().Err(err).Msg("gRPC server failed")
		}
	}()

	// Start HTTP health check server
	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.HTTPPort),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Health check handlers
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		// Liveness probe - always returns 200 if service is running
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","service":"otel-worker"}`)) //nolint:errcheck // HTTP response write failure is not recoverable
	})

	http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// Readiness probe - checks database connectivity
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if err := dbClient.HealthCheck(ctx); err != nil {
			log.Warn().Err(err).Msg("Readiness check failed: database unhealthy")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"status":"not_ready","reason":"database_unavailable","error":"%s"}`, err.Error()))) //nolint:errcheck // HTTP response write failure is not recoverable
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready","service":"otel-worker","database":"connected"}`)) //nolint:errcheck // HTTP response write failure is not recoverable
	})

	// CSV download endpoint
	http.HandleFunc("/download/", func(w http.ResponseWriter, r *http.Request) {
		// Extract filename from path
		filename := r.URL.Path[len("/download/"):]
		if filename == "" {
			http.Error(w, "Filename required", http.StatusBadRequest)
			return
		}

		// Security: only allow distance_*.csv files
		if len(filename) < 13 || filename[:9] != "distance_" || filename[len(filename)-4:] != ".csv" {
			http.Error(w, "Invalid filename format", http.StatusBadRequest)
			return
		}

		// Construct file path using configured CSV output path
		csvPath := fmt.Sprintf("%s/%s", cfg.CSVOutputPath, filename)

		// Serve the file
		http.ServeFile(w, r, csvPath)
	})

	go func() {
		log.Info().Str("port", cfg.HTTPPort).Msg("HTTP health server listening")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info().Msg("Shutdown signal received, gracefully stopping...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("HTTP server shutdown failed")
	} else {
		log.Info().Msg("HTTP server stopped")
	}

	// Stop gRPC server
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-shutdownCtx.Done():
		log.Warn().Msg("Shutdown timeout exceeded, forcing stop")
		grpcServer.Stop()
	case <-stopped:
		log.Info().Msg("gRPC server stopped")
	}

	// Shutdown distance service workers
	if err := distanceServer.Shutdown(10 * time.Second); err != nil {
		log.Error().Err(err).Msg("Failed to shutdown distance service")
	}

	log.Info().Msg("Service shutdown complete")
}

// setLogLevel configures the global log level
func setLogLevel(level string) {
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	log.Info().Str("level", level).Msg("Log level set")
}
