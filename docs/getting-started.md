# Getting Started with Gorax

This comprehensive guide will walk you through setting up Gorax for local development, from prerequisites to creating your first workflow.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation Methods](#installation-methods)
- [Quick Start (Docker Compose)](#quick-start-docker-compose)
- [Manual Installation](#manual-installation)
- [Environment Configuration](#environment-configuration)
- [Verification](#verification)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Troubleshooting](#troubleshooting)
- [Platform-Specific Instructions](#platform-specific-instructions)
- [Next Steps](#next-steps)

## Prerequisites

Before you begin, ensure you have the following installed:

### Required Software

| Tool | Minimum Version | Recommended | Purpose |
|------|-----------------|-------------|---------|
| **Go** | 1.21+ | 1.21.5+ | Backend API |
| **Node.js** | 18+ | 20 LTS | Frontend UI |
| **PostgreSQL** | 14+ | 15+ | Primary database |
| **Redis** | 6+ | 7+ | Caching & queues |
| **Git** | 2.30+ | Latest | Version control |
| **Docker** (optional) | 20+ | Latest | Container runtime |
| **Docker Compose** (optional) | 2.0+ | Latest | Multi-container orchestration |

### Check Your Versions

```bash
# Go
go version
# Should output: go version go1.21.x ...

# Node.js
node --version
# Should output: v20.x.x or higher

# PostgreSQL
psql --version
# Should output: psql (PostgreSQL) 15.x or higher

# Redis
redis-server --version
# Should output: Redis server v=7.x.x or higher

# Git
git --version
# Should output: git version 2.x.x

# Docker (if using Docker method)
docker --version
docker-compose --version
```

### Optional Tools

- **[Air](https://github.com/cosmtrek/air)**: Hot reload for Go (development)
- **[golangci-lint](https://golangci-lint.run/)**: Go linter
- **[pgAdmin](https://www.pgadmin.org/)** or **[Postico](https://eggerapps.at/postico/)**: PostgreSQL GUI
- **[RedisInsight](https://redis.com/redis-enterprise/redis-insight/)**: Redis GUI
- **[Postman](https://www.postman.com/)** or **[HTTPie](https://httpie.io/)**: API testing

## Installation Methods

Choose the installation method that best fits your needs:

| Method | Best For | Setup Time | Flexibility |
|--------|----------|------------|-------------|
| **Docker Compose** | Quick start, isolated environment | 5 min | Low |
| **Manual Installation** | Full control, production-like setup | 15 min | High |
| **Development Mode** | Active development, debugging | 10 min | Very High |

## Quick Start (Docker Compose)

The fastest way to get Gorax running locally.

### 1. Clone the Repository

```bash
git clone https://github.com/stherrien/gorax.git
cd gorax
```

### 2. Start All Services

```bash
# Start PostgreSQL, Redis, API, and Frontend
docker-compose up -d

# View logs
docker-compose logs -f

# Check status
docker-compose ps
```

This starts:
- **PostgreSQL** on port `5432`
- **Redis** on port `6379`
- **API Server** on port `8080`
- **Frontend** on port `5173`

### 3. Verify Installation

```bash
# Health check
curl http://localhost:8080/health

# Open UI
open http://localhost:5173
```

### 4. Stop Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (data will be lost)
docker-compose down -v
```

**Pros:**
- ✅ Quick setup
- ✅ Isolated environment
- ✅ No local dependencies needed

**Cons:**
- ❌ Slower iteration (rebuild on code changes)
- ❌ Harder to debug
- ❌ Less control over configuration

## Manual Installation

For full control and production-like setup.

### Step 1: Clone and Setup

```bash
# Clone repository
git clone https://github.com/stherrien/gorax.git
cd gorax

# Create a dedicated directory for Gorax
mkdir -p ~/gorax
cd ~/gorax
```

### Step 2: Install Go Dependencies

```bash
# Download Go modules
go mod download

# Verify dependencies
go mod verify

# Install development tools
go install github.com/cosmtrek/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Step 3: Setup PostgreSQL

#### Option A: Docker

```bash
# Start PostgreSQL container
docker run -d \
  --name gorax-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=gorax \
  -p 5432:5432 \
  postgres:15-alpine

# Verify connection
docker exec -it gorax-postgres psql -U postgres -d gorax -c "SELECT version();"
```

#### Option B: Local PostgreSQL

```bash
# Create database
createdb gorax

# Create user (if needed)
createuser -P gorax_user
# Enter password: gorax_pass

# Grant permissions
psql -c "GRANT ALL PRIVILEGES ON DATABASE gorax TO gorax_user;"
```

### Step 4: Setup Redis

#### Option A: Docker

```bash
# Start Redis container
docker run -d \
  --name gorax-redis \
  -p 6379:6379 \
  redis:7-alpine

# Verify connection
docker exec -it gorax-redis redis-cli ping
# Should output: PONG
```

#### Option B: Local Redis

```bash
# macOS (Homebrew)
brew install redis
brew services start redis

# Linux (Ubuntu/Debian)
sudo apt-get install redis-server
sudo systemctl start redis-server

# Verify
redis-cli ping
# Should output: PONG
```

### Step 5: Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit configuration
# Use your preferred editor: vim, nano, code, etc.
vim .env
```

Edit `.env` with your settings (see [Environment Configuration](#environment-configuration) below).

### Step 6: Run Database Migrations

```bash
# Apply all migrations
make migrate-up

# Verify migrations
psql -U postgres -d gorax -c "\dt"
# You should see: tenants, workflows, executions, webhooks, etc.
```

### Step 7: Install Frontend Dependencies

```bash
cd web

# Install npm packages
npm install

# Verify installation
npm list --depth=0
```

### Step 8: Start Backend API

```bash
# From project root
make run-api

# Or with hot reload (development)
air

# Verify API
curl http://localhost:8080/health
```

### Step 9: Start Frontend

```bash
# From web/ directory
cd web
npm run dev

# Open browser
open http://localhost:5173
```

## Environment Configuration

### Essential Variables

Edit `.env` with these required variables:

```env
#=============================================================================
# Core Configuration
#=============================================================================

# Application Environment
APP_ENV=development                    # development | staging | production
LOG_LEVEL=debug                        # debug | info | warn | error

# Server Configuration
SERVER_ADDRESS=:8080                   # API server port
FRONTEND_URL=http://localhost:5173     # Frontend URL for CORS

#=============================================================================
# Database Configuration
#=============================================================================

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gorax
DB_SSL_MODE=disable                    # disable | require | verify-full

# Connection Pool
DB_MAX_OPEN_CONNS=25                   # Max open connections
DB_MAX_IDLE_CONNS=5                    # Max idle connections
DB_MAX_CONN_LIFETIME=5m                # Max connection lifetime

#=============================================================================
# Redis Configuration
#=============================================================================

REDIS_ADDRESS=localhost:6379
REDIS_PASSWORD=                        # Leave empty if no password
REDIS_DB=0                             # Redis database number
REDIS_POOL_SIZE=10                     # Connection pool size

#=============================================================================
# Authentication (Ory Kratos)
#=============================================================================

# Kratos Public Endpoint
KRATOS_PUBLIC_URL=http://localhost:4433

# Kratos Admin Endpoint
KRATOS_ADMIN_URL=http://localhost:4434

# Session Management
SESSION_COOKIE_NAME=gorax_session
SESSION_COOKIE_DOMAIN=localhost
SESSION_COOKIE_SECURE=false            # true in production
SESSION_MAX_AGE=86400                  # 24 hours

#=============================================================================
# Security
#=============================================================================

# JWT Configuration
JWT_SECRET=your-secret-key-change-me-in-production
JWT_EXPIRATION=24h

# Encryption
MASTER_KEY=32-byte-encryption-key-change-in-production

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
CORS_ALLOW_CREDENTIALS=true

#=============================================================================
# Worker Configuration
#=============================================================================

WORKER_CONCURRENCY=10                  # Max parallel workflow executions
WORKER_MAX_CONCURRENCY_PER_TENANT=5    # Per-tenant limit
WORKER_HEALTH_PORT=:8081              # Health check port

#=============================================================================
# Queue Configuration (SQS)
#=============================================================================

QUEUE_URL=http://localhost:9324/queue/gorax-executions  # LocalStack
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
AWS_ENDPOINT=http://localhost:9324    # LocalStack endpoint

#=============================================================================
# Observability
#=============================================================================

# OpenTelemetry
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
OTEL_SERVICE_NAME=gorax-api
OTEL_TRACES_ENABLED=true
OTEL_METRICS_ENABLED=true

# Sentry (Error Tracking)
SENTRY_DSN=                            # Optional: Add your Sentry DSN
SENTRY_ENVIRONMENT=development
SENTRY_TRACES_SAMPLE_RATE=1.0

#=============================================================================
# Rate Limiting
#=============================================================================

RATE_LIMIT_REQUESTS_PER_MINUTE=60
RATE_LIMIT_REQUESTS_PER_HOUR=1000
RATE_LIMIT_REQUESTS_PER_DAY=10000

#=============================================================================
# Feature Flags
#=============================================================================

FEATURE_WEBHOOK_REPLAY=true
FEATURE_COLLABORATION=true
FEATURE_ANALYTICS=true
FEATURE_MARKETPLACE=true

#=============================================================================
# Development Only
#=============================================================================

# Enable debug features
DEBUG_SQL=false                        # Log all SQL queries
DEBUG_HTTP=false                       # Log HTTP requests/responses
HOT_RELOAD=true                        # Enable hot reload

# Mock external services
MOCK_KRATOS=false                      # Use mock authentication
MOCK_SQS=false                         # Use in-memory queue
```

### Environment Variable Reference

See full reference: [docs/configuration.md](configuration.md)

### Validation

Verify your environment configuration:

```bash
# Check required variables are set
make validate-env

# Test database connection
make test-db

# Test Redis connection
make test-redis
```

## Verification

### 1. Check Services

```bash
# PostgreSQL
psql -U postgres -d gorax -c "SELECT 1;"
# Expected: Returns 1

# Redis
redis-cli ping
# Expected: PONG

# API Server
curl http://localhost:8080/health
# Expected: {"status":"healthy","version":"x.x.x"}

# Frontend
curl http://localhost:5173
# Expected: HTML response
```

### 2. Run Health Checks

```bash
# Full system health check
make health-check

# Individual component checks
make check-api      # API server
make check-db       # Database
make check-redis    # Redis
make check-frontend # Frontend build
```

### 3. Verify Database Schema

```bash
# List all tables
psql -U postgres -d gorax -c "\dt"

# Expected tables:
# - tenants
# - users
# - workflows
# - executions
# - execution_steps
# - webhooks
# - webhook_events
# - webhook_filters
# - credentials
# - schedules
# - templates
# - integrations
```

### 4. Test API Endpoints

```bash
# Health endpoint
curl http://localhost:8080/health

# API version
curl http://localhost:8080/api/v1/version

# Metrics (Prometheus format)
curl http://localhost:8080/metrics
```

### 5. Open Web UI

1. Navigate to http://localhost:5173
2. You should see the Gorax login page
3. Create an account or log in

## Development Workflow

### Running Services Individually

```bash
# Terminal 1: PostgreSQL (if using Docker)
docker start gorax-postgres

# Terminal 2: Redis (if using Docker)
docker start gorax-redis

# Terminal 3: API Server with hot reload
air

# Terminal 4: Frontend with hot reload
cd web && npm run dev
```

### Running Tests

```bash
# Backend unit tests
make test

# Backend integration tests (requires TEST_DATABASE_URL)
TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/gorax_test make test-integration

# Frontend tests
cd web && npm test

# Run specific test
go test ./internal/workflow -v -run TestValidateWorkflow

# Test with coverage
make test-coverage
open coverage.html
```

### Linting and Formatting

```bash
# Backend linting
make lint

# Fix auto-fixable issues
make lint-fix

# Frontend linting
cd web && npm run lint

# Format code
make fmt          # Backend (gofmt)
cd web && npm run format  # Frontend (prettier)
```

### Database Management

```bash
# Create a new migration
make migration name=add_webhooks_table

# Apply migrations
make migrate-up

# Rollback last migration
make migrate-down

# Rollback all migrations
make migrate-reset

# Check migration status
make migrate-status

# Seed development data
make seed
```

### Debugging

#### Go Debugging (VS Code)

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch API Server",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/api/main.go",
      "env": {
        "APP_ENV": "development"
      },
      "args": []
    }
  ]
}
```

#### React Debugging (Chrome DevTools)

1. Open http://localhost:5173
2. Open Chrome DevTools (F12)
3. Go to Sources tab
4. Set breakpoints in TypeScript files

#### Database Debugging

```bash
# View PostgreSQL logs
docker logs gorax-postgres -f

# Connect to database shell
psql -U postgres -d gorax

# View slow queries
psql -U postgres -d gorax -c "SELECT * FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"
```

## Testing

### Unit Tests

```bash
# Run all unit tests
make test

# Run with verbose output
go test -v ./...

# Run specific package
go test ./internal/workflow

# Run with race detector
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Setup test database
createdb gorax_test

# Run integration tests
TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/gorax_test \
  go test -v ./internal/webhook -run TestIntegration_

# Run all integration tests
make test-integration
```

### Frontend Tests

```bash
cd web

# Run all tests
npm test

# Run in watch mode
npm test -- --watch

# Run with coverage
npm test -- --coverage

# Run specific test file
npm test -- WorkflowList.test.tsx
```

### End-to-End Tests

```bash
# Install Playwright
cd web
npm run playwright:install

# Run E2E tests
npm run test:e2e

# Run in UI mode
npm run test:e2e:ui
```

## Troubleshooting

### Common Issues

#### 1. Database Connection Failed

**Error**: `connection refused` or `could not connect to server`

**Solutions**:

```bash
# Check if PostgreSQL is running
docker ps | grep postgres
# Or
pg_isready -h localhost -p 5432

# Check PostgreSQL logs
docker logs gorax-postgres

# Verify connection manually
psql -U postgres -h localhost -p 5432 -d gorax

# Check port availability
lsof -i :5432
```

**Configuration Issues**:
- Verify `DB_HOST` and `DB_PORT` in `.env`
- Check `DB_USER` and `DB_PASSWORD`
- Ensure database `gorax` exists: `createdb gorax`

#### 2. Redis Connection Failed

**Error**: `dial tcp: connect: connection refused`

**Solutions**:

```bash
# Check if Redis is running
redis-cli ping
# Expected: PONG

# Check Redis logs
docker logs gorax-redis

# Restart Redis
docker restart gorax-redis
# Or
brew services restart redis  # macOS
sudo systemctl restart redis-server  # Linux
```

#### 3. Port Already in Use

**Error**: `bind: address already in use`

**Solutions**:

```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>

# Or change port in .env
SERVER_ADDRESS=:8081
```

#### 4. Migration Errors

**Error**: `migration failed` or `duplicate table`

**Solutions**:

```bash
# Check current migration version
make migrate-status

# Rollback and retry
make migrate-down
make migrate-up

# Force to specific version
migrate -database ${DATABASE_URL} -path migrations force 20231201000000

# Reset and reapply (WARNING: data loss)
make migrate-reset
make migrate-up
```

#### 5. Frontend Build Errors

**Error**: `npm ERR! code ELIFECYCLE` or module not found

**Solutions**:

```bash
cd web

# Clear cache and reinstall
rm -rf node_modules package-lock.json
npm cache clean --force
npm install

# Check Node version
node --version  # Should be 18+

# Use specific Node version (nvm)
nvm use 20
npm install
```

#### 6. API Returns 401 Unauthorized

**Error**: Authentication failures

**Solutions**:

```bash
# Check Kratos is running
curl http://localhost:4433/health/ready

# Verify session cookie
# Open DevTools > Application > Cookies

# Check CORS configuration
# Ensure FRONTEND_URL matches in .env

# Regenerate JWT secret
openssl rand -base64 32  # Use in JWT_SECRET
```

#### 7. Tests Failing

**Error**: Test timeouts or failures

**Solutions**:

```bash
# Clean test cache
go clean -testcache

# Run with verbose output
go test -v ./internal/workflow

# Check test database
TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/gorax_test

# Frontend tests
cd web
rm -rf node_modules
npm install
npm test
```

### Getting Help

If you're still stuck:

1. **Search Issues**: https://github.com/stherrien/gorax/issues
2. **Ask in Discussions**: https://github.com/stherrien/gorax/discussions
3. **Join Slack**: [Gorax Community Slack](https://gorax.slack.com)
4. **Email**: dev@gorax.dev

When asking for help, include:
- OS and version
- Go, Node, PostgreSQL, Redis versions
- Relevant error messages
- Steps to reproduce
- Your `.env` (without secrets!)

## Platform-Specific Instructions

### macOS

```bash
# Install dependencies via Homebrew
brew install go node postgresql@15 redis

# Start services
brew services start postgresql@15
brew services start redis

# Verify installations
go version
node --version
psql --version
redis-cli --version
```

### Linux (Ubuntu/Debian)

```bash
# Update packages
sudo apt-get update

# Install Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install Node.js
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Install PostgreSQL
sudo apt-get install -y postgresql-15 postgresql-contrib

# Install Redis
sudo apt-get install -y redis-server

# Start services
sudo systemctl start postgresql
sudo systemctl start redis-server
sudo systemctl enable postgresql
sudo systemctl enable redis-server
```

### Windows (WSL2 Recommended)

```powershell
# Install WSL2 (if not already installed)
wsl --install

# Inside WSL2, follow Linux instructions

# Or use Windows native tools:
# - Install Go from https://go.dev/dl/
# - Install Node.js from https://nodejs.org/
# - Install PostgreSQL from https://www.postgresql.org/download/windows/
# - Install Redis via Docker Desktop
```

## Next Steps

Now that you have Gorax running, here's what to do next:

### 1. Create Your First Workflow
📖 [First Workflow Tutorial](first-workflow.md)

Learn to:
- Create a workflow in the UI
- Add trigger and action nodes
- Test workflow execution
- View execution logs

### 2. Configure Integrations
🔧 [Slack Integration Guide](integrations/slack.md)

Set up:
- Webhook triggers
- HTTP actions
- Credential management
- Event filtering

### 3. Explore Examples
💡 [Example Workflows](../examples/)

Browse examples:
- CI/CD pipeline automation
- Incident response workflows
- Data sync workflows
- Notification workflows

### 4. Read the Developer Guide
📚 [Developer Guide](DEVELOPER_GUIDE.md)

Learn about:
- Architecture overview
- Testing patterns
- Code review process
- Deployment strategies

### 5. Join the Community
👥 [Community Resources](../README.md#community)

Connect with:
- GitHub Discussions
- Slack workspace
- Monthly community calls
- Contributor meetups

## Quick Reference

### Essential Commands

```bash
# Start everything (Docker)
docker-compose up -d

# Start API server
make run-api

# Start frontend
cd web && npm run dev

# Run tests
make test
cd web && npm test

# Database migrations
make migrate-up
make migrate-down

# Health checks
curl http://localhost:8080/health
curl http://localhost:5173

# Stop everything (Docker)
docker-compose down
```

### Key URLs

- **Frontend**: http://localhost:5173
- **API**: http://localhost:8080
- **API Docs**: http://localhost:8080/api/v1/docs
- **Health**: http://localhost:8080/health
- **Metrics**: http://localhost:8080/metrics

### Documentation Index

- [README](../README.md) - Project overview
- [CONTRIBUTING](../CONTRIBUTING.md) - How to contribute
- [DEVELOPER_GUIDE](DEVELOPER_GUIDE.md) - Comprehensive dev guide
- [API_REFERENCE](API_REFERENCE.md) - API documentation
- [SECURITY](SECURITY.md) - Security documentation

---

**Welcome to Gorax!** 🚀

Questions? Join our [Slack](https://gorax.slack.com) or open a [Discussion](https://github.com/stherrien/gorax/discussions).
