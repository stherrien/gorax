#!/bin/bash

# Database Migration Runner Script
# Runs pending migrations for enterprise features

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5433}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-gorax}"

# PostgreSQL connection string
PGPASSWORD="${DB_PASSWORD}"
export PGPASSWORD

PSQL_CMD="psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d ${DB_NAME}"

echo -e "${GREEN}=== Gorax Database Migration Runner ===${NC}"
echo "Database: ${DB_NAME}@${DB_HOST}:${DB_PORT}"
echo ""

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo -e "${RED}Error: psql command not found. Please install PostgreSQL client.${NC}"
    exit 1
fi

# Test database connection
echo "Testing database connection..."
if ! $PSQL_CMD -c "SELECT 1" > /dev/null 2>&1; then
    echo -e "${RED}Error: Cannot connect to database. Please check your configuration.${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Database connection successful${NC}"
echo ""

# Create migrations tracking table if it doesn't exist
echo "Creating migrations tracking table..."
$PSQL_CMD -c "
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    description TEXT
);
" > /dev/null 2>&1
echo -e "${GREEN}✓ Migrations tracking table ready${NC}"
echo ""

# List of migrations to run (in order)
# Note: Only new enterprise feature migrations
MIGRATIONS=(
    "025_error_handling.sql"
    "026_marketplace_enhancements.sql"
    "027_template_reviews_enhanced.sql"
    "028_oauth_connections.sql"
    "029_database_connectors.sql"
    "030_sso_providers.sql"
    "031_audit_logs.sql"
)

# Track statistics
TOTAL_MIGRATIONS=${#MIGRATIONS[@]}
APPLIED_COUNT=0
SKIPPED_COUNT=0
FAILED_COUNT=0

echo "Found ${TOTAL_MIGRATIONS} migrations to process"
echo ""

# Process each migration
for migration in "${MIGRATIONS[@]}"; do
    MIGRATION_FILE="migrations/${migration}"
    MIGRATION_NAME="${migration%.sql}"

    echo -e "Processing: ${YELLOW}${migration}${NC}"

    # Check if migration file exists
    if [ ! -f "${MIGRATION_FILE}" ]; then
        echo -e "${RED}  ✗ Migration file not found: ${MIGRATION_FILE}${NC}"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        continue
    fi

    # Check if migration has already been applied
    APPLIED=$($PSQL_CMD -t -c "SELECT COUNT(*) FROM schema_migrations WHERE version = '${MIGRATION_NAME}'" 2>/dev/null | tr -d ' ')

    if [ "${APPLIED}" = "1" ]; then
        echo -e "${YELLOW}  ⊘ Already applied, skipping${NC}"
        SKIPPED_COUNT=$((SKIPPED_COUNT + 1))
        continue
    fi

    # Run the migration in a transaction
    echo "  → Applying migration..."
    if $PSQL_CMD -v ON_ERROR_STOP=1 -f "${MIGRATION_FILE}" > /dev/null 2>&1; then
        # Record successful migration
        $PSQL_CMD -c "INSERT INTO schema_migrations (version, description) VALUES ('${MIGRATION_NAME}', 'Enterprise feature migration')" > /dev/null 2>&1
        echo -e "${GREEN}  ✓ Successfully applied${NC}"
        APPLIED_COUNT=$((APPLIED_COUNT + 1))
    else
        echo -e "${RED}  ✗ Failed to apply migration${NC}"
        echo -e "${RED}     Check the migration file for errors${NC}"
        FAILED_COUNT=$((FAILED_COUNT + 1))

        # Show error details
        echo "  → Running with verbose output for debugging..."
        $PSQL_CMD -f "${MIGRATION_FILE}"
        exit 1
    fi

    echo ""
done

# Print summary
echo -e "${GREEN}=== Migration Summary ===${NC}"
echo "Total migrations: ${TOTAL_MIGRATIONS}"
echo -e "Applied: ${GREEN}${APPLIED_COUNT}${NC}"
echo -e "Skipped: ${YELLOW}${SKIPPED_COUNT}${NC}"
if [ ${FAILED_COUNT} -gt 0 ]; then
    echo -e "Failed: ${RED}${FAILED_COUNT}${NC}"
    exit 1
else
    echo -e "Failed: ${FAILED_COUNT}"
fi

echo ""
if [ ${APPLIED_COUNT} -gt 0 ]; then
    echo -e "${GREEN}✓ All migrations completed successfully!${NC}"
else
    echo -e "${YELLOW}ⓘ No new migrations to apply${NC}"
fi

# Show current schema version
echo ""
echo "Current schema migrations:"
$PSQL_CMD -c "SELECT version, applied_at FROM schema_migrations ORDER BY applied_at DESC LIMIT 10"

exit 0
