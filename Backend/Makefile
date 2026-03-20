.PHONY: help build run test test-verbose test-coverage test-module clean

help:
	@echo "Pulse Backend - Available Commands"
	@echo "===================================="
	@echo "make build              - Build the application"
	@echo "make run                - Run the application"
	@echo "make test               - Run all tests"
	@echo "make test-verbose       - Run tests with verbose output"
	@echo "make test-coverage      - Run tests with coverage report"
	@echo "make test-module MOD=<package> - Run tests for specific module (e.g., MOD=crypto)"
	@echo "make clean              - Remove build artifacts"
	@echo "make help               - Show this help message"

build:
	@echo "Building application..."
	go build -o pulse-backend ./cmd/server
	@echo "✅ Build complete: ./pulse-backend"

run: build
	@echo "Starting server..."
	./pulse-backend

test:
	@echo "Running all tests..."
	go test -v ./...
	@echo "✅ All tests passed"

test-verbose:
	@echo "Running tests with verbose output..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "✅ Tests complete"

test-coverage: test-verbose
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

test-module:
	@if [ -z "$(MOD)" ]; then \
		echo "Error: Please specify MOD variable"; \
		echo "Example: make test-module MOD=crypto"; \
		exit 1; \
	fi
	@echo "Running tests for module: $(MOD)"
	go test -v ./internal/$(MOD) ./cmd/server
	@echo "✅ Module tests passed"

clean:
	@echo "Cleaning up..."
	rm -f pulse-backend coverage.out coverage.html
	go clean
	@echo "✅ Clean complete"
