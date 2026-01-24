#!/bin/bash
set -e

echo "=== otel-worker Project Bootstrap ==="
cd /home/ubuntu/git/otel-worker

# Create directories
mkdir -p cmd/server internal/{grpc,database,calculator,config,queue} proto/distance/v1 scripts docs .github/workflows bin data/csv

# Create setup.sh (with proper escaping)
cat > scripts/setup.sh << 'SETUP_EOF'
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
SETUP_EOF

chmod +x scripts/setup.sh

# Create other files...
cat > Makefile << 'MAKEFILE_EOF'
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
MAKEFILE_EOF

cat > .env.example << 'ENV_EOF'
SERVICE_NAME=otel-worker
ENVIRONMENT=development
GRPC_PORT=50051
HTTP_PORT=8080
POSTGRES_HOST=192.168.1.175
POSTGRES_PORT=5432
POSTGRES_DB=owntracks
POSTGRES_USER=development
POSTGRES_PASSWORD=development
HOME_LATITUDE=40.736097
HOME_LONGITUDE=-74.039373
AWAY_THRESHOLD_KM=0.5
CSV_OUTPUT_PATH=/data/csv
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
LOG_LEVEL=info
ENV_EOF

cat > .gitignore << 'GIT_EOF'
bin/
*.exe
*.test
*.out
.env
.env.local
.idea/
.vscode/
.DS_Store
/data/
*.csv
*.log
GIT_EOF

echo "✓ Bootstrap complete!"
echo "Run: ./scripts/setup.sh"
