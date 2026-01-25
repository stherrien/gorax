package security

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountLockoutService_RecordFailedAttempt(t *testing.T) {
	config := AccountLockoutConfig{
		MaxFailedAttempts:  3,
		LockoutDuration:    time.Minute,
		AttemptWindow:      time.Minute,
		ResetOnSuccess:     true,
		ProgressiveLockout: false,
	}
	service := NewAccountLockoutService(config)
	ctx := context.Background()
	accountID := "test-user-1"

	// First two attempts should not lock
	err := service.RecordFailedAttempt(ctx, accountID)
	assert.NoError(t, err)

	err = service.RecordFailedAttempt(ctx, accountID)
	assert.NoError(t, err)

	// Third attempt should lock
	err = service.RecordFailedAttempt(ctx, accountID)
	require.Error(t, err)
	assert.True(t, IsLockoutError(err))

	lockoutErr, ok := err.(*LockoutError)
	require.True(t, ok)
	assert.Equal(t, accountID, lockoutErr.AccountID)
	assert.Equal(t, 1, lockoutErr.LockoutCount)
}

func TestAccountLockoutService_CheckLockout(t *testing.T) {
	config := AccountLockoutConfig{
		MaxFailedAttempts: 2,
		LockoutDuration:   time.Second * 2,
		AttemptWindow:     time.Minute,
	}
	service := NewAccountLockoutService(config)
	ctx := context.Background()
	accountID := "test-user-2"

	// Should not be locked initially
	err := service.CheckLockout(ctx, accountID)
	assert.NoError(t, err)

	// Lock the account
	_ = service.RecordFailedAttempt(ctx, accountID)
	_ = service.RecordFailedAttempt(ctx, accountID)

	// Should be locked now
	err = service.CheckLockout(ctx, accountID)
	require.Error(t, err)
	assert.True(t, IsLockoutError(err))

	// Wait for lockout to expire
	time.Sleep(time.Second * 3)

	// Should not be locked anymore
	err = service.CheckLockout(ctx, accountID)
	assert.NoError(t, err)
}

func TestAccountLockoutService_RecordSuccessfulAttempt(t *testing.T) {
	config := AccountLockoutConfig{
		MaxFailedAttempts: 3,
		LockoutDuration:   time.Minute,
		AttemptWindow:     time.Minute,
		ResetOnSuccess:    true,
	}
	service := NewAccountLockoutService(config)
	ctx := context.Background()
	accountID := "test-user-3"

	// Record two failed attempts
	_ = service.RecordFailedAttempt(ctx, accountID)
	_ = service.RecordFailedAttempt(ctx, accountID)

	// Record successful attempt
	service.RecordSuccessfulAttempt(ctx, accountID)

	// State should be reset
	state, exists := service.GetAccountState(accountID)
	assert.False(t, exists)
	assert.Nil(t, state)
}

func TestAccountLockoutService_ProgressiveLockout(t *testing.T) {
	config := AccountLockoutConfig{
		MaxFailedAttempts:  2,
		LockoutDuration:    time.Second,
		AttemptWindow:      time.Minute,
		ProgressiveLockout: true,
		MaxLockoutDuration: time.Hour,
	}
	service := NewAccountLockoutService(config)
	ctx := context.Background()
	accountID := "test-user-4"

	// First lockout
	_ = service.RecordFailedAttempt(ctx, accountID)
	err := service.RecordFailedAttempt(ctx, accountID)
	require.Error(t, err)

	lockoutErr := err.(*LockoutError)
	firstDuration := lockoutErr.RemainingTime

	// Wait for lockout to expire
	time.Sleep(time.Second * 2)

	// Second lockout should be longer
	_ = service.RecordFailedAttempt(ctx, accountID)
	err = service.RecordFailedAttempt(ctx, accountID)
	require.Error(t, err)

	lockoutErr = err.(*LockoutError)
	secondDuration := lockoutErr.RemainingTime

	// Second lockout should be longer (2x first)
	assert.Greater(t, secondDuration, firstDuration)
}

func TestAccountLockoutService_UnlockAccount(t *testing.T) {
	config := AccountLockoutConfig{
		MaxFailedAttempts: 2,
		LockoutDuration:   time.Hour,
		AttemptWindow:     time.Minute,
	}
	service := NewAccountLockoutService(config)
	ctx := context.Background()
	accountID := "test-user-5"

	// Lock the account
	_ = service.RecordFailedAttempt(ctx, accountID)
	_ = service.RecordFailedAttempt(ctx, accountID)

	// Verify locked
	err := service.CheckLockout(ctx, accountID)
	require.Error(t, err)

	// Admin unlock
	service.UnlockAccount(accountID)

	// Should not be locked anymore
	err = service.CheckLockout(ctx, accountID)
	assert.NoError(t, err)
}

func TestAccountLockoutService_AttemptWindowReset(t *testing.T) {
	config := AccountLockoutConfig{
		MaxFailedAttempts: 2,
		LockoutDuration:   time.Minute,
		AttemptWindow:     time.Millisecond * 100,
	}
	service := NewAccountLockoutService(config)
	ctx := context.Background()
	accountID := "test-user-6"

	// Record one failed attempt
	_ = service.RecordFailedAttempt(ctx, accountID)

	// Wait for window to expire
	time.Sleep(time.Millisecond * 200)

	// Next attempt should not cause lockout (counter reset)
	err := service.RecordFailedAttempt(ctx, accountID)
	assert.NoError(t, err) // Only 1 attempt in new window
}

func TestIPRateLimiter_Allow(t *testing.T) {
	limiter := NewIPRateLimiter(3, time.Second*2)

	ip := "192.168.1.1"

	// First 3 requests should be allowed
	assert.True(t, limiter.Allow(ip))
	assert.True(t, limiter.Allow(ip))
	assert.True(t, limiter.Allow(ip))

	// Fourth should be denied
	assert.False(t, limiter.Allow(ip))

	// Different IP should be allowed
	assert.True(t, limiter.Allow("192.168.1.2"))
}

func TestIPRateLimiter_WindowExpiry(t *testing.T) {
	limiter := NewIPRateLimiter(2, time.Millisecond*100)

	ip := "192.168.1.3"

	// Use up quota
	assert.True(t, limiter.Allow(ip))
	assert.True(t, limiter.Allow(ip))
	assert.False(t, limiter.Allow(ip))

	// Wait for window to expire
	time.Sleep(time.Millisecond * 150)

	// Should be allowed again
	assert.True(t, limiter.Allow(ip))
}

func TestIPRateLimiter_RemainingRequests(t *testing.T) {
	limiter := NewIPRateLimiter(5, time.Second)

	ip := "192.168.1.4"

	// Check initial remaining
	assert.Equal(t, 5, limiter.RemainingRequests(ip))

	// Make some requests
	limiter.Allow(ip)
	limiter.Allow(ip)

	// Check remaining
	assert.Equal(t, 3, limiter.RemainingRequests(ip))
}
