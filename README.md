# otel-worker

[![Lint](https://github.com/stuartshay/otel-worker/actions/workflows/lint.yml/badge.svg)](https://github.com/stuartshay/otel-worker/actions/workflows/lint.yml)
[![Test](https://github.com/stuartshay/otel-worker/actions/workflows/test.yml/badge.svg)](https://github.com/stuartshay/otel-worker/actions/workflows/test.yml)
[![Docker](https://github.com/stuartshay/otel-worker/actions/workflows/docker.yml/badge.svg)](https://github.com/stuartshay/otel-worker/actions/workflows/docker.yml)
[![Docker Hub](https://img.shields.io/badge/Docker%20Hub-stuartshay%2Fotel--worker-blue?logo=docker)](https://hub.docker.com/repository/docker/stuartshay/otel-worker)
[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Renovate](https://img.shields.io/badge/renovate-enabled-brightgreen.svg?logo=renovatebot)](https://renovatebot.com)

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

# Run pre-commit checks
make pre-commit-run
```

## Configuration

See [`.env.example`](.env.example) for all configuration options.

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `HOME_LATITUDE` | - | Reference latitude for distance calculations |
| `HOME_LONGITUDE` | - | Reference longitude for distance calculations |
| `POSTGRES_HOST` | `localhost` | PostgreSQL host |
| `POSTGRES_PORT` | `5432` | PostgreSQL port |
| `POSTGRES_DB` | `owntracks` | Database name |
| `POSTGRES_USER` | - | Database username |
| `POSTGRES_PASSWORD` | - | Database password |
| `AWAY_THRESHOLD_KM` | `0.5` | Distance threshold for trip detection |
| `GRPC_PORT` | `50051` | gRPC server port |
| `HTTP_PORT` | `8080` | HTTP health check port |

## API

### gRPC Service (port 50051)

| Method | Description |
| ------ | ----------- |
| `CalculateDistanceFromHome` | Submit calculation job for a date |
| `GetJobStatus` | Poll job status and retrieve results |
| `ListJobs` | List all jobs |

### Health Checks (port 8080)

| Endpoint | Description |
| -------- | ----------- |
| `GET /healthz` | Liveness probe |
| `GET /readyz` | Readiness probe |

## Output

CSV files: `distance_YYYYMMDD.csv`

| Column | Description |
| ------ | ----------- |
| `timestamp` | Location timestamp |
| `latitude` | GPS latitude |
| `longitude` | GPS longitude |
| `distance_from_home_km` | Distance from home in kilometers |
| `is_away` | Boolean indicating if outside threshold |
| `accuracy` | GPS accuracy in meters |
| `cumulative_away_time_minutes` | Total time away from home |

## Docker

Docker image: **stuartshay/otel-worker:latest** (14.9 MB)

```bash
# Build Docker image
make docker-build

# Run with Docker
docker run -p 50051:50051 -p 8080:8080 --env-file .env stuartshay/otel-worker:latest
```

## Kubernetes

Kubernetes manifests are managed in the [k8s-gitops](https://github.com/stuartshay/k8s-gitops) repository.

| Resource | Value |
| -------- | ----- |
| Namespace | `otel-worker` |
| Docker Hub | `stuartshay/otel-worker` |
| gRPC Port | 50051 |
| HTTP Port | 8080 |

## Documentation

- [PROJECT_PLAN.md](docs/PROJECT_PLAN.md) - Implementation roadmap
- [DATABASE.md](docs/DATABASE.md) - Database schema and query patterns
- [DOCKER_BUILD.md](docs/DOCKER_BUILD.md) - Docker build and deployment guide
- [GITHUB_ACTIONS.md](docs/GITHUB_ACTIONS.md) - CI/CD workflows documentation
- [AGENTS.md](AGENTS.md) - Quick reference for automation/developers

## Related Projects

| Project | Description |
| ------- | ----------- |
| [otel-demo](https://github.com/stuartshay/otel-demo) | Flask API backend |
| [otel-ui](https://github.com/stuartshay/otel-ui) | React frontend |
| [k8s-gitops](https://github.com/stuartshay/k8s-gitops) | Kubernetes GitOps manifests |

## License

MIT
