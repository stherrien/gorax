package integrations

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// RetryableFunc is a function that can be retried
type RetryableFunc func() error

// IsRetryableError determines if an error should trigger a retry
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for known retryable errors
	if errors.Is(err, ErrRateLimitExceeded) {
		return true
	}

	// Add more retryable error patterns as needed
	return false
}

// WithRetry executes a function with exponential backoff retry logic
func WithRetry(ctx context.Context, config RetryConfig, fn RetryableFunc) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Check context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry if error is not retryable
		if !IsRetryableError(err) {
			return err
		}

		// Don't sleep after the last attempt
		if attempt < config.MaxAttempts-1 {
			delay := calculateDelay(attempt, config)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}

	return fmt.Errorf("max retry attempts exceeded: %w", lastErr)
}

// calculateDelay calculates the delay for exponential backoff
func calculateDelay(attempt int, config RetryConfig) time.Duration {
	// Exponential backoff: baseDelay * 2^attempt
	delay := config.BaseDelay * time.Duration(1<<uint(attempt))

	// Cap at max delay
	if delay > config.MaxDelay {
		return config.MaxDelay
	}

	return delay
}
