-- Rollback for 004_execution_history_enhancements.sql
-- Removes indexes and retention_until column

-- Drop indexes
DROP INDEX IF EXISTS idx_executions_retention_until;
DROP INDEX IF EXISTS idx_executions_tenant_created_at;
DROP INDEX IF EXISTS idx_executions_tenant_trigger_type;
DROP INDEX IF EXISTS idx_executions_tenant_workflow;
DROP INDEX IF EXISTS idx_executions_tenant_status;
DROP INDEX IF EXISTS idx_executions_cursor_pagination;

-- Drop retention_until column
ALTER TABLE executions DROP COLUMN IF EXISTS retention_until;
