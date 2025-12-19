package executor

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"time"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts (0 means no retries)
	MaxRetries int
	// InitialBackoff is the initial backoff duration
	InitialBackoff time.Duration
	// MaxBackoff is the maximum backoff duration
	MaxBackoff time.Duration
	// BackoffMultiplier is the multiplier for exponential backoff
	BackoffMultiplier float64
	// Jitter adds randomness to backoff to prevent thundering herd
	Jitter bool
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        30 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            true,
	}
}

// NodeRetryConfig extends RetryConfig with node-specific settings
type NodeRetryConfig struct {
	RetryConfig
	// Enabled determines if retry is enabled for this node
	Enabled bool
	// RetryableStatusCodes for HTTP actions (optional)
	RetryableStatusCodes []int
}

// DefaultNodeRetryConfig returns the default node retry configuration
func DefaultNodeRetryConfig() NodeRetryConfig {
	return NodeRetryConfig{
		RetryConfig:          DefaultRetryConfig(),
		Enabled:              true,
		RetryableStatusCodes: []int{408, 429, 500, 502, 503, 504},
	}
}

// RetryableOperation is a function that can be retried
type RetryableOperation func(ctx context.Context, attempt int) error

// RetryStrategy handles retry logic with exponential backoff
type RetryStrategy struct {
	config RetryConfig
	logger *slog.Logger
}

// NewRetryStrategy creates a new retry strategy
func NewRetryStrategy(config RetryConfig, logger *slog.Logger) *RetryStrategy {
	return &RetryStrategy{
		config: config,
		logger: logger,
	}
}

// Execute runs an operation with retry logic
func (r *RetryStrategy) Execute(ctx context.Context, operation RetryableOperation) error {
	var lastErr error

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		// Execute the operation
		err := operation(ctx, attempt)
		if err == nil {
			// Success
			if attempt > 0 {
				r.logger.Info("operation succeeded after retry",
					"attempt", attempt,
					"max_retries", r.config.MaxRetries,
				)
			}
			return nil
		}

		lastErr = err

		// Check if we should retry
		if attempt >= r.config.MaxRetries {
			r.logger.Error("operation failed after all retries",
				"attempts", attempt+1,
				"max_retries", r.config.MaxRetries,
				"error", err,
			)
			break
		}

		// Check if error is retryable
		if !ShouldRetry(err, attempt, r.config.MaxRetries) {
			r.logger.Info("operation failed with non-retryable error",
				"attempt", attempt+1,
				"error", err,
			)
			return err
		}

		// Calculate backoff duration
		backoff := r.calculateBackoff(attempt)

		r.logger.Info("operation failed, retrying",
			"attempt", attempt+1,
			"max_retries", r.config.MaxRetries,
			"backoff", backoff,
			"error", err,
		)

		// Wait for backoff duration or context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(backoff):
			// Continue to next retry
		}
	}

	return lastErr
}

// ExecuteWithResult runs an operation that returns a result with retry logic
func (r *RetryStrategy) ExecuteWithResult(ctx context.Context, operation func(ctx context.Context, attempt int) (interface{}, error)) (interface{}, error) {
	var lastErr error
	var result interface{}

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		// Execute the operation
		res, err := operation(ctx, attempt)
		if err == nil {
			// Success
			if attempt > 0 {
				r.logger.Info("operation succeeded after retry",
					"attempt", attempt,
					"max_retries", r.config.MaxRetries,
				)
			}
			return res, nil
		}

		lastErr = err
		result = res

		// Check if we should retry
		if attempt >= r.config.MaxRetries {
			r.logger.Error("operation failed after all retries",
				"attempts", attempt+1,
				"max_retries", r.config.MaxRetries,
				"error", err,
			)
			break
		}

		// Check if error is retryable
		if !ShouldRetry(err, attempt, r.config.MaxRetries) {
			r.logger.Info("operation failed with non-retryable error",
				"attempt", attempt+1,
				"error", err,
			)
			return result, err
		}

		// Calculate backoff duration
		backoff := r.calculateBackoff(attempt)

		r.logger.Info("operation failed, retrying",
			"attempt", attempt+1,
			"max_retries", r.config.MaxRetries,
			"backoff", backoff,
			"error", err,
		)

		// Wait for backoff duration or context cancellation
		select {
		case <-ctx.Done():
			return result, fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(backoff):
			// Continue to next retry
		}
	}

	return result, lastErr
}

// calculateBackoff calculates the backoff duration for the given attempt
func (r *RetryStrategy) calculateBackoff(attempt int) time.Duration {
	// Calculate exponential backoff: initialBackoff * (multiplier ^ attempt)
	backoff := float64(r.config.InitialBackoff) * math.Pow(r.config.BackoffMultiplier, float64(attempt))

	// Apply max backoff limit
	if backoff > float64(r.config.MaxBackoff) {
		backoff = float64(r.config.MaxBackoff)
	}

	duration := time.Duration(backoff)

	// Add jitter if enabled (random variation of ±25%)
	if r.config.Jitter {
		jitter := float64(duration) * 0.25
		variation := (rand.Float64() * 2 * jitter) - jitter // Random value between -jitter and +jitter
		duration = time.Duration(float64(duration) + variation)
	}

	return duration
}

// GetAttemptNumber returns the current attempt number (0-indexed)
func GetAttemptNumber(ctx context.Context) int {
	if attempt, ok := ctx.Value(attemptKey).(int); ok {
		return attempt
	}
	return 0
}

// attemptKey is the context key for storing retry attempt number
type contextKey string

const attemptKey contextKey = "retry_attempt"

// withAttempt returns a context with the attempt number
func withAttempt(ctx context.Context, attempt int) context.Context {
	return context.WithValue(ctx, attemptKey, attempt)
}
