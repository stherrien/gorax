package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

// PublisherAdapter adapts the queue Publisher to the workflow service interface
type PublisherAdapter struct {
	publisher *Publisher
	logger    *slog.Logger
}

// NewPublisherAdapter creates a new publisher adapter
func NewPublisherAdapter(publisher *Publisher, logger *slog.Logger) *PublisherAdapter {
	return &PublisherAdapter{
		publisher: publisher,
		logger:    logger,
	}
}

// PublishExecution publishes an execution message to the queue
// This method accepts a generic interface{} to match the workflow service interface
func (a *PublisherAdapter) PublishExecution(ctx context.Context, msg interface{}) error {
	// Convert the interface to a map
	msgMap, ok := msg.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid message type: expected map[string]interface{}")
	}

	// Extract required fields
	executionID, ok := msgMap["execution_id"].(string)
	if !ok {
		return fmt.Errorf("execution_id is required")
	}

	tenantID, ok := msgMap["tenant_id"].(string)
	if !ok {
		return fmt.Errorf("tenant_id is required")
	}

	workflowID, ok := msgMap["workflow_id"].(string)
	if !ok {
		return fmt.Errorf("workflow_id is required")
	}

	workflowVersion, ok := msgMap["workflow_version"].(int)
	if !ok {
		return fmt.Errorf("workflow_version is required")
	}

	triggerType, ok := msgMap["trigger_type"].(string)
	if !ok {
		return fmt.Errorf("trigger_type is required")
	}

	// Extract optional trigger data
	var triggerData json.RawMessage
	if td, exists := msgMap["trigger_data"]; exists {
		switch v := td.(type) {
		case []byte:
			triggerData = v
		case string:
			triggerData = []byte(v)
		case json.RawMessage:
			triggerData = v
		default:
			// Try to marshal it
			data, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("failed to marshal trigger_data: %w", err)
			}
			triggerData = data
		}
	}

	// Create execution message
	execMsg := NewExecutionMessage(
		executionID,
		tenantID,
		workflowID,
		workflowVersion,
		triggerType,
		triggerData,
	)

	// Publish to queue
	return a.publisher.PublishExecution(ctx, execMsg)
}
