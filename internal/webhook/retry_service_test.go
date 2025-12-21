package webhook

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryConfig_Defaults(t *testing.T) {
	config := DefaultRetryConfig()

	assert.Equal(t, 3, config.MaxAttempts)
	assert.Equal(t, time.Second, config.BaseDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.Multiplier)
}

func TestRetryConfig_Custom(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 5,
		BaseDelay:   500 * time.Millisecond,
		MaxDelay:    60 * time.Second,
		Multiplier:  3.0,
	}

	assert.Equal(t, 5, config.MaxAttempts)
	assert.Equal(t, 500*time.Millisecond, config.BaseDelay)
	assert.Equal(t, 60*time.Second, config.MaxDelay)
	assert.Equal(t, 3.0, config.Multiplier)
}

func TestCalculateBackoff(t *testing.T) {
	config := RetryConfig{
		BaseDelay:  time.Second,
		MaxDelay:   30 * time.Second,
		Multiplier: 2.0,
	}

	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{"first attempt", 0, time.Second},
		{"second attempt", 1, 2 * time.Second},
		{"third attempt", 2, 4 * time.Second},
		{"fourth attempt", 3, 8 * time.Second},
		{"fifth attempt", 4, 16 * time.Second},
		{"capped at max", 10, 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := CalculateBackoff(tt.attempt, config)
			assert.Equal(t, tt.expected, delay)
		})
	}
}

func TestCalculateBackoffWithJitter(t *testing.T) {
	config := RetryConfig{
		BaseDelay:  time.Second,
		MaxDelay:   30 * time.Second,
		Multiplier: 2.0,
		Jitter:     0.1, // 10% jitter
	}

	// Run multiple times to verify jitter introduces variability
	delays := make(map[time.Duration]bool)
	for i := 0; i < 100; i++ {
		delay := CalculateBackoffWithJitter(1, config)
		delays[delay] = true

		// Should be around 2 seconds +/- 10%
		assert.GreaterOrEqual(t, delay, time.Duration(float64(2*time.Second)*0.9))
		assert.LessOrEqual(t, delay, time.Duration(float64(2*time.Second)*1.1))
	}

	// With jitter, we should see some variability
	// (unless we get very unlucky with random numbers)
	assert.Greater(t, len(delays), 1, "Expected jitter to produce different delay values")
}

func TestIsRetryableWebhookError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{"nil error", nil, false},
		{"timeout error", ErrWebhookTimeout, true},
		{"rate limit error", ErrWebhookRateLimited, true},
		{"connection error", ErrWebhookConnectionFailed, true},
		{"server error 500", ErrWebhookServerError, true},
		{"validation error", ErrWebhookValidationFailed, false},
		{"auth error", ErrWebhookAuthFailed, false},
		{"generic error", errors.New("some error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableWebhookError(tt.err)
			assert.Equal(t, tt.retryable, result)
		})
	}
}

func TestRetryableOperation_Success(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   10 * time.Millisecond,
		MaxDelay:    100 * time.Millisecond,
		Multiplier:  2.0,
	}

	attempts := 0
	result, err := WithWebhookRetry(ctx, config, func(ctx context.Context, attempt int) (*RetryResult, error) {
		attempts++
		return &RetryResult{
			Success:    true,
			StatusCode: 200,
		}, nil
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 200, result.StatusCode)
	assert.Equal(t, 1, attempts, "Should succeed on first attempt")
}

func TestRetryableOperation_SucceedsAfterRetries(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   10 * time.Millisecond,
		MaxDelay:    100 * time.Millisecond,
		Multiplier:  2.0,
	}

	attempts := 0
	result, err := WithWebhookRetry(ctx, config, func(ctx context.Context, attempt int) (*RetryResult, error) {
		attempts++
		if attempts < 3 {
			return nil, ErrWebhookTimeout
		}
		return &RetryResult{
			Success:    true,
			StatusCode: 200,
		}, nil
	})

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 3, attempts)
}

func TestRetryableOperation_ExhaustsRetries(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   10 * time.Millisecond,
		MaxDelay:    100 * time.Millisecond,
		Multiplier:  2.0,
	}

	attempts := 0
	result, err := WithWebhookRetry(ctx, config, func(ctx context.Context, attempt int) (*RetryResult, error) {
		attempts++
		return nil, ErrWebhookTimeout
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 3, attempts)
	assert.ErrorIs(t, err, ErrMaxRetriesExceeded)
}

func TestRetryableOperation_NonRetryableError(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   10 * time.Millisecond,
		MaxDelay:    100 * time.Millisecond,
		Multiplier:  2.0,
	}

	attempts := 0
	result, err := WithWebhookRetry(ctx, config, func(ctx context.Context, attempt int) (*RetryResult, error) {
		attempts++
		return nil, ErrWebhookAuthFailed // Not retryable
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 1, attempts, "Should not retry non-retryable errors")
	assert.ErrorIs(t, err, ErrWebhookAuthFailed)
}

func TestRetryableOperation_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := RetryConfig{
		MaxAttempts: 5,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    1 * time.Second,
		Multiplier:  2.0,
	}

	attempts := 0
	var wg sync.WaitGroup
	wg.Add(1)

	var result *RetryResult
	var err error

	go func() {
		defer wg.Done()
		result, err = WithWebhookRetry(ctx, config, func(ctx context.Context, attempt int) (*RetryResult, error) {
			attempts++
			return nil, ErrWebhookTimeout
		})
	}()

	// Cancel after first attempt
	time.Sleep(50 * time.Millisecond)
	cancel()
	wg.Wait()

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, result)
	assert.LessOrEqual(t, attempts, 2, "Should stop retrying after context cancellation")
}

func TestRetryableOperation_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	config := RetryConfig{
		MaxAttempts: 10,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    1 * time.Second,
		Multiplier:  2.0,
	}

	attempts := 0
	result, err := WithWebhookRetry(ctx, config, func(ctx context.Context, attempt int) (*RetryResult, error) {
		attempts++
		return nil, ErrWebhookTimeout
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.LessOrEqual(t, attempts, 2, "Should stop when context times out")
}

func TestRetryableOperation_TracksAttempts(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   10 * time.Millisecond,
		MaxDelay:    100 * time.Millisecond,
		Multiplier:  2.0,
	}

	attemptsSeen := []int{}
	_, _ = WithWebhookRetry(ctx, config, func(ctx context.Context, attempt int) (*RetryResult, error) {
		attemptsSeen = append(attemptsSeen, attempt)
		if attempt < 2 {
			return nil, ErrWebhookTimeout
		}
		return &RetryResult{Success: true}, nil
	})

	assert.Equal(t, []int{0, 1, 2}, attemptsSeen)
}

func TestWebhookRetryState(t *testing.T) {
	state := NewWebhookRetryState("webhook-123", "event-456")

	assert.Equal(t, "webhook-123", state.WebhookID)
	assert.Equal(t, "event-456", state.EventID)
	assert.Equal(t, 0, state.Attempts)
	assert.Nil(t, state.LastError)
	assert.False(t, state.NextRetryAt.IsZero())
}

func TestWebhookRetryState_IncrementAttempt(t *testing.T) {
	config := DefaultRetryConfig()
	state := NewWebhookRetryState("webhook-123", "event-456")

	state.IncrementAttempt(ErrWebhookTimeout, config)

	assert.Equal(t, 1, state.Attempts)
	assert.Equal(t, ErrWebhookTimeout, state.LastError)
	assert.False(t, state.NextRetryAt.IsZero())
	assert.True(t, state.NextRetryAt.After(time.Now()))
}

func TestWebhookRetryState_ShouldRetry(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   time.Second,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
	}

	tests := []struct {
		name        string
		attempts    int
		lastError   error
		shouldRetry bool
	}{
		{"first attempt", 0, ErrWebhookTimeout, true},
		{"second attempt", 1, ErrWebhookTimeout, true},
		{"max attempts", 3, ErrWebhookTimeout, false},
		{"non-retryable error", 0, ErrWebhookAuthFailed, false},
		{"nil error", 0, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &WebhookRetryState{
				Attempts:  tt.attempts,
				LastError: tt.lastError,
			}
			assert.Equal(t, tt.shouldRetry, state.ShouldRetry(config))
		})
	}
}

func TestWebhookRetryState_GetNextRetryDelay(t *testing.T) {
	config := RetryConfig{
		BaseDelay:  time.Second,
		MaxDelay:   30 * time.Second,
		Multiplier: 2.0,
	}

	state := &WebhookRetryState{Attempts: 2}
	delay := state.GetNextRetryDelay(config)

	// After 2 attempts, delay should be 4 seconds (1 * 2^2)
	assert.Equal(t, 4*time.Second, delay)
}
