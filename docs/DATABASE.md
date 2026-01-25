# Database Schema Documentation

## Overview

otel-worker connects to the OwnTracks PostgreSQL database on Node-5 via PgBouncer
connection pooler for efficient connection management.

## Connection Details

| Property | Value |
|----------|-------|
| Host | 192.168.1.175 |
| Database | owntracks |
| User | development |
| Password | development |
| Direct Port | 5432 (PostgreSQL) |
| Pooled Port | 6432 (PgBouncer) |

**⚠️ Important**: Always use PgBouncer port (6432) for application connections.
Direct PostgreSQL connections (5432) should only be used for migrations and admin tasks.

## Schema: public.locations

The `public.locations` table stores GPS tracking data from OwnTracks devices.

### Table Structure

```sql
CREATE TABLE public.locations (
    id               SERIAL PRIMARY KEY,
    device_id        VARCHAR,
    tid              VARCHAR,
    latitude         DOUBLE PRECISION,
    longitude        DOUBLE PRECISION,
    accuracy         INTEGER,
    altitude         INTEGER,
    velocity         INTEGER,
    battery          INTEGER,
    battery_status   VARCHAR,
    connection_type  VARCHAR,
    trigger          VARCHAR,
    timestamp        TIMESTAMP WITH TIME ZONE,
    created_at       TIMESTAMP,
    raw_payload      JSONB
);
```

### Column Descriptions

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL | Auto-incrementing primary key |
| device_id | VARCHAR | OwnTracks device identifier (e.g., "iphone", "android") |
| tid | VARCHAR | Tracker ID (2-character device abbreviation) |
| latitude | DOUBLE PRECISION | GPS latitude in decimal degrees |
| longitude | DOUBLE PRECISION | GPS longitude in decimal degrees |
| accuracy | INTEGER | GPS accuracy in meters |
| altitude | INTEGER | Altitude in meters above sea level |
| velocity | INTEGER | Speed in km/h |
| battery | INTEGER | Battery level percentage (0-100) |
| battery_status | VARCHAR | Battery status (e.g., "charging", "full", "unplugged") |
| connection_type | VARCHAR | Network connection type (e.g., "wifi", "mobile") |
| trigger | VARCHAR | Event trigger type (e.g., "manual", "timer", "region") |
| timestamp | TIMESTAMP WITH TIME ZONE | Location event timestamp (extracted as Unix epoch in queries) |
| created_at | TIMESTAMP | Database insertion timestamp |
| raw_payload | JSONB | Original OwnTracks JSON payload |

### Query Examples

#### Get locations for a specific date

```sql
SELECT
    id,
    device_id,
    latitude,
    longitude,
    accuracy,
    timestamp,
    created_at
FROM public.locations
WHERE DATE(created_at) = '2026-01-24'
ORDER BY created_at ASC;
```

#### Get latest location per device

```sql
SELECT DISTINCT ON (device_id)
    device_id,
    latitude,
    longitude,
    created_at
FROM public.locations
ORDER BY device_id, created_at DESC;
```

#### Get locations within date range

```sql
SELECT *
FROM public.locations
WHERE created_at >= '2026-01-24 00:00:00'
  AND created_at < '2026-01-25 00:00:00'
ORDER BY created_at ASC;
```

### Indexes

(To be determined - check for existing indexes on device_id, created_at, timestamp)

Recommended indexes for otel-worker queries:

```sql
-- For date-based filtering
CREATE INDEX idx_locations_created_at ON public.locations (created_at);

-- For device filtering
CREATE INDEX idx_locations_device_id ON public.locations (device_id);

-- Composite index for common queries
CREATE INDEX idx_locations_device_created ON public.locations (device_id, created_at);
```

## Data Samples

### Sample Row

```text
id:              123456
device_id:       iphone
tid:             ip
latitude:        40.748817
longitude:       -73.985428
accuracy:        10
altitude:        15
velocity:        0
battery:         85
battery_status:  unplugged
connection_type: wifi
trigger:         timer
timestamp:       1737720000
created_at:      2026-01-24 07:00:00
raw_payload:     {"_type":"location","acc":10,"alt":15,...}
```

## Distance Calculation

otel-worker calculates distance from a fixed home location:

**Home Coordinates**: 40.736097°N, 74.039373°W

The Haversine formula is used to calculate great-circle distance between
two points on a sphere given their longitudes and latitudes.

### Distance Metrics

For each date, the service calculates:

- **Total Distance (km)**: Sum of distances from home for all location points
- **Max Distance (km)**: Furthest point from home
- **Min Distance (km)**: Closest point to home
- **Total Locations**: Number of GPS records processed

## Performance Considerations

- Use PgBouncer (port 6432) for connection pooling
- Filter by `created_at` instead of `timestamp` (indexed)
- Limit queries to specific date ranges
- Use prepared statements for repeated queries
- Consider batching for large date ranges

## Data Volume Estimates

(To be determined based on actual usage patterns)

- Average records per day: TBD
- Average records per device per day: TBD
- Data retention period: TBD
