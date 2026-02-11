# Copilot Rules for otel-worker Repo

These rules ensure Copilot/assistants follow best practices for Go gRPC
microservice development.

## Always Read First

- **README**: Read `README.md` for project overview and quick start
- **docs**: Check `docs/DATABASE.md` for schema and query patterns
- **env**: Load `.env` for database credentials (PgBouncer on port 6432)
- **pre-commit**: ALWAYS run `pre-commit run -a` before commit/PR

## Target Infrastructure

| Property            | Value                          |
|---------------------|--------------------------------|
| Language            | Go 1.23+                       |
| Protocol            | gRPC + Protobuf                |
| Database            | PostgreSQL (PgBouncer pooling) |
| K8s Cluster         | k8s-pi5-cluster                |
| Namespace           | otel-worker                    |
| Home Coordinates    | 40.736097°N, 74.039373°W       |

## Development Workflow

### Branch Strategy

⚠️ **CRITICAL RULE**: NEVER commit directly to `main` branch. All changes
MUST go through `develop` or `feature/*` branches.

- **main**: Protected branch, production-only (PR required, direct commits
  FORBIDDEN)
- **develop**: Primary development branch (work here by default)
- **feature/\***: Feature branches (use for isolated features, PR to `main`)

### Before Starting Any Work

**ALWAYS sync your working branch with the remote before making changes:**

```bash
# If working on develop:
git checkout develop && git fetch origin && git pull origin develop

# If creating a new feature branch:
git checkout main && git fetch origin && git pull origin main
git checkout -b feature/my-feature
```

### Daily Workflow

1. **ALWAYS** start from `develop` or create a feature branch
2. **Sync with remote** before making any changes (see above)
3. Make changes to Go source files
4. Run `go test ./...` to validate
5. Run `pre-commit run -a` before commit
6. Commit and push to `develop` or `feature/*` branch
7. Create PR to `main` when ready for deployment
8. **NEVER**: `git push origin main` or commit directly to main

## Writing Go Code

### Project Structure

- `cmd/server/` - Application entrypoint
- `internal/config/` - Configuration management
- `internal/database/` - PostgreSQL client
- `internal/calculator/` - Distance calculations
- `internal/queue/` - Job queue system
- `internal/grpc/` - gRPC service handlers
- `proto/distance/v1/` - Protobuf definitions

### Best Practices

- Use standard Go project layout
- Write unit tests alongside implementation (target 80%+ coverage)
- Use context.Context for cancellation and timeouts
- Handle errors explicitly, never ignore them
- Use structured logging with zerolog
- Follow effective Go guidelines
- Use gofmt, goimports for formatting

### Database

- Always use PgBouncer (port 6432) for connections
- Use prepared statements for repeated queries
- Handle connection pooling with max connections (25) and idle (5)
- Filter by `created_at` (indexed) instead of `timestamp`
- Close database connections in defer statements

### gRPC

- Define services in `proto/distance/v1/distance.proto`
- Regenerate code with `make proto` after proto changes
- Use `timestamppb` for timestamp fields
- Implement health checks with `grpc_health_v1`
- Enable server reflection for debugging

## Safety Rules (Do Not)

- ⛔ **NEVER commit directly to main branch** - ALWAYS use develop or feature
  branches
- Do not commit secrets or database credentials
- Do not use `latest` tag in Docker images
- Do not skip `pre-commit run -a` before commits
- Do not hardcode database connection strings
- Do not ignore context cancellation
- Do not use blocking operations without timeouts

## Quick Commands

```bash
# Generate protobuf code
make proto

# Run tests
make test
go test -v -race -coverprofile=coverage.out ./...

# Build binary
make build
go build -o bin/otel-worker cmd/server/main.go

# Run locally (requires .env)
./bin/otel-worker

# Docker
docker build -t stuartshay/otel-worker .
docker run -p 50051:50051 --env-file .env stuartshay/otel-worker

# Pre-commit checks
pre-commit run --all-files
```

## Long-Running Commands

| Command Type     | Estimated Duration |
|------------------|--------------------|
| `go test ./...`  | 1-5 seconds        |
| `go build`       | 3-10 seconds       |
| `make proto`     | 1-2 seconds        |
| `docker build`   | 30-90 seconds      |
| Pre-commit hooks | 5-15 seconds       |

## CI/CD Pipeline

- **lint.yml**: Runs golangci-lint, markdownlint, yamllint, hadolint
- **test.yml**: Unit tests + integration tests with PostgreSQL service
- **docker.yml**: Multi-arch builds (amd64/arm64), push to Docker Hub

## Related Repositories

- [k8s-gitops](https://github.com/stuartshay/k8s-gitops) - Kubernetes manifests
- [otel-demo](https://github.com/stuartshay/otel-demo) - Flask API (may integrate)
- [AnsiblePlaybooks](https://github.com/stuartshay/AnsiblePlaybooks) - Infrastructure

## When Unsure

- Check existing code patterns in `internal/`
- Reference Go standard library documentation
- Validate with `go test ./...` before committing
- Test with real database before pushing

## Database Credentials

```bash
# PgBouncer (recommended for application connections)
POSTGRES_HOST=192.168.1.175
POSTGRES_PORT=6432
POSTGRES_DB=owntracks
POSTGRES_USER=development
POSTGRES_PASSWORD=development

# Direct PostgreSQL (only for migrations/admin)
POSTGRES_PORT=5432
```

## Home Location

All distance calculations use a fixed home location:

- **Latitude**: 40.736097°N
- **Longitude**: 74.039373°W
- **Away Threshold**: 0.5 km (configurable via `AWAY_THRESHOLD_KM`)
