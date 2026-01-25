#!/bin/bash
# Post-deployment smoke test script
# Run this after deploying to verify the deployment was successful

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Default configuration
ENVIRONMENT="${ENVIRONMENT:-production}"
NOTIFICATION_WEBHOOK="${SLACK_WEBHOOK_URL:-}"

# Function to send notification
send_notification() {
    local status=$1
    local message=$2

    if [ -n "$NOTIFICATION_WEBHOOK" ]; then
        local emoji=""
        local color=""

        if [ "$status" = "success" ]; then
            emoji="✅"
            color="good"
        else
            emoji="❌"
            color="danger"
        fi

        curl -X POST "$NOTIFICATION_WEBHOOK" \
            -H 'Content-Type: application/json' \
            -d "{
                \"text\": \"${emoji} Gorax Post-Deployment Smoke Test\",
                \"attachments\": [{
                    \"color\": \"${color}\",
                    \"fields\": [
                        {\"title\": \"Environment\", \"value\": \"${ENVIRONMENT}\", \"short\": true},
                        {\"title\": \"Status\", \"value\": \"${status}\", \"short\": true},
                        {\"title\": \"Message\", \"value\": \"${message}\", \"short\": false}
                    ]
                }]
            }" >/dev/null 2>&1 || true
    fi
}

echo ""
echo "========================================"
echo "   🚀 Post-Deployment Smoke Test"
echo "========================================"
echo ""
echo "Environment: $ENVIRONMENT"
echo "Base URL: ${BASE_URL}"
echo ""

# Ensure required environment variables are set
if [ -z "$BASE_URL" ]; then
    echo -e "${RED}ERROR: BASE_URL environment variable is required${NC}"
    echo "Example: export BASE_URL=https://api.example.com"
    exit 1
fi

# Run smoke tests
echo "Running smoke tests..."
echo ""

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_DIR="$SCRIPT_DIR/tests/smoke"

# Skip Go tests for post-deployment (requires database access)
export SKIP_GO=true

# Skip database tests if no DATABASE_URL is provided
if [ -z "$DATABASE_URL" ]; then
    echo -e "${YELLOW}⚠ DATABASE_URL not set, skipping database tests${NC}"
    export SKIP_DB=true
fi

# Skip Redis tests if no REDIS_HOST is provided
if [ -z "$REDIS_HOST" ]; then
    echo -e "${YELLOW}⚠ REDIS_HOST not set, skipping some service tests${NC}"
fi

# Run tests
if cd "$TEST_DIR" && ./run-all.sh; then
    echo ""
    echo -e "${GREEN}✓ Post-deployment smoke tests PASSED${NC}"
    echo ""
    echo "Deployment verification successful!"

    send_notification "success" "All post-deployment smoke tests passed"
    exit 0
else
    echo ""
    echo -e "${RED}✗ Post-deployment smoke tests FAILED${NC}"
    echo ""
    echo "⚠️  WARNING: Deployment may have issues!"
    echo "Review logs and consider rolling back if necessary."

    send_notification "failed" "Post-deployment smoke tests failed - review deployment"
    exit 1
fi
