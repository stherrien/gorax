package schedule

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// ConflictStrategy defines how to handle schedule conflicts
type ConflictStrategy string

const (
	// ConflictStrategySkip skips the new execution if one is already running
	ConflictStrategySkip ConflictStrategy = "skip"
	// ConflictStrategyQueue queues the new execution to run after the current one
	ConflictStrategyQueue ConflictStrategy = "queue"
	// ConflictStrategyReplace cancels the running execution and starts a new one
	ConflictStrategyReplace ConflictStrategy = "replace"
	// ConflictStrategyParallel allows parallel executions
	ConflictStrategyParallel ConflictStrategy = "parallel"
)

// ConflictDetector detects and manages schedule conflicts
type ConflictDetector struct {
	repo   ConflictRepository
	logger *slog.Logger
	parser *CronParser
}

// ConflictRepository defines the interface for conflict-related database operations
type ConflictRepository interface {
	// GetRunningExecutions returns all currently running executions for a schedule
	GetRunningExecutions(ctx context.Context, scheduleID string) ([]*ScheduleExecution, error)
	// GetSchedulesByWorkflow returns all schedules for a workflow
	GetSchedulesByWorkflow(ctx context.Context, tenantID, workflowID string) ([]*Schedule, error)
	// GetOverlappingSchedules returns schedules that may overlap with the given time window
	GetOverlappingSchedules(ctx context.Context, tenantID, workflowID string, startTime, endTime time.Time) ([]*Schedule, error)
}

// NewConflictDetector creates a new conflict detector
func NewConflictDetector(repo ConflictRepository, logger *slog.Logger) *ConflictDetector {
	return &ConflictDetector{
		repo:   repo,
		logger: logger,
		parser: NewCronParser(),
	}
}

// ConflictCheckResult contains the result of a conflict check
type ConflictCheckResult struct {
	HasConflict       bool             `json:"has_conflict"`
	ConflictType      string           `json:"conflict_type,omitempty"`
	ConflictingIDs    []string         `json:"conflicting_ids,omitempty"`
	Message           string           `json:"message,omitempty"`
	RecommendedAction ConflictStrategy `json:"recommended_action,omitempty"`
	Details           *ConflictDetails `json:"details,omitempty"`
}

// ConflictDetails provides detailed information about a conflict
type ConflictDetails struct {
	RunningExecutions     int          `json:"running_executions,omitempty"`
	OverlappingWindows    []TimeWindow `json:"overlapping_windows,omitempty"`
	SameTimeSchedules     []string     `json:"same_time_schedules,omitempty"`
	EstimatedNextConflict *time.Time   `json:"estimated_next_conflict,omitempty"`
}

// TimeWindow represents a time window for conflict detection
type TimeWindow struct {
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
	ScheduleID string    `json:"schedule_id"`
}

// CheckExecutionConflict checks if there's a conflict for executing a schedule
func (cd *ConflictDetector) CheckExecutionConflict(ctx context.Context, schedule *Schedule) (*ConflictCheckResult, error) {
	result := &ConflictCheckResult{
		HasConflict: false,
		Details:     &ConflictDetails{},
	}

	// Check for running executions
	runningExecs, err := cd.repo.GetRunningExecutions(ctx, schedule.ID)
	if err != nil {
		cd.logger.Error("failed to check running executions",
			"schedule_id", schedule.ID,
			"error", err,
		)
		return nil, fmt.Errorf("failed to check running executions: %w", err)
	}

	if len(runningExecs) > 0 {
		result.HasConflict = true
		result.ConflictType = "running_execution"
		result.Message = fmt.Sprintf("schedule has %d running execution(s)", len(runningExecs))
		result.Details.RunningExecutions = len(runningExecs)
		result.RecommendedAction = ConflictStrategySkip

		for _, exec := range runningExecs {
			result.ConflictingIDs = append(result.ConflictingIDs, exec.ID)
		}

		cd.logger.Warn("execution conflict detected",
			"schedule_id", schedule.ID,
			"running_count", len(runningExecs),
		)
	}

	return result, nil
}

// CheckScheduleConflict checks if a new schedule conflicts with existing schedules
func (cd *ConflictDetector) CheckScheduleConflict(ctx context.Context, tenantID, workflowID string, cronExpr, timezone string, excludeScheduleID string) (*ConflictCheckResult, error) {
	result := &ConflictCheckResult{
		HasConflict: false,
		Details:     &ConflictDetails{},
	}

	// Get all schedules for the workflow
	existingSchedules, err := cd.repo.GetSchedulesByWorkflow(ctx, tenantID, workflowID)
	if err != nil {
		cd.logger.Error("failed to get existing schedules",
			"workflow_id", workflowID,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get existing schedules: %w", err)
	}

	// Calculate next 10 execution times for the new schedule
	newTimes, err := cd.parser.CalculateNextRuns(cronExpr, timezone, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate next runs for new schedule: %w", err)
	}

	// Check each existing schedule for conflicts
	for _, existing := range existingSchedules {
		// Skip the schedule being updated
		if existing.ID == excludeScheduleID {
			continue
		}

		// Skip disabled schedules
		if !existing.Enabled {
			continue
		}

		// Calculate next 10 execution times for existing schedule
		existingTimes, err := cd.parser.CalculateNextRuns(existing.CronExpression, existing.Timezone, 10)
		if err != nil {
			cd.logger.Warn("failed to calculate next runs for existing schedule",
				"schedule_id", existing.ID,
				"error", err,
			)
			continue
		}

		// Check for overlapping execution times (within 1 minute tolerance)
		for _, newTime := range newTimes {
			for _, existingTime := range existingTimes {
				diff := newTime.Sub(existingTime)
				if diff < 0 {
					diff = -diff
				}

				// If executions are within 1 minute of each other, it's a potential conflict
				if diff < time.Minute {
					result.HasConflict = true
					result.ConflictType = "overlapping_schedule"
					result.ConflictingIDs = append(result.ConflictingIDs, existing.ID)
					result.Details.SameTimeSchedules = append(result.Details.SameTimeSchedules, existing.Name)

					if result.Details.EstimatedNextConflict == nil {
						conflictTime := newTime
						result.Details.EstimatedNextConflict = &conflictTime
					}

					result.Details.OverlappingWindows = append(result.Details.OverlappingWindows, TimeWindow{
						Start:      newTime.Add(-time.Minute),
						End:        newTime.Add(time.Minute),
						ScheduleID: existing.ID,
					})
					break // Found conflict with this schedule, move to next
				}
			}
		}
	}

	if result.HasConflict {
		result.Message = fmt.Sprintf("schedule conflicts with %d existing schedule(s)", len(result.ConflictingIDs))
		result.RecommendedAction = ConflictStrategyQueue

		cd.logger.Warn("schedule conflict detected",
			"workflow_id", workflowID,
			"conflicting_schedules", len(result.ConflictingIDs),
		)
	}

	return result, nil
}

// ResolveConflict applies a conflict resolution strategy
func (cd *ConflictDetector) ResolveConflict(ctx context.Context, schedule *Schedule, conflict *ConflictCheckResult, strategy ConflictStrategy) (*ConflictResolution, error) {
	resolution := &ConflictResolution{
		Strategy:   strategy,
		ScheduleID: schedule.ID,
		Timestamp:  time.Now(),
	}

	switch strategy {
	case ConflictStrategySkip:
		resolution.Action = "skipped"
		resolution.Message = "execution skipped due to running instance"
		resolution.ShouldExecute = false

		cd.logger.Info("conflict resolved by skipping",
			"schedule_id", schedule.ID,
			"strategy", strategy,
		)

	case ConflictStrategyQueue:
		resolution.Action = "queued"
		resolution.Message = "execution queued for later"
		resolution.ShouldExecute = false
		resolution.QueuedForLater = true

		cd.logger.Info("conflict resolved by queueing",
			"schedule_id", schedule.ID,
			"strategy", strategy,
		)

	case ConflictStrategyReplace:
		resolution.Action = "replacing"
		resolution.Message = "cancelling running execution and starting new one"
		resolution.ShouldExecute = true
		resolution.CancelRunning = true
		resolution.ExecutionsToCancel = conflict.ConflictingIDs

		cd.logger.Info("conflict resolved by replacing",
			"schedule_id", schedule.ID,
			"strategy", strategy,
			"cancelling", len(conflict.ConflictingIDs),
		)

	case ConflictStrategyParallel:
		resolution.Action = "parallel"
		resolution.Message = "executing in parallel with existing instance"
		resolution.ShouldExecute = true

		cd.logger.Info("conflict resolved by parallel execution",
			"schedule_id", schedule.ID,
			"strategy", strategy,
		)

	default:
		return nil, fmt.Errorf("unknown conflict strategy: %s", strategy)
	}

	return resolution, nil
}

// ConflictResolution represents the result of conflict resolution
type ConflictResolution struct {
	Strategy           ConflictStrategy `json:"strategy"`
	Action             string           `json:"action"`
	Message            string           `json:"message"`
	ScheduleID         string           `json:"schedule_id"`
	ShouldExecute      bool             `json:"should_execute"`
	QueuedForLater     bool             `json:"queued_for_later,omitempty"`
	CancelRunning      bool             `json:"cancel_running,omitempty"`
	ExecutionsToCancel []string         `json:"executions_to_cancel,omitempty"`
	Timestamp          time.Time        `json:"timestamp"`
}

// ConflictConfig contains configuration for conflict handling
type ConflictConfig struct {
	DefaultStrategy    ConflictStrategy `json:"default_strategy"`
	MaxParallelExecs   int              `json:"max_parallel_executions"`
	MaxQueuedExecs     int              `json:"max_queued_executions"`
	QueueTimeout       time.Duration    `json:"queue_timeout"`
	ConflictWindowSecs int              `json:"conflict_window_seconds"`
}

// DefaultConflictConfig returns the default conflict configuration
func DefaultConflictConfig() *ConflictConfig {
	return &ConflictConfig{
		DefaultStrategy:    ConflictStrategySkip,
		MaxParallelExecs:   5,
		MaxQueuedExecs:     100,
		QueueTimeout:       time.Hour,
		ConflictWindowSecs: 60,
	}
}

// ValidateConflictStrategy validates a conflict strategy string
func ValidateConflictStrategy(strategy string) (ConflictStrategy, error) {
	switch ConflictStrategy(strategy) {
	case ConflictStrategySkip, ConflictStrategyQueue, ConflictStrategyReplace, ConflictStrategyParallel:
		return ConflictStrategy(strategy), nil
	default:
		return "", fmt.Errorf("invalid conflict strategy: %s (valid: skip, queue, replace, parallel)", strategy)
	}
}

// ScheduleConflictInfo contains information about schedule conflicts for API responses
type ScheduleConflictInfo struct {
	ScheduleID        string           `json:"schedule_id"`
	WorkflowID        string           `json:"workflow_id"`
	HasActiveConflict bool             `json:"has_active_conflict"`
	ConflictStrategy  ConflictStrategy `json:"conflict_strategy"`
	RunningCount      int              `json:"running_count"`
	QueuedCount       int              `json:"queued_count"`
	LastConflict      *time.Time       `json:"last_conflict,omitempty"`
}
