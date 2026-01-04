#!/bin/bash
# Performance Smoke Tests - Test API response times

set -e

# Load library functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib.sh"

print_header "Performance Smoke Tests"

# Maximum acceptable response times (in milliseconds)
MAX_HEALTH_TIME=100
MAX_API_TIME=500
MAX_QUERY_TIME=1000

# Health endpoint should be very fast
test_endpoint_with_timing "Health endpoint response time" "GET" "/health" 200 "$MAX_HEALTH_TIME"

# Ready endpoint
test_endpoint_with_timing "Ready endpoint response time" "GET" "/ready" 200 "$MAX_HEALTH_TIME"

# API endpoints (will return 401 but should be fast)
test_endpoint_with_timing "Marketplace list response time" "GET" "/api/v1/marketplace/templates" 401 "$MAX_API_TIME"

test_endpoint_with_timing "OAuth providers response time" "GET" "/api/v1/oauth/providers" 401 "$MAX_API_TIME"

# Frontend static files
test_endpoint_with_timing "Frontend load time" "GET" "/" 200 "$MAX_API_TIME"

# Database query performance
if command -v psql >/dev/null 2>&1 && [ -n "$DATABASE_URL" ]; then
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # Measure query time
    start_time=$(date +%s%N)
    psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM workflows" >/dev/null 2>&1
    query_result=$?
    end_time=$(date +%s%N)

    if [ $query_result -eq 0 ]; then
        # Calculate duration in milliseconds
        duration_ms=$(( (end_time - start_time) / 1000000 ))

        if [ $duration_ms -lt $MAX_QUERY_TIME ]; then
            print_status "PASS" "Database query performance (${duration_ms}ms)"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            print_status "FAIL" "Database query performance (${duration_ms}ms > ${MAX_QUERY_TIME}ms)"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            FAILED_TEST_NAMES+=("Database query performance")
        fi
    else
        print_status "FAIL" "Database query performance (query failed)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("Database query performance")
    fi
else
    print_status "SKIP" "Database query performance (psql not available)"
fi

# Redis performance
if command -v redis-cli >/dev/null 2>&1; then
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    redis_host="${REDIS_HOST:-localhost}"
    redis_port="${REDIS_PORT:-6379}"

    # Measure Redis SET/GET time
    start_time=$(date +%s%N)
    redis-cli -h "$redis_host" -p "$redis_port" SET perf_test_key "test" >/dev/null 2>&1
    redis-cli -h "$redis_host" -p "$redis_port" GET perf_test_key >/dev/null 2>&1
    redis-cli -h "$redis_host" -p "$redis_port" DEL perf_test_key >/dev/null 2>&1
    redis_result=$?
    end_time=$(date +%s%N)

    if [ $redis_result -eq 0 ]; then
        duration_ms=$(( (end_time - start_time) / 1000000 ))

        if [ $duration_ms -lt 50 ]; then
            print_status "PASS" "Redis operations performance (${duration_ms}ms)"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            print_status "FAIL" "Redis operations performance (${duration_ms}ms > 50ms)"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            FAILED_TEST_NAMES+=("Redis operations performance")
        fi
    else
        print_status "FAIL" "Redis operations performance (operations failed)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("Redis operations performance")
    fi
else
    print_status "SKIP" "Redis performance (redis-cli not available)"
fi

# Print results
print_summary
