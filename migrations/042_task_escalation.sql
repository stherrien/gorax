-- Add escalation tracking table for human task timeout escalation
CREATE TABLE IF NOT EXISTS task_escalations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES human_tasks(id) ON DELETE CASCADE,
    escalation_level INT NOT NULL DEFAULT 1,
    escalated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    escalated_from JSONB NOT NULL DEFAULT '[]'::JSONB, -- Previous assignees
    escalated_to JSONB NOT NULL DEFAULT '[]'::JSONB,   -- New assignees (backup approvers)
    escalation_reason VARCHAR(255) NOT NULL DEFAULT 'timeout', -- timeout, manual
    timeout_minutes INT, -- Original timeout that triggered this escalation
    auto_action_taken VARCHAR(50), -- auto_approve, auto_reject, or null if escalated to backup
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, superseded, completed
    completed_at TIMESTAMPTZ,
    completed_by UUID REFERENCES users(id),
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB, -- Additional context
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for efficient queries
CREATE INDEX idx_task_escalations_task ON task_escalations(task_id);
CREATE INDEX idx_task_escalations_status ON task_escalations(status);
CREATE INDEX idx_task_escalations_level ON task_escalations(task_id, escalation_level);
CREATE INDEX idx_task_escalations_created ON task_escalations(created_at DESC);

-- Add constraint for valid escalation reasons
ALTER TABLE task_escalations ADD CONSTRAINT chk_escalation_reason
    CHECK (escalation_reason IN ('timeout', 'manual'));

-- Add constraint for valid status values
ALTER TABLE task_escalations ADD CONSTRAINT chk_escalation_status
    CHECK (status IN ('active', 'superseded', 'completed'));

-- Add constraint for valid auto actions
ALTER TABLE task_escalations ADD CONSTRAINT chk_auto_action
    CHECK (auto_action_taken IS NULL OR auto_action_taken IN ('auto_approve', 'auto_reject'));

-- Trigger to update timestamps
CREATE OR REPLACE FUNCTION update_task_escalations_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    -- No updated_at column, but we track completed_at
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add columns to human_tasks for enhanced escalation configuration
-- These are stored in the config JSONB field, but we add columns for indexing/querying
ALTER TABLE human_tasks ADD COLUMN IF NOT EXISTS escalation_level INT NOT NULL DEFAULT 0;
ALTER TABLE human_tasks ADD COLUMN IF NOT EXISTS max_escalation_level INT NOT NULL DEFAULT 0;
ALTER TABLE human_tasks ADD COLUMN IF NOT EXISTS last_escalated_at TIMESTAMPTZ;

-- Index for finding tasks that need escalation processing
CREATE INDEX idx_human_tasks_escalation ON human_tasks(tenant_id, escalation_level, status)
    WHERE status = 'pending';

-- Comments
COMMENT ON TABLE task_escalations IS 'Tracks escalation history for human tasks that timeout';
COMMENT ON COLUMN task_escalations.escalation_level IS 'Current escalation level (1 = first escalation, 2 = second, etc.)';
COMMENT ON COLUMN task_escalations.escalated_from IS 'JSON array of previous assignees before escalation';
COMMENT ON COLUMN task_escalations.escalated_to IS 'JSON array of backup approvers who received the escalation';
COMMENT ON COLUMN task_escalations.auto_action_taken IS 'If final escalation, indicates auto-approve or auto-reject action taken';
COMMENT ON COLUMN task_escalations.status IS 'active = current escalation, superseded = replaced by higher level, completed = task resolved';
COMMENT ON COLUMN human_tasks.escalation_level IS 'Current escalation level for this task (0 = not escalated)';
COMMENT ON COLUMN human_tasks.max_escalation_level IS 'Maximum escalation level configured for this task';
COMMENT ON COLUMN human_tasks.last_escalated_at IS 'Timestamp of last escalation';
