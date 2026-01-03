package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/gorax/gorax/internal/workflow"
)

// TryAction implements try/catch/finally error handling
type TryAction struct {
	executeNodeFunc func(ctx context.Context, nodeID string) (interface{}, error)
	getNodeFunc     func(nodeID string) (*workflow.Node, error)
}

// NewTryAction creates a new try action
func NewTryAction(
	executeNode func(ctx context.Context, nodeID string) (interface{}, error),
	getNode func(nodeID string) (*workflow.Node, error),
) *TryAction {
	return &TryAction{
		executeNodeFunc: executeNode,
		getNodeFunc:     getNode,
	}
}

// Execute executes the try/catch/finally action
func (a *TryAction) Execute(ctx context.Context, input *ActionInput) (*ActionOutput, error) {
	// Parse configuration
	var config workflow.TryConfig
	configBytes, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to parse try config: %w", err)
	}

	// Validate configuration
	if len(config.TryNodes) == 0 {
		return nil, fmt.Errorf("try block must have at least one node")
	}

	var tryError error
	var tryOutput interface{}
	var errorMetadata *workflow.ErrorHandlingMetadata

	// Execute try block
	tryOutput, tryError = a.executeTryBlock(ctx, config.TryNodes)

	// If error occurred, execute catch block
	if tryError != nil {
		// Create error metadata
		errorMetadata = createErrorMetadata(tryError, config.TryNodes[0], "try_block")

		// Check if we have a catch block
		if len(config.CatchNodes) > 0 {
			// Bind error to context if error binding is specified
			catchContext := make(map[string]interface{})
			for k, v := range input.Context {
				catchContext[k] = v
			}

			// Add error to context
			if config.ErrorBinding != "" {
				catchContext[config.ErrorBinding] = errorMetadata
			} else {
				catchContext["error"] = errorMetadata
			}

			// Execute catch block
			catchOutput, catchErr := a.executeCatchBlock(ctx, config.CatchNodes, catchContext)
			if catchErr != nil {
				// Catch block failed, record it
				errorMetadata.RecoveryAction = "failed"
				// Execute finally block if present
				if len(config.FinallyNodes) > 0 {
					a.executeFinallyBlock(ctx, config.FinallyNodes, input.Context)
				}
				return nil, fmt.Errorf("catch block failed: %w (original error: %v)", catchErr, tryError)
			}

			// Catch block succeeded
			errorMetadata.RecoveryAction = "handled"
			errorMetadata.CaughtBy = config.CatchNodes[0]

			// Execute finally block if present
			if len(config.FinallyNodes) > 0 {
				finallyOutput, finallyErr := a.executeFinallyBlock(ctx, config.FinallyNodes, catchContext)
				if finallyErr != nil {
					return nil, fmt.Errorf("finally block failed: %w", finallyErr)
				}

				// Return combined output
				return NewActionOutput(map[string]interface{}{
					"try_error":      errorMetadata,
					"catch_output":   catchOutput,
					"finally_output": finallyOutput,
					"success":        true,
					"error_handled":  true,
				}), nil
			}

			// No finally block, return catch output
			return NewActionOutput(map[string]interface{}{
				"try_error":     errorMetadata,
				"catch_output":  catchOutput,
				"success":       true,
				"error_handled": true,
			}), nil
		}

		// No catch block, execute finally and propagate error
		if len(config.FinallyNodes) > 0 {
			a.executeFinallyBlock(ctx, config.FinallyNodes, input.Context)
		}
		errorMetadata.RecoveryAction = "propagate"
		return nil, tryError
	}

	// Try block succeeded, execute finally block if present
	if len(config.FinallyNodes) > 0 {
		finallyOutput, finallyErr := a.executeFinallyBlock(ctx, config.FinallyNodes, input.Context)
		if finallyErr != nil {
			return nil, fmt.Errorf("finally block failed: %w", finallyErr)
		}

		// Return combined output
		return NewActionOutput(map[string]interface{}{
			"try_output":     tryOutput,
			"finally_output": finallyOutput,
			"success":        true,
			"error_handled":  false,
		}), nil
	}

	// No finally block, return try output
	return NewActionOutput(map[string]interface{}{
		"try_output":    tryOutput,
		"success":       true,
		"error_handled": false,
	}), nil
}

// executeTryBlock executes the try block nodes
func (a *TryAction) executeTryBlock(ctx context.Context, nodeIDs []string) (interface{}, error) {
	outputs := make(map[string]interface{})

	for _, nodeID := range nodeIDs {
		output, err := a.executeNodeFunc(ctx, nodeID)
		if err != nil {
			return outputs, fmt.Errorf("try block node %s failed: %w", nodeID, err)
		}
		outputs[nodeID] = output
	}

	return outputs, nil
}

// executeCatchBlock executes the catch block nodes
func (a *TryAction) executeCatchBlock(ctx context.Context, nodeIDs []string, contextWithError map[string]interface{}) (interface{}, error) {
	outputs := make(map[string]interface{})

	for _, nodeID := range nodeIDs {
		output, err := a.executeNodeFunc(ctx, nodeID)
		if err != nil {
			return outputs, fmt.Errorf("catch block node %s failed: %w", nodeID, err)
		}
		outputs[nodeID] = output
	}

	return outputs, nil
}

// executeFinallyBlock executes the finally block nodes
func (a *TryAction) executeFinallyBlock(ctx context.Context, nodeIDs []string, context map[string]interface{}) (interface{}, error) {
	outputs := make(map[string]interface{})

	for _, nodeID := range nodeIDs {
		output, err := a.executeNodeFunc(ctx, nodeID)
		if err != nil {
			return outputs, fmt.Errorf("finally block node %s failed: %w", nodeID, err)
		}
		outputs[nodeID] = output
	}

	return outputs, nil
}

// CatchAction implements catch block error filtering
type CatchAction struct {
	executeNodeFunc func(ctx context.Context, nodeID string) (interface{}, error)
}

// NewCatchAction creates a new catch action
func NewCatchAction(executeNode func(ctx context.Context, nodeID string) (interface{}, error)) *CatchAction {
	return &CatchAction{
		executeNodeFunc: executeNode,
	}
}

// Execute executes the catch action with error filtering
func (a *CatchAction) Execute(ctx context.Context, input *ActionInput) (*ActionOutput, error) {
	// Parse configuration
	var config workflow.CatchConfig
	configBytes, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to parse catch config: %w", err)
	}

	// Get error from context
	var errorData *workflow.ErrorHandlingMetadata
	errorBinding := config.ErrorBinding
	if errorBinding == "" {
		errorBinding = "error"
	}

	if errData, ok := input.Context[errorBinding]; ok {
		if errMetadata, ok := errData.(*workflow.ErrorHandlingMetadata); ok {
			errorData = errMetadata
		} else if errMap, ok := errData.(map[string]interface{}); ok {
			// Try to convert from map
			bytes, _ := json.Marshal(errMap)
			json.Unmarshal(bytes, &errorData)
		}
	}

	if errorData == nil {
		return nil, fmt.Errorf("no error data found in context")
	}

	// Check if this catch block should handle the error
	if !a.shouldCatchError(errorData, config) {
		// This catch block doesn't handle this error, propagate it
		return nil, fmt.Errorf("error not caught by this catch block: %s", errorData.ErrorMessage)
	}

	// Error should be caught, return success with error details
	return NewActionOutput(map[string]interface{}{
		"caught_error":  errorData,
		"error_handled": true,
	}), nil
}

// shouldCatchError determines if the catch block should handle the given error
func (a *CatchAction) shouldCatchError(errorData *workflow.ErrorHandlingMetadata, config workflow.CatchConfig) bool {
	// If no filters specified, catch all errors
	if len(config.ErrorTypes) == 0 && len(config.ErrorPatterns) == 0 {
		return true
	}

	// Check error type filters
	if len(config.ErrorTypes) > 0 {
		matched := false
		for _, errorType := range config.ErrorTypes {
			if errorData.ErrorType == errorType || errorData.Classification == errorType {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check error pattern filters
	if len(config.ErrorPatterns) > 0 {
		matched := false
		for _, pattern := range config.ErrorPatterns {
			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}
			if re.MatchString(errorData.ErrorMessage) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// createErrorMetadata creates error metadata from an error
func createErrorMetadata(err error, nodeID, nodeType string) *workflow.ErrorHandlingMetadata {
	classification := "unknown"

	// Try to determine classification
	if execErr, ok := err.(interface{ IsRetryable() bool }); ok {
		if execErr.IsRetryable() {
			classification = "transient"
		} else {
			classification = "permanent"
		}
	}

	return &workflow.ErrorHandlingMetadata{
		ErrorType:      fmt.Sprintf("%T", err),
		ErrorMessage:   err.Error(),
		Classification: classification,
		NodeID:         nodeID,
		NodeType:       nodeType,
		RetryAttempt:   0,
		MaxRetries:     0,
		Timestamp:      time.Now().Format(time.RFC3339),
		Context:        make(map[string]interface{}),
	}
}
