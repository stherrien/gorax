package schedule

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// OverlapHandler handles overlap policy decisions for schedule executions
type OverlapHandler struct {
	repo   *Repository
	logger *slog.Logger
}

// NewOverlapHandler creates a new overlap handler
func NewOverlapHandler(repo *Repository, logger *slog.Logger) *OverlapHandler {
	return &OverlapHandler{
		repo:   repo,
		logger: logger,
	}
}

// OverlapDecision represents the decision made by the overlap handler
type OverlapDecision struct {
	ShouldExecute    bool
	ShouldTerminate  bool
	SkipReason       string
	RunningExecution *string
}

// CheckOverlap checks if execution should proceed based on overlap policy
func (h *OverlapHandler) CheckOverlap(ctx context.Context, schedule *Schedule) (*OverlapDecision, error) {
	decision := &OverlapDecision{
		ShouldExecute:   true,
		ShouldTerminate: false,
	}

	// Check if there's a running execution
	hasRunning, runningID, err := h.repo.HasRunningExecution(ctx, schedule.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check running execution: %w", err)
	}

	if !hasRunning {
		// No running execution, proceed normally
		return decision, nil
	}

	decision.RunningExecution = runningID

	// Apply overlap policy
	switch schedule.OverlapPolicy {
	case OverlapPolicySkip:
		decision.ShouldExecute = false
		decision.SkipReason = fmt.Sprintf("previous execution %s still running (policy: skip)", *runningID)
		h.logger.Info("skipping execution due to overlap policy",
			"schedule_id", schedule.ID,
			"running_execution_id", *runningID,
			"policy", schedule.OverlapPolicy,
		)

	case OverlapPolicyQueue:
		// Queue policy - skip this trigger but don't permanently skip
		// The next check cycle will attempt again
		decision.ShouldExecute = false
		decision.SkipReason = fmt.Sprintf("previous execution %s still running (policy: queue, will retry)", *runningID)
		h.logger.Info("queuing execution due to overlap policy",
			"schedule_id", schedule.ID,
			"running_execution_id", *runningID,
			"policy", schedule.OverlapPolicy,
		)

	case OverlapPolicyTerminate:
		decision.ShouldTerminate = true
		h.logger.Info("terminating previous execution due to overlap policy",
			"schedule_id", schedule.ID,
			"running_execution_id", *runningID,
			"policy", schedule.OverlapPolicy,
		)

	default:
		// Default to skip if unknown policy
		decision.ShouldExecute = false
		decision.SkipReason = fmt.Sprintf("unknown overlap policy: %s", schedule.OverlapPolicy)
	}

	return decision, nil
}

// RecordExecutionStart records the start of an execution
func (h *OverlapHandler) RecordExecutionStart(ctx context.Context, schedule *Schedule, executionID string, triggerTime time.Time) (*ExecutionLog, error) {
	// Create execution log
	log, err := h.repo.CreateExecutionLog(ctx, schedule.TenantID, schedule.ID, triggerTime)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution log: %w", err)
	}

	// Mark as started
	if err := h.repo.UpdateExecutionLogStarted(ctx, log.ID, executionID); err != nil {
		return nil, fmt.Errorf("failed to update execution log started: %w", err)
	}

	// Set running execution on schedule
	if err := h.repo.SetRunningExecution(ctx, schedule.ID, executionID); err != nil {
		return nil, fmt.Errorf("failed to set running execution: %w", err)
	}

	h.logger.Info("execution started",
		"schedule_id", schedule.ID,
		"execution_id", executionID,
		"log_id", log.ID,
	)

	return log, nil
}

// RecordExecutionComplete records the completion of an execution
func (h *OverlapHandler) RecordExecutionComplete(ctx context.Context, scheduleID, logID string) error {
	if err := h.repo.UpdateExecutionLogCompleted(ctx, logID); err != nil {
		return fmt.Errorf("failed to update execution log completed: %w", err)
	}

	if err := h.repo.ClearRunningExecution(ctx, scheduleID); err != nil {
		return fmt.Errorf("failed to clear running execution: %w", err)
	}

	h.logger.Info("execution completed",
		"schedule_id", scheduleID,
		"log_id", logID,
	)

	return nil
}

// RecordExecutionFailed records a failed execution
func (h *OverlapHandler) RecordExecutionFailed(ctx context.Context, scheduleID, logID, errorMsg string) error {
	if err := h.repo.UpdateExecutionLogFailed(ctx, logID, errorMsg); err != nil {
		return fmt.Errorf("failed to update execution log failed: %w", err)
	}

	if err := h.repo.ClearRunningExecution(ctx, scheduleID); err != nil {
		return fmt.Errorf("failed to clear running execution: %w", err)
	}

	h.logger.Info("execution failed",
		"schedule_id", scheduleID,
		"log_id", logID,
		"error", errorMsg,
	)

	return nil
}

// RecordExecutionSkipped records a skipped execution
func (h *OverlapHandler) RecordExecutionSkipped(ctx context.Context, schedule *Schedule, triggerTime time.Time, reason string) error {
	log, err := h.repo.CreateExecutionLog(ctx, schedule.TenantID, schedule.ID, triggerTime)
	if err != nil {
		return fmt.Errorf("failed to create execution log: %w", err)
	}

	if err := h.repo.UpdateExecutionLogSkipped(ctx, log.ID, reason); err != nil {
		return fmt.Errorf("failed to update execution log skipped: %w", err)
	}

	h.logger.Info("execution skipped",
		"schedule_id", schedule.ID,
		"log_id", log.ID,
		"reason", reason,
	)

	return nil
}

// RecordExecutionTerminated records that an execution was terminated
func (h *OverlapHandler) RecordExecutionTerminated(ctx context.Context, scheduleID string) error {
	// Get the running execution log
	log, err := h.repo.GetRunningExecutionLogBySchedule(ctx, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to get running execution log: %w", err)
	}

	if log != nil {
		if err := h.repo.UpdateExecutionLogTerminated(ctx, log.ID); err != nil {
			return fmt.Errorf("failed to update execution log terminated: %w", err)
		}
	}

	if err := h.repo.ClearRunningExecution(ctx, scheduleID); err != nil {
		return fmt.Errorf("failed to clear running execution: %w", err)
	}

	h.logger.Info("execution terminated",
		"schedule_id", scheduleID,
	)

	return nil
}
