-- Fix for migration 023: Create immutable wrapper functions for DATE_TRUNC
-- This allows us to use these functions in functional indexes

-- Create immutable wrapper for hourly truncation
CREATE OR REPLACE FUNCTION date_trunc_hour_immutable(timestamp with time zone)
RETURNS timestamp with time zone
LANGUAGE sql
IMMUTABLE
PARALLEL SAFE
AS $$
    SELECT DATE_TRUNC('hour', $1);
$$;

-- Create immutable wrapper for daily truncation
CREATE OR REPLACE FUNCTION date_trunc_day_immutable(timestamp with time zone)
RETURNS timestamp with time zone
LANGUAGE sql
IMMUTABLE
PARALLEL SAFE
AS $$
    SELECT DATE_TRUNC('day', $1);
$$;

-- Now create the functional indexes using the immutable wrappers
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_executions_tenant_hour_trunc
ON executions(tenant_id, date_trunc_hour_immutable(created_at));

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_executions_tenant_day_trunc
ON executions(tenant_id, date_trunc_day_immutable(created_at));

-- Add comments
COMMENT ON FUNCTION date_trunc_hour_immutable IS 'Immutable wrapper for DATE_TRUNC(hour) to enable functional indexes';
COMMENT ON FUNCTION date_trunc_day_immutable IS 'Immutable wrapper for DATE_TRUNC(day) to enable functional indexes';
COMMENT ON INDEX idx_executions_tenant_hour_trunc IS 'Optimizes hourly execution trend queries';
COMMENT ON INDEX idx_executions_tenant_day_trunc IS 'Optimizes daily execution trend queries';
