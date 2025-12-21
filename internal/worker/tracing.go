package worker

import (
	"context"

	"github.com/gorax/gorax/internal/queue"
	"github.com/gorax/gorax/internal/tracing"
)

// handleMessageWithTracing wraps message handling with distributed tracing
func (h *QueueMessageHandler) handleMessageWithTracing(ctx context.Context, msg *queue.ExecutionMessage, receiptHandle string) error {
	return tracing.TraceQueueMessage(
		ctx,
		"workflow-executions",
		msg.ExecutionID,
		func(ctx context.Context) error {
			// Add message attributes to span
			tracing.AddWorkflowAttributes(ctx, map[string]interface{}{
				"tenant_id":      msg.TenantID,
				"workflow_id":    msg.WorkflowID,
				"execution_id":   msg.ExecutionID,
				"retry_count":    msg.RetryCount,
				"receipt_handle": receiptHandle,
			})

			return h.HandleMessage(ctx, msg, receiptHandle)
		},
	)
}
