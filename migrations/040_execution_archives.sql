-- Execution archives migration
-- Creates the execution_archives table for storing archived execution data
-- This enables the retention service to archive executions before deletion

-- Create execution_archives table
CREATE TABLE IF NOT EXISTS execution_archives (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    workflow_id UUID,
    execution_data JSONB NOT NULL,
    archived_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    original_created_at TIMESTAMPTZ NOT NULL
);

-- Create indexes for efficient querying
CREATE INDEX idx_execution_archives_tenant_id ON execution_archives(tenant_id);
CREATE INDEX idx_execution_archives_workflow_id ON execution_archives(workflow_id);
CREATE INDEX idx_execution_archives_archived_at ON execution_archives(archived_at DESC);
CREATE INDEX idx_execution_archives_original_created_at ON execution_archives(original_created_at);

-- Enable RLS on execution_archives
ALTER TABLE execution_archives ENABLE ROW LEVEL SECURITY;

-- Create RLS policy for execution_archives
CREATE POLICY tenant_isolation_execution_archives ON execution_archives
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Add column comments for documentation
COMMENT ON TABLE execution_archives IS 'Stores archived workflow execution data for compliance and historical analysis';
COMMENT ON COLUMN execution_archives.id IS 'Original execution ID (same as the deleted execution)';
COMMENT ON COLUMN execution_archives.tenant_id IS 'Tenant that owned the execution';
COMMENT ON COLUMN execution_archives.workflow_id IS 'Workflow that was executed';
COMMENT ON COLUMN execution_archives.execution_data IS 'Complete execution data including step executions in JSONB format';
COMMENT ON COLUMN execution_archives.archived_at IS 'Timestamp when the execution was archived';
COMMENT ON COLUMN execution_archives.original_created_at IS 'Original creation timestamp of the execution';

-- Add executions_archived column to retention_cleanup_logs if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'retention_cleanup_logs'
        AND column_name = 'executions_archived'
    ) THEN
        ALTER TABLE retention_cleanup_logs ADD COLUMN executions_archived INTEGER NOT NULL DEFAULT 0;
        COMMENT ON COLUMN retention_cleanup_logs.executions_archived IS 'Number of executions archived during cleanup';
    END IF;
END $$;
