package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	// ErrNotFound is returned when an audit event is not found
	ErrNotFound = errors.New("audit event not found")
	// ErrInvalidFilter is returned when a query filter is invalid
	ErrInvalidFilter = errors.New("invalid query filter")
)

// Repository handles audit log database operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new audit repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateAuditEvent creates a single audit event
func (r *Repository) CreateAuditEvent(ctx context.Context, event *AuditEvent) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query := `
		INSERT INTO audit_events (
			id, tenant_id, user_id, user_email, category, event_type, action,
			resource_type, resource_id, resource_name, ip_address, user_agent,
			severity, status, error_message, metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
	`

	_, err = r.db.ExecContext(ctx, query,
		event.ID, event.TenantID, event.UserID, event.UserEmail,
		event.Category, event.EventType, event.Action,
		event.ResourceType, event.ResourceID, event.ResourceName,
		event.IPAddress, event.UserAgent, event.Severity, event.Status,
		event.ErrorMessage, metadataJSON, event.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}

	return nil
}

// CreateAuditEventBatch creates multiple audit events in a single transaction
func (r *Repository) CreateAuditEventBatch(ctx context.Context, events []*AuditEvent) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PreparexContext(ctx, `
		INSERT INTO audit_events (
			id, tenant_id, user_id, user_email, category, event_type, action,
			resource_type, resource_id, resource_name, ip_address, user_agent,
			severity, status, error_message, metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
	`)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, event := range events {
		if event.ID == "" {
			event.ID = uuid.New().String()
		}
		if event.CreatedAt.IsZero() {
			event.CreatedAt = now
		}

		metadataJSON, err := json.Marshal(event.Metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}

		_, err = stmt.ExecContext(ctx,
			event.ID, event.TenantID, event.UserID, event.UserEmail,
			event.Category, event.EventType, event.Action,
			event.ResourceType, event.ResourceID, event.ResourceName,
			event.IPAddress, event.UserAgent, event.Severity, event.Status,
			event.ErrorMessage, metadataJSON, event.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert audit event batch: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetAuditEvent retrieves a single audit event by ID
func (r *Repository) GetAuditEvent(ctx context.Context, tenantID, eventID string) (*AuditEvent, error) {
	query := `
		SELECT
			id, tenant_id, user_id, user_email, category, event_type, action,
			resource_type, resource_id, resource_name, ip_address, user_agent,
			severity, status, error_message, metadata, created_at
		FROM audit_events
		WHERE id = $1 AND tenant_id = $2
	`

	var event AuditEvent
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, eventID, tenantID).Scan(
		&event.ID, &event.TenantID, &event.UserID, &event.UserEmail,
		&event.Category, &event.EventType, &event.Action,
		&event.ResourceType, &event.ResourceID, &event.ResourceName,
		&event.IPAddress, &event.UserAgent, &event.Severity, &event.Status,
		&event.ErrorMessage, &metadataJSON, &event.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get audit event: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}

	return &event, nil
}

// QueryAuditEvents queries audit events with filters
func (r *Repository) QueryAuditEvents(ctx context.Context, filter QueryFilter) ([]AuditEvent, int, error) {
	if err := validateQueryFilter(filter); err != nil {
		return nil, 0, err
	}

	whereClause, args := buildWhereClause(filter)

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_events WHERE %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("count audit events: %w", err)
	}

	// Data query
	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortDirection := filter.SortDirection
	if sortDirection == "" {
		sortDirection = "DESC"
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	query := fmt.Sprintf(`
		SELECT
			id, tenant_id, user_id, user_email, category, event_type, action,
			resource_type, resource_id, resource_name, ip_address, user_agent,
			severity, status, error_message, metadata, created_at
		FROM audit_events
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortDirection, len(args)+1, len(args)+2)

	args = append(args, limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query audit events: %w", err)
	}
	defer rows.Close()

	var events []AuditEvent
	for rows.Next() {
		var event AuditEvent
		var metadataJSON []byte

		err := rows.Scan(
			&event.ID, &event.TenantID, &event.UserID, &event.UserEmail,
			&event.Category, &event.EventType, &event.Action,
			&event.ResourceType, &event.ResourceID, &event.ResourceName,
			&event.IPAddress, &event.UserAgent, &event.Severity, &event.Status,
			&event.ErrorMessage, &metadataJSON, &event.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan audit event: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
			return nil, 0, fmt.Errorf("unmarshal metadata: %w", err)
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate rows: %w", err)
	}

	return events, total, nil
}

// GetAuditStats retrieves aggregate statistics for audit logs
func (r *Repository) GetAuditStats(ctx context.Context, tenantID string, timeRange TimeRange) (*AuditStats, error) {
	stats := &AuditStats{
		EventsByCategory: make(map[Category]int),
		EventsBySeverity: make(map[Severity]int),
		EventsByStatus:   make(map[Status]int),
		TimeRange:        timeRange,
	}

	// Total events
	err := r.db.GetContext(ctx, &stats.TotalEvents, `
		SELECT COUNT(*) FROM audit_events
		WHERE tenant_id = $1 AND created_at >= $2 AND created_at <= $3
	`, tenantID, timeRange.StartDate, timeRange.EndDate)
	if err != nil {
		return nil, fmt.Errorf("count total events: %w", err)
	}

	// Events by category
	categoryRows, err := r.db.QueryContext(ctx, `
		SELECT category, COUNT(*) as count
		FROM audit_events
		WHERE tenant_id = $1 AND created_at >= $2 AND created_at <= $3
		GROUP BY category
	`, tenantID, timeRange.StartDate, timeRange.EndDate)
	if err != nil {
		return nil, fmt.Errorf("query category stats: %w", err)
	}
	defer categoryRows.Close()

	for categoryRows.Next() {
		var category Category
		var count int
		if err := categoryRows.Scan(&category, &count); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		stats.EventsByCategory[category] = count
	}

	// Events by severity
	severityRows, err := r.db.QueryContext(ctx, `
		SELECT severity, COUNT(*) as count
		FROM audit_events
		WHERE tenant_id = $1 AND created_at >= $2 AND created_at <= $3
		GROUP BY severity
	`, tenantID, timeRange.StartDate, timeRange.EndDate)
	if err != nil {
		return nil, fmt.Errorf("query severity stats: %w", err)
	}
	defer severityRows.Close()

	for severityRows.Next() {
		var severity Severity
		var count int
		if err := severityRows.Scan(&severity, &count); err != nil {
			return nil, fmt.Errorf("scan severity: %w", err)
		}
		stats.EventsBySeverity[severity] = count
	}

	// Events by status
	statusRows, err := r.db.QueryContext(ctx, `
		SELECT status, COUNT(*) as count
		FROM audit_events
		WHERE tenant_id = $1 AND created_at >= $2 AND created_at <= $3
		GROUP BY status
	`, tenantID, timeRange.StartDate, timeRange.EndDate)
	if err != nil {
		return nil, fmt.Errorf("query status stats: %w", err)
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var status Status
		var count int
		if err := statusRows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan status: %w", err)
		}
		stats.EventsByStatus[status] = count
	}

	// Critical and failed events counts
	stats.CriticalEvents = stats.EventsBySeverity[SeverityCritical]
	stats.FailedEvents = stats.EventsByStatus[StatusFailure]

	// Top users
	err = r.db.SelectContext(ctx, &stats.TopUsers, `
		SELECT user_id, user_email, COUNT(*) as event_count
		FROM audit_events
		WHERE tenant_id = $1 AND created_at >= $2 AND created_at <= $3
			AND user_id IS NOT NULL
		GROUP BY user_id, user_email
		ORDER BY event_count DESC
		LIMIT 10
	`, tenantID, timeRange.StartDate, timeRange.EndDate)
	if err != nil {
		return nil, fmt.Errorf("query top users: %w", err)
	}

	// Top actions
	err = r.db.SelectContext(ctx, &stats.TopActions, `
		SELECT action, COUNT(*) as count
		FROM audit_events
		WHERE tenant_id = $1 AND created_at >= $2 AND created_at <= $3
		GROUP BY action
		ORDER BY count DESC
		LIMIT 10
	`, tenantID, timeRange.StartDate, timeRange.EndDate)
	if err != nil {
		return nil, fmt.Errorf("query top actions: %w", err)
	}

	// Recent critical events
	err = r.db.SelectContext(ctx, &stats.RecentCritical, `
		SELECT
			id, tenant_id, user_id, user_email, category, event_type, action,
			resource_type, resource_id, resource_name, ip_address, user_agent,
			severity, status, error_message, metadata, created_at
		FROM audit_events
		WHERE tenant_id = $1 AND severity = $2 AND created_at >= $3 AND created_at <= $4
		ORDER BY created_at DESC
		LIMIT 20
	`, tenantID, SeverityCritical, timeRange.StartDate, timeRange.EndDate)
	if err != nil {
		return nil, fmt.Errorf("query recent critical events: %w", err)
	}

	return stats, nil
}

// GetRetentionPolicy retrieves the retention policy for a tenant
func (r *Repository) GetRetentionPolicy(ctx context.Context, tenantID string) (*RetentionPolicy, error) {
	query := `
		SELECT
			id, tenant_id, hot_retention_days, warm_retention_days, cold_retention_days,
			archive_enabled, archive_bucket, archive_path, purge_enabled,
			last_archive_at, last_purge_at, created_at, updated_at
		FROM audit_retention_policies
		WHERE tenant_id = $1
	`

	var policy RetentionPolicy
	err := r.db.GetContext(ctx, &policy, query, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get retention policy: %w", err)
	}

	return &policy, nil
}

// UpdateRetentionPolicy updates the retention policy for a tenant
func (r *Repository) UpdateRetentionPolicy(ctx context.Context, policy *RetentionPolicy) error {
	query := `
		UPDATE audit_retention_policies
		SET hot_retention_days = $1, warm_retention_days = $2, cold_retention_days = $3,
			archive_enabled = $4, archive_bucket = $5, archive_path = $6, purge_enabled = $7
		WHERE tenant_id = $8
	`

	result, err := r.db.ExecContext(ctx, query,
		policy.HotRetentionDays, policy.WarmRetentionDays, policy.ColdRetentionDays,
		policy.ArchiveEnabled, policy.ArchiveBucket, policy.ArchivePath, policy.PurgeEnabled,
		policy.TenantID,
	)
	if err != nil {
		return fmt.Errorf("update retention policy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteOldAuditEvents deletes audit events older than the cutoff date
func (r *Repository) DeleteOldAuditEvents(ctx context.Context, tenantID string, cutoffDate time.Time) (int64, error) {
	query := `
		DELETE FROM audit_events
		WHERE tenant_id = $1 AND created_at < $2
	`

	result, err := r.db.ExecContext(ctx, query, tenantID, cutoffDate)
	if err != nil {
		return 0, fmt.Errorf("delete old audit events: %w", err)
	}

	deletedCount, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("get rows affected: %w", err)
	}

	return deletedCount, nil
}

// buildWhereClause builds the WHERE clause and arguments for audit event queries
func buildWhereClause(filter QueryFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argPos := 1

	// Tenant ID is required
	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argPos))
	args = append(args, filter.TenantID)
	argPos++

	if filter.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argPos))
		args = append(args, filter.UserID)
		argPos++
	}

	if filter.UserEmail != "" {
		conditions = append(conditions, fmt.Sprintf("user_email = $%d", argPos))
		args = append(args, filter.UserEmail)
		argPos++
	}

	if len(filter.Categories) > 0 {
		conditions = append(conditions, fmt.Sprintf("category = ANY($%d)", argPos))
		args = append(args, pq.Array(filter.Categories))
		argPos++
	}

	if len(filter.EventTypes) > 0 {
		conditions = append(conditions, fmt.Sprintf("event_type = ANY($%d)", argPos))
		args = append(args, pq.Array(filter.EventTypes))
		argPos++
	}

	if len(filter.Actions) > 0 {
		conditions = append(conditions, fmt.Sprintf("action = ANY($%d)", argPos))
		args = append(args, pq.Array(filter.Actions))
		argPos++
	}

	if filter.ResourceType != "" {
		conditions = append(conditions, fmt.Sprintf("resource_type = $%d", argPos))
		args = append(args, filter.ResourceType)
		argPos++
	}

	if filter.ResourceID != "" {
		conditions = append(conditions, fmt.Sprintf("resource_id = $%d", argPos))
		args = append(args, filter.ResourceID)
		argPos++
	}

	if filter.IPAddress != "" {
		conditions = append(conditions, fmt.Sprintf("ip_address = $%d", argPos))
		args = append(args, filter.IPAddress)
		argPos++
	}

	if len(filter.Severities) > 0 {
		conditions = append(conditions, fmt.Sprintf("severity = ANY($%d)", argPos))
		args = append(args, pq.Array(filter.Severities))
		argPos++
	}

	if len(filter.Statuses) > 0 {
		conditions = append(conditions, fmt.Sprintf("status = ANY($%d)", argPos))
		args = append(args, pq.Array(filter.Statuses))
		argPos++
	}

	if !filter.StartDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argPos))
		args = append(args, filter.StartDate)
		argPos++
	}

	if !filter.EndDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argPos))
		args = append(args, filter.EndDate)
		_ = argPos // Final argPos value not used but maintains pattern for future conditions
	}

	whereClause := strings.Join(conditions, " AND ")
	return whereClause, args
}

// validateQueryFilter validates the query filter
func validateQueryFilter(filter QueryFilter) error {
	if filter.TenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidFilter)
	}

	if filter.SortBy != "" {
		validSortFields := map[string]bool{
			"created_at":    true,
			"category":      true,
			"event_type":    true,
			"severity":      true,
			"status":        true,
			"action":        true,
			"resource_type": true,
		}
		if !validSortFields[filter.SortBy] {
			return fmt.Errorf("%w: invalid sort_by field", ErrInvalidFilter)
		}
	}

	if filter.SortDirection != "" {
		if filter.SortDirection != "ASC" && filter.SortDirection != "DESC" {
			return fmt.Errorf("%w: sort_direction must be ASC or DESC", ErrInvalidFilter)
		}
	}

	return nil
}
