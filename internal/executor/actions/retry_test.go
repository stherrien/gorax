package actions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gorax/gorax/internal/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryAction_Execute_Success(t *testing.T) {
	// Test successful execution on first attempt
	callCount := 0
	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		callCount++
		return map[string]interface{}{"result": "success"}, nil
	}

	action := NewRetryAction(executeNode)

	config := workflow.RetryNodeConfig{
		Strategy:       "fixed",
		MaxAttempts:    3,
		InitialDelayMs: 100,
	}

	input := NewActionInput(config, map[string]interface{}{
		"retry_node_id": "node1",
	})

	output, err := action.Execute(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 1, callCount)

	data, ok := output.Data.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, data["success"].(bool))
	assert.Equal(t, 0, data["retries"].(int))
}

func TestRetryAction_Execute_SuccessAfterRetries(t *testing.T) {
	// Test success after multiple retries
	callCount := 0
	testErr := errors.New("temporary error")

	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		callCount++
		if callCount < 3 {
			return nil, testErr
		}
		return map[string]interface{}{"result": "success"}, nil
	}

	action := NewRetryAction(executeNode)

	config := workflow.RetryNodeConfig{
		Strategy:        "fixed",
		MaxAttempts:     5,
		InitialDelayMs:  10, // Short delay for testing
		RetryableErrors: []string{"temporary"},
	}

	input := NewActionInput(config, map[string]interface{}{
		"retry_node_id": "node1",
	})

	start := time.Now()
	output, err := action.Execute(context.Background(), input)
	duration := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 3, callCount)

	// Should have delayed at least 2 times (10ms each)
	assert.GreaterOrEqual(t, duration, 20*time.Millisecond)

	data, ok := output.Data.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, data["success"].(bool))
	assert.Equal(t, 2, data["retries"].(int))

	attempts := data["attempts"].([]RetryAttempt)
	assert.Len(t, attempts, 3)
	assert.False(t, attempts[0].Success)
	assert.False(t, attempts[1].Success)
	assert.True(t, attempts[2].Success)
}

func TestRetryAction_Execute_MaxRetriesExceeded(t *testing.T) {
	// Test max retries exceeded
	callCount := 0
	testErr := errors.New("persistent error")

	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		callCount++
		return nil, testErr
	}

	action := NewRetryAction(executeNode)

	config := workflow.RetryNodeConfig{
		Strategy:        "fixed",
		MaxAttempts:     3,
		InitialDelayMs:  10,
		RetryableErrors: []string{"error"},
	}

	input := NewActionInput(config, map[string]interface{}{
		"retry_node_id": "node1",
	})

	output, err := action.Execute(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "max retries exceeded")
	assert.Equal(t, 4, callCount) // Initial + 3 retries
}

func TestRetryAction_Execute_NonRetryableError(t *testing.T) {
	// Test non-retryable error
	callCount := 0
	testErr := errors.New("invalid request")

	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		callCount++
		return nil, testErr
	}

	action := NewRetryAction(executeNode)

	config := workflow.RetryNodeConfig{
		Strategy:           "fixed",
		MaxAttempts:        3,
		InitialDelayMs:     10,
		NonRetryableErrors: []string{"invalid"},
	}

	input := NewActionInput(config, map[string]interface{}{
		"retry_node_id": "node1",
	})

	output, err := action.Execute(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "non-retryable error")
	assert.Equal(t, 1, callCount) // Should not retry
}

func TestRetryAction_Execute_FixedDelay(t *testing.T) {
	// Test fixed delay strategy
	callCount := 0
	testErr := errors.New("temporary error")

	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		callCount++
		if callCount < 3 {
			return nil, testErr
		}
		return map[string]interface{}{"result": "success"}, nil
	}

	action := NewRetryAction(executeNode)

	config := workflow.RetryNodeConfig{
		Strategy:        "fixed",
		MaxAttempts:     5,
		InitialDelayMs:  50,
		RetryableErrors: []string{"temporary"},
	}

	input := NewActionInput(config, map[string]interface{}{
		"retry_node_id": "node1",
	})

	start := time.Now()
	output, err := action.Execute(context.Background(), input)
	duration := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, output)

	// Should have delayed 2 times with fixed 50ms
	assert.GreaterOrEqual(t, duration, 100*time.Millisecond)
	assert.Less(t, duration, 150*time.Millisecond) // Allow some margin

	data, ok := output.Data.(map[string]interface{})
	require.True(t, ok)
	attempts := data["attempts"].([]RetryAttempt)
	assert.Equal(t, 50, attempts[0].DelayMs)
	assert.Equal(t, 50, attempts[1].DelayMs)
}

func TestRetryAction_Execute_ExponentialBackoff(t *testing.T) {
	// Test exponential backoff strategy
	callCount := 0
	testErr := errors.New("temporary error")

	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		callCount++
		if callCount < 3 {
			return nil, testErr
		}
		return map[string]interface{}{"result": "success"}, nil
	}

	action := NewRetryAction(executeNode)

	config := workflow.RetryNodeConfig{
		Strategy:        "exponential",
		MaxAttempts:     5,
		InitialDelayMs:  20,
		MaxDelayMs:      1000,
		Multiplier:      2.0,
		RetryableErrors: []string{"temporary"},
	}

	input := NewActionInput(config, map[string]interface{}{
		"retry_node_id": "node1",
	})

	start := time.Now()
	output, err := action.Execute(context.Background(), input)
	duration := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, output)

	// Should have exponential delays: 20ms, 40ms
	assert.GreaterOrEqual(t, duration, 60*time.Millisecond)

	data, ok := output.Data.(map[string]interface{})
	require.True(t, ok)
	attempts := data["attempts"].([]RetryAttempt)

	// Verify exponential growth
	assert.Equal(t, 20, attempts[0].DelayMs)
	assert.Equal(t, 40, attempts[1].DelayMs)
}

func TestRetryAction_Execute_ExponentialWithJitter(t *testing.T) {
	// Test exponential backoff with jitter
	callCount := 0
	testErr := errors.New("temporary error")

	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		callCount++
		if callCount < 3 {
			return nil, testErr
		}
		return map[string]interface{}{"result": "success"}, nil
	}

	action := NewRetryAction(executeNode)

	config := workflow.RetryNodeConfig{
		Strategy:        "exponential_jitter",
		MaxAttempts:     5,
		InitialDelayMs:  20,
		MaxDelayMs:      1000,
		Multiplier:      2.0,
		Jitter:          true,
		RetryableErrors: []string{"temporary"},
	}

	input := NewActionInput(config, map[string]interface{}{
		"retry_node_id": "node1",
	})

	output, err := action.Execute(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, output)

	data, ok := output.Data.(map[string]interface{})
	require.True(t, ok)
	attempts := data["attempts"].([]RetryAttempt)

	// With jitter, delays should vary but be roughly exponential
	// First delay should be around 20ms ± 25%
	assert.GreaterOrEqual(t, attempts[0].DelayMs, 15)
	assert.LessOrEqual(t, attempts[0].DelayMs, 25)

	// Second delay should be around 40ms ± 25%
	assert.GreaterOrEqual(t, attempts[1].DelayMs, 30)
	assert.LessOrEqual(t, attempts[1].DelayMs, 50)
}

func TestRetryAction_Execute_MaxDelayLimit(t *testing.T) {
	// Test max delay limit
	callCount := 0
	testErr := errors.New("temporary error")

	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		callCount++
		if callCount <= 5 {
			return nil, testErr
		}
		return map[string]interface{}{"result": "success"}, nil
	}

	action := NewRetryAction(executeNode)

	config := workflow.RetryNodeConfig{
		Strategy:        "exponential",
		MaxAttempts:     10,
		InitialDelayMs:  10,
		MaxDelayMs:      50, // Cap at 50ms
		Multiplier:      2.0,
		RetryableErrors: []string{"temporary"},
	}

	input := NewActionInput(config, map[string]interface{}{
		"retry_node_id": "node1",
	})

	output, err := action.Execute(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, output)

	data, ok := output.Data.(map[string]interface{})
	require.True(t, ok)
	attempts := data["attempts"].([]RetryAttempt)

	// Delays should not exceed max: 10, 20, 40, 50, 50, 50
	for i, attempt := range attempts {
		if i > 0 {
			assert.LessOrEqual(t, attempt.DelayMs, 50, "Delay at attempt %d exceeded max", i)
		}
	}
}

func TestRetryAction_Execute_ContextCancellation(t *testing.T) {
	// Test context cancellation during retry
	testErr := errors.New("temporary error")

	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		return nil, testErr
	}

	action := NewRetryAction(executeNode)

	config := workflow.RetryNodeConfig{
		Strategy:        "fixed",
		MaxAttempts:     10,
		InitialDelayMs:  100, // Long delay
		RetryableErrors: []string{"temporary"},
	}

	input := NewActionInput(config, map[string]interface{}{
		"retry_node_id": "node1",
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context after short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	output, err := action.Execute(ctx, input)
	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "retry cancelled")
}

func TestRetryAction_Execute_MissingNodeID(t *testing.T) {
	// Test missing retry_node_id in context
	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		return nil, nil
	}

	action := NewRetryAction(executeNode)

	config := workflow.RetryNodeConfig{
		Strategy:       "fixed",
		MaxAttempts:    3,
		InitialDelayMs: 100,
	}

	input := NewActionInput(config, map[string]interface{}{
		// Missing retry_node_id
	})

	output, err := action.Execute(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "retry_node_id not found")
}

func TestRetryAction_ValidateConfig(t *testing.T) {
	action := NewRetryAction(nil)

	tests := []struct {
		name    string
		config  workflow.RetryNodeConfig
		wantErr bool
	}{
		{
			name: "valid fixed strategy",
			config: workflow.RetryNodeConfig{
				Strategy:       "fixed",
				MaxAttempts:    3,
				InitialDelayMs: 100,
			},
			wantErr: false,
		},
		{
			name: "valid exponential strategy",
			config: workflow.RetryNodeConfig{
				Strategy:       "exponential",
				MaxAttempts:    3,
				InitialDelayMs: 100,
				MaxDelayMs:     1000,
				Multiplier:     2.0,
			},
			wantErr: false,
		},
		{
			name: "negative max attempts",
			config: workflow.RetryNodeConfig{
				Strategy:       "fixed",
				MaxAttempts:    -1,
				InitialDelayMs: 100,
			},
			wantErr: true,
		},
		{
			name: "negative initial delay",
			config: workflow.RetryNodeConfig{
				Strategy:       "fixed",
				MaxAttempts:    3,
				InitialDelayMs: -100,
			},
			wantErr: true,
		},
		{
			name: "invalid strategy",
			config: workflow.RetryNodeConfig{
				Strategy:       "unknown",
				MaxAttempts:    3,
				InitialDelayMs: 100,
			},
			wantErr: true,
		},
		{
			name: "max delay less than initial delay",
			config: workflow.RetryNodeConfig{
				Strategy:       "exponential",
				MaxAttempts:    3,
				InitialDelayMs: 1000,
				MaxDelayMs:     100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := action.validateConfig(&tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFixedDelayStrategy_CalculateDelay(t *testing.T) {
	strategy := &FixedDelayStrategy{
		Delay: 100 * time.Millisecond,
	}

	// All delays should be fixed
	for i := 0; i < 5; i++ {
		delay := strategy.CalculateDelay(i)
		assert.Equal(t, 100*time.Millisecond, delay)
	}
}

func TestExponentialBackoffStrategy_CalculateDelay(t *testing.T) {
	strategy := &ExponentialBackoffStrategy{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1000 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	// Test exponential growth
	delays := []time.Duration{
		100 * time.Millisecond,  // 100 * 2^0
		200 * time.Millisecond,  // 100 * 2^1
		400 * time.Millisecond,  // 100 * 2^2
		800 * time.Millisecond,  // 100 * 2^3
		1000 * time.Millisecond, // Capped at max
		1000 * time.Millisecond, // Stays at max
	}

	for i, expected := range delays {
		delay := strategy.CalculateDelay(i)
		assert.Equal(t, expected, delay, "Attempt %d", i)
	}
}

func TestExponentialBackoffStrategy_WithJitter(t *testing.T) {
	strategy := &ExponentialBackoffStrategy{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1000 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       true,
	}

	// With jitter, delays should vary but stay within bounds
	delay := strategy.CalculateDelay(0)
	// Should be around 100ms ± 25%
	assert.GreaterOrEqual(t, delay, 75*time.Millisecond)
	assert.LessOrEqual(t, delay, 125*time.Millisecond)

	delay2 := strategy.CalculateDelay(1)
	// Should be around 200ms ± 25%
	assert.GreaterOrEqual(t, delay2, 150*time.Millisecond)
	assert.LessOrEqual(t, delay2, 250*time.Millisecond)
}
