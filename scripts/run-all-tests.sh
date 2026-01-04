#!/bin/bash

# Script to run all integration and E2E tests locally
# Usage: ./scripts/run-all-tests.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Gorax Comprehensive Test Suite${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Step 1: Check prerequisites
echo -e "${YELLOW}[1/7] Checking prerequisites...${NC}"
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}Error: Docker Compose is not installed${NC}"
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Prerequisites check passed${NC}"
echo ""

# Step 2: Start test dependencies
echo -e "${YELLOW}[2/7] Starting test dependencies (Docker)...${NC}"
cd tests
docker-compose -f docker-compose.test.yml up -d

# Wait for services to be healthy
echo "Waiting for services to be ready..."
sleep 10

# Check service health
for service in postgres-test redis-test mysql-test mongo-test; do
    if ! docker-compose -f docker-compose.test.yml ps | grep "$service" | grep -q "Up (healthy)"; then
        echo -e "${YELLOW}Warning: $service may not be fully ready${NC}"
    fi
done

cd ..
echo -e "${GREEN}✓ Test dependencies started${NC}"
echo ""

# Step 3: Run database migrations
echo -e "${YELLOW}[3/7] Running database migrations...${NC}"
export DATABASE_URL="postgres://gorax:gorax_test@localhost:5433/gorax_test?sslmode=disable"
go run cmd/migrate/main.go up || echo -e "${YELLOW}Warning: Migration may have already been applied${NC}"
echo -e "${GREEN}✓ Migrations completed${NC}"
echo ""

# Step 4: Run integration tests
echo -e "${YELLOW}[4/7] Running integration tests...${NC}"
export DATABASE_URL="postgres://gorax:gorax_test@localhost:5433/gorax_test?sslmode=disable"
export REDIS_URL="localhost:6380"
export ENV="test"

go test -v -tags=integration -race -coverprofile=integration-coverage.out ./tests/integration/... || {
    echo -e "${RED}Integration tests failed${NC}"
    exit 1
}

echo -e "${GREEN}✓ Integration tests passed${NC}"
echo ""

# Step 5: Generate integration coverage report
echo -e "${YELLOW}[5/7] Generating integration coverage report...${NC}"
go tool cover -html=integration-coverage.out -o integration-coverage.html
echo -e "${GREEN}✓ Coverage report generated: integration-coverage.html${NC}"
echo ""

# Step 6: Check if E2E tests should run
read -p "Run E2E tests? This will start the backend server (y/n): " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}[6/7] Installing Playwright and dependencies...${NC}"
    cd web
    npm ci
    npx playwright install --with-deps chromium || {
        echo -e "${YELLOW}Warning: Could not install Playwright browsers${NC}"
    }

    echo -e "${YELLOW}[6/7] Starting backend server...${NC}"
    cd ..
    go run cmd/api/main.go &
    BACKEND_PID=$!
    echo "Backend PID: $BACKEND_PID"

    # Wait for backend to be ready
    echo "Waiting for backend to start..."
    timeout 30 sh -c 'until curl -f http://localhost:8080/health; do sleep 1; done' || {
        echo -e "${RED}Backend failed to start${NC}"
        kill $BACKEND_PID 2>/dev/null || true
        exit 1
    }

    echo -e "${GREEN}✓ Backend server started${NC}"
    echo ""

    echo -e "${YELLOW}[7/7] Running E2E tests...${NC}"
    cd web
    export BASE_URL="http://localhost:8080"
    npm run test:e2e || {
        echo -e "${RED}E2E tests failed${NC}"
        kill $BACKEND_PID 2>/dev/null || true
        exit 1
    }

    # Stop backend server
    echo "Stopping backend server..."
    kill $BACKEND_PID 2>/dev/null || true

    echo -e "${GREEN}✓ E2E tests passed${NC}"
    echo ""
else
    echo -e "${YELLOW}Skipping E2E tests${NC}"
    echo ""
fi

# Summary
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Test Summary${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${GREEN}✓ Integration tests: PASSED${NC}"
echo -e "  Coverage report: integration-coverage.html"
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${GREEN}✓ E2E tests: PASSED${NC}"
    echo -e "  Test results: web/test-results/"
    echo ""
fi

echo -e "${YELLOW}To view results:${NC}"
echo "  Integration coverage: open integration-coverage.html"
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "  E2E test report: cd web && npx playwright show-report"
fi
echo ""

echo -e "${YELLOW}To stop test services:${NC}"
echo "  cd tests && docker-compose -f docker-compose.test.yml down"
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  All tests completed successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
