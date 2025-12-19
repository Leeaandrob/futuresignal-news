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
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o futuresignals ./cmd/futuresignals

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /build/futuresignals /app/futuresignals

# Non-root user for security
RUN adduser -D -g '' appuser
RUN chown -R appuser:appuser /app
USER appuser

# Environment defaults
ENV HTTP_ADDR=:8080
ENV DEBUG=false

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/api/health || exit 1

ENTRYPOINT ["/app/futuresignals"]
