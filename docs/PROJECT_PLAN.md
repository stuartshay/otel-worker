# otel-worker Project Plan

## Project Overview

**Name**: otel-worker
**Type**: Go gRPC Microservice
**Purpose**: Calculate distance-from-home metrics using OwnTracks GPS data
**Started**: 2026-01-24
**Last Updated**: 2026-01-25
**Target Cluster**: k8s-pi5-cluster
**Home Location**: 40.736097°N, 74.039373°W
**Database**: PostgreSQL (192.168.1.175:6432/owntracks via PgBouncer)
**Docker Hub**: stuartshay/otel-worker
**Buf Registry**: stuartshay-consulting/otel-worker

## Current Status

**Overall Progress**: ~55% complete

| Phase | Status | Completion |
|-------|--------|------------|
| Phase 1: Investigation & Setup | ✅ COMPLETE | 100% |
| Phase 2: Core Implementation | ✅ COMPLETE | 100% |
| Phase 3: Testing & Quality | 🟡 IN PROGRESS | 50% |
| Phase 4: Containerization & Deployment | 🟡 IN PROGRESS | 50% |
| Phase 5: Observability & Monitoring | ❌ PLANNED | 0% |
| Phase 6: Integration & Documentation | ❌ PLANNED | 0% |

**Next Immediate Task**: Implement health check endpoints (30 min) - **Blocks K8s deployment**

### Recent Accomplishments (2026-01-25)

- ✅ Fixed all 27 golangci-lint errors
- ✅ Added semantic versioning to Docker builds (1.0.{BUILD_NUMBER})
- ✅ Fixed pre-commit hooks (converted to local)
- ✅ Added zsh support to setup.sh
- ✅ All GitHub Actions workflows passing
- ✅ Docker image optimized to 14.9MB

### Critical Path Forward

1. **Now**: Health checks (30 min) ← **YOU ARE HERE**
2. **Next**: gRPC handler tests (2-3 hrs)
3. **Then**: K8s manifests (2 hrs)
4. **Finally**: Deploy to k8s-pi5-cluster (1 hr)

See [QUICK_WINS.md](QUICK_WINS.md) for detailed task breakdown.

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

### Phase 3: Testing & Quality ⏳ IN PROGRESS (50% complete)

**Last Updated**: 2026-01-25

#### Unit Tests

- [x] Test haversine distance calculations (100% coverage)
- [x] Test job queue operations (92% coverage)
- [x] Test configuration loading (100% coverage)
- [x] Test database client basic operations (13.3% coverage - needs expansion)
- [ ] Test trip detection logic (not implemented - simplified design)
- [ ] Test CSV generation
- [ ] **HIGH PRIORITY**: Test gRPC handlers (currently 0% coverage)
- [ ] Achieve 80%+ code coverage (current: ~50% overall)

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

### Phase 4: Containerization & Deployment ⏳ IN PROGRESS (50% complete)

**Last Updated**: 2026-01-25

#### Docker

- [x] Create multi-stage Dockerfile
- [x] Use golang:1.24-alpine for builder
- [x] Use distroless/static for runtime
- [x] Optimize image size (< 20MB) - **achieved 14.9MB** ✨
- [x] Add non-root user for security
- [x] Build and test locally
- [x] Configure multi-arch builds (amd64/arm64)
- [x] Add semantic versioning: `1.0.{BUILD_NUMBER}`
- [ ] **HIGH PRIORITY**: Push to Docker Hub (stuartshay/otel-worker) - repo exists, needs images

#### GitHub Actions

- [x] Create .github/workflows/lint.yml (Go, Markdown, YAML, Dockerfile, Protobuf) ✅
- [x] Create .github/workflows/test.yml (unit tests, integration tests with PostgreSQL) ✅
- [x] Create .github/workflows/docker.yml (multi-arch builds, Trivy scanning) ✅
- [x] Create .github/workflows/proto-publish.yml (Buf Schema Registry) ✅
- [x] Add version tagging on merge to main (semantic versioning)
- [x] Add Trivy vulnerability scanning
- [x] Add database schema initialization for integration tests
- [x] Fix all golangci-lint errors (27 errors resolved) ✨
- [x] Configure local pre-commit hooks for reliability
- [ ] Configure Docker Hub credentials (secrets: DOCKERHUB_USERNAME, DOCKERHUB_TOKEN)
- [x] Add build status badges to README

**Status**: All workflows passing ✅

#### Kubernetes Manifests

**Status**: Not started - **HIGH PRIORITY** 🚨
**Repository**: k8s-gitops (k8s-pi5-cluster)
**Reference**: Follow patterns from otel-demo and otel-ui (already deployed)

- [ ] Create apps/base/otel-worker/deployment.yaml
- [ ] Create apps/base/otel-worker/service.yaml
- [ ] Create apps/base/otel-worker/configmap.yaml
- [ ] Create apps/base/otel-worker/secret.yaml (SOPS-encrypted)
- [ ] Create apps/base/otel-worker/pvc.yaml
- [ ] Create apps/base/otel-worker/kustomization.yaml
- [ ] Add cluster overlay in clusters/k8s-pi5-cluster/
- [ ] Configure resource limits (requests: 100m CPU, 256Mi; limits: 500m CPU, 512Mi)
- [ ] Add health/readiness probes (requires health check endpoints first)

**Blockers**: Health check endpoints not implemented yet (see Phase 5)

#### GitOps Integration

- [ ] Add otel-worker to k8s-gitops repository
- [ ] Configure Argo CD application
- [ ] Set sync policy (automated)
- [ ] Test deployment to k8s-pi5-cluster
- [ ] Verify PVC mount and CSV file creation
- [ ] Check logs for errors
- [ ] Test gRPC connectivity

### Phase 5: Observability & Monitoring 📊 PLANNED

**Status**: Not started - Health checks are **HIGHEST PRIORITY** 🚨
**Blocker**: Health endpoints required for Kubernetes deployment

#### Health Checks - **NEXT IMMEDIATE TASK**

- [ ] **HIGH PRIORITY**: Implement /healthz (liveness probe)
- [ ] **HIGH PRIORITY**: Implement /readyz (readiness probe)
- [ ] Start HTTP server on port 8080
- [ ] Check database connectivity for readiness
- [ ] Check queue health
- [ ] Return proper HTTP status codes (200, 503)
- [ ] Add graceful HTTP server shutdown
- [ ] Document in README.md

**Estimated Time**: 30 minutes
**Impact**: Unblocks Kubernetes deployment (Phase 4)

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
### Phase 7: Database Migrations & Multi-Source GPS 🗄️ PLANNED

#### Migration Repository Reference

**Repository**: [home lab-database-migrations](https://github.com/stuartshay/homelab-database-migrations)

This application now relies on a centralized migration system for database schema management. See the homelab-database-migrations repository for:
- Database setup instructions (dev/prod)
- Schema migration files
- Garmin Connect integration schema
- Data seeding scripts

#### Database Environment Updates

- [ ] **Update development environment**
  - Change `.env` to use `owntracks_dev` database
  - Test connectivity to 192.168.1.175:5432/owntracks_dev
  - Verify application works with dev schema

- [ ] **Document database dependencies**
  - Link to homelab-database-migrations README
  - Add database setup section to README.md
  - Document required schema version

#### Garmin Connect Integration (Future)

The database now supports dual-source GPS tracking:

1. **OwnTracks** (existing) - Real-time location tracking
   - `public.locations` table
   - Device-based tracking
   - Battery, accuracy, velocity metrics

2. **Garmin Connect** (new) - Activity-based GPS tracks
   - `public.garmin_activities` - Workout metadata
   - `public.garmin_track_points` - GPS coordinates from activities
   - Rich metrics: heart rate, cadence, speed, elevation

**Application Updates Required** (future work):

- [ ] Add gRPC methods for Garmin activity queries
  - `GetActivitiesByDateRange`
  - `GetActivityDetails`
  - `GetActivitiesWithinRadius`

- [ ] Update distance calculations to handle Garmin data
  - Query `garmin_track_points` table
  - Calculate distance metrics per activity
  - Generate CSV with activity metadata

- [ ] Add unified GPS endpoints
  - Query `unified_gps_points` view
  - Combine OwnTracks + Garmin data
  - Filter by date, device, sport type

#### Migration Workflow

1. **Developers**: Apply migrations locally to `owntracks_dev`
   ```bash
   git clone https://github.com/stuartshay/homelab-database-migrations.git
   cd homelab-database-migrations
   ./scripts/migrate.sh dev up
   ```

2. **CI/CD**: Production migrations applied automatically
   ```bash
   # Triggered by git tag in homelab-database-migrations
   ./scripts/migrate.sh prod up
   ```

3. **otel-worker deployment**: Always follows successful migration

#### Data Seeding for Development

The migrations repo includes rich test data:
- **Garmin activity**: 50.6km cycling with 10,707 GPS points
- **OwnTracks data**: Synthetic tracks for realistic testing
- **Seed script**: `./scripts/seed_dev_data.sh`

This enables local testing with realistic GPS movement patterns instead of static data.

#### Related Documentation

- **[homelab-database-migrations README](https://github.com/stuartshay/homelab-database-migrations/blob/main/README.md)**
- **[Migration Guide](https://github.com/stuartshay/homelab-database-migrations/blob/main/docs/MIGRATION_GUIDE.md)**
- **[Schema Reference](https://github.com/stuartshay/homelab-database-migrations/blob/main/docs/schema_reference.sql)**

**Estimated Time**: 2-3 hours for dev environment updates, 6-8 hours for Garmin integration