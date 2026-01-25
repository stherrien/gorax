-- Add overlap policy and execution tracking to schedules table
-- Enables handling of overlapping executions with configurable policies

-- Add overlap policy column to schedules table
ALTER TABLE schedules
ADD COLUMN IF NOT EXISTS overlap_policy VARCHAR(20) NOT NULL DEFAULT 'skip';

-- Add constraint for valid overlap policy values
ALTER TABLE schedules
ADD CONSTRAINT valid_overlap_policy CHECK (overlap_policy IN ('skip', 'queue', 'terminate'));

-- Add column to track if schedule has running execution
ALTER TABLE schedules
ADD COLUMN IF NOT EXISTS running_execution_id UUID;

-- Create schedule_execution_logs table for tracking execution history
CREATE TABLE IF NOT EXISTS schedule_execution_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    schedule_id UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    execution_id UUID,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    trigger_time TIMESTAMPTZ NOT NULL,
    skipped_reason VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_execution_status CHECK (status IN ('pending', 'running', 'completed', 'failed', 'skipped', 'terminated'))
);

-- Indexes for schedule_execution_logs
CREATE INDEX IF NOT EXISTS idx_schedule_execution_logs_tenant_id
    ON schedule_execution_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_schedule_execution_logs_schedule_id
    ON schedule_execution_logs(schedule_id);
CREATE INDEX IF NOT EXISTS idx_schedule_execution_logs_execution_id
    ON schedule_execution_logs(execution_id) WHERE execution_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_schedule_execution_logs_status
    ON schedule_execution_logs(status);
CREATE INDEX IF NOT EXISTS idx_schedule_execution_logs_trigger_time
    ON schedule_execution_logs(trigger_time);
CREATE INDEX IF NOT EXISTS idx_schedule_execution_logs_created_at
    ON schedule_execution_logs(created_at);

-- Enable Row Level Security
ALTER TABLE schedule_execution_logs ENABLE ROW LEVEL SECURITY;

-- Create RLS policy for tenant isolation
CREATE POLICY tenant_isolation_schedule_execution_logs ON schedule_execution_logs
    USING (tenant_id = current_setting('app.current_tenant_id', true)::UUID);

-- Create trigger for updated_at
CREATE TRIGGER update_schedule_execution_logs_updated_at
    BEFORE UPDATE ON schedule_execution_logs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add index on schedules for running executions lookup
CREATE INDEX IF NOT EXISTS idx_schedules_running_execution
    ON schedules(running_execution_id) WHERE running_execution_id IS NOT NULL;
