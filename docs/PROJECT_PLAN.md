# otel-worker Project Plan

## Project Overview

**Name**: otel-worker
**Type**: Go gRPC Microservice
**Purpose**: Calculate distance-from-home metrics using OwnTracks GPS data
**Started**: 2026-01-24
**Target Cluster**: k8s-pi5-cluster
**Home Location**: 40.736097°N, 74.039373°W
**Database**: PostgreSQL (192.168.1.175:5432/owntracks)

## Implementation Roadmap

### Phase 1: Investigation & Setup ⏳ IN PROGRESS

#### Database Investigation

- [ ] Connect to owntracks database (192.168.1.175:5432)
- [ ] Query `public.locations` table schema
- [ ] Analyze sample data (column names, data types, indexes)
- [ ] Determine device_id/user_id handling strategy
- [ ] Test query performance with date filters
- [ ] Document actual schema in DATABASE.md

#### Development Environment

- [x] Create project directory structure
- [x] Write setup.sh script for Go SDK installation
- [x] Create Makefile with build targets
- [x] Create .env.example with configuration
- [x] Create .gitignore
- [ ] Initialize go.mod with module path
- [ ] Write README.md with quick start guide
- [ ] Write project documentation

### Phase 2: Core Implementation 🔜 NEXT

#### gRPC API Definition

- [ ] Define proto/distance/v1/distance.proto
- [ ] Create DistanceService with CalculateDistanceFromHome RPC
- [ ] Define request/response message structures
- [ ] Add GetJobStatus RPC for async polling
- [ ] Add ListJobs RPC for job management
- [ ] Generate Go code with protoc
- [ ] Document API in docs/API.md

**Database Layer** `internal/database/`

- [ ] Implement client.go (PostgreSQL connection pool)
- [ ] Create Config struct with connection parameters
- [ ] Implement Connect() with retry logic
- [ ] Create query functions for location data retrieval
- [ ] Add connection health check
- [ ] Implement graceful connection closing
- [ ] Write unit tests with mock database

**Distance Calculator** `internal/calculator/`

- [ ] Implement haversine.go with distance formula
- [ ] Create DistanceFromHome() function
- [ ] Implement trips.go for trip detection logic
- [ ] Create trip segmentation (departure/return detection)
- [ ] Add data quality filters (accuracy, deduplication)
- [ ] Implement csv.go for file generation
- [ ] Calculate cumulative away-time metrics
- [ ] Write unit tests with known coordinates

**Configuration** `internal/config/`

- [ ] Implement config.go with environment loading
- [ ] Create Config struct with all settings
- [ ] Add validation for required fields
- [ ] Support .env file loading (godotenv)
- [ ] Add default values
- [ ] Implement getters for configuration

**Job Queue System** `internal/queue/`

- [ ] Implement queue.go with channel-based queue
- [ ] Create Job struct with status tracking
- [ ] Implement worker pool (configurable goroutines)
- [ ] Add job status tracking (sync.Map)
- [ ] Implement graceful shutdown with context
- [ ] Add queue metrics (pending, processing, completed)
- [ ] Handle job failures and retries

**gRPC Server** `internal/grpc/`

- [ ] Implement handler.go with service implementation
- [ ] Create CalculateDistanceFromHome handler
- [ ] Create GetJobStatus handler
- [ ] Create ListJobs handler
- [ ] Add gRPC middleware for logging
- [ ] Add OpenTelemetry instrumentation
- [ ] Implement error handling and status codes

**Main Application** `cmd/server/`

- [ ] Implement main.go with initialization
- [ ] Load configuration from environment
- [ ] Set up database connection
- [ ] Initialize job queue and workers
- [ ] Start gRPC server on port 50051
- [ ] Start HTTP health server on port 8080
- [ ] Handle OS signals for graceful shutdown
- [ ] Add structured logging (zerolog)

### Phase 3: Testing & Quality 📝 PLANNED

#### Unit Tests

- [ ] Test haversine distance calculations
- [ ] Test trip detection logic
- [ ] Test data quality filters
- [ ] Test CSV generation
- [ ] Test job queue operations
- [ ] Test configuration loading
- [ ] Achieve 80%+ code coverage

#### Integration Tests

- [ ] Test database queries with test database
- [ ] Test gRPC handlers end-to-end
- [ ] Test concurrent job processing
- [ ] Test file output to temporary volume
- [ ] Test error scenarios (DB down, invalid input)
- [ ] Test graceful shutdown

#### Performance Testing

- [ ] Benchmark distance calculations (1000+ points)
- [ ] Test concurrent request handling (10+ jobs)
- [ ] Measure memory usage with large datasets
- [ ] Test CSV file size limits
- [ ] Profile CPU usage under load
- [ ] Optimize database queries

#### Code Quality

- [ ] Set up golangci-lint configuration
- [ ] Add pre-commit hooks
- [ ] Format code with gofmt
- [ ] Run go vet for issues
- [ ] Check for race conditions (go test -race)
- [ ] Run static analysis

### Phase 4: Containerization & Deployment 🚀 PLANNED

#### Docker

- [ ] Create multi-stage Dockerfile
- [ ] Use golang:1.23-alpine for builder
- [ ] Use distroless/static for runtime
- [ ] Optimize image size (< 20MB)
- [ ] Add non-root user for security
- [ ] Build and test locally
- [ ] Push to Docker Hub (stuartshay/otel-worker)

#### GitHub Actions

- [ ] Create .github/workflows/lint.yml
- [ ] Create .github/workflows/test.yml
- [ ] Create .github/workflows/docker.yml
- [ ] Add version tagging on merge to main
- [ ] Configure Docker Hub credentials
- [ ] Add build status badges

#### Kubernetes Manifests

- [ ] Create apps/base/otel-worker/deployment.yaml
- [ ] Create apps/base/otel-worker/service.yaml
- [ ] Create apps/base/otel-worker/configmap.yaml
- [ ] Create apps/base/otel-worker/secret.yaml
- [ ] Create apps/base/otel-worker/pvc.yaml
- [ ] Create apps/base/otel-worker/kustomization.yaml
- [ ] Add cluster overlay in clusters/k8s-pi5-cluster/
- [ ] Configure resource limits (200m CPU, 512Mi memory)
- [ ] Add health/readiness probes

#### GitOps Integration

- [ ] Add otel-worker to k8s-gitops repository
- [ ] Configure Argo CD application
- [ ] Set sync policy (automated)
- [ ] Test deployment to k8s-pi5-cluster
- [ ] Verify PVC mount and CSV file creation
- [ ] Check logs for errors
- [ ] Test gRPC connectivity

### Phase 5: Observability & Monitoring 📊 PLANNED

#### OpenTelemetry Integration

- [ ] Add OTel SDK dependency
- [ ] Instrument gRPC server with otelgrpc
- [ ] Trace database queries
- [ ] Trace distance calculations
- [ ] Export to otel-collector
- [ ] Configure trace sampling (1.0 for dev)

#### Metrics

- [ ] Expose Prometheus metrics endpoint (/metrics)
- [ ] Add custom metrics (jobs_total, jobs_duration_seconds)
- [ ] Track database query duration
- [ ] Monitor CSV file sizes
- [ ] Track queue depth and worker utilization
- [ ] Add histogram for distance calculations

#### Logging

- [ ] Implement structured logging (JSON format)
- [ ] Include trace_id in all logs
- [ ] Log levels: DEBUG, INFO, WARN, ERROR
- [ ] Log to stdout (captured by Kubernetes)
- [ ] Configure log sampling for high-volume logs
- [ ] Add contextual fields (job_id, date, etc.)

#### Health Checks

- [ ] Implement /healthz (liveness probe)
- [ ] Implement /readyz (readiness probe)
- [ ] Check database connectivity
- [ ] Check queue health
- [ ] Return proper HTTP status codes

### Phase 6: Integration & Documentation 📖 PLANNED

#### otel-demo API Integration

- [ ] Add gRPC client to otel-demo Flask app
- [ ] Create REST endpoint: POST /api/distance/calculate
- [ ] Implement async job polling
- [ ] Return CSV download link
- [ ] Add authentication (Cognito JWT validation)
- [ ] Add rate limiting

#### otel-ui Frontend

- [ ] Create distance calculation UI component
- [ ] Add date picker for calculation
- [ ] Display job status (pending/processing/completed)
- [ ] Show summary metrics (max distance, trips)
- [ ] Provide CSV download button
- [ ] Add distance visualization (map/chart)
- [ ] Add historical data view

#### Documentation

- [x] Complete PROJECT_PLAN.md (this file)
- [ ] Write FUNCTIONALITY.md with business logic
- [ ] Write API.md with gRPC examples
- [ ] Write DATABASE.md with schema and queries
- [ ] Write DEPLOYMENT.md with kubectl commands
- [ ] Write DEVELOPMENT.md with local setup
