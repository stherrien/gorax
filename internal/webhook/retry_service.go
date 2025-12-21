package webhook

import (
	"context"
	cryptoRand "crypto/rand"
	"encoding/binary"
	"errors"
	"math"
	"math/rand"
	"time"
)

// secureRand is a math/rand source seeded with crypto/rand for jitter calculations
var secureRand *rand.Rand

func init() {
	var seed int64
	if err := binary.Read(cryptoRand.Reader, binary.BigEndian, &seed); err != nil {
		// Fallback to time-based seed if crypto/rand fails
		seed = time.Now().UnixNano()
	}
	secureRand = rand.New(rand.NewSource(seed))
}

// Webhook-specific retry errors
var (
	ErrWebhookTimeout          = errors.New("webhook delivery timeout")
	ErrWebhookRateLimited      = errors.New("webhook rate limited")
	ErrWebhookConnectionFailed = errors.New("webhook connection failed")
	ErrWebhookServerError      = errors.New("webhook server error")
	ErrWebhookValidationFailed = errors.New("webhook validation failed")
	ErrWebhookAuthFailed       = errors.New("webhook authentication failed")
	ErrMaxRetriesExceeded      = errors.New("maximum retry attempts exceeded")
)

// RetryConfig holds configuration for webhook retry behavior
type RetryConfig struct {
	MaxAttempts int           // Maximum number of retry attempts (including initial attempt)
	BaseDelay   time.Duration // Base delay for exponential backoff
	MaxDelay    time.Duration // Maximum delay cap
	Multiplier  float64       // Multiplier for exponential backoff
	Jitter      float64       // Jitter factor (0.0 - 1.0)
}

// DefaultRetryConfig returns sensible default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   time.Second,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
		Jitter:      0.0,
	}
}

// RetryResult represents the result of a webhook delivery attempt
type RetryResult struct {
	Success      bool
	StatusCode   int
	ResponseBody []byte
	Error        error
}

// RetryableFunc is a function that can be retried
type RetryableFunc func(ctx context.Context, attempt int) (*RetryResult, error)

// CalculateBackoff calculates the delay for exponential backoff
func CalculateBackoff(attempt int, config RetryConfig) time.Duration {
	// Exponential backoff: baseDelay * multiplier^attempt
	delay := time.Duration(float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(attempt)))

	// Cap at max delay
	if delay > config.MaxDelay {
		return config.MaxDelay
	}

	return delay
}

// CalculateBackoffWithJitter calculates the delay with jitter
func CalculateBackoffWithJitter(attempt int, config RetryConfig) time.Duration {
	baseDelay := CalculateBackoff(attempt, config)

	if config.Jitter <= 0 {
		return baseDelay
	}

	// Add jitter: +/- jitter% of the base delay
	// Using secureRand which is seeded from crypto/rand
	jitterRange := float64(baseDelay) * config.Jitter
	jitter := (secureRand.Float64()*2 - 1) * jitterRange // Random between -jitterRange and +jitterRange

	finalDelay := time.Duration(float64(baseDelay) + jitter)
	if finalDelay < 0 {
		return baseDelay
	}

	return finalDelay
}

// IsRetryableWebhookError determines if an error should trigger a retry
func IsRetryableWebhookError(err error) bool {
	if err == nil {
		return false
	}

	// These errors are retryable
	if errors.Is(err, ErrWebhookTimeout) ||
		errors.Is(err, ErrWebhookRateLimited) ||
		errors.Is(err, ErrWebhookConnectionFailed) ||
		errors.Is(err, ErrWebhookServerError) {
		return true
	}

	// These errors are NOT retryable
	if errors.Is(err, ErrWebhookValidationFailed) ||
		errors.Is(err, ErrWebhookAuthFailed) {
		return false
	}

	return false
}

// WithWebhookRetry executes a function with exponential backoff retry logic
func WithWebhookRetry(ctx context.Context, config RetryConfig, fn RetryableFunc) (*RetryResult, error) {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Check context cancellation before each attempt
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Execute the function
		result, err := fn(ctx, attempt)
		if err == nil && result != nil && result.Success {
			return result, nil
		}

		lastErr = err
		if lastErr == nil && result != nil && result.Error != nil {
			lastErr = result.Error
		}

		// Don't retry if error is not retryable
		if !IsRetryableWebhookError(lastErr) {
			return nil, lastErr
		}

		// Don't sleep after the last attempt
		if attempt < config.MaxAttempts-1 {
			delay := CalculateBackoffWithJitter(attempt, config)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}

	return nil, errors.Join(ErrMaxRetriesExceeded, lastErr)
}

// WebhookRetryState tracks the retry state for a webhook delivery
type WebhookRetryState struct {
	WebhookID   string
	EventID     string
	Attempts    int
	LastError   error
	NextRetryAt time.Time
	CreatedAt   time.Time
}

// NewWebhookRetryState creates a new retry state tracker
func NewWebhookRetryState(webhookID, eventID string) *WebhookRetryState {
	return &WebhookRetryState{
		WebhookID:   webhookID,
		EventID:     eventID,
		Attempts:    0,
		NextRetryAt: time.Now(),
		CreatedAt:   time.Now(),
	}
}

// IncrementAttempt increments the attempt counter and calculates next retry time
func (s *WebhookRetryState) IncrementAttempt(err error, config RetryConfig) {
	s.Attempts++
	s.LastError = err
	s.NextRetryAt = time.Now().Add(s.GetNextRetryDelay(config))
}

// ShouldRetry determines if the webhook should be retried
func (s *WebhookRetryState) ShouldRetry(config RetryConfig) bool {
	if s.LastError == nil {
		return false
	}

	if s.Attempts >= config.MaxAttempts {
		return false
	}

	return IsRetryableWebhookError(s.LastError)
}

// GetNextRetryDelay returns the delay until the next retry
func (s *WebhookRetryState) GetNextRetryDelay(config RetryConfig) time.Duration {
	return CalculateBackoff(s.Attempts, config)
}
