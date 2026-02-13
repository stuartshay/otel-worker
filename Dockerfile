# Build stage
FROM golang:1.26-alpine AS builder

# Install build dependencies
# hadolint ignore=DL3018
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for multi-arch support
ARG TARGETARCH

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build \
    -ldflags="-w -s" \
    -o otel-worker \
    cmd/server/main.go

# Create directory structure for CSV output
RUN mkdir -p /data/csv && chmod 755 /data/csv

# Runtime stage
FROM gcr.io/distroless/static-debian12:nonroot

# Copy timezone data and CA certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary from builder
COPY --from=builder /build/otel-worker /app/otel-worker

# Copy CSV directory with proper ownership
COPY --from=builder --chown=nonroot:nonroot /data/csv /data/csv

WORKDIR /app

# Expose gRPC and HTTP ports
EXPOSE 50051 8080

# Set environment defaults
ENV GRPC_PORT=50051
ENV HTTP_PORT=8080
ENV CSV_OUTPUT_PATH=/data/csv
ENV LOG_LEVEL=info

# Health check (requires grpc_health_probe in production)
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#   CMD ["/bin/grpc_health_probe", "-addr=:50051"]

ENTRYPOINT ["/app/otel-worker"]
