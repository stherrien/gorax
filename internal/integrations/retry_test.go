package integrations

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithRetry_Success(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   time.Millisecond,
		MaxDelay:    time.Millisecond * 10,
	}

	attempts := 0
	fn := func() error {
		attempts++
		return nil
	}

	err := WithRetry(ctx, config, fn)

	assert.NoError(t, err)
	assert.Equal(t, 1, attempts, "should succeed on first attempt")
}

func TestWithRetry_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   time.Millisecond,
		MaxDelay:    time.Millisecond * 10,
	}

	attempts := 0
	fn := func() error {
		attempts++
		if attempts < 3 {
			return ErrRateLimitExceeded
		}
		return nil
	}

	err := WithRetry(ctx, config, fn)

	assert.NoError(t, err)
	assert.Equal(t, 3, attempts, "should succeed on third attempt")
}

func TestWithRetry_MaxAttemptsExceeded(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   time.Millisecond,
		MaxDelay:    time.Millisecond * 10,
	}

	attempts := 0
	fn := func() error {
		attempts++
		return ErrRateLimitExceeded
	}

	err := WithRetry(ctx, config, fn)

	assert.Error(t, err)
	assert.Equal(t, 3, attempts, "should attempt max times")
	assert.Contains(t, err.Error(), "max retry attempts exceeded")
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   time.Millisecond,
		MaxDelay:    time.Millisecond * 10,
	}

	attempts := 0
	expectedErr := errors.New("non-retryable error")
	fn := func() error {
		attempts++
		return expectedErr
	}

	err := WithRetry(ctx, config, fn)

	assert.Error(t, err)
	assert.Equal(t, 1, attempts, "should not retry non-retryable errors")
	assert.Equal(t, expectedErr, err)
}

func TestWithRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   time.Millisecond * 100,
		MaxDelay:    time.Second,
	}

	attempts := 0
	fn := func() error {
		attempts++
		if attempts == 1 {
			cancel()
		}
		return ErrRateLimitExceeded
	}

	err := WithRetry(ctx, config, fn)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "rate limit exceeded",
			err:      ErrRateLimitExceeded,
			expected: true,
		},
		{
			name:     "auth failed",
			err:      ErrAuthFailed,
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("generic error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateDelay(t *testing.T) {
	config := RetryConfig{
		BaseDelay: time.Second,
		MaxDelay:  time.Second * 30,
	}

	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{
			name:     "first retry",
			attempt:  0,
			expected: time.Second,
		},
		{
			name:     "second retry",
			attempt:  1,
			expected: time.Second * 2,
		},
		{
			name:     "third retry",
			attempt:  2,
			expected: time.Second * 4,
		},
		{
			name:     "capped at max delay",
			attempt:  10,
			expected: time.Second * 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateDelay(tt.attempt, config)
			assert.Equal(t, tt.expected, result)
		})
	}
}
