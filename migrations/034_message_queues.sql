-- +goose Up
-- SQL in this section is executed when the migration is applied

-- Create table for tracking queue messages
CREATE TABLE IF NOT EXISTS queue_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    workflow_id UUID REFERENCES workflows(id) ON DELETE CASCADE,
    execution_id UUID REFERENCES executions(id) ON DELETE CASCADE,
    queue_type VARCHAR(50) NOT NULL,
    destination VARCHAR(500) NOT NULL,
    message_id VARCHAR(500),
    status VARCHAR(50) NOT NULL DEFAULT 'sent',
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX idx_queue_messages_tenant ON queue_messages(tenant_id);
CREATE INDEX idx_queue_messages_workflow ON queue_messages(workflow_id);
CREATE INDEX idx_queue_messages_execution ON queue_messages(execution_id);
CREATE INDEX idx_queue_messages_status ON queue_messages(status);
CREATE INDEX idx_queue_messages_queue_type ON queue_messages(queue_type);
CREATE INDEX idx_queue_messages_sent_at ON queue_messages(sent_at DESC);

-- Add comments for documentation
COMMENT ON TABLE queue_messages IS 'Audit log for message queue operations';
COMMENT ON COLUMN queue_messages.queue_type IS 'Type of message queue: sqs, kafka, rabbitmq';
COMMENT ON COLUMN queue_messages.destination IS 'Queue URL, topic name, or exchange name';
COMMENT ON COLUMN queue_messages.message_id IS 'Message ID returned by the queue system';
COMMENT ON COLUMN queue_messages.status IS 'Message status: sent, acknowledged, failed';

-- +goose Down
-- SQL in this section is executed when the migration is rolled back

DROP INDEX IF EXISTS idx_queue_messages_sent_at;
DROP INDEX IF EXISTS idx_queue_messages_queue_type;
DROP INDEX IF EXISTS idx_queue_messages_status;
DROP INDEX IF EXISTS idx_queue_messages_execution;
DROP INDEX IF EXISTS idx_queue_messages_workflow;
DROP INDEX IF EXISTS idx_queue_messages_tenant;

DROP TABLE IF EXISTS queue_messages;
