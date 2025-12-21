# Load Testing Quick Start Guide

Quick reference for running performance and load tests on the Gorax platform.

## Prerequisites

```bash
# Install k6 (macOS)
brew install k6

# Verify installation
k6 version
```

## Quick Test Commands

### Run All Tests (Load Scenario)
```bash
cd tests/load
./run_tests.sh
```

### Run Smoke Test (1 minute)
```bash
./run_tests.sh --scenario smoke
```

### Run Specific Test
```bash
# Workflow API only
./run_tests.sh workflow

# Execution API with stress scenario
./run_tests.sh execution --scenario stress

# Webhook ingestion test
./run_tests.sh webhook
```

### Test Against Different Environment
```bash
# Staging
./run_tests.sh --url https://staging.gorax.io --scenario smoke

# Local with custom port
./run_tests.sh --url http://localhost:3000
```

## Test Scenarios

| Scenario | Duration | VUs | Use Case |
|----------|----------|-----|----------|
| `smoke` | 1 min | 1 | Quick health check |
| `load` | 9 min | 10 | Normal traffic |
| `stress` | 26 min | 10→100 | Find breaking points |
| `spike` | 8 min | 10→100→10 | Traffic spikes |
| `soak` | 70 min | 20 | Memory leak detection |

## Go Benchmarks

```bash
# Benchmark workflow executor
go test -bench=. -benchmem ./internal/executor/

# Benchmark formula evaluator
go test -bench=. -benchmem ./internal/workflow/formula/

# Save results for comparison
go test -bench=. -benchmem ./internal/executor/ > before.txt
# ... make changes ...
go test -bench=. -benchmem ./internal/executor/ > after.txt
```

## Expected Performance

### Local Development
- Workflow Create: < 500ms
- Workflow Execute: < 1s
- Webhook Ingestion: < 100ms
- Auth Login: < 200ms

### Production
- Workflow Create: < 300ms
- Workflow Execute: < 500ms
- Webhook Ingestion: < 50ms
- Auth Login: < 100ms

## Results

Results are saved in `./results/` directory:
- `*_results.json` - Raw metrics
- `*_summary.txt` - Text summary
- `combined_report_*.html` - HTML report

```bash
# View results
cat results/*_summary.txt

# Open HTML report
open results/combined_report_*.html
```

## Environment Variables

```bash
# Custom configuration
export BASE_URL="http://localhost:8080"
export TEST_USER_EMAIL="loadtest@example.com"
export TEST_USER_PASSWORD="loadtest123"
export SCENARIO="load"

./run_tests.sh
```

## Common Issues

### k6 Not Found
```bash
# Install k6
brew install k6  # macOS
# See README.md for other platforms
```

### Connection Refused
```bash
# Start the Gorax server first
cd /Users/shawntherrien/Projects/gorax
make run

# Or use Docker
docker-compose up
```

### Authentication Failures
```bash
# Verify test user exists
# Create via API or admin interface
# Email: loadtest@example.com
# Password: loadtest123
```

## Need Help?

See full documentation: [tests/load/README.md](./README.md)
