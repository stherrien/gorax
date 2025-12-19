package workflow

import (
	"context"
	"encoding/json"
	"log/slog"
)

// WorkflowExecutor interface to avoid circular dependencies
type WorkflowExecutor interface {
	Execute(ctx context.Context, execution *Execution) error
}

// QueuePublisher interface for publishing execution messages
type QueuePublisher interface {
	PublishExecution(ctx context.Context, msg interface{}) error
}

// RepositoryInterface defines the repository methods used by Service
type RepositoryInterface interface {
	Create(ctx context.Context, tenantID, createdBy string, input CreateWorkflowInput) (*Workflow, error)
	GetByID(ctx context.Context, tenantID, id string) (*Workflow, error)
	Update(ctx context.Context, tenantID, id string, input UpdateWorkflowInput) (*Workflow, error)
	Delete(ctx context.Context, tenantID, id string) error
	List(ctx context.Context, tenantID string, limit, offset int) ([]*Workflow, error)
	CreateExecution(ctx context.Context, tenantID, workflowID string, workflowVersion int, triggerType string, triggerData []byte) (*Execution, error)
	GetExecutionByID(ctx context.Context, tenantID, id string) (*Execution, error)
	GetStepExecutionsByExecutionID(ctx context.Context, executionID string) ([]*StepExecution, error)
	ListExecutions(ctx context.Context, tenantID string, workflowID string, limit, offset int) ([]*Execution, error)
	ListExecutionsAdvanced(ctx context.Context, tenantID string, filter ExecutionFilter, cursor string, limit int) (*ExecutionListResult, error)
	GetExecutionWithSteps(ctx context.Context, tenantID, executionID string) (*ExecutionWithSteps, error)
	CountExecutions(ctx context.Context, tenantID string, filter ExecutionFilter) (int, error)
}

// Service handles workflow business logic
type Service struct {
	repo           RepositoryInterface
	executor       WorkflowExecutor
	webhookService WebhookService
	queuePublisher QueuePublisher
	logger         *slog.Logger
}

// NewService creates a new workflow service
func NewService(repo *Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// SetExecutor sets the workflow executor (called after initialization to avoid import cycles)
func (s *Service) SetExecutor(executor WorkflowExecutor) {
	s.executor = executor
}

// SetWebhookService sets the webhook service (called after both services are created)
func (s *Service) SetWebhookService(webhookService WebhookService) {
	s.webhookService = webhookService
}

// SetQueuePublisher sets the queue publisher (optional, for queue-based execution)
func (s *Service) SetQueuePublisher(publisher QueuePublisher) {
	s.queuePublisher = publisher
}

// Create creates a new workflow
func (s *Service) Create(ctx context.Context, tenantID, userID string, input CreateWorkflowInput) (*Workflow, error) {
	// Validate definition structure
	if err := s.validateDefinition(input.Definition); err != nil {
		return nil, err
	}

	workflow, err := s.repo.Create(ctx, tenantID, userID, input)
	if err != nil {
		s.logger.Error("failed to create workflow", "error", err, "tenant_id", tenantID)
		return nil, err
	}

	// Sync webhooks if webhook service is available
	if s.webhookService != nil {
		webhookNodes := s.extractWebhookNodes(input.Definition)
		if len(webhookNodes) > 0 {
			if err := s.webhookService.SyncWorkflowWebhooks(ctx, tenantID, workflow.ID, webhookNodes); err != nil {
				s.logger.Error("failed to sync webhooks", "error", err, "workflow_id", workflow.ID)
				// Don't fail the workflow creation if webhook sync fails
			}
		}
	}

	s.logger.Info("workflow created", "workflow_id", workflow.ID, "tenant_id", tenantID)
	return workflow, nil
}

// GetByID retrieves a workflow by ID
func (s *Service) GetByID(ctx context.Context, tenantID, id string) (*Workflow, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

// Update updates a workflow
func (s *Service) Update(ctx context.Context, tenantID, id string, input UpdateWorkflowInput) (*Workflow, error) {
	// Validate definition if provided
	if input.Definition != nil {
		if err := s.validateDefinition(input.Definition); err != nil {
			return nil, err
		}
	}

	workflow, err := s.repo.Update(ctx, tenantID, id, input)
	if err != nil {
		s.logger.Error("failed to update workflow", "error", err, "workflow_id", id)
		return nil, err
	}

	// Sync webhooks if definition was updated and webhook service is available
	if input.Definition != nil && s.webhookService != nil {
		webhookNodes := s.extractWebhookNodes(input.Definition)
		if err := s.webhookService.SyncWorkflowWebhooks(ctx, tenantID, workflow.ID, webhookNodes); err != nil {
			s.logger.Error("failed to sync webhooks", "error", err, "workflow_id", workflow.ID)
			// Don't fail the workflow update if webhook sync fails
		}
	}

	s.logger.Info("workflow updated", "workflow_id", workflow.ID, "version", workflow.Version)
	return workflow, nil
}

// Delete deletes a workflow
func (s *Service) Delete(ctx context.Context, tenantID, id string) error {
	// Delete associated webhooks first if webhook service is available
	if s.webhookService != nil {
		if err := s.webhookService.DeleteByWorkflowID(ctx, id); err != nil {
			s.logger.Error("failed to delete webhooks", "error", err, "workflow_id", id)
			// Continue with workflow deletion even if webhook deletion fails
		}
	}

	err := s.repo.Delete(ctx, tenantID, id)
	if err != nil {
		s.logger.Error("failed to delete workflow", "error", err, "workflow_id", id)
		return err
	}

	s.logger.Info("workflow deleted", "workflow_id", id)
	return nil
}

// List retrieves all workflows for a tenant
func (s *Service) List(ctx context.Context, tenantID string, limit, offset int) ([]*Workflow, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.List(ctx, tenantID, limit, offset)
}

// Execute starts a workflow execution
func (s *Service) Execute(ctx context.Context, tenantID, workflowID, triggerType string, triggerData []byte) (*Execution, error) {
	// Get workflow
	workflow, err := s.repo.GetByID(ctx, tenantID, workflowID)
	if err != nil {
		s.logger.Error("failed to get workflow", "error", err, "workflow_id", workflowID)
		return nil, err
	}

	// Validate workflow is active
	if workflow.Status != string(WorkflowStatusActive) {
		s.logger.Warn("attempted to execute inactive workflow",
			"workflow_id", workflowID,
			"status", workflow.Status,
		)
		return nil, &ValidationError{Message: "workflow must be active to execute"}
	}

	// Create execution record
	execution, err := s.repo.CreateExecution(ctx, tenantID, workflowID, workflow.Version, triggerType, triggerData)
	if err != nil {
		s.logger.Error("failed to create execution", "error", err, "workflow_id", workflowID)
		return nil, err
	}

	s.logger.Info("execution created", "execution_id", execution.ID, "workflow_id", workflowID)

	// If queue publisher is configured, publish to queue
	// Otherwise, execute in goroutine (backward compatibility)
	if s.queuePublisher != nil {
		// Import json for RawMessage
		var triggerDataPtr *[]byte
		if triggerData != nil {
			data := []byte(triggerData)
			triggerDataPtr = &data
		}

		// Create execution message for queue
		execMsg := map[string]interface{}{
			"execution_id":     execution.ID,
			"tenant_id":        tenantID,
			"workflow_id":      workflowID,
			"workflow_version": workflow.Version,
			"trigger_type":     triggerType,
		}
		if triggerDataPtr != nil {
			execMsg["trigger_data"] = *triggerDataPtr
		}

		// Publish to queue
		if err := s.queuePublisher.PublishExecution(ctx, execMsg); err != nil {
			s.logger.Error("failed to publish execution to queue",
				"error", err,
				"execution_id", execution.ID,
				"workflow_id", workflowID,
			)
			// Don't fail the request, fall back to goroutine execution
			s.executeInGoroutine(execution, workflowID)
		} else {
			s.logger.Info("execution published to queue", "execution_id", execution.ID, "workflow_id", workflowID)
		}
	} else {
		// Backward compatibility: execute in goroutine
		s.executeInGoroutine(execution, workflowID)
	}

	return execution, nil
}

// executeInGoroutine executes workflow in a goroutine (backward compatibility)
func (s *Service) executeInGoroutine(execution *Execution, workflowID string) {
	go func() {
		// Create a new context for the execution to avoid cancellation
		execCtx := context.Background()

		if s.executor != nil {
			if err := s.executor.Execute(execCtx, execution); err != nil {
				s.logger.Error("workflow execution failed",
					"error", err,
					"execution_id", execution.ID,
					"workflow_id", workflowID,
				)
			}
		} else {
			s.logger.Warn("executor not set, execution will remain in pending state", "execution_id", execution.ID)
		}
	}()
}

// ExecuteSync executes a workflow synchronously (useful for testing)
func (s *Service) ExecuteSync(ctx context.Context, tenantID, workflowID, triggerType string, triggerData []byte) (*Execution, error) {
	// Get workflow
	workflow, err := s.repo.GetByID(ctx, tenantID, workflowID)
	if err != nil {
		s.logger.Error("failed to get workflow", "error", err, "workflow_id", workflowID)
		return nil, err
	}

	// Validate workflow is active
	if workflow.Status != string(WorkflowStatusActive) {
		s.logger.Warn("attempted to execute inactive workflow",
			"workflow_id", workflowID,
			"status", workflow.Status,
		)
		return nil, &ValidationError{Message: "workflow must be active to execute"}
	}

	// Create execution record
	execution, err := s.repo.CreateExecution(ctx, tenantID, workflowID, workflow.Version, triggerType, triggerData)
	if err != nil {
		s.logger.Error("failed to create execution", "error", err, "workflow_id", workflowID)
		return nil, err
	}

	s.logger.Info("execution created", "execution_id", execution.ID, "workflow_id", workflowID)

	// Execute the workflow synchronously
	if s.executor == nil {
		err := &ValidationError{Message: "executor not configured"}
		s.logger.Error("executor not set", "execution_id", execution.ID)
		return nil, err
	}

	if err := s.executor.Execute(ctx, execution); err != nil {
		s.logger.Error("workflow execution failed",
			"error", err,
			"execution_id", execution.ID,
			"workflow_id", workflowID,
		)
		return execution, err
	}

	// Reload execution to get final state
	return s.repo.GetExecutionByID(ctx, tenantID, execution.ID)
}

// GetExecution retrieves an execution by ID
func (s *Service) GetExecution(ctx context.Context, tenantID, executionID string) (*Execution, error) {
	return s.repo.GetExecutionByID(ctx, tenantID, executionID)
}

// GetStepExecutions retrieves all step executions for an execution
func (s *Service) GetStepExecutions(ctx context.Context, executionID string) ([]*StepExecution, error) {
	return s.repo.GetStepExecutionsByExecutionID(ctx, executionID)
}

// ListExecutions retrieves executions for a tenant
func (s *Service) ListExecutions(ctx context.Context, tenantID, workflowID string, limit, offset int) ([]*Execution, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.ListExecutions(ctx, tenantID, workflowID, limit, offset)
}

// ListExecutionsAdvanced retrieves executions with advanced filtering and cursor-based pagination
func (s *Service) ListExecutionsAdvanced(ctx context.Context, tenantID string, filter ExecutionFilter, cursor string, limit int) (*ExecutionListResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	if err := filter.Validate(); err != nil {
		return nil, &ValidationError{Message: "invalid filter: " + err.Error()}
	}

	return s.repo.ListExecutionsAdvanced(ctx, tenantID, filter, cursor, limit)
}

// GetExecutionWithSteps retrieves an execution with all its step executions
func (s *Service) GetExecutionWithSteps(ctx context.Context, tenantID, executionID string) (*ExecutionWithSteps, error) {
	return s.repo.GetExecutionWithSteps(ctx, tenantID, executionID)
}

// GetExecutionStats retrieves execution statistics grouped by status
func (s *Service) GetExecutionStats(ctx context.Context, tenantID string, filter ExecutionFilter) (*ExecutionStats, error) {
	if err := filter.Validate(); err != nil {
		return nil, &ValidationError{Message: "invalid filter: " + err.Error()}
	}

	stats := &ExecutionStats{
		StatusCounts: make(map[string]int),
	}

	statuses := []ExecutionStatus{
		ExecutionStatusPending,
		ExecutionStatusRunning,
		ExecutionStatusCompleted,
		ExecutionStatusFailed,
		ExecutionStatusCancelled,
	}

	for _, status := range statuses {
		statusFilter := filter
		statusFilter.Status = string(status)

		count, err := s.repo.CountExecutions(ctx, tenantID, statusFilter)
		if err != nil {
			return nil, err
		}

		stats.StatusCounts[string(status)] = count
		stats.TotalCount += count
	}

	return stats, nil
}

// validateDefinition validates a workflow definition
func (s *Service) validateDefinition(definition json.RawMessage) error {
	var def WorkflowDefinition
	if err := json.Unmarshal(definition, &def); err != nil {
		return &ValidationError{Message: "invalid definition JSON: " + err.Error()}
	}

	// Validate nodes exist
	if len(def.Nodes) == 0 {
		return &ValidationError{Message: "workflow must have at least one node"}
	}

	// Validate at least one trigger
	hasTrigger := false
	nodeIDs := make(map[string]bool)
	for _, node := range def.Nodes {
		nodeIDs[node.ID] = true
		if node.Type == string(NodeTypeTriggerWebhook) || node.Type == string(NodeTypeTriggerSchedule) {
			hasTrigger = true
		}
	}

	if !hasTrigger {
		return &ValidationError{Message: "workflow must have at least one trigger"}
	}

	// Validate edges reference existing nodes
	for _, edge := range def.Edges {
		if !nodeIDs[edge.Source] {
			return &ValidationError{Message: "edge references non-existent source node: " + edge.Source}
		}
		if !nodeIDs[edge.Target] {
			return &ValidationError{Message: "edge references non-existent target node: " + edge.Target}
		}
	}

	return nil
}

// extractWebhookNodes extracts webhook trigger nodes from a workflow definition
func (s *Service) extractWebhookNodes(definition json.RawMessage) []WebhookNodeConfig {
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
			if node.Config != nil {
				if err := json.Unmarshal(node.Config, &config); err == nil {
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

// GetWebhooks retrieves all webhooks for a workflow
func (s *Service) GetWebhooks(ctx context.Context, workflowID string) ([]*WebhookInfo, error) {
	if s.webhookService == nil {
		return nil, nil
	}
	return s.webhookService.GetByWorkflowID(ctx, workflowID)
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// AuthType constants
const (
	AuthTypeNone      = "none"
	AuthTypeSignature = "signature"
	AuthTypeBasic     = "basic"
	AuthTypeAPIKey    = "api_key"
)
