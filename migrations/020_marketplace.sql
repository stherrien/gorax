-- Migration: 020_marketplace.sql
-- Description: Create marketplace templates tables for template sharing

-- Marketplace templates table
CREATE TABLE IF NOT EXISTS marketplace_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    category VARCHAR(100) NOT NULL,
    definition JSONB NOT NULL,
    tags TEXT[] DEFAULT '{}',
    author_id VARCHAR(255) NOT NULL,
    author_name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    download_count INTEGER DEFAULT 0,
    average_rating DECIMAL(3,2) DEFAULT 0.0,
    total_ratings INTEGER DEFAULT 0,
    is_verified BOOLEAN DEFAULT FALSE,
    source_tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
    source_template_id UUID,  -- No foreign key since workflow_templates doesn't exist
    published_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_rating CHECK (average_rating >= 0 AND average_rating <= 5),
    CONSTRAINT valid_download_count CHECK (download_count >= 0),
    CONSTRAINT valid_total_ratings CHECK (total_ratings >= 0)
);

-- Template installations table
CREATE TABLE IF NOT EXISTS marketplace_installations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES marketplace_templates(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    installed_version VARCHAR(50) NOT NULL,
    installed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_tenant_template UNIQUE (tenant_id, template_id, workflow_id)
);

-- Template reviews table
CREATE TABLE IF NOT EXISTS marketplace_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES marketplace_templates(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    user_name VARCHAR(255) NOT NULL,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_user_template_review UNIQUE (tenant_id, user_id, template_id)
);

-- Template versions table (for version history)
CREATE TABLE IF NOT EXISTS marketplace_template_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES marketplace_templates(id) ON DELETE CASCADE,
    version VARCHAR(50) NOT NULL,
    definition JSONB NOT NULL,
    change_notes TEXT,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_template_version UNIQUE (template_id, version)
);

-- Create indexes for better query performance
CREATE INDEX idx_marketplace_templates_category ON marketplace_templates(category);
CREATE INDEX idx_marketplace_templates_author ON marketplace_templates(author_id);
CREATE INDEX idx_marketplace_templates_download_count ON marketplace_templates(download_count DESC);
CREATE INDEX idx_marketplace_templates_rating ON marketplace_templates(average_rating DESC);
CREATE INDEX idx_marketplace_templates_published_at ON marketplace_templates(published_at DESC);
CREATE INDEX idx_marketplace_templates_tags ON marketplace_templates USING GIN(tags);
CREATE INDEX idx_marketplace_templates_name ON marketplace_templates(name);

CREATE INDEX idx_marketplace_installations_template ON marketplace_installations(template_id);
CREATE INDEX idx_marketplace_installations_tenant ON marketplace_installations(tenant_id);
CREATE INDEX idx_marketplace_installations_installed_at ON marketplace_installations(installed_at DESC);

CREATE INDEX idx_marketplace_reviews_template ON marketplace_reviews(template_id);
CREATE INDEX idx_marketplace_reviews_tenant_user ON marketplace_reviews(tenant_id, user_id);
CREATE INDEX idx_marketplace_reviews_created_at ON marketplace_reviews(created_at DESC);

CREATE INDEX idx_marketplace_template_versions_template ON marketplace_template_versions(template_id);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_marketplace_template_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at
CREATE TRIGGER marketplace_templates_updated_at
    BEFORE UPDATE ON marketplace_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_marketplace_template_updated_at();

CREATE TRIGGER marketplace_reviews_updated_at
    BEFORE UPDATE ON marketplace_reviews
    FOR EACH ROW
    EXECUTE FUNCTION update_marketplace_template_updated_at();

-- Insert some seed data for popular categories
COMMENT ON TABLE marketplace_templates IS 'Stores publicly shared workflow templates in the marketplace';
COMMENT ON TABLE marketplace_installations IS 'Tracks template installations by tenants';
COMMENT ON TABLE marketplace_reviews IS 'Stores user reviews and ratings for marketplace templates';
COMMENT ON TABLE marketplace_template_versions IS 'Version history for marketplace templates';
