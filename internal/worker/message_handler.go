package worker

import (
	"context"
	"errors"

	"github.com/gorax/gorax/internal/queue"
)

// MessageContext holds context for message processing including receipt handle
type MessageContext struct {
	ReceiptHandle string
	RetryCount    int
}

// processExecutionMessageWithRequeue wraps message processing with requeue capability
func (w *Worker) processExecutionMessageWithRequeue(ctx context.Context, msg *queue.ExecutionMessage, receiptHandle string) error {
	w.logger.Info("processing execution message",
		"execution_id", msg.ExecutionID,
		"workflow_id", msg.WorkflowID,
		"tenant_id", msg.TenantID,
		"retry_count", msg.RetryCount,
	)

	// Load execution from database
	execution, err := w.workflowRepo.GetExecutionByID(ctx, msg.TenantID, msg.ExecutionID)
	if err != nil {
		w.logger.Error("failed to load execution",
			"error", err,
			"execution_id", msg.ExecutionID,
		)
		return err
	}

	// Process the execution
	if err := w.processExecution(ctx, execution); err != nil {
		// Check if tenant is at capacity - if so, requeue with delay
		if errors.Is(err, ErrTenantAtCapacity) {
			w.logger.Info("tenant at capacity, requeueing message",
				"tenant_id", msg.TenantID,
				"execution_id", msg.ExecutionID,
				"retry_count", msg.RetryCount,
			)

			// Only requeue if we have SQS client (queue mode)
			if w.sqsClient != nil {
				if requeueErr := w.requeueMessageWithDelay(ctx, receiptHandle, msg.RetryCount); requeueErr != nil {
					w.logger.Error("failed to requeue message",
						"error", requeueErr,
						"execution_id", msg.ExecutionID,
					)
					// Return original error so consumer knows processing failed
					return err
				}

				// Return nil to prevent consumer from deleting the message
				// The message will become visible again after the visibility timeout
				return nil
			}

			// In non-queue mode, return error (execution stays pending)
			return err
		}

		w.logger.Error("execution processing failed",
			"error", err,
			"execution_id", msg.ExecutionID,
		)
		return err
	}

	return nil
}

// requeueMessageWithDelay extends message visibility timeout to implement delay
func (w *Worker) requeueMessageWithDelay(ctx context.Context, receiptHandle string, retryCount int) error {
	if w.sqsClient == nil {
		return errors.New("SQS client not initialized")
	}

	return requeueMessage(ctx, w.sqsClient, receiptHandle, retryCount)
}
