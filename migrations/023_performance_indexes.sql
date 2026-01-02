-- Performance optimization indexes
-- This migration adds specialized indexes to improve query performance for common analytics and reporting queries

-- Index for execution trend queries with hourly granularity
-- Improves performance of time-series queries grouped by hour
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_executions_tenant_hour_trunc
ON executions(tenant_id, (DATE_TRUNC('hour', created_at)));

-- Index for execution trend queries with daily granularity
-- Improves performance of time-series queries grouped by day
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_executions_tenant_day_trunc
ON executions(tenant_id, (DATE_TRUNC('day', created_at)));

-- Covering index for workflow list queries
-- Includes commonly selected columns to avoid table lookups
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_workflows_tenant_status_updated
ON workflows(tenant_id, status, updated_at DESC)
INCLUDE (name, description);

-- Index for workflow execution status queries
-- Improves performance of status-based filtering and sorting
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_executions_workflow_status_created
ON executions(workflow_id, status, created_at DESC);

-- Index for execution filtering by trigger type
-- Useful for analytics queries filtering by how workflows were triggered
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_executions_tenant_trigger_created
ON executions(tenant_id, trigger_type, created_at DESC);

-- Add comments for documentation
COMMENT ON INDEX idx_executions_tenant_hour_trunc IS 'Optimizes hourly execution trend queries';
COMMENT ON INDEX idx_executions_tenant_day_trunc IS 'Optimizes daily execution trend queries';
COMMENT ON INDEX idx_workflows_tenant_status_updated IS 'Covering index for workflow list queries';
COMMENT ON INDEX idx_executions_workflow_status_created IS 'Optimizes workflow execution status queries';
COMMENT ON INDEX idx_executions_tenant_trigger_created IS 'Optimizes trigger type filtering in analytics';

-- Analyze tables to update statistics after index creation
ANALYZE executions;
ANALYZE workflows;
