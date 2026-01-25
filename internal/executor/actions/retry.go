package actions

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	mathRand "math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/gorax/gorax/internal/workflow"
)

// RetryAction implements configurable retry logic for actions
type RetryAction struct {
	executeNodeFunc func(ctx context.Context, nodeID string) (interface{}, error)
}

// NewRetryAction creates a new retry action
func NewRetryAction(executeNode func(ctx context.Context, nodeID string) (interface{}, error)) *RetryAction {
	return &RetryAction{
		executeNodeFunc: executeNode,
	}
}

// Execute executes an action with retry logic
func (a *RetryAction) Execute(ctx context.Context, input *ActionInput) (*ActionOutput, error) {
	// Parse configuration
	var config workflow.RetryNodeConfig
	configBytes, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to parse retry config: %w", err)
	}

	// Validate configuration
	if err := a.validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid retry config: %w", err)
	}

	// Get the node ID to retry from context
	nodeID, ok := input.Context["retry_node_id"].(string)
	if !ok || nodeID == "" {
		return nil, fmt.Errorf("retry_node_id not found in context")
	}

	// Create retry strategy
	strategy := a.createRetryStrategy(&config)

	// Execute with retry
	var lastError error
	var attempts []RetryAttempt

	for attempt := 0; attempt <= config.MaxAttempts; attempt++ {
		// Execute the node
		output, err := a.executeNodeFunc(ctx, nodeID)

		// Record attempt
		attemptData := RetryAttempt{
			Attempt:   attempt,
			Timestamp: time.Now(),
			Success:   err == nil,
		}

		if err != nil {
			attemptData.Error = err.Error()
			attemptData.ErrorType = fmt.Sprintf("%T", err)

			// Classify error
			if execErr, ok := err.(interface{ IsRetryable() bool }); ok {
				if execErr.IsRetryable() {
					attemptData.Classification = "transient"
				} else {
					attemptData.Classification = "permanent"
				}
			} else {
				attemptData.Classification = "unknown"
			}
		}

		// Success
		if err == nil {
			attempts = append(attempts, attemptData)
			return NewActionOutput(map[string]interface{}{
				"output":   output,
				"attempts": attempts,
				"success":  true,
				"retries":  attempt,
			}), nil
		}

		lastError = err

		// Check if we should retry this error
		if !a.shouldRetryError(err, &config) {
			_ = append(attempts, attemptData) // Track attempt for potential future logging
			return nil, fmt.Errorf("non-retryable error: %w", err)
		}

		// Check if we have more attempts
		if attempt >= config.MaxAttempts {
			_ = append(attempts, attemptData) // Track attempt for potential future logging
			break
		}

		// Calculate delay
		delay := strategy.CalculateDelay(attempt)
		attemptData.DelayMs = int(delay.Milliseconds())

		// Append attempt with delay info
		attempts = append(attempts, attemptData)

		// Wait for delay or context cancellation
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(delay):
			// Continue to next retry
		}
	}

	// All retries exhausted
	return nil, fmt.Errorf("max retries exceeded (%d attempts): %w", config.MaxAttempts+1, lastError)
}

// validateConfig validates retry configuration
func (a *RetryAction) validateConfig(config *workflow.RetryNodeConfig) error {
	if config.MaxAttempts < 0 {
		return fmt.Errorf("max_attempts must be non-negative")
	}
	if config.InitialDelayMs < 0 {
		return fmt.Errorf("initial_delay_ms must be non-negative")
	}
	if config.MaxDelayMs > 0 && config.MaxDelayMs < config.InitialDelayMs {
		return fmt.Errorf("max_delay_ms must be greater than or equal to initial_delay_ms")
	}

	// Validate strategy
	validStrategies := map[string]bool{
		"fixed":              true,
		"exponential":        true,
		"exponential_jitter": true,
	}
	if !validStrategies[config.Strategy] {
		return fmt.Errorf("invalid strategy: %s (must be fixed, exponential, or exponential_jitter)", config.Strategy)
	}

	// Set defaults
	if config.Strategy == "exponential" || config.Strategy == "exponential_jitter" {
		if config.Multiplier == 0 {
			config.Multiplier = 2.0
		}
		if config.MaxDelayMs == 0 {
			config.MaxDelayMs = 60000 // 60 seconds
		}
	}

	return nil
}

// shouldRetryError determines if an error should be retried
func (a *RetryAction) shouldRetryError(err error, config *workflow.RetryNodeConfig) bool {
	errorStr := err.Error()
	errorType := fmt.Sprintf("%T", err)

	// Check non-retryable errors first
	for _, pattern := range config.NonRetryableErrors {
		if a.matchesPattern(errorStr, pattern) || a.matchesPattern(errorType, pattern) {
			return false
		}
	}

	// If retryable errors are specified, check if error matches
	if len(config.RetryableErrors) > 0 {
		for _, pattern := range config.RetryableErrors {
			if a.matchesPattern(errorStr, pattern) || a.matchesPattern(errorType, pattern) {
				return true
			}
		}
		return false // Doesn't match any retryable patterns
	}

	// Check if error is transient by default
	if execErr, ok := err.(interface{ IsRetryable() bool }); ok {
		return execErr.IsRetryable()
	}

	// Default: retry transient-looking errors
	transientPatterns := []string{
		"timeout",
		"temporary",
		"connection",
		"unavailable",
		"throttle",
		"rate limit",
	}

	errorLower := strings.ToLower(errorStr)
	for _, pattern := range transientPatterns {
		if strings.Contains(errorLower, pattern) {
			return true
		}
	}

	return false
}

// matchesPattern checks if a string matches a pattern (regex or simple contains)
func (a *RetryAction) matchesPattern(str, pattern string) bool {
	// Try as regex first
	if re, err := regexp.Compile(pattern); err == nil {
		return re.MatchString(str)
	}

	// Fallback to simple contains
	return strings.Contains(strings.ToLower(str), strings.ToLower(pattern))
}

// RetryStrategy defines the interface for retry strategies
type RetryStrategy interface {
	CalculateDelay(attempt int) time.Duration
}

// createRetryStrategy creates a retry strategy based on configuration
func (a *RetryAction) createRetryStrategy(config *workflow.RetryNodeConfig) RetryStrategy {
	switch config.Strategy {
	case "fixed":
		return &FixedDelayStrategy{
			Delay: time.Duration(config.InitialDelayMs) * time.Millisecond,
		}
	case "exponential":
		return &ExponentialBackoffStrategy{
			InitialDelay: time.Duration(config.InitialDelayMs) * time.Millisecond,
			MaxDelay:     time.Duration(config.MaxDelayMs) * time.Millisecond,
			Multiplier:   config.Multiplier,
			Jitter:       false,
		}
	case "exponential_jitter":
		return &ExponentialBackoffStrategy{
			InitialDelay: time.Duration(config.InitialDelayMs) * time.Millisecond,
			MaxDelay:     time.Duration(config.MaxDelayMs) * time.Millisecond,
			Multiplier:   config.Multiplier,
			Jitter:       true,
		}
	default:
		// Default to exponential backoff with jitter
		return &ExponentialBackoffStrategy{
			InitialDelay: time.Duration(config.InitialDelayMs) * time.Millisecond,
			MaxDelay:     time.Duration(config.MaxDelayMs) * time.Millisecond,
			Multiplier:   2.0,
			Jitter:       true,
		}
	}
}

// FixedDelayStrategy implements fixed delay retry
type FixedDelayStrategy struct {
	Delay time.Duration
}

// CalculateDelay returns a fixed delay
func (s *FixedDelayStrategy) CalculateDelay(attempt int) time.Duration {
	return s.Delay
}

// ExponentialBackoffStrategy implements exponential backoff retry
type ExponentialBackoffStrategy struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       bool
}

// secureRandSource is a cryptographically secure random source for jitter
var secureRandSource *mathRand.Rand

func init() {
	var seed int64
	if err := binary.Read(rand.Reader, binary.BigEndian, &seed); err != nil {
		seed = time.Now().UnixNano()
	}
	// #nosec G404 -- Using math/rand seeded with crypto/rand for jitter; cryptographic security not required
	secureRandSource = mathRand.New(mathRand.NewSource(seed))
}

// CalculateDelay calculates exponential backoff delay
func (s *ExponentialBackoffStrategy) CalculateDelay(attempt int) time.Duration {
	// Calculate exponential backoff
	delay := float64(s.InitialDelay) * math.Pow(s.Multiplier, float64(attempt))

	// Apply max delay
	if s.MaxDelay > 0 && delay > float64(s.MaxDelay) {
		delay = float64(s.MaxDelay)
	}

	duration := time.Duration(delay)

	// Add jitter if enabled (±25%)
	if s.Jitter {
		jitter := float64(duration) * 0.25
		variation := (secureRandSource.Float64() * 2 * jitter) - jitter
		duration = time.Duration(float64(duration) + variation)
	}

	return duration
}

// RetryAttempt represents a single retry attempt
type RetryAttempt struct {
	Attempt        int       `json:"attempt"`
	Timestamp      time.Time `json:"timestamp"`
	Success        bool      `json:"success"`
	Error          string    `json:"error,omitempty"`
	ErrorType      string    `json:"error_type,omitempty"`
	Classification string    `json:"classification,omitempty"`
	DelayMs        int       `json:"delay_ms,omitempty"`
}
