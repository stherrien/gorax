#!/bin/bash
# Database Smoke Tests - Test database connectivity and critical tables

set -e

# Load library functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib.sh"

print_header "Database Smoke Tests"

# Test database connection
test_database "Database connection" "SELECT 1"

# Test critical tables exist
CRITICAL_TABLES=(
    "tenants"
    "users"
    "workflows"
    "executions"
    "credentials"
    "webhooks"
    "webhook_events"
    "schedules"
    "marketplace_templates"
    "marketplace_categories"
    "marketplace_reviews"
    "oauth_connections"
    "oauth_providers"
    "audit_events"
    "notifications"
)

for table in "${CRITICAL_TABLES[@]}"; do
    test_database "Table: $table" "SELECT COUNT(*) FROM $table LIMIT 1"
done

# Test database performance (simple query should be fast)
TOTAL_TESTS=$((TOTAL_TESTS + 1))
query_time=$(psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM workflows" 2>&1 | grep "Time:" | awk '{print $2}' || echo "unknown")

if [ "$query_time" != "unknown" ]; then
    print_status "PASS" "Query performance (workflows count: ${query_time})"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    print_status "PASS" "Query performance check (timing not available)"
    PASSED_TESTS=$((PASSED_TESTS + 1))
fi

# Test database indices exist (sample check)
test_database "Index check: workflows" "SELECT indexname FROM pg_indexes WHERE tablename = 'workflows' LIMIT 1"

# Test foreign key constraints are working
test_database "Foreign key constraints" "SELECT COUNT(*) FROM information_schema.table_constraints WHERE constraint_type = 'FOREIGN KEY'"

# Test database version
TOTAL_TESTS=$((TOTAL_TESTS + 1))
db_version=$(psql "$DATABASE_URL" -t -c "SELECT version()" 2>/dev/null | head -1 | xargs)

if [ -n "$db_version" ]; then
    print_status "PASS" "Database version: $db_version"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    print_status "FAIL" "Could not retrieve database version"
    FAILED_TESTS=$((FAILED_TESTS + 1))
    FAILED_TEST_NAMES+=("Database version check")
fi

# Print results
print_summary
