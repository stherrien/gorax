-- Add usage count tracking to workflow templates
-- Tracks how many times a template has been instantiated

ALTER TABLE workflow_templates
ADD COLUMN usage_count INTEGER NOT NULL DEFAULT 0;

CREATE INDEX idx_templates_usage_count ON workflow_templates(usage_count DESC);

-- Update any existing templates to have 0 usage count
UPDATE workflow_templates
SET usage_count = 0
WHERE usage_count IS NULL;
