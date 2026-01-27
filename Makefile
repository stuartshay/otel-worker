.PHONY: proto proto-buf buf-lint buf-format build run grpcui grpcui-k8s test coverage lint clean docker-build help pre-commit-install pre-commit-run pre-commit-update install

BINARY_NAME=otel-worker
DOCKER_IMAGE=stuartshay/otel-worker
VERSION?=latest

# Proto generation using protoc (default)
proto:
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/distance/v1/distance.proto

# Proto generation using buf (alternative, more modern)
proto-buf:
	@buf generate

build:
	@go build -o bin/$(BINARY_NAME) cmd/server/main.go

run: build
	@./bin/$(BINARY_NAME)

grpcui:
	@echo "Starting grpcui on http://localhost:8080..."
	@grpcui -plaintext localhost:50051

grpcui-k8s:
	@echo "Port-forwarding otel-worker service and starting grpcui..."
	@echo "Press Ctrl+C to stop"
	@kubectl port-forward -n otel-worker svc/otel-worker 50051:50051 & \
	PF_PID=$$!; \
	sleep 2; \
	grpcui -plaintext localhost:50051; \
	kill $$PF_PID 2>/dev/null || true

# Buf linting
buf-lint:
	@buf lint

# Buf formatting
buf-format:
	@buf format -w

test:
	@go test -v -race -coverprofile=coverage.out ./...

# Show coverage report
coverage: test
	@go tool cover -func=coverage.out
	@echo ""
	@echo "For HTML coverage report, run: go tool cover -html=coverage.out"

# Run linting
lint:
	@golangci-lint run --timeout=5m

clean:
	@rm -rf bin/ coverage.out

docker-build:
	@docker build -t $(DOCKER_IMAGE):$(VERSION) .

help:
	@echo "Available targets:"
	@echo "  proto           - Generate protobuf code using protoc"
	@echo "  proto-buf       - Generate protobuf code using buf (alternative)"
	@echo "  buf-lint        - Lint protobuf files with buf"
	@echo "  buf-format      - Format protobuf files with buf"
	@echo "  build           - Build the otel-worker binary"
	@echo "  run             - Build and run the application locally"
	@echo "  grpcui          - Launch grpcui web interface (requires local server on :50051)"
	@echo "  grpcui-k8s      - Port-forward k8s service and launch grpcui"
	@echo "  test            - Run tests with race detection and coverage"
	@echo "  coverage        - Run tests and show coverage report"
	@echo "  lint            - Run golangci-lint"
	@echo "  clean           - Remove build artifacts"
	@echo "  docker-build    - Build Docker image"
	@echo "  install         - Install development dependencies"
	@echo "  pre-commit-*    - Pre-commit hook management"

# Pre-commit
pre-commit-install:
	@echo "Installing pre-commit hooks..."
	@command -v pre-commit >/dev/null 2>&1 || { echo "Installing pre-commit..."; pip3 install --user pre-commit; }
	@if [ -f ~/.local/bin/pre-commit ]; then ~/.local/bin/pre-commit install; else pre-commit install; fi
	@echo "✓ Pre-commit hooks installed"

pre-commit-run:
	@echo "Running pre-commit checks..."
	@if [ -f ~/.local/bin/pre-commit ]; then ~/.local/bin/pre-commit run --all-files; else pre-commit run --all-files; fi
	@echo "✓ Pre-commit checks complete"

pre-commit-update:
	@if [ -f ~/.local/bin/pre-commit ]; then ~/.local/bin/pre-commit autoupdate; else pre-commit autoupdate; fi

# Install development dependencies
install:
	@echo "Installing Go dependencies..."
	@go mod download
	@echo "Installing Go tools..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/fullstorydev/grpcui/cmd/grpcui@latest
	@echo "✓ All dependencies installed"
