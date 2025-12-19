# FutureSignals Backend Makefile

.PHONY: all build run test clean deps lint docker-build docker-run help

# Variables
BINARY_NAME=signald
BUILD_DIR=./build
CMD_DIR=./cmd/signald
GO=go

# Build flags
LDFLAGS=-ldflags "-s -w"
CGO_ENABLED=0

# Default target
all: deps build

## help: Show this help message
help:
	@echo "FutureSignals Backend"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'

## deps: Download dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

## build: Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

## run: Run the application
run:
	$(GO) run $(CMD_DIR)/main.go

## dev: Run with live reload (requires air)
dev:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	air

## test: Run tests
test:
	$(GO) test -v -race ./...

## test-coverage: Run tests with coverage
test-coverage:
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run linter
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

## clean: Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## docker-build: Build Docker image
docker-build:
	docker build -t futuresignals-backend:latest .

## docker-run: Run Docker container
docker-run:
	docker run --rm -it \
		--env-file .env \
		-v $(PWD)/output:/app/output \
		futuresignals-backend:latest

## fetch-markets: Quick test - fetch top markets
fetch-markets:
	@$(GO) run $(CMD_DIR)/main.go --mode=fetch-once 2>/dev/null || \
		echo "Run with: DASHSCOPE_API_KEY=xxx make run"

## setup: Initial setup
setup: deps
	@cp -n .env.example .env 2>/dev/null || true
	@mkdir -p output
	@echo "Setup complete. Edit .env with your credentials."
