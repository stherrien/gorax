package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/workflow"
)

// MaxSubWorkflowDepth is the maximum depth of nested sub-workflow execution
const MaxSubWorkflowDepth = 10

// Broadcaster defines the interface for broadcasting execution events
type Broadcaster interface {
	BroadcastExecutionStarted(tenantID, workflowID, executionID string, totalSteps int)
	BroadcastExecutionCompleted(tenantID, workflowID, executionID string, output json.RawMessage)
	BroadcastExecutionFailed(tenantID, workflowID, executionID string, errorMsg string)
	BroadcastStepStarted(tenantID, workflowID, executionID, nodeID, nodeType string)
	BroadcastStepCompleted(tenantID, workflowID, executionID, nodeID string, output json.RawMessage, durationMs int)
	BroadcastStepFailed(tenantID, workflowID, executionID, nodeID string, errorMsg string)
	BroadcastProgress(tenantID, workflowID, executionID string, completedSteps, totalSteps int)
}

// Executor handles workflow execution
type Executor struct {
	repo               *workflow.Repository
	logger             *slog.Logger
	broadcaster        Broadcaster
	retryStrategy      *RetryStrategy
	circuitBreakers    *CircuitBreakerRegistry
	defaultRetryConfig NodeRetryConfig
	credentialInjector *credential.Injector // Optional credential injector
	credentialService  credential.Service   // Optional credential service for Slack actions
}

// New creates a new executor without broadcasting
func New(repo *workflow.Repository, logger *slog.Logger) *Executor {
	retryConfig := DefaultRetryConfig()
	circuitConfig := DefaultCircuitBreakerConfig()

	return &Executor{
		repo:               repo,
		logger:             logger,
		broadcaster:        nil,
		retryStrategy:      NewRetryStrategy(retryConfig, logger),
		circuitBreakers:    NewCircuitBreakerRegistry(circuitConfig, logger),
		defaultRetryConfig: DefaultNodeRetryConfig(),
	}
}

// NewWithBroadcaster creates a new executor with event broadcasting
func NewWithBroadcaster(repo *workflow.Repository, logger *slog.Logger, broadcaster Broadcaster) *Executor {
	retryConfig := DefaultRetryConfig()
	circuitConfig := DefaultCircuitBreakerConfig()

	return &Executor{
		repo:               repo,
		logger:             logger,
		broadcaster:        broadcaster,
		retryStrategy:      NewRetryStrategy(retryConfig, logger),
		circuitBreakers:    NewCircuitBreakerRegistry(circuitConfig, logger),
		defaultRetryConfig: DefaultNodeRetryConfig(),
	}
}

// NewWithCredentials creates a new executor with credential injection support
func NewWithCredentials(repo *workflow.Repository, logger *slog.Logger, broadcaster Broadcaster, injector *credential.Injector, credService credential.Service) *Executor {
	retryConfig := DefaultRetryConfig()
	circuitConfig := DefaultCircuitBreakerConfig()

	return &Executor{
		repo:               repo,
		logger:             logger,
		broadcaster:        broadcaster,
		retryStrategy:      NewRetryStrategy(retryConfig, logger),
		circuitBreakers:    NewCircuitBreakerRegistry(circuitConfig, logger),
		defaultRetryConfig: DefaultNodeRetryConfig(),
		credentialInjector: injector,
		credentialService:  credService,
	}
}

// ExecutionContext holds context for a workflow execution
type ExecutionContext struct {
	TenantID          string
	ExecutionID       string
	WorkflowID        string
	TriggerType       string
	TriggerData       map[string]interface{}
	StepOutputs       map[string]interface{}
	CredentialValues  []string // Decrypted credential values for masking
	UserID            string   // User who triggered the execution
	Depth             int      // Execution depth for sub-workflow tracking
	WorkflowChain     []string // Chain of workflow IDs to detect circular dependencies
	ParentExecutionID string   // Parent execution ID for sub-workflows
}

// GetUserID returns the user ID from the execution context
// Returns "system" if no user ID is found (for automated executions)
func (ec *ExecutionContext) GetUserID() string {
	// If UserID is already set, use it
	if ec.UserID != "" {
		return ec.UserID
	}

	// Try to extract from trigger data
	if ec.TriggerData != nil {
		// Check for user_id at top level (manual triggers)
		if userID, ok := ec.TriggerData["user_id"].(string); ok && userID != "" {
			return userID
		}

		// Check for user_id in auth context (webhook triggers)
		if authData, ok := ec.TriggerData["_auth"].(map[string]interface{}); ok {
			if userID, ok := authData["user_id"].(string); ok && userID != "" {
				return userID
			}
		}
	}

	// Default to system for automated executions (schedules, etc.)
	return "system"
}

// SetUserID sets the user ID in the execution context
func (ec *ExecutionContext) SetUserID(userID string) {
	ec.UserID = userID
}

// Execute runs a workflow execution
func (e *Executor) Execute(ctx context.Context, execution *workflow.Execution) error {
	e.logger.Info("starting workflow execution",
		"execution_id", execution.ID,
		"workflow_id", execution.WorkflowID,
	)

	// Update status to running
	if err := e.repo.UpdateExecutionStatus(ctx, execution.ID, workflow.ExecutionStatusRunning, nil, nil); err != nil {
		return err
	}

	// Load workflow definition
	wf, err := e.repo.GetByID(ctx, execution.TenantID, execution.WorkflowID)
	if err != nil {
		return e.failExecution(ctx, execution)
	}

	// Parse workflow definition
	var definition workflow.WorkflowDefinition
	if err := json.Unmarshal(wf.Definition, &definition); err != nil {
		return e.failExecution(ctx, execution, fmt.Errorf("failed to parse workflow definition: %w", err))
	}

	// Validate workflow has nodes
	if len(definition.Nodes) == 0 {
		return e.failExecution(ctx, execution, fmt.Errorf("workflow has no nodes to execute"))
	}

	// Count non-trigger nodes for progress tracking
	totalSteps := 0
	for _, node := range definition.Nodes {
		if !isTriggerNode(node.Type) {
			totalSteps++
		}
	}

	// Broadcast execution started with total steps
	if e.broadcaster != nil {
		e.broadcaster.BroadcastExecutionStarted(execution.TenantID, execution.WorkflowID, execution.ID, totalSteps)
	}

	// Parse trigger data
	var triggerData map[string]interface{}
	if execution.TriggerData != nil {
		if err := json.Unmarshal(*execution.TriggerData, &triggerData); err != nil {
			triggerData = make(map[string]interface{})
		}
	} else {
		triggerData = make(map[string]interface{})
	}

	// Create execution context
	execCtx := &ExecutionContext{
		TenantID:          execution.TenantID,
		ExecutionID:       execution.ID,
		WorkflowID:        execution.WorkflowID,
		TriggerData:       triggerData,
		StepOutputs:       make(map[string]interface{}),
		CredentialValues:  []string{}, // Will be populated during execution
		Depth:             execution.ExecutionDepth,
		WorkflowChain:     []string{execution.WorkflowID},
		ParentExecutionID: "",
	}

	// Set parent execution ID if this is a sub-workflow
	if execution.ParentExecutionID != nil {
		execCtx.ParentExecutionID = *execution.ParentExecutionID
	}

	// Build execution order from DAG
	nodeMap := buildNodeMap(definition.Nodes)
	executionOrder, err := topologicalSort(definition.Nodes, definition.Edges)
	if err != nil {
		return e.failExecution(ctx, execution, fmt.Errorf("failed to determine execution order: %w", err))
	}

	e.logger.Info("determined execution order",
		"execution_id", execution.ID,
		"node_count", len(executionOrder),
		"order", executionOrder,
	)

	// Execute nodes in order
	completedSteps := 0
	for _, nodeID := range executionOrder {
		node, exists := nodeMap[nodeID]
		if !exists {
			e.logger.Warn("node not found in map", "node_id", nodeID)
			continue
		}

		e.logger.Info("executing node", "node_id", node.ID, "node_type", node.Type)

		// Skip triggers (they've already fired)
		if isTriggerNode(node.Type) {
			execCtx.StepOutputs[node.ID] = triggerData
			e.logger.Info("skipping trigger node", "node_id", node.ID)
			continue
		}

		// Broadcast step started
		if e.broadcaster != nil {
			e.broadcaster.BroadcastStepStarted(execution.TenantID, execution.WorkflowID, execution.ID, node.ID, node.Type)
		}

		// Execute the node with step tracking
		startTime := time.Now()
		var output interface{}
		var err error

		// Handle control nodes specially (they need workflow definition)
		if node.Type == string(workflow.NodeTypeControlLoop) {
			output, err = e.executeLoopAction(ctx, node, execCtx, &definition)
		} else if node.Type == string(workflow.NodeTypeControlParallel) {
			output, err = e.executeParallelAction(ctx, node, execCtx, &definition)
		} else if node.Type == string(workflow.NodeTypeControlFork) {
			output, err = e.executeForkAction(ctx, node, execCtx)
		} else if node.Type == string(workflow.NodeTypeControlJoin) {
			output, err = e.executeJoinAction(ctx, node, execCtx, &definition)
		} else {
			output, err = e.executeNodeWithTracking(ctx, node, execCtx)
		}
		durationMs := int(time.Since(startTime).Milliseconds())

		if err != nil {
			e.logger.Error("node execution failed", "node_id", node.ID, "error", err)
			// Broadcast step failure
			if e.broadcaster != nil {
				e.broadcaster.BroadcastStepFailed(execution.TenantID, execution.WorkflowID, execution.ID, node.ID, err.Error())
			}
			return e.failExecution(ctx, execution, fmt.Errorf("node %s failed: %w", node.ID, err))
		}

		// Store output for downstream nodes
		execCtx.StepOutputs[node.ID] = output

		// Broadcast step completion
		if e.broadcaster != nil {
			outputJSON, _ := json.Marshal(output)
			e.broadcaster.BroadcastStepCompleted(execution.TenantID, execution.WorkflowID, execution.ID, node.ID, outputJSON, durationMs)
		}

		// Update and broadcast progress
		completedSteps++
		if e.broadcaster != nil {
			e.broadcaster.BroadcastProgress(execution.TenantID, execution.WorkflowID, execution.ID, completedSteps, totalSteps)
		}
	}

	// Mark execution as completed
	outputData, _ := json.Marshal(execCtx.StepOutputs)
	if err := e.repo.UpdateExecutionStatus(ctx, execution.ID, workflow.ExecutionStatusCompleted, outputData, nil); err != nil {
		return err
	}

	// Broadcast execution completed
	if e.broadcaster != nil {
		e.broadcaster.BroadcastExecutionCompleted(execution.TenantID, execution.WorkflowID, execution.ID, outputData)
	}

	e.logger.Info("workflow execution completed", "execution_id", execution.ID)
	return nil
}

// executeNodeWithTracking executes a node and tracks the execution in the database
func (e *Executor) executeNodeWithTracking(ctx context.Context, node workflow.Node, execCtx *ExecutionContext) (interface{}, error) {
	// Prepare input data for the node
	inputData := buildInputData(execCtx)
	inputDataJSON, _ := json.Marshal(inputData)

	// Create step execution record
	stepExecution, err := e.repo.CreateStepExecution(ctx, execCtx.ExecutionID, node.ID, node.Type, inputDataJSON)
	if err != nil {
		e.logger.Error("failed to create step execution record", "error", err, "node_id", node.ID)
		// Continue execution even if we can't track it
	}

	// Parse retry configuration from node data if available
	retryConfig := e.defaultRetryConfig
	// Try to parse retry config from node data
	var configMap map[string]interface{}
	if len(node.Data.Config) > 0 {
		if err := json.Unmarshal(node.Data.Config, &configMap); err == nil {
			if retryData, exists := configMap["retry"]; exists {
				// Parse custom retry config from node
				retryConfig = e.parseRetryConfig(retryData)
			}
		}
	}

	// Execute the node with retry logic
	var output interface{}
	var execErr error
	retryCount := 0

	if retryConfig.Enabled {
		// Create retry strategy for this node
		nodeRetryStrategy := NewRetryStrategy(retryConfig.RetryConfig, e.logger)

		// Execute with retry
		result, err := nodeRetryStrategy.ExecuteWithResult(ctx, func(ctx context.Context, attempt int) (interface{}, error) {
			retryCount = attempt
			return e.executeNode(ctx, node, execCtx)
		})
		output = result
		execErr = err
	} else {
		// Execute without retry
		output, execErr = e.executeNode(ctx, node, execCtx)
	}

	// Wrap error with execution context
	if execErr != nil {
		execErr = WrapError(execErr, node.ID, node.Type, retryCount)
	}

	// Update step execution record
	if stepExecution != nil {
		var status string
		var errorMsg *string

		if execErr != nil {
			status = "failed"
			errStr := execErr.Error()
			errorMsg = &errStr

			// Add error classification to the error message
			if execError, ok := execErr.(*ExecutionError); ok {
				classificationStr := "unknown"
				switch execError.Classification {
				case ErrorClassificationTransient:
					classificationStr = "transient"
				case ErrorClassificationPermanent:
					classificationStr = "permanent"
				}
				detailedErr := fmt.Sprintf("%s (classification: %s, retry_count: %d)", errStr, classificationStr, retryCount)
				errorMsg = &detailedErr
			}
		} else {
			status = "completed"
		}

		outputDataJSON, _ := json.Marshal(output)
		if err := e.repo.UpdateStepExecution(ctx, stepExecution.ID, status, outputDataJSON, errorMsg); err != nil {
			e.logger.Error("failed to update step execution record", "error", err, "step_id", stepExecution.ID)
		}
	}

	return output, execErr
}

// parseRetryConfig parses retry configuration from node data
func (e *Executor) parseRetryConfig(data interface{}) NodeRetryConfig {
	config := e.defaultRetryConfig

	if retryMap, ok := data.(map[string]interface{}); ok {
		if enabled, ok := retryMap["enabled"].(bool); ok {
			config.Enabled = enabled
		}
		if maxRetries, ok := retryMap["max_retries"].(float64); ok {
			config.MaxRetries = int(maxRetries)
		}
		if initialBackoff, ok := retryMap["initial_backoff_ms"].(float64); ok {
			config.InitialBackoff = time.Duration(initialBackoff) * time.Millisecond
		}
		if maxBackoff, ok := retryMap["max_backoff_ms"].(float64); ok {
			config.MaxBackoff = time.Duration(maxBackoff) * time.Millisecond
		}
		if multiplier, ok := retryMap["backoff_multiplier"].(float64); ok {
			config.BackoffMultiplier = multiplier
		}
	}

	return config
}

// executeNode executes a single node
func (e *Executor) executeNode(ctx context.Context, node workflow.Node, execCtx *ExecutionContext) (interface{}, error) {
	startTime := time.Now()

	// Inject credentials if injector is available
	nodeToExecute := node
	var credentialValues []string

	if e.credentialInjector != nil && len(node.Data.Config) > 0 {
		injCtx := &credential.InjectionContext{
			TenantID:    execCtx.TenantID,
			WorkflowID:  execCtx.WorkflowID,
			ExecutionID: execCtx.ExecutionID,
			AccessedBy:  execCtx.GetUserID(),
		}

		injectResult, err := e.credentialInjector.InjectCredentials(ctx, node.Data.Config, injCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to inject credentials: %w", err)
		}

		// Update node config with injected credentials
		nodeToExecute = node
		nodeToExecute.Data.Config = injectResult.Config

		// Store credential values for masking
		credentialValues = injectResult.Values
		execCtx.CredentialValues = append(execCtx.CredentialValues, credentialValues...)
	}

	var output interface{}
	var err error

	switch nodeToExecute.Type {
	case string(workflow.NodeTypeActionHTTP):
		output, err = e.executeHTTPAction(ctx, nodeToExecute, execCtx)
	case string(workflow.NodeTypeActionTransform):
		output, err = e.executeTransformAction(ctx, nodeToExecute, execCtx)
	case string(workflow.NodeTypeActionFormula):
		output, err = e.executeFormulaAction(ctx, nodeToExecute, execCtx)
	case string(workflow.NodeTypeActionCode):
		output, err = e.executeCodeAction(ctx, nodeToExecute, execCtx)
	case string(workflow.NodeTypeActionSlackSendMessage):
		output, err = e.executeSlackSendMessageAction(ctx, nodeToExecute, execCtx)
	case string(workflow.NodeTypeActionSlackSendDM):
		output, err = e.executeSlackSendDMAction(ctx, nodeToExecute, execCtx)
	case string(workflow.NodeTypeActionSlackUpdateMessage):
		output, err = e.executeSlackUpdateMessageAction(ctx, nodeToExecute, execCtx)
	case string(workflow.NodeTypeActionSlackAddReaction):
		output, err = e.executeSlackAddReactionAction(ctx, nodeToExecute, execCtx)
	case string(workflow.NodeTypeControlDelay):
		output, err = e.executeDelayAction(ctx, nodeToExecute, execCtx)
	case string(workflow.NodeTypeControlSubWorkflow):
		output, err = e.executeSubWorkflowAction(ctx, nodeToExecute, execCtx)
	case string(workflow.NodeTypeControlLoop):
		// Loop nodes need access to workflow definition
		// For now, return error - will be handled separately in execution flow
		err = fmt.Errorf("loop nodes must be handled in execution flow, not as individual nodes")
	case string(workflow.NodeTypeControlParallel):
		// Parallel nodes need access to workflow definition
		// For now, return error - will be handled separately in execution flow
		err = fmt.Errorf("parallel nodes must be handled in execution flow, not as individual nodes")
	default:
		err = fmt.Errorf("unknown node type: %s", nodeToExecute.Type)
	}

	// Mask credentials in output if any were injected
	if len(credentialValues) > 0 && e.credentialInjector != nil {
		output = e.credentialInjector.MaskOutput(output, credentialValues)
	}

	duration := time.Since(startTime)
	e.logger.Info("node executed",
		"node_id", node.ID,
		"node_type", node.Type,
		"duration_ms", duration.Milliseconds(),
		"success", err == nil,
	)

	return output, err
}

// failExecution marks an execution as failed
func (e *Executor) failExecution(ctx context.Context, execution *workflow.Execution, err ...error) error {
	var errMsg string
	if len(err) > 0 && err[0] != nil {
		errMsg = err[0].Error()
	} else {
		errMsg = "execution failed"
	}

	e.repo.UpdateExecutionStatus(ctx, execution.ID, workflow.ExecutionStatusFailed, nil, &errMsg)

	// Broadcast execution failure
	if e.broadcaster != nil {
		e.broadcaster.BroadcastExecutionFailed(execution.TenantID, execution.WorkflowID, execution.ID, errMsg)
	}

	if len(err) > 0 && err[0] != nil {
		return err[0]
	}
	return fmt.Errorf("%s", errMsg)
}

// Helper functions

func buildNodeMap(nodes []workflow.Node) map[string]workflow.Node {
	nodeMap := make(map[string]workflow.Node)
	for _, node := range nodes {
		nodeMap[node.ID] = node
	}
	return nodeMap
}

func isTriggerNode(nodeType string) bool {
	return nodeType == string(workflow.NodeTypeTriggerWebhook) ||
		nodeType == string(workflow.NodeTypeTriggerSchedule)
}

// topologicalSort performs a topological sort on the workflow DAG
func topologicalSort(nodes []workflow.Node, edges []workflow.Edge) ([]string, error) {
	// Build adjacency list and in-degree map
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

	// Find nodes with no incoming edges (start nodes)
	var queue []string
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	// Process nodes in topological order
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

	// Check for cycles
	if len(result) != len(nodes) {
		return nil, fmt.Errorf("workflow contains cycles")
	}

	return result, nil
}

// buildInputData creates the input data context for a node execution
func buildInputData(execCtx *ExecutionContext) map[string]interface{} {
	return map[string]interface{}{
		"trigger": execCtx.TriggerData,
		"steps":   execCtx.StepOutputs,
		"env": map[string]interface{}{
			"tenant_id":    execCtx.TenantID,
			"execution_id": execCtx.ExecutionID,
			"workflow_id":  execCtx.WorkflowID,
		},
	}
}
