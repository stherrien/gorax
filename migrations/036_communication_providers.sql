-- Migration: Communication Providers
-- Description: Add tables for tracking email and SMS communications sent through workflow actions
-- Version: 036

-- Table for tracking communication events
CREATE TABLE IF NOT EXISTS communication_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    execution_id UUID REFERENCES executions(id) ON DELETE CASCADE,
    workflow_id UUID REFERENCES workflows(id) ON DELETE SET NULL,

    -- Communication details
    communication_type VARCHAR(50) NOT NULL CHECK (communication_type IN ('email', 'sms')),
    provider VARCHAR(50) NOT NULL,

    -- Recipient information (stored for audit purposes)
    recipient VARCHAR(500) NOT NULL,
    sender VARCHAR(500),

    -- Message identifiers from providers
    message_id VARCHAR(500),
    external_id VARCHAR(500), -- Provider-specific ID

    -- Status tracking
    status VARCHAR(50) NOT NULL CHECK (status IN ('sent', 'failed', 'queued', 'delivered', 'bounced')),
    error_message TEXT,

    -- Cost tracking (for SMS primarily)
    cost DECIMAL(10, 4),
    currency VARCHAR(3) DEFAULT 'USD',

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Timestamps
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    delivered_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient querying
CREATE INDEX idx_communication_events_tenant ON communication_events(tenant_id);
CREATE INDEX idx_communication_events_execution ON communication_events(execution_id);
CREATE INDEX idx_communication_events_workflow ON communication_events(workflow_id);
CREATE INDEX idx_communication_events_type ON communication_events(communication_type);
CREATE INDEX idx_communication_events_status ON communication_events(status);
CREATE INDEX idx_communication_events_sent_at ON communication_events(sent_at);
CREATE INDEX idx_communication_events_recipient ON communication_events(recipient);

-- Composite index for common queries
CREATE INDEX idx_communication_events_tenant_type_sent ON communication_events(tenant_id, communication_type, sent_at DESC);

-- Table for storing communication templates (optional - for future use)
CREATE TABLE IF NOT EXISTS communication_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Template details
    name VARCHAR(255) NOT NULL,
    description TEXT,
    communication_type VARCHAR(50) NOT NULL CHECK (communication_type IN ('email', 'sms')),

    -- Template content
    subject VARCHAR(500), -- For email only
    body TEXT NOT NULL,
    body_html TEXT, -- For email only

    -- Template variables (e.g., {{name}}, {{code}})
    variables JSONB DEFAULT '[]',

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Status
    is_active BOOLEAN DEFAULT true,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),

    -- Ensure unique template names per tenant
    CONSTRAINT unique_template_name_per_tenant UNIQUE (tenant_id, name)
);

-- Indexes for templates
CREATE INDEX idx_communication_templates_tenant ON communication_templates(tenant_id);
CREATE INDEX idx_communication_templates_type ON communication_templates(communication_type);
CREATE INDEX idx_communication_templates_active ON communication_templates(is_active);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_communication_events_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER communication_events_updated_at
    BEFORE UPDATE ON communication_events
    FOR EACH ROW
    EXECUTE FUNCTION update_communication_events_updated_at();

CREATE TRIGGER communication_templates_updated_at
    BEFORE UPDATE ON communication_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_communication_events_updated_at();

-- Add comments for documentation
COMMENT ON TABLE communication_events IS 'Tracks all email and SMS communications sent through workflow actions';
COMMENT ON COLUMN communication_events.communication_type IS 'Type of communication: email or sms';
COMMENT ON COLUMN communication_events.provider IS 'Communication provider used (e.g., sendgrid, twilio)';
COMMENT ON COLUMN communication_events.status IS 'Current status of the communication';
COMMENT ON COLUMN communication_events.cost IS 'Cost of sending the communication (primarily for SMS)';

COMMENT ON TABLE communication_templates IS 'Reusable templates for email and SMS communications';
COMMENT ON COLUMN communication_templates.variables IS 'Array of variable names used in the template';
