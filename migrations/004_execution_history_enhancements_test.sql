-- Test script for 004_execution_history_enhancements.sql migration
-- This validates that the migration creates all necessary indexes and columns

-- Test 1: Verify retention_until column exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'executions'
        AND column_name = 'retention_until'
        AND data_type = 'timestamp with time zone'
    ) THEN
        RAISE EXCEPTION 'retention_until column not found on executions table';
    END IF;
END $$;

-- Test 2: Verify composite index for cursor-based pagination exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE tablename = 'executions'
        AND indexname = 'idx_executions_cursor_pagination'
    ) THEN
        RAISE EXCEPTION 'idx_executions_cursor_pagination index not found';
    END IF;
END $$;

-- Test 3: Verify status filter index exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE tablename = 'executions'
        AND indexname = 'idx_executions_tenant_status'
    ) THEN
        RAISE EXCEPTION 'idx_executions_tenant_status index not found';
    END IF;
END $$;

-- Test 4: Verify workflow_id filter index exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE tablename = 'executions'
        AND indexname = 'idx_executions_tenant_workflow'
    ) THEN
        RAISE EXCEPTION 'idx_executions_tenant_workflow index not found';
    END IF;
END $$;

-- Test 5: Verify trigger_type filter index exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE tablename = 'executions'
        AND indexname = 'idx_executions_tenant_trigger_type'
    ) THEN
        RAISE EXCEPTION 'idx_executions_tenant_trigger_type index not found';
    END IF;
END $$;

-- Test 6: Verify date range filter index exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE tablename = 'executions'
        AND indexname = 'idx_executions_tenant_created_at'
    ) THEN
        RAISE EXCEPTION 'idx_executions_tenant_created_at index not found';
    END IF;
END $$;

-- Test 7: Verify retention_until index exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE tablename = 'executions'
        AND indexname = 'idx_executions_retention_until'
    ) THEN
        RAISE EXCEPTION 'idx_executions_retention_until index not found';
    END IF;
END $$;

-- Test 8: Verify retention_until can be set and queried
DO $$
DECLARE
    test_tenant_id UUID := uuid_generate_v4();
    test_workflow_id UUID := uuid_generate_v4();
    test_execution_id UUID;
    test_retention_date TIMESTAMPTZ := NOW() + INTERVAL '90 days';
BEGIN
    -- Create test tenant
    INSERT INTO tenants (id, name, subdomain)
    VALUES (test_tenant_id, 'Test Tenant', 'test-tenant-migration');

    -- Create test workflow
    INSERT INTO workflows (id, tenant_id, name, definition, created_by)
    VALUES (test_workflow_id, test_tenant_id, 'Test Workflow', '{}', test_tenant_id);

    -- Create test execution with retention_until
    INSERT INTO executions (id, tenant_id, workflow_id, workflow_version, trigger_type, retention_until)
    VALUES (uuid_generate_v4(), test_tenant_id, test_workflow_id, 1, 'manual', test_retention_date)
    RETURNING id INTO test_execution_id;

    -- Verify retention_until was set correctly
    IF NOT EXISTS (
        SELECT 1 FROM executions
        WHERE id = test_execution_id
        AND retention_until = test_retention_date
    ) THEN
        RAISE EXCEPTION 'retention_until not set correctly';
    END IF;

    -- Clean up test data
    DELETE FROM executions WHERE id = test_execution_id;
    DELETE FROM workflows WHERE id = test_workflow_id;
    DELETE FROM tenants WHERE id = test_tenant_id;
END $$;

-- All tests passed
SELECT 'All migration tests passed' AS result;
