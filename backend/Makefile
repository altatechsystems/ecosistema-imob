.PHONY: help build run test test-unit test-integration test-all test-coverage clean install dev

help:
	@echo "Available commands:"
	@echo "  make install           - Install dependencies"
	@echo "  make build             - Build the server binary"
	@echo "  make run               - Run the server"
	@echo "  make dev               - Run the server in development mode"
	@echo "  make test              - Run all tests (unit + integration)"
	@echo "  make test-unit         - Run unit tests only"
	@echo "  make test-integration  - Run integration tests only"
	@echo "  make test-coverage     - Run tests with coverage report"
	@echo "  make clean             - Clean build artifacts"

install:
	go mod download
	go mod tidy

build:
	@echo "Building server..."
	go build -o bin/server.exe ./cmd/server

run: build
	@echo "Starting server..."
	./bin/server.exe

dev:
	@echo "Starting server in development mode..."
	go run ./cmd/server/main.go

test-unit:
	@echo "Running unit tests..."
	go test -v ./internal/...

test-integration:
	@echo "Running integration tests..."
	@echo "NOTE: This requires Firebase credentials configured"
	set RUN_INTEGRATION_TESTS=1 && go test -v -timeout 5m ./tests/integration/...

test: test-unit
	@echo "All unit tests completed"

test-all:
	@echo "Running all tests (unit + integration)..."
	set RUN_INTEGRATION_TESTS=1 && go test -v -timeout 5m ./...

test-coverage:
	@echo "Running tests with coverage..."
	set RUN_INTEGRATION_TESTS=1 && go test -v -timeout 5m -cover -coverprofile=coverage.out ./...
	@echo "Generating HTML coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean
