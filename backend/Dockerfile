# FutureSignals Backend Dockerfile
# Multi-stage build for minimal image size

# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o signald ./cmd/signald

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /build/signald /app/signald

# Create output directory
RUN mkdir -p /app/output

# Non-root user for security
RUN adduser -D -g '' appuser
RUN chown -R appuser:appuser /app
USER appuser

# Environment defaults
ENV OUTPUT_DIR=/app/output
ENV DEBUG=false

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD pgrep signald || exit 1

ENTRYPOINT ["/app/signald"]
