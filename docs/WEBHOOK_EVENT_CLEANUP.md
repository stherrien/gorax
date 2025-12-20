# Webhook Event Cleanup

## Overview

The webhook event cleanup feature automatically removes old webhook event records from the database based on a configurable retention period. This helps manage database storage and maintain optimal performance by preventing unbounded growth of audit logs.

## Features

- **Configurable Retention Period**: Set how long webhook events should be retained (default: 30 days)
- **Batch Processing**: Deletes events in configurable batches to prevent database overload
- **Scheduled Execution**: Runs on a cron schedule (default: daily at midnight)
- **Graceful Handling**: Respects context cancellation for clean shutdowns
- **Comprehensive Logging**: Logs cleanup statistics including events deleted, batches processed, and execution time
- **Zero Downtime**: Runs alongside the worker without impacting workflow execution

## Configuration

Configure the cleanup service using environment variables:

```bash
# Enable/disable cleanup (default: true)
CLEANUP_ENABLED=true

# Number of days to retain webhook events (default: 30)
CLEANUP_RETENTION_DAYS=30

# Number of events to delete per batch (default: 1000)
CLEANUP_BATCH_SIZE=1000

# Cron schedule for cleanup (default: "0 0 * * *" - daily at midnight)
CLEANUP_SCHEDULE=0 0 * * *
```

### Cron Schedule Format

The cleanup schedule uses standard cron format:

```
* * * * *
│ │ │ │ │
│ │ │ │ └─── Day of week (0-6, Sunday=0)
│ │ │ └───── Month (1-12)
│ │ └─────── Day of month (1-31)
│ └───────── Hour (0-23)
└─────────── Minute (0-59)
```

**Examples**:
- `0 0 * * *` - Daily at midnight
- `0 2 * * *` - Daily at 2 AM
- `0 0 * * 0` - Weekly on Sunday at midnight
- `0 */6 * * *` - Every 6 hours

## Architecture

### Components

1. **CleanupService** (`internal/webhook/cleanup.go`)
   - Core business logic for deleting old events
   - Handles batch processing and error management
   - Returns detailed statistics about the cleanup operation

2. **CleanupScheduler** (`internal/webhook/cleanup_scheduler.go`)
   - Manages scheduled execution using cron
   - Integrates with the worker lifecycle
   - Logs cleanup results and warnings

3. **Repository Method** (`internal/webhook/repository.go`)
   - `DeleteOldEvents()` - Efficiently deletes events older than retention period
   - Uses batching to prevent database locks on large deletions

### Database Query

The cleanup uses an efficient SQL query that:
1. Identifies events older than the retention cutoff
2. Deletes them in batches using a subquery with LIMIT
3. Returns the count of deleted rows

```sql
DELETE FROM webhook_events
WHERE id IN (
    SELECT id FROM webhook_events
    WHERE created_at < $1
    ORDER BY created_at
    LIMIT $2
)
```

## Integration

The cleanup service is integrated into the worker process:

```go
// In cmd/worker/main.go
if cfg.Cleanup.Enabled {
    webhookRepo := webhook.NewRepository(db)
    retentionPeriod := time.Duration(cfg.Cleanup.RetentionDays) * 24 * time.Hour
    cleanupService := webhook.NewCleanupService(webhookRepo, cfg.Cleanup.BatchSize, retentionPeriod)
    cleanupScheduler = webhook.NewCleanupScheduler(cleanupService, cfg.Cleanup.Schedule, logger)
}
```

## Monitoring

### Log Output

Successful cleanup:
```json
{
  "level": "info",
  "msg": "cleanup completed",
  "total_deleted": 1500,
  "batches_processed": 2,
  "duration_ms": 245,
  "retention_period": "720h0m0s"
}
```

Failed cleanup:
```json
{
  "level": "error",
  "msg": "cleanup failed",
  "error": "delete batch failed: ...",
  "total_deleted": 1000,
  "batches_processed": 1,
  "duration_ms": 120
}
```

Long-running cleanup warning:
```json
{
  "level": "warn",
  "msg": "cleanup took longer than expected",
  "duration": "6m30s"
}
```

### Metrics to Monitor

- **total_deleted**: Number of events removed
- **batches_processed**: Number of batches executed
- **duration_ms**: Total execution time in milliseconds
- **retention_period**: Configured retention period

## Performance Considerations

### Batch Size Selection

Choose batch size based on your database performance:

- **Small databases (< 1M events)**: 1000-5000 events/batch
- **Medium databases (1-10M events)**: 5000-10000 events/batch
- **Large databases (> 10M events)**: 10000-50000 events/batch

**Trade-offs**:
- Smaller batches: Lower database impact, longer total cleanup time
- Larger batches: Faster cleanup, higher momentary database load

### Scheduling Recommendations

- Run during low-traffic periods (e.g., 2-4 AM)
- Avoid running during peak workflow execution times
- Consider weekly execution for low-volume systems
- For high-volume systems, consider daily or even multiple times per day

## Testing

### Unit Tests

Tests are located in `internal/webhook/cleanup_test.go`:

```bash
go test ./internal/webhook/cleanup_test.go ./internal/webhook/cleanup.go -v
```

Test coverage includes:
- Successful multi-batch cleanup
- Empty database (no events to delete)
- Context cancellation handling
- Error handling
- Custom retention periods
- Large batch processing

### Integration Tests

Repository tests in `internal/webhook/repository_test.go`:

```bash
# Requires DB_TEST_URL environment variable
export DB_TEST_URL="postgres://..."
go test ./internal/webhook/repository_test.go
```

## Troubleshooting

### Cleanup Not Running

1. Check if cleanup is enabled:
   ```bash
   echo $CLEANUP_ENABLED
   ```

2. Verify cron schedule is valid:
   ```bash
   # Test cron expression at https://crontab.guru/
   ```

3. Check worker logs for initialization errors

### Cleanup Taking Too Long

1. Reduce batch size to decrease per-batch execution time
2. Add database indexes on `webhook_events.created_at`
3. Consider running cleanup more frequently with shorter retention

### High Database Load

1. Increase interval between cleanup runs
2. Reduce batch size
3. Run during off-peak hours only

## Future Enhancements

Potential improvements for future versions:

1. **Per-Tenant Retention Policies**
   - Allow different retention periods per tenant
   - Support tenant-specific cleanup schedules

2. **Selective Cleanup**
   - Retain events based on status (e.g., keep failed events longer)
   - Preserve events linked to active executions

3. **Archive Before Delete**
   - Export old events to S3 before deletion
   - Provide API for archived event retrieval

4. **Dynamic Batch Sizing**
   - Adjust batch size based on database load
   - Implement adaptive cleanup speed

5. **Cleanup Metrics API**
   - Expose cleanup statistics via REST API
   - Add Prometheus metrics for monitoring

## See Also

- [Webhook Requirements](./WEBHOOK_REQUIREMENTS.md) - Complete webhook system documentation
- [Database Migrations](../migrations/002_webhook_events.sql) - Webhook events table schema
- [Configuration Guide](../README.md) - Full configuration reference
