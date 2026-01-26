#!/bin/bash
# Post-create command for otel-worker DevContainer
# This script runs after the container is created
#
# Tools installed via DevContainer Features (not installed here):
# - Go SDK (base image)
# - docker-in-docker (ghcr.io/devcontainers/features/docker-in-docker)
# - common-utils/zsh (ghcr.io/devcontainers/features/common-utils)
# - github-cli (ghcr.io/devcontainers/features/github-cli)
# - kubectl & helm (ghcr.io/devcontainers/features/kubectl-helm-minikube)
# - golangci-lint (ghcr.io/guiyomh/features/golangci-lint)
# - pre-commit (ghcr.io/prulloac/devcontainer-features/pre-commit)
# - protoc (ghcr.io/devcontainers-extra/features/protoc)
# - grpcurl (ghcr.io/devcontainers-extra/features/grpcurl-asdf) - gRPC CLI
# - buf CLI (ghcr.io/devcontainers-extra/features/buf)
# - hadolint (ghcr.io/dhoeric/features/hadolint) - Dockerfile linting
# - shellcheck (ghcr.io/lukewiwa/features/shellcheck) - Shell script linting
# - actionlint (ghcr.io/jsburckhardt/devcontainer-features/actionlint) - GitHub Actions linting
#
# Tools installed via Go in this script:
# - grpcui (installed via go install)

set -e

echo "🔧 Setting up otel-worker development environment..."

# Install PostgreSQL client for remote database connectivity
echo "📦 Installing PostgreSQL client..."
sudo apt-get update
sudo apt-get install -y --no-install-recommends postgresql-client

# Install grpcui (grpcurl installed via devcontainer feature)
echo "📦 Installing grpcui..."
go install github.com/fullstorydev/grpcui/cmd/grpcui@latest

# Install Go protobuf plugins
echo "📦 Installing Go protobuf plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Install goimports for code formatting
echo "📦 Installing goimports..."
go install golang.org/x/tools/cmd/goimports@latest

# Initialize pre-commit hooks if .pre-commit-config.yaml exists
# (pre-commit is installed via DevContainer Feature)
if [ -f ".pre-commit-config.yaml" ]; then
    echo "🔗 Installing pre-commit hooks..."
    pre-commit install
fi

# Download Go module dependencies
echo "📦 Downloading Go dependencies..."
go mod download

# Create .env from .env.example if not exists
if [ ! -f ".env" ] && [ -f ".env.example" ]; then
    echo "📄 Creating .env from .env.example..."
    cp .env.example .env
    echo "   Please update .env with your configuration"
fi

# Clean up apt cache
sudo apt-get clean
sudo rm -rf /var/lib/apt/lists/*

echo ""
echo "✅ Development environment setup complete!"
echo ""
echo "Available commands:"
echo "  make build     - Build the otel-worker binary"
echo "  make run       - Build and run locally"
echo "  make test      - Run tests with coverage"
echo "  make proto     - Generate protobuf code"
echo "  make grpcui    - Launch gRPC UI"
echo "  make grpcui-k8s - Port-forward k8s and launch gRPC UI"
echo "  make help      - Show all available targets"
echo ""
echo "Tools available:"
echo "  protoc, buf, grpcurl, grpcui, golangci-lint, kubectl, helm, gh"
echo "  hadolint, shellcheck, actionlint, pre-commit"
