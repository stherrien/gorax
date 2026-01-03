#!/bin/bash
# Smoke test library - shared functions for smoke tests

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Base URL for API tests
BASE_URL="${BASE_URL:-http://localhost:8080}"
TENANT_ID="${TEST_TENANT_ID:-test-tenant-1}"

# Database connection
DATABASE_URL="${DATABASE_URL:-postgres://postgres:postgres@localhost:5433/gorax?sslmode=disable}"

# Test result tracking
declare -a FAILED_TEST_NAMES=()

# Print colored output
print_status() {
    local status=$1
    local message=$2

    case "$status" in
        "PASS")
            echo -e "${GREEN}✓${NC} $message"
            ;;
        "FAIL")
            echo -e "${RED}✗${NC} $message"
            ;;
        "SKIP")
            echo -e "${YELLOW}⊘${NC} $message"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ${NC} $message"
            ;;
        *)
            echo "$message"
            ;;
    esac
}

# Print section header
print_header() {
    echo ""
    echo "========================================="
    echo "$1"
    echo "========================================="
}

# Test an HTTP endpoint
test_endpoint() {
    local name="$1"
    local method="$2"
    local url="$3"
    local expected_status="$4"
    local headers="${5:-}"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # Build curl command
    local curl_cmd="curl -s -o /dev/null -w '%{http_code}' -X $method"

    # Add headers if provided
    if [ -n "$headers" ]; then
        curl_cmd="$curl_cmd $headers"
    fi

    curl_cmd="$curl_cmd $BASE_URL$url"

    # Execute request
    local status
    status=$(eval "$curl_cmd" 2>/dev/null)
    local curl_exit=$?

    # Check if curl succeeded
    if [ $curl_exit -ne 0 ]; then
        print_status "FAIL" "$name (connection failed)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("$name")
        return 1
    fi

    # Check status code
    if [ "$status" -eq "$expected_status" ]; then
        print_status "PASS" "$name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        print_status "FAIL" "$name (got $status, expected $expected_status)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("$name")
        return 1
    fi
}

# Test endpoint with response time check
test_endpoint_with_timing() {
    local name="$1"
    local method="$2"
    local url="$3"
    local expected_status="$4"
    local max_time_ms="$5"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # Execute request and capture timing
    local response
    response=$(curl -s -w '\n%{http_code}\n%{time_total}' -X "$method" "$BASE_URL$url" 2>/dev/null)
    local curl_exit=$?

    if [ $curl_exit -ne 0 ]; then
        print_status "FAIL" "$name (connection failed)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("$name")
        return 1
    fi

    # Parse response (last two lines are status and time)
    local status time_total
    status=$(echo "$response" | tail -2 | head -1)
    time_total=$(echo "$response" | tail -1)

    # Convert time to milliseconds
    local time_ms
    time_ms=$(echo "$time_total * 1000" | bc -l | cut -d. -f1)

    # Check status code
    if [ "$status" -ne "$expected_status" ]; then
        print_status "FAIL" "$name (got $status, expected $expected_status)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("$name")
        return 1
    fi

    # Check timing
    if [ "$time_ms" -gt "$max_time_ms" ]; then
        print_status "FAIL" "$name (${time_ms}ms > ${max_time_ms}ms)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("$name")
        return 1
    fi

    print_status "PASS" "$name (${time_ms}ms)"
    PASSED_TESTS=$((PASSED_TESTS + 1))
    return 0
}

# Test database connectivity
test_database() {
    local name="$1"
    local query="$2"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # Test query
    if psql "$DATABASE_URL" -c "$query" >/dev/null 2>&1; then
        print_status "PASS" "$name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        print_status "FAIL" "$name"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("$name")
        return 1
    fi
}

# Test Redis connectivity
test_redis() {
    local name="$1"
    local redis_host="${REDIS_HOST:-localhost}"
    local redis_port="${REDIS_PORT:-6379}"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if command -v redis-cli >/dev/null 2>&1; then
        if redis-cli -h "$redis_host" -p "$redis_port" ping >/dev/null 2>&1; then
            print_status "PASS" "$name"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            return 0
        fi
    fi

    print_status "FAIL" "$name"
    FAILED_TESTS=$((FAILED_TESTS + 1))
    FAILED_TEST_NAMES+=("$name")
    return 1
}

# Test Docker container
test_container() {
    local name="$1"
    local container="$2"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
        print_status "PASS" "$name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        print_status "FAIL" "$name (container not running)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("$name")
        return 1
    fi
}

# Wait for service to be ready
wait_for_service() {
    local name="$1"
    local url="$2"
    local max_attempts="${3:-30}"
    local wait_seconds="${4:-2}"

    print_status "INFO" "Waiting for $name to be ready..."

    for i in $(seq 1 "$max_attempts"); do
        if curl -s -f "$url" >/dev/null 2>&1; then
            print_status "PASS" "$name is ready"
            return 0
        fi

        if [ "$i" -lt "$max_attempts" ]; then
            sleep "$wait_seconds"
        fi
    done

    print_status "FAIL" "$name failed to start within $((max_attempts * wait_seconds)) seconds"
    return 1
}

# Print test summary
print_summary() {
    echo ""
    echo "========================================="
    echo "Test Summary"
    echo "========================================="
    echo "Total:  $TOTAL_TESTS"
    echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
    echo ""

    if [ $FAILED_TESTS -gt 0 ]; then
        echo "Failed tests:"
        for test_name in "${FAILED_TEST_NAMES[@]}"; do
            echo "  - $test_name"
        done
        echo ""
        echo -e "${RED}✗ SMOKE TESTS FAILED${NC}"
        return 1
    else
        echo -e "${GREEN}✓ ALL SMOKE TESTS PASSED${NC}"
        return 0
    fi
}

# Reset counters (useful for multiple test runs)
reset_counters() {
    TOTAL_TESTS=0
    PASSED_TESTS=0
    FAILED_TESTS=0
    FAILED_TEST_NAMES=()
}
