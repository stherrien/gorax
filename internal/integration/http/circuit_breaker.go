package http

import (
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	// StateClosed means the circuit is closed and requests flow normally.
	StateClosed CircuitState = iota
	// StateOpen means the circuit is open and requests are blocked.
	StateOpen
	// StateHalfOpen means the circuit is testing if the service has recovered.
	StateHalfOpen
)

// String returns the string representation of the circuit state.
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

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	mu sync.RWMutex

	// Configuration
	failureThreshold   int           // Number of failures before opening
	successThreshold   int           // Number of successes in half-open to close
	timeout            time.Duration // How long to wait before half-open
	halfOpenMaxAllowed int           // Max concurrent requests in half-open

	// State
	state            CircuitState
	failures         int
	successes        int
	lastFailureTime  time.Time
	halfOpenRequests int

	// Callbacks
	onStateChange func(from, to CircuitState)
}

// CircuitBreakerConfig holds configuration for a circuit breaker.
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of consecutive failures before opening.
	FailureThreshold int

	// SuccessThreshold is the number of successes needed to close from half-open.
	SuccessThreshold int

	// Timeout is how long to wait before transitioning from open to half-open.
	Timeout time.Duration

	// HalfOpenMaxAllowed is the max number of concurrent requests in half-open state.
	HalfOpenMaxAllowed int

	// OnStateChange is called when the circuit state changes.
	OnStateChange func(from, to CircuitState)
}

// DefaultCircuitBreakerConfig returns the default circuit breaker configuration.
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold:   5,
		SuccessThreshold:   2,
		Timeout:            30 * time.Second,
		HalfOpenMaxAllowed: 1,
	}
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration.
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	return &CircuitBreaker{
		failureThreshold:   config.FailureThreshold,
		successThreshold:   config.SuccessThreshold,
		timeout:            config.Timeout,
		halfOpenMaxAllowed: config.HalfOpenMaxAllowed,
		state:              StateClosed,
		onStateChange:      config.OnStateChange,
	}
}

// Allow checks if a request is allowed through the circuit breaker.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		// Check if timeout has elapsed
		if time.Since(cb.lastFailureTime) >= cb.timeout {
			cb.transitionTo(StateHalfOpen)
			cb.halfOpenRequests = 1
			return true
		}
		return false

	case StateHalfOpen:
		// Allow limited requests in half-open state
		if cb.halfOpenRequests < cb.halfOpenMaxAllowed {
			cb.halfOpenRequests++
			return true
		}
		return false

	default:
		return false
	}
}

// RecordSuccess records a successful request.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		cb.failures = 0 // Reset failure count on success

	case StateHalfOpen:
		cb.successes++
		cb.halfOpenRequests--
		if cb.successes >= cb.successThreshold {
			cb.transitionTo(StateClosed)
		}

	case StateOpen:
		// Shouldn't happen if Allow() was called first
	}
}

// RecordFailure records a failed request.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.failureThreshold {
			cb.transitionTo(StateOpen)
		}

	case StateHalfOpen:
		cb.halfOpenRequests--
		cb.transitionTo(StateOpen)

	case StateOpen:
		// Already open, just update last failure time
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.transitionTo(StateClosed)
}

// transitionTo changes the circuit breaker state.
// Must be called with lock held.
func (cb *CircuitBreaker) transitionTo(newState CircuitState) {
	if cb.state == newState {
		return
	}

	oldState := cb.state
	cb.state = newState

	// Reset counters based on new state
	switch newState {
	case StateClosed:
		cb.failures = 0
		cb.successes = 0
		cb.halfOpenRequests = 0

	case StateOpen:
		cb.successes = 0
		cb.halfOpenRequests = 0

	case StateHalfOpen:
		cb.successes = 0
		cb.halfOpenRequests = 0
	}

	// Call state change callback
	if cb.onStateChange != nil {
		cb.onStateChange(oldState, newState)
	}
}

// Stats returns statistics about the circuit breaker.
func (cb *CircuitBreaker) Stats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return CircuitBreakerStats{
		State:           cb.state,
		Failures:        cb.failures,
		Successes:       cb.successes,
		LastFailureTime: cb.lastFailureTime,
	}
}

// CircuitBreakerStats holds statistics about a circuit breaker.
type CircuitBreakerStats struct {
	State           CircuitState
	Failures        int
	Successes       int
	LastFailureTime time.Time
}
