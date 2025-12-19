package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
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
	CreateWorkflowVersion(ctx context.Context, workflowID string, version int, definition json.RawMessage, createdBy string) (*WorkflowVersion, error)
	ListWorkflowVersions(ctx context.Context, workflowID string) ([]*WorkflowVersion, error)
	GetWorkflowVersion(ctx context.Context, workflowID string, version int) (*WorkflowVersion, error)
	RestoreWorkflowVersion(ctx context.Context, tenantID, workflowID string, version int) (*Workflow, error)
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
	// Create version record if definition was updated
	if input.Definition != nil {
		_, err := s.repo.CreateWorkflowVersion(ctx, workflow.ID, workflow.Version, workflow.Definition, workflow.CreatedBy)
		if err != nil {
			s.logger.Error("failed to create workflow version", "error", err, "workflow_id", workflow.ID, "version", workflow.Version)
			// Don't fail the workflow update if version creation fails
		}
	}

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

// ListWorkflowVersions retrieves all versions for a workflow
func (s *Service) ListWorkflowVersions(ctx context.Context, workflowID string) ([]*WorkflowVersion, error) {
	return s.repo.ListWorkflowVersions(ctx, workflowID)
}

// GetWorkflowVersion retrieves a specific version of a workflow
func (s *Service) GetWorkflowVersion(ctx context.Context, workflowID string, version int) (*WorkflowVersion, error) {
	return s.repo.GetWorkflowVersion(ctx, workflowID, version)
}

// RestoreWorkflowVersion restores a workflow to a previous version
func (s *Service) RestoreWorkflowVersion(ctx context.Context, tenantID, workflowID string, version int) (*Workflow, error) {
	// Validate workflow exists
	workflow, err := s.repo.GetByID(ctx, tenantID, workflowID)
	if err != nil {
		return nil, err
	}

	// Validate version exists
	versionData, err := s.repo.GetWorkflowVersion(ctx, workflowID, version)
	if err != nil {
		return nil, err
	}

	// Restore the version
	restoredWorkflow, err := s.repo.RestoreWorkflowVersion(ctx, tenantID, workflowID, version)
	if err != nil {
		s.logger.Error("failed to restore workflow version",
			"error", err,
			"workflow_id", workflowID,
			"version", version,
		)
		return nil, err
	}

	// Create version record for the restored state
	_, err = s.repo.CreateWorkflowVersion(ctx, workflowID, restoredWorkflow.Version, restoredWorkflow.Definition, workflow.CreatedBy)
	if err != nil {
		s.logger.Error("failed to create version record after restore",
			"error", err,
			"workflow_id", workflowID,
			"version", restoredWorkflow.Version,
		)
		// Don't fail the restore if version record creation fails
	}

	// Sync webhooks if webhook service is available
	if s.webhookService != nil {
		webhookNodes := s.extractWebhookNodes(versionData.Definition)
		if err := s.webhookService.SyncWorkflowWebhooks(ctx, tenantID, workflowID, webhookNodes); err != nil {
			s.logger.Error("failed to sync webhooks after restore", "error", err, "workflow_id", workflowID)
			// Don't fail the restore if webhook sync fails
		}
	}

	s.logger.Info("workflow version restored",
		"workflow_id", workflowID,
		"restored_from_version", version,
		"new_version", restoredWorkflow.Version,
	)

	return restoredWorkflow, nil
}

// DryRun performs a dry-run validation of a workflow without executing it
func (s *Service) DryRun(ctx context.Context, tenantID, workflowID string, testData map[string]interface{}) (*DryRunResult, error) {
	result := &DryRunResult{
		Valid:           true,
		ExecutionOrder:  []string{},
		VariableMapping: make(map[string]string),
		Warnings:        []DryRunWarning{},
		Errors:          []DryRunError{},
	}

	workflow, err := s.repo.GetByID(ctx, tenantID, workflowID)
	if err != nil {
		return nil, err
	}

	var definition WorkflowDefinition
	if err := json.Unmarshal(workflow.Definition, &definition); err != nil {
		return nil, &ValidationError{Message: "failed to parse workflow definition: " + err.Error()}
	}

	if len(definition.Nodes) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, DryRunError{
			NodeID:  "",
			Field:   "nodes",
			Message: "workflow has no nodes",
		})
		return result, nil
	}

	nodeMap := s.buildNodeMapForDryRun(definition.Nodes)

	executionOrder, err := s.validateTopologicalOrder(definition.Nodes, definition.Edges)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, DryRunError{
			NodeID:  "",
			Field:   "edges",
			Message: err.Error(),
		})
		return result, nil
	}
	result.ExecutionOrder = executionOrder

	availableVars := make(map[string]bool)
	if testData != nil {
		for key := range testData {
			availableVars["trigger."+key] = true
			result.VariableMapping["trigger."+key] = "test_data"
		}
	}
	availableVars["trigger"] = true
	result.VariableMapping["trigger"] = "trigger_data"

	for _, nodeID := range executionOrder {
		node, exists := nodeMap[nodeID]
		if !exists {
			continue
		}

		if s.isTriggerNodeType(node.Type) {
			continue
		}

		nodeErrors := s.validateNodeConfig(node, availableVars)
		if len(nodeErrors) > 0 {
			result.Valid = false
			result.Errors = append(result.Errors, nodeErrors...)
		}

		nodeWarnings := s.validateNodeWarnings(node)
		if len(nodeWarnings) > 0 {
			result.Warnings = append(result.Warnings, nodeWarnings...)
		}

		availableVars["steps."+nodeID] = true
		result.VariableMapping["steps."+nodeID] = "node:" + nodeID

		if node.Type == string(NodeTypeControlLoop) {
			var loopConfig LoopActionConfig
			if len(node.Data.Config) > 0 {
				if err := json.Unmarshal(node.Data.Config, &loopConfig); err == nil {
					if loopConfig.ItemVariable != "" {
						availableVars[loopConfig.ItemVariable] = true
						result.VariableMapping[loopConfig.ItemVariable] = "loop:"+nodeID
					}
					if loopConfig.IndexVariable != "" {
						availableVars[loopConfig.IndexVariable] = true
						result.VariableMapping[loopConfig.IndexVariable] = "loop:"+nodeID
					}
				}
			}
		}
	}

	return result, nil
}

func (s *Service) buildNodeMapForDryRun(nodes []Node) map[string]Node {
	nodeMap := make(map[string]Node)
	for _, node := range nodes {
		nodeMap[node.ID] = node
	}
	return nodeMap
}

func (s *Service) isTriggerNodeType(nodeType string) bool {
	return nodeType == string(NodeTypeTriggerWebhook) ||
		nodeType == string(NodeTypeTriggerSchedule)
}

func (s *Service) validateTopologicalOrder(nodes []Node, edges []Edge) ([]string, error) {
	inDegree := make(map[string]int)
	adjList := make(map[string][]string)

	for _, node := range nodes {
		inDegree[node.ID] = 0
		adjList[node.ID] = []string{}
	}

	for _, edge := range edges {
		adjList[edge.Source] = append(adjList[edge.Source], edge.Target)
		inDegree[edge.Target]++
	}

	var queue []string
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	var result []string
	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		result = append(result, nodeID)

		for _, neighbor := range adjList[nodeID] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(result) != len(nodes) {
		return nil, &ValidationError{Message: "workflow contains cycles"}
	}

	return result, nil
}

func (s *Service) validateNodeConfig(node Node, availableVars map[string]bool) []DryRunError {
	var errors []DryRunError

	switch node.Type {
	case string(NodeTypeActionHTTP):
		errors = append(errors, s.validateHTTPConfig(node, availableVars)...)
	case string(NodeTypeActionTransform):
		errors = append(errors, s.validateTransformConfig(node, availableVars)...)
	case string(NodeTypeActionFormula):
		errors = append(errors, s.validateFormulaConfig(node, availableVars)...)
	case string(NodeTypeControlIf):
		errors = append(errors, s.validateConditionalConfig(node, availableVars)...)
	case string(NodeTypeControlLoop):
		errors = append(errors, s.validateLoopConfig(node, availableVars)...)
	}

	return errors
}

func (s *Service) validateHTTPConfig(node Node, availableVars map[string]bool) []DryRunError {
	var errors []DryRunError
	var config HTTPActionConfig

	if len(node.Data.Config) == 0 {
		errors = append(errors, DryRunError{
			NodeID:  node.ID,
			Field:   "config",
			Message: "HTTP action requires configuration",
		})
		return errors
	}

	if err := json.Unmarshal(node.Data.Config, &config); err != nil {
		errors = append(errors, DryRunError{
			NodeID:  node.ID,
			Field:   "config",
			Message: "invalid HTTP configuration: " + err.Error(),
		})
		return errors
	}

	if config.Method == "" {
		errors = append(errors, DryRunError{
			NodeID:  node.ID,
			Field:   "method",
			Message: "HTTP method is required",
		})
	}

	if config.URL == "" {
		errors = append(errors, DryRunError{
			NodeID:  node.ID,
			Field:   "url",
			Message: "URL is required",
		})
	}

	configStr := string(node.Data.Config)
	errors = append(errors, s.validateVariableReferences(node.ID, configStr, availableVars)...)

	return errors
}

func (s *Service) validateTransformConfig(node Node, availableVars map[string]bool) []DryRunError {
	var errors []DryRunError
	var config TransformActionConfig

	if len(node.Data.Config) == 0 {
		return errors
	}

	if err := json.Unmarshal(node.Data.Config, &config); err != nil {
		errors = append(errors, DryRunError{
			NodeID:  node.ID,
			Field:   "config",
			Message: "invalid transform configuration: " + err.Error(),
		})
		return errors
	}

	configStr := string(node.Data.Config)
	errors = append(errors, s.validateVariableReferences(node.ID, configStr, availableVars)...)

	return errors
}

func (s *Service) validateFormulaConfig(node Node, availableVars map[string]bool) []DryRunError {
	var errors []DryRunError
	var config FormulaActionConfig

	if len(node.Data.Config) == 0 {
		errors = append(errors, DryRunError{
			NodeID:  node.ID,
			Field:   "config",
			Message: "formula action requires configuration",
		})
		return errors
	}

	if err := json.Unmarshal(node.Data.Config, &config); err != nil {
		errors = append(errors, DryRunError{
			NodeID:  node.ID,
			Field:   "config",
			Message: "invalid formula configuration: " + err.Error(),
		})
		return errors
	}

	if config.Expression == "" {
		errors = append(errors, DryRunError{
			NodeID:  node.ID,
			Field:   "expression",
			Message: "expression is required",
		})
	}

	configStr := string(node.Data.Config)
	errors = append(errors, s.validateVariableReferences(node.ID, configStr, availableVars)...)

	return errors
}

func (s *Service) validateConditionalConfig(node Node, availableVars map[string]bool) []DryRunError {
	var errors []DryRunError
	var config ConditionalActionConfig

	if len(node.Data.Config) == 0 {
		errors = append(errors, DryRunError{
			NodeID:  node.ID,
			Field:   "config",
			Message: "conditional action requires configuration",
		})
		return errors
	}

	if err := json.Unmarshal(node.Data.Config, &config); err != nil {
		errors = append(errors, DryRunError{
			NodeID:  node.ID,
			Field:   "config",
			Message: "invalid conditional configuration: " + err.Error(),
		})
		return errors
	}

	if config.Condition == "" {
		errors = append(errors, DryRunError{
			NodeID:  node.ID,
			Field:   "condition",
			Message: "condition is required",
		})
	}

	configStr := string(node.Data.Config)
	errors = append(errors, s.validateVariableReferences(node.ID, configStr, availableVars)...)

	return errors
}

func (s *Service) validateLoopConfig(node Node, availableVars map[string]bool) []DryRunError {
	var errors []DryRunError
	var config LoopActionConfig

	if len(node.Data.Config) == 0 {
		errors = append(errors, DryRunError{
			NodeID:  node.ID,
			Field:   "config",
			Message: "loop action requires configuration",
		})
		return errors
	}

	if err := json.Unmarshal(node.Data.Config, &config); err != nil {
		errors = append(errors, DryRunError{
			NodeID:  node.ID,
			Field:   "config",
			Message: "invalid loop configuration: " + err.Error(),
		})
		return errors
	}

	if config.Source == "" {
		errors = append(errors, DryRunError{
			NodeID:  node.ID,
			Field:   "source",
			Message: "loop source is required",
		})
	}

	if config.ItemVariable == "" {
		errors = append(errors, DryRunError{
			NodeID:  node.ID,
			Field:   "item_variable",
			Message: "item variable name is required",
		})
	}

	configStr := string(node.Data.Config)
	errors = append(errors, s.validateVariableReferences(node.ID, configStr, availableVars)...)

	return errors
}

func (s *Service) validateVariableReferences(nodeID, configStr string, availableVars map[string]bool) []DryRunError {
	var errors []DryRunError

	varPattern := `\$\{([^}]+)\}`
	re := regexp.MustCompile(varPattern)
	matches := re.FindAllStringSubmatch(configStr, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		varName := match[1]

		parts := strings.Split(varName, ".")
		if len(parts) == 0 {
			continue
		}

		rootVar := parts[0]
		if len(parts) > 1 {
			rootVar = parts[0] + "." + parts[1]
		}

		if !availableVars[parts[0]] && !availableVars[rootVar] {
			errors = append(errors, DryRunError{
				NodeID:  nodeID,
				Field:   "mapping",
				Message: fmt.Sprintf("undefined variable reference: %s", varName),
			})
		}
	}

	return errors
}

func (s *Service) validateNodeWarnings(node Node) []DryRunWarning {
	var warnings []DryRunWarning

	var configMap map[string]interface{}
	if len(node.Data.Config) > 0 {
		if err := json.Unmarshal(node.Data.Config, &configMap); err == nil {
			if credID, exists := configMap["credential_id"]; exists && credID != nil {
				warnings = append(warnings, DryRunWarning{
					NodeID:  node.ID,
					Message: "references credential (not validated during dry-run)",
				})
			}
		}
	}

	return warnings
}
