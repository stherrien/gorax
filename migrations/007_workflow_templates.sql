-- Workflow Templates Library
-- Enables reusable workflow patterns across tenants

-- Workflow templates table
CREATE TABLE workflow_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    definition JSONB NOT NULL,
    tags TEXT[],
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_template_name_per_tenant UNIQUE (tenant_id, name)
);

CREATE INDEX idx_templates_tenant ON workflow_templates(tenant_id);
CREATE INDEX idx_templates_category ON workflow_templates(category);
CREATE INDEX idx_templates_tags ON workflow_templates USING GIN(tags);
CREATE INDEX idx_templates_is_public ON workflow_templates(is_public);
CREATE INDEX idx_templates_created_at ON workflow_templates(created_at DESC);

-- Enable Row Level Security
ALTER TABLE workflow_templates ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Users can see their tenant's templates + public templates
CREATE POLICY tenant_isolation_templates ON workflow_templates
    USING (
        tenant_id = current_setting('app.current_tenant_id', true)::UUID
        OR is_public = true
    );

-- Trigger for updated_at
CREATE TRIGGER update_templates_updated_at BEFORE UPDATE ON workflow_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
