package quota

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrInvalidTenantID is returned when tenant ID is empty
	ErrInvalidTenantID = errors.New("tenant ID cannot be empty")
	// ErrInvalidPeriod is returned when period is invalid
	ErrInvalidPeriod = errors.New("invalid period")
	// ErrInvalidDateRange is returned when date range is invalid
	ErrInvalidDateRange = errors.New("end date must be after start date")
)

// Period represents a time period for quota tracking
type Period string

const (
	// PeriodDaily represents daily quota tracking
	PeriodDaily Period = "daily"
	// PeriodMonthly represents monthly quota tracking
	PeriodMonthly Period = "monthly"
)

// UsageByDate represents usage statistics for a specific date
type UsageByDate struct {
	Date               string `json:"date"`
	WorkflowExecutions int64  `json:"workflow_executions"`
	StepExecutions     int64  `json:"step_executions"`
}

// Tracker handles quota tracking using Redis
type Tracker struct {
	client *redis.Client
}

// NewTracker creates a new quota tracker
func NewTracker(client *redis.Client) *Tracker {
	return &Tracker{
		client: client,
	}
}

// IncrementWorkflowExecutions increments workflow execution count
func (t *Tracker) IncrementWorkflowExecutions(ctx context.Context, tenantID string, period Period) error {
	if tenantID == "" {
		return ErrInvalidTenantID
	}
	if !isValidPeriod(period) {
		return ErrInvalidPeriod
	}

	key := t.periodKey(tenantID, period, "workflow")
	ttl := t.getTTL(period)

	pipe := t.client.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to increment workflow executions: %w", err)
	}

	return nil
}

// IncrementStepExecutions increments step execution count
func (t *Tracker) IncrementStepExecutions(ctx context.Context, tenantID string, period Period) error {
	if tenantID == "" {
		return ErrInvalidTenantID
	}
	if !isValidPeriod(period) {
		return ErrInvalidPeriod
	}

	key := t.periodKey(tenantID, period, "step")
	ttl := t.getTTL(period)

	pipe := t.client.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to increment step executions: %w", err)
	}

	return nil
}

// DecrementWorkflowExecutions decrements workflow execution count
func (t *Tracker) DecrementWorkflowExecutions(ctx context.Context, tenantID string, period Period) error {
	if tenantID == "" {
		return ErrInvalidTenantID
	}
	if !isValidPeriod(period) {
		return ErrInvalidPeriod
	}

	key := t.periodKey(tenantID, period, "workflow")

	// Don't let it go below 0
	count, err := t.client.Get(ctx, key).Int64()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to get current count: %w", err)
	}

	if count > 0 {
		err = t.client.Decr(ctx, key).Err()
		if err != nil {
			return fmt.Errorf("failed to decrement workflow executions: %w", err)
		}
	}

	return nil
}

// GetWorkflowExecutions returns workflow execution count for a period
func (t *Tracker) GetWorkflowExecutions(ctx context.Context, tenantID string, period Period) (int64, error) {
	if tenantID == "" {
		return 0, ErrInvalidTenantID
	}
	if !isValidPeriod(period) {
		return 0, ErrInvalidPeriod
	}

	key := t.periodKey(tenantID, period, "workflow")
	count, err := t.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get workflow executions: %w", err)
	}

	return count, nil
}

// GetStepExecutions returns step execution count for a period
func (t *Tracker) GetStepExecutions(ctx context.Context, tenantID string, period Period) (int64, error) {
	if tenantID == "" {
		return 0, ErrInvalidTenantID
	}
	if !isValidPeriod(period) {
		return 0, ErrInvalidPeriod
	}

	key := t.periodKey(tenantID, period, "step")
	count, err := t.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get step executions: %w", err)
	}

	return count, nil
}

// CheckQuota checks if quota is exceeded
// Returns: exceeded (bool), remaining (int64), error
func (t *Tracker) CheckQuota(ctx context.Context, tenantID string, period Period, quota int64) (bool, int64, error) {
	if tenantID == "" {
		return false, 0, ErrInvalidTenantID
	}
	if !isValidPeriod(period) {
		return false, 0, ErrInvalidPeriod
	}

	// -1 means unlimited
	if quota == -1 {
		return false, -1, nil
	}

	current, err := t.GetWorkflowExecutions(ctx, tenantID, period)
	if err != nil {
		return false, 0, err
	}

	exceeded := current >= quota
	remaining := quota - current
	if remaining < 0 {
		remaining = 0
	}

	return exceeded, remaining, nil
}

// GetUsageByDateRange returns usage statistics for a date range
func (t *Tracker) GetUsageByDateRange(ctx context.Context, tenantID string, startDate, endDate time.Time) ([]UsageByDate, error) {
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}
	if endDate.Before(startDate) {
		return nil, ErrInvalidDateRange
	}

	var usage []UsageByDate

	// Iterate through each day in the range
	currentDate := startDate
	for !currentDate.After(endDate) {
		workflowKey := t.dailyKey(tenantID, currentDate)
		stepKey := t.dailyKey(tenantID, currentDate) + ":step"

		workflowCount, err := t.client.Get(ctx, workflowKey).Int64()
		if err != nil && err != redis.Nil {
			return nil, fmt.Errorf("failed to get workflow count for %s: %w", currentDate.Format("2006-01-02"), err)
		}
		if err == redis.Nil {
			workflowCount = 0
		}

		stepCount, err := t.client.Get(ctx, stepKey).Int64()
		if err != nil && err != redis.Nil {
			return nil, fmt.Errorf("failed to get step count for %s: %w", currentDate.Format("2006-01-02"), err)
		}
		if err == redis.Nil {
			stepCount = 0
		}

		usage = append(usage, UsageByDate{
			Date:               currentDate.Format("2006-01-02"),
			WorkflowExecutions: workflowCount,
			StepExecutions:     stepCount,
		})

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return usage, nil
}

// Reset clears all quota data for a tenant
func (t *Tracker) Reset(ctx context.Context, tenantID string) error {
	if tenantID == "" {
		return ErrInvalidTenantID
	}

	pattern := fmt.Sprintf("quota:%s:*", tenantID)
	iter := t.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		if err := t.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete key: %w", err)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	return nil
}

// periodKey generates Redis key for period-based tracking
func (t *Tracker) periodKey(tenantID string, period Period, counter string) string {
	now := time.Now()

	switch period {
	case PeriodDaily:
		return fmt.Sprintf("quota:%s:daily:%s:%s", tenantID, now.Format("2006-01-02"), counter)
	case PeriodMonthly:
		return fmt.Sprintf("quota:%s:monthly:%s:%s", tenantID, now.Format("2006-01"), counter)
	default:
		return fmt.Sprintf("quota:%s:%s:%s", tenantID, period, counter)
	}
}

// dailyKey generates Redis key for a specific date
func (t *Tracker) dailyKey(tenantID string, date time.Time) string {
	return fmt.Sprintf("quota:%s:daily:%s:workflow", tenantID, date.Format("2006-01-02"))
}

// getTTL returns appropriate TTL for a period
func (t *Tracker) getTTL(period Period) time.Duration {
	switch period {
	case PeriodDaily:
		return 48 * time.Hour // 2 days to handle timezone differences
	case PeriodMonthly:
		return 62 * 24 * time.Hour // ~2 months
	default:
		return 48 * time.Hour
	}
}

// isValidPeriod checks if period is valid
func isValidPeriod(period Period) bool {
	return period == PeriodDaily || period == PeriodMonthly
}
