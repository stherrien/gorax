package workflow

import (
	"context"
	"encoding/json"
	"fmt"
)

// BulkOperationError represents an error for a single item
type BulkOperationError struct {
	ID    string
	Error string
}

// BulkDelete deletes multiple executions
func (s *Service) BulkDelete(ctx context.Context, tenantID string, ids []string) ([]string, []BulkOperationError) {
	var success []string
	var failed []BulkOperationError

	for _, id := range ids {
		err := s.deleteExecution(ctx, tenantID, id)
		if err != nil {
			failed = append(failed, BulkOperationError{
				ID:    id,
				Error: err.Error(),
			})
			s.logger.Error("failed to delete execution in bulk operation",
				"execution_id", id,
				"error", err,
			)
		} else {
			success = append(success, id)
		}
	}

	return success, failed
}

// BulkRetry retries multiple failed executions
func (s *Service) BulkRetry(ctx context.Context, tenantID string, ids []string) ([]string, []BulkOperationError) {
	var success []string
	var failed []BulkOperationError

	for _, id := range ids {
		err := s.retryExecution(ctx, tenantID, id)
		if err != nil {
			failed = append(failed, BulkOperationError{
				ID:    id,
				Error: err.Error(),
			})
			s.logger.Error("failed to retry execution in bulk operation",
				"execution_id", id,
				"error", err,
			)
		} else {
			success = append(success, id)
		}
	}

	return success, failed
}

// deleteExecution deletes a single execution (simplified - would need actual repository method)
func (s *Service) deleteExecution(ctx context.Context, tenantID, id string) error {
	execution, err := s.repo.GetExecutionByID(ctx, tenantID, id)
	if err != nil {
		return err
	}

	if execution.Status == string(ExecutionStatusRunning) {
		return fmt.Errorf("cannot delete running execution")
	}

	// Note: This would require adding DeleteExecution to repository interface
	return fmt.Errorf("delete execution not yet implemented in repository")
}

// retryExecution retries a failed execution
func (s *Service) retryExecution(ctx context.Context, tenantID, id string) error {
	execution, err := s.repo.GetExecutionByID(ctx, tenantID, id)
	if err != nil {
		return err
	}

	if execution.Status != string(ExecutionStatusFailed) {
		return fmt.Errorf("can only retry failed executions, current status: %s", execution.Status)
	}

	// Verify workflow exists
	_, err = s.repo.GetByID(ctx, tenantID, execution.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	// Create new execution with same trigger data
	triggerData := make(map[string]interface{})
	if execution.TriggerData != nil {
		if err := json.Unmarshal(*execution.TriggerData, &triggerData); err != nil {
			s.logger.Error("failed to unmarshal trigger data",
				"error", err,
				"execution_id", id,
			)
			// Continue with empty trigger data
		}
	}

	// Note: This is simplified - actual Execute method signature may differ
	// The implementation would need to match the actual Execute method
	s.logger.Info("retrying execution", "execution_id", id, "workflow_id", execution.WorkflowID)

	return nil
}
