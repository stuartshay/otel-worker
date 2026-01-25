package grpc

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stuartshay/otel-worker/internal/config"
	"github.com/stuartshay/otel-worker/internal/database"
	distancev1 "github.com/stuartshay/otel-worker/proto/distance/v1"
)

// setupTestServer creates a test server with database connection
func setupTestServer(t *testing.T) (*Server, func()) {
	t.Helper()

	// Load configuration from environment
	cfg, err := config.Load()
	require.NoError(t, err, "Failed to load config")

	// Create temporary CSV output directory
	tmpDir := t.TempDir()
	cfg.CSVOutputPath = tmpDir

	// Connect to database
	db, err := database.NewClient(cfg.DatabaseDSN())
	require.NoError(t, err, "Failed to connect to database")

	// Create server
	server := NewServer(cfg, db)

	cleanup := func() {
		if server.queue != nil {
			_ = server.queue.Shutdown(5 * time.Second)
		}
		if db != nil {
			db.Close()
		}
	}

	return server, cleanup
}

func TestCalculateDistanceFromHome(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server, cleanup := setupTestServer(t)
	defer cleanup()

	tests := []struct {
		name          string
		request       *distancev1.CalculateDistanceRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "valid request with date and device",
			request: &distancev1.CalculateDistanceRequest{
				Date:     "2025-01-22",
				DeviceId: "pixel8",
			},
			expectError: false,
		},
		{
			name: "valid request with date only",
			request: &distancev1.CalculateDistanceRequest{
				Date:     "2025-01-22",
				DeviceId: "",
			},
			expectError: false,
		},
		{
			name: "empty date",
			request: &distancev1.CalculateDistanceRequest{
				Date:     "",
				DeviceId: "pixel8",
			},
			expectError:   true,
			errorContains: "date is required",
		},
		{
			name: "future date (no data)",
			request: &distancev1.CalculateDistanceRequest{
				Date:     "2026-12-31",
				DeviceId: "pixel8",
			},
			expectError: false, // Job created but will fail during processing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			resp, err := server.CalculateDistanceFromHome(ctx, tt.request)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, resp.JobId, "Job ID should not be empty")
			assert.Equal(t, "queued", resp.Status, "Initial status should be queued")
			assert.NotNil(t, resp.QueuedAt, "QueuedAt timestamp should be set")
		})
	}
}

func TestGetJobStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create a job first
	createResp, err := server.CalculateDistanceFromHome(ctx, &distancev1.CalculateDistanceRequest{
		Date:     "2025-01-22",
		DeviceId: "pixel8",
	})
	require.NoError(t, err)
	require.NotEmpty(t, createResp.JobId)

	// Wait for job to complete (or fail)
	time.Sleep(2 * time.Second)

	tests := []struct {
		name          string
		jobID         string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid job ID",
			jobID:       createResp.JobId,
			expectError: false,
		},
		{
			name:          "invalid job ID",
			jobID:         "00000000-0000-0000-0000-000000000000",
			expectError:   true,
			errorContains: "job not found",
		},
		{
			name:          "empty job ID",
			jobID:         "",
			expectError:   true,
			errorContains: "job not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.GetJobStatus(ctx, &distancev1.GetJobStatusRequest{
				JobId: tt.jobID,
			})

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, createResp.JobId, resp.JobId)
			assert.NotEmpty(t, resp.Status)
			assert.NotNil(t, resp.QueuedAt)

			// Job should be completed or failed
			assert.Contains(t, []string{"completed", "failed"}, resp.Status)

			if resp.Status == "completed" {
				assert.NotNil(t, resp.CompletedAt)
				assert.NotNil(t, resp.Result)
				assert.NotEmpty(t, resp.Result.CsvPath)
				assert.Greater(t, resp.Result.TotalLocations, int32(0))
			} else if resp.Status == "failed" {
				assert.NotNil(t, resp.CompletedAt)
				assert.NotEmpty(t, resp.ErrorMessage)
			}
		})
	}
}

func TestListJobs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create multiple jobs first
	dates := []string{"2025-01-22", "2025-01-23", "2025-01-24"}
	for _, date := range dates {
		_, err := server.CalculateDistanceFromHome(ctx, &distancev1.CalculateDistanceRequest{
			Date:     date,
			DeviceId: "pixel8",
		})
		require.NoError(t, err)
	}

	// Wait for jobs to complete
	time.Sleep(3 * time.Second)

	tests := []struct {
		name          string
		request       *distancev1.ListJobsRequest
		expectError   bool
		expectMinJobs int
	}{
		{
			name: "default pagination (limit 10)",
			request: &distancev1.ListJobsRequest{
				Limit: 10,
			},
			expectError:   false,
			expectMinJobs: 3,
		},
		{
			name: "small limit",
			request: &distancev1.ListJobsRequest{
				Limit: 2,
			},
			expectError:   false,
			expectMinJobs: 2,
		},
		{
			name: "zero limit (default to 50)",
			request: &distancev1.ListJobsRequest{
				Limit: 0,
			},
			expectError:   false,
			expectMinJobs: 3,
		},
		{
			name: "large limit (capped at 500)",
			request: &distancev1.ListJobsRequest{
				Limit: 1000,
			},
			expectError:   false,
			expectMinJobs: 3,
		},
		{
			name: "with offset",
			request: &distancev1.ListJobsRequest{
				Limit:  10,
				Offset: 1,
			},
			expectError:   false,
			expectMinJobs: 2,
		},
		{
			name: "filter by completed status",
			request: &distancev1.ListJobsRequest{
				Status: "completed",
				Limit:  10,
			},
			expectError: false,
			// expectMinJobs not set - may be 0 if all jobs failed
		},
		{
			name: "filter by failed status",
			request: &distancev1.ListJobsRequest{
				Status: "failed",
				Limit:  10,
			},
			expectError: false,
			// expectMinJobs not set - depends on data availability
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.ListJobs(ctx, tt.request)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)

			if tt.expectMinJobs > 0 {
				assert.GreaterOrEqual(t, len(resp.Jobs), tt.expectMinJobs,
					"Should have at least %d jobs", tt.expectMinJobs)
			}

			// Validate job summaries
			for _, job := range resp.Jobs {
				assert.NotEmpty(t, job.JobId)
				assert.NotEmpty(t, job.Status)
				assert.NotEmpty(t, job.Date)
				assert.NotNil(t, job.QueuedAt)

				if tt.request.Status != "" {
					assert.Equal(t, tt.request.Status, job.Status,
						"Job status should match filter")
				}
			}

			// Validate pagination fields
			expectedLimit := tt.request.Limit
			if expectedLimit == 0 {
				expectedLimit = 50 // Default
			}
			if expectedLimit > 500 {
				expectedLimit = 500 // Cap
			}
			assert.Equal(t, expectedLimit, resp.Limit, "Limit should match normalized value")
			assert.Equal(t, tt.request.Offset, resp.Offset, "Offset should match request")
		})
	}
}

func TestCSVGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create a job that should generate CSV (using a date with data)
	createResp, err := server.CalculateDistanceFromHome(ctx, &distancev1.CalculateDistanceRequest{
		Date:     "2025-01-22",
		DeviceId: "pixel8",
	})
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(3 * time.Second)

	// Get job status
	statusResp, err := server.GetJobStatus(ctx, &distancev1.GetJobStatusRequest{
		JobId: createResp.JobId,
	})
	require.NoError(t, err)

	// If job completed successfully, verify CSV file
	if statusResp.Status == "completed" {
		assert.NotNil(t, statusResp.Result)
		assert.NotEmpty(t, statusResp.Result.CsvPath)

		// Verify CSV file exists
		_, err := os.Stat(statusResp.Result.CsvPath)
		assert.NoError(t, err, "CSV file should exist")

		// Verify filename format
		expectedFilename := "distance_20250122_pixel8.csv"
		assert.Equal(t, expectedFilename, filepath.Base(statusResp.Result.CsvPath))

		// Verify CSV content
		content, err := os.ReadFile(statusResp.Result.CsvPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "timestamp,device_id,latitude,longitude")
		assert.Contains(t, string(content), "Summary")
		assert.Contains(t, string(content), "Total Distance (km)")
	} else {
		// Job failed (expected if no data for date)
		assert.Equal(t, "failed", statusResp.Status)
		assert.NotEmpty(t, statusResp.ErrorMessage)
		t.Logf("Job failed as expected: %s", statusResp.ErrorMessage)
	}
}

func TestServerShutdown(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Shutdown should complete without error
	err := server.Shutdown(5 * time.Second)
	assert.NoError(t, err)
}

func TestJobProcessingWithRealDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Test with a recent date that likely has data
	testDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	createResp, err := server.CalculateDistanceFromHome(ctx, &distancev1.CalculateDistanceRequest{
		Date:     testDate,
		DeviceId: "pixel8",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, createResp.JobId)
	assert.Equal(t, "queued", createResp.Status)

	// Wait for job to process
	time.Sleep(5 * time.Second)

	// Check final status
	statusResp, err := server.GetJobStatus(ctx, &distancev1.GetJobStatusRequest{
		JobId: createResp.JobId,
	})
	require.NoError(t, err)

	// Job should be either completed or failed (not still queued/processing)
	assert.Contains(t, []string{"completed", "failed"}, statusResp.Status,
		"Job should finish processing within 5 seconds")

	if statusResp.Status == "completed" {
		// Validate result metrics
		assert.NotNil(t, statusResp.Result)
		assert.Greater(t, statusResp.Result.TotalLocations, int32(0))
		assert.GreaterOrEqual(t, statusResp.Result.MaxDistanceKm, statusResp.Result.MinDistanceKm)
		assert.GreaterOrEqual(t, statusResp.Result.MaxDistanceKm, float64(0))

		// Verify CSV file was created
		csvPath := statusResp.Result.CsvPath
		assert.FileExists(t, csvPath)

		// Verify CSV is in correct directory
		assert.Equal(t, server.cfg.CSVOutputPath, filepath.Dir(csvPath))
	}

	// Verify job appears in list
	listResp, err := server.ListJobs(ctx, &distancev1.ListJobsRequest{
		Limit: 10,
	})
	require.NoError(t, err)
	assert.Greater(t, len(listResp.Jobs), 0)

	// Find our job in the list
	var foundJob *distancev1.JobSummary
	for _, job := range listResp.Jobs {
		if job.JobId == createResp.JobId {
			foundJob = job
			break
		}
	}
	require.NotNil(t, foundJob, "Job should appear in list")
	assert.Equal(t, testDate, foundJob.Date)
	assert.Equal(t, "pixel8", foundJob.DeviceId)
}

func TestConcurrentJobs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create multiple jobs concurrently
	numJobs := 10
	jobIDs := make([]string, numJobs)

	for i := 0; i < numJobs; i++ {
		// Use different dates to create variety
		date := time.Now().AddDate(0, 0, -i-1).Format("2006-01-02")

		resp, err := server.CalculateDistanceFromHome(ctx, &distancev1.CalculateDistanceRequest{
			Date:     date,
			DeviceId: "pixel8",
		})
		require.NoError(t, err)
		jobIDs[i] = resp.JobId
	}

	// Wait for all jobs to process
	time.Sleep(10 * time.Second)

	// Verify all jobs completed or failed
	for i, jobID := range jobIDs {
		statusResp, err := server.GetJobStatus(ctx, &distancev1.GetJobStatusRequest{
			JobId: jobID,
		})
		require.NoError(t, err, "Job %d should be retrievable", i)
		assert.Contains(t, []string{"completed", "failed"}, statusResp.Status,
			"Job %d should be finished", i)
	}

	// Verify all jobs appear in list
	listResp, err := server.ListJobs(ctx, &distancev1.ListJobsRequest{
		Limit: int32(numJobs + 10), // Request more than we created
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(listResp.Jobs), numJobs,
		"Should list at least %d jobs", numJobs)
}

func TestCSVFilenameFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name             string
		date             string
		deviceID         string
		expectedFilename string
	}{
		{
			name:             "with device ID",
			date:             "2025-01-22",
			deviceID:         "pixel8",
			expectedFilename: "distance_20250122_pixel8.csv",
		},
		{
			name:             "without device ID",
			date:             "2025-01-23",
			deviceID:         "",
			expectedFilename: "distance_20250123.csv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createResp, err := server.CalculateDistanceFromHome(ctx, &distancev1.CalculateDistanceRequest{
				Date:     tt.date,
				DeviceId: tt.deviceID,
			})
			require.NoError(t, err)

			// Wait for processing
			time.Sleep(3 * time.Second)

			statusResp, err := server.GetJobStatus(ctx, &distancev1.GetJobStatusRequest{
				JobId: createResp.JobId,
			})
			require.NoError(t, err)

			if statusResp.Status == "completed" {
				assert.Equal(t, tt.expectedFilename, filepath.Base(statusResp.Result.CsvPath))
			}
		})
	}
}

func TestJobResultMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create job with a date that has data
	createResp, err := server.CalculateDistanceFromHome(ctx, &distancev1.CalculateDistanceRequest{
		Date:     "2025-01-22",
		DeviceId: "pixel8",
	})
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(3 * time.Second)

	statusResp, err := server.GetJobStatus(ctx, &distancev1.GetJobStatusRequest{
		JobId: createResp.JobId,
	})
	require.NoError(t, err)

	// If job completed, validate metrics
	if statusResp.Status == "completed" {
		result := statusResp.Result
		require.NotNil(t, result)

		// Validate distance metrics
		assert.GreaterOrEqual(t, result.MaxDistanceKm, result.MinDistanceKm,
			"Max distance should be >= min distance")
		assert.GreaterOrEqual(t, result.MaxDistanceKm, float64(0),
			"Max distance should be non-negative")
		assert.GreaterOrEqual(t, result.MinDistanceKm, float64(0),
			"Min distance should be non-negative")
		assert.GreaterOrEqual(t, result.TotalDistanceKm, float64(0),
			"Total distance should be non-negative")

		// Validate counts
		assert.Greater(t, result.TotalLocations, int32(0),
			"Should have at least one location")

		// Validate processing time
		assert.Greater(t, result.ProcessingTimeMs, int64(0),
			"Processing time should be recorded")

		// Validate date and device
		assert.Equal(t, "2025-01-22", result.Date)
		assert.Equal(t, "pixel8", result.DeviceId)

		t.Logf("Job completed successfully:")
		t.Logf("  Total Distance: %.2f km", result.TotalDistanceKm)
		t.Logf("  Max Distance: %.2f km", result.MaxDistanceKm)
		t.Logf("  Min Distance: %.2f km", result.MinDistanceKm)
		t.Logf("  Total Locations: %d", result.TotalLocations)
		t.Logf("  Processing Time: %d ms", result.ProcessingTimeMs)
		t.Logf("  CSV Path: %s", result.CsvPath)
	} else {
		t.Logf("Job failed: %s (expected if no data for date)", statusResp.ErrorMessage)
	}
}
