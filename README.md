# otel-worker

[![Lint](https://github.com/stuartshay/otel-worker/actions/workflows/lint.yml/badge.svg)](https://github.com/stuartshay/otel-worker/actions/workflows/lint.yml)
[![Test](https://github.com/stuartshay/otel-worker/actions/workflows/test.yml/badge.svg)](https://github.com/stuartshay/otel-worker/actions/workflows/test.yml)
[![Docker](https://github.com/stuartshay/otel-worker/actions/workflows/docker.yml/badge.svg)](https://github.com/stuartshay/otel-worker/actions/workflows/docker.yml)

Go-based gRPC microservice that calculates distance-from-home metrics using OwnTracks GPS data.

## Overview

otel-worker processes GPS location data from an OwnTracks database and calculates
distance-based metrics from a fixed home location. It provides a gRPC API for async job
processing and generates CSV reports with detailed location and distance information.

## Features

- 🚀 **gRPC API** for async distance calculations
- 📊 **PostgreSQL** integration with OwnTracks database
- 📐 **Haversine formula** for accurate GPS distance calculations
- 🚶 **Trip detection** with configurable away-from-home thresholds
- 📝 **CSV output** with detailed metrics
- ⚙️ **Concurrent processing** with worker pool and job queue
- 📡 **OpenTelemetry** instrumentation for observability
- ☸️ **Kubernetes-ready** deployment

## Quick Start

### Prerequisites

- Go 1.24+
- PostgreSQL with OwnTracks data (PgBouncer recommended)
- Docker (optional)
- protoc 3.12+ with Go plugins (for protobuf generation)

### Installation

```bash
# Clone repository
git clone https://github.com/stuartshay/otel-worker.git
cd otel-worker

# Run setup script
./setup.sh

# Copy and configure environment
cp .env.example .env
# Edit .env with your database credentials
```

### Development

```bash
# Generate protobuf code
make proto

# Build
make build

# Run locally
make run

# Run tests
make test
```

## Configuration

See [`.env.example`](.env.example) for all configuration options.

Key settings:

- `HOME_LATITUDE`, `HOME_LONGITUDE` - Reference point for distance calculations
- `POSTGRES_*` - Database connection details
- `AWAY_THRESHOLD_KM` - Distance threshold for trip detection (default: 0.5 km)

## API

gRPC service on port **50051**:

- `CalculateDistanceFromHome` - Submit calculation job for a date
- `GetJobStatus` - Poll job status and retrieve results
- `ListJobs` - List all jobs

Health checks on port **8080**:

- `GET /healthz` - Liveness probe
- `GET /readyz` - Readiness probe

## Output

CSV files: `distance_YYYYMMDD.csv`

Columns:

- timestamp
- latitude, longitude
- distance_from_home_km
- is_away (boolean)
- accuracy
- cumulative_away_time_minutes

## Documentation

- [PROJECT_PLAN.md](docs/PROJECT_PLAN.md) - Implementation roadmap
- [DATABASE.md](docs/DATABASE.md) - Database schema and query patterns
- [DOCKER_BUILD.md](docs/DOCKER_BUILD.md) - Docker build and deployment guide
- [GITHUB_ACTIONS.md](docs/GITHUB_ACTIONS.md) - CI/CD workflows documentation
- [AGENTS.md](AGENTS.md) - Quick reference for automation/developers
- [FUNCTIONALITY.md](docs/FUNCTIONALITY.md) - Business logic specification (coming soon)

## Deployment

Docker image: **stuartshay/otel-worker:latest** (14.9 MB)

See [DOCKER_BUILD.md](docs/DOCKER_BUILD.md) for build instructions.

Kubernetes manifests are managed in the [k8s-gitops](https://github.com/stuartshay/k8s-gitops) repository.

```bash
# Build Docker image
make docker-build

# Push to Docker Hub
make docker-push

# Run with Docker
docker run -p 50051:50051 -p 8080:8080 --env-file .env stuartshay/otel-worker:latest
```

## Related Projects

- [otel-demo](https://github.com/stuartshay/otel-demo) - Flask API backend
- [otel-ui](https://github.com/stuartshay/otel-ui) - React frontend
- [k8s-gitops](https://github.com/stuartshay/k8s-gitops) - Kubernetes GitOps manifests

## License

MIT

## Status

✅ **Phase 1 Complete** - Database integration and development environment
✅ **Phase 2 Complete** - Core implementation (config, calculator, database, queue, gRPC, server)
⏳ **Phase 3 In Progress** - Testing & quality (Docker containerization complete)
📋 **Phase 4 Planned** - GitHub Actions CI/CD and Kubernetes deployment
