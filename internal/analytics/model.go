package analytics

import "time"

// TimeRange represents a time range for analytics queries
type TimeRange struct {
	StartDate time.Time
	EndDate   time.Time
}

// Granularity represents the time granularity for trend data
type Granularity string

const (
	GranularityHour  Granularity = "hour"
	GranularityDay   Granularity = "day"
	GranularityWeek  Granularity = "week"
	GranularityMonth Granularity = "month"
)

// WorkflowStats represents statistics for a specific workflow
type WorkflowStats struct {
	WorkflowID       string    `db:"workflow_id" json:"workflowId"`
	WorkflowName     string    `db:"workflow_name" json:"workflowName"`
	ExecutionCount   int       `db:"execution_count" json:"executionCount"`
	SuccessCount     int       `db:"success_count" json:"successCount"`
	FailureCount     int       `db:"failure_count" json:"failureCount"`
	SuccessRate      float64   `db:"success_rate" json:"successRate"`
	AvgDurationMs    int64     `db:"avg_duration_ms" json:"avgDurationMs"`
	MinDurationMs    int64     `db:"min_duration_ms" json:"minDurationMs"`
	MaxDurationMs    int64     `db:"max_duration_ms" json:"maxDurationMs"`
	LastExecutedAt   time.Time `db:"last_executed_at" json:"lastExecutedAt"`
	TotalDurationMs  int64     `db:"total_duration_ms" json:"totalDurationMs"`
	CancelledCount   int       `db:"cancelled_count" json:"cancelledCount"`
	PendingCount     int       `db:"pending_count" json:"pendingCount"`
	RunningCount     int       `db:"running_count" json:"runningCount"`
}

// TenantOverview represents overall statistics for a tenant
type TenantOverview struct {
	TotalExecutions     int     `db:"total_executions" json:"totalExecutions"`
	SuccessfulExecutions int     `db:"successful_executions" json:"successfulExecutions"`
	FailedExecutions    int     `db:"failed_executions" json:"failedExecutions"`
	CancelledExecutions int     `db:"cancelled_executions" json:"cancelledExecutions"`
	PendingExecutions   int     `db:"pending_executions" json:"pendingExecutions"`
	RunningExecutions   int     `db:"running_executions" json:"runningExecutions"`
	SuccessRate         float64 `db:"success_rate" json:"successRate"`
	AvgDurationMs       int64   `db:"avg_duration_ms" json:"avgDurationMs"`
	ActiveWorkflows     int     `db:"active_workflows" json:"activeWorkflows"`
	TotalWorkflows      int     `db:"total_workflows" json:"totalWorkflows"`
}

// TimeSeriesPoint represents a single data point in a time series
type TimeSeriesPoint struct {
	Timestamp        time.Time `db:"timestamp" json:"timestamp"`
	ExecutionCount   int       `db:"execution_count" json:"executionCount"`
	SuccessCount     int       `db:"success_count" json:"successCount"`
	FailureCount     int       `db:"failure_count" json:"failureCount"`
	SuccessRate      float64   `db:"success_rate" json:"successRate"`
	AvgDurationMs    int64     `db:"avg_duration_ms" json:"avgDurationMs"`
}

// ExecutionTrends represents execution trends over time
type ExecutionTrends struct {
	Granularity Granularity        `json:"granularity"`
	StartDate   time.Time          `json:"startDate"`
	EndDate     time.Time          `json:"endDate"`
	DataPoints  []TimeSeriesPoint  `json:"dataPoints"`
}

// TopWorkflow represents a workflow in the top workflows list
type TopWorkflow struct {
	WorkflowID      string  `db:"workflow_id" json:"workflowId"`
	WorkflowName    string  `db:"workflow_name" json:"workflowName"`
	ExecutionCount  int     `db:"execution_count" json:"executionCount"`
	SuccessRate     float64 `db:"success_rate" json:"successRate"`
	AvgDurationMs   int64   `db:"avg_duration_ms" json:"avgDurationMs"`
	LastExecutedAt  time.Time `db:"last_executed_at" json:"lastExecutedAt"`
}

// TopWorkflows represents the most frequently executed workflows
type TopWorkflows struct {
	Workflows []TopWorkflow `json:"workflows"`
	Total     int           `json:"total"`
}

// ErrorInfo represents error information with frequency
type ErrorInfo struct {
	ErrorMessage   string    `db:"error_message" json:"errorMessage"`
	ErrorCount     int       `db:"error_count" json:"errorCount"`
	WorkflowID     string    `db:"workflow_id" json:"workflowId"`
	WorkflowName   string    `db:"workflow_name" json:"workflowName"`
	LastOccurrence time.Time `db:"last_occurrence" json:"lastOccurrence"`
	Percentage     float64   `db:"percentage" json:"percentage"`
}

// ErrorBreakdown represents error analysis
type ErrorBreakdown struct {
	TotalErrors  int         `json:"totalErrors"`
	ErrorsByType []ErrorInfo `json:"errorsByType"`
}

// NodeStats represents statistics for a specific node
type NodeStats struct {
	NodeID          string  `db:"node_id" json:"nodeId"`
	NodeType        string  `db:"node_type" json:"nodeType"`
	ExecutionCount  int     `db:"execution_count" json:"executionCount"`
	SuccessCount    int     `db:"success_count" json:"successCount"`
	FailureCount    int     `db:"failure_count" json:"failureCount"`
	SuccessRate     float64 `db:"success_rate" json:"successRate"`
	AvgDurationMs   int64   `db:"avg_duration_ms" json:"avgDurationMs"`
	MinDurationMs   int64   `db:"min_duration_ms" json:"minDurationMs"`
	MaxDurationMs   int64   `db:"max_duration_ms" json:"maxDurationMs"`
}

// NodePerformance represents node-level performance statistics
type NodePerformance struct {
	WorkflowID   string      `json:"workflowId"`
	WorkflowName string      `json:"workflowName"`
	Nodes        []NodeStats `json:"nodes"`
}
