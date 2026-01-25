#!/bin/bash
set -e

echo "=== otel-worker Setup Script ==="
echo "This script will install Go SDK, protobuf compiler, and initialize the project"
echo ""

# Detect OS
OS=$(uname -s)
# ARCH is available for future architecture-specific logic
# shellcheck disable=SC2034
ARCH=$(uname -m)

# Check if Go is already installed
if command -v go &> /dev/null; then
    INSTALLED_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo "✓ Go is already installed: $INSTALLED_VERSION"

    # Check if version is >= 1.23
    REQUIRED="1.23"
    if [ "$(printf '%s\n' "$REQUIRED" "$INSTALLED_VERSION" | sort -V | head -n1)" = "$REQUIRED" ]; then
        echo "  Version is sufficient (>= 1.23)"
        GO_INSTALLED=true
    else
        echo "  Version $INSTALLED_VERSION is too old, need >= 1.23"
        GO_INSTALLED=false
    fi
else
    echo "✗ Go is not installed"
    GO_INSTALLED=false
fi

# Install/Update Go if needed
if [ "$GO_INSTALLED" = false ]; then
    echo ""
    echo "Installing Go 1.23.5..."

    if [ "$OS" = "Linux" ]; then
        GO_VERSION="1.23.5"
        GO_TARBALL="go${GO_VERSION}.linux-amd64.tar.gz"

        echo "  Downloading Go $GO_VERSION for Linux..."
        wget -q "https://go.dev/dl/$GO_TARBALL" -O "/tmp/$GO_TARBALL"

        echo "  Removing old Go installation (if exists)..."
        sudo rm -rf /usr/local/go

        echo "  Extracting to /usr/local/go..."
        sudo tar -C /usr/local -xzf "/tmp/$GO_TARBALL"

        # Detect shell and add to PATH if not already there
        SHELL_RC_FILES=()

        # Check for bash
        if [ -f "$HOME/.bashrc" ]; then
            SHELL_RC_FILES+=("$HOME/.bashrc")
        fi

        # Check for zsh
        if [ -n "$ZSH_VERSION" ] || [ "$SHELL" = "/bin/zsh" ] || [ "$SHELL" = "/usr/bin/zsh" ]; then
            if [ -f "$HOME/.zshrc" ]; then
                SHELL_RC_FILES+=("$HOME/.zshrc")
            else
                touch "$HOME/.zshrc"
                SHELL_RC_FILES+=("$HOME/.zshrc")
            fi
        fi

        # Add Go to PATH in all detected shell configs
        for RC_FILE in "${SHELL_RC_FILES[@]}"; do
            if ! grep -q "/usr/local/go/bin" "$RC_FILE"; then
                echo '' >> "$RC_FILE"
                echo '# Go programming language' >> "$RC_FILE"
                echo 'export PATH=$PATH:/usr/local/go/bin' >> "$RC_FILE"
                echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> "$RC_FILE"
                echo "  Added Go to PATH in $RC_FILE"
            fi
        done

        export PATH=$PATH:/usr/local/go/bin
        GOPATH_BIN=$(go env GOPATH)/bin
        export PATH=$PATH:$GOPATH_BIN

        rm "/tmp/$GO_TARBALL"
        echo "✓ Go $GO_VERSION installed successfully"

    elif [ "$OS" = "Darwin" ]; then
        if command -v brew &> /dev/null; then
            brew install go
            echo "✓ Go installed via Homebrew"
        else
            echo "Error: Homebrew not found. Please install Go manually from https://go.dev/dl/"
            exit 1
        fi
    else
        echo "Error: Unsupported OS: $OS"
        exit 1
    fi
fi

# Verify Go installation
echo ""
echo "Verifying Go installation..."
go version

# Install protobuf compiler
echo ""
echo "Checking protobuf compiler (protoc)..."
if command -v protoc &> /dev/null; then
    echo "✓ protoc is already installed: $(protoc --version)"
else
    echo "Installing protobuf compiler..."

    if [ "$OS" = "Linux" ]; then
        if command -v apt-get &> /dev/null; then
            sudo apt-get update -qq
            sudo apt-get install -y protobuf-compiler
            echo "✓ protoc installed via apt"
        elif command -v yum &> /dev/null; then
            sudo yum install -y protobuf-compiler
            echo "✓ protoc installed via yum"
        else
            echo "Warning: No package manager found. Downloading protoc manually..."
            PROTOC_VERSION="25.2"
            PROTOC_ZIP="protoc-${PROTOC_VERSION}-linux-x86_64.zip"
            wget -q "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/${PROTOC_ZIP}" -O "/tmp/${PROTOC_ZIP}"
            sudo unzip -q "/tmp/${PROTOC_ZIP}" -d /usr/local
            rm "/tmp/${PROTOC_ZIP}"
            echo "✓ protoc installed to /usr/local/bin"
        fi
    elif [ "$OS" = "Darwin" ]; then
        brew install protobuf
        echo "✓ protoc installed via Homebrew"
    fi
fi

# Ensure GOPATH/bin is in PATH
if [ -z "$GOPATH" ]; then
    GOPATH=$(go env GOPATH)
    export GOPATH
fi
export PATH=$PATH:$GOPATH/bin

# Install Go protobuf plugins
echo ""
echo "Installing Go protobuf plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
echo "✓ protoc-gen-go and protoc-gen-go-grpc installed"

# Install goimports for code formatting
echo ""
echo "Installing goimports..."
go install golang.org/x/tools/cmd/goimports@latest
echo "✓ goimports installed"

# Install golangci-lint for linting
echo ""
echo "Checking golangci-lint..."
if command -v golangci-lint &> /dev/null; then
    echo "✓ golangci-lint is already installed: $(golangci-lint --version | head -n1)"
else
    echo "Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    echo "✓ golangci-lint installed"
fi

# Install grpcurl for testing gRPC services
echo ""
echo "Checking grpcurl..."
if command -v grpcurl &> /dev/null; then
    echo "✓ grpcurl is already installed: $(grpcurl --version 2>&1 | head -n1)"
else
    echo "Installing grpcurl..."
    go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
    echo "✓ grpcurl installed"
fi

# Install grpcui for interactive gRPC testing
echo ""
echo "Checking grpcui..."
if command -v grpcui &> /dev/null; then
    echo "✓ grpcui is already installed"
else
    echo "Installing grpcui..."
    go install github.com/fullstorydev/grpcui/cmd/grpcui@latest
    echo "✓ grpcui installed"
fi

# Ensure ~/.local/bin is in PATH for pip user installs
echo ""
echo "Checking ~/.local/bin in PATH..."
SHELL_RC_FILES_LOCAL=()

# Check for bash
if [ -f "$HOME/.bashrc" ]; then
    SHELL_RC_FILES_LOCAL+=("$HOME/.bashrc")
fi

# Check for zsh
if [ -n "$ZSH_VERSION" ] || [ "$SHELL" = "/bin/zsh" ] || [ "$SHELL" = "/usr/bin/zsh" ]; then
    if [ -f "$HOME/.zshrc" ]; then
        SHELL_RC_FILES_LOCAL+=("$HOME/.zshrc")
    else
        touch "$HOME/.zshrc"
        SHELL_RC_FILES_LOCAL+=("$HOME/.zshrc")
    fi
fi

# Add ~/.local/bin to PATH in all detected shell configs
for RC_FILE in "${SHELL_RC_FILES_LOCAL[@]}"; do
    if ! grep -q 'export PATH="\$HOME/.local/bin:\$PATH"' "$RC_FILE" && ! grep -q "export PATH=\"\$HOME/.local/bin:\$PATH\"" "$RC_FILE" && ! grep -q '\$HOME/.local/bin' "$RC_FILE"; then
        echo '' >> "$RC_FILE"
        echo '# User local binaries (pip --user installs)' >> "$RC_FILE"
        echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$RC_FILE"
        echo "  Added ~/.local/bin to PATH in $RC_FILE"
    fi
done

# Export for current session
export PATH="$HOME/.local/bin:$PATH"
echo "✓ ~/.local/bin added to PATH"

# Install pre-commit
echo ""
echo "Checking pre-commit..."
if command -v pre-commit &> /dev/null || [ -x "$HOME/.local/bin/pre-commit" ]; then
    PRE_COMMIT_CMD="pre-commit"
    [ -x "$HOME/.local/bin/pre-commit" ] && PRE_COMMIT_CMD="$HOME/.local/bin/pre-commit"
    echo "✓ pre-commit is already installed: $($PRE_COMMIT_CMD --version)"
else
    echo "Installing pre-commit..."

    # Check if we're in a virtualenv
    if [ -n "$VIRTUAL_ENV" ]; then
        # In virtualenv, install without --user flag
        if command -v pip3 &> /dev/null; then
            pip3 install pre-commit
            echo "✓ pre-commit installed via pip3 (virtualenv)"
        elif command -v pip &> /dev/null; then
            pip install pre-commit
            echo "✓ pre-commit installed via pip (virtualenv)"
        else
            echo "Warning: Could not install pre-commit. Install manually: pip install pre-commit"
        fi
    else
        # Not in virtualenv, use --user flag
        if command -v pip3 &> /dev/null; then
            pip3 install --user pre-commit
            echo "✓ pre-commit installed via pip3"
        elif command -v pip &> /dev/null; then
            pip install --user pre-commit
            echo "✓ pre-commit installed via pip"
        elif [ "$OS" = "Darwin" ] && command -v brew &> /dev/null; then
            brew install pre-commit
            echo "✓ pre-commit installed via Homebrew"
        else
            echo "Warning: Could not install pre-commit. Install manually: pip install pre-commit"
        fi
    fi
fi

# Install pre-commit hooks
echo ""
# Ensure ~/.local/bin is in PATH (where pip --user installs binaries)
export PATH="$HOME/.local/bin:$PATH"

if command -v pre-commit &> /dev/null && [ -f ".pre-commit-config.yaml" ]; then
    echo "Installing pre-commit hooks..."
    pre-commit install
    echo "✓ pre-commit hooks installed"
else
    if ! command -v pre-commit &> /dev/null; then
        echo "⚠ Skipping pre-commit hook installation (pre-commit not found in PATH)"
        echo "  Try running: export PATH=\$HOME/.local/bin:\$PATH && pre-commit install"
    elif [ ! -f ".pre-commit-config.yaml" ]; then
        echo "⚠ Skipping pre-commit hook installation (.pre-commit-config.yaml not found)"
    fi
fi

# Initialize Go module if not exists
echo ""
if [ -f "go.mod" ]; then
    echo "✓ go.mod already exists"
else
    echo "Initializing Go module..."
    go mod init github.com/stuartshay/otel-worker
    echo "✓ go.mod created"
fi

# Install dependencies
echo ""
echo "Installing Go dependencies..."
go get google.golang.org/grpc@latest
go get google.golang.org/protobuf@latest
go get github.com/lib/pq@latest
go get github.com/umahmood/haversine@latest
go get github.com/joho/godotenv@latest
go get github.com/rs/zerolog@latest
go get go.opentelemetry.io/otel@latest
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@latest
go get go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc@latest

echo "✓ Dependencies installed"

# Run go mod tidy
echo ""
echo "Running go mod tidy..."
go mod tidy

# Create .env from .env.example if not exists
echo ""
if [ -f ".env" ]; then
    echo "✓ .env already exists"
else
    if [ -f ".env.example" ]; then
        cp .env.example .env
        echo "✓ Created .env from .env.example"
        echo "  Please update .env with your configuration"
    else
        echo "⚠ .env.example not found, skipping .env creation"
    fi
fi

# Generate protobuf code if proto files exist
echo ""
if [ -f "proto/distance/v1/distance.proto" ]; then
    echo "Generating protobuf code..."
    if make proto 2>/dev/null; then
        echo "✓ Protobuf code generated"
    else
        echo "⚠ Could not generate protobuf code. Run 'make proto' manually"
    fi
else
    echo "⚠ proto/distance/v1/distance.proto not found, skipping generation"
fi

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Next steps:"
echo "  1. Update .env with your database credentials"
echo "  2. Run 'make proto' to generate protobuf code (if not done)"
echo "  3. Implement Go packages in internal/"
echo "  4. Run 'make build' to compile the application"
echo "  5. Run 'make run' to start the server"
echo ""
