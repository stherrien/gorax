# Retention Policy Service

The retention policy service provides automatic cleanup of old workflow execution records based on configurable retention periods. This helps manage database size and comply with data retention policies.

## Features

- **Configurable Retention Periods**: Set retention periods per tenant (stored in tenant settings)
- **Batch Processing**: Deletes records in configurable batches to avoid database locks
- **Tenant Isolation**: Respects row-level security (RLS) policies
- **Audit Logging**: Tracks all cleanup operations for compliance
- **Scheduled Cleanup**: Runs on a configurable interval (default: daily)
- **Foreign Key Safety**: Deletes step_executions before executions to maintain referential integrity

## Architecture

### Components

1. **Service** (`service.go`): Business logic for retention policies
   - Gets retention policy per tenant
   - Orchestrates cleanup operations
   - Handles audit logging

2. **Repository** (`repository.go`): Database operations
   - Batch deletion of old executions
   - Retrieves retention policies from tenant settings
   - Logs cleanup operations

3. **Scheduler** (`scheduler.go`): Background job scheduler
   - Runs cleanup on a configurable interval
   - Supports one-time manual execution
   - Graceful start/stop

4. **Models** (`model.go`): Data structures
   - RetentionPolicy
   - CleanupResult
   - CleanupLog

## Database Schema

### Tenant Settings (JSONB)

Retention settings are stored in the `tenants.settings` JSONB column:

```json
{
  "retention_days": 90,
  "retention_enabled": true
}
```

### Cleanup Logs Table

Audit logs are stored in `retention_cleanup_logs`:

```sql
CREATE TABLE retention_cleanup_logs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    executions_deleted INTEGER,
    step_executions_deleted INTEGER,
    retention_days INTEGER,
    cutoff_date TIMESTAMPTZ,
    duration_ms INTEGER,
    status VARCHAR(50), -- 'completed' or 'failed'
    error_message TEXT,
    created_at TIMESTAMPTZ
);
```

## Configuration

Environment variables:

```bash
# Enable/disable retention cleanup
RETENTION_ENABLED=true

# Default retention period in days
RETENTION_DEFAULT_DAYS=90

# Number of executions to delete per batch
RETENTION_BATCH_SIZE=1000

# How often to run cleanup (Go duration format)
RETENTION_RUN_INTERVAL=24h

# Enable audit logging
RETENTION_ENABLE_AUDIT_LOG=true
```

## Usage

### Basic Setup

```go
package main

import (
    "context"
    "log/slog"
    "time"

    "github.com/gorax/gorax/internal/retention"
    "github.com/jmoiron/sqlx"
)

func main() {
    // Connect to database
    db, _ := sqlx.Connect("postgres", connectionString)

    // Create repository
    repo := retention.NewRepository(db)

    // Create service with config
    config := retention.Config{
        DefaultRetentionDays: 90,
        BatchSize:            1000,
        EnableAuditLog:       true,
    }
    logger := slog.Default()
    service := retention.NewService(repo, logger, config)

    // Create and start scheduler
    scheduler := retention.NewScheduler(service, logger, 24*time.Hour)
    ctx := context.Background()
    scheduler.Start(ctx)

    // ... application runs ...

    // Graceful shutdown
    scheduler.Stop()
}
```

### Manual Cleanup

```go
// Run cleanup for a specific tenant
result, err := service.CleanupOldExecutions(ctx, tenantID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Deleted %d executions and %d step executions\n",
    result.ExecutionsDeleted,
    result.StepExecutionsDeleted)
```

### One-Time Cleanup

```go
// Run cleanup once for all tenants
scheduler := retention.NewScheduler(service, logger, 24*time.Hour)
result, err := scheduler.RunOnce(ctx)
```

### Custom Retention Policy

Update tenant settings in the database:

```sql
UPDATE tenants
SET settings = jsonb_set(
    jsonb_set(
        settings,
        '{retention_days}',
        '60'::jsonb
    ),
    '{retention_enabled}',
    'true'::jsonb
)
WHERE id = 'tenant-id';
```

Or programmatically:

```go
err := repo.SetRetentionPolicy(ctx, tenantID, 60, true)
```

## Batch Processing

The service processes deletions in batches to avoid long-running database locks:

1. Queries for execution IDs to delete (batch size limit)
2. Deletes associated step_executions (foreign key order)
3. Deletes executions
4. Commits transaction
5. Repeats until no more records to delete

Each batch includes a small delay (100ms) between iterations to reduce database load.

## Retention Rules

- Only deletes executions with status `completed` or `failed`
- Never deletes executions with status `running` or `pending`
- Cutoff date is calculated as: `NOW() - retention_days`
- Executions created before cutoff date are eligible for deletion

## Monitoring

### Metrics

The service logs the following information:

- Number of executions deleted
- Number of step executions deleted
- Number of batches processed
- Duration of cleanup operation
- Any errors encountered

### Audit Logs

Query cleanup history:

```sql
SELECT
    tenant_id,
    executions_deleted,
    step_executions_deleted,
    retention_days,
    cutoff_date,
    duration_ms,
    status,
    created_at
FROM retention_cleanup_logs
WHERE tenant_id = 'tenant-id'
ORDER BY created_at DESC
LIMIT 10;
```

## Testing

Run tests:

```bash
# Unit tests (no database required)
go test ./internal/retention/...

# Integration tests (requires TEST_DATABASE_URL)
TEST_DATABASE_URL="postgres://user:pass@localhost/test_db" go test ./internal/retention/...
```

## Performance Considerations

1. **Batch Size**: Larger batches are faster but hold locks longer (default: 1000)
2. **Run Interval**: More frequent runs process smaller amounts of data (default: 24h)
3. **Indexes**: Ensure indexes exist on `executions(tenant_id, created_at, status)`
4. **Off-Peak Hours**: Consider running cleanup during low-traffic periods

## Security

- **Row-Level Security**: All queries respect tenant isolation via RLS policies
- **Audit Trail**: All cleanup operations are logged for compliance
- **No Data Modification**: Only deletes old records, never modifies existing data
- **Configurable**: Tenants can disable cleanup or adjust retention periods

## Migration

Apply the migration to add retention support:

```bash
# Run migration 015_retention_policy.sql
psql -d gorax -f migrations/015_retention_policy.sql
```

This creates:
- `retention_cleanup_logs` table
- Indexes for efficient queries
- Default retention settings for existing tenants

## Troubleshooting

### Cleanup not running

Check:
1. `RETENTION_ENABLED=true` in environment
2. Scheduler is started: `scheduler.Start(ctx)`
3. Check logs for errors

### Too many/too few deletions

Adjust:
- `RETENTION_DEFAULT_DAYS`: Change default retention period
- Tenant-specific settings: Update `tenants.settings`
- `RETENTION_BATCH_SIZE`: Adjust batch size

### Performance issues

Try:
- Reduce batch size
- Increase run interval
- Run during off-peak hours
- Add database indexes

## Example: Setting Up Retention in Main Application

```go
// cmd/server/main.go

func setupRetention(db *sqlx.DB, cfg *config.Config) *retention.Scheduler {
    if !cfg.Retention.Enabled {
        return nil
    }

    repo := retention.NewRepository(db)

    serviceConfig := retention.Config{
        DefaultRetentionDays: cfg.Retention.DefaultRetentionDays,
        BatchSize:            cfg.Retention.BatchSize,
        EnableAuditLog:       cfg.Retention.EnableAuditLog,
    }

    service := retention.NewService(repo, slog.Default(), serviceConfig)

    interval, err := time.ParseDuration(cfg.Retention.RunInterval)
    if err != nil {
        interval = 24 * time.Hour // Default to daily
    }

    scheduler := retention.NewScheduler(service, slog.Default(), interval)

    return scheduler
}

func main() {
    // ... setup database ...

    // Setup retention scheduler
    retentionScheduler := setupRetention(db, cfg)
    if retentionScheduler != nil {
        ctx := context.Background()
        if err := retentionScheduler.Start(ctx); err != nil {
            log.Fatal("failed to start retention scheduler:", err)
        }
        defer retentionScheduler.Stop()
    }

    // ... start server ...
}
```

## API Integration (Optional)

To allow tenants to manage their retention policies via API, add handlers:

```go
// GET /api/v1/settings/retention
func (h *Handler) GetRetentionPolicy(c *gin.Context) {
    tenantID := getTenantID(c)

    policy, err := h.retentionService.GetRetentionPolicy(c.Request.Context(), tenantID)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, policy)
}

// PUT /api/v1/settings/retention
func (h *Handler) UpdateRetentionPolicy(c *gin.Context) {
    tenantID := getTenantID(c)

    var req struct {
        RetentionDays int  `json:"retention_days"`
        Enabled       bool `json:"enabled"`
    }

    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    err := h.retentionRepo.SetRetentionPolicy(c.Request.Context(), tenantID, req.RetentionDays, req.Enabled)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"message": "retention policy updated"})
}
```
