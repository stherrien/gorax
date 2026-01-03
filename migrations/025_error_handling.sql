-- Error handling enhancements for Gorax workflows
-- Adds support for try/catch/finally, retry, and circuit breaker nodes

-- Add error handling configuration to execution_steps
-- This tracks retry attempts and error context for each step
ALTER TABLE execution_steps
ADD COLUMN IF NOT EXISTS retry_count INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS error_classification VARCHAR(50),
ADD COLUMN IF NOT EXISTS error_context JSONB;

-- Create error_handling_history table to track error details
-- This stores detailed error information for debugging and analysis
CREATE TABLE IF NOT EXISTS error_handling_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    step_execution_id UUID REFERENCES execution_steps(id) ON DELETE CASCADE,
    node_id VARCHAR(255) NOT NULL,
    node_type VARCHAR(100) NOT NULL,
    error_type VARCHAR(100) NOT NULL,
    error_message TEXT NOT NULL,
    error_classification VARCHAR(50) NOT NULL,
    retry_attempt INTEGER DEFAULT 0,
    max_retries INTEGER,
    retry_strategy VARCHAR(50),
    error_metadata JSONB,
    caught_by_node_id VARCHAR(255), -- ID of the catch/try node that handled this error
    recovery_action VARCHAR(100), -- retry, fallback, propagate, handled
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_error_handling_history_tenant ON error_handling_history(tenant_id);
CREATE INDEX idx_error_handling_history_execution ON error_handling_history(execution_id);
CREATE INDEX idx_error_handling_history_node ON error_handling_history(node_id);
CREATE INDEX idx_error_handling_history_type ON error_handling_history(error_type);
CREATE INDEX idx_error_handling_history_classification ON error_handling_history(error_classification);
CREATE INDEX idx_error_handling_history_created ON error_handling_history(created_at);

-- Add comments for documentation
COMMENT ON TABLE error_handling_history IS 'Tracks detailed error history for workflow executions including retry attempts and error handling actions';
COMMENT ON COLUMN error_handling_history.error_classification IS 'Classification of error: transient, permanent, or unknown';
COMMENT ON COLUMN error_handling_history.retry_strategy IS 'Retry strategy used: fixed, exponential, exponential_jitter';
COMMENT ON COLUMN error_handling_history.caught_by_node_id IS 'ID of the try/catch node that handled this error';
COMMENT ON COLUMN error_handling_history.recovery_action IS 'Action taken: retry (retried), fallback (ran fallback), propagate (rethrown), handled (caught and handled)';

-- Add error handling statistics to workflows table
-- This helps track overall error patterns per workflow
ALTER TABLE workflows
ADD COLUMN IF NOT EXISTS error_statistics JSONB DEFAULT '{"total_errors": 0, "transient_errors": 0, "permanent_errors": 0, "caught_errors": 0}';

COMMENT ON COLUMN workflows.error_statistics IS 'Aggregated error statistics for the workflow';
