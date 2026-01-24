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
	@pip install pre-commit || pip3 install pre-commit
	@pre-commit install
	@echo "✓ Pre-commit hooks installed"

pre-commit-run:
	@echo "Running pre-commit checks..."
	@pre-commit run --all-files
	@echo "✓ Pre-commit checks complete"

pre-commit-update:
	@pre-commit autoupdate
