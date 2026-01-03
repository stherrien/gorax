#!/bin/bash
# API Smoke Tests - Test critical API endpoints

set -e

# Load library functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib.sh"

print_header "API Smoke Tests"

# Health check endpoints
test_endpoint "Health check" "GET" "/health" 200
test_endpoint "Ready check" "GET" "/ready" 200

# Marketplace endpoints (public)
test_endpoint "List marketplace templates" "GET" "/api/v1/marketplace/templates" 401
test_endpoint "List categories" "GET" "/api/v1/marketplace/categories" 401

# OAuth endpoints (public listing)
test_endpoint "List OAuth providers" "GET" "/api/v1/oauth/providers" 401

# Metrics endpoint (if enabled)
if curl -s -f "$BASE_URL/metrics" >/dev/null 2>&1; then
    test_endpoint "Prometheus metrics" "GET" "/metrics" 200
else
    print_status "SKIP" "Prometheus metrics (not enabled)"
fi

# GraphQL endpoint (should require auth)
test_endpoint "GraphQL endpoint" "POST" "/api/v1/graphql" 401

# Static files / Frontend
test_endpoint "Frontend app root" "GET" "/" 200

# API versioning check
test_endpoint "API v1 base (should redirect or require auth)" "GET" "/api/v1" 404

# Print results
print_summary
