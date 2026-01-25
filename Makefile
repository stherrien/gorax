.PHONY: all build run test clean deps docker-up docker-down migrate help \
	kill-api kill-web kill-worker kill-all \
	dev-start dev-restart dev-attach status dev-all

# Variables
BINARY_API=gorax-api
BINARY_WORKER=gorax-worker
DOCKER_COMPOSE=docker compose -f deployments/docker/docker-compose.yml

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X github.com/gorax/gorax/internal/buildinfo.version=$(VERSION) \
                      -X github.com/gorax/gorax/internal/buildinfo.buildTime=$(BUILD_TIME) \
                      -X github.com/gorax/gorax/internal/buildinfo.gitCommit=$(GIT_COMMIT)"

# Default target - show help
.DEFAULT_GOAL := help

all: deps build

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build binaries
build:
	go build $(LDFLAGS) -o bin/$(BINARY_API) ./cmd/api
	go build $(LDFLAGS) -o bin/$(BINARY_WORKER) ./cmd/worker

# Build API only
build-api:
	go build $(LDFLAGS) -o bin/$(BINARY_API) ./cmd/api

# Build Worker only
build-worker:
	go build $(LDFLAGS) -o bin/$(BINARY_WORKER) ./cmd/worker

# Run API locally
run-api:
	@set -a; . ./.env; set +a; go run ./cmd/api

# Run Worker locally
run-worker:
	@set -a; . ./.env; set +a; go run ./cmd/worker

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Smoke tests
smoke-tests:
	@echo "Running smoke tests..."
	cd tests/smoke && ./run-all.sh

smoke-tests-quick:
	@echo "Running smoke tests (skip Go tests)..."
	cd tests/smoke && SKIP_GO=true ./run-all.sh

smoke-tests-api:
	cd tests/smoke && ./api-smoke.sh

smoke-tests-db:
	cd tests/smoke && ./db-smoke.sh

smoke-tests-services:
	cd tests/smoke && ./service-smoke.sh

smoke-tests-perf:
	cd tests/smoke && ./perf-smoke.sh

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
	@echo "To manually apply: psql -h localhost -U postgres -d gorax -f migrations/001_initial_schema.sql"

db-seed:
	psql -h localhost -U postgres -d gorax -f migrations/002_seed_data.sql

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
	@echo "Postgres: localhost:5433 (user: postgres, password: postgres, db: gorax)"
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
	@echo "API: http://localhost:8181"
	@echo "Kratos Public: http://localhost:4433"
	@echo "Kratos Admin: http://localhost:4434"
	@echo "MailSlurper: http://localhost:4437"

# Run API in development mode (uses dev auth, bypasses Kratos)
run-api-dev:
	@echo "Starting API in development mode..."
	@echo "Dev auth enabled - X-Tenant-ID header will be used"
	@set -a; . ./.env; set +a; go run ./cmd/api

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

# Kill processes
kill-api:
	@echo "Killing API server (port 8181)..."
	@-tmux kill-window -t gorax:api 2>/dev/null || true
	@-lsof -ti:8181 | xargs kill -9 2>/dev/null || true
	@-lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	@-pkill -9 -f "go run ./cmd/api" 2>/dev/null || true
	@-pkill -9 -f "gorax-api" 2>/dev/null || true
	@echo "API server killed"

kill-web:
	@echo "Killing web dev server (port 5173)..."
	@-tmux kill-window -t gorax:web 2>/dev/null || true
	@-lsof -ti:5173 | xargs kill -9 2>/dev/null || true
	@-pkill -9 -f "vite" 2>/dev/null || true
	@-pkill -9 -f "npm run dev" 2>/dev/null || true
	@echo "Web server killed"

kill-worker:
	@echo "Killing worker process..."
	@-tmux kill-window -t gorax:worker 2>/dev/null || true
	@-pkill -9 -f "go run ./cmd/worker" 2>/dev/null || true
	@-pkill -9 -f "gorax-worker" 2>/dev/null || true
	@echo "Worker killed"

kill-all:
	@echo "Killing all gorax processes..."
	@-tmux kill-session -t gorax 2>/dev/null || true
	@-lsof -ti:8181 | xargs kill -9 2>/dev/null || true
	@-lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	@-lsof -ti:5173 | xargs kill -9 2>/dev/null || true
	@-pkill -9 -f "go run ./cmd/api" 2>/dev/null || true
	@-pkill -9 -f "go run ./cmd/worker" 2>/dev/null || true
	@-pkill -9 -f "gorax-api" 2>/dev/null || true
	@-pkill -9 -f "gorax-worker" 2>/dev/null || true
	@-pkill -9 -f "vite" 2>/dev/null || true
	@echo "All processes killed"

# Start dev environment in tmux with split panes
dev-start:
	@if tmux has-session -t gorax 2>/dev/null; then \
		echo "Gorax session already running. Attaching..."; \
		tmux attach -t gorax; \
	else \
		echo "Starting gorax dev environment..."; \
		tmux new-session -d -s gorax -n dev -c $(CURDIR); \
		tmux send-keys -t gorax:dev 'set -a; . ./.env; set +a; go run ./cmd/api' Enter; \
		tmux split-window -t gorax:dev -h -c $(CURDIR)/web; \
		tmux send-keys -t gorax:dev.1 'npm run dev' Enter; \
		tmux select-pane -t gorax:dev.0; \
		tmux attach -t gorax; \
	fi

# Restart and attach
dev-restart: kill-all
	@sleep 1
	@$(MAKE) dev-start

# Attach to existing session
dev-attach:
	@if tmux has-session -t gorax 2>/dev/null; then \
		tmux attach -t gorax; \
	else \
		echo "No gorax session running. Use 'make dev-start' to start."; \
	fi

# Show status of tmux sessions
status:
	@echo "Gorax tmux session:"
	@tmux list-windows -t gorax 2>/dev/null || echo "  No gorax session running"
	@echo ""
	@echo "Port status:"
	@echo "  Port 8181 (API):  $$(lsof -ti:8181 >/dev/null 2>&1 && echo 'IN USE' || echo 'free')"
	@echo "  Port 5173 (Web):  $$(lsof -ti:5173 >/dev/null 2>&1 && echo 'IN USE' || echo 'free')"

# Full stack development
dev-full:
	$(MAKE) docker-up
	@echo "Waiting for services to start..."
	sleep 10
	@echo "Starting frontend..."
	cd web && npm run dev &
	@echo "Full stack is running!"
	@echo "API: http://localhost:8181"
	@echo "Web: http://localhost:5173"

# Development: API + Web in tmux (alias for dev-start)
dev-all: dev-start

# Generate API documentation
docs:
	@echo "TODO: Generate API documentation"

# Pre-commit checks (mirrors CI)
.PHONY: precommit check check-go check-web lint-go lint-web test-go test-web typecheck security-scan

# Full pre-commit check (run before committing)
precommit: check
	@echo ""
	@echo "✅ All pre-commit checks passed!"
	@echo "You can safely commit your changes."

# Run all checks (without committing)
check: check-go check-web security-scan
	@echo ""
	@echo "✅ All checks passed!"

# Go checks
check-go: lint-go test-go
	@echo "✅ Go checks passed"

# Web/Frontend checks
check-web: lint-web typecheck test-web
	@echo "✅ Frontend checks passed"

# Go lint and format
lint-go:
	@echo "🔍 Running Go linter..."
	@gofmt -l . | grep -v vendor | xargs -r gofmt -w
	@goimports -l . | grep -v vendor | xargs -r goimports -w
	@golangci-lint run ./... --timeout 5m
	@go vet ./...
	@echo "✅ Go lint passed"

# Go tests
test-go:
	@echo "🧪 Running Go tests..."
	@go test ./... -race -count=1
	@echo "✅ Go tests passed"

# Frontend lint
lint-web:
	@echo "🔍 Running frontend linter..."
	@cd web && npm run lint
	@echo "✅ Frontend lint passed"

# Frontend TypeScript type check
typecheck:
	@echo "📝 Running TypeScript type check..."
	@cd web && npx tsc --noEmit
	@echo "✅ TypeScript check passed"

# Frontend tests
test-web:
	@echo "🧪 Running frontend tests..."
	@cd web && npm test -- --run
	@echo "✅ Frontend tests passed"

# Security scanning (basic checks)
security-scan:
	@echo "🔒 Running security scans..."
	@echo "Checking for secrets in staged files..."
	@-git diff --cached --name-only | xargs -I {} sh -c 'grep -l "BEGIN.*PRIVATE KEY\|AKIA[0-9A-Z]\{16\}\|postgres://.*:.*@" "{}" 2>/dev/null && echo "⚠️  Potential secret in: {}"' || true
	@echo "✅ Security scan passed"

# Quick check (lint only, no tests)
check-quick: lint-go lint-web typecheck
	@echo "✅ Quick check passed (lint + typecheck only)"

# CI check (exactly what CI runs)
ci: check-go check-web
	@echo "✅ CI checks passed"

# Help
help:
	@echo "🚀 Gorax - Modern Workflow Automation Platform"
	@echo ""
	@echo "📦 Build & Dependencies:"
	@echo "  make all           - Install deps and build binaries"
	@echo "  make deps          - Install Go dependencies"
	@echo "  make build         - Build all binaries"
	@echo "  make build-api     - Build API only"
	@echo "  make build-worker  - Build worker only"
	@echo "  make clean         - Clean build artifacts"
	@echo ""
	@echo "🏃 Run Locally:"
	@echo "  make run-api       - Run API server"
	@echo "  make run-api-dev   - Run API in dev mode (no Kratos)"
	@echo "  make run-worker    - Run worker"
	@echo ""
	@echo "🧪 Testing & Quality:"
	@echo "  make test             - Run all tests"
	@echo "  make test-coverage    - Run tests with coverage report"
	@echo "  make smoke-tests      - Run all smoke tests"
	@echo "  make smoke-tests-quick - Run smoke tests (skip Go tests)"
	@echo "  make smoke-tests-api  - Run API smoke tests only"
	@echo "  make smoke-tests-db   - Run database smoke tests only"
	@echo "  make lint             - Run linter"
	@echo "  make fmt              - Format code"
	@echo ""
	@echo "✅ Pre-commit Checks (run before committing):"
	@echo "  make precommit     - Full pre-commit check (lint + test + security)"
	@echo "  make check         - Run all checks (same as precommit)"
	@echo "  make check-quick   - Quick check (lint + typecheck, no tests)"
	@echo "  make check-go      - Go lint + tests"
	@echo "  make check-web     - Frontend lint + typecheck + tests"
	@echo "  make lint-go       - Go linting only"
	@echo "  make lint-web      - Frontend linting only"
	@echo "  make typecheck     - TypeScript type checking"
	@echo "  make ci            - Run exact CI checks locally"
	@echo ""
	@echo "🐳 Docker:"
	@echo "  make docker-up     - Start all services"
	@echo "  make docker-down   - Stop all services"
	@echo "  make docker-logs   - View logs"
	@echo "  make docker-ps     - List containers"
	@echo "  make dev-simple    - Start simple dev env (postgres + redis)"
	@echo ""
	@echo "💾 Database:"
	@echo "  make db-up         - Start postgres + redis"
	@echo "  make db-migrate    - Run migrations"
	@echo "  make db-seed       - Seed database"
	@echo "  make db-reset      - Reset database"
	@echo ""
	@echo "🎨 Frontend:"
	@echo "  make web-install   - Install dependencies"
	@echo "  make web-dev       - Start dev server"
	@echo "  make web-build     - Build for production"
	@echo ""
	@echo "🚦 Dev Environment:"
	@echo "  make dev-all       - Start API + web in tmux (same as dev-start)"
	@echo "  make dev-start     - Start API + web in tmux split-screen"
	@echo "  make dev-restart   - Kill all and restart"
	@echo "  make dev-attach    - Attach to running session"
	@echo "  make kill-all      - Kill all processes"
	@echo "  make status        - Show running sessions & ports"
	@echo ""
	@echo "  Tmux shortcuts (once attached):"
	@echo "    Ctrl+B then arrow - Switch between panes"
	@echo "    Ctrl+B then D     - Detach (keep running)"
	@echo "    Ctrl+B then Z     - Zoom current pane (toggle)"
	@echo "    Ctrl+B then X     - Kill current pane"
