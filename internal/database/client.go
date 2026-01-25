// Package database provides PostgreSQL client functionality for accessing
// OwnTracks location data with connection pooling and health checks.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Client wraps a PostgreSQL database connection
type Client struct {
	db *sql.DB
}

// Location represents a GPS location record from the database
type Location struct {
	ID             int64
	DeviceID       string
	TID            string
	Latitude       float64
	Longitude      float64
	Accuracy       int
	Altitude       int
	Velocity       int
	Battery        int
	BatteryStatus  string
	ConnectionType string
	Trigger        string
	Timestamp      int64
	CreatedAt      time.Time
}

// NewClient creates a new database client with connection pooling
func NewClient(dsn string) (*Client, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to ping database: %w (also failed to close: %w)", err, closeErr)
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Client{db: db}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.db.Close()
}

// GetLocationsByDate retrieves GPS locations for a specific date
// Date should be in YYYY-MM-DD format
func (c *Client) GetLocationsByDate(ctx context.Context, date string, deviceID string) ([]Location, error) {
	query := `
		SELECT
			id, device_id, tid, latitude, longitude, accuracy,
			altitude, velocity, battery, battery_status,
			connection_type, trigger, EXTRACT(EPOCH FROM timestamp)::bigint AS timestamp, created_at
		FROM public.locations
		WHERE DATE(created_at) = $1
	`

	args := []interface{}{date}

	// Add device_id filter if specified
	if deviceID != "" {
		query += " AND device_id = $2"
		args = append(args, deviceID)
	}

	query += " ORDER BY created_at ASC"

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer func() { _ = rows.Close() }() // nolint:errcheck // Close in defer, error not actionable

	var locations []Location
	for rows.Next() {
		var loc Location
		var accuracy, altitude, velocity, battery, timestamp sql.NullInt64
		var batteryStatus, connectionType, trigger sql.NullString

		err := rows.Scan(
			&loc.ID,
			&loc.DeviceID,
			&loc.TID,
			&loc.Latitude,
			&loc.Longitude,
			&accuracy,
			&altitude,
			&velocity,
			&battery,
			&batteryStatus,
			&connectionType,
			&trigger,
			&timestamp,
			&loc.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		// Convert NULL values to zero values
		if accuracy.Valid {
			loc.Accuracy = int(accuracy.Int64)
		}
		if altitude.Valid {
			loc.Altitude = int(altitude.Int64)
		}
		if velocity.Valid {
			loc.Velocity = int(velocity.Int64)
		}
		if battery.Valid {
			loc.Battery = int(battery.Int64)
		}
		if timestamp.Valid {
			loc.Timestamp = timestamp.Int64
		}
		if batteryStatus.Valid {
			loc.BatteryStatus = batteryStatus.String
		}
		if connectionType.Valid {
			loc.ConnectionType = connectionType.String
		}
		if trigger.Valid {
			loc.Trigger = trigger.String
		}

		locations = append(locations, loc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return locations, nil
}

// GetLocationsByDateRange retrieves GPS locations within a date range
func (c *Client) GetLocationsByDateRange(ctx context.Context, startDate, endDate string, deviceID string) ([]Location, error) {
	query := `
		SELECT
			id, device_id, tid, latitude, longitude, accuracy,
			altitude, velocity, battery, battery_status,
			connection_type, trigger, EXTRACT(EPOCH FROM timestamp)::bigint AS timestamp, created_at
		FROM public.locations
		WHERE created_at >= $1::date AND created_at < $2::date + interval '1 day'
	`

	args := []interface{}{startDate, endDate}

	if deviceID != "" {
		query += " AND device_id = $3"
		args = append(args, deviceID)
	}

	query += " ORDER BY created_at ASC"

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer func() { _ = rows.Close() }() // nolint:errcheck // Close in defer, error not actionable

	var locations []Location
	for rows.Next() {
		var loc Location
		var accuracy, altitude, velocity, battery, timestamp sql.NullInt64
		var batteryStatus, connectionType, trigger sql.NullString

		err := rows.Scan(
			&loc.ID,
			&loc.DeviceID,
			&loc.TID,
			&loc.Latitude,
			&loc.Longitude,
			&accuracy,
			&altitude,
			&velocity,
			&battery,
			&batteryStatus,
			&connectionType,
			&trigger,
			&timestamp,
			&loc.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		// Convert NULL values to zero values
		if accuracy.Valid {
			loc.Accuracy = int(accuracy.Int64)
		}
		if altitude.Valid {
			loc.Altitude = int(altitude.Int64)
		}
		if velocity.Valid {
			loc.Velocity = int(velocity.Int64)
		}
		if battery.Valid {
			loc.Battery = int(battery.Int64)
		}
		if timestamp.Valid {
			loc.Timestamp = timestamp.Int64
		}
		if batteryStatus.Valid {
			loc.BatteryStatus = batteryStatus.String
		}
		if connectionType.Valid {
			loc.ConnectionType = connectionType.String
		}
		if trigger.Valid {
			loc.Trigger = trigger.String
		}

		locations = append(locations, loc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return locations, nil
}

// GetDevices returns a list of unique device IDs from the database
func (c *Client) GetDevices(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT device_id
		FROM public.locations
		WHERE device_id IS NOT NULL
		ORDER BY device_id
	`

	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer func() { _ = rows.Close() }() // nolint:errcheck // Close in defer, error not actionable

	var devices []string
	for rows.Next() {
		var deviceID string
		if err := rows.Scan(&deviceID); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		devices = append(devices, deviceID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return devices, nil
}

// GetLocationCount returns the count of locations for a specific date
func (c *Client) GetLocationCount(ctx context.Context, date string, deviceID string) (int, error) {
	query := `SELECT COUNT(*) FROM public.locations WHERE DATE(created_at) = $1`
	args := []interface{}{date}

	if deviceID != "" {
		query += " AND device_id = $2"
		args = append(args, deviceID)
	}

	var count int
	err := c.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count query failed: %w", err)
	}

	return count, nil
}

// HealthCheck verifies database connectivity
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.db.PingContext(ctx)
}
