package executor

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestRetryStrategy_Execute_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}
	strategy := NewRetryStrategy(config, logger)

	ctx := context.Background()
	attempts := 0

	operation := func(ctx context.Context, attempt int) error {
		attempts++
		return nil
	}

	err := strategy.Execute(ctx, operation)
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if attempts != 1 {
		t.Errorf("attempts = %d, want 1", attempts)
	}
}

func TestRetryStrategy_Execute_RetryAndSuccess(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}
	strategy := NewRetryStrategy(config, logger)

	ctx := context.Background()
	attempts := 0

	operation := func(ctx context.Context, attempt int) error {
		attempts++
		if attempts < 3 {
			return errors.New("connection timeout")
		}
		return nil
	}

	start := time.Now()
	err := strategy.Execute(ctx, operation)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}

	// Check that backoff was applied (at least 10ms + 20ms = 30ms)
	minDuration := 30 * time.Millisecond
	if duration < minDuration {
		t.Errorf("duration = %v, want >= %v", duration, minDuration)
	}
}

func TestRetryStrategy_Execute_AllRetriesFailed(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}
	strategy := NewRetryStrategy(config, logger)

	ctx := context.Background()
	attempts := 0
	expectedErr := errors.New("connection timeout")

	operation := func(ctx context.Context, attempt int) error {
		attempts++
		return expectedErr
	}

	err := strategy.Execute(ctx, operation)
	if err != expectedErr {
		t.Errorf("Execute() error = %v, want %v", err, expectedErr)
	}

	// Should attempt: initial + 3 retries = 4 total
	if attempts != 4 {
		t.Errorf("attempts = %d, want 4", attempts)
	}
}

func TestRetryStrategy_Execute_PermanentError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}
	strategy := NewRetryStrategy(config, logger)

	ctx := context.Background()
	attempts := 0
	expectedErr := errors.New("invalid request")

	operation := func(ctx context.Context, attempt int) error {
		attempts++
		return expectedErr
	}

	err := strategy.Execute(ctx, operation)
	if err != expectedErr {
		t.Errorf("Execute() error = %v, want %v", err, expectedErr)
	}

	// Should not retry permanent errors
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1 (no retries for permanent errors)", attempts)
	}
}

func TestRetryStrategy_Execute_ContextCanceled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        1 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}
	strategy := NewRetryStrategy(config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0

	operation := func(ctx context.Context, attempt int) error {
		attempts++
		// Cancel after first attempt
		if attempts == 1 {
			cancel()
		}
		return errors.New("connection timeout")
	}

	err := strategy.Execute(ctx, operation)
	if err == nil {
		t.Error("Execute() error = nil, want error")
	}

	// Should stop retrying after context cancellation
	if attempts > 2 {
		t.Errorf("attempts = %d, want <= 2 (should stop after context cancel)", attempts)
	}
}

func TestRetryStrategy_ExecuteWithResult(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}
	strategy := NewRetryStrategy(config, logger)

	ctx := context.Background()
	attempts := 0
	expectedResult := "success"

	operation := func(ctx context.Context, attempt int) (interface{}, error) {
		attempts++
		if attempts < 2 {
			return nil, errors.New("connection timeout")
		}
		return expectedResult, nil
	}

	result, err := strategy.ExecuteWithResult(ctx, operation)
	if err != nil {
		t.Errorf("ExecuteWithResult() error = %v, want nil", err)
	}

	if result != expectedResult {
		t.Errorf("ExecuteWithResult() result = %v, want %v", result, expectedResult)
	}

	if attempts != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
}

func TestRetryStrategy_CalculateBackoff(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := RetryConfig{
		MaxRetries:        5,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        1 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}
	strategy := NewRetryStrategy(config, logger)

	tests := []struct {
		attempt         int
		expectedBackoff time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 400 * time.Millisecond},
		{3, 800 * time.Millisecond},
		{4, 1 * time.Second}, // Capped at MaxBackoff
		{5, 1 * time.Second}, // Capped at MaxBackoff
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.attempt)), func(t *testing.T) {
			backoff := strategy.calculateBackoff(tt.attempt)
			if backoff != tt.expectedBackoff {
				t.Errorf("calculateBackoff(%d) = %v, want %v", tt.attempt, backoff, tt.expectedBackoff)
			}
		})
	}
}

func TestRetryStrategy_CalculateBackoffWithJitter(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        1 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            true,
	}
	strategy := NewRetryStrategy(config, logger)

	// Calculate backoff multiple times and ensure jitter is applied
	backoff1 := strategy.calculateBackoff(1)
	backoff2 := strategy.calculateBackoff(1)

	// With jitter, backoffs should be in range [150ms, 250ms] (200ms ± 25%)
	minBackoff := 150 * time.Millisecond
	maxBackoff := 250 * time.Millisecond

	if backoff1 < minBackoff || backoff1 > maxBackoff {
		t.Errorf("backoff1 = %v, want in range [%v, %v]", backoff1, minBackoff, maxBackoff)
	}

	if backoff2 < minBackoff || backoff2 > maxBackoff {
		t.Errorf("backoff2 = %v, want in range [%v, %v]", backoff2, minBackoff, maxBackoff)
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", config.MaxRetries)
	}
	if config.InitialBackoff != 1*time.Second {
		t.Errorf("InitialBackoff = %v, want 1s", config.InitialBackoff)
	}
	if config.MaxBackoff != 30*time.Second {
		t.Errorf("MaxBackoff = %v, want 30s", config.MaxBackoff)
	}
	if config.BackoffMultiplier != 2.0 {
		t.Errorf("BackoffMultiplier = %f, want 2.0", config.BackoffMultiplier)
	}
	if !config.Jitter {
		t.Error("Jitter = false, want true")
	}
}

func TestDefaultNodeRetryConfig(t *testing.T) {
	config := DefaultNodeRetryConfig()

	if !config.Enabled {
		t.Error("Enabled = false, want true")
	}
	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", config.MaxRetries)
	}

	expectedCodes := []int{408, 429, 500, 502, 503, 504}
	if len(config.RetryableStatusCodes) != len(expectedCodes) {
		t.Errorf("len(RetryableStatusCodes) = %d, want %d", len(config.RetryableStatusCodes), len(expectedCodes))
	}

	for i, code := range expectedCodes {
		if config.RetryableStatusCodes[i] != code {
			t.Errorf("RetryableStatusCodes[%d] = %d, want %d", i, config.RetryableStatusCodes[i], code)
		}
	}
}
