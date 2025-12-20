package worker

import (
	"context"
	"fmt"
	"math"
)

// SQSClientInterface defines the interface for SQS operations needed for requeue
type SQSClientInterface interface {
	ChangeMessageVisibility(ctx context.Context, receiptHandle string, visibilityTimeout int32) error
}

// requeueMessage requeues a message by extending its visibility timeout
// This allows the message to become visible again after a delay for retry
func requeueMessage(ctx context.Context, sqsClient SQSClientInterface, receiptHandle string, retryCount int) error {
	delay := calculateRequeueDelay(retryCount)

	if err := sqsClient.ChangeMessageVisibility(ctx, receiptHandle, delay); err != nil {
		return fmt.Errorf("failed to requeue message: %w", err)
	}

	return nil
}

// calculateRequeueDelay calculates the delay before retry using exponential backoff
// Returns delay in seconds, capped at 5 minutes (300 seconds)
func calculateRequeueDelay(retryCount int) int32 {
	// Exponential backoff: 30s, 60s, 120s, 240s, 300s (capped)
	baseDelay := 30.0 // 30 seconds base
	maxDelay := 300   // 5 minutes maximum

	delay := baseDelay * math.Pow(2, float64(retryCount))

	if delay > float64(maxDelay) {
		return int32(maxDelay)
	}

	return int32(delay)
}
