package executor

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"
)

func TestCircuitBreaker_InitialState(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := DefaultCircuitBreakerConfig()
	cb := NewCircuitBreaker("test", config, logger)

	if cb.GetState() != StateClosed {
		t.Errorf("initial state = %v, want %v", cb.GetState(), StateClosed)
	}

	if cb.GetFailureCount() != 0 {
		t.Errorf("initial failure count = %d, want 0", cb.GetFailureCount())
	}
}

func TestCircuitBreaker_SuccessfulRequests(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := DefaultCircuitBreakerConfig()
	cb := NewCircuitBreaker("test", config, logger)

	ctx := context.Background()

	// Execute multiple successful requests
	for i := 0; i < 10; i++ {
		err := cb.Execute(ctx, func(ctx context.Context) error {
			return nil
		})
		if err != nil {
			t.Errorf("Execute() error = %v, want nil", err)
		}
	}

	if cb.GetState() != StateClosed {
		t.Errorf("state = %v, want %v", cb.GetState(), StateClosed)
	}

	if cb.GetFailureCount() != 0 {
		t.Errorf("failure count = %d, want 0", cb.GetFailureCount())
	}
}

func TestCircuitBreaker_OpensAfterFailures(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := CircuitBreakerConfig{
		MaxFailures:       3,
		Timeout:           100 * time.Millisecond,
		MaxRequests:       2,
		FailureThreshold:  0.5,
		SlidingWindowSize: 10,
	}
	cb := NewCircuitBreaker("test", config, logger)

	ctx := context.Background()
	testErr := errors.New("test error")

	// Execute requests that fail
	for i := 0; i < 3; i++ {
		err := cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
		if err != testErr {
			t.Errorf("Execute() error = %v, want %v", err, testErr)
		}
	}

	// Circuit should be open after MaxFailures
	if cb.GetState() != StateOpen {
		t.Errorf("state = %v, want %v", cb.GetState(), StateOpen)
	}

	// Next request should fail immediately with ErrCircuitOpen
	err := cb.Execute(ctx, func(ctx context.Context) error {
		t.Error("operation should not be executed when circuit is open")
		return nil
	})

	if err != ErrCircuitOpen {
		t.Errorf("Execute() error = %v, want %v", err, ErrCircuitOpen)
	}
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := CircuitBreakerConfig{
		MaxFailures:       3,
		Timeout:           50 * time.Millisecond,
		MaxRequests:       2,
		FailureThreshold:  0.5,
		SlidingWindowSize: 10,
	}
	cb := NewCircuitBreaker("test", config, logger)

	ctx := context.Background()
	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 3; i++ {
		cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	if cb.GetState() != StateOpen {
		t.Fatalf("state = %v, want %v", cb.GetState(), StateOpen)
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Next request should transition to half-open
	err := cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if cb.GetState() != StateHalfOpen {
		t.Errorf("state = %v, want %v", cb.GetState(), StateHalfOpen)
	}
}

func TestCircuitBreaker_HalfOpenToCloseOnSuccess(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := CircuitBreakerConfig{
		MaxFailures:       3,
		Timeout:           50 * time.Millisecond,
		MaxRequests:       5,
		FailureThreshold:  0.5,
		SlidingWindowSize: 10,
	}
	cb := NewCircuitBreaker("test", config, logger)

	ctx := context.Background()
	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 3; i++ {
		cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	// Wait for timeout to transition to half-open
	time.Sleep(60 * time.Millisecond)

	// Execute enough successful requests to close the circuit
	for i := 0; i < 10; i++ {
		err := cb.Execute(ctx, func(ctx context.Context) error {
			return nil
		})
		if err != nil {
			t.Errorf("Execute() error = %v, want nil", err)
		}
	}

	// Circuit should close after successful requests
	if cb.GetState() != StateClosed {
		t.Errorf("state = %v, want %v", cb.GetState(), StateClosed)
	}
}

func TestCircuitBreaker_HalfOpenToOpenOnFailure(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := CircuitBreakerConfig{
		MaxFailures:       3,
		Timeout:           50 * time.Millisecond,
		MaxRequests:       2,
		FailureThreshold:  0.5,
		SlidingWindowSize: 10,
	}
	cb := NewCircuitBreaker("test", config, logger)

	ctx := context.Background()
	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 3; i++ {
		cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	// Wait for timeout to transition to half-open
	time.Sleep(60 * time.Millisecond)

	// Fail a request in half-open state
	err := cb.Execute(ctx, func(ctx context.Context) error {
		return testErr
	})

	if err != testErr {
		t.Errorf("Execute() error = %v, want %v", err, testErr)
	}

	// Circuit should reopen after failure in half-open state
	if cb.GetState() != StateOpen {
		t.Errorf("state = %v, want %v", cb.GetState(), StateOpen)
	}
}

func TestCircuitBreaker_MaxRequestsInHalfOpen(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := CircuitBreakerConfig{
		MaxFailures:       3,
		Timeout:           50 * time.Millisecond,
		MaxRequests:       2,
		FailureThreshold:  0.5,
		SlidingWindowSize: 10,
	}
	cb := NewCircuitBreaker("test", config, logger)

	ctx := context.Background()
	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 3; i++ {
		cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	// Wait for timeout to transition to half-open
	time.Sleep(60 * time.Millisecond)

	// Start MaxRequests concurrent operations
	var wg sync.WaitGroup
	successCount := 0
	tooManyRequestsCount := 0
	var mu sync.Mutex

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cb.Execute(ctx, func(ctx context.Context) error {
				time.Sleep(20 * time.Millisecond)
				return nil
			})

			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				successCount++
			} else if err == ErrTooManyRequests {
				tooManyRequestsCount++
			}
		}()
	}

	wg.Wait()

	// Should allow MaxRequests and reject others
	// Note: due to race conditions, we allow some tolerance
	if successCount < config.MaxRequests || successCount > config.MaxRequests+1 {
		t.Errorf("successCount = %d, want around %d", successCount, config.MaxRequests)
	}

	if tooManyRequestsCount < 5-config.MaxRequests-1 {
		t.Errorf("tooManyRequestsCount = %d, want at least %d", tooManyRequestsCount, 5-config.MaxRequests-1)
	}
}

func TestCircuitBreaker_ExecuteWithResult(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := DefaultCircuitBreakerConfig()
	cb := NewCircuitBreaker("test", config, logger)

	ctx := context.Background()
	expectedResult := "success"

	result, err := cb.ExecuteWithResult(ctx, func(ctx context.Context) (interface{}, error) {
		return expectedResult, nil
	})

	if err != nil {
		t.Errorf("ExecuteWithResult() error = %v, want nil", err)
	}

	if result != expectedResult {
		t.Errorf("ExecuteWithResult() result = %v, want %v", result, expectedResult)
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := CircuitBreakerConfig{
		MaxFailures:       3,
		Timeout:           100 * time.Millisecond,
		MaxRequests:       2,
		FailureThreshold:  0.5,
		SlidingWindowSize: 10,
	}
	cb := NewCircuitBreaker("test", config, logger)

	ctx := context.Background()
	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 3; i++ {
		cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	if cb.GetState() != StateOpen {
		t.Fatalf("state = %v, want %v", cb.GetState(), StateOpen)
	}

	// Reset the circuit breaker
	cb.Reset()

	if cb.GetState() != StateClosed {
		t.Errorf("state after reset = %v, want %v", cb.GetState(), StateClosed)
	}

	if cb.GetFailureCount() != 0 {
		t.Errorf("failure count after reset = %d, want 0", cb.GetFailureCount())
	}
}

func TestCircuitBreakerRegistry_GetOrCreate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := DefaultCircuitBreakerConfig()
	registry := NewCircuitBreakerRegistry(config, logger)

	cb1 := registry.GetOrCreate("breaker1")
	if cb1 == nil {
		t.Fatal("GetOrCreate() returned nil")
	}

	cb2 := registry.GetOrCreate("breaker1")
	if cb2 != cb1 {
		t.Error("GetOrCreate() should return the same instance for the same name")
	}

	cb3 := registry.GetOrCreate("breaker2")
	if cb3 == cb1 {
		t.Error("GetOrCreate() should return different instances for different names")
	}
}

func TestCircuitBreakerRegistry_Get(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := DefaultCircuitBreakerConfig()
	registry := NewCircuitBreakerRegistry(config, logger)

	// Get non-existent breaker
	cb, exists := registry.Get("nonexistent")
	if exists {
		t.Error("Get() returned exists=true for non-existent breaker")
	}
	if cb != nil {
		t.Error("Get() returned non-nil breaker for non-existent breaker")
	}

	// Create and get breaker
	registry.GetOrCreate("breaker1")
	cb, exists = registry.Get("breaker1")
	if !exists {
		t.Error("Get() returned exists=false for existing breaker")
	}
	if cb == nil {
		t.Fatal("Get() returned nil for existing breaker")
	}
}

func TestCircuitBreakerRegistry_Reset(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := CircuitBreakerConfig{
		MaxFailures:       3,
		Timeout:           100 * time.Millisecond,
		MaxRequests:       2,
		FailureThreshold:  0.5,
		SlidingWindowSize: 10,
	}
	registry := NewCircuitBreakerRegistry(config, logger)

	ctx := context.Background()
	testErr := errors.New("test error")

	// Create and open multiple breakers
	cb1 := registry.GetOrCreate("breaker1")
	cb2 := registry.GetOrCreate("breaker2")

	for i := 0; i < 3; i++ {
		cb1.Execute(ctx, func(ctx context.Context) error { return testErr })
		cb2.Execute(ctx, func(ctx context.Context) error { return testErr })
	}

	if cb1.GetState() != StateOpen {
		t.Fatalf("cb1 state = %v, want %v", cb1.GetState(), StateOpen)
	}
	if cb2.GetState() != StateOpen {
		t.Fatalf("cb2 state = %v, want %v", cb2.GetState(), StateOpen)
	}

	// Reset all breakers
	registry.Reset()

	if cb1.GetState() != StateClosed {
		t.Errorf("cb1 state after reset = %v, want %v", cb1.GetState(), StateClosed)
	}
	if cb2.GetState() != StateClosed {
		t.Errorf("cb2 state after reset = %v, want %v", cb2.GetState(), StateClosed)
	}
}

func TestCircuitBreakerRegistry_GetStats(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := DefaultCircuitBreakerConfig()
	registry := NewCircuitBreakerRegistry(config, logger)

	ctx := context.Background()

	// Create some breakers with different states
	cb1 := registry.GetOrCreate("breaker1")
	cb2 := registry.GetOrCreate("breaker2")

	// Keep cb1 closed with successful requests
	cb1.Execute(ctx, func(ctx context.Context) error { return nil })

	// Open cb2 with failures
	testErr := errors.New("test error")
	for i := 0; i < 5; i++ {
		cb2.Execute(ctx, func(ctx context.Context) error { return testErr })
	}

	stats := registry.GetStats()

	if len(stats) != 2 {
		t.Errorf("len(stats) = %d, want 2", len(stats))
	}

	breaker1Stats, ok := stats["breaker1"]
	if !ok {
		t.Fatal("stats missing breaker1")
	}
	if breaker1Stats.State != "closed" {
		t.Errorf("breaker1 state = %s, want closed", breaker1Stats.State)
	}

	breaker2Stats, ok := stats["breaker2"]
	if !ok {
		t.Fatal("stats missing breaker2")
	}
	if breaker2Stats.State != "open" {
		t.Errorf("breaker2 state = %s, want open", breaker2Stats.State)
	}
	if breaker2Stats.Failures < 5 {
		t.Errorf("breaker2 failures = %d, want >= 5", breaker2Stats.Failures)
	}
}

func TestCircuitState_String(t *testing.T) {
	tests := []struct {
		state    CircuitState
		expected string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{CircuitState(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.state.String()
			if result != tt.expected {
				t.Errorf("String() = %s, want %s", result, tt.expected)
			}
		})
	}
}
