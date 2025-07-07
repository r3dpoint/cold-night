# Makefile for Securities Marketplace
.PHONY: help setup test test-unit test-integration test-coverage test-benchmark build clean run dev migrate-up migrate-down lint format check

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Setup development environment
setup: ## Setup development environment
	@echo "Setting up development environment..."
	go mod download
	go install github.com/air-verse/air@latest
	go install gotest.tools/gotestsum@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Development environment setup complete!"

# Testing targets
test: ## Run all tests with coverage
	@echo "Running all tests..."
	gotestsum --format testname -- -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Tests complete! Coverage report generated at coverage.html"

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	gotestsum --format testname -- -short -race ./domains/...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	gotestsum --format testname -- -run Integration ./...

test-coverage: ## Generate detailed coverage report
	@echo "Generating coverage report..."
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | grep total:

test-benchmark: ## Run benchmark tests
	@echo "Running benchmark tests..."
	go test -bench=. -benchmem ./domains/...

test-watch: ## Run tests in watch mode
	@echo "Running tests in watch mode..."
	gotestsum --watch -- -short ./domains/...

# Domain-specific testing
test-users: ## Test users domain only
	@echo "Testing users domain..."
	gotestsum --format testname -- -race ./domains/users/...

test-trading: ## Test trading domain only
	@echo "Testing trading domain..."
	gotestsum --format testname -- -race ./domains/trading/...

test-execution: ## Test execution domain only
	@echo "Testing execution domain..."
	gotestsum --format testname -- -race ./domains/trading/execution/...

# Build targets
build: ## Build production binaries
	@echo "Building application..."
	CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o bin/api cmd/api/main.go
	CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o bin/worker cmd/worker/main.go
	@echo "Build complete! Binaries in bin/"

build-dev: ## Build development binaries with debug info
	@echo "Building for development..."
	go build -race -o bin/api-dev cmd/api/main.go
	go build -race -o bin/worker-dev cmd/worker/main.go

# Development targets
dev: ## Run development server with live reload
	@echo "Starting development server..."
	air -c .air.toml

run: ## Run the application directly
	@echo "Starting application..."
	go run cmd/api/main.go

# Database targets
migrate-up: ## Run database migrations up
	@echo "Running database migrations..."
	migrate -verbose -path migrations/ -database "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" up

migrate-down: ## Run database migrations down
	@echo "Reverting database migrations..."
	migrate -verbose -path migrations/ -database "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" down

migrate-create: ## Create new migration (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=migration_name"; exit 1; fi
	migrate create -ext sql -dir migrations $(NAME)

# Code quality targets
lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

lint-fix: ## Run linter with auto-fix
	@echo "Running linter with auto-fix..."
	golangci-lint run --fix

format: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

check: lint vet test-unit ## Run all code quality checks

# Container targets
podman-dev: ## Start development environment with Podman
	@echo "Starting development environment..."
	podman-compose up -d postgres redis
	make dev

podman-test: ## Run tests with Podman
	@echo "Running tests with containers..."
	podman-compose up -d postgres redis
	sleep 5
	make test

podman-down: ## Stop Podman containers
	@echo "Stopping containers..."
	podman-compose down

# Performance testing
perf-test: ## Run performance tests
	@echo "Running performance tests..."
	go test -bench=. -benchmem ./domains/...

profile-cpu: ## Run CPU profiling
	@echo "Running CPU profiling..."
	go test -cpuprofile=cpu.prof -bench=. ./domains/trading/execution
	go tool pprof cpu.prof

profile-mem: ## Run memory profiling
	@echo "Running memory profiling..."
	go test -memprofile=mem.prof -bench=. ./domains/trading/execution
	go tool pprof mem.prof

# Utility targets
clean: ## Clean build artifacts and temporary files
	@echo "Cleaning up..."
	rm -rf bin/ coverage.out coverage.html
	rm -f cpu.prof mem.prof
	go clean -testcache
	@echo "Cleanup complete!"

deps: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

verify: ## Verify project integrity
	@echo "Verifying project..."
	go mod verify
	go mod tidy
	make test-unit
	make build
	@echo "Project verification complete!"

# Quick development workflow
quick-test: ## Quick test for fast feedback
	@echo "Running quick tests..."
	go test -short -race ./domains/shared/testutil/... ./domains/users/... ./domains/trading/execution/...

fix: format lint-fix ## Fix common code issues

# CI/CD simulation
ci: ## Simulate CI pipeline
	@echo "Running CI pipeline..."
	make check
	make test-coverage
	make build
	@echo "CI pipeline complete!"
