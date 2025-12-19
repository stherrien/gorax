package workflow

import (
	"context"
	"fmt"
	"time"
)

// ExecutionTrend represents execution counts grouped by time period
type ExecutionTrend struct {
	Date    string `json:"date" db:"date"`
	Count   int    `json:"count" db:"count"`
	Success int    `json:"success" db:"success"`
	Failed  int    `json:"failed" db:"failed"`
}

// DurationStats represents duration statistics for a workflow
type DurationStats struct {
	WorkflowID   string  `json:"workflowId" db:"workflow_id"`
	WorkflowName string  `json:"workflowName" db:"workflow_name"`
	AvgDuration  float64 `json:"avgDuration" db:"avg_duration"`
	P50Duration  float64 `json:"p50Duration" db:"p50_duration"`
	P90Duration  float64 `json:"p90Duration" db:"p90_duration"`
	P99Duration  float64 `json:"p99Duration" db:"p99_duration"`
	TotalRuns    int     `json:"totalRuns" db:"total_runs"`
}

// TopFailure represents a workflow with high failure count
type TopFailure struct {
	WorkflowID    string     `json:"workflowId" db:"workflow_id"`
	WorkflowName  string     `json:"workflowName" db:"workflow_name"`
	FailureCount  int        `json:"failureCount" db:"failure_count"`
	LastFailedAt  *time.Time `json:"lastFailedAt" db:"last_failed_at"`
	ErrorPreview  *string    `json:"errorPreview,omitempty" db:"error_preview"`
}

// TriggerTypeBreakdown represents execution count by trigger type
type TriggerTypeBreakdown struct {
	TriggerType string  `json:"triggerType" db:"trigger_type"`
	Count       int     `json:"count" db:"count"`
	Percentage  float64 `json:"percentage" db:"percentage"`
}

// GetExecutionTrends returns execution counts grouped by time period
func (r *Repository) GetExecutionTrends(ctx context.Context, tenantID string, startDate, endDate time.Time, groupBy string) ([]ExecutionTrend, error) {
	var dateFormat string
	var truncFunc string

	switch groupBy {
	case "hour":
		dateFormat = "YYYY-MM-DD HH24:00"
		truncFunc = "hour"
	case "day":
		dateFormat = "YYYY-MM-DD"
		truncFunc = "day"
	default:
		return nil, fmt.Errorf("invalid groupBy value: %s (must be 'hour' or 'day')", groupBy)
	}

	query := fmt.Sprintf(`
		SELECT
			TO_CHAR(DATE_TRUNC('%s', created_at), '%s') as date,
			COUNT(*) as count,
			COUNT(*) FILTER (WHERE status = 'completed') as success,
			COUNT(*) FILTER (WHERE status = 'failed') as failed
		FROM executions
		WHERE tenant_id = $1
			AND created_at >= $2
			AND created_at < $3
		GROUP BY DATE_TRUNC('%s', created_at)
		ORDER BY DATE_TRUNC('%s', created_at) ASC
	`, truncFunc, dateFormat, truncFunc, truncFunc)

	var trends []ExecutionTrend
	err := r.db.SelectContext(ctx, &trends, query, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("get execution trends: %w", err)
	}

	return trends, nil
}

// GetDurationStats returns duration statistics grouped by workflow
func (r *Repository) GetDurationStats(ctx context.Context, tenantID string, startDate, endDate time.Time) ([]DurationStats, error) {
	query := `
		SELECT
			e.workflow_id,
			w.name as workflow_name,
			AVG(EXTRACT(EPOCH FROM (e.completed_at - e.started_at)) * 1000) as avg_duration,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (e.completed_at - e.started_at)) * 1000) as p50_duration,
			PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (e.completed_at - e.started_at)) * 1000) as p90_duration,
			PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (e.completed_at - e.started_at)) * 1000) as p99_duration,
			COUNT(*) as total_runs
		FROM executions e
		INNER JOIN workflows w ON e.workflow_id = w.id
		WHERE e.tenant_id = $1
			AND e.created_at >= $2
			AND e.created_at < $3
			AND e.started_at IS NOT NULL
			AND e.completed_at IS NOT NULL
			AND e.status IN ('completed', 'failed')
		GROUP BY e.workflow_id, w.name
		ORDER BY total_runs DESC
	`

	var stats []DurationStats
	err := r.db.SelectContext(ctx, &stats, query, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("get duration stats: %w", err)
	}

	return stats, nil
}

// GetTopFailures returns workflows with the most failures
func (r *Repository) GetTopFailures(ctx context.Context, tenantID string, startDate, endDate time.Time, limit int) ([]TopFailure, error) {
	query := `
		SELECT
			e.workflow_id,
			w.name as workflow_name,
			COUNT(*) as failure_count,
			MAX(e.completed_at) as last_failed_at,
			(
				SELECT error_message
				FROM executions
				WHERE workflow_id = e.workflow_id
					AND status = 'failed'
					AND error_message IS NOT NULL
				ORDER BY completed_at DESC
				LIMIT 1
			) as error_preview
		FROM executions e
		INNER JOIN workflows w ON e.workflow_id = w.id
		WHERE e.tenant_id = $1
			AND e.created_at >= $2
			AND e.created_at < $3
			AND e.status = 'failed'
		GROUP BY e.workflow_id, w.name
		ORDER BY failure_count DESC
		LIMIT $4
	`

	var failures []TopFailure
	err := r.db.SelectContext(ctx, &failures, query, tenantID, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("get top failures: %w", err)
	}

	return failures, nil
}

// GetTriggerTypeBreakdown returns execution count by trigger type
func (r *Repository) GetTriggerTypeBreakdown(ctx context.Context, tenantID string, startDate, endDate time.Time) ([]TriggerTypeBreakdown, error) {
	query := `
		WITH trigger_counts AS (
			SELECT
				trigger_type,
				COUNT(*) as count
			FROM executions
			WHERE tenant_id = $1
				AND created_at >= $2
				AND created_at < $3
			GROUP BY trigger_type
		),
		total_count AS (
			SELECT SUM(count) as total FROM trigger_counts
		)
		SELECT
			tc.trigger_type,
			tc.count,
			ROUND((tc.count::numeric / t.total * 100), 2) as percentage
		FROM trigger_counts tc
		CROSS JOIN total_count t
		ORDER BY tc.count DESC
	`

	var breakdown []TriggerTypeBreakdown
	err := r.db.SelectContext(ctx, &breakdown, query, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("get trigger type breakdown: %w", err)
	}

	return breakdown, nil
}
