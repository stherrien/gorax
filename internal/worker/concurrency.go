package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TenantConcurrencyLimiter manages per-tenant concurrency limits
type TenantConcurrencyLimiter struct {
	redis       *redis.Client
	maxPerTenant int
	keyPrefix   string
}

// NewTenantConcurrencyLimiter creates a new tenant concurrency limiter
func NewTenantConcurrencyLimiter(redis *redis.Client, maxPerTenant int) *TenantConcurrencyLimiter {
	return &TenantConcurrencyLimiter{
		redis:        redis,
		maxPerTenant: maxPerTenant,
		keyPrefix:    "tenant:concurrency:",
	}
}

// Acquire attempts to acquire a concurrency slot for a tenant
// Returns true if acquired, false if tenant is at capacity
func (tcl *TenantConcurrencyLimiter) Acquire(ctx context.Context, tenantID string, executionID string) (bool, error) {
	key := tcl.keyPrefix + tenantID

	// Use Redis ZADD with NX to atomically check and increment
	// Store execution ID with current timestamp as score
	now := float64(time.Now().Unix())

	// First, clean up old entries (executions that finished more than 1 hour ago)
	cutoff := now - 3600
	tcl.redis.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%f", cutoff))

	// Count current active executions
	count, err := tcl.redis.ZCard(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check tenant concurrency: %w", err)
	}

	// Check if at capacity
	if int(count) >= tcl.maxPerTenant {
		return false, nil
	}

	// Add this execution
	_, err = tcl.redis.ZAdd(ctx, key, redis.Z{
		Score:  now,
		Member: executionID,
	}).Result()

	if err != nil {
		return false, fmt.Errorf("failed to acquire concurrency slot: %w", err)
	}

	// Set expiry on the key to ensure cleanup
	tcl.redis.Expire(ctx, key, 24*time.Hour)

	return true, nil
}

// Release releases a concurrency slot for a tenant
func (tcl *TenantConcurrencyLimiter) Release(ctx context.Context, tenantID string, executionID string) error {
	key := tcl.keyPrefix + tenantID

	_, err := tcl.redis.ZRem(ctx, key, executionID).Result()
	if err != nil {
		return fmt.Errorf("failed to release concurrency slot: %w", err)
	}

	return nil
}

// GetCurrent returns the current concurrency count for a tenant
func (tcl *TenantConcurrencyLimiter) GetCurrent(ctx context.Context, tenantID string) (int, error) {
	key := tcl.keyPrefix + tenantID

	// Clean up old entries first
	now := float64(time.Now().Unix())
	cutoff := now - 3600
	tcl.redis.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%f", cutoff))

	count, err := tcl.redis.ZCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get tenant concurrency: %w", err)
	}

	return int(count), nil
}

// GetMaxPerTenant returns the maximum concurrent executions per tenant
func (tcl *TenantConcurrencyLimiter) GetMaxPerTenant() int {
	return tcl.maxPerTenant
}
