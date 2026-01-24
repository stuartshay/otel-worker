#!/bin/bash
set -e
echo "=== otel-worker Setup Script ==="
OS=$(uname -s)

if command -v go &> /dev/null; then
    INSTALLED_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo "✓ Go installed: $INSTALLED_VERSION"
    GO_INSTALLED=true
else
    echo "✗ Go not installed"
    GO_INSTALLED=false
fi

if [ "$GO_INSTALLED" = false ]; then
    echo "Installing Go 1.23.5..."
    GO_VERSION="1.23.5"
    GO_TARBALL="go${GO_VERSION}.linux-amd64.tar.gz"
    wget -q "https://go.dev/dl/$GO_TARBALL" -O "/tmp/$GO_TARBALL"
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "/tmp/$GO_TARBALL"
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin:$(go env GOPATH)/bin' >> ~/.bashrc
    fi
    export PATH=$PATH:/usr/local/go/bin:$(go env GOPATH)/bin
    rm "/tmp/$GO_TARBALL"
    echo "✓ Go installed"
fi

go version

if ! command -v protoc &> /dev/null; then
    echo "Installing protoc..."
    sudo apt-get update -qq && sudo apt-get install -y protobuf-compiler
fi

echo "Installing Go plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

if [ ! -f "go.mod" ]; then
    go mod init github.com/stuartshay/otel-worker
fi

echo "Installing dependencies..."
go get google.golang.org/grpc@latest google.golang.org/protobuf@latest github.com/lib/pq@latest github.com/umahmood/haversine@latest github.com/joho/godotenv@latest github.com/rs/zerolog@latest go.opentelemetry.io/otel@latest go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@latest go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc@latest

go mod tidy

[ -f ".env.example" ] && [ ! -f ".env" ] && cp .env.example .env

echo "✓ Setup complete"
