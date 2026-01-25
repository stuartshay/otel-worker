# Quick Wins - otel-worker

**Last Updated**: 2026-01-25
**Status**: Phase 3 (Testing) & Phase 4 (Deployment) in progress

## Overview

This document tracks immediate, high-value tasks that can be completed quickly to
move the project forward. Focus on these before larger integration work.

## Current Status Summary

### ✅ Completed (Phase 1-2)

- Core gRPC service implementation
- Database layer with PgBouncer integration
- Distance calculation logic (Haversine)
- Job queue system with worker pool
- CI/CD pipelines (Lint, Test, Docker, Proto)
- Docker multi-arch builds (14.9MB image)
- Semantic versioning (1.0.{BUILD_NUMBER})
- Pre-commit hooks (local, reliable)
- Buf Schema Registry integration

### 🟡 In Progress (Phase 3-4)

- Test coverage: 50% overall (92-100% in core logic, 0-13% in handlers)
- Docker images built but not pushed to Docker Hub
- No Kubernetes manifests created yet
- No health check endpoints

### ❌ Not Started (Phase 5-6)

- OpenTelemetry instrumentation
- Integration with otel-demo REST API
- otel-ui frontend components

## Quick Wins (Prioritized)

### 🎯 Week 1 Focus

#### 1. Health Check Endpoints ⚡ (30 minutes) - ✅ COMPLETE

**Why**: Required for Kubernetes liveness/readiness probes. Blocks deployment.

**Tasks**:

- [x] Add HTTP server on port 8080 in `cmd/server/main.go`
- [x] Implement `GET /healthz` (liveness) - always returns 200
- [x] Implement `GET /readyz` (readiness) - checks database connectivity
- [x] Add graceful HTTP server shutdown
- [x] Test locally: `curl http://localhost:8080/healthz`
- [x] Update README.md with health check documentation

**Completed**: 2026-01-25
**Estimated Time**: 30 minutes
**Priority**: P0 - Blocks deployment
**Dependencies**: None

**Implementation Details**:

- HTTP server runs on port 8080 with proper timeouts (10s read header, 30s read/write, 60s idle)
- `/healthz` returns `{"status":"healthy","service":"otel-worker"}` (200 OK)
- `/readyz` returns `{"status":"ready","database":"connected"}` (200 OK) or 503 if DB down
- Graceful shutdown with 30-second timeout integrates with existing shutdown logic
- Unit test added in `cmd/server/main_health_test.go`

---

#### 2. gRPC Handler Integration Tests ⚡ (2-3 hours)

**Why**: Currently 0% handler coverage. Need confidence before production deployment.

**Tasks**:

- [ ] Create `internal/grpc/handler_test.go`
- [ ] Add test database setup (use PostgreSQL service from test workflow)
- [ ] Test `CalculateDistanceFromHome` with real data
- [ ] Test `GetJobStatus` polling behavior
- [ ] Test `ListJobs` pagination
- [ ] Test CSV file generation and content
- [ ] Test error scenarios (invalid date, missing device)
- [ ] Verify job queue integration

**Estimated Time**: 2-3 hours
**Priority**: P0 - Critical for production confidence
**Dependencies**: Test database available in CI

---

#### 3. Kubernetes Manifests ⚡ (2 hours) - ✅ COMPLETE

**Why**: Required to deploy to k8s-pi5-cluster. Sister projects (otel-demo, otel-ui) already deployed.

**Tasks**:

- [x] Create `k8s-gitops/apps/base/otel-worker/` directory
- [x] Create `deployment.yaml` (2 replicas, gRPC port 50051, HTTP port 8080)
- [x] Create `service.yaml` (ClusterIP, gRPC endpoint)
- [x] Create `configmap.yaml` (home coordinates, threshold)
- [x] Create `secret.yaml` (SOPS-encrypted database credentials)
- [x] Create `pvc.yaml` (10Gi for CSV files)
- [x] Create `kustomization.yaml`
- [x] Add health/readiness probes using new endpoints
- [x] Configure resources (requests: 100m CPU, 256Mi; limits: 500m CPU, 512Mi)
- [x] Create cluster overlay in `clusters/k8s-pi5-cluster/apps/`

**Completed**: 2026-01-25
**Estimated Time**: 2 hours
**Priority**: P0 - Blocks deployment
**Dependencies**: Health check endpoints (task #1) ✅
**Reference**: Follow patterns from otel-demo and otel-ui

**Implementation Details**:

- Created 11 files in k8s-gitops repository
- Validated with kubeconform (6/7 resources valid, SealedSecret is CRD)
- Resources: 100m/256Mi requests, 500m/512Mi limits
- 2 replicas for high availability
- Health probes configured (10s/10s liveness, 5s/5s readiness)
- 10Gi PVC on nfs-general storage class
- Service exposes gRPC (50051) and HTTP (8080) ports
- Committed to k8s-gitops main branch

---

#### 4. Push Docker Images to Docker Hub (15 minutes)

**Why**: Required for Kubernetes deployment. Images exist but not pushed.

**Tasks**:

- [ ] Configure GitHub secrets: `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`
- [ ] Verify docker.yml workflow pushes on merge to main
- [ ] Trigger workflow by merging a branch
- [ ] Verify image at <https://hub.docker.com/r/stuartshay/otel-worker>
- [ ] Test pull: `docker pull stuartshay/otel-worker:1.0.{BUILD_NUMBER}`

**Estimated Time**: 15 minutes
**Priority**: P0 - Blocks deployment
**Dependencies**: None (Docker Hub repo already exists)

---

### 🚀 Week 2 Focus

#### 5. Deploy to k8s-pi5-cluster (1 hour)

**Why**: Get service running in production to validate end-to-end functionality.

**Tasks**:

- [ ] Commit K8s manifests to k8s-gitops repo
- [ ] Create Argo CD application
- [ ] Set sync policy (manual first, then auto)
- [ ] Apply changes: `kubectl apply -k clusters/k8s-pi5-cluster/apps/otel-worker`
- [ ] Verify pods running: `kubectl get pods -n otel-worker`
- [ ] Check logs: `kubectl logs -n otel-worker -l app=otel-worker`
- [ ] Test gRPC endpoint from within cluster
- [ ] Verify CSV files written to PVC
- [ ] Monitor for errors in first 24 hours

**Estimated Time**: 1 hour
**Priority**: P1 - Production deployment
**Dependencies**: Tasks #1-4 complete

---

#### 6. OpenTelemetry Instrumentation (3-4 hours)

**Why**: Observability for production debugging and performance monitoring.

**Tasks**:

- [ ] Add OpenTelemetry SDK dependencies
- [ ] Configure OTel exporter (otel-collector endpoint)
- [ ] Instrument gRPC server with `otelgrpc` interceptors
- [ ] Trace database queries (wrap with spans)
- [ ] Trace distance calculations
- [ ] Add custom span attributes (device_id, date, job_id)
- [ ] Configure trace sampling (1.0 for dev)
- [ ] Verify traces in New Relic or Jaeger
- [ ] Add trace_id to log output

**Estimated Time**: 3-4 hours
**Priority**: P1 - Production observability
**Dependencies**: Task #5 (deployed to K8s)

---

#### 7. Database Test Coverage (2 hours)

**Why**: Currently 13.3% coverage. Need confidence in query logic.

**Tasks**:

- [ ] Add more tests in `internal/database/client_test.go`
- [ ] Test `GetLocationsByDate` with various scenarios
- [ ] Test `GetLocationsByDateAndDevice` filtering
- [ ] Test connection retry logic
- [ ] Test connection pool behavior
- [ ] Test query timeout handling
- [ ] Achieve 80%+ coverage in database package

**Estimated Time**: 2 hours
**Priority**: P2 - Quality improvement
**Dependencies**: None

---

### 🌟 Week 3+ Focus

#### 8. otel-demo REST API Integration (4-5 hours)

**Why**: Provide REST interface for frontend. Complete the service chain.

**Tasks**:

- [ ] Add gRPC client to otel-demo Flask app
- [ ] Create REST endpoint: `POST /api/distance/calculate`
- [ ] Implement async job polling endpoint: `GET /api/distance/jobs/{job_id}`
- [ ] Return CSV download link when complete
- [ ] Add Cognito JWT validation
- [ ] Add rate limiting (10 requests/min)
- [ ] Add OpenTelemetry context propagation
- [ ] Test end-to-end flow

**Estimated Time**: 4-5 hours
**Priority**: P2 - User-facing functionality
**Dependencies**: Task #5 (otel-worker deployed)

---

#### 9. otel-ui Frontend Components (6-8 hours)

**Why**: Provide user interface for distance calculations.

**Tasks**:

- [ ] Create `DistanceCalculator.tsx` component
- [ ] Add date picker for calculation date
- [ ] Add device selector (auto-populate from API)
- [ ] Display job status (pending/processing/completed/failed)
- [ ] Show summary metrics (max distance, total locations)
- [ ] Add CSV download button
- [ ] Add error handling and loading states
- [ ] Use existing Cognito auth patterns
- [ ] Follow otel-ui styling/patterns

**Estimated Time**: 6-8 hours
**Priority**: P3 - UI enhancement
**Dependencies**: Task #8 (otel-demo API ready)

---

## Success Metrics

### Week 1 Goals

- ✅ Health checks implemented and tested
- ✅ Handler test coverage > 50%
- ✅ K8s manifests created
- ✅ Docker images pushed to Docker Hub
- ✅ All GitHub Actions passing
- ✅ Pre-commit hooks passing

### Week 2 Goals

- ✅ Deployed to k8s-pi5-cluster
- ✅ Service responds to gRPC requests
- ✅ CSV files generated successfully
- ✅ OpenTelemetry traces visible
- ✅ Zero critical errors in logs
- ✅ Database coverage > 80%

### Week 3+ Goals

- ✅ REST API integration complete
- ✅ Frontend UI working
- ✅ End-to-end user workflow functional
- ✅ Documentation complete

## Notes

- **Focus on P0 tasks first** - These block further progress
- **Test locally before pushing** - Saves CI/CD time
- **Follow sister project patterns** - otel-demo and otel-ui have proven patterns
- **Document as you go** - Update README.md and docs/ files
- **Run pre-commit hooks** - Ensure code quality: `make pre-commit-run`

## Next Action

**Start with Task #1**: Health Check Endpoints (30 minutes)

This is the foundation for Kubernetes deployment and blocks all deployment tasks.
