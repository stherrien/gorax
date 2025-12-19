.PHONY: all build run test clean deps docker-up docker-down migrate help

# Variables
BINARY_API=rflow-api
BINARY_WORKER=rflow-worker
DOCKER_COMPOSE=docker compose -f deployments/docker/docker-compose.yml

# Default target
all: deps build

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build binaries
build:
	go build -o bin/$(BINARY_API) ./cmd/api
	go build -o bin/$(BINARY_WORKER) ./cmd/worker

# Build API only
build-api:
	go build -o bin/$(BINARY_API) ./cmd/api

# Build Worker only
build-worker:
	go build -o bin/$(BINARY_WORKER) ./cmd/worker

# Run API locally
run-api:
	go run ./cmd/api

# Run Worker locally
run-worker:
	go run ./cmd/worker

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint code
lint:
	golangci-lint run ./...

# Format code
fmt:
	go fmt ./...
	gofmt -s -w .

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker commands
docker-up:
	$(DOCKER_COMPOSE) up -d

docker-down:
	$(DOCKER_COMPOSE) down

docker-logs:
	$(DOCKER_COMPOSE) logs -f

docker-ps:
	$(DOCKER_COMPOSE) ps

docker-build:
	$(DOCKER_COMPOSE) build

docker-restart:
	$(DOCKER_COMPOSE) restart

# Database commands
db-up:
	$(DOCKER_COMPOSE) up -d postgres redis

db-migrate:
	@echo "Migrations are auto-applied via docker-compose init scripts"
	@echo "To manually apply: psql -h localhost -U rflow -d rflow -f migrations/001_initial_schema.sql"

db-seed:
	psql -h localhost -U rflow -d rflow -f migrations/002_seed_data.sql

db-reset:
	$(DOCKER_COMPOSE) down -v
	$(DOCKER_COMPOSE) up -d postgres
	sleep 5
	$(MAKE) db-seed

# Kratos commands
kratos-up:
	$(DOCKER_COMPOSE) up -d kratos kratos-migrate mailslurper

# Development environment (simplified - no Kratos)
dev-simple:
	docker compose -f docker-compose.dev.yml up -d
	@echo "Simplified development environment started (postgres + redis only)"
	@echo "Postgres: localhost:5433 (user: postgres, password: postgres, db: rflow)"
	@echo "Redis: localhost:6379"
	@echo ""
	@echo "Next steps:"
	@echo "  1. make run-api-dev    # Start API server on :8080"
	@echo "  2. cd web && npm run dev  # Start frontend on :5173"

dev-simple-down:
	docker compose -f docker-compose.dev.yml down

# Development environment (full stack with Kratos)
dev: docker-up
	@echo "Development environment started"
	@echo "API: http://localhost:8080"
	@echo "Kratos Public: http://localhost:4433"
	@echo "Kratos Admin: http://localhost:4434"
	@echo "MailSlurper: http://localhost:4437"

# Run API in development mode (uses dev auth, bypasses Kratos)
run-api-dev:
	@echo "Starting API in development mode..."
	@echo "Dev auth enabled - X-Tenant-ID header will be used"
	APP_ENV=development go run ./cmd/api

dev-api: db-up
	go run ./cmd/api

dev-worker: db-up
	go run ./cmd/worker

# Web frontend commands
web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

# Full stack development
dev-full:
	$(MAKE) docker-up
	@echo "Waiting for services to start..."
	sleep 10
	@echo "Starting frontend..."
	cd web && npm run dev &
	@echo "Full stack is running!"
	@echo "API: http://localhost:8080"
	@echo "Web: http://localhost:5173"

# Generate API documentation
docs:
	@echo "TODO: Generate API documentation"

# Help
help:
	@echo "rflow - Workflow Automation Platform"
	@echo ""
	@echo "Usage:"
	@echo "  make deps          - Install Go dependencies"
	@echo "  make build         - Build all binaries"
	@echo "  make run-api       - Run API server locally"
	@echo "  make run-worker    - Run worker locally"
	@echo "  make test          - Run tests"
	@echo "  make lint          - Run linter"
	@echo "  make fmt           - Format code"
	@echo "  make clean         - Clean build artifacts"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up     - Start all Docker services"
	@echo "  make docker-down   - Stop all Docker services"
	@echo "  make docker-logs   - View Docker logs"
	@echo "  make dev           - Start development environment"
	@echo ""
	@echo "Database:"
	@echo "  make db-up         - Start database services"
	@echo "  make db-seed       - Seed database with sample data"
	@echo "  make db-reset      - Reset database"
	@echo ""
	@echo "Frontend:"
	@echo "  make web-install   - Install frontend dependencies"
	@echo "  make web-dev       - Start frontend dev server"
	@echo "  make web-build     - Build frontend for production"
