package analytics

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

var (
	ErrNotFound = errors.New("not found")
)

// Repository handles analytics database operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new analytics repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// GetWorkflowStats retrieves statistics for a specific workflow
func (r *Repository) GetWorkflowStats(ctx context.Context, tenantID, workflowID string, timeRange TimeRange) (*WorkflowStats, error) {
	query := `
		SELECT
			e.workflow_id,
			w.name as workflow_name,
			COUNT(*) as execution_count,
			COUNT(*) FILTER (WHERE e.status = 'completed') as success_count,
			COUNT(*) FILTER (WHERE e.status = 'failed') as failure_count,
			COUNT(*) FILTER (WHERE e.status = 'cancelled') as cancelled_count,
			COUNT(*) FILTER (WHERE e.status = 'pending') as pending_count,
			COUNT(*) FILTER (WHERE e.status = 'running') as running_count,
			COALESCE(
				CAST(COUNT(*) FILTER (WHERE e.status = 'completed') AS FLOAT) /
				NULLIF(COUNT(*), 0),
				0
			) as success_rate,
			COALESCE(
				AVG(EXTRACT(EPOCH FROM (e.completed_at - e.started_at)) * 1000) FILTER (
					WHERE e.completed_at IS NOT NULL AND e.started_at IS NOT NULL
				),
				0
			) as avg_duration_ms,
			COALESCE(
				MIN(EXTRACT(EPOCH FROM (e.completed_at - e.started_at)) * 1000) FILTER (
					WHERE e.completed_at IS NOT NULL AND e.started_at IS NOT NULL
				),
				0
			) as min_duration_ms,
			COALESCE(
				MAX(EXTRACT(EPOCH FROM (e.completed_at - e.started_at)) * 1000) FILTER (
					WHERE e.completed_at IS NOT NULL AND e.started_at IS NOT NULL
				),
				0
			) as max_duration_ms,
			COALESCE(
				SUM(EXTRACT(EPOCH FROM (e.completed_at - e.started_at)) * 1000) FILTER (
					WHERE e.completed_at IS NOT NULL AND e.started_at IS NOT NULL
				),
				0
			) as total_duration_ms,
			MAX(e.created_at) as last_executed_at
		FROM executions e
		JOIN workflows w ON e.workflow_id = w.id
		WHERE e.tenant_id = $1
			AND e.workflow_id = $2
			AND e.created_at >= $3
			AND e.created_at <= $4
		GROUP BY e.workflow_id, w.name
	`

	var stats WorkflowStats
	err := r.db.GetContext(ctx, &stats, query, tenantID, workflowID, timeRange.StartDate, timeRange.EndDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get workflow stats: %w", err)
	}

	return &stats, nil
}

// GetTenantOverview retrieves overall statistics for a tenant
func (r *Repository) GetTenantOverview(ctx context.Context, tenantID string, timeRange TimeRange) (*TenantOverview, error) {
	query := `
		SELECT
			COUNT(*) as total_executions,
			COUNT(*) FILTER (WHERE status = 'completed') as successful_executions,
			COUNT(*) FILTER (WHERE status = 'failed') as failed_executions,
			COUNT(*) FILTER (WHERE status = 'cancelled') as cancelled_executions,
			COUNT(*) FILTER (WHERE status = 'pending') as pending_executions,
			COUNT(*) FILTER (WHERE status = 'running') as running_executions,
			COALESCE(
				CAST(COUNT(*) FILTER (WHERE status = 'completed') AS FLOAT) /
				NULLIF(COUNT(*), 0),
				0
			) as success_rate,
			COALESCE(
				AVG(EXTRACT(EPOCH FROM (completed_at - started_at)) * 1000) FILTER (
					WHERE completed_at IS NOT NULL AND started_at IS NOT NULL
				),
				0
			) as avg_duration_ms
		FROM executions
		WHERE tenant_id = $1
			AND created_at >= $2
			AND created_at <= $3
	`

	var overview TenantOverview
	err := r.db.GetContext(ctx, &overview, query, tenantID, timeRange.StartDate, timeRange.EndDate)
	if err != nil {
		return nil, fmt.Errorf("get tenant overview: %w", err)
	}

	workflowQuery := `
		SELECT
			COUNT(*) FILTER (WHERE status = 'active') as active_workflows,
			COUNT(*) as total_workflows
		FROM workflows
		WHERE tenant_id = $1
			AND status != 'archived'
	`

	var workflowCounts struct {
		ActiveWorkflows int `db:"active_workflows"`
		TotalWorkflows  int `db:"total_workflows"`
	}
	err = r.db.GetContext(ctx, &workflowCounts, workflowQuery, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get workflow counts: %w", err)
	}

	overview.ActiveWorkflows = workflowCounts.ActiveWorkflows
	overview.TotalWorkflows = workflowCounts.TotalWorkflows

	return &overview, nil
}

// GetExecutionTrends retrieves execution trends over time
func (r *Repository) GetExecutionTrends(ctx context.Context, tenantID string, timeRange TimeRange, granularity Granularity) (*ExecutionTrends, error) {
	truncFunc := getTruncFunction(granularity)

	query := fmt.Sprintf(`
		SELECT
			DATE_TRUNC('%s', created_at) as timestamp,
			COUNT(*) as execution_count,
			COUNT(*) FILTER (WHERE status = 'completed') as success_count,
			COUNT(*) FILTER (WHERE status = 'failed') as failure_count,
			COALESCE(
				CAST(COUNT(*) FILTER (WHERE status = 'completed') AS FLOAT) /
				NULLIF(COUNT(*), 0),
				0
			) as success_rate,
			COALESCE(
				AVG(EXTRACT(EPOCH FROM (completed_at - started_at)) * 1000) FILTER (
					WHERE completed_at IS NOT NULL AND started_at IS NOT NULL
				),
				0
			) as avg_duration_ms
		FROM executions
		WHERE tenant_id = $1
			AND created_at >= $2
			AND created_at <= $3
		GROUP BY DATE_TRUNC('%s', created_at)
		ORDER BY timestamp ASC
	`, truncFunc, truncFunc)

	var dataPoints []TimeSeriesPoint
	err := r.db.SelectContext(ctx, &dataPoints, query, tenantID, timeRange.StartDate, timeRange.EndDate)
	if err != nil {
		return nil, fmt.Errorf("get execution trends: %w", err)
	}

	trends := &ExecutionTrends{
		Granularity: granularity,
		StartDate:   timeRange.StartDate,
		EndDate:     timeRange.EndDate,
		DataPoints:  dataPoints,
	}

	return trends, nil
}

// GetTopWorkflows retrieves the most frequently executed workflows
func (r *Repository) GetTopWorkflows(ctx context.Context, tenantID string, timeRange TimeRange, limit int) (*TopWorkflows, error) {
	query := `
		SELECT
			e.workflow_id,
			w.name as workflow_name,
			COUNT(*) as execution_count,
			COALESCE(
				CAST(COUNT(*) FILTER (WHERE e.status = 'completed') AS FLOAT) /
				NULLIF(COUNT(*), 0),
				0
			) as success_rate,
			COALESCE(
				AVG(EXTRACT(EPOCH FROM (e.completed_at - e.started_at)) * 1000) FILTER (
					WHERE e.completed_at IS NOT NULL AND e.started_at IS NOT NULL
				),
				0
			) as avg_duration_ms,
			MAX(e.created_at) as last_executed_at
		FROM executions e
		JOIN workflows w ON e.workflow_id = w.id
		WHERE e.tenant_id = $1
			AND e.created_at >= $2
			AND e.created_at <= $3
		GROUP BY e.workflow_id, w.name
		ORDER BY execution_count DESC
		LIMIT $4
	`

	var workflows []TopWorkflow
	err := r.db.SelectContext(ctx, &workflows, query, tenantID, timeRange.StartDate, timeRange.EndDate, limit)
	if err != nil {
		return nil, fmt.Errorf("get top workflows: %w", err)
	}

	countQuery := `
		SELECT COUNT(DISTINCT workflow_id) as total
		FROM executions
		WHERE tenant_id = $1
			AND created_at >= $2
			AND created_at <= $3
	`

	var total int
	err = r.db.GetContext(ctx, &total, countQuery, tenantID, timeRange.StartDate, timeRange.EndDate)
	if err != nil {
		return nil, fmt.Errorf("get workflow count: %w", err)
	}

	return &TopWorkflows{
		Workflows: workflows,
		Total:     total,
	}, nil
}

// GetErrorBreakdown retrieves error analysis
func (r *Repository) GetErrorBreakdown(ctx context.Context, tenantID string, timeRange TimeRange) (*ErrorBreakdown, error) {
	query := `
		SELECT
			e.error_message,
			COUNT(*) as error_count,
			e.workflow_id,
			w.name as workflow_name,
			MAX(e.created_at) as last_occurrence,
			COALESCE(
				CAST(COUNT(*) AS FLOAT) * 100.0 /
				NULLIF(SUM(COUNT(*)) OVER (), 0),
				0
			) as percentage
		FROM executions e
		JOIN workflows w ON e.workflow_id = w.id
		WHERE e.tenant_id = $1
			AND e.status = 'failed'
			AND e.error_message IS NOT NULL
			AND e.created_at >= $2
			AND e.created_at <= $3
		GROUP BY e.error_message, e.workflow_id, w.name
		ORDER BY error_count DESC
		LIMIT 20
	`

	var errors []ErrorInfo
	err := r.db.SelectContext(ctx, &errors, query, tenantID, timeRange.StartDate, timeRange.EndDate)
	if err != nil {
		return nil, fmt.Errorf("get error breakdown: %w", err)
	}

	countQuery := `
		SELECT COUNT(*) as total_errors
		FROM executions
		WHERE tenant_id = $1
			AND status = 'failed'
			AND created_at >= $2
			AND created_at <= $3
	`

	var totalErrors int
	err = r.db.GetContext(ctx, &totalErrors, countQuery, tenantID, timeRange.StartDate, timeRange.EndDate)
	if err != nil {
		return nil, fmt.Errorf("get total errors: %w", err)
	}

	return &ErrorBreakdown{
		TotalErrors:  totalErrors,
		ErrorsByType: errors,
	}, nil
}

// GetNodePerformance retrieves node-level performance statistics
func (r *Repository) GetNodePerformance(ctx context.Context, tenantID, workflowID string) (*NodePerformance, error) {
	var workflowName string
	err := r.db.GetContext(ctx, &workflowName, "SELECT name FROM workflows WHERE id = $1 AND tenant_id = $2", workflowID, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get workflow name: %w", err)
	}

	query := `
		SELECT
			se.node_id,
			se.node_type,
			COUNT(*) as execution_count,
			COUNT(*) FILTER (WHERE se.status = 'completed') as success_count,
			COUNT(*) FILTER (WHERE se.status = 'failed') as failure_count,
			COALESCE(
				CAST(COUNT(*) FILTER (WHERE se.status = 'completed') AS FLOAT) /
				NULLIF(COUNT(*), 0),
				0
			) as success_rate,
			COALESCE(AVG(se.duration_ms) FILTER (WHERE se.duration_ms IS NOT NULL), 0) as avg_duration_ms,
			COALESCE(MIN(se.duration_ms) FILTER (WHERE se.duration_ms IS NOT NULL), 0) as min_duration_ms,
			COALESCE(MAX(se.duration_ms) FILTER (WHERE se.duration_ms IS NOT NULL), 0) as max_duration_ms
		FROM step_executions se
		JOIN executions e ON se.execution_id = e.id
		WHERE e.workflow_id = $1
			AND e.tenant_id = $2
		GROUP BY se.node_id, se.node_type
		ORDER BY execution_count DESC
	`

	var nodes []NodeStats
	err = r.db.SelectContext(ctx, &nodes, query, workflowID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get node performance: %w", err)
	}

	return &NodePerformance{
		WorkflowID:   workflowID,
		WorkflowName: workflowName,
		Nodes:        nodes,
	}, nil
}

// getTruncFunction returns the appropriate PostgreSQL date truncation string
func getTruncFunction(granularity Granularity) string {
	switch granularity {
	case GranularityHour:
		return "hour"
	case GranularityDay:
		return "day"
	case GranularityWeek:
		return "week"
	case GranularityMonth:
		return "month"
	default:
		return "day"
	}
}
