package quota

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gorax/gorax/internal/workflow"
)

// ExecutorMiddleware wraps an executor and tracks quota usage
type ExecutorMiddleware struct {
	tracker *Tracker
	logger  *slog.Logger
}

// NewExecutorMiddleware creates a new executor middleware
func NewExecutorMiddleware(tracker *Tracker, logger *slog.Logger) *ExecutorMiddleware {
	return &ExecutorMiddleware{
		tracker: tracker,
		logger:  logger,
	}
}

// BeforeExecute is called before workflow execution starts
func (m *ExecutorMiddleware) BeforeExecute(ctx context.Context, execution *workflow.Execution) error {
	// Increment workflow execution counter
	if err := m.tracker.IncrementWorkflowExecutions(ctx, execution.TenantID, PeriodDaily); err != nil {
		m.logger.Error("failed to increment daily workflow executions",
			"error", err,
			"tenant_id", execution.TenantID,
			"execution_id", execution.ID,
		)
		// Don't fail execution on tracking error
	}

	if err := m.tracker.IncrementWorkflowExecutions(ctx, execution.TenantID, PeriodMonthly); err != nil {
		m.logger.Error("failed to increment monthly workflow executions",
			"error", err,
			"tenant_id", execution.TenantID,
			"execution_id", execution.ID,
		)
	}

	return nil
}

// AfterExecute is called after workflow execution completes
func (m *ExecutorMiddleware) AfterExecute(ctx context.Context, execution *workflow.Execution, err error) {
	// If execution was cancelled or failed early, optionally decrement
	// For now, we count all started executions regardless of outcome
	if err != nil {
		m.logger.Debug("execution completed with error",
			"error", err,
			"execution_id", execution.ID,
		)
	}
}

// OnStepExecute is called for each step execution
func (m *ExecutorMiddleware) OnStepExecute(ctx context.Context, tenantID, executionID, nodeID string) {
	// Increment step execution counter
	if err := m.tracker.IncrementStepExecutions(ctx, tenantID, PeriodDaily); err != nil {
		m.logger.Error("failed to increment daily step executions",
			"error", err,
			"tenant_id", tenantID,
			"execution_id", executionID,
			"node_id", nodeID,
		)
	}

	if err := m.tracker.IncrementStepExecutions(ctx, tenantID, PeriodMonthly); err != nil {
		m.logger.Error("failed to increment monthly step executions",
			"error", err,
			"tenant_id", tenantID,
			"execution_id", executionID,
			"node_id", nodeID,
		)
	}
}

// CheckQuotaBeforeExecution checks if tenant can execute workflow
func (m *ExecutorMiddleware) CheckQuotaBeforeExecution(ctx context.Context, tenantID string, quotas TenantQuotas) error {
	// Check daily execution quota
	if quotas.MaxExecutionsPerDay > 0 {
		exceeded, remaining, err := m.tracker.CheckQuota(ctx, tenantID, PeriodDaily, int64(quotas.MaxExecutionsPerDay))
		if err != nil {
			return fmt.Errorf("failed to check quota: %w", err)
		}

		if exceeded {
			return fmt.Errorf("daily execution quota exceeded (0 remaining)")
		}

		m.logger.Debug("quota check passed",
			"tenant_id", tenantID,
			"remaining", remaining,
			"period", "daily",
		)
	}

	return nil
}

// TenantQuotas represents tenant quota limits
type TenantQuotas struct {
	MaxExecutionsPerDay int `json:"max_executions_per_day"`
}
