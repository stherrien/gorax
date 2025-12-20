-- Migration: Add support for sub-workflow execution
-- Phase 4.2: Sub-workflows

-- Add parent_execution_id column to executions table for nested workflow tracking
ALTER TABLE executions
ADD COLUMN parent_execution_id UUID REFERENCES executions(id) ON DELETE CASCADE,
ADD COLUMN execution_depth INTEGER NOT NULL DEFAULT 0;

-- Create index for parent-child execution queries
CREATE INDEX idx_executions_parent_id ON executions(parent_execution_id);
CREATE INDEX idx_executions_depth ON executions(execution_depth);

-- Add comment to document the purpose
COMMENT ON COLUMN executions.parent_execution_id IS 'References parent execution if this is a sub-workflow execution';
COMMENT ON COLUMN executions.execution_depth IS 'Depth of nested execution (0 = root, 1+ = sub-workflow)';
