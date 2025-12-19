# Schedule Triggers

This package implements scheduled workflow executions using cron expressions.

## Features

- **Cron Expression Support**: Standard cron syntax with optional seconds field
- **Timezone Aware**: Schedule executions in any timezone
- **Persistent Storage**: Schedules stored in PostgreSQL with tenant isolation
- **Automatic Next Run Calculation**: Automatically calculates and updates next run times
- **Missed Execution Handling**: Runs immediately when a schedule is due
- **Graceful Shutdown**: Properly handles shutdown signals to avoid lost executions

## Architecture

### Components

1. **Model** (`model.go`): Schedule data structures and types
2. **Repository** (`repository.go`): Database operations with tenant isolation
3. **Service** (`service.go`): Business logic and cron validation
4. **Scheduler** (`scheduler.go`): Background service that checks and executes due schedules
5. **Handler** (`handlers/schedule.go`): REST API endpoints

### Database Schema

```sql
CREATE TABLE schedules (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    workflow_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    cron_expression VARCHAR(255) NOT NULL,
    timezone VARCHAR(100) NOT NULL DEFAULT 'UTC',
    enabled BOOLEAN NOT NULL DEFAULT true,
    next_run_at TIMESTAMPTZ,
    last_run_at TIMESTAMPTZ,
    last_execution_id UUID,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
```

## API Endpoints

### Create Schedule

```
POST /api/v1/workflows/{workflowID}/schedules
```

**Request Body:**
```json
{
  "name": "Daily Report",
  "cron_expression": "0 9 * * *",
  "timezone": "America/New_York",
  "enabled": true
}
```

### List Schedules for Workflow

```
GET /api/v1/workflows/{workflowID}/schedules?limit=20&offset=0
```

### List All Schedules

```
GET /api/v1/schedules?limit=20&offset=0
```

### Get Schedule

```
GET /api/v1/schedules/{scheduleID}
```

### Update Schedule

```
PUT /api/v1/schedules/{scheduleID}
```

**Request Body:**
```json
{
  "name": "Updated Name",
  "cron_expression": "0 10 * * *",
  "timezone": "UTC",
  "enabled": false
}
```

### Delete Schedule

```
DELETE /api/v1/schedules/{scheduleID}
```

### Parse Cron Expression

```
POST /api/v1/schedules/parse-cron
```

**Request Body:**
```json
{
  "cron_expression": "0 */2 * * *",
  "timezone": "UTC"
}
```

**Response:**
```json
{
  "valid": true,
  "next_run": "2024-01-15T14:00:00Z"
}
```

## Cron Expression Format

The scheduler supports standard cron expressions with an optional seconds field:

### Standard Format (5 fields)
```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of week (0 - 6) (Sunday to Saturday)
│ │ │ │ │
* * * * *
```

### Extended Format (6 fields with seconds)
```
┌───────────── second (0 - 59)
│ ┌───────────── minute (0 - 59)
│ │ ┌───────────── hour (0 - 23)
│ │ │ ┌───────────── day of month (1 - 31)
│ │ │ │ ┌───────────── month (1 - 12)
│ │ │ │ │ ┌───────────── day of week (0 - 6)
│ │ │ │ │ │
* * * * * *
```

### Special Characters

- `*` - any value
- `,` - value list separator (e.g., `1,3,5`)
- `-` - range of values (e.g., `1-5`)
- `/` - step values (e.g., `*/15` for every 15 units)

### Predefined Schedules

- `@yearly` or `@annually` - Run once a year at midnight on January 1st
- `@monthly` - Run once a month at midnight on the first day
- `@weekly` - Run once a week at midnight on Sunday
- `@daily` or `@midnight` - Run once a day at midnight
- `@hourly` - Run once an hour at the beginning of the hour

### Examples

```
0 9 * * *           # Every day at 9:00 AM
0 */2 * * *         # Every 2 hours
30 8 * * 1-5        # Weekdays at 8:30 AM
0 0 1 * *           # First day of every month at midnight
0 12 * * 0          # Every Sunday at noon
*/30 * * * *        # Every 30 minutes
0 0,12 * * *        # Twice a day at midnight and noon
0 9-17 * * 1-5      # Every hour from 9 AM to 5 PM on weekdays
@daily              # Once per day at midnight
```

## Timezone Support

Schedules support any valid IANA timezone identifier:

- `UTC`
- `America/New_York`
- `America/Los_Angeles`
- `Europe/London`
- `Asia/Tokyo`
- etc.

If no timezone is specified, UTC is used by default.

## Scheduler Configuration

The scheduler runs in the worker process and checks for due schedules every 30 seconds by default.

### Configuration Options

- **Check Interval**: How often to check for due schedules (default: 30 seconds)
- **Batch Size**: Maximum number of schedules to process per check (default: 100)
- **Concurrency**: Number of schedules to execute concurrently (default: 10)

## Integration

### In Worker Process

The scheduler is integrated into the worker process (`cmd/worker/main.go`):

```go
// Initialize scheduler
scheduler := schedule.NewScheduler(scheduleService, executorAdapter, logger)

// Start scheduler
go scheduler.Start(ctx)

// Stop on shutdown
scheduler.Stop()
scheduler.Wait()
```

### With Workflow Service

The schedule service requires a workflow getter to validate workflows:

```go
// Create adapter
workflowGetter := &workflowServiceAdapter{
    workflowService: workflowService,
}

// Set on schedule service
scheduleService.SetWorkflowService(workflowGetter)
```

## Error Handling

### Missed Executions

When a schedule is due, it runs immediately. The scheduler:

1. Queries for all enabled schedules where `next_run_at <= NOW()`
2. Executes each schedule
3. Updates `last_run_at` and calculates new `next_run_at`
4. If execution fails, still updates the schedule to avoid repeated failures

### Workflow Validation

When creating a schedule:

1. Validates cron expression syntax
2. Validates timezone identifier
3. Verifies workflow exists and is accessible by tenant
4. Calculates initial `next_run_at` if enabled

## Testing

Run tests with:

```bash
go test ./internal/schedule/... -v
```

### Test Coverage

- Cron expression validation
- Next run time calculation
- Timezone handling
- Scheduler start/stop
- Schedule execution
- Disabled schedule handling
- Multiple concurrent schedules

## Security

### Tenant Isolation

All schedule operations are tenant-scoped:

- Database queries include tenant_id
- Row-level security policies enforce isolation
- API handlers validate tenant context

### Access Control

- Schedule creation requires authentication
- Only tenant members can manage their schedules
- Workflow access is validated before schedule creation

## Performance Considerations

### Database Indexes

The following indexes optimize schedule queries:

```sql
CREATE INDEX idx_schedules_next_run_at ON schedules(next_run_at) WHERE enabled = true;
CREATE INDEX idx_schedules_workflow_id ON schedules(workflow_id);
CREATE INDEX idx_schedules_tenant_id ON schedules(tenant_id);
```

### Query Optimization

- Due schedules query uses indexed `next_run_at` column
- WHERE clause filters on `enabled = true` to use partial index
- LIMIT 100 prevents excessive memory usage

### Concurrency

- Scheduler executes up to 10 schedules concurrently
- Semaphore pattern prevents resource exhaustion
- Each execution runs in its own goroutine

## Future Enhancements

- [ ] Schedule execution history tracking
- [ ] Schedule pause/resume functionality
- [ ] Schedule dry-run mode for testing
- [ ] Email notifications for schedule failures
- [ ] Schedule overlap prevention
- [ ] Advanced scheduling rules (skip holidays, business days only, etc.)
- [ ] Schedule templates
- [ ] Bulk schedule operations
