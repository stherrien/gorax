package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gorax/gorax/internal/executor/javascript"
)

// ScriptAction implements the Action interface for sandboxed JavaScript execution
//
// Security Considerations:
// - Uses the secure JavaScript sandbox engine with comprehensive security restrictions
// - Blocks dangerous APIs: require, process, eval, Function, file system, network
// - Enforces execution timeout to prevent infinite loops
// - Enforces memory limits to prevent memory exhaustion attacks
// - Enforces call stack limits to prevent stack overflow attacks
// - Scripts run in isolated VM instances (no shared state between executions)
// - VM pooling for efficient resource management
// - Comprehensive audit logging for security events
type ScriptAction struct {
	engine     *javascript.Engine
	engineOnce sync.Once
	engineErr  error
	logger     *slog.Logger
}

// ScriptActionConfig represents the configuration for a script action
type ScriptActionConfig struct {
	Script      string `json:"script"`                 // JavaScript code to execute
	Timeout     int    `json:"timeout,omitempty"`      // Max execution time in seconds (default: 30)
	MemoryLimit int64  `json:"memory_limit,omitempty"` // Max memory in MB (default: 128)
}

// ScriptActionResult represents the result of a script execution
type ScriptActionResult struct {
	Result      any                       `json:"result"`                 // The value returned by the script
	ConsoleLogs []javascript.ConsoleEntry `json:"console_logs,omitempty"` // Captured console output
	Duration    time.Duration             `json:"duration"`               // Execution duration
	MemoryDelta int64                     `json:"memory_delta,omitempty"` // Memory change during execution
}

const (
	defaultScriptTimeout  = 30 // seconds
	maxScriptTimeout      = 60 // seconds
	defaultEnginePoolSize = 10
)

// NewScriptAction creates a new ScriptAction with default configuration
func NewScriptAction() *ScriptAction {
	return &ScriptAction{
		logger: slog.Default(),
	}
}

// NewScriptActionWithLogger creates a new ScriptAction with a custom logger
func NewScriptActionWithLogger(logger *slog.Logger) *ScriptAction {
	if logger == nil {
		logger = slog.Default()
	}
	return &ScriptAction{
		logger: logger,
	}
}

// getEngine returns the shared JavaScript engine, creating it if necessary
func (a *ScriptAction) getEngine() (*javascript.Engine, error) {
	a.engineOnce.Do(func() {
		config := &javascript.EngineConfig{
			Limits:               javascript.DefaultLimits(),
			SandboxConfig:        javascript.DefaultSandboxConfig(),
			PoolSize:             defaultEnginePoolSize,
			Logger:               a.logger,
			EnableConsoleCapture: true,
		}
		a.engine, a.engineErr = javascript.NewEngine(config)
	})
	return a.engine, a.engineErr
}

// Execute implements the Action interface
func (a *ScriptAction) Execute(ctx context.Context, input *ActionInput) (*ActionOutput, error) {
	// Parse config
	configBytes, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var config ScriptActionConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to parse script action config: %w", err)
	}

	// Validate script
	if config.Script == "" {
		return nil, fmt.Errorf("script is required")
	}

	// Get the JavaScript engine
	engine, err := a.getEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JavaScript engine: %w", err)
	}

	// Determine timeout
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout <= 0 {
		timeout = defaultScriptTimeout * time.Second
	}
	if timeout > maxScriptTimeout*time.Second {
		timeout = maxScriptTimeout * time.Second
	}

	// Build execution context
	execCtx := javascript.FromWorkflowContext(input.Context)

	// Extract metadata for tracing
	tenantID, _ := extractString(input.Context, "env", "tenant_id")
	workflowID, _ := extractString(input.Context, "env", "workflow_id")
	nodeID, _ := extractString(input.Context, "env", "node_id")
	executionID, _ := extractString(input.Context, "env", "execution_id")

	// Configure execution
	executeConfig := &javascript.ExecuteConfig{
		Script:      config.Script,
		Context:     execCtx,
		Timeout:     timeout,
		ExecutionID: executionID,
		TenantID:    tenantID,
		WorkflowID:  workflowID,
		NodeID:      nodeID,
	}

	// Execute the script in the sandboxed environment
	result, err := engine.Execute(ctx, executeConfig)
	if err != nil {
		return nil, err
	}

	// Create output
	output := NewActionOutput(&ScriptActionResult{
		Result:      result.Result,
		ConsoleLogs: result.ConsoleLogs,
		Duration:    result.Duration,
		MemoryDelta: result.MemoryDelta,
	})
	output.WithMetadata("execution_time_ms", result.Duration.Milliseconds())
	output.WithMetadata("execution_id", result.ExecutionID)
	output.WithMetadata("memory_delta_bytes", result.MemoryDelta)
	output.WithMetadata("console_log_count", len(result.ConsoleLogs))

	return output, nil
}

// Close releases resources held by the ScriptAction
func (a *ScriptAction) Close() error {
	if a.engine != nil {
		return a.engine.Close()
	}
	return nil
}

// extractString extracts a string value from a nested map
func extractString(m map[string]any, keys ...string) (string, bool) {
	if m == nil || len(keys) == 0 {
		return "", false
	}

	current := any(m)
	for i, key := range keys {
		currentMap, ok := current.(map[string]any)
		if !ok {
			return "", false
		}

		val, exists := currentMap[key]
		if !exists {
			return "", false
		}

		if i == len(keys)-1 {
			if str, ok := val.(string); ok {
				return str, true
			}
			return "", false
		}

		current = val
	}

	return "", false
}
