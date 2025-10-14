# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy Go workspace files
COPY ../../go.work ../../go.work.sum ./

# Copy proto module (needed for builds)
COPY ../../proto ../proto

# Copy service files
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -a -installsuffix cgo \
    -o /build/bin/accounts \
    ./cmd/accounts

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS and curl for health checks
RUN apk --no-cache add ca-certificates curl

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bin/accounts .

# Create non-root user
RUN addgroup -g 1001 appuser && \
    adduser -D -u 1001 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 7101

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:7101/healthz || exit 1

CMD ["./accounts"]
