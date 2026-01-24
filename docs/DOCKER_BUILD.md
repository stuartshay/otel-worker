# Docker Build Documentation

## Overview

The otel-worker service is containerized using a multi-stage Docker build that produces a minimal, secure image.

## Build Details

### Base Images

- **Builder**: `golang:1.24-alpine` - Minimal Alpine Linux with Go 1.24
- **Runtime**: `gcr.io/distroless/static-debian12:nonroot` - Distroless image with no shell, minimal attack surface

### Image Size

- **Final Image**: 14.9 MB
- **Target**: < 20 MB ✅

### Build Features

- Static binary compilation (`CGO_ENABLED=0`)
- Stripped binary (`-ldflags="-w -s"`)
- Linux/amd64 architecture
- Non-root user execution
- Timezone data included
- CA certificates included

## Build Commands

### Local Build

```bash
# Standard build
docker build -t stuartshay/otel-worker:latest .

# Build without cache
docker build --no-cache -t stuartshay/otel-worker:latest .

# Multi-arch build (when pushing to registry)
docker buildx build --platform linux/amd64,linux/arm64 -t stuartshay/otel-worker:latest --push .
```

### Run Locally

```bash
# Run with environment variables
docker run -p 50051:50051 -p 8080:8080 \
  --env-file .env \
  stuartshay/otel-worker:latest

# Run in background
docker run -d -p 50051:50051 -p 8080:8080 \
  --name otel-worker \
  --env-file .env \
  stuartshay/otel-worker:latest

# View logs
docker logs -f otel-worker
```

## Exposed Ports

| Port  | Protocol | Purpose        |
|-------|----------|----------------|
| 50051 | gRPC     | Service API    |
| 8080  | HTTP     | Health checks  |

## Volume Mounts

- `/data/csv` - CSV output directory (owned by nonroot:nonroot)

Example with volume:

```bash
docker run -p 50051:50051 -p 8080:8080 \
  -v $(pwd)/csv-output:/data/csv \
  --env-file .env \
  stuartshay/otel-worker:latest
```

## Environment Variables

Required:

- `POSTGRES_HOST` - Database host (default: localhost)
- `POSTGRES_PORT` - Database port (default: 6432 for PgBouncer)
- `POSTGRES_DB` - Database name (default: owntracks)
- `POSTGRES_USER` - Database user
- `POSTGRES_PASSWORD` - Database password

Optional:

- `GRPC_PORT` - gRPC server port (default: 50051)
- `HTTP_PORT` - HTTP health check port (default: 8080)
- `CSV_OUTPUT_PATH` - CSV output directory (default: /data/csv)
- `LOG_LEVEL` - Logging level (default: info)
- `HOME_LATITUDE` - Home latitude (default: 40.736097)
- `HOME_LONGITUDE` - Home longitude (default: -74.039373)
- `AWAY_THRESHOLD_KM` - Away threshold in km (default: 0.5)

## Security

- Runs as non-root user (uid: 65532, gid: 65532)
- No shell in runtime image
- Minimal attack surface (distroless)
- Static binary with no external dependencies
- CA certificates included for HTTPS connections

## Dockerfile Stages

### Stage 1: Builder

1. Install build dependencies (git, ca-certificates, tzdata)
2. Download Go modules
3. Copy source code
4. Build static binary with optimizations
5. Create CSV output directory structure

### Stage 2: Runtime

1. Copy timezone data from builder
2. Copy CA certificates from builder
3. Copy compiled binary from builder
4. Copy CSV directory with nonroot ownership
5. Set working directory to /app
6. Expose ports 50051 and 8080
7. Set environment variable defaults
8. Run as nonroot user

## CI/CD Integration

The Dockerfile is used by GitHub Actions workflows:

- **docker.yml** - Builds and pushes multi-arch images on merge to main
- **test.yml** - Builds image for integration testing

## Troubleshooting

### Build Failures

**Issue**: `CGO_ENABLED=0` build fails

- **Solution**: Ensure no C dependencies in Go code

**Issue**: Permission denied on /data/csv

- **Solution**: Directory owned by nonroot:nonroot, check volume permissions

### Runtime Issues

**Issue**: Cannot connect to database

- **Solution**: Verify POSTGRES_* environment variables, check network connectivity

**Issue**: CSV files not persisted

- **Solution**: Mount volume at /data/csv, verify directory permissions

## Next Steps

- [ ] Push image to Docker Hub
- [ ] Set up automated multi-arch builds
- [ ] Add health check probe binary
- [ ] Configure vulnerability scanning
- [ ] Set up image signing

## References

- [Dockerfile Best Practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- [Distroless Images](https://github.com/GoogleContainerTools/distroless)
- [Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/)
