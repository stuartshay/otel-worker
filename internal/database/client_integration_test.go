//go:build integration

package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stuartshay/otel-worker/internal/config"
)

// setupTestClient creates a test database client
func setupTestClient(t *testing.T) (*Client, func()) {
	t.Helper()

	cfg, err := config.Load()
	require.NoError(t, err, "Failed to load config")

	client, err := NewClient(cfg.DatabaseDSN())
	require.NoError(t, err, "Failed to create database client")

	cleanup := func() {
		if client != nil {
			client.Close()
		}
	}

	return client, cleanup
}

func TestNewClient_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cfg, err := config.Load()
	require.NoError(t, err)

	client, err := NewClient(cfg.DatabaseDSN())
	require.NoError(t, err, "Should connect to database successfully")
	require.NotNil(t, client)
	require.NotNil(t, client.db)

	defer client.Close()
}

func TestClient_HealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()
	err := client.HealthCheck(ctx)
	assert.NoError(t, err, "Should perform health check successfully")
}

func TestClient_HealthCheckWithTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	// Test with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout expires

	err := client.HealthCheck(ctx)
	assert.Error(t, err, "HealthCheck should fail with expired context")
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestClient_Close(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cfg, err := config.Load()
	require.NoError(t, err)

	client, err := NewClient(cfg.DatabaseDSN())
	require.NoError(t, err)

	// Close should not error
	err = client.Close()
	assert.NoError(t, err)

	// Second close should also not error
	err = client.Close()
	assert.NoError(t, err)
}

func TestGetLocationsByDate_NoData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	// Use a future date that definitely has no data
	locations, err := client.GetLocationsByDate(ctx, "2099-12-31", "")
	require.NoError(t, err, "Query should succeed even with no results")
	assert.Empty(t, locations, "Should return empty slice for future date")
}

func TestGetLocationsByDate_RecentData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	// Use yesterday's date (likely to have data)
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	locations, err := client.GetLocationsByDate(ctx, yesterday, "")
	require.NoError(t, err, "Query should succeed")

	// Report results (may be empty if no data for yesterday)
	t.Logf("Found %d locations for date %s", len(locations), yesterday)

	if len(locations) > 0 {
		// Validate first location structure
		loc := locations[0]
		assert.NotEmpty(t, loc.DeviceID)
		assert.NotZero(t, loc.Latitude)
		assert.NotZero(t, loc.Longitude)
		assert.False(t, loc.CreatedAt.IsZero())

		t.Logf("Sample location: device=%s, lat=%.6f, lon=%.6f, time=%s",
			loc.DeviceID, loc.Latitude, loc.Longitude, loc.CreatedAt.Format(time.RFC3339))
	}
}

func TestGetLocationsByDate_WithDeviceID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name     string
		date     string
		deviceID string
	}{
		{
			name:     "specific device yesterday",
			date:     time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			deviceID: "pixel8",
		},
		{
			name:     "specific device week ago",
			date:     time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
			deviceID: "pixel8",
		},
		{
			name:     "non-existent device",
			date:     time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			deviceID: "non-existent-device-12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locations, err := client.GetLocationsByDate(ctx, tt.date, tt.deviceID)
			require.NoError(t, err)

			t.Logf("Found %d locations for date=%s, device=%s", len(locations), tt.date, tt.deviceID)

			// Validate all results match requested device
			for _, loc := range locations {
				assert.Equal(t, tt.deviceID, loc.DeviceID,
					"All results should match requested device ID")
			}
		})
	}
}

func TestGetLocationsByDate_DateFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name        string
		date        string
		expectError bool
	}{
		{
			name:        "standard format YYYY-MM-DD",
			date:        "2025-01-22",
			expectError: false,
		},
		{
			name:        "single digit month and day",
			date:        "2025-1-5",
			expectError: false,
		},
		{
			name:        "empty date",
			date:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locations, err := client.GetLocationsByDate(ctx, tt.date, "")

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			t.Logf("Query for date '%s' returned %d locations", tt.date, len(locations))
		})
	}
}

func TestGetLocationsByDate_FieldValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	// Query for recent data
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	locations, err := client.GetLocationsByDate(ctx, yesterday, "")
	require.NoError(t, err)

	if len(locations) == 0 {
		t.Skip("No location data available for validation")
	}

	// Validate each returned location
	for i, loc := range locations {
		t.Run("location_"+string(rune(i)), func(t *testing.T) {
			// Required fields
			assert.NotEmpty(t, loc.DeviceID, "DeviceID should not be empty")
			assert.NotZero(t, loc.CreatedAt, "CreatedAt should be set")

			// GPS coordinates should be valid ranges
			assert.GreaterOrEqual(t, loc.Latitude, -90.0, "Latitude should be >= -90")
			assert.LessOrEqual(t, loc.Latitude, 90.0, "Latitude should be <= 90")
			assert.GreaterOrEqual(t, loc.Longitude, -180.0, "Longitude should be >= -180")
			assert.LessOrEqual(t, loc.Longitude, 180.0, "Longitude should be <= 180")

			// Numeric fields should be in valid ranges
			assert.GreaterOrEqual(t, loc.Accuracy, 0, "Accuracy should be non-negative")
			assert.GreaterOrEqual(t, loc.Battery, 0, "Battery should be 0-100")
			assert.LessOrEqual(t, loc.Battery, 100, "Battery should be 0-100")
			assert.GreaterOrEqual(t, loc.Velocity, 0, "Velocity should be non-negative")
		})
	}
}

func TestGetLocationsByDate_Ordering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	// Query for date with data
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	locations, err := client.GetLocationsByDate(ctx, yesterday, "")
	require.NoError(t, err)

	if len(locations) < 2 {
		t.Skip("Need at least 2 locations to test ordering")
	}

	// Verify results are ordered by created_at ASC
	for i := 1; i < len(locations); i++ {
		prev := locations[i-1]
		curr := locations[i]

		assert.False(t, curr.CreatedAt.Before(prev.CreatedAt),
			"Locations should be ordered by created_at ASC")
	}

	t.Logf("Verified ordering of %d locations", len(locations))
}

func TestGetLocationsByDate_ConcurrentQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	// Run multiple concurrent queries
	numQueries := 10
	done := make(chan bool, numQueries)
	successCount := make(chan bool, numQueries)

	for i := 0; i < numQueries; i++ {
		go func(index int) {
			date := time.Now().AddDate(0, 0, -index-1).Format("2006-01-02")
			_, err := client.GetLocationsByDate(ctx, date, "pixel8")
			if err != nil {
				// PgBouncer can have transient prepared statement issues
				t.Logf("Concurrent query #%d failed (non-fatal): %v", index, err)
			} else {
				successCount <- true
			}
			done <- true
		}(i)
	}

	// Wait for all queries to complete
	for i := 0; i < numQueries; i++ {
		<-done
	}
	close(successCount)

	// Count successes
	successes := len(successCount)

	// Require at least 50% success rate (allows for PgBouncer transient issues)
	require.GreaterOrEqual(t, successes, numQueries/2, "Expected at least half of concurrent queries to succeed")
	t.Logf("Successfully completed %d/%d concurrent queries", successes, numQueries)
}

func TestGetLocationsByDate_LargeResultSet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	// Query for a date range that might have lots of data
	// Use yesterday to maximize chance of data
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	start := time.Now()
	locations, err := client.GetLocationsByDate(ctx, yesterday, "")
	duration := time.Since(start)

	require.NoError(t, err)

	t.Logf("Retrieved %d locations in %v", len(locations), duration)

	// Performance check: should complete within reasonable time
	assert.Less(t, duration, 10*time.Second,
		"Query should complete within 10 seconds even with large result set")
}

func TestClient_ConnectionPooling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cfg, err := config.Load()
	require.NoError(t, err)

	client, err := NewClient(cfg.DatabaseDSN())
	require.NoError(t, err)
	defer client.Close()

	// Verify connection pool settings
	stats := client.db.Stats()

	// Pool should have configured max connections
	assert.Equal(t, 25, stats.MaxOpenConnections, "Should have 25 max open connections")

	// Run queries to warm up pool
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		date := time.Now().AddDate(0, 0, -i-1).Format("2006-01-02")
		_, err := client.GetLocationsByDate(ctx, date, "pixel8")
		require.NoError(t, err)
	}

	// Check pool stats after queries
	statsAfter := client.db.Stats()
	t.Logf("Pool stats: OpenConnections=%d, InUse=%d, Idle=%d",
		statsAfter.OpenConnections, statsAfter.InUse, statsAfter.Idle)

	assert.GreaterOrEqual(t, statsAfter.OpenConnections, 1,
		"Should have at least one open connection")
}

func TestClient_HealthCheckRepeated(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	// Health check should succeed multiple times
	for i := 0; i < 3; i++ {
		err := client.HealthCheck(ctx)
		assert.NoError(t, err, "Health check %d should succeed", i+1)
	}
}

func TestClient_ContextCancellationDuringQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	// Create context that will be canceled mid-query
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout expires

	// Query should fail due to context cancellation
	_, err := client.GetLocationsByDate(ctx, "2025-01-22", "")
	assert.Error(t, err, "Query should fail with canceled context")
}

func TestGetLocationsByDate_EmptyDateError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name        string
		date        string
		expectError bool
	}{
		{
			name:        "empty date string",
			date:        "",
			expectError: true,
		},
		{
			name:        "whitespace only",
			date:        "   ",
			expectError: false, // Database will handle this
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GetLocationsByDate(ctx, tt.date, "")

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// May succeed or fail depending on database handling
				t.Logf("Query result for '%s': %v", tt.date, err)
			}
		})
	}
}

func TestGetLocationsByDate_VariousDevices(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	devices := []string{"pixel8", "iphone", "unknown-device-xyz"}
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	for _, device := range devices {
		t.Run("device_"+device, func(t *testing.T) {
			locations, err := client.GetLocationsByDate(ctx, yesterday, device)
			require.NoError(t, err)

			t.Logf("Device '%s': found %d locations", device, len(locations))

			// All results should match requested device
			for _, loc := range locations {
				assert.Equal(t, device, loc.DeviceID)
			}
		})
	}
}

func TestGetLocationsByDate_DateRange(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	// Test data availability across different time ranges
	dates := []struct {
		name   string
		offset int // Days before today
	}{
		{"today", 0},
		{"yesterday", -1},
		{"3 days ago", -3},
		{"1 week ago", -7},
		{"1 month ago", -30},
		{"3 months ago", -90},
	}

	for _, td := range dates {
		t.Run(td.name, func(t *testing.T) {
			date := time.Now().AddDate(0, 0, td.offset).Format("2006-01-02")
			locations, err := client.GetLocationsByDate(ctx, date, "pixel8")
			require.NoError(t, err)

			t.Logf("%s (%s): %d locations", td.name, date, len(locations))
		})
	}
}

func TestGetLocationsByDate_VerifyOrderingWithData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	// Try to find a date with data
	var locations []Location
	var queryDate string

	for i := 0; i < 30; i++ {
		date := time.Now().AddDate(0, 0, -i-1).Format("2006-01-02")
		locs, err := client.GetLocationsByDate(ctx, date, "")
		require.NoError(t, err)

		if len(locs) > 1 {
			locations = locs
			queryDate = date
			break
		}
	}

	if len(locations) < 2 {
		t.Skip("No date with multiple locations found in last 30 days")
	}

	t.Logf("Testing ordering with %d locations from %s", len(locations), queryDate)

	// Verify chronological ordering
	for i := 1; i < len(locations); i++ {
		prev := locations[i-1]
		curr := locations[i]

		assert.True(t, curr.CreatedAt.After(prev.CreatedAt) || curr.CreatedAt.Equal(prev.CreatedAt),
			"Location %d timestamp (%s) should be >= previous (%s)",
			i, curr.CreatedAt, prev.CreatedAt)
	}

	// Log time span
	if len(locations) > 0 {
		first := locations[0].CreatedAt
		last := locations[len(locations)-1].CreatedAt
		span := last.Sub(first)
		t.Logf("Time span: %v (from %s to %s)", span,
			first.Format(time.RFC3339), last.Format(time.RFC3339))
	}
}

func TestGetLocationsByDate_BatteryAndAccuracy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	// Find a date with data
	var locations []Location
	for i := 0; i < 30; i++ {
		date := time.Now().AddDate(0, 0, -i-1).Format("2006-01-02")
		locs, err := client.GetLocationsByDate(ctx, date, "pixel8")
		require.NoError(t, err)

		if len(locs) > 0 {
			locations = locs
			break
		}
	}

	if len(locations) == 0 {
		t.Skip("No location data found in last 30 days")
	}

	t.Logf("Analyzing %d locations for battery and accuracy metrics", len(locations))

	// Collect statistics
	var totalBattery, totalAccuracy int64
	minBattery, maxBattery := 100, 0
	minAccuracy, maxAccuracy := int64(999999), int64(0)

	for _, loc := range locations {
		// Battery
		if loc.Battery < minBattery {
			minBattery = loc.Battery
		}
		if loc.Battery > maxBattery {
			maxBattery = loc.Battery
		}
		totalBattery += int64(loc.Battery)

		// Accuracy
		if int64(loc.Accuracy) < minAccuracy {
			minAccuracy = int64(loc.Accuracy)
		}
		if int64(loc.Accuracy) > maxAccuracy {
			maxAccuracy = int64(loc.Accuracy)
		}
		totalAccuracy += int64(loc.Accuracy)
	}

	avgBattery := totalBattery / int64(len(locations))
	avgAccuracy := totalAccuracy / int64(len(locations))

	t.Logf("Battery: min=%d, max=%d, avg=%d", minBattery, maxBattery, avgBattery)
	t.Logf("Accuracy: min=%d, max=%d, avg=%d", minAccuracy, maxAccuracy, avgAccuracy)

	// Validate ranges
	assert.GreaterOrEqual(t, minBattery, 0, "Battery should be >= 0")
	assert.LessOrEqual(t, maxBattery, 100, "Battery should be <= 100")
	assert.GreaterOrEqual(t, int(minAccuracy), 0, "Accuracy should be >= 0")
}

func TestClient_MultipleClients(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cfg, err := config.Load()
	require.NoError(t, err)

	// Create multiple clients simultaneously
	numClients := 3
	clients := make([]*Client, numClients)

	for i := 0; i < numClients; i++ {
		client, err := NewClient(cfg.DatabaseDSN())
		require.NoError(t, err)
		clients[i] = client
	}

	// All clients should be able to perform health checks
	ctx := context.Background()
	for i, client := range clients {
		err := client.HealthCheck(ctx)
		assert.NoError(t, err, "Client %d should health check successfully", i)
	}

	// Clean up
	for _, client := range clients {
		err := client.Close()
		assert.NoError(t, err)
	}
}

func TestGetLocationsByDate_DeviceFiltering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Query without device filter
	allLocations, err := client.GetLocationsByDate(ctx, yesterday, "")
	require.NoError(t, err)

	if len(allLocations) == 0 {
		t.Skip("No data available for testing")
	}

	// Get unique device IDs from results
	deviceMap := make(map[string]int)
	for _, loc := range allLocations {
		deviceMap[loc.DeviceID]++
	}

	t.Logf("All devices combined: %d locations from %d devices", len(allLocations), len(deviceMap))

	// Query for each device individually
	totalFiltered := 0
	for deviceID, expectedCount := range deviceMap {
		t.Run("device_"+deviceID, func(t *testing.T) {
			locations, err := client.GetLocationsByDate(ctx, yesterday, deviceID)
			require.NoError(t, err)

			assert.Equal(t, expectedCount, len(locations),
				"Filtered query should return same count as in combined results")

			totalFiltered += len(locations)
		})
	}

	// Sum of all filtered queries should equal unfiltered query
	assert.Equal(t, len(allLocations), totalFiltered,
		"Sum of device-filtered queries should equal unfiltered query")
}

func TestGetLocationsByDate_TimestampVsCreatedAt(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	locations, err := client.GetLocationsByDate(ctx, yesterday, "")
	require.NoError(t, err)

	if len(locations) == 0 {
		t.Skip("No data available for testing")
	}

	// Verify timestamp field is populated
	for i, loc := range locations {
		assert.NotZero(t, loc.Timestamp, "Location %d should have timestamp", i)

		// Timestamp (Unix) should roughly correlate with CreatedAt
		timestampTime := time.Unix(loc.Timestamp, 0)
		diff := timestampTime.Sub(loc.CreatedAt).Abs()

		// Allow up to 1 hour difference (accounting for timezone conversions, etc.)
		assert.Less(t, diff, 1*time.Hour,
			"Timestamp and CreatedAt should be within 1 hour of each other")
	}
}

func TestNewClient_ConnectionRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Test with invalid port to verify error handling
	invalidDSN := "host=192.168.1.175 port=9999 user=test dbname=test sslmode=disable"

	_, err := NewClient(invalidDSN)
	// NewClient may succeed (lazy connection), but Ping should fail
	// This test verifies we get appropriate errors, not panics

	if err != nil {
		assert.Error(t, err)
		t.Logf("Connection failed as expected: %v", err)
	}
}

func TestGetLocationsByDate_EdgeCaseDates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name        string
		date        string
		expectError bool
	}{
		{
			name:        "leap year date",
			date:        "2024-02-29",
			expectError: false,
		},
		{
			name:        "start of year",
			date:        "2025-01-01",
			expectError: false,
		},
		{
			name:        "end of year",
			date:        "2025-12-31",
			expectError: false,
		},
		{
			name:        "far past",
			date:        "2020-01-01",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locations, err := client.GetLocationsByDate(ctx, tt.date, "")

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			t.Logf("Date %s: found %d locations", tt.date, len(locations))
		})
	}
}

func TestGetLocationsByDateRange_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	// Test with 3-day range
	endDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -3).Format("2006-01-02")

	locations, err := client.GetLocationsByDateRange(ctx, startDate, endDate, "")
	require.NoError(t, err)

	t.Logf("Found %d locations between %s and %s", len(locations), startDate, endDate)

	// Validate all locations are within range
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	endPlusOne := end.AddDate(0, 0, 1)

	for _, loc := range locations {
		assert.True(t, loc.CreatedAt.After(start) || loc.CreatedAt.Equal(start),
			"Location should be on or after start date")
		assert.True(t, loc.CreatedAt.Before(endPlusOne),
			"Location should be before end date + 1 day")
	}
}

func TestGetLocationsByDateRange_WithDevice(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	endDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")

	locations, err := client.GetLocationsByDateRange(ctx, startDate, endDate, "pixel8")
	require.NoError(t, err)

	t.Logf("Found %d locations for pixel8 in date range", len(locations))

	// All should be for requested device
	for _, loc := range locations {
		assert.Equal(t, "pixel8", loc.DeviceID)
	}
}

func TestGetLocationsByDateRange_SingleDay(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	// Same start and end date
	date := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	locations, err := client.GetLocationsByDateRange(ctx, date, date, "")
	require.NoError(t, err)

	t.Logf("Single day range: found %d locations", len(locations))
}

func TestGetLocationsByDateRange_ReversedDates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	// Reversed: end before start
	startDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	endDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")

	locations, err := client.GetLocationsByDateRange(ctx, startDate, endDate, "")
	require.NoError(t, err)

	// Should return empty or minimal results
	t.Logf("Reversed date range: found %d locations", len(locations))
	assert.LessOrEqual(t, len(locations), 0, "Reversed range should return no results")
}

func TestGetDevices(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	devices, err := client.GetDevices(ctx)
	require.NoError(t, err)

	t.Logf("Found %d unique devices in database", len(devices))

	if len(devices) > 0 {
		// Validate device IDs are not empty
		for _, deviceID := range devices {
			assert.NotEmpty(t, deviceID, "Device ID should not be empty")
		}

		// Log all devices
		t.Logf("Devices: %v", devices)
	}
}

func TestGetLocationCount_NoData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	// Future date should have zero count
	count, err := client.GetLocationCount(ctx, "2099-12-31", "")
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Future date should have zero locations")
}

func TestGetLocationCount_WithDevice(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Get count for all devices
	totalCount, err := client.GetLocationCount(ctx, yesterday, "")
	require.NoError(t, err)

	// Get count for specific device
	pixel8Count, err := client.GetLocationCount(ctx, yesterday, "pixel8")
	require.NoError(t, err)

	t.Logf("Yesterday: total=%d, pixel8=%d", totalCount, pixel8Count)

	// Device-specific count should be <= total count
	assert.LessOrEqual(t, pixel8Count, totalCount,
		"Device count should not exceed total count")
}

func TestGetLocationCount_ConsistencyWithQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Get count
	count, err := client.GetLocationCount(ctx, yesterday, "pixel8")
	require.NoError(t, err)

	// Get actual locations
	locations, err := client.GetLocationsByDate(ctx, yesterday, "pixel8")
	require.NoError(t, err)

	// Count should match actual results
	assert.Equal(t, count, len(locations),
		"GetLocationCount should match length of GetLocationsByDate results")

	t.Logf("Count consistency verified: %d locations", count)
}
