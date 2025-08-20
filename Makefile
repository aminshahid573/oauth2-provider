# File: Makefile
.PHONY: all build run test tidy docker-up docker-down help

BINARY_NAME=oauth2-provider

all: build

# Build the Go application
build:
	@echo "Building binary..."
	@go build -o bin/$(BINARY_NAME) ./cmd/server

# Run the Go application
run:
	@echo "Running application..."
	@go run ./cmd/server

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Tidy go.mod and go.sum
tidy:
	@echo "Tidying modules..."
	@go mod tidy

# Start Docker containers
docker-up:
	@echo "Starting Docker services..."
	@docker-compose -f docker/docker-compose.yml up -d

# Stop Docker containers
docker-down:
	@echo "Stopping Docker services..."
	@docker-compose -f docker/docker-compose.yml down

# Display help
help:
	@echo "Available commands:"
	@echo "  build        - Build the application binary"
	@echo "  run          - Run the application"
	@echo "  test         - Run all tests"
	@echo "  tidy         - Tidy go modules"
	@echo "  docker-up    - Start required Docker services (MongoDB, Redis)"
	@echo "  docker-down  - Stop Docker services"