.PHONY: proto build run test clean docker-build help

BINARY_NAME=otel-worker
DOCKER_IMAGE=stuartshay/otel-worker
VERSION?=latest

proto:
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/distance/v1/distance.proto

build:
	@go build -o bin/$(BINARY_NAME) cmd/server/main.go

run: build
	@./bin/$(BINARY_NAME)

test:
	@go test -v -race -coverprofile=coverage.out ./...

clean:
	@rm -rf bin/ coverage.out

docker-build:
	@docker build -t $(DOCKER_IMAGE):$(VERSION) .

help:
	@echo "Targets: proto build run test clean docker-build"

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
