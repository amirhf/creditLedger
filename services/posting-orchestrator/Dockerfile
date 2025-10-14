# Build stage
FROM golang:1.24-alpine AS builder

# Install git for go mod tidy
RUN apk add --no-cache git

WORKDIR /build

# Copy proto and common modules to expected locations
WORKDIR /
COPY proto ./proto
COPY services/common ./common

# Set working directory for service
WORKDIR /build

# Copy service files
COPY services/posting-orchestrator/ .

# Download dependencies and build
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -a -installsuffix cgo \
    -o /build/bin/orchestrator \
    ./cmd/orchestrator

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS and curl for health checks
RUN apk --no-cache add ca-certificates curl

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bin/orchestrator .

# Create non-root user
RUN addgroup -g 1001 appuser && \
    adduser -D -u 1001 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 7103

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:7103/healthz || exit 1

CMD ["./orchestrator"]
