package schedule

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// WorkflowExecutor interface for triggering workflow executions
type WorkflowExecutor interface {
	ExecuteScheduled(ctx context.Context, tenantID, workflowID, scheduleID string) (executionID string, err error)
}

// ExecutionTerminator interface for terminating workflow executions
type ExecutionTerminator interface {
	TerminateExecution(ctx context.Context, executionID string) error
}

// ScheduleProvider interface for getting due schedules
type ScheduleProvider interface {
	GetDueSchedules(ctx context.Context) ([]*Schedule, error)
	MarkScheduleRun(ctx context.Context, scheduleID, executionID string) error
}

// Scheduler manages scheduled workflow executions
type Scheduler struct {
	provider       ScheduleProvider
	executor       WorkflowExecutor
	terminator     ExecutionTerminator
	overlapHandler *OverlapHandler
	logger         *slog.Logger

	// Scheduler configuration
	checkInterval time.Duration
	batchSize     int

	// Running state
	running bool
	mu      sync.Mutex
	wg      sync.WaitGroup
	stopCh  chan struct{}
}

// NewScheduler creates a new scheduler instance
func NewScheduler(provider ScheduleProvider, executor WorkflowExecutor, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		provider:      provider,
		executor:      executor,
		logger:        logger,
		checkInterval: 30 * time.Second, // Check every 30 seconds
		batchSize:     100,              // Process up to 100 schedules per check
		stopCh:        make(chan struct{}),
	}
}

// SetOverlapHandler sets the overlap handler for the scheduler
func (s *Scheduler) SetOverlapHandler(handler *OverlapHandler) {
	s.overlapHandler = handler
}

// SetTerminator sets the execution terminator for the scheduler
func (s *Scheduler) SetTerminator(terminator ExecutionTerminator) {
	s.terminator = terminator
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	s.logger.Info("scheduler started", "check_interval", s.checkInterval)

	s.wg.Add(1)
	go s.run(ctx)

	return nil
}

// Stop stops the scheduler gracefully
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	s.logger.Info("stopping scheduler...")
	close(s.stopCh)
	s.wg.Wait()
	s.logger.Info("scheduler stopped")
}

// Wait waits for the scheduler to finish
func (s *Scheduler) Wait() {
	s.wg.Wait()
}

// run is the main scheduler loop
func (s *Scheduler) run(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	// Run initial check immediately
	s.checkAndExecuteSchedules(ctx)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("scheduler context cancelled")
			return
		case <-s.stopCh:
			s.logger.Info("scheduler stop signal received")
			return
		case <-ticker.C:
			s.checkAndExecuteSchedules(ctx)
		}
	}
}

// checkAndExecuteSchedules checks for due schedules and executes them
func (s *Scheduler) checkAndExecuteSchedules(ctx context.Context) {
	schedules, err := s.provider.GetDueSchedules(ctx)
	if err != nil {
		s.logger.Error("failed to get due schedules", "error", err)
		return
	}

	if len(schedules) == 0 {
		s.logger.Debug("no schedules due for execution")
		return
	}

	s.logger.Info("found schedules due for execution", "count", len(schedules))

	// Execute schedules concurrently with a limit
	semaphore := make(chan struct{}, 10) // Limit to 10 concurrent executions
	var wg sync.WaitGroup

	for _, schedule := range schedules {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(sch *Schedule) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			s.executeSchedule(ctx, sch)
		}(schedule)
	}

	wg.Wait()
	s.logger.Info("finished processing due schedules", "count", len(schedules))
}

// executeSchedule executes a single schedule with overlap policy handling
func (s *Scheduler) executeSchedule(ctx context.Context, schedule *Schedule) {
	triggerTime := time.Now()

	s.logger.Info("executing schedule",
		"schedule_id", schedule.ID,
		"workflow_id", schedule.WorkflowID,
		"tenant_id", schedule.TenantID,
		"name", schedule.Name,
		"overlap_policy", schedule.OverlapPolicy,
	)

	// Check if schedule is still enabled (may have been disabled since query)
	if !schedule.Enabled {
		s.logger.Warn("schedule is disabled, skipping",
			"schedule_id", schedule.ID,
		)
		return
	}

	// Check overlap policy if handler is available
	if s.overlapHandler != nil {
		decision, err := s.overlapHandler.CheckOverlap(ctx, schedule)
		if err != nil {
			s.logger.Error("failed to check overlap policy",
				"error", err,
				"schedule_id", schedule.ID,
			)
			return
		}

		// Handle terminate policy
		if decision.ShouldTerminate {
			if err := s.terminatePreviousExecution(ctx, schedule, decision.RunningExecution); err != nil {
				s.logger.Error("failed to terminate previous execution",
					"error", err,
					"schedule_id", schedule.ID,
					"running_execution_id", decision.RunningExecution,
				)
				return
			}
		}

		// Skip if overlap policy says so
		if !decision.ShouldExecute {
			if err := s.overlapHandler.RecordExecutionSkipped(ctx, schedule, triggerTime, decision.SkipReason); err != nil {
				s.logger.Error("failed to record skipped execution",
					"error", err,
					"schedule_id", schedule.ID,
				)
			}
			// Still mark schedule run to update next run time for skip policy
			if schedule.OverlapPolicy == OverlapPolicySkip {
				if err := s.provider.MarkScheduleRun(ctx, schedule.ID, ""); err != nil {
					s.logger.Error("failed to mark schedule run after skip",
						"error", err,
						"schedule_id", schedule.ID,
					)
				}
			}
			return
		}
	}

	// Execute workflow
	executionID, err := s.executor.ExecuteScheduled(ctx, schedule.TenantID, schedule.WorkflowID, schedule.ID)
	if err != nil {
		s.logger.Error("failed to execute scheduled workflow",
			"error", err,
			"schedule_id", schedule.ID,
			"workflow_id", schedule.WorkflowID,
		)

		// Record failure if overlap handler available
		if s.overlapHandler != nil {
			log, logErr := s.overlapHandler.RecordExecutionStart(ctx, schedule, "", triggerTime)
			if logErr == nil && log != nil {
				_ = s.overlapHandler.RecordExecutionFailed(ctx, schedule.ID, log.ID, err.Error())
			}
		}

		// Still mark the schedule as run to avoid repeated failures
		if err := s.provider.MarkScheduleRun(ctx, schedule.ID, ""); err != nil {
			s.logger.Error("failed to mark schedule run after error",
				"error", err,
				"schedule_id", schedule.ID,
			)
		}
		return
	}

	// Record execution start if overlap handler available
	var execLog *ExecutionLog
	if s.overlapHandler != nil {
		execLog, err = s.overlapHandler.RecordExecutionStart(ctx, schedule, executionID, triggerTime)
		if err != nil {
			s.logger.Error("failed to record execution start",
				"error", err,
				"schedule_id", schedule.ID,
				"execution_id", executionID,
			)
		}
	}

	s.logger.Info("schedule executed successfully",
		"schedule_id", schedule.ID,
		"execution_id", executionID,
	)

	// Mark schedule as run and update next run time
	if err := s.provider.MarkScheduleRun(ctx, schedule.ID, executionID); err != nil {
		s.logger.Error("failed to mark schedule run",
			"error", err,
			"schedule_id", schedule.ID,
		)
	}

	// Note: In a production system, you would want to monitor the execution
	// and call RecordExecutionComplete/RecordExecutionFailed when it finishes.
	// For now, we mark it as complete immediately since we don't have
	// async execution monitoring in this implementation.
	if s.overlapHandler != nil && execLog != nil {
		if err := s.overlapHandler.RecordExecutionComplete(ctx, schedule.ID, execLog.ID); err != nil {
			s.logger.Error("failed to record execution complete",
				"error", err,
				"schedule_id", schedule.ID,
			)
		}
	}
}

// terminatePreviousExecution terminates a running execution
func (s *Scheduler) terminatePreviousExecution(ctx context.Context, schedule *Schedule, runningExecutionID *string) error {
	if runningExecutionID == nil {
		return nil
	}

	// Terminate the execution if terminator is available
	if s.terminator != nil {
		if err := s.terminator.TerminateExecution(ctx, *runningExecutionID); err != nil {
			s.logger.Warn("failed to terminate execution",
				"error", err,
				"execution_id", *runningExecutionID,
			)
			// Continue anyway - we'll record the termination
		}
	}

	// Record the termination
	if s.overlapHandler != nil {
		if err := s.overlapHandler.RecordExecutionTerminated(ctx, schedule.ID); err != nil {
			return err
		}
	}

	return nil
}

// SetCheckInterval sets the interval between schedule checks
func (s *Scheduler) SetCheckInterval(interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkInterval = interval
}

// IsRunning returns whether the scheduler is currently running
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
