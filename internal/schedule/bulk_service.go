package schedule

import (
	"context"
	"fmt"
)

// BulkOperationError represents an error for a single item
type BulkOperationError struct {
	ID    string
	Error string
}

// BulkUpdate updates multiple schedules (enable/disable)
func (s *Service) BulkUpdate(ctx context.Context, tenantID string, ids []string, enabled bool) ([]string, []BulkOperationError) {
	var success []string
	var failed []BulkOperationError

	for _, id := range ids {
		err := s.updateScheduleEnabled(ctx, tenantID, id, enabled)
		if err != nil {
			failed = append(failed, BulkOperationError{
				ID:    id,
				Error: err.Error(),
			})
			s.logger.Error("failed to update schedule in bulk operation",
				"schedule_id", id,
				"enabled", enabled,
				"error", err,
			)
		} else {
			success = append(success, id)
		}
	}

	return success, failed
}

// BulkDelete deletes multiple schedules
func (s *Service) BulkDelete(ctx context.Context, tenantID string, ids []string) ([]string, []BulkOperationError) {
	var success []string
	var failed []BulkOperationError

	for _, id := range ids {
		err := s.Delete(ctx, tenantID, id)
		if err != nil {
			failed = append(failed, BulkOperationError{
				ID:    id,
				Error: err.Error(),
			})
			s.logger.Error("failed to delete schedule in bulk operation",
				"schedule_id", id,
				"error", err,
			)
		} else {
			success = append(success, id)
		}
	}

	return success, failed
}

// ExportSchedules exports schedules as a list
func (s *Service) ExportSchedules(ctx context.Context, tenantID string, ids []string) ([]*Schedule, error) {
	if len(ids) == 0 {
		// Export all schedules for tenant
		allSchedules, err := s.repo.ListAll(ctx, tenantID, 1000, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to list all schedules: %w", err)
		}
		// Convert from ScheduleWithWorkflow to Schedule
		schedules := make([]*Schedule, len(allSchedules))
		for i, sched := range allSchedules {
			schedules[i] = &sched.Schedule
		}
		return schedules, nil
	}

	// Export specific schedules
	var schedules []*Schedule
	for _, id := range ids {
		schedule, err := s.repo.GetByID(ctx, tenantID, id)
		if err != nil {
			if err == ErrNotFound {
				continue // Skip not found schedules
			}
			return nil, fmt.Errorf("failed to get schedule %s: %w", id, err)
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

// updateScheduleEnabled updates just the enabled field of a schedule
func (s *Service) updateScheduleEnabled(ctx context.Context, tenantID, id string, enabled bool) error {
	// Get existing schedule
	existing, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}

	// Update enabled status
	input := UpdateScheduleInput{
		Enabled: &enabled,
	}

	_, err = s.Update(ctx, tenantID, id, input)
	if err != nil {
		return err
	}

	// If enabling, recalculate next run time
	if enabled && !existing.Enabled {
		nextRun, err := s.calculateNextRun(existing.CronExpression, existing.Timezone)
		if err != nil {
			s.logger.Error("failed to calculate next run time",
				"error", err,
				"schedule_id", id,
			)
			return err
		}

		if err := s.repo.UpdateNextRunTime(ctx, id, nextRun); err != nil {
			s.logger.Error("failed to update next run time",
				"error", err,
				"schedule_id", id,
			)
			return err
		}
	}

	return nil
}
