package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrInvalidLimit is returned when limit is <= 0
	ErrInvalidLimit = errors.New("limit must be greater than 0")
	// ErrInvalidWindow is returned when window is <= 0
	ErrInvalidWindow = errors.New("window must be greater than 0")
	// ErrInvalidTenantID is returned when tenant ID is empty
	ErrInvalidTenantID = errors.New("tenant ID cannot be empty")
)

// SlidingWindowLimiter implements rate limiting using Redis sorted sets
// with sliding window algorithm for accurate rate limiting
type SlidingWindowLimiter struct {
	client *redis.Client
}

// NewSlidingWindowLimiter creates a new sliding window rate limiter
func NewSlidingWindowLimiter(client *redis.Client) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		client: client,
	}
}

// Allow checks if a request is allowed under the rate limit
// Returns true if allowed, false if rate limit exceeded
func (l *SlidingWindowLimiter) Allow(ctx context.Context, tenantID string, limit int64, window time.Duration) (bool, error) {
	if err := l.validate(tenantID, limit, window); err != nil {
		return false, err
	}

	key := l.key(tenantID, window)
	now := time.Now().UnixNano()
	windowStart := now - window.Nanoseconds()

	// Use Lua script for atomic operations
	script := redis.NewScript(`
		-- Remove expired entries
		redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', ARGV[1])

		-- Count current entries
		local count = redis.call('ZCARD', KEYS[1])

		-- Check if under limit
		if tonumber(count) < tonumber(ARGV[3]) then
			-- Add new entry
			redis.call('ZADD', KEYS[1], ARGV[2], ARGV[2])
			-- Set expiration
			redis.call('EXPIRE', KEYS[1], ARGV[4])
			return 1
		else
			return 0
		end
	`)

	result, err := script.Run(ctx, l.client, []string{key},
		windowStart,             // ARGV[1] - window start
		now,                     // ARGV[2] - current timestamp
		limit,                   // ARGV[3] - limit
		int(window.Seconds())+1, // ARGV[4] - TTL
	).Result()

	if err != nil {
		return false, fmt.Errorf("rate limit check failed: %w", err)
	}

	// Result is 1 if allowed, 0 if blocked
	resultInt, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected result type from rate limit script")
	}
	return resultInt == 1, nil
}

// GetUsage returns the current usage count within the window
func (l *SlidingWindowLimiter) GetUsage(ctx context.Context, tenantID string, window time.Duration) (int64, error) {
	if tenantID == "" {
		return 0, ErrInvalidTenantID
	}
	if window <= 0 {
		return 0, ErrInvalidWindow
	}

	key := l.key(tenantID, window)
	now := time.Now().UnixNano()
	windowStart := now - window.Nanoseconds()

	// Count entries in current window
	count, err := l.client.ZCount(ctx, key, fmt.Sprint(windowStart), "+inf").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get usage: %w", err)
	}

	return count, nil
}

// Reset clears all rate limit data for a tenant
func (l *SlidingWindowLimiter) Reset(ctx context.Context, tenantID string) error {
	if tenantID == "" {
		return ErrInvalidTenantID
	}

	// Delete all keys for this tenant
	pattern := fmt.Sprintf("ratelimit:%s:*", tenantID)
	iter := l.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		if err := l.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete key: %w", err)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	return nil
}

// validate checks if parameters are valid
func (l *SlidingWindowLimiter) validate(tenantID string, limit int64, window time.Duration) error {
	if tenantID == "" {
		return ErrInvalidTenantID
	}
	if limit <= 0 {
		return ErrInvalidLimit
	}
	if window <= 0 {
		return ErrInvalidWindow
	}
	return nil
}

// key generates a Redis key for rate limiting
func (l *SlidingWindowLimiter) key(tenantID string, window time.Duration) string {
	return fmt.Sprintf("ratelimit:%s:window_%d", tenantID, int64(window.Seconds()))
}
