-- Migration: 026_marketplace_enhancements.sql
-- Description: Add categories, featured templates, and enhanced search capabilities

-- Create categories table with hierarchy support
CREATE TABLE IF NOT EXISTS marketplace_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    icon VARCHAR(100),
    parent_id UUID REFERENCES marketplace_categories(id) ON DELETE CASCADE,
    display_order INTEGER DEFAULT 0,
    template_count INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_category_name UNIQUE (name, parent_id)
);

-- Create template_categories junction table (many-to-many)
CREATE TABLE IF NOT EXISTS marketplace_template_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES marketplace_templates(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES marketplace_categories(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_template_category UNIQUE (template_id, category_id)
);

-- Add featured columns to marketplace_templates
ALTER TABLE marketplace_templates
ADD COLUMN IF NOT EXISTS is_featured BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS featured_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS featured_by VARCHAR(255);

-- Add full-text search column
ALTER TABLE marketplace_templates
ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Create indexes for categories
CREATE INDEX idx_marketplace_categories_slug ON marketplace_categories(slug);
CREATE INDEX idx_marketplace_categories_parent ON marketplace_categories(parent_id);
CREATE INDEX idx_marketplace_categories_order ON marketplace_categories(display_order);

-- Create indexes for template_categories junction
CREATE INDEX idx_template_categories_template ON marketplace_template_categories(template_id);
CREATE INDEX idx_template_categories_category ON marketplace_template_categories(category_id);

-- Create indexes for featured templates
CREATE INDEX idx_marketplace_templates_featured ON marketplace_templates(is_featured, featured_at DESC) WHERE is_featured = TRUE;

-- Create full-text search index
CREATE INDEX idx_marketplace_templates_search ON marketplace_templates USING GIN(search_vector);

-- Function to update search vector
CREATE OR REPLACE FUNCTION update_marketplace_template_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.name, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(array_to_string(NEW.tags, ' '), '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for search vector updates
CREATE TRIGGER marketplace_templates_search_vector_update
    BEFORE INSERT OR UPDATE OF name, description, tags
    ON marketplace_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_marketplace_template_search_vector();

-- Function to update category updated_at timestamp
CREATE OR REPLACE FUNCTION update_marketplace_category_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for category updates
CREATE TRIGGER marketplace_categories_updated_at
    BEFORE UPDATE ON marketplace_categories
    FOR EACH ROW
    EXECUTE FUNCTION update_marketplace_category_updated_at();

-- Function to update category template counts
CREATE OR REPLACE FUNCTION update_category_template_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE marketplace_categories
        SET template_count = template_count + 1
        WHERE id = NEW.category_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE marketplace_categories
        SET template_count = GREATEST(template_count - 1, 0)
        WHERE id = OLD.category_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger to maintain category counts
CREATE TRIGGER update_category_count_on_template_category_change
    AFTER INSERT OR DELETE ON marketplace_template_categories
    FOR EACH ROW
    EXECUTE FUNCTION update_category_template_count();

-- Insert seed categories
INSERT INTO marketplace_categories (name, slug, description, icon, display_order) VALUES
    ('Integration', 'integration', 'Templates for integrating with external services and APIs', 'link', 1),
    ('Automation', 'automation', 'Templates for automating repetitive tasks and workflows', 'zap', 2),
    ('Data Processing', 'data-processing', 'Templates for data transformation and processing', 'database', 3),
    ('Notifications', 'notifications', 'Templates for sending notifications across different channels', 'bell', 4),
    ('DevOps', 'devops', 'Templates for CI/CD, deployment, and infrastructure automation', 'server', 5),
    ('Security', 'security', 'Templates for security monitoring and compliance checks', 'shield', 6),
    ('Analytics', 'analytics', 'Templates for data analysis and reporting', 'bar-chart', 7),
    ('Communication', 'communication', 'Templates for team communication and collaboration', 'message-circle', 8),
    ('Monitoring', 'monitoring', 'Templates for system and application monitoring', 'activity', 9),
    ('Scheduling', 'scheduling', 'Templates for task scheduling and time-based automation', 'clock', 10)
ON CONFLICT (slug) DO NOTHING;

-- Update existing templates' search vectors
UPDATE marketplace_templates
SET search_vector =
    setweight(to_tsvector('english', COALESCE(name, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(description, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(array_to_string(tags, ' '), '')), 'C')
WHERE search_vector IS NULL;

-- Migrate existing category field to new structure
-- (Only if there are templates - this will map old string categories to new category table)
DO $$
DECLARE
    template_record RECORD;
    category_id UUID;
BEGIN
    FOR template_record IN SELECT id, category FROM marketplace_templates WHERE category IS NOT NULL AND category != ''
    LOOP
        -- Find or create category from old category string
        SELECT id INTO category_id
        FROM marketplace_categories
        WHERE LOWER(name) = LOWER(template_record.category)
        LIMIT 1;

        -- If found, create association
        IF category_id IS NOT NULL THEN
            INSERT INTO marketplace_template_categories (template_id, category_id)
            VALUES (template_record.id, category_id)
            ON CONFLICT (template_id, category_id) DO NOTHING;
        END IF;
    END LOOP;
END $$;

-- Add comments for documentation
COMMENT ON TABLE marketplace_categories IS 'Categories for organizing marketplace templates with hierarchy support';
COMMENT ON TABLE marketplace_template_categories IS 'Junction table for many-to-many relationship between templates and categories';
COMMENT ON COLUMN marketplace_templates.is_featured IS 'Whether this template is featured on the marketplace homepage';
COMMENT ON COLUMN marketplace_templates.featured_at IS 'Timestamp when the template was featured';
COMMENT ON COLUMN marketplace_templates.search_vector IS 'Full-text search index for template name, description, and tags';
