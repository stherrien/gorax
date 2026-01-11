#!/bin/bash
# Wait for services to be ready before running tests

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Configuration
API_URL="${BASE_URL:-http://localhost:8080}"
DB_URL="${DATABASE_URL:-postgres://postgres:postgres@localhost:5433/gorax?sslmode=disable}"
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"

MAX_ATTEMPTS=30
WAIT_SECONDS=2

echo "Waiting for services to be ready..."
echo ""

# Wait for API
echo -n "API ($API_URL): "
attempt=0
while [ $attempt -lt $MAX_ATTEMPTS ]; do
    if curl -s -f "$API_URL/health" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ ready${NC}"
        break
    fi
    attempt=$((attempt + 1))
    if [ $attempt -eq $MAX_ATTEMPTS ]; then
        echo -e "${RED}✗ failed${NC}"
        exit 1
    fi
    sleep $WAIT_SECONDS
done

# Wait for Database
echo -n "Database: "
attempt=0
while [ $attempt -lt $MAX_ATTEMPTS ]; do
    if psql "$DB_URL" -c "SELECT 1" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ ready${NC}"
        break
    fi
    attempt=$((attempt + 1))
    if [ $attempt -eq $MAX_ATTEMPTS ]; then
        echo -e "${RED}✗ failed${NC}"
        exit 1
    fi
    sleep $WAIT_SECONDS
done

# Wait for Redis
if command -v redis-cli >/dev/null 2>&1; then
    echo -n "Redis: "
    attempt=0
    while [ $attempt -lt $MAX_ATTEMPTS ]; do
        if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping >/dev/null 2>&1; then
            echo -e "${GREEN}✓ ready${NC}"
            break
        fi
        attempt=$((attempt + 1))
        if [ $attempt -eq $MAX_ATTEMPTS ]; then
            echo -e "${RED}✗ failed${NC}"
            exit 1
        fi
        sleep $WAIT_SECONDS
    done
else
    echo -e "Redis: ${YELLOW}⊘ skipped (redis-cli not available)${NC}"
fi

echo ""
echo -e "${GREEN}All services are ready!${NC}"
exit 0
