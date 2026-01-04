#!/bin/bash
# Master Smoke Test Runner - Runs all smoke tests

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Configuration
SKIP_API="${SKIP_API:-false}"
SKIP_DB="${SKIP_DB:-false}"
SKIP_SERVICES="${SKIP_SERVICES:-false}"
SKIP_PERF="${SKIP_PERF:-false}"
SKIP_GO="${SKIP_GO:-false}"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Track results
TOTAL_SUITES=0
PASSED_SUITES=0
FAILED_SUITES=0
declare -a FAILED_SUITE_NAMES=()

echo ""
echo "========================================"
echo "   🔥 Gorax Smoke Test Suite"
echo "========================================"
echo ""
echo "Base URL: ${BASE_URL:-http://localhost:8080}"
echo "Database: ${DATABASE_URL:-postgres://postgres:postgres@localhost:5433/gorax}"
echo "Redis: ${REDIS_HOST:-localhost}:${REDIS_PORT:-6379}"
echo ""

# Function to run a test suite
run_suite() {
    local name="$1"
    local script="$2"
    local skip_var="$3"

    if [ "$skip_var" = "true" ]; then
        echo -e "${YELLOW}⊘${NC} Skipping: $name"
        return 0
    fi

    TOTAL_SUITES=$((TOTAL_SUITES + 1))

    echo ""
    if bash "$SCRIPT_DIR/$script"; then
        echo -e "${GREEN}✓${NC} $name: PASSED"
        PASSED_SUITES=$((PASSED_SUITES + 1))
        return 0
    else
        echo -e "${RED}✗${NC} $name: FAILED"
        FAILED_SUITES=$((FAILED_SUITES + 1))
        FAILED_SUITE_NAMES+=("$name")
        return 1
    fi
}

# Wait for services to be ready (optional)
if [ "${WAIT_FOR_SERVICES}" = "true" ]; then
    echo "Waiting for services to be ready..."

    # Wait for API
    max_attempts=30
    attempt=0
    while [ $attempt -lt $max_attempts ]; do
        if curl -s -f "${BASE_URL:-http://localhost:8080}/health" >/dev/null 2>&1; then
            echo -e "${GREEN}✓${NC} API is ready"
            break
        fi
        attempt=$((attempt + 1))
        if [ $attempt -lt $max_attempts ]; then
            sleep 2
        fi
    done

    if [ $attempt -eq $max_attempts ]; then
        echo -e "${RED}✗${NC} API failed to start within $((max_attempts * 2)) seconds"
        exit 1
    fi

    echo ""
fi

# Run test suites
run_suite "API Smoke Tests" "api-smoke.sh" "$SKIP_API"
run_suite "Database Smoke Tests" "db-smoke.sh" "$SKIP_DB"
run_suite "Service Dependency Tests" "service-smoke.sh" "$SKIP_SERVICES"
run_suite "Performance Tests" "perf-smoke.sh" "$SKIP_PERF"

# Run Go smoke tests if available
if [ "$SKIP_GO" != "true" ] && [ -f "$SCRIPT_DIR/go/workflow_smoke_test.go" ]; then
    TOTAL_SUITES=$((TOTAL_SUITES + 1))
    echo ""
    echo "========================================="
    echo "Go Workflow Smoke Tests"
    echo "========================================="

    if cd "$SCRIPT_DIR/go" && go test -v -tags=smoke -timeout=2m .; then
        echo -e "${GREEN}✓${NC} Go Workflow Tests: PASSED"
        PASSED_SUITES=$((PASSED_SUITES + 1))
    else
        echo -e "${RED}✗${NC} Go Workflow Tests: FAILED"
        FAILED_SUITES=$((FAILED_SUITES + 1))
        FAILED_SUITE_NAMES+=("Go Workflow Tests")
    fi
fi

# Print final summary
echo ""
echo "========================================="
echo "   📊 Final Summary"
echo "========================================="
echo "Total Suites:  $TOTAL_SUITES"
echo -e "Passed:        ${GREEN}$PASSED_SUITES${NC}"
echo -e "Failed:        ${RED}$FAILED_SUITES${NC}"
echo ""

# Print execution time
if [ -n "$START_TIME" ]; then
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))
    echo "Execution time: ${DURATION}s"
    echo ""
fi

# Print failed suites
if [ $FAILED_SUITES -gt 0 ]; then
    echo "Failed test suites:"
    for suite in "${FAILED_SUITE_NAMES[@]}"; do
        echo "  - $suite"
    done
    echo ""
    echo -e "${RED}✗ SMOKE TESTS FAILED${NC}"
    echo ""
    exit 1
else
    echo -e "${GREEN}✓ ALL SMOKE TESTS PASSED${NC}"
    echo ""
    echo "All critical paths are working correctly!"
    exit 0
fi
