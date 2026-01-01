package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

// BulkOperationResult represents the result of a bulk operation
type BulkOperationResult struct {
	SuccessCount int                    `json:"success_count"`
	Failures     []BulkOperationFailure `json:"failures"`
}

// BulkOperationFailure represents a failure in a bulk operation
type BulkOperationFailure struct {
	WorkflowID string `json:"workflow_id"`
	Error      string `json:"error"`
}

// WorkflowExport represents a workflow export structure
type WorkflowExport struct {
	Workflows  []WorkflowExportItem `json:"workflows"`
	ExportedAt time.Time            `json:"exported_at"`
	Version    string               `json:"version"`
}

// WorkflowExportItem represents a single workflow in an export
type WorkflowExportItem struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Definition  json.RawMessage `json:"definition"`
	Status      string          `json:"status"`
	Version     int             `json:"version"`
}

// BulkService handles bulk workflow operations
type BulkService struct {
	repo           RepositoryInterface
	webhookService WebhookService
	logger         *slog.Logger
}

// NewBulkService creates a new bulk service
func NewBulkService(repo RepositoryInterface, webhookService WebhookService, logger *slog.Logger) *BulkService {
	return &BulkService{
		repo:           repo,
		webhookService: webhookService,
		logger:         logger,
	}
}

// BulkDelete deletes multiple workflows
func (s *BulkService) BulkDelete(ctx context.Context, tenantID string, workflowIDs []string) BulkOperationResult {
	result := BulkOperationResult{
		SuccessCount: 0,
		Failures:     make([]BulkOperationFailure, 0),
	}

	for _, workflowID := range workflowIDs {
		if err := s.deleteWorkflow(ctx, tenantID, workflowID); err != nil {
			result.Failures = append(result.Failures, BulkOperationFailure{
				WorkflowID: workflowID,
				Error:      err.Error(),
			})
		} else {
			result.SuccessCount++
		}
	}

	s.logger.Info("bulk delete completed",
		"success_count", result.SuccessCount,
		"failure_count", len(result.Failures),
		"tenant_id", tenantID,
	)

	return result
}

// deleteWorkflow deletes a single workflow with webhooks
func (s *BulkService) deleteWorkflow(ctx context.Context, tenantID, workflowID string) error {
	// Delete associated webhooks first if webhook service is available
	if s.webhookService != nil {
		if err := s.webhookService.DeleteByWorkflowID(ctx, workflowID); err != nil {
			s.logger.Error("failed to delete webhooks", "error", err, "workflow_id", workflowID)
			// Continue with workflow deletion even if webhook deletion fails
		}
	}

	// Delete the workflow
	if err := s.repo.Delete(ctx, tenantID, workflowID); err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	return nil
}

// BulkEnable enables multiple workflows
func (s *BulkService) BulkEnable(ctx context.Context, tenantID string, workflowIDs []string) BulkOperationResult {
	return s.bulkUpdateStatus(ctx, tenantID, workflowIDs, "active", "enable")
}

// BulkDisable disables multiple workflows
func (s *BulkService) BulkDisable(ctx context.Context, tenantID string, workflowIDs []string) BulkOperationResult {
	return s.bulkUpdateStatus(ctx, tenantID, workflowIDs, "inactive", "disable")
}

// bulkUpdateStatus updates the status of multiple workflows
func (s *BulkService) bulkUpdateStatus(ctx context.Context, tenantID string, workflowIDs []string, status, operation string) BulkOperationResult {
	result := BulkOperationResult{
		SuccessCount: 0,
		Failures:     make([]BulkOperationFailure, 0),
	}

	for _, workflowID := range workflowIDs {
		input := UpdateWorkflowInput{
			Status: status,
		}

		_, err := s.repo.Update(ctx, tenantID, workflowID, input)
		if err != nil {
			result.Failures = append(result.Failures, BulkOperationFailure{
				WorkflowID: workflowID,
				Error:      err.Error(),
			})
		} else {
			result.SuccessCount++
		}
	}

	s.logger.Info("bulk "+operation+" completed",
		"success_count", result.SuccessCount,
		"failure_count", len(result.Failures),
		"tenant_id", tenantID,
	)

	return result
}

// BulkExport exports multiple workflows as JSON
func (s *BulkService) BulkExport(ctx context.Context, tenantID string, workflowIDs []string) (WorkflowExport, BulkOperationResult) {
	export := WorkflowExport{
		Workflows:  make([]WorkflowExportItem, 0),
		ExportedAt: time.Now(),
		Version:    "1.0",
	}

	result := BulkOperationResult{
		SuccessCount: 0,
		Failures:     make([]BulkOperationFailure, 0),
	}

	for _, workflowID := range workflowIDs {
		workflow, err := s.repo.GetByID(ctx, tenantID, workflowID)
		if err != nil {
			result.Failures = append(result.Failures, BulkOperationFailure{
				WorkflowID: workflowID,
				Error:      err.Error(),
			})
			continue
		}

		export.Workflows = append(export.Workflows, WorkflowExportItem{
			ID:          workflow.ID,
			Name:        workflow.Name,
			Description: workflow.Description,
			Definition:  workflow.Definition,
			Status:      workflow.Status,
			Version:     workflow.Version,
		})

		result.SuccessCount++
	}

	s.logger.Info("bulk export completed",
		"success_count", result.SuccessCount,
		"failure_count", len(result.Failures),
		"tenant_id", tenantID,
	)

	return export, result
}

// BulkClone clones multiple workflows
func (s *BulkService) BulkClone(ctx context.Context, tenantID, userID string, workflowIDs []string) ([]*Workflow, BulkOperationResult) {
	clones := make([]*Workflow, 0)
	result := BulkOperationResult{
		SuccessCount: 0,
		Failures:     make([]BulkOperationFailure, 0),
	}

	for _, workflowID := range workflowIDs {
		clone, err := s.cloneWorkflow(ctx, tenantID, userID, workflowID)
		if err != nil {
			result.Failures = append(result.Failures, BulkOperationFailure{
				WorkflowID: workflowID,
				Error:      err.Error(),
			})
			continue
		}

		clones = append(clones, clone)
		result.SuccessCount++
	}

	s.logger.Info("bulk clone completed",
		"success_count", result.SuccessCount,
		"failure_count", len(result.Failures),
		"tenant_id", tenantID,
	)

	return clones, result
}

// cloneWorkflow clones a single workflow
func (s *BulkService) cloneWorkflow(ctx context.Context, tenantID, userID, workflowID string) (*Workflow, error) {
	// Get the original workflow
	original, err := s.repo.GetByID(ctx, tenantID, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original workflow: %w", err)
	}

	// Create input for the clone
	input := CreateWorkflowInput{
		Name:        original.Name + " (Copy)",
		Description: original.Description,
		Definition:  original.Definition,
	}

	// Create the clone with draft status
	clone, err := s.repo.Create(ctx, tenantID, userID, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create clone: %w", err)
	}

	// Sync webhooks if webhook service is available
	if s.webhookService != nil {
		webhookNodes := s.extractWebhookNodes(original.Definition)
		if len(webhookNodes) > 0 {
			if err := s.webhookService.SyncWorkflowWebhooks(ctx, tenantID, clone.ID, webhookNodes); err != nil {
				s.logger.Error("failed to sync webhooks for clone",
					"error", err,
					"clone_id", clone.ID,
					"original_id", workflowID,
				)
				// Don't fail the clone operation if webhook sync fails
			}
		}
	}

	return clone, nil
}

// extractWebhookNodes extracts webhook trigger nodes from a workflow definition
func (s *BulkService) extractWebhookNodes(definition json.RawMessage) []WebhookNodeConfig {
	var def WorkflowDefinition
	if err := json.Unmarshal(definition, &def); err != nil {
		return nil
	}

	var webhookNodes []WebhookNodeConfig
	for _, node := range def.Nodes {
		if node.Type == string(NodeTypeTriggerWebhook) {
			// Parse config to get auth type
			var config WebhookTriggerConfig
			authType := AuthTypeSignature // Default
			if node.Data.Config != nil {
				if err := json.Unmarshal(node.Data.Config, &config); err == nil {
					if config.AuthType != "" {
						authType = config.AuthType
					}
				}
			}

			webhookNodes = append(webhookNodes, WebhookNodeConfig{
				NodeID:   node.ID,
				AuthType: authType,
			})
		}
	}

	return webhookNodes
}
