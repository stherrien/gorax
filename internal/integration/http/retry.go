package http

import (
	"math"
	"math/rand/v2"
	"time"
)

// RetryConfig defines the configuration for retry behavior.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts.
	MaxRetries int

	// BaseDelay is the initial delay before the first retry.
	BaseDelay time.Duration

	// MaxDelay is the maximum delay between retries.
	MaxDelay time.Duration

	// Multiplier is the factor by which the delay increases after each retry.
	Multiplier float64

	// JitterFactor is the percentage of jitter to add (0.0 to 1.0).
	JitterFactor float64

	// ShouldRetry is a custom function to determine if a retry should occur.
	ShouldRetry func(err error, resp *Response) bool
}

// DefaultRetryConfig returns the default retry configuration.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:   DefaultMaxRetries,
		BaseDelay:    DefaultBaseDelay,
		MaxDelay:     DefaultMaxDelay,
		Multiplier:   2.0,
		JitterFactor: 0.2,
		ShouldRetry:  nil, // Use default retry logic
	}
}

// CalculateDelay calculates the delay for a given retry attempt.
// Uses exponential backoff with jitter.
func (c *RetryConfig) CalculateDelay(attempt int) time.Duration {
	if c == nil {
		return DefaultBaseDelay
	}

	// Calculate exponential delay
	delay := float64(c.BaseDelay) * math.Pow(c.Multiplier, float64(attempt))

	// Cap at max delay
	if delay > float64(c.MaxDelay) {
		delay = float64(c.MaxDelay)
	}

	// Add jitter
	if c.JitterFactor > 0 {
		jitter := delay * c.JitterFactor * (rand.Float64()*2 - 1) // -jitter to +jitter
		delay += jitter
	}

	// Ensure delay is non-negative
	if delay < 0 {
		delay = 0
	}

	return time.Duration(delay)
}

// RetryConfigBuilder provides a fluent interface for building RetryConfig.
type RetryConfigBuilder struct {
	config *RetryConfig
}

// NewRetryConfigBuilder creates a new RetryConfigBuilder with default values.
func NewRetryConfigBuilder() *RetryConfigBuilder {
	return &RetryConfigBuilder{
		config: DefaultRetryConfig(),
	}
}

// WithMaxRetries sets the maximum number of retries.
func (b *RetryConfigBuilder) WithMaxRetries(maxRetries int) *RetryConfigBuilder {
	b.config.MaxRetries = maxRetries
	return b
}

// WithBaseDelay sets the base delay.
func (b *RetryConfigBuilder) WithBaseDelay(delay time.Duration) *RetryConfigBuilder {
	b.config.BaseDelay = delay
	return b
}

// WithMaxDelay sets the maximum delay.
func (b *RetryConfigBuilder) WithMaxDelay(delay time.Duration) *RetryConfigBuilder {
	b.config.MaxDelay = delay
	return b
}

// WithMultiplier sets the delay multiplier.
func (b *RetryConfigBuilder) WithMultiplier(multiplier float64) *RetryConfigBuilder {
	b.config.Multiplier = multiplier
	return b
}

// WithJitterFactor sets the jitter factor.
func (b *RetryConfigBuilder) WithJitterFactor(factor float64) *RetryConfigBuilder {
	b.config.JitterFactor = factor
	return b
}

// WithShouldRetry sets the custom retry predicate.
func (b *RetryConfigBuilder) WithShouldRetry(fn func(err error, resp *Response) bool) *RetryConfigBuilder {
	b.config.ShouldRetry = fn
	return b
}

// Build returns the configured RetryConfig.
func (b *RetryConfigBuilder) Build() *RetryConfig {
	return b.config
}

// NoRetry returns a retry config that disables retries.
func NoRetry() *RetryConfig {
	return &RetryConfig{
		MaxRetries: 1,
		BaseDelay:  0,
		MaxDelay:   0,
	}
}

// AggressiveRetry returns a retry config with more aggressive retry behavior.
func AggressiveRetry() *RetryConfig {
	return &RetryConfig{
		MaxRetries:   5,
		BaseDelay:    50 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   1.5,
		JitterFactor: 0.3,
	}
}

// ConservativeRetry returns a retry config with more conservative retry behavior.
func ConservativeRetry() *RetryConfig {
	return &RetryConfig{
		MaxRetries:   2,
		BaseDelay:    500 * time.Millisecond,
		MaxDelay:     30 * time.Second,
		Multiplier:   3.0,
		JitterFactor: 0.1,
	}
}
