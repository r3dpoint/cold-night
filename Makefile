# Makefile
.PHONY: setup test build run clean

# Setup development environment
setup:
	go mod download
	go install github.com/cosmtrek/air@latest
	go install gotest.tools/gotestsum@latest
	make migrate-up

# Run tests with coverage
test:
	gotestsum --format testname -- -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Fast unit tests only
test-unit:
	gotestsum --format testname -- -short -race ./domains/...

# Integration tests
test-integration:
	gotestsum --format testname -- -run Integration ./...

# Database migrations
migrate-up:
	migrate -path migrations -database "postgres://user:pass@localhost:5432/securities?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://user:pass@localhost:5432/securities?sslmode=disable" down

# Development server with live reload
dev:
	air -c .air.toml

# Build production binary
build:
	CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o bin/api cmd/api/main.go
	CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o bin/worker cmd/worker/main.go

# Run with Podman
podman-dev:
	podman-compose up -d postgres redis
	make dev

# Performance testing
perf-test:
	go test -bench=. -benchmem ./domains/...

clean:
	rm -rf bin/ coverage.out coverage.html