package security

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AccountLockoutConfig holds configuration for account lockout
type AccountLockoutConfig struct {
	// MaxFailedAttempts is the number of failed attempts before lockout (default: 5)
	MaxFailedAttempts int
	// LockoutDuration is how long the account is locked (default: 15 minutes)
	LockoutDuration time.Duration
	// AttemptWindow is the time window for counting failed attempts (default: 15 minutes)
	AttemptWindow time.Duration
	// ResetOnSuccess resets the failed attempt count on successful login (default: true)
	ResetOnSuccess bool
	// ProgressiveLockout increases lockout duration with each subsequent lockout (default: true)
	ProgressiveLockout bool
	// MaxLockoutDuration is the maximum lockout duration for progressive lockout (default: 24 hours)
	MaxLockoutDuration time.Duration
}

// DefaultAccountLockoutConfig returns secure default configuration
func DefaultAccountLockoutConfig() AccountLockoutConfig {
	return AccountLockoutConfig{
		MaxFailedAttempts:  5,
		LockoutDuration:    15 * time.Minute,
		AttemptWindow:      15 * time.Minute,
		ResetOnSuccess:     true,
		ProgressiveLockout: true,
		MaxLockoutDuration: 24 * time.Hour,
	}
}

// AccountState represents the state of an account's login attempts
type AccountState struct {
	FailedAttempts int
	LastAttempt    time.Time
	LockedUntil    time.Time
	LockoutCount   int // Number of times account has been locked
}

// IsLocked returns true if the account is currently locked
func (s *AccountState) IsLocked() bool {
	return time.Now().Before(s.LockedUntil)
}

// RemainingLockoutTime returns the remaining lockout time
func (s *AccountState) RemainingLockoutTime() time.Duration {
	if !s.IsLocked() {
		return 0
	}
	return time.Until(s.LockedUntil)
}

// AccountLockoutService manages account lockout state
type AccountLockoutService struct {
	config AccountLockoutConfig
	states map[string]*AccountState
	mu     sync.RWMutex
}

// NewAccountLockoutService creates a new account lockout service
func NewAccountLockoutService(config AccountLockoutConfig) *AccountLockoutService {
	service := &AccountLockoutService{
		config: config,
		states: make(map[string]*AccountState),
	}

	// Start cleanup goroutine
	go service.cleanupExpiredStates()

	return service
}

// CheckLockout checks if an account is locked
// Returns nil if not locked, error with remaining time if locked
func (s *AccountLockoutService) CheckLockout(_ context.Context, accountID string) error {
	s.mu.RLock()
	state, exists := s.states[accountID]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	if state.IsLocked() {
		return &LockoutError{
			AccountID:     accountID,
			RemainingTime: state.RemainingLockoutTime(),
			LockoutCount:  state.LockoutCount,
		}
	}

	return nil
}

// RecordFailedAttempt records a failed login attempt
// Returns error if account is now locked
func (s *AccountLockoutService) RecordFailedAttempt(_ context.Context, accountID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, exists := s.states[accountID]
	if !exists {
		state = &AccountState{}
		s.states[accountID] = state
	}

	now := time.Now()

	// Check if we should reset the counter (outside attempt window)
	if now.Sub(state.LastAttempt) > s.config.AttemptWindow {
		state.FailedAttempts = 0
	}

	state.FailedAttempts++
	state.LastAttempt = now

	// Check if we should lock the account
	if state.FailedAttempts >= s.config.MaxFailedAttempts {
		state.LockoutCount++
		lockoutDuration := s.calculateLockoutDuration(state.LockoutCount)
		state.LockedUntil = now.Add(lockoutDuration)
		state.FailedAttempts = 0 // Reset for next window

		return &LockoutError{
			AccountID:     accountID,
			RemainingTime: lockoutDuration,
			LockoutCount:  state.LockoutCount,
		}
	}

	return nil
}

// RecordSuccessfulAttempt records a successful login attempt
func (s *AccountLockoutService) RecordSuccessfulAttempt(_ context.Context, accountID string) {
	if !s.config.ResetOnSuccess {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove the state completely on successful login
	delete(s.states, accountID)
}

// GetAccountState returns the current state of an account
func (s *AccountLockoutService) GetAccountState(accountID string) (*AccountState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, exists := s.states[accountID]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent mutation
	stateCopy := *state
	return &stateCopy, true
}

// UnlockAccount manually unlocks an account (admin action)
func (s *AccountLockoutService) UnlockAccount(accountID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if state, exists := s.states[accountID]; exists {
		state.LockedUntil = time.Time{}
		state.FailedAttempts = 0
		// Note: We don't reset LockoutCount so progressive lockout still applies
	}
}

// ResetAccount completely resets an account's state (admin action)
func (s *AccountLockoutService) ResetAccount(accountID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.states, accountID)
}

func (s *AccountLockoutService) calculateLockoutDuration(lockoutCount int) time.Duration {
	if !s.config.ProgressiveLockout {
		return s.config.LockoutDuration
	}

	// Double the lockout duration for each subsequent lockout
	duration := s.config.LockoutDuration * time.Duration(1<<(lockoutCount-1))

	// Cap at maximum
	if duration > s.config.MaxLockoutDuration {
		return s.config.MaxLockoutDuration
	}

	return duration
}

func (s *AccountLockoutService) cleanupExpiredStates() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for accountID, state := range s.states {
			// Remove if not locked and no recent attempts
			if !state.IsLocked() && now.Sub(state.LastAttempt) > s.config.AttemptWindow*2 {
				delete(s.states, accountID)
			}
		}
		s.mu.Unlock()
	}
}

// LockoutError represents an account lockout error
type LockoutError struct {
	AccountID     string
	RemainingTime time.Duration
	LockoutCount  int
}

func (e *LockoutError) Error() string {
	return fmt.Sprintf("account %s is locked for %v", e.AccountID, e.RemainingTime.Round(time.Second))
}

// IsLockoutError checks if an error is a lockout error
func IsLockoutError(err error) bool {
	_, ok := err.(*LockoutError)
	return ok
}

// --- Rate Limiting by IP ---

// IPRateLimiter provides IP-based rate limiting for authentication endpoints
type IPRateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// NewIPRateLimiter creates a new IP rate limiter
func NewIPRateLimiter(limit int, window time.Duration) *IPRateLimiter {
	limiter := &IPRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// Allow checks if a request from the given IP should be allowed
func (l *IPRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	// Get existing requests and filter old ones
	requests := l.requests[ip]
	var valid []time.Time
	for _, t := range requests {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	// Check if limit exceeded
	if len(valid) >= l.limit {
		l.requests[ip] = valid
		return false
	}

	// Add new request
	valid = append(valid, now)
	l.requests[ip] = valid
	return true
}

// RemainingRequests returns the number of remaining requests for an IP
func (l *IPRateLimiter) RemainingRequests(ip string) int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	requests := l.requests[ip]
	var count int
	for _, t := range requests {
		if t.After(cutoff) {
			count++
		}
	}

	remaining := l.limit - count
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (l *IPRateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-l.window)

		for ip, requests := range l.requests {
			var valid []time.Time
			for _, t := range requests {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(l.requests, ip)
			} else {
				l.requests[ip] = valid
			}
		}
		l.mu.Unlock()
	}
}
