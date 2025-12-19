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

# Build the binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o futuresignals ./cmd/futuresignals
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o backfill ./cmd/backfill
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o backfill-probability ./cmd/backfill-probability
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o backfill-enrichment ./cmd/backfill-enrichment

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binaries from builder
COPY --from=builder /build/futuresignals /app/futuresignals
COPY --from=builder /build/backfill /app/backfill
COPY --from=builder /build/backfill-probability /app/backfill-probability
COPY --from=builder /build/backfill-enrichment /app/backfill-enrichment

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
