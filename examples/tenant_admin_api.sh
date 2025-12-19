#!/bin/bash
# Tenant Admin API Examples for rflow
# This script demonstrates how to use the tenant admin endpoints

BASE_URL="http://localhost:8080/api/v1"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== rflow Tenant Admin API Examples ===${NC}\n"

# 1. Create a new tenant
echo -e "${GREEN}1. Creating a new tenant (Free tier)${NC}"
TENANT_RESPONSE=$(curl -s -X POST "$BASE_URL/admin/tenants" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Example Corp",
    "subdomain": "example",
    "tier": "free"
  }')

echo "$TENANT_RESPONSE" | jq .
TENANT_ID=$(echo "$TENANT_RESPONSE" | jq -r '.id')
echo -e "Created tenant ID: ${TENANT_ID}\n"

# 2. List all tenants
echo -e "${GREEN}2. Listing all tenants (with pagination)${NC}"
curl -s "$BASE_URL/admin/tenants?limit=10&offset=0" | jq .
echo ""

# 3. Get specific tenant details
echo -e "${GREEN}3. Getting tenant details${NC}"
curl -s "$BASE_URL/admin/tenants/$TENANT_ID" | jq .
echo ""

# 4. Update tenant (upgrade to professional)
echo -e "${GREEN}4. Updating tenant tier to Professional${NC}"
curl -s -X PUT "$BASE_URL/admin/tenants/$TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "tier": "professional"
  }' | jq .
echo ""

# 5. Update tenant quotas (custom limits)
echo -e "${GREEN}5. Updating tenant quotas${NC}"
curl -s -X PUT "$BASE_URL/admin/tenants/$TENANT_ID/quotas" \
  -H "Content-Type: application/json" \
  -d '{
    "max_workflows": 75,
    "max_executions_per_day": 7500,
    "max_concurrent_executions": 15,
    "max_storage_bytes": 10737418240,
    "max_api_calls_per_minute": 400,
    "execution_history_retention_days": 60
  }' | jq .
echo ""

# 6. Get tenant usage statistics
echo -e "${GREEN}6. Getting tenant usage statistics${NC}"
curl -s "$BASE_URL/admin/tenants/$TENANT_ID/usage" | jq .
echo ""

# 7. Create workflows to test quotas (requires tenant context)
echo -e "${GREEN}7. Testing quota enforcement (simulated)${NC}"
echo "Note: This would require setting X-Tenant-ID header and creating workflows"
echo "The quota middleware will block creation once limits are reached"
echo ""

# 8. Deactivate tenant (soft delete)
echo -e "${GREEN}8. Deactivating tenant${NC}"
read -p "Press enter to deactivate tenant (or Ctrl+C to skip)..."
curl -s -X DELETE "$BASE_URL/admin/tenants/$TENANT_ID"
echo -e "${RED}Tenant deactivated${NC}\n"

echo -e "${BLUE}=== Examples Complete ===${NC}"
echo ""
echo "Tips:"
echo "  - Use jq to pretty-print JSON responses"
echo "  - Set X-Tenant-ID header for tenant-scoped operations"
echo "  - Monitor usage with GET /admin/tenants/{id}/usage"
echo "  - Quotas are enforced automatically via middleware"
