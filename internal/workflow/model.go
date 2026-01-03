package workflow

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Workflow represents a workflow definition
type Workflow struct {
	ID          string          `db:"id" json:"id"`
	TenantID    string          `db:"tenant_id" json:"tenant_id"`
	Name        string          `db:"name" json:"name"`
	Description string          `db:"description" json:"description"`
	Definition  json.RawMessage `db:"definition" json:"definition"`
	Status      string          `db:"status" json:"status"`
	Version     int             `db:"version" json:"version"`
	CreatedBy   string          `db:"created_by" json:"created_by"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`
}

// WorkflowDefinition represents the full workflow structure
type WorkflowDefinition struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Node represents a node in the workflow
type NodeData struct {
	Name   string          `json:"name"`
	Config json.RawMessage `json:"config"`
}

type Node struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Position Position `json:"position"`
	Data     NodeData `json:"data"`
	// Config is extracted from Data.Config for backward compatibility
	Config json.RawMessage `json:"-"`
}

// Position represents node position on the canvas
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Edge represents a connection between nodes
type Edge struct {
	ID       string `json:"id"`
	Source   string `json:"source"`
	Target   string `json:"target"`
	SourceID string `json:"sourceHandle,omitempty"`
	TargetID string `json:"targetHandle,omitempty"`
	Label    string `json:"label,omitempty"` // Used for conditional branches: "true" or "false"
}

// NodeType represents the type of a node
type NodeType string

const (
	NodeTypeTriggerWebhook           NodeType = "trigger:webhook"
	NodeTypeTriggerSchedule          NodeType = "trigger:schedule"
	NodeTypeActionHTTP               NodeType = "action:http"
	NodeTypeActionTransform          NodeType = "action:transform"
	NodeTypeActionFormula            NodeType = "action:formula"
	NodeTypeActionCode               NodeType = "action:code"
	NodeTypeActionEmail              NodeType = "action:email"
	NodeTypeActionSlackSendMessage   NodeType = "slack:send_message"
	NodeTypeActionSlackSendDM        NodeType = "slack:send_dm"
	NodeTypeActionSlackUpdateMessage NodeType = "slack:update_message"
	NodeTypeActionSlackAddReaction   NodeType = "slack:add_reaction"
	NodeTypeActionSubworkflow        NodeType = "action:subworkflow"
	NodeTypeControlIf                NodeType = "control:if"
	NodeTypeControlLoop              NodeType = "control:loop"
	NodeTypeControlParallel          NodeType = "control:parallel"
	NodeTypeControlFork              NodeType = "control:fork"
	NodeTypeControlJoin              NodeType = "control:join"
	NodeTypeControlDelay             NodeType = "control:delay"
	NodeTypeControlSubWorkflow       NodeType = "control:sub_workflow"
	NodeTypeControlTry               NodeType = "control:try"
	NodeTypeControlCatch             NodeType = "control:catch"
	NodeTypeControlFinally           NodeType = "control:finally"
	NodeTypeControlRetry             NodeType = "control:retry"
	NodeTypeControlCircuitBreaker    NodeType = "control:circuit_breaker"
)

// HTTPActionConfig represents HTTP action configuration
type HTTPActionConfig struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    json.RawMessage   `json:"body,omitempty"`
	Timeout int               `json:"timeout,omitempty"`
}

// TransformActionConfig represents transform action configuration
type TransformActionConfig struct {
	Expression string            `json:"expression"`
	Mapping    map[string]string `json:"mapping,omitempty"`
}

// FormulaActionConfig represents formula action configuration
type FormulaActionConfig struct {
	Expression     string `json:"expression"`
	OutputVariable string `json:"output_variable,omitempty"`
}

// ScriptActionConfig represents script (JavaScript) action configuration
type ScriptActionConfig struct {
	Script      string `json:"script"`                 // JavaScript code to execute
	Timeout     int    `json:"timeout,omitempty"`      // Max execution time in seconds (default: 30)
	MemoryLimit int    `json:"memory_limit,omitempty"` // Max memory in MB (future enhancement)
}

// WebhookTriggerConfig represents webhook trigger configuration
type WebhookTriggerConfig struct {
	Path        string `json:"path,omitempty"`
	AuthType    string `json:"auth_type,omitempty"` // none, basic, signature, api_key
	Secret      string `json:"secret,omitempty"`
	AllowedIPs  string `json:"allowed_ips,omitempty"`
	ResponseURL string `json:"response_url,omitempty"`
}

// ScheduleTriggerConfig represents schedule trigger configuration
type ScheduleTriggerConfig struct {
	Cron     string `json:"cron"`
	Timezone string `json:"timezone,omitempty"`
}

// ConditionalActionConfig represents conditional (if/else) action configuration
type ConditionalActionConfig struct {
	Condition   string `json:"condition"`               // Boolean expression to evaluate
	TrueBranch  string `json:"true_branch,omitempty"`   // Edge ID or label for true branch
	FalseBranch string `json:"false_branch,omitempty"`  // Edge ID or label for false branch
	Description string `json:"description,omitempty"`   // Optional description of the condition
	StopOnTrue  bool   `json:"stop_on_true,omitempty"`  // Stop workflow if condition is true
	StopOnFalse bool   `json:"stop_on_false,omitempty"` // Stop workflow if condition is false
}

// LoopActionConfig represents loop (for-each) action configuration
type LoopActionConfig struct {
	Source        string `json:"source"`                   // JSONPath to array (e.g., ${steps.node1.output.items})
	ItemVariable  string `json:"item_variable"`            // Variable name for current item (e.g., "item")
	IndexVariable string `json:"index_variable,omitempty"` // Variable name for current index (e.g., "index")
	MaxIterations int    `json:"max_iterations,omitempty"` // Safety limit (default 1000)
	OnError       string `json:"on_error,omitempty"`       // "continue" or "stop" (default "stop")
}

// ParallelConfig represents parallel execution configuration
type ParallelConfig struct {
	// Branches to execute in parallel (named branches for better organization)
	Branches []ParallelBranch `json:"branches,omitempty"`
	// Wait for all branches or first completion
	WaitMode string `json:"wait_mode,omitempty"` // "all" or "first" (default "all")
	// Maximum concurrent executions (0 = unlimited, >0 = max concurrent branches)
	MaxConcurrency int `json:"max_concurrency,omitempty"`
	// Timeout for parallel execution (e.g., "30s", "1m", "2h")
	Timeout string `json:"timeout,omitempty"`
	// What to do if one branch fails
	FailureMode string `json:"failure_mode,omitempty"` // "stop_all" or "continue" (default "stop_all")
	// Legacy: ErrorStrategy for backward compatibility (maps to FailureMode)
	ErrorStrategy string `json:"error_strategy,omitempty"` // "fail_fast" or "wait_all" (deprecated, use FailureMode)
}

// ParallelBranch represents a named branch in parallel execution
type ParallelBranch struct {
	Name  string   `json:"name"`  // Branch name for identification
	Nodes []string `json:"nodes"` // Node IDs in this branch (sequential execution within branch)
}

// DelayConfig represents delay action configuration
type DelayConfig struct {
	Duration string `json:"duration"` // Duration string (e.g., "5s", "1m", "2h") or template variable (e.g., "{{steps.node1.delay}}")
}

// ForkConfig represents fork node configuration
type ForkConfig struct {
	BranchCount int `json:"branch_count"` // Number of parallel branches to create
}

// JoinConfig represents join node configuration
type JoinConfig struct {
	JoinStrategy  string `json:"join_strategy"`            // "wait_all" or "wait_n"
	RequiredCount int    `json:"required_count,omitempty"` // Number of branches required for wait_n strategy
	TimeoutMs     int    `json:"timeout_ms,omitempty"`     // Optional timeout for waiting (0 = no timeout)
	OnTimeout     string `json:"on_timeout,omitempty"`     // "fail" or "continue" (default "fail")
}

// SubWorkflowConfig represents sub-workflow action configuration
type SubWorkflowConfig struct {
	WorkflowID     string            `json:"workflow_id"`              // ID of the workflow to execute
	WorkflowName   string            `json:"workflow_name,omitempty"`  // Optional, for display purposes
	InputMapping   map[string]string `json:"input_mapping,omitempty"`  // Map parent context to sub-workflow input
	OutputMapping  map[string]string `json:"output_mapping,omitempty"` // Map sub-workflow output to parent context
	Mode           string            `json:"mode"`                     // "sync" or "async" execution mode
	Timeout        string            `json:"timeout,omitempty"`        // Timeout duration (e.g., "5m", "30s")
	TimeoutMs      int               `json:"timeout_ms,omitempty"`     // Timeout in milliseconds (deprecated, use Timeout)
	WaitForResult  bool              `json:"wait_for_result"`          // Sync (true) vs async (false) execution (deprecated, use Mode)
	InheritContext bool              `json:"inherit_context"`          // Whether to inherit parent context
}

// TryConfig represents try/catch/finally error handling configuration
type TryConfig struct {
	TryNodes     []string          `json:"try_nodes"`               // Node IDs to execute in the try block
	CatchNodes   []string          `json:"catch_nodes,omitempty"`   // Node IDs to execute on error (catch block)
	FinallyNodes []string          `json:"finally_nodes,omitempty"` // Node IDs to always execute (finally block)
	ErrorBinding string            `json:"error_binding,omitempty"` // Variable name to bind error details (e.g., "error")
	RetryConfig  *RetryNodeConfig  `json:"retry_config,omitempty"`  // Optional retry configuration
	Metadata     map[string]string `json:"metadata,omitempty"`      // Additional metadata
}

// RetryNodeConfig represents retry configuration for error handling
type RetryNodeConfig struct {
	Strategy             string   `json:"strategy"`                         // "fixed", "exponential", "exponential_jitter"
	MaxAttempts          int      `json:"max_attempts"`                     // Maximum number of retry attempts
	InitialDelayMs       int      `json:"initial_delay_ms"`                 // Initial delay in milliseconds
	MaxDelayMs           int      `json:"max_delay_ms,omitempty"`           // Maximum delay in milliseconds (for exponential)
	Multiplier           float64  `json:"multiplier,omitempty"`             // Backoff multiplier (for exponential, default 2.0)
	Jitter               bool     `json:"jitter,omitempty"`                 // Add random jitter to delays
	RetryableErrors      []string `json:"retryable_errors,omitempty"`       // Error types/patterns to retry (empty = all transient)
	NonRetryableErrors   []string `json:"non_retryable_errors,omitempty"`   // Error types/patterns to never retry
	RetryableStatusCodes []int    `json:"retryable_status_codes,omitempty"` // HTTP status codes to retry
}

// CatchConfig represents catch block configuration
type CatchConfig struct {
	ErrorTypes    []string          `json:"error_types,omitempty"`    // Specific error types to catch (empty = all)
	ErrorPatterns []string          `json:"error_patterns,omitempty"` // Regex patterns for error messages
	ErrorBinding  string            `json:"error_binding,omitempty"`  // Variable name to bind error details
	PropagateAs   string            `json:"propagate_as,omitempty"`   // Rethrow as different error type
	Metadata      map[string]string `json:"metadata,omitempty"`       // Additional metadata
}

// FinallyConfig represents finally block configuration
type FinallyConfig struct {
	AlwaysRun bool              `json:"always_run"` // Run even if try/catch nodes fail (default true)
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// CircuitBreakerConfig represents circuit breaker configuration for nodes
type CircuitBreakerConfig struct {
	Enabled           bool    `json:"enabled"`                       // Enable circuit breaker
	MaxFailures       int     `json:"max_failures"`                  // Consecutive failures to open circuit
	TimeoutMs         int     `json:"timeout_ms"`                    // Time to wait before half-open (milliseconds)
	MaxRequests       int     `json:"max_requests,omitempty"`        // Max requests in half-open state
	FailureThreshold  float64 `json:"failure_threshold,omitempty"`   // Failure ratio to open (0.0-1.0)
	SlidingWindowSize int     `json:"sliding_window_size,omitempty"` // Window size for failure tracking
	Name              string  `json:"name,omitempty"`                // Circuit breaker name (defaults to node ID)
}

// ErrorHandlingMetadata represents error metadata captured during execution
type ErrorHandlingMetadata struct {
	ErrorType      string                 `json:"error_type"`
	ErrorMessage   string                 `json:"error_message"`
	ErrorStack     string                 `json:"error_stack,omitempty"`
	Classification string                 `json:"classification"` // "transient", "permanent", "unknown"
	NodeID         string                 `json:"node_id"`
	NodeType       string                 `json:"node_type"`
	RetryAttempt   int                    `json:"retry_attempt"`
	MaxRetries     int                    `json:"max_retries"`
	Timestamp      string                 `json:"timestamp"`
	Context        map[string]interface{} `json:"context,omitempty"`
	CaughtBy       string                 `json:"caught_by,omitempty"`        // Node ID that caught the error
	RecoveryAction string                 `json:"recovery_action,omitempty"`  // "retry", "fallback", "propagate", "handled"
	HTTPStatusCode int                    `json:"http_status_code,omitempty"` // For HTTP errors
	OriginalError  *ErrorHandlingMetadata `json:"original_error,omitempty"`   // For wrapped/rethrown errors
}

// CreateWorkflowInput represents input for creating a workflow
type CreateWorkflowInput struct {
	Name        string          `json:"name" validate:"required,min=1,max=255"`
	Description string          `json:"description"`
	Definition  json.RawMessage `json:"definition" validate:"required"`
}

// UpdateWorkflowInput represents input for updating a workflow
type UpdateWorkflowInput struct {
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	Definition  json.RawMessage `json:"definition,omitempty"`
	Status      string          `json:"status,omitempty"`
}

// WorkflowStatus represents workflow status
type WorkflowStatus string

const (
	WorkflowStatusDraft    WorkflowStatus = "draft"
	WorkflowStatusActive   WorkflowStatus = "active"
	WorkflowStatusInactive WorkflowStatus = "inactive"
	WorkflowStatusArchived WorkflowStatus = "archived"
)

// Execution represents a workflow execution
type Execution struct {
	ID                string           `db:"id" json:"id"`
	TenantID          string           `db:"tenant_id" json:"tenant_id"`
	WorkflowID        string           `db:"workflow_id" json:"workflow_id"`
	WorkflowVersion   int              `db:"workflow_version" json:"workflow_version"`
	Status            string           `db:"status" json:"status"`
	TriggerType       string           `db:"trigger_type" json:"trigger_type"`
	TriggerData       *json.RawMessage `db:"trigger_data" json:"trigger_data,omitempty"`
	OutputData        *json.RawMessage `db:"output_data" json:"output_data,omitempty"`
	ErrorMessage      *string          `db:"error_message" json:"error_message,omitempty"`
	ParentExecutionID *string          `db:"parent_execution_id" json:"parent_execution_id,omitempty"`
	ExecutionDepth    int              `db:"execution_depth" json:"execution_depth"`
	StartedAt         *time.Time       `db:"started_at" json:"started_at,omitempty"`
	CompletedAt       *time.Time       `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt         time.Time        `db:"created_at" json:"created_at"`
}

// StepExecution represents a single step in an execution
type StepExecution struct {
	ID           string           `db:"id" json:"id"`
	ExecutionID  string           `db:"execution_id" json:"execution_id"`
	NodeID       string           `db:"node_id" json:"node_id"`
	NodeType     string           `db:"node_type" json:"node_type"`
	Status       string           `db:"status" json:"status"`
	InputData    *json.RawMessage `db:"input_data" json:"input_data,omitempty"`
	OutputData   *json.RawMessage `db:"output_data" json:"output_data,omitempty"`
	ErrorMessage *string          `db:"error_message" json:"error_message,omitempty"`
	RetryCount   int              `db:"retry_count" json:"retry_count"`
	StartedAt    *time.Time       `db:"started_at" json:"started_at,omitempty"`
	CompletedAt  *time.Time       `db:"completed_at" json:"completed_at,omitempty"`
	DurationMs   *int             `db:"duration_ms" json:"duration_ms,omitempty"`
}

// ExecutionStatus represents execution status
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
)

// ExecutionFilter represents filters for listing executions
type ExecutionFilter struct {
	WorkflowID        string     `json:"workflow_id,omitempty"`
	Status            string     `json:"status,omitempty"`
	TriggerType       string     `json:"trigger_type,omitempty"`
	StartDate         *time.Time `json:"start_date,omitempty"`
	EndDate           *time.Time `json:"end_date,omitempty"`
	ErrorSearch       string     `json:"error_search,omitempty"`
	ExecutionIDPrefix string     `json:"execution_id_prefix,omitempty"`
	MinDurationMs     *int64     `json:"min_duration_ms,omitempty"`
	MaxDurationMs     *int64     `json:"max_duration_ms,omitempty"`
}

// Validate validates the execution filter
func (f ExecutionFilter) Validate() error {
	if f.StartDate != nil && f.EndDate != nil {
		if f.EndDate.Before(*f.StartDate) {
			return errors.New("end_date must be after start_date")
		}
	}

	if f.MinDurationMs != nil && *f.MinDurationMs < 0 {
		return errors.New("min_duration_ms must be non-negative")
	}

	if f.MaxDurationMs != nil && *f.MaxDurationMs < 0 {
		return errors.New("max_duration_ms must be non-negative")
	}

	if f.MinDurationMs != nil && f.MaxDurationMs != nil {
		if *f.MaxDurationMs < *f.MinDurationMs {
			return errors.New("max_duration_ms must be greater than or equal to min_duration_ms")
		}
	}

	return nil
}

// PaginationCursor represents a cursor for pagination
type PaginationCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
}

// Encode encodes the cursor to a base64 string
func (c PaginationCursor) Encode() string {
	data, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(data)
}

// DecodePaginationCursor decodes a base64 cursor string
func DecodePaginationCursor(encoded string) (PaginationCursor, error) {
	if encoded == "" {
		return PaginationCursor{}, errors.New("empty cursor")
	}

	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return PaginationCursor{}, fmt.Errorf("invalid cursor encoding: %w", err)
	}

	var cursor PaginationCursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return PaginationCursor{}, fmt.Errorf("invalid cursor format: %w", err)
	}

	return cursor, nil
}

// ExecutionListResult represents a paginated list of executions
type ExecutionListResult struct {
	Data       []*Execution `json:"data"`
	Cursor     string       `json:"cursor,omitempty"`
	HasMore    bool         `json:"has_more"`
	TotalCount int          `json:"total_count"`
}

// ExecutionWithSteps represents an execution with its step executions
type ExecutionWithSteps struct {
	Execution *Execution       `json:"execution"`
	Steps     []*StepExecution `json:"steps"`
}

// ExecutionStats represents statistics about executions
type ExecutionStats struct {
	TotalCount   int            `json:"total_count"`
	StatusCounts map[string]int `json:"status_counts"`
}

// DryRunResult represents the result of a workflow dry-run validation
type DryRunResult struct {
	Valid           bool              `json:"valid"`
	ExecutionOrder  []string          `json:"execution_order"`  // Node IDs in execution order
	VariableMapping map[string]string `json:"variable_mapping"` // Variable -> source mapping
	Warnings        []DryRunWarning   `json:"warnings"`
	Errors          []DryRunError     `json:"errors"`
}

// DryRunWarning represents a warning found during dry-run
type DryRunWarning struct {
	NodeID  string `json:"node_id"`
	Message string `json:"message"`
}

// DryRunError represents an error found during dry-run
type DryRunError struct {
	NodeID  string `json:"node_id"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

// DryRunInput represents input data for dry-run testing
type DryRunInput struct {
	TestData map[string]interface{} `json:"test_data"`
}

// WorkflowVersion represents a version of a workflow definition
type WorkflowVersion struct {
	ID         string          `db:"id" json:"id"`
	WorkflowID string          `db:"workflow_id" json:"workflow_id"`
	Version    int             `db:"version" json:"version"`
	Definition json.RawMessage `db:"definition" json:"definition"`
	CreatedBy  string          `db:"created_by" json:"created_by"`
	CreatedAt  time.Time       `db:"created_at" json:"created_at"`
}
