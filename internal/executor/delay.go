package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gorax/gorax/internal/executor/actions"
	"github.com/gorax/gorax/internal/workflow"
)

// executeDelayAction executes a delay node
func (e *Executor) executeDelayAction(ctx context.Context, node workflow.Node, execCtx *ExecutionContext) (interface{}, error) {
	// Parse delay configuration
	var config workflow.DelayConfig
	if err := json.Unmarshal(node.Data.Config, &config); err != nil {
		return nil, fmt.Errorf("failed to parse delay configuration: %w", err)
	}

	// Validate duration is provided
	if config.Duration == "" {
		return nil, fmt.Errorf("duration is required")
	}

	// Interpolate duration string if it contains variables
	durationStr := config.Duration
	if e.containsVariable(durationStr) {
		interpolatedDuration := e.interpolateDuration(durationStr, execCtx)
		if interpolatedDuration == "" {
			return nil, fmt.Errorf("failed to interpolate duration: resolved to empty string")
		}
		// Check if interpolation actually happened (variable wasn't resolved)
		if e.containsVariable(interpolatedDuration) {
			return nil, fmt.Errorf("failed to interpolate duration: variable not found in context")
		}
		durationStr = interpolatedDuration
	}

	// Parse duration
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return nil, fmt.Errorf("invalid duration format: %w", err)
	}

	// Validate duration is not negative
	if duration < 0 {
		return nil, fmt.Errorf("duration must be positive, got: %s", duration)
	}

	// Log delay start
	e.logger.Info("delay started",
		"node_id", node.ID,
		"duration", durationStr,
		"duration_ms", duration.Milliseconds(),
	)

	// Broadcast delay started event if broadcaster is available
	if e.broadcaster != nil {
		e.broadcastDelayStarted(execCtx, node.ID, duration)
	}

	// Wait for the specified duration or context cancellation
	startTime := time.Now()
	err = e.waitWithContext(ctx, duration)
	actualDelay := time.Since(startTime)

	if err != nil {
		e.logger.Warn("delay interrupted",
			"node_id", node.ID,
			"requested_duration_ms", duration.Milliseconds(),
			"actual_duration_ms", actualDelay.Milliseconds(),
			"error", err,
		)
		return nil, fmt.Errorf("delay interrupted: %w", err)
	}

	// Log delay completion
	e.logger.Info("delay completed",
		"node_id", node.ID,
		"duration_ms", actualDelay.Milliseconds(),
	)

	// Broadcast delay completed event if broadcaster is available
	if e.broadcaster != nil {
		e.broadcastDelayCompleted(execCtx, node.ID, actualDelay)
	}

	// Return delay metadata
	return map[string]interface{}{
		"duration":   config.Duration,
		"delayed_ms": actualDelay.Milliseconds(),
		"completed":  true,
	}, nil
}

// containsVariable checks if a string contains variable syntax
func (e *Executor) containsVariable(s string) bool {
	return strings.Contains(s, "{{") || strings.Contains(s, "${")
}

// interpolateDuration interpolates variables in duration string
// Supports both {{path}} and ${path} syntax
func (e *Executor) interpolateDuration(durationTemplate string, execCtx *ExecutionContext) string {
	context := buildInterpolationContext(execCtx)

	// Try {{}} syntax first (standard interpolation)
	result := actions.InterpolateString(durationTemplate, context)
	if result != durationTemplate {
		return result
	}

	// Handle ${} syntax (used in some other parts of the system)
	if strings.HasPrefix(durationTemplate, "${") && strings.HasSuffix(durationTemplate, "}") {
		path := durationTemplate[2 : len(durationTemplate)-1]
		value, err := actions.GetValueByPath(context, path)
		if err != nil {
			// Return original if not found
			return durationTemplate
		}
		// Convert to string
		if str, ok := value.(string); ok {
			return str
		}
		// Try to convert other types to string
		return fmt.Sprintf("%v", value)
	}

	return result
}

// waitWithContext waits for duration or until context is cancelled
func (e *Executor) waitWithContext(ctx context.Context, duration time.Duration) error {
	if duration == 0 {
		return nil
	}

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// broadcastDelayStarted broadcasts delay started event
func (e *Executor) broadcastDelayStarted(execCtx *ExecutionContext, nodeID string, duration time.Duration) {
	// Note: Current broadcaster interface doesn't have a specific delay event method
	// This is a placeholder for future enhancement
	// For now, we could extend the broadcaster interface to support:
	// BroadcastDelayStarted(tenantID, workflowID, executionID, nodeID string, durationMs int)
}

// broadcastDelayCompleted broadcasts delay completed event
func (e *Executor) broadcastDelayCompleted(execCtx *ExecutionContext, nodeID string, actualDelay time.Duration) {
	// Note: Current broadcaster interface doesn't have a specific delay event method
	// This is a placeholder for future enhancement
	// For now, we could extend the broadcaster interface to support:
	// BroadcastDelayCompleted(tenantID, workflowID, executionID, nodeID string, actualDelayMs int)
}
