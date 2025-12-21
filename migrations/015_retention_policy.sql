-- Retention policy migration
-- Adds support for configurable retention policies and cleanup audit logs

-- Add retention settings to tenants table (if not already present in settings JSONB)
-- The settings JSONB field already exists in the tenants table from initial schema
-- We'll use it to store: retention_days and retention_enabled

-- Update existing tenants to have default retention settings if not set
UPDATE tenants
SET settings = jsonb_set(
    jsonb_set(
        COALESCE(settings, '{}'::jsonb),
        '{retention_days}',
        '90'::jsonb,
        true
    ),
    '{retention_enabled}',
    'true'::jsonb,
    true
)
WHERE settings->>'retention_days' IS NULL;

-- Create retention_cleanup_logs table for audit trail
CREATE TABLE IF NOT EXISTS retention_cleanup_logs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    executions_deleted INTEGER NOT NULL DEFAULT 0,
    step_executions_deleted INTEGER NOT NULL DEFAULT 0,
    retention_days INTEGER NOT NULL,
    cutoff_date TIMESTAMPTZ NOT NULL,
    duration_ms INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL, -- 'completed' or 'failed'
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for cleanup logs
CREATE INDEX idx_retention_cleanup_logs_tenant_id ON retention_cleanup_logs(tenant_id);
CREATE INDEX idx_retention_cleanup_logs_created_at ON retention_cleanup_logs(created_at DESC);
CREATE INDEX idx_retention_cleanup_logs_status ON retention_cleanup_logs(status);

-- Enable RLS on cleanup logs
ALTER TABLE retention_cleanup_logs ENABLE ROW LEVEL SECURITY;

-- Create RLS policy for cleanup logs
CREATE POLICY tenant_isolation_retention_cleanup_logs ON retention_cleanup_logs
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Add index on executions.created_at for efficient cleanup queries
-- This may already exist, so we use IF NOT EXISTS
CREATE INDEX IF NOT EXISTS idx_executions_tenant_created_at ON executions(tenant_id, created_at);

-- Add index on executions status for cleanup filtering
CREATE INDEX IF NOT EXISTS idx_executions_tenant_status_created ON executions(tenant_id, status, created_at);

-- Comment on retention settings
COMMENT ON TABLE retention_cleanup_logs IS 'Audit log for retention policy cleanup operations';
COMMENT ON COLUMN retention_cleanup_logs.executions_deleted IS 'Number of execution records deleted';
COMMENT ON COLUMN retention_cleanup_logs.step_executions_deleted IS 'Number of step_execution records deleted';
COMMENT ON COLUMN retention_cleanup_logs.retention_days IS 'Retention period in days applied during cleanup';
COMMENT ON COLUMN retention_cleanup_logs.cutoff_date IS 'Cutoff date used for deletion (executions older than this were deleted)';
COMMENT ON COLUMN retention_cleanup_logs.duration_ms IS 'Duration of cleanup operation in milliseconds';
