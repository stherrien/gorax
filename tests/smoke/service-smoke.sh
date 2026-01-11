#!/bin/bash
# Service Smoke Tests - Test external service dependencies

set -e

# Load library functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib.sh"

print_header "Service Dependency Smoke Tests"

# Redis connectivity
test_redis "Redis connection"

# Test Redis basic operations
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if command -v redis-cli >/dev/null 2>&1; then
    redis_host="${REDIS_HOST:-localhost}"
    redis_port="${REDIS_PORT:-6379}"

    # Set and get a test key
    redis-cli -h "$redis_host" -p "$redis_port" SET smoke_test_key "test_value" >/dev/null 2>&1
    value=$(redis-cli -h "$redis_host" -p "$redis_port" GET smoke_test_key 2>/dev/null | tr -d '\r\n')
    redis-cli -h "$redis_host" -p "$redis_port" DEL smoke_test_key >/dev/null 2>&1

    if [ "$value" = "test_value" ]; then
        print_status "PASS" "Redis read/write operations"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        print_status "FAIL" "Redis read/write operations"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("Redis read/write operations")
    fi
else
    print_status "SKIP" "Redis read/write operations (redis-cli not available)"
fi

# Docker containers (if running in Docker environment)
if command -v docker >/dev/null 2>&1; then
    print_status "INFO" "Checking Docker containers..."

    # Check if we're using Docker Compose
    if docker compose ps >/dev/null 2>&1 || docker-compose ps >/dev/null 2>&1; then
        # Common container names
        CONTAINERS=(
            "gorax-postgres"
            "gorax-redis"
        )

        for container in "${CONTAINERS[@]}"; do
            test_container "Container: $container" "$container" || true
        done
    else
        print_status "SKIP" "Docker Compose check (not in Docker environment)"
    fi
else
    print_status "SKIP" "Docker checks (Docker not available)"
fi

# AWS LocalStack (if configured for testing)
if [ -n "${AWS_ENDPOINT}" ] && [[ "${AWS_ENDPOINT}" == *"localstack"* ]]; then
    print_status "INFO" "Checking LocalStack services..."

    # Test S3
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if aws --endpoint-url="$AWS_ENDPOINT" s3 ls >/dev/null 2>&1; then
        print_status "PASS" "LocalStack S3"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        print_status "FAIL" "LocalStack S3"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("LocalStack S3")
    fi

    # Test SQS
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if aws --endpoint-url="$AWS_ENDPOINT" sqs list-queues >/dev/null 2>&1; then
        print_status "PASS" "LocalStack SQS"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        print_status "FAIL" "LocalStack SQS"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("LocalStack SQS")
    fi
else
    print_status "SKIP" "LocalStack checks (not configured)"
fi

# Kratos (if configured)
if [ -n "${KRATOS_PUBLIC_URL}" ]; then
    KRATOS_URL="${KRATOS_PUBLIC_URL}/health/ready"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if curl -s -f "$KRATOS_URL" >/dev/null 2>&1; then
        print_status "PASS" "Kratos health check"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        print_status "FAIL" "Kratos health check"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("Kratos health check")
    fi
else
    print_status "SKIP" "Kratos health check (not configured)"
fi

# Print results
print_summary
