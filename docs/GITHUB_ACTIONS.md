# GitHub Actions CI/CD Workflows

## Overview

The otel-worker project uses GitHub Actions for continuous integration and continuous deployment.
Three workflows automate testing, linting, and Docker image builds.

## Workflows

### 1. Lint (`lint.yml`)

**Trigger**: Push and pull requests to `main` and `develop` branches

**Jobs**:

- **Go Lint** (`golangci`) - Runs golangci-lint with 5-minute timeout
- **Markdown Lint** (`markdown`) - Validates all Markdown files
- **YAML Lint** (`yaml`) - Validates YAML files with 120 char line length
- **Dockerfile Lint** (`dockerfile`) - Runs hadolint on Dockerfile
- **Protobuf Lint** (`protobuf`) - Runs buf format check and buf lint

**Requirements**:

- Go 1.24
- golangci-lint action
- markdownlint-cli2
- yamllint
- hadolint
- buf CLI

### 2. Test (`test.yml`)

**Trigger**: Push and pull requests to `main` and `develop` branches

**Jobs**:

- **Unit Tests** (`test`)
  - Runs all unit tests with race detector
  - Generates coverage report
  - Uploads coverage to Codecov
  - Outputs: `coverage.out`, `coverage.html`

- **Integration Tests** (`integration`)
  - Spins up PostgreSQL 16 service container
  - Initializes database schema (`public.locations` table)
  - Runs integration tests with `-tags=integration`
  - Environment: localhost:5432 PostgreSQL

**Database Schema**:

```sql
CREATE TABLE public.locations (
  id SERIAL PRIMARY KEY,
  device_id VARCHAR(255) NOT NULL,
  latitude DOUBLE PRECISION NOT NULL,
  longitude DOUBLE PRECISION NOT NULL,
  accuracy DOUBLE PRECISION,
  battery INTEGER,
  velocity DOUBLE PRECISION,
  timestamp BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  altitude DOUBLE PRECISION,
  vertical_accuracy DOUBLE PRECISION,
  course DOUBLE PRECISION,
  topic VARCHAR(255),
  tracker_id VARCHAR(255),
  description TEXT
);
CREATE INDEX idx_locations_created_at ON public.locations(created_at);
CREATE INDEX idx_locations_device_id ON public.locations(device_id);
```

**Requirements**:

- Go 1.24
- PostgreSQL 16 (alpine) service container
- psql command-line tool

### 3. Docker (`docker.yml`)

**Trigger**:

- Push to `main` branch
- Push tags matching `v*` (e.g., v1.0.0)
- Pull requests to `main` (build only, no push)

**Jobs**:

- **Build and Push** (`build`)
  - Sets up Docker Buildx for multi-platform builds
  - Extracts metadata for tags and labels
  - Builds for linux/amd64 and linux/arm64
  - Pushes to Docker Hub (on non-PR events)
  - Runs Trivy vulnerability scanner
  - Uploads security scan results to GitHub Security

**Image Tags**:

| Event | Tag Example |
|-------|-------------|
| Branch push | `main`, `develop` |
| Pull request | `pr-123` |
| Semver tag | `v1.2.3`, `1.2`, `1` |
| Commit SHA | `main-abc1234` |
| Latest (main) | `latest` |

**Multi-Platform**:

- linux/amd64 (x86_64)
- linux/arm64 (ARM 64-bit, Raspberry Pi)

**Security**:

- Trivy scans for vulnerabilities in final image
- Results uploaded to GitHub Security tab (SARIF format)
- Scans run only on pushes (not PRs)

**Requirements**:

- Docker Hub credentials:
  - `DOCKERHUB_USERNAME`
  - `DOCKERHUB_TOKEN`

## Secrets Configuration

Configure these secrets in GitHub repository settings (Settings → Secrets and variables → Actions):

| Secret | Description | Required By |
|--------|-------------|-------------|
| `DOCKERHUB_USERNAME` | Docker Hub username | docker.yml |
| `DOCKERHUB_TOKEN` | Docker Hub access token | docker.yml |

**Note**: Codecov token is optional; the action will work without it but may hit rate limits.

## Status Badges

Add to README.md:

```markdown
[![Lint](https://github.com/stuartshay/otel-worker/actions/workflows/lint.yml/badge.svg)](https://github.com/stuartshay/otel-worker/actions/workflows/lint.yml)
[![Test](https://github.com/stuartshay/otel-worker/actions/workflows/test.yml/badge.svg)](https://github.com/stuartshay/otel-worker/actions/workflows/test.yml)
[![Docker](https://github.com/stuartshay/otel-worker/actions/workflows/docker.yml/badge.svg)](https://github.com/stuartshay/otel-worker/actions/workflows/docker.yml)
```

## Workflow Outputs

### Artifacts

- **Unit Tests**: `coverage.html` (HTML coverage report)
- **Security Scan**: `trivy-results.sarif` (uploaded to GitHub Security)

### Codecov Integration

Coverage reports are uploaded to Codecov for trend analysis:

- **URL**: `https://codecov.io/gh/stuartshay/otel-worker`
- **Flag**: `unittests`
- **File**: `coverage.out`

## Cache Strategy

All workflows use GitHub Actions cache to speed up builds:

- **Go modules**: `~/go/pkg/mod`
- **Go build cache**: `~/.cache/go-build`
- **Docker layers**: `type=gha` (GitHub Actions cache)

## Manual Triggers

All workflows can be manually triggered via GitHub Actions UI:

1. Go to Actions tab
2. Select workflow
3. Click "Run workflow"
4. Choose branch
5. Click "Run workflow"

## Troubleshooting

### Lint Failures

**Issue**: golangci-lint times out

- **Solution**: Increase timeout in workflow (currently 5m)

**Issue**: Markdown lint fails on line length

- **Solution**: YAML lint allows 120 chars; check markdownlint config

### Test Failures

**Issue**: Integration tests fail to connect to PostgreSQL

- **Solution**: Wait for health check to pass (10s interval, 5 retries)

**Issue**: Database schema not initialized

- **Solution**: Check psql command in "Initialize database schema" step

### Docker Build Failures

**Issue**: Multi-platform build fails

- **Solution**: Ensure Buildx is set up correctly

**Issue**: Push fails with authentication error

- **Solution**: Verify DOCKERHUB_USERNAME and DOCKERHUB_TOKEN secrets

**Issue**: Trivy scan fails

- **Solution**: Check image was built successfully; verify image-ref format

## Next Steps

- [ ] Add build status badges to README
- [ ] Configure Docker Hub credentials in repository secrets
- [ ] Set up Codecov project (optional)
- [ ] Add integration test data fixtures
- [ ] Configure branch protection rules (require passing CI)

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [golangci-lint](https://golangci-lint.run/)
- [Trivy](https://aquasecurity.github.io/trivy/)
- [Codecov](https://about.codecov.io/)
