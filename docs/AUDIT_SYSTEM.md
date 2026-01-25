# Audit Logging System Documentation

## Overview

The Gorax audit logging system provides comprehensive, compliance-focused auditing for all user actions, system events, and data access. The system is designed to meet SOC 2, HIPAA, and GDPR compliance requirements while maintaining high performance through async batching.

## Architecture

### Components

1. **Database Schema** (`migrations/031_audit_logs.sql`)
   - `audit_events`: Main audit log table with comprehensive event tracking
   - `audit_retention_policies`: Configurable retention policies per tenant
   - `audit_log_integrity`: Daily integrity hashes for tamper detection

2. **Domain Model** (`internal/audit/model.go`)
   - Event categories, types, severities, and statuses
   - Audit event structure with metadata support
   - Retention policy configuration
   - Query filters and statistics

3. **Repository** (`internal/audit/repository.go`)
   - Database operations for audit events
   - Batch insert support for performance
   - Complex query filtering
   - Aggregate statistics

4. **Service** (`internal/audit/service.go`)
   - Business logic layer
   - Async event buffering and batching
   - Automatic flush on timer or buffer full
   - Graceful shutdown with event preservation

## Event Categories

The system categorizes events into the following types:

- **authentication**: Login, logout, session management
- **authorization**: Permission changes, access control
- **data_access**: Reading sensitive data, exports
- **configuration**: System and tenant configuration changes
- **workflow**: Workflow creation, execution, modification
- **integration**: External system connections
- **credential**: Credential access and management
- **user_management**: User creation, updates, deletions
- **system**: System-level events

## Event Types

- **create**: Resource creation
- **read**: Data access
- **update**: Resource modification
- **delete**: Resource deletion
- **execute**: Workflow or action execution
- **login**: User authentication
- **logout**: User session termination
- **permission_change**: Authorization modifications
- **export**: Data export operations
- **import**: Data import operations
- **access**: Credential or sensitive data access
- **configure**: Configuration changes

## Severity Levels

- **info**: Normal operations
- **warning**: Potential issues
- **error**: Operation failures
- **critical**: Security events, critical failures

## Usage

### Basic Event Logging

```go
import "github.com/gorax/gorax/internal/audit"

// Create service
auditService := audit.NewService(repo, 100, 5*time.Second)
defer auditService.Close()

// Log event asynchronously (non-blocking)
event := &audit.AuditEvent{
    TenantID:     "tenant-1",
    UserID:       "user-1",
    UserEmail:    "user@example.com",
    Category:     audit.CategoryWorkflow,
    EventType:    audit.EventTypeExecute,
    Action:       "workflow.executed",
    ResourceType: "workflow",
    ResourceID:   "wf-123",
    ResourceName: "User Onboarding",
    IPAddress:    "192.168.1.1",
    UserAgent:    "Mozilla/5.0...",
    Severity:     audit.SeverityInfo,
    Status:       audit.StatusSuccess,
    Metadata: map[string]interface{}{
        "execution_id": "exec-123",
        "duration_ms":  1250,
    },
}

err := auditService.LogEvent(ctx, event)

// Log event synchronously (blocking, for critical events)
err := auditService.LogEventSync(ctx, event)

// Log multiple events in batch
events := []*audit.AuditEvent{event1, event2, event3}
err := auditService.LogEventBatch(ctx, events)
```

### Querying Audit Events

```go
// Query with filters
filter := audit.QueryFilter{
    TenantID:      "tenant-1",
    UserID:        "user-1",
    Categories:    []audit.Category{audit.CategoryWorkflow},
    Severities:    []audit.Severity{audit.SeverityCritical},
    StartDate:     time.Now().Add(-24 * time.Hour),
    EndDate:       time.Now(),
    Limit:         100,
    Offset:        0,
    SortBy:        "created_at",
    SortDirection: "DESC",
}

events, total, err := auditService.QueryAuditEvents(ctx, filter)
```

### Statistics and Analytics

```go
// Get aggregate statistics
timeRange := audit.TimeRange{
    StartDate: time.Now().Add(-30 * 24 * time.Hour),
    EndDate:   time.Now(),
}

stats, err := auditService.GetAuditStats(ctx, tenantID, timeRange)

// Access statistics
fmt.Printf("Total events: %d\n", stats.TotalEvents)
fmt.Printf("Critical events: %d\n", stats.CriticalEvents)
fmt.Printf("Failed events: %d\n", stats.FailedEvents)

// Events by category
for category, count := range stats.EventsByCategory {
    fmt.Printf("%s: %d\n", category, count)
}

// Top active users
for _, user := range stats.TopUsers {
    fmt.Printf("%s: %d events\n", user.UserEmail, user.EventCount)
}
```

### Retention Policy Management

```go
// Get retention policy
policy, err := auditService.GetRetentionPolicy(ctx, tenantID)

// Update retention policy
policy.HotRetentionDays = 180
policy.WarmRetentionDays = 365
policy.ColdRetentionDays = 2555
policy.ArchiveEnabled = true
policy.ArchiveBucket = "audit-archive"
policy.ArchivePath = "tenant-1/audit"

err = auditService.UpdateRetentionPolicy(ctx, policy)

// Cleanup old logs
deletedCount, err := auditService.CleanupOldLogs(ctx, tenantID)
```

## Performance Considerations

### Async Batching

The service uses async batching to minimize database load:

- Events are buffered in memory
- Automatic flush when buffer is full (default: 100 events)
- Periodic flush on timer (default: 5 seconds)
- Graceful shutdown preserves all buffered events

### Buffer Sizing

Configure buffer size and flush interval based on your load:

```go
// High-volume: larger buffer, longer timer
service := audit.NewService(repo, 1000, 30*time.Second)

// Low-latency: smaller buffer, shorter timer
service := audit.NewService(repo, 50, 1*time.Second)
```

### Database Indexes

The migration creates comprehensive indexes for common query patterns:

- Tenant + created_at (time-series queries)
- Tenant + category (category filtering)
- Tenant + user (user activity)
- Tenant + severity (critical event queries)
- IP address (security monitoring)
- Metadata GIN index (JSON queries)

## Compliance Features

### Tamper-Proof Logging

- Append-only table (no updates or deletes via application)
- Daily integrity hashes in `audit_log_integrity` table
- Cryptographic verification of log completeness

### Data Retention

Three-tier retention model:

1. **Hot storage** (default 90 days): Full-speed queries in main database
2. **Warm storage** (default 365 days): Compressed storage, slower queries
3. **Cold storage** (default 7 years): Archive to S3/object storage

### Compliance Reports

Generate reports for:

- **SOC 2**: User activity, access control changes, system configuration
- **HIPAA**: PHI access logs, user authentication, data exports
- **GDPR**: Personal data access, user consent, data deletion requests

## Security Best Practices

### What to Log

**DO log:**
- All authentication events (login, logout, session creation)
- Authorization changes (role assignments, permission modifications)
- Data access (reading PII, PHI, financial data)
- Data modifications (create, update, delete)
- Configuration changes (system settings, integrations)
- Security events (failed logins, permission denials)
- Data exports (CSV, JSON, API access)

**DO NOT log:**
- Passwords or credentials (even encrypted)
- API keys or tokens
- PII in metadata without masking
- Request bodies containing sensitive data

### Example Security Events

```go
// Failed login attempt
auditService.LogEvent(ctx, &audit.AuditEvent{
    TenantID:  tenantID,
    UserEmail: attemptedEmail,
    Category:  audit.CategoryAuthentication,
    EventType: audit.EventTypeLogin,
    Action:    "login.failed",
    Severity:  audit.SeverityWarning,
    Status:    audit.StatusFailure,
    ErrorMessage: "Invalid credentials",
    IPAddress: ipAddress,
})

// Permission change
auditService.LogEvent(ctx, &audit.AuditEvent{
    TenantID:     tenantID,
    UserID:       adminUserID,
    Category:     audit.CategoryAuthorization,
    EventType:    audit.EventTypePermissionChange,
    Action:       "role.assigned",
    ResourceType: "user",
    ResourceID:   targetUserID,
    Severity:     audit.SeverityCritical,
    Status:       audit.StatusSuccess,
    Metadata: map[string]interface{}{
        "role": "admin",
        "previous_role": "user",
    },
})

// Data export
auditService.LogEvent(ctx, &audit.AuditEvent{
    TenantID:     tenantID,
    UserID:       userID,
    Category:     audit.CategoryDataAccess,
    EventType:    audit.EventTypeExport,
    Action:       "data.exported",
    ResourceType: "audit_logs",
    Severity:     audit.SeverityWarning,
    Status:       audit.StatusSuccess,
    Metadata: map[string]interface{}{
        "format": "csv",
        "record_count": 1000,
    },
})
```

## Database Schema

### audit_events Table

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Unique event ID |
| tenant_id | UUID | Tenant identifier |
| user_id | VARCHAR(255) | User who performed action |
| user_email | VARCHAR(255) | User email |
| category | VARCHAR(50) | Event category |
| event_type | VARCHAR(50) | Type of event |
| action | VARCHAR(255) | Specific action taken |
| resource_type | VARCHAR(100) | Type of resource affected |
| resource_id | VARCHAR(255) | ID of resource |
| resource_name | VARCHAR(255) | Name of resource |
| ip_address | INET | Client IP address |
| user_agent | TEXT | Client user agent |
| severity | VARCHAR(20) | Event severity |
| status | VARCHAR(20) | Event outcome |
| error_message | TEXT | Error details if failed |
| metadata | JSONB | Additional context |
| created_at | TIMESTAMP | Event timestamp |

### Indexes

- Primary key on `id`
- Index on `tenant_id` (required for all queries)
- Composite index on `(tenant_id, created_at)` for time-series
- Composite index on `(tenant_id, category)` for filtering
- Composite index on `(tenant_id, severity)` for critical event queries
- Partial index on critical events
- Partial index on failed events
- GIN index on `metadata` for JSON queries

## Testing

### Unit Tests

All components have comprehensive unit tests:

```bash
# Run all audit tests
go test ./internal/audit/...

# Run with coverage
go test ./internal/audit/... -cover

# Run with verbose output
go test ./internal/audit/... -v
```

### Integration Tests

Integration tests require a database connection (skipped by default):

```bash
# Set up test database
export TEST_DB_URL="postgres://user:pass@localhost/gorax_test"

# Run integration tests
go test ./internal/audit/... -tags=integration
```

## Monitoring and Alerting

### Key Metrics

Monitor these metrics for audit system health:

1. **Event ingestion rate**: Events per second
2. **Buffer utilization**: Percentage of buffer full
3. **Flush latency**: Time to flush batch to database
4. **Query performance**: P50, P95, P99 latencies
5. **Failed inserts**: Count of database errors

### Alert Conditions

Set up alerts for:

- Multiple failed login attempts (5+ in 5 minutes)
- Critical events (immediate notification)
- Permission changes (admin role assignments)
- Large data exports (> 10,000 records)
- Unusual activity patterns (time-series anomalies)

## Migration and Deployment

### Running the Migration

```bash
# Apply migration
psql -U postgres -d gorax -f migrations/031_audit_logs.sql

# Verify tables created
psql -U postgres -d gorax -c "\dt audit_*"
```

### Rollback

To remove the audit system (not recommended for production):

```sql
DROP TABLE IF EXISTS audit_events CASCADE;
DROP TABLE IF EXISTS audit_retention_policies CASCADE;
DROP TABLE IF EXISTS audit_log_integrity CASCADE;
DROP FUNCTION IF EXISTS update_audit_retention_policy_updated_at CASCADE;
```

## Future Enhancements

### Planned Features

1. **Log archival service**: Automatic archival to S3/GCS
2. **Compliance report generation**: SOC 2, HIPAA, GDPR reports
3. **Anomaly detection**: ML-based unusual activity detection
4. **Real-time alerting**: WebSocket notifications for critical events
5. **Log export**: CSV and JSON export functionality
6. **Audit log viewer UI**: React component for browsing logs
7. **Audit dashboard**: Statistics and visualizations

### API Endpoints (To Be Implemented)

- `GET /api/v1/audit/events`: Query audit logs
- `GET /api/v1/audit/events/:id`: Get event details
- `GET /api/v1/audit/export`: Export logs
- `GET /api/v1/audit/stats`: Audit statistics
- `POST /api/v1/audit/retention-policy`: Update retention
- `GET /api/v1/audit/compliance/report`: Generate compliance report

## Troubleshooting

### Common Issues

**Issue**: Events not appearing in database

**Solution**: Check buffer flush timing. Force flush with `service.Flush()` or use `LogEventSync()` for immediate writes.

---

**Issue**: Slow audit queries

**Solution**: Ensure indexes are present. Use `EXPLAIN ANALYZE` to check query plans. Consider increasing `hot_retention_days` to keep more data in fast storage.

---

**Issue**: High database load

**Solution**: Increase buffer size and flush interval. Consider using read replicas for queries.

---

**Issue**: Missing events after crash

**Solution**: Use `LogEventSync()` for critical events that must not be lost. Ensure graceful shutdown calls `service.Close()`.

## Support and Contributing

For questions or issues with the audit system:

1. Check this documentation
2. Review test files for usage examples
3. Check database logs for errors
4. Open an issue with reproduction steps

When contributing:

1. Follow TDD: write tests first
2. Maintain cognitive complexity < 15
3. Use meaningful variable names
4. Handle all errors explicitly
5. Add appropriate audit logging to new features
