package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/stuartshay/otel-worker/internal/config"
	"github.com/stuartshay/otel-worker/internal/database"
	grpcserver "github.com/stuartshay/otel-worker/internal/grpc"
	distancev1 "github.com/stuartshay/otel-worker/proto/distance/v1"
)

func main() {
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

	// Initialize database client
	dbClient, err := database.NewClient(cfg.DatabaseDSN())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database client")
	}
	defer dbClient.Close()

	log.Info().Msg("Database connection established")

	// Verify database connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := dbClient.HealthCheck(ctx); err != nil {
		log.Fatal().Err(err).Msg("Database health check failed")
	}

	log.Info().Msg("Database health check passed")

	// Initialize gRPC server
	grpcServer := grpc.NewServer()

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

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info().Msg("Shutdown signal received, gracefully stopping...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

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
