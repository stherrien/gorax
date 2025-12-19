-- Migration: Add indexes for execution filtering performance
-- Description: Adds indexes to support advanced execution filtering with error search, duration, and ID prefix

-- Add GIN index for error message full-text search
-- This enables fast ILIKE queries on error_message field
CREATE INDEX IF NOT EXISTS idx_executions_error_message_gin
ON executions
USING gin (error_message gin_trgm_ops)
WHERE error_message IS NOT NULL;

-- Add index for execution ID prefix searches
-- This supports fast LIKE 'prefix%' queries
CREATE INDEX IF NOT EXISTS idx_executions_id_prefix
ON executions (id text_pattern_ops);

-- Add composite index for duration-based filtering
-- This supports queries filtering by execution duration
CREATE INDEX IF NOT EXISTS idx_executions_duration
ON executions ((EXTRACT(EPOCH FROM (completed_at - started_at)) * 1000))
WHERE started_at IS NOT NULL AND completed_at IS NOT NULL;

-- Add composite indexes for common filter combinations
-- Status + created_at (most common combination)
CREATE INDEX IF NOT EXISTS idx_executions_status_created_at
ON executions (tenant_id, status, created_at DESC);

-- Trigger type + created_at
CREATE INDEX IF NOT EXISTS idx_executions_trigger_type_created_at
ON executions (tenant_id, trigger_type, created_at DESC);

-- Workflow ID + status + created_at
CREATE INDEX IF NOT EXISTS idx_executions_workflow_status_created_at
ON executions (tenant_id, workflow_id, status, created_at DESC);

-- Enable pg_trgm extension if not already enabled (required for GIN index)
-- Note: This requires superuser privileges
-- CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Rollback instructions:
-- DROP INDEX IF EXISTS idx_executions_error_message_gin;
-- DROP INDEX IF EXISTS idx_executions_id_prefix;
-- DROP INDEX IF EXISTS idx_executions_duration;
-- DROP INDEX IF EXISTS idx_executions_status_created_at;
-- DROP INDEX IF EXISTS idx_executions_trigger_type_created_at;
-- DROP INDEX IF EXISTS idx_executions_workflow_status_created_at;
