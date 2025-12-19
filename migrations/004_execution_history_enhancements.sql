-- Execution history enhancements for Phase 2.4
-- Adds indexes for cursor-based pagination and filtering
-- Adds retention_until column for execution lifecycle management

-- Add retention_until column to executions table
ALTER TABLE executions ADD COLUMN retention_until TIMESTAMPTZ;

-- Composite index for cursor-based pagination (tenant_id, created_at DESC, id)
-- This index supports efficient cursor-based pagination queries
CREATE INDEX idx_executions_cursor_pagination ON executions(tenant_id, created_at DESC, id);

-- Index for filtering by status within tenant
CREATE INDEX idx_executions_tenant_status ON executions(tenant_id, status, created_at DESC);

-- Index for filtering by workflow_id within tenant
CREATE INDEX idx_executions_tenant_workflow ON executions(tenant_id, workflow_id, created_at DESC);

-- Index for filtering by trigger_type within tenant
CREATE INDEX idx_executions_tenant_trigger_type ON executions(tenant_id, trigger_type, created_at DESC);

-- Index for date range queries within tenant
CREATE INDEX idx_executions_tenant_created_at ON executions(tenant_id, created_at DESC);

-- Index for retention policy queries
-- Used to efficiently find executions that should be deleted
CREATE INDEX idx_executions_retention_until ON executions(retention_until) WHERE retention_until IS NOT NULL;

-- Comment on the new column
COMMENT ON COLUMN executions.retention_until IS 'Timestamp after which the execution can be deleted by retention policy';
