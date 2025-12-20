-- Create human tasks table for approval workflows
CREATE TABLE IF NOT EXISTS human_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    step_id VARCHAR(100) NOT NULL,
    task_type VARCHAR(50) NOT NULL, -- 'approval', 'input', 'review'
    title VARCHAR(255) NOT NULL,
    description TEXT,
    assignees JSONB NOT NULL DEFAULT '[]'::JSONB, -- Array of user IDs or roles
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, approved, rejected, expired, cancelled
    due_date TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    completed_by UUID REFERENCES users(id),
    response_data JSONB,
    config JSONB NOT NULL DEFAULT '{}'::JSONB, -- Store task-specific configuration
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for efficient queries
CREATE INDEX idx_human_tasks_tenant ON human_tasks(tenant_id);
CREATE INDEX idx_human_tasks_execution ON human_tasks(execution_id);
CREATE INDEX idx_human_tasks_status ON human_tasks(status);
CREATE INDEX idx_human_tasks_assignee ON human_tasks USING GIN(assignees);
CREATE INDEX idx_human_tasks_due_date ON human_tasks(due_date) WHERE status = 'pending';
CREATE INDEX idx_human_tasks_created_at ON human_tasks(created_at DESC);

-- Add check constraints
ALTER TABLE human_tasks ADD CONSTRAINT chk_task_type
    CHECK (task_type IN ('approval', 'input', 'review'));

ALTER TABLE human_tasks ADD CONSTRAINT chk_status
    CHECK (status IN ('pending', 'approved', 'rejected', 'expired', 'cancelled'));

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_human_tasks_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_human_tasks_updated_at
    BEFORE UPDATE ON human_tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_human_tasks_updated_at();

-- Comments
COMMENT ON TABLE human_tasks IS 'Human approval tasks in workflow executions';
COMMENT ON COLUMN human_tasks.task_type IS 'Type of task: approval, input, or review';
COMMENT ON COLUMN human_tasks.assignees IS 'JSON array of user IDs or role names who can complete the task';
COMMENT ON COLUMN human_tasks.status IS 'Current status of the task';
COMMENT ON COLUMN human_tasks.response_data IS 'JSON data containing the user response (approval/rejection reason, form data, etc.)';
COMMENT ON COLUMN human_tasks.config IS 'Task configuration including form fields, escalation settings, etc.';
