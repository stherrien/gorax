# Smoke Tests - Quick Start

## TL;DR

```bash
# Run all smoke tests
make smoke-tests

# Expected: ✓ ALL SMOKE TESTS PASSED
```

## 30-Second Setup

```bash
# 1. Start services
make dev-simple

# 2. Start API
make run-api-dev

# 3. Run tests (in new terminal)
make smoke-tests
```

## Commands

| Command | Description |
|---------|-------------|
| `make smoke-tests` | Run all smoke tests (2 min) |
| `make smoke-tests-quick` | Skip Go tests (30 sec) |
| `make smoke-tests-api` | API only (10 sec) |
| `make smoke-tests-db` | Database only (15 sec) |
| `make smoke-tests-services` | Services only (10 sec) |
| `make smoke-tests-perf` | Performance only (20 sec) |

## What's Tested?

- ✅ API health and endpoints
- ✅ Database connectivity
- ✅ Redis connection
- ✅ Response times
- ✅ Workflow execution

## Configuration

```bash
# Set these if needed
export BASE_URL=http://localhost:8080
export DATABASE_URL=postgres://postgres:postgres@localhost:5433/gorax?sslmode=disable
export REDIS_HOST=localhost
export REDIS_PORT=6379
```

## Skip Tests

```bash
export SKIP_API=true        # Skip API tests
export SKIP_DB=true         # Skip database tests
export SKIP_SERVICES=true   # Skip service tests
export SKIP_PERF=true       # Skip performance tests
export SKIP_GO=true         # Skip Go tests
```

## Troubleshooting

### Connection Refused
```bash
# Check services are running
docker ps
curl http://localhost:8080/health
```

### Database Error
```bash
# Test database connection
psql postgres://postgres:postgres@localhost:5433/gorax -c "SELECT 1"
```

### Redis Error
```bash
# Test Redis connection
redis-cli -h localhost -p 6379 ping
```

## CI/CD

Tests run automatically on:
- Every PR to main/dev
- Every push to main/dev
- View in GitHub Actions → "Smoke Tests"

## Full Documentation

See: `docs/SMOKE_TESTS.md`
