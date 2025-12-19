package schedule

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

// WorkflowGetter interface to avoid circular dependencies
type WorkflowGetter interface {
	GetByID(ctx context.Context, tenantID, id string) (interface{}, error)
}

// Service handles schedule business logic
type Service struct {
	repo           *Repository
	workflowGetter WorkflowGetter
	logger         *slog.Logger
	cronParser     cron.Parser
}

// NewService creates a new schedule service
func NewService(repo *Repository, logger *slog.Logger) *Service {
	// Create parser that supports standard cron format with seconds
	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

	return &Service{
		repo:       repo,
		logger:     logger,
		cronParser: parser,
	}
}

// SetWorkflowService sets the workflow service (called after initialization to avoid import cycles)
func (s *Service) SetWorkflowService(workflowGetter WorkflowGetter) {
	s.workflowGetter = workflowGetter
}

// Create creates a new schedule
func (s *Service) Create(ctx context.Context, tenantID, workflowID, userID string, input CreateScheduleInput) (*Schedule, error) {
	// Validate cron expression
	if err := s.validateCronExpression(input.CronExpression); err != nil {
		return nil, err
	}

	// Validate timezone
	if input.Timezone != "" {
		if _, err := time.LoadLocation(input.Timezone); err != nil {
			return nil, &ValidationError{Message: "invalid timezone: " + err.Error()}
		}
	}

	// Verify workflow exists
	if s.workflowGetter != nil {
		if _, err := s.workflowGetter.GetByID(ctx, tenantID, workflowID); err != nil {
			return nil, &ValidationError{Message: "workflow not found"}
		}
	}

	// Create schedule
	schedule, err := s.repo.Create(ctx, tenantID, workflowID, userID, input)
	if err != nil {
		s.logger.Error("failed to create schedule", "error", err, "tenant_id", tenantID, "workflow_id", workflowID)
		return nil, err
	}

	// Calculate and set next run time if enabled
	if schedule.Enabled {
		nextRun, err := s.calculateNextRun(schedule.CronExpression, schedule.Timezone)
		if err != nil {
			s.logger.Error("failed to calculate next run time", "error", err, "schedule_id", schedule.ID)
		} else {
			if err := s.repo.UpdateNextRunTime(ctx, schedule.ID, nextRun); err != nil {
				s.logger.Error("failed to update next run time", "error", err, "schedule_id", schedule.ID)
			} else {
				schedule.NextRunAt = &nextRun
			}
		}
	}

	s.logger.Info("schedule created", "schedule_id", schedule.ID, "workflow_id", workflowID)
	return schedule, nil
}

// GetByID retrieves a schedule by ID
func (s *Service) GetByID(ctx context.Context, tenantID, id string) (*Schedule, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

// Update updates a schedule
func (s *Service) Update(ctx context.Context, tenantID, id string, input UpdateScheduleInput) (*Schedule, error) {
	// Get existing schedule
	existing, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	// Validate cron expression if provided
	if input.CronExpression != nil {
		if err := s.validateCronExpression(*input.CronExpression); err != nil {
			return nil, err
		}
	}

	// Validate timezone if provided
	if input.Timezone != nil && *input.Timezone != "" {
		if _, err := time.LoadLocation(*input.Timezone); err != nil {
			return nil, &ValidationError{Message: "invalid timezone: " + err.Error()}
		}
	}

	// Update schedule
	schedule, err := s.repo.Update(ctx, tenantID, id, input)
	if err != nil {
		s.logger.Error("failed to update schedule", "error", err, "schedule_id", id)
		return nil, err
	}

	// Recalculate next run time if cron expression, timezone, or enabled status changed
	shouldRecalculate := false
	if input.CronExpression != nil || input.Timezone != nil {
		shouldRecalculate = true
	}
	if input.Enabled != nil && *input.Enabled && !existing.Enabled {
		shouldRecalculate = true
	}

	if shouldRecalculate && schedule.Enabled {
		nextRun, err := s.calculateNextRun(schedule.CronExpression, schedule.Timezone)
		if err != nil {
			s.logger.Error("failed to calculate next run time", "error", err, "schedule_id", schedule.ID)
		} else {
			if err := s.repo.UpdateNextRunTime(ctx, schedule.ID, nextRun); err != nil {
				s.logger.Error("failed to update next run time", "error", err, "schedule_id", schedule.ID)
			} else {
				schedule.NextRunAt = &nextRun
			}
		}
	}

	s.logger.Info("schedule updated", "schedule_id", schedule.ID)
	return schedule, nil
}

// Delete deletes a schedule
func (s *Service) Delete(ctx context.Context, tenantID, id string) error {
	err := s.repo.Delete(ctx, tenantID, id)
	if err != nil {
		s.logger.Error("failed to delete schedule", "error", err, "schedule_id", id)
		return err
	}

	s.logger.Info("schedule deleted", "schedule_id", id)
	return nil
}

// List retrieves all schedules for a workflow
func (s *Service) List(ctx context.Context, tenantID, workflowID string, limit, offset int) ([]*Schedule, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.List(ctx, tenantID, workflowID, limit, offset)
}

// ListAll retrieves all schedules for a tenant
func (s *Service) ListAll(ctx context.Context, tenantID string, limit, offset int) ([]*ScheduleWithWorkflow, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.ListAll(ctx, tenantID, limit, offset)
}

// GetDueSchedules retrieves schedules that need to be executed
func (s *Service) GetDueSchedules(ctx context.Context) ([]*Schedule, error) {
	return s.repo.GetDueSchedules(ctx, time.Now())
}

// MarkScheduleRun updates schedule after execution
func (s *Service) MarkScheduleRun(ctx context.Context, scheduleID, executionID string) error {
	schedule, err := s.repo.GetByIDWithoutTenant(ctx, scheduleID)
	if err != nil {
		return err
	}

	// Calculate next run time
	nextRun, err := s.calculateNextRun(schedule.CronExpression, schedule.Timezone)
	if err != nil {
		s.logger.Error("failed to calculate next run time", "error", err, "schedule_id", scheduleID)
		return err
	}

	// Update last run and next run
	err = s.repo.UpdateLastRun(ctx, scheduleID, time.Now(), executionID, nextRun)
	if err != nil {
		s.logger.Error("failed to update schedule run info", "error", err, "schedule_id", scheduleID)
		return err
	}

	s.logger.Info("schedule run marked", "schedule_id", scheduleID, "execution_id", executionID, "next_run", nextRun)
	return nil
}

// validateCronExpression validates a cron expression
func (s *Service) validateCronExpression(expression string) error {
	_, err := s.cronParser.Parse(expression)
	if err != nil {
		return &ValidationError{Message: "invalid cron expression: " + err.Error()}
	}
	return nil
}

// calculateNextRun calculates the next run time for a cron expression
func (s *Service) calculateNextRun(expression, timezone string) (time.Time, error) {
	// Parse cron expression
	sched, err := s.cronParser.Parse(expression)
	if err != nil {
		return time.Time{}, err
	}

	// Load timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	// Calculate next run time
	now := time.Now().In(loc)
	nextRun := sched.Next(now)

	return nextRun, nil
}

// ParseNextRunTime is a helper to parse and return next run time (useful for API responses)
func (s *Service) ParseNextRunTime(expression, timezone string) (time.Time, error) {
	return s.calculateNextRun(expression, timezone)
}
