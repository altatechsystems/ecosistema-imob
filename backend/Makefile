.PHONY: help build run test clean install dev

help:
	@echo "Available commands:"
	@echo "  make install    - Install dependencies"
	@echo "  make build      - Build the server binary"
	@echo "  make run        - Run the server"
	@echo "  make dev        - Run the server in development mode with auto-reload"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Clean build artifacts"

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

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean
