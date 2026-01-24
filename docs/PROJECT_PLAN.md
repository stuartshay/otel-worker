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

### Phase 1: Investigation & Setup ✅ COMPLETE

#### Database Investigation

- [x] Connect to owntracks database (192.168.1.175:6432 via PgBouncer)
- [x] Query `public.locations` table schema
- [x] Analyze sample data (column names, data types, indexes)
- [x] Determine device_id/user_id handling strategy
- [x] Test query performance with date filters
- [x] Document actual schema in DATABASE.md

#### Development Environment

- [x] Create project directory structure
- [x] Write setup.sh script for Go SDK installation
- [x] Create Makefile with build targets
- [x] Create .env.example with configuration
- [x] Create .gitignore
- [x] Initialize go.mod with module path
- [x] Write README.md with quick start guide
- [x] Write project documentation

### Phase 2: Core Implementation ✅ COMPLETE

#### gRPC API Definition

- [x] Define proto/distance/v1/distance.proto
- [x] Create DistanceService with CalculateDistanceFromHome RPC
- [x] Define request/response message structures
- [x] Add GetJobStatus RPC for async polling
- [x] Add ListJobs RPC for job management
- [x] Generate Go code with protoc
- [x] Document API in protobuf comments

#### Database Layer `internal/database/`

- [x] Implement client.go (PostgreSQL connection pool)
- [x] Create Config struct with connection parameters
- [x] Implement Connect() with retry logic
- [x] Create query functions for location data retrieval
- [x] Add connection health check
- [x] Implement graceful connection closing
- [x] Write unit tests with mock database

#### Distance Calculator `internal/calculator/`

- [x] Implement haversine.go with distance formula
- [x] Create DistanceFromHome() function
- [x] Implement CalculateMetrics() for distance statistics
- [x] Add data quality filters (accuracy, deduplication)
- [x] Calculate distance metrics (total, max, min, avg)
- [x] Write unit tests with known coordinates
- [x] Add benchmark tests for performance

#### Configuration `internal/config/`

- [x] Implement config.go with environment loading
- [x] Create Config struct with all settings
- [x] Add validation for required fields
- [x] Support .env file loading (godotenv)
- [x] Add default values
- [x] Implement DatabaseDSN() for connection string
- [x] Write unit tests for config loading

#### Job Queue System `internal/queue/`

- [x] Implement queue.go with channel-based queue
- [x] Create Job struct with status tracking
- [x] Implement worker pool (5 goroutines)
- [x] Add job status tracking (sync.Map)
- [x] Implement graceful shutdown with context
- [x] Add queue statistics (total, queued, processing, completed, failed)
- [x] Handle job failures and error messages
- [x] Write comprehensive unit tests

#### gRPC Server `internal/grpc/`

- [x] Implement handler.go with service implementation
- [x] Create CalculateDistanceFromHome handler
- [x] Create GetJobStatus handler
- [x] Create ListJobs handler
- [x] Add CSV generation with distance metrics
- [x] Implement job processor integration
- [x] Add structured logging with zerolog
- [x] Implement error handling and status codes

#### Main Application `cmd/server/`

- [x] Implement main.go with initialization
- [x] Load configuration from environment
- [x] Set up database connection
- [x] Initialize job queue and workers
- [x] Start gRPC server on port 50051
- [x] Start HTTP health server on port 8080
- [x] Handle OS signals for graceful shutdown
- [x] Add structured logging (zerolog)

### Phase 3: Testing & Quality ⏳ IN PROGRESS

#### Unit Tests

- [x] Test haversine distance calculations
- [ ] Test trip detection logic (not implemented - simplified design)
- [ ] Test data quality filters
- [ ] Test CSV generation
- [x] Test job queue operations
- [x] Test configuration loading
- [ ] Achieve 80%+ code coverage (partial - calculator, config, database, queue tested)

#### Integration Tests

- [ ] Test database queries with test database
- [ ] Test gRPC handlers end-to-end
- [ ] Test concurrent job processing
- [ ] Test file output to temporary volume
- [ ] Test error scenarios (DB down, invalid input)
- [ ] Test graceful shutdown

#### Performance Testing

- [x] Benchmark distance calculations (1000+ points)
- [ ] Test concurrent request handling (10+ jobs)
- [ ] Measure memory usage with large datasets
- [ ] Test CSV file size limits
- [ ] Profile CPU usage under load
- [ ] Optimize database queries

#### Code Quality

- [x] Set up golangci-lint configuration
- [x] Add pre-commit hooks
- [x] Format code with gofmt
- [x] Run go vet for issues
- [ ] Check for race conditions (go test -race)
- [ ] Run static analysis

### Phase 4: Containerization & Deployment ⏳ IN PROGRESS

#### Docker

- [x] Create multi-stage Dockerfile
- [x] Use golang:1.24-alpine for builder
- [x] Use distroless/static for runtime
- [x] Optimize image size (< 20MB) - achieved 14.9MB
- [x] Add non-root user for security
- [x] Build and test locally
- [ ] Push to Docker Hub (stuartshay/otel-worker)

#### GitHub Actions

- [x] Create .github/workflows/lint.yml (Go, Markdown, YAML, Dockerfile, Protobuf)
- [x] Create .github/workflows/test.yml (unit tests, integration tests with PostgreSQL)
- [x] Create .github/workflows/docker.yml (multi-arch builds, Trivy scanning)
- [x] Add version tagging on merge to main
- [x] Add Trivy vulnerability scanning
- [x] Add database schema initialization for integration tests
- [ ] Configure Docker Hub credentials (secrets: DOCKERHUB_USERNAME, DOCKERHUB_TOKEN)
- [ ] Add build status badges to README

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
