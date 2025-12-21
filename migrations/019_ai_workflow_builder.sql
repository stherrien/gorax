-- AI Workflow Builder Schema
-- Supports natural language to workflow generation with multi-turn conversations

-- Conversations table for tracking multi-turn workflow building sessions
CREATE TABLE IF NOT EXISTS aibuilder_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    current_workflow JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_aibuilder_conversation_status CHECK (status IN ('active', 'completed', 'abandoned'))
);

-- Messages table for conversation history
CREATE TABLE IF NOT EXISTS aibuilder_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES aibuilder_conversations(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    workflow JSONB,
    prompt_tokens INTEGER DEFAULT 0,
    completion_tokens INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_aibuilder_message_role CHECK (role IN ('user', 'assistant', 'system'))
);

-- Generated workflows table for tracking successful generations
CREATE TABLE IF NOT EXISTS aibuilder_generated_workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    conversation_id UUID NOT NULL REFERENCES aibuilder_conversations(id) ON DELETE CASCADE,
    workflow_id UUID REFERENCES workflows(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    definition JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'generated',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    applied_at TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT chk_aibuilder_workflow_status CHECK (status IN ('generated', 'applied', 'discarded'))
);

-- Node templates table for LLM context
CREATE TABLE IF NOT EXISTS aibuilder_node_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID, -- NULL for global templates
    node_type VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(50) NOT NULL,
    config_schema JSONB NOT NULL,
    example_config JSONB,
    llm_description TEXT NOT NULL, -- Description optimized for LLM understanding
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_aibuilder_node_category CHECK (
        category IN ('trigger', 'action', 'control', 'integration')
    ),

    -- Unique node type per tenant (or global)
    CONSTRAINT uq_aibuilder_node_type UNIQUE (tenant_id, node_type)
);

-- Usage tracking for analytics
CREATE TABLE IF NOT EXISTS aibuilder_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    conversation_id UUID REFERENCES aibuilder_conversations(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL,
    prompt_tokens INTEGER DEFAULT 0,
    completion_tokens INTEGER DEFAULT 0,
    model VARCHAR(100),
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_aibuilder_usage_action CHECK (
        action IN ('generate', 'refine', 'apply', 'abandon')
    )
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_aibuilder_conversations_tenant ON aibuilder_conversations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_aibuilder_conversations_user ON aibuilder_conversations(tenant_id, user_id);
CREATE INDEX IF NOT EXISTS idx_aibuilder_conversations_status ON aibuilder_conversations(status);
CREATE INDEX IF NOT EXISTS idx_aibuilder_conversations_created ON aibuilder_conversations(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_aibuilder_messages_conversation ON aibuilder_messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_aibuilder_messages_created ON aibuilder_messages(created_at);

CREATE INDEX IF NOT EXISTS idx_aibuilder_generated_tenant ON aibuilder_generated_workflows(tenant_id);
CREATE INDEX IF NOT EXISTS idx_aibuilder_generated_conversation ON aibuilder_generated_workflows(conversation_id);
CREATE INDEX IF NOT EXISTS idx_aibuilder_generated_workflow ON aibuilder_generated_workflows(workflow_id);

CREATE INDEX IF NOT EXISTS idx_aibuilder_node_templates_type ON aibuilder_node_templates(node_type);
CREATE INDEX IF NOT EXISTS idx_aibuilder_node_templates_category ON aibuilder_node_templates(category);
CREATE INDEX IF NOT EXISTS idx_aibuilder_node_templates_active ON aibuilder_node_templates(is_active) WHERE is_active = true;

CREATE INDEX IF NOT EXISTS idx_aibuilder_usage_tenant ON aibuilder_usage(tenant_id);
CREATE INDEX IF NOT EXISTS idx_aibuilder_usage_user ON aibuilder_usage(tenant_id, user_id);
CREATE INDEX IF NOT EXISTS idx_aibuilder_usage_created ON aibuilder_usage(created_at DESC);

-- Seed data: Node templates for LLM context
INSERT INTO aibuilder_node_templates (tenant_id, node_type, name, description, category, config_schema, example_config, llm_description)
VALUES
    -- Triggers
    (NULL, 'trigger:webhook', 'Webhook Trigger', 'Starts workflow when HTTP webhook is received', 'trigger',
     '{"type": "object", "properties": {"path": {"type": "string"}, "auth_type": {"type": "string", "enum": ["none", "basic", "signature", "api_key"]}, "secret": {"type": "string"}}}'::jsonb,
     '{"path": "/my-webhook", "auth_type": "signature", "secret": "my-secret"}'::jsonb,
     'Use this to start a workflow when an external system sends an HTTP POST request. Configure authentication (none, basic auth, signature verification, or API key) for security. The webhook payload will be available in subsequent steps.'),

    (NULL, 'trigger:schedule', 'Schedule Trigger', 'Starts workflow on a cron schedule', 'trigger',
     '{"type": "object", "properties": {"cron": {"type": "string"}, "timezone": {"type": "string"}}}'::jsonb,
     '{"cron": "0 9 * * 1-5", "timezone": "America/New_York"}'::jsonb,
     'Use this to run a workflow automatically at scheduled times. Specify a cron expression (e.g., "0 9 * * 1-5" for 9am weekdays) and optional timezone.'),

    -- Actions
    (NULL, 'action:http', 'HTTP Request', 'Makes HTTP API calls to external services', 'action',
     '{"type": "object", "properties": {"method": {"type": "string", "enum": ["GET", "POST", "PUT", "PATCH", "DELETE"]}, "url": {"type": "string"}, "headers": {"type": "object"}, "body": {"type": "object"}, "timeout": {"type": "integer"}}}'::jsonb,
     '{"method": "POST", "url": "https://api.example.com/data", "headers": {"Content-Type": "application/json"}, "body": {"key": "value"}, "timeout": 30}'::jsonb,
     'Use this to call REST APIs. Configure the HTTP method, URL, headers, and request body. Supports template variables like ${steps.trigger.body.data} for dynamic values.'),

    (NULL, 'action:transform', 'Transform Data', 'Transforms data using JSONPath or mapping', 'action',
     '{"type": "object", "properties": {"expression": {"type": "string"}, "mapping": {"type": "object"}}}'::jsonb,
     '{"mapping": {"user_name": "${steps.trigger.body.user.name}", "user_email": "${steps.trigger.body.user.email}"}}'::jsonb,
     'Use this to reshape, filter, or combine data. Create mappings to extract and rename fields from previous steps. Use ${steps.nodeName.output.field} syntax.'),

    (NULL, 'action:code', 'JavaScript Code', 'Executes custom JavaScript code', 'action',
     '{"type": "object", "properties": {"script": {"type": "string"}, "timeout": {"type": "integer"}}}'::jsonb,
     '{"script": "const input = context.steps.transform.output;\nreturn { processed: input.data.map(x => x * 2) };", "timeout": 30}'::jsonb,
     'Use this for complex logic that cannot be expressed with other nodes. Write JavaScript code that returns an object. Access previous steps via context.steps.nodeName.output.'),

    (NULL, 'action:formula', 'Formula', 'Evaluates mathematical or logical expressions', 'action',
     '{"type": "object", "properties": {"expression": {"type": "string"}, "output_variable": {"type": "string"}}}'::jsonb,
     '{"expression": "${steps.data.output.price} * ${steps.data.output.quantity} * 1.1", "output_variable": "total_with_tax"}'::jsonb,
     'Use this for calculations. Write expressions using ${variable} syntax. Supports math operators, comparisons, and logical operators.'),

    (NULL, 'action:email', 'Send Email', 'Sends email notifications', 'action',
     '{"type": "object", "properties": {"to": {"type": "string"}, "subject": {"type": "string"}, "body": {"type": "string"}, "cc": {"type": "string"}, "bcc": {"type": "string"}}}'::jsonb,
     '{"to": "user@example.com", "subject": "Workflow Alert", "body": "Data processed: ${steps.process.output.count} items"}'::jsonb,
     'Use this to send email notifications. Configure recipients, subject, and body. Use template variables for dynamic content.'),

    -- Control flow
    (NULL, 'control:if', 'Conditional (If/Else)', 'Branches workflow based on a condition', 'control',
     '{"type": "object", "properties": {"condition": {"type": "string"}, "description": {"type": "string"}}}'::jsonb,
     '{"condition": "${steps.data.output.status} == \"approved\"", "description": "Check if status is approved"}'::jsonb,
     'Use this to branch workflow logic. Write a condition that evaluates to true/false. Connect "true" and "false" edges to different paths.'),

    (NULL, 'control:loop', 'Loop (For Each)', 'Iterates over an array', 'control',
     '{"type": "object", "properties": {"source": {"type": "string"}, "item_variable": {"type": "string"}, "index_variable": {"type": "string"}, "max_iterations": {"type": "integer"}}}'::jsonb,
     '{"source": "${steps.data.output.items}", "item_variable": "item", "index_variable": "index", "max_iterations": 100}'::jsonb,
     'Use this to process each item in an array. Specify the source array and variable names. Nodes inside the loop can access ${loop.item} and ${loop.index}.'),

    (NULL, 'control:delay', 'Delay', 'Pauses workflow execution', 'control',
     '{"type": "object", "properties": {"duration": {"type": "string"}}}'::jsonb,
     '{"duration": "5s"}'::jsonb,
     'Use this to pause the workflow. Specify duration as "5s" (seconds), "2m" (minutes), or "1h" (hours). Can use variables: "${steps.config.output.delay}".'),

    (NULL, 'control:parallel', 'Parallel', 'Executes multiple branches in parallel', 'control',
     '{"type": "object", "properties": {"error_strategy": {"type": "string", "enum": ["fail_fast", "wait_all"]}, "max_concurrency": {"type": "integer"}}}'::jsonb,
     '{"error_strategy": "fail_fast", "max_concurrency": 5}'::jsonb,
     'Use this to run multiple branches simultaneously. Choose "fail_fast" to stop on first error or "wait_all" to complete all branches regardless of errors.'),

    -- Integrations
    (NULL, 'slack:send_message', 'Slack: Send Message', 'Sends a message to a Slack channel', 'integration',
     '{"type": "object", "properties": {"channel": {"type": "string"}, "text": {"type": "string"}, "blocks": {"type": "array"}}}'::jsonb,
     '{"channel": "#general", "text": "New alert: ${steps.data.output.message}"}'::jsonb,
     'Use this to send messages to Slack channels. Specify channel name or ID and message text. Supports Block Kit for rich formatting.'),

    (NULL, 'slack:send_dm', 'Slack: Send Direct Message', 'Sends a direct message to a Slack user', 'integration',
     '{"type": "object", "properties": {"user_id": {"type": "string"}, "text": {"type": "string"}}}'::jsonb,
     '{"user_id": "${steps.lookup.output.slack_user_id}", "text": "You have a new task assigned"}'::jsonb,
     'Use this to send direct messages to Slack users. Specify user ID (not username) and message text.')

ON CONFLICT (tenant_id, node_type) DO NOTHING;

-- Add trigger function for updated_at
CREATE OR REPLACE FUNCTION update_aibuilder_conversations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for conversations
DROP TRIGGER IF EXISTS trg_aibuilder_conversations_updated ON aibuilder_conversations;
CREATE TRIGGER trg_aibuilder_conversations_updated
    BEFORE UPDATE ON aibuilder_conversations
    FOR EACH ROW
    EXECUTE FUNCTION update_aibuilder_conversations_updated_at();

-- Add trigger for node templates
CREATE OR REPLACE FUNCTION update_aibuilder_node_templates_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_aibuilder_node_templates_updated ON aibuilder_node_templates;
CREATE TRIGGER trg_aibuilder_node_templates_updated
    BEFORE UPDATE ON aibuilder_node_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_aibuilder_node_templates_updated_at();

COMMENT ON TABLE aibuilder_conversations IS 'Multi-turn conversations for AI workflow building';
COMMENT ON TABLE aibuilder_messages IS 'Message history for AI builder conversations';
COMMENT ON TABLE aibuilder_generated_workflows IS 'Workflows generated by the AI builder';
COMMENT ON TABLE aibuilder_node_templates IS 'Node type definitions for LLM context';
COMMENT ON TABLE aibuilder_usage IS 'Usage tracking for AI builder analytics';
