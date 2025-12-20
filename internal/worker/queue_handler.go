package worker

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gorax/gorax/internal/queue"
)

// QueueMessageHandler wraps the standard consumer to add requeue capability
type QueueMessageHandler struct {
	worker    *Worker
	sqsClient *queue.SQSClient
	logger    *slog.Logger
}

// NewQueueMessageHandler creates a handler that supports message requeue
func NewQueueMessageHandler(worker *Worker, sqsClient *queue.SQSClient, logger *slog.Logger) *QueueMessageHandler {
	return &QueueMessageHandler{
		worker:    worker,
		sqsClient: sqsClient,
		logger:    logger,
	}
}

// HandleMessage processes a message with receipt handle for requeue support
// This is called by a custom consumer that exposes receipt handles
func (h *QueueMessageHandler) HandleMessage(ctx context.Context, msg *queue.ExecutionMessage, receiptHandle string) error {
	h.logger.Info("handling queue message",
		"execution_id", msg.ExecutionID,
		"tenant_id", msg.TenantID,
		"retry_count", msg.RetryCount,
	)

	// Load execution from database
	execution, err := h.worker.workflowRepo.GetExecutionByID(ctx, msg.TenantID, msg.ExecutionID)
	if err != nil {
		h.logger.Error("failed to load execution",
			"error", err,
			"execution_id", msg.ExecutionID,
		)
		return err
	}

	// Process the execution
	err = h.worker.processExecution(ctx, execution)
	if err != nil {
		// Check if tenant is at capacity
		if errors.Is(err, ErrTenantAtCapacity) {
			h.logger.Info("tenant at capacity, requeueing message with delay",
				"tenant_id", msg.TenantID,
				"execution_id", msg.ExecutionID,
				"retry_count", msg.RetryCount,
			)

			// Requeue by extending visibility timeout
			if requeueErr := h.requeueWithDelay(ctx, receiptHandle, msg.RetryCount); requeueErr != nil {
				h.logger.Error("failed to requeue message",
					"error", requeueErr,
					"execution_id", msg.ExecutionID,
				)
				// Return original capacity error
				return err
			}

			// Return nil to indicate message was requeued successfully
			// Consumer should NOT delete this message
			return ErrMessageRequeued
		}

		// Other errors - let consumer handle normally (retry or DLQ)
		return err
	}

	// Success - consumer will delete the message
	return nil
}

// requeueWithDelay extends message visibility timeout to delay retry
func (h *QueueMessageHandler) requeueWithDelay(ctx context.Context, receiptHandle string, retryCount int) error {
	delay := calculateRequeueDelay(retryCount)

	h.logger.Debug("extending message visibility",
		"receipt_handle", receiptHandle,
		"delay_seconds", delay,
	)

	return h.sqsClient.ChangeMessageVisibility(ctx, receiptHandle, delay)
}

// ErrMessageRequeued indicates message was requeued and should not be deleted
var ErrMessageRequeued = errors.New("message requeued with delay")
