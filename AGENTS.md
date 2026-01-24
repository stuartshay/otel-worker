# Agent Operating Guide

All automation, assistants, and developers must follow
`.github/copilot-instructions.md` for workflow, safety, and environment rules.

## How to Use

- Read `.github/copilot-instructions.md` before making changes
- Apply every rule in that file as-is; do not redefine or override them here
- If updates are needed, edit `.github/copilot-instructions.md` and keep this
  file pointing to it

## Quick Reference

- **Language**: Go 1.23+
- **Protocol**: gRPC + Protobuf
- **Database**: PostgreSQL via PgBouncer (192.168.1.175:6432)
- **Linting**: `make pre-commit-run` or `pre-commit run --all-files`
- **Build**: `make build` or `go build -o bin/otel-worker cmd/server/main.go`
- **Test**: `make test` or `go test ./...`
- **Docker**: `make docker-build` or `docker build -t stuartshay/otel-worker .`

## Project Structure

```text
otel-worker/
├── cmd/server/          # Application entrypoint
├── internal/            # Private application code
│   ├── calculator/      # Distance calculations (Haversine)
│   ├── config/          # Configuration management
│   ├── database/        # PostgreSQL client
│   ├── grpc/            # gRPC service handlers
│   └── queue/           # Job queue (5-worker pool)
├── proto/distance/v1/   # gRPC protobuf definitions
└── docs/                # Documentation
```

## Development Workflow

1. Make changes to Go source files in `internal/` or `cmd/`
2. Run unit tests: `go test ./...`
3. Run lint checks: `pre-commit run --all-files`
4. Build binary: `go build -o bin/otel-worker cmd/server/main.go`
5. Test locally (requires `.env` with DB credentials)
6. Commit with descriptive message
7. Push to trigger CI/CD pipeline

## gRPC Service

**DistanceService** (port 50051):

- `CalculateDistanceFromHome(date, device_id)` → Returns job_id
- `GetJobStatus(job_id)` → Returns status, result with CSV path and metrics
- `ListJobs(status, limit, offset)` → Returns job list with pagination

**Home Location**: 40.736097°N, 74.039373°W (used for all distance calculations)

## Database Schema

```sql
-- public.locations (owntracks database)
SELECT id, device_id, latitude, longitude, accuracy,
       battery, velocity, timestamp, created_at
FROM public.locations
WHERE DATE(created_at) = '2026-01-24'
ORDER BY created_at ASC;
```

See [docs/DATABASE.md](docs/DATABASE.md) for full schema documentation.

## CSV Output Format

Generated files: `distance_YYYYMMDD.csv` or `distance_YYYYMMDD_<device_id>.csv`

Columns: timestamp, device_id, latitude, longitude, distance_from_home_km,
accuracy, battery, velocity

Summary footer: Total Distance, Max Distance, Min Distance, Total Locations,
Average Distance
