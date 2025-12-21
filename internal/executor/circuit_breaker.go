package executor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

var (
	// ErrCircuitOpen is returned when the circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")
	// ErrTooManyRequests is returned when the circuit breaker is half-open and at capacity
	ErrTooManyRequests = errors.New("circuit breaker is half-open: too many requests")
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	// StateClosed allows all requests through
	StateClosed CircuitState = iota
	// StateOpen blocks all requests
	StateOpen
	// StateHalfOpen allows limited requests to test if the service has recovered
	StateHalfOpen
)

// String returns the string representation of the circuit state
func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig holds configuration for circuit breaker behavior
type CircuitBreakerConfig struct {
	// MaxFailures is the number of consecutive failures before opening the circuit
	MaxFailures int
	// Timeout is how long to wait before transitioning from Open to Half-Open
	Timeout time.Duration
	// MaxRequests is the max concurrent requests allowed in Half-Open state
	MaxRequests int
	// FailureThreshold is the ratio of failures to trigger open state (0.0-1.0)
	FailureThreshold float64
	// SlidingWindowSize is the number of recent requests to consider
	SlidingWindowSize int
}

// DefaultCircuitBreakerConfig returns the default circuit breaker configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:       5,
		Timeout:           60 * time.Second,
		MaxRequests:       3,
		FailureThreshold:  0.5,
		SlidingWindowSize: 10,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name   string
	config CircuitBreakerConfig
	logger *slog.Logger

	mu            sync.RWMutex
	state         CircuitState
	failures      int
	lastFailTime  time.Time
	lastStateTime time.Time
	halfOpenReqs  int

	// Sliding window for tracking recent requests
	window       []bool // true = success, false = failure
	windowIndex  int
	windowFilled bool
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, config CircuitBreakerConfig, logger *slog.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		name:          name,
		config:        config,
		logger:        logger,
		state:         StateClosed,
		lastStateTime: time.Now(),
		window:        make([]bool, config.SlidingWindowSize),
	}
}

// Execute runs an operation through the circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, operation func(context.Context) error) error {
	// Check if we can execute
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	// Execute the operation
	err := operation(ctx)

	// Record the result
	cb.afterRequest(err)

	return err
}

// ExecuteWithResult runs an operation that returns a result through the circuit breaker
func (cb *CircuitBreaker) ExecuteWithResult(ctx context.Context, operation func(context.Context) (interface{}, error)) (interface{}, error) {
	// Check if we can execute
	if err := cb.beforeRequest(); err != nil {
		return nil, err
	}

	// Execute the operation
	result, err := operation(ctx)

	// Record the result
	cb.afterRequest(err)

	return result, err
}

// beforeRequest checks if a request can proceed based on circuit state
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	switch cb.state {
	case StateClosed:
		// Allow request
		return nil

	case StateOpen:
		// Check if timeout has elapsed to transition to half-open
		if now.Sub(cb.lastStateTime) >= cb.config.Timeout {
			cb.setState(StateHalfOpen)
			cb.halfOpenReqs = 0
			cb.logger.Info("circuit breaker transitioning to half-open",
				"name", cb.name,
				"timeout", cb.config.Timeout,
			)
			return nil
		}
		// Still open, reject request
		return ErrCircuitOpen

	case StateHalfOpen:
		// Allow limited concurrent requests
		if cb.halfOpenReqs >= cb.config.MaxRequests {
			return ErrTooManyRequests
		}
		cb.halfOpenReqs++
		return nil

	default:
		return fmt.Errorf("unknown circuit state: %v", cb.state)
	}
}

// afterRequest records the result of a request and updates circuit state
func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	success := err == nil

	// Add to sliding window
	cb.window[cb.windowIndex] = success
	cb.windowIndex = (cb.windowIndex + 1) % len(cb.window)
	if cb.windowIndex == 0 {
		cb.windowFilled = true
	}

	if cb.state == StateHalfOpen {
		cb.halfOpenReqs--
	}

	if success {
		cb.onSuccess()
	} else {
		cb.onFailure()
	}
}

// onSuccess handles a successful request
func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		// Reset failure count on success
		if cb.failures > 0 {
			cb.failures = 0
		}

	case StateHalfOpen:
		// Check if we should transition to closed
		successCount, totalCount := cb.getWindowStats()
		if totalCount >= cb.config.SlidingWindowSize {
			failureRatio := 1.0 - (float64(successCount) / float64(totalCount))
			if failureRatio < cb.config.FailureThreshold {
				cb.setState(StateClosed)
				cb.failures = 0
				cb.logger.Info("circuit breaker closed after recovery",
					"name", cb.name,
					"success_count", successCount,
					"total_count", totalCount,
				)
			}
		}
	}
}

// onFailure handles a failed request
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	switch cb.state {
	case StateClosed:
		// Check if we should open the circuit
		if cb.shouldOpen() {
			cb.setState(StateOpen)
			cb.logger.Warn("circuit breaker opened due to failures",
				"name", cb.name,
				"failures", cb.failures,
				"max_failures", cb.config.MaxFailures,
			)
		}

	case StateHalfOpen:
		// Any failure in half-open state reopens the circuit
		cb.setState(StateOpen)
		cb.logger.Warn("circuit breaker reopened after half-open failure",
			"name", cb.name,
		)
	}
}

// shouldOpen determines if the circuit should open based on failure criteria
func (cb *CircuitBreaker) shouldOpen() bool {
	// Check consecutive failures
	if cb.failures >= cb.config.MaxFailures {
		return true
	}

	// Check failure threshold in sliding window
	if cb.windowFilled || cb.windowIndex >= cb.config.SlidingWindowSize {
		successCount, totalCount := cb.getWindowStats()
		if totalCount > 0 {
			failureRatio := 1.0 - (float64(successCount) / float64(totalCount))
			if failureRatio >= cb.config.FailureThreshold {
				return true
			}
		}
	}

	return false
}

// getWindowStats returns the success count and total count from the sliding window
func (cb *CircuitBreaker) getWindowStats() (successCount, totalCount int) {
	limit := len(cb.window)
	if !cb.windowFilled {
		limit = cb.windowIndex
	}

	for i := 0; i < limit; i++ {
		if cb.window[i] {
			successCount++
		}
		totalCount++
	}

	return successCount, totalCount
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(state CircuitState) {
	if cb.state != state {
		oldState := cb.state
		cb.state = state
		cb.lastStateTime = time.Now()
		cb.logger.Info("circuit breaker state changed",
			"name", cb.name,
			"old_state", oldState.String(),
			"new_state", state.String(),
		)
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetFailureCount returns the current failure count
func (cb *CircuitBreaker) GetFailureCount() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.setState(StateClosed)
	cb.failures = 0
	cb.halfOpenReqs = 0
	cb.windowIndex = 0
	cb.windowFilled = false
	cb.window = make([]bool, cb.config.SlidingWindowSize)

	cb.logger.Info("circuit breaker reset",
		"name", cb.name,
	)
}

// CircuitBreakerRegistry manages multiple circuit breakers
type CircuitBreakerRegistry struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker
	config   CircuitBreakerConfig
	logger   *slog.Logger
}

// NewCircuitBreakerRegistry creates a new circuit breaker registry
func NewCircuitBreakerRegistry(config CircuitBreakerConfig, logger *slog.Logger) *CircuitBreakerRegistry {
	return &CircuitBreakerRegistry{
		breakers: make(map[string]*CircuitBreaker),
		config:   config,
		logger:   logger,
	}
}

// GetOrCreate gets an existing circuit breaker or creates a new one
func (r *CircuitBreakerRegistry) GetOrCreate(name string) *CircuitBreaker {
	r.mu.RLock()
	breaker, exists := r.breakers[name]
	r.mu.RUnlock()

	if exists {
		return breaker
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if breaker, exists := r.breakers[name]; exists {
		return breaker
	}

	// Create new circuit breaker
	breaker = NewCircuitBreaker(name, r.config, r.logger)
	r.breakers[name] = breaker

	return breaker
}

// Get retrieves a circuit breaker by name
func (r *CircuitBreakerRegistry) Get(name string) (*CircuitBreaker, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	breaker, exists := r.breakers[name]
	return breaker, exists
}

// Reset resets all circuit breakers
func (r *CircuitBreakerRegistry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, breaker := range r.breakers {
		breaker.Reset()
	}
}

// GetStats returns statistics for all circuit breakers
func (r *CircuitBreakerRegistry) GetStats() map[string]CircuitBreakerStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]CircuitBreakerStats)
	for name, breaker := range r.breakers {
		breaker.mu.RLock()
		stats[name] = CircuitBreakerStats{
			Name:         name,
			State:        breaker.state.String(),
			Failures:     breaker.failures,
			LastFailTime: breaker.lastFailTime,
		}
		breaker.mu.RUnlock()
	}

	return stats
}

// CircuitBreakerStats holds statistics for a circuit breaker
type CircuitBreakerStats struct {
	Name         string    `json:"name"`
	State        string    `json:"state"`
	Failures     int       `json:"failures"`
	LastFailTime time.Time `json:"last_fail_time"`
}
