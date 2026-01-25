package webhook

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// WebhookDeliverer is an interface for delivering webhooks
type WebhookDeliverer interface {
	DeliverWebhook(ctx context.Context, webhook *Webhook, event *WebhookEvent) (*RetryResult, error)
}

// RetryWorker processes failed webhook events and retries them
type RetryWorker struct {
	repo      *Repository
	deliverer WebhookDeliverer
	config    RetryConfig
	logger    *slog.Logger
}

// NewRetryWorker creates a new retry worker
func NewRetryWorker(repo *Repository, deliverer WebhookDeliverer, config RetryConfig) *RetryWorker {
	return &RetryWorker{
		repo:      repo,
		deliverer: deliverer,
		config:    config,
		logger:    slog.Default(),
	}
}

// WithLogger sets a custom logger for the retry worker
func (w *RetryWorker) WithLogger(logger *slog.Logger) *RetryWorker {
	w.logger = logger
	return w
}

// Start starts the retry worker in a loop
func (w *RetryWorker) Start(ctx context.Context, interval time.Duration) error {
	w.logger.Info("starting webhook retry worker", "interval", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("webhook retry worker stopping")
			return ctx.Err()
		case <-ticker.C:
			processed, err := w.ProcessRetries(ctx, 100)
			if err != nil {
				w.logger.Error("error processing retries", "error", err)
				continue
			}
			if processed > 0 {
				w.logger.Info("processed webhook retries", "count", processed)
			}
		}
	}
}

// ProcessRetries processes a batch of webhook events ready for retry
func (w *RetryWorker) ProcessRetries(ctx context.Context, batchSize int) (int, error) {
	// Get events ready for retry
	events, err := w.repo.GetEventsForRetry(ctx, batchSize)
	if err != nil {
		return 0, fmt.Errorf("get events for retry: %w", err)
	}

	if len(events) == 0 {
		return 0, nil
	}

	w.logger.Debug("processing webhook retries", "count", len(events))

	processed := 0
	for _, event := range events {
		if err := w.processEvent(ctx, event); err != nil {
			w.logger.Error("error processing retry",
				"event_id", event.ID,
				"webhook_id", event.WebhookID,
				"error", err)
			continue
		}
		processed++
	}

	return processed, nil
}

// processEvent processes a single retry attempt
func (w *RetryWorker) processEvent(ctx context.Context, event *WebhookEvent) error {
	// Get webhook details
	webhook, err := w.repo.GetByID(ctx, event.WebhookID)
	if err != nil {
		// If webhook doesn't exist, mark event as non-retryable
		markErr := w.repo.MarkEventAsNonRetryable(ctx, event.ID, "webhook not found")
		if markErr != nil {
			return fmt.Errorf("mark as non-retryable: %w", markErr)
		}
		return fmt.Errorf("webhook not found: %w", err)
	}

	// Check if webhook is enabled
	if !webhook.Enabled {
		// Mark as non-retryable if webhook is disabled
		err := w.repo.MarkEventAsNonRetryable(ctx, event.ID, "webhook disabled")
		if err != nil {
			return fmt.Errorf("mark as non-retryable: %w", err)
		}
		return fmt.Errorf("webhook disabled: %s", webhook.ID)
	}

	startTime := time.Now()

	// Attempt delivery
	result, err := w.deliverer.DeliverWebhook(ctx, webhook, event)

	processingTime := int(time.Since(startTime).Milliseconds())

	// Handle result
	if err == nil && result != nil && result.Success {
		// Success - mark as processed
		executionID := ""
		if result.Error == nil {
			executionID = fmt.Sprintf("retry-%s", event.ID)
		}
		markErr := w.repo.MarkEventRetrySucceeded(ctx, event.ID, executionID, processingTime)
		if markErr != nil {
			return fmt.Errorf("mark retry succeeded: %w", markErr)
		}

		w.logger.Info("webhook retry succeeded",
			"event_id", event.ID,
			"webhook_id", event.WebhookID,
			"retry_count", event.RetryCount+1)

		return nil
	}

	// Handle failure
	if err == nil && result != nil && result.Error != nil {
		err = result.Error
	}

	// Classify the error
	statusCode := 0
	if result != nil {
		statusCode = result.StatusCode
	}

	retryable := ClassifyWebhookError(err, statusCode)

	if !retryable {
		// Non-retryable error - mark as permanently failed
		errorMsg := "unknown error"
		if err != nil {
			errorMsg = err.Error()
		}
		markErr := w.repo.MarkEventAsNonRetryable(ctx, event.ID, errorMsg)
		if markErr != nil {
			return fmt.Errorf("mark as non-retryable: %w", markErr)
		}

		w.logger.Warn("webhook retry failed with non-retryable error",
			"event_id", event.ID,
			"webhook_id", event.WebhookID,
			"error", err)

		return fmt.Errorf("non-retryable error: %w", err)
	}

	// Retryable error - schedule for next retry
	errorMsg := "unknown error"
	if err != nil {
		errorMsg = err.Error()
	}

	markErr := w.repo.MarkEventForRetry(ctx, event.ID, errorMsg)
	if markErr != nil {
		return fmt.Errorf("mark for retry: %w", markErr)
	}

	w.logger.Info("webhook retry failed, will retry again",
		"event_id", event.ID,
		"webhook_id", event.WebhookID,
		"retry_count", event.RetryCount+1,
		"error", err)

	return nil
}

// ClassifyWebhookError determines if an error is retryable based on error type and status code
func ClassifyWebhookError(err error, statusCode int) bool {
	// If no error and successful status code, not retryable
	if err == nil && statusCode >= 200 && statusCode < 300 {
		return false
	}

	// Check specific error types
	if IsRetryableWebhookError(err) {
		return true
	}

	// Classify by status code
	if statusCode >= 500 && statusCode < 600 {
		// 5xx errors are retryable
		return true
	}

	if statusCode == 429 {
		// Rate limit errors are retryable
		return true
	}

	if statusCode >= 400 && statusCode < 500 {
		// 4xx errors (except 429) are not retryable
		return false
	}

	// If we have an error but no status code, assume it's a network error (retryable)
	if err != nil && statusCode == 0 {
		return true
	}

	// Default to not retryable
	return false
}
