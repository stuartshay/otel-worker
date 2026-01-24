// Package grpc implements the DistanceService gRPC server handlers
// for distance calculation job management and CSV generation.
package grpc

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/stuartshay/otel-worker/internal/calculator"
	"github.com/stuartshay/otel-worker/internal/config"
	"github.com/stuartshay/otel-worker/internal/database"
	"github.com/stuartshay/otel-worker/internal/queue"
	distancev1 "github.com/stuartshay/otel-worker/proto/distance/v1"
)

// Server implements the DistanceService gRPC server
type Server struct {
	distancev1.UnimplementedDistanceServiceServer
	cfg   *config.Config
	db    *database.Client
	queue *queue.Queue
}

// NewServer creates a new gRPC server instance
func NewServer(cfg *config.Config, db *database.Client) *Server {
	s := &Server{
		cfg: cfg,
		db:  db,
	}

	// Initialize job queue with processor
	s.queue = queue.NewQueue(5, s.processDistanceJob)

	return s
}

// CalculateDistanceFromHome initiates an async distance calculation job
func (s *Server) CalculateDistanceFromHome(ctx context.Context, req *distancev1.CalculateDistanceRequest) (*distancev1.CalculateDistanceResponse, error) {
	log.Info().
		Str("date", req.Date).
		Str("device_id", req.DeviceId).
		Msg("Received distance calculation request")

	// Validate date format
	if req.Date == "" {
		return nil, fmt.Errorf("date is required")
	}

	// Enqueue job
	jobID, err := s.queue.Enqueue(req.Date, req.DeviceId)
	if err != nil {
		log.Error().Err(err).Msg("Failed to enqueue job")
		return nil, fmt.Errorf("failed to enqueue job: %w", err)
	}

	return &distancev1.CalculateDistanceResponse{
		JobId:    jobID,
		Status:   "queued",
		QueuedAt: timestamppb.Now(),
	}, nil
}

// GetJobStatus returns the current status of a distance calculation job
func (s *Server) GetJobStatus(ctx context.Context, req *distancev1.GetJobStatusRequest) (*distancev1.GetJobStatusResponse, error) {
	job, err := s.queue.GetJob(req.JobId)
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}

	resp := &distancev1.GetJobStatusResponse{
		JobId:    job.ID,
		Status:   string(job.Status),
		QueuedAt: timestamppb.New(job.QueuedAt),
	}

	if job.StartedAt != nil {
		resp.StartedAt = timestamppb.New(*job.StartedAt)
	}

	if job.CompletedAt != nil {
		resp.CompletedAt = timestamppb.New(*job.CompletedAt)
	}

	if job.ErrorMessage != "" {
		resp.ErrorMessage = job.ErrorMessage
	}

	if job.Result != nil {
		resp.Result = &distancev1.JobResult{
			CsvPath:          job.Result.CSVPath,
			TotalDistanceKm:  job.Result.TotalDistanceKM,
			MaxDistanceKm:    job.Result.MaxDistanceKM,
			MinDistanceKm:    job.Result.MinDistanceKM,
			TotalLocations:   int32(job.Result.TotalLocations),
			Date:             job.Date,
			DeviceId:         job.DeviceID,
			ProcessingTimeMs: job.Result.ProcessingTimeMS,
		}
	}

	return resp, nil
}

// ListJobs returns a list of distance calculation jobs with optional filtering
func (s *Server) ListJobs(ctx context.Context, req *distancev1.ListJobsRequest) (*distancev1.ListJobsResponse, error) {
	limit := int(req.Limit)
	if limit == 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	offset := int(req.Offset)

	jobs := s.queue.ListJobs(queue.JobStatus(req.Status), limit, offset)

	resp := &distancev1.ListJobsResponse{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	for _, job := range jobs {
		summary := &distancev1.JobSummary{
			JobId:    job.ID,
			Status:   string(job.Status),
			Date:     job.Date,
			DeviceId: job.DeviceID,
			QueuedAt: timestamppb.New(job.QueuedAt),
		}

		if job.CompletedAt != nil {
			summary.CompletedAt = timestamppb.New(*job.CompletedAt)
		}

		resp.Jobs = append(resp.Jobs, summary)
	}

	resp.TotalCount = int32(len(jobs))

	return resp, nil
}

// processDistanceJob is the worker function that processes distance calculation jobs
func (s *Server) processDistanceJob(ctx context.Context, job *queue.Job) (*queue.JobResult, error) {
	log.Info().
		Str("job_id", job.ID).
		Str("date", job.Date).
		Str("device_id", job.DeviceID).
		Msg("Processing distance calculation job")

	// Fetch locations from database
	locations, err := s.db.GetLocationsByDate(ctx, job.Date, job.DeviceID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch locations from database")
		return nil, fmt.Errorf("database query failed: %w", err)
	}

	if len(locations) == 0 {
		log.Warn().Str("date", job.Date).Msg("No locations found for date")
		return nil, fmt.Errorf("no locations found for date %s", job.Date)
	}

	// Convert database locations to calculator locations
	calcLocations := make([]calculator.Location, len(locations))
	for i, loc := range locations {
		calcLocations[i] = calculator.Location{
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
		}
	}

	// Calculate distance metrics
	metrics := calculator.CalculateMetrics(
		s.cfg.HomeLatitude,
		s.cfg.HomeLongitude,
		calcLocations,
	)

	log.Info().
		Float64("total_distance_km", metrics.TotalDistanceKM).
		Float64("max_distance_km", metrics.MaxDistanceKM).
		Float64("min_distance_km", metrics.MinDistanceKM).
		Int("total_locations", metrics.TotalLocations).
		Msg("Distance metrics calculated")

	// Generate CSV file
	csvPath, err := s.generateCSV(job.Date, job.DeviceID, locations, metrics)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate CSV file")
		return nil, fmt.Errorf("CSV generation failed: %w", err)
	}

	return &queue.JobResult{
		CSVPath:         csvPath,
		TotalDistanceKM: metrics.TotalDistanceKM,
		MaxDistanceKM:   metrics.MaxDistanceKM,
		MinDistanceKM:   metrics.MinDistanceKM,
		TotalLocations:  metrics.TotalLocations,
	}, nil
}

// generateCSV creates a CSV file with location data and distance calculations
func (s *Server) generateCSV(date, deviceID string, locations []database.Location, metrics calculator.DistanceMetrics) (string, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(s.cfg.CSVOutputPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate filename
	dateStr := date
	if len(date) == 10 {
		dateStr = date[0:4] + date[5:7] + date[8:10] // YYYYMMDD
	}
	filename := fmt.Sprintf("distance_%s.csv", dateStr)
	if deviceID != "" {
		filename = fmt.Sprintf("distance_%s_%s.csv", dateStr, deviceID)
	}
	csvPath := filepath.Join(s.cfg.CSVOutputPath, filename)

	// Create CSV file
	file, err := os.Create(csvPath)
	if err != nil {
		return "", fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("Failed to close CSV file")
		}
	}()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"timestamp", "device_id", "latitude", "longitude",
		"distance_from_home_km", "accuracy", "battery", "velocity",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, loc := range locations {
		distance := calculator.DistanceFromHome(
			s.cfg.HomeLatitude,
			s.cfg.HomeLongitude,
			loc.Latitude,
			loc.Longitude,
		)

		row := []string{
			loc.CreatedAt.Format(time.RFC3339),
			loc.DeviceID,
			fmt.Sprintf("%.6f", loc.Latitude),
			fmt.Sprintf("%.6f", loc.Longitude),
			fmt.Sprintf("%.2f", distance),
			fmt.Sprintf("%d", loc.Accuracy),
			fmt.Sprintf("%d", loc.Battery),
			fmt.Sprintf("%d", loc.Velocity),
		}

		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	// Write summary footer
	_ = writer.Write([]string{})
	_ = writer.Write([]string{"Summary"})
	_ = writer.Write([]string{"Total Distance (km)", fmt.Sprintf("%.2f", metrics.TotalDistanceKM)})
	_ = writer.Write([]string{"Max Distance (km)", fmt.Sprintf("%.2f", metrics.MaxDistanceKM)})
	_ = writer.Write([]string{"Min Distance (km)", fmt.Sprintf("%.2f", metrics.MinDistanceKM)})
	_ = writer.Write([]string{"Total Locations", fmt.Sprintf("%d", metrics.TotalLocations)})
	_ = writer.Write([]string{"Average Distance (km)", fmt.Sprintf("%.2f", metrics.AvgDistanceKM)})

	log.Info().Str("csv_path", csvPath).Msg("CSV file generated successfully")

	return csvPath, nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(timeout time.Duration) error {
	return s.queue.Shutdown(timeout)
}
