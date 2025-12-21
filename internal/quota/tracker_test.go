package quota

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr
}

func TestNewTracker(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	tracker := NewTracker(client)
	assert.NotNil(t, tracker)
	assert.Equal(t, client, tracker.client)
}

func TestTracker_IncrementWorkflowExecutions(t *testing.T) {
	tests := []struct {
		name     string
		tenantID string
		period   Period
		wantErr  bool
		errType  error
	}{
		{
			name:     "increment daily executions",
			tenantID: "tenant-1",
			period:   PeriodDaily,
			wantErr:  false,
		},
		{
			name:     "increment monthly executions",
			tenantID: "tenant-1",
			period:   PeriodMonthly,
			wantErr:  false,
		},
		{
			name:     "empty tenant ID",
			tenantID: "",
			period:   PeriodDaily,
			wantErr:  true,
			errType:  ErrInvalidTenantID,
		},
		{
			name:     "invalid period",
			tenantID: "tenant-1",
			period:   Period("invalid"),
			wantErr:  true,
			errType:  ErrInvalidPeriod,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mr := setupTestRedis(t)
			defer mr.Close()

			tracker := NewTracker(client)
			ctx := context.Background()

			err := tracker.IncrementWorkflowExecutions(ctx, tt.tenantID, tt.period)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)

				// Verify the counter was incremented
				count, err := tracker.GetWorkflowExecutions(ctx, tt.tenantID, tt.period)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), count)
			}
		})
	}
}

func TestTracker_IncrementStepExecutions(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	tracker := NewTracker(client)
	ctx := context.Background()

	tenantID := "tenant-1"

	err := tracker.IncrementStepExecutions(ctx, tenantID, PeriodDaily)
	assert.NoError(t, err)

	count, err := tracker.GetStepExecutions(ctx, tenantID, PeriodDaily)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestTracker_GetWorkflowExecutions(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	tracker := NewTracker(client)
	ctx := context.Background()

	tenantID := "tenant-1"

	// Initially should be 0
	count, err := tracker.GetWorkflowExecutions(ctx, tenantID, PeriodDaily)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Increment multiple times
	for i := 0; i < 5; i++ {
		err = tracker.IncrementWorkflowExecutions(ctx, tenantID, PeriodDaily)
		assert.NoError(t, err)
	}

	// Should return 5
	count, err = tracker.GetWorkflowExecutions(ctx, tenantID, PeriodDaily)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestTracker_GetStepExecutions(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	tracker := NewTracker(client)
	ctx := context.Background()

	tenantID := "tenant-1"

	// Increment step executions
	for i := 0; i < 10; i++ {
		err := tracker.IncrementStepExecutions(ctx, tenantID, PeriodMonthly)
		assert.NoError(t, err)
	}

	count, err := tracker.GetStepExecutions(ctx, tenantID, PeriodMonthly)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), count)
}

func TestTracker_GetUsageByDateRange(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	tracker := NewTracker(client)
	ctx := context.Background()

	tenantID := "tenant-1"

	// Set up data for multiple days
	now := time.Now()
	dates := []time.Time{
		now.AddDate(0, 0, -2), // 2 days ago
		now.AddDate(0, 0, -1), // yesterday
		now,                   // today
	}

	// Manually set counters for different dates
	for i, date := range dates {
		key := tracker.dailyKey(tenantID, date)
		err := client.Set(ctx, key, i+1, 48*time.Hour).Err()
		require.NoError(t, err)
	}

	startDate := now.AddDate(0, 0, -2)
	endDate := now

	usage, err := tracker.GetUsageByDateRange(ctx, tenantID, startDate, endDate)
	assert.NoError(t, err)
	assert.Len(t, usage, 3)

	// Verify dates and counts
	assert.Equal(t, startDate.Format("2006-01-02"), usage[0].Date)
	assert.Equal(t, int64(1), usage[0].WorkflowExecutions)

	assert.Equal(t, dates[1].Format("2006-01-02"), usage[1].Date)
	assert.Equal(t, int64(2), usage[1].WorkflowExecutions)

	assert.Equal(t, endDate.Format("2006-01-02"), usage[2].Date)
	assert.Equal(t, int64(3), usage[2].WorkflowExecutions)
}

func TestTracker_GetUsageByDateRange_EmptyRange(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	tracker := NewTracker(client)
	ctx := context.Background()

	tenantID := "tenant-1"
	now := time.Now()

	usage, err := tracker.GetUsageByDateRange(ctx, tenantID, now, now.AddDate(0, 0, -1))
	assert.Error(t, err)
	assert.Nil(t, usage)
	assert.ErrorIs(t, err, ErrInvalidDateRange)
}

func TestTracker_CheckQuota(t *testing.T) {
	tests := []struct {
		name       string
		tenantID   string
		period     Period
		quota      int64
		current    int64
		wantExceed bool
	}{
		{
			name:       "within quota",
			tenantID:   "tenant-1",
			period:     PeriodDaily,
			quota:      100,
			current:    50,
			wantExceed: false,
		},
		{
			name:       "at quota limit",
			tenantID:   "tenant-2",
			period:     PeriodDaily,
			quota:      100,
			current:    100,
			wantExceed: true,
		},
		{
			name:       "exceeded quota",
			tenantID:   "tenant-3",
			period:     PeriodDaily,
			quota:      100,
			current:    150,
			wantExceed: true,
		},
		{
			name:       "unlimited quota (-1)",
			tenantID:   "tenant-4",
			period:     PeriodDaily,
			quota:      -1,
			current:    999999,
			wantExceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mr := setupTestRedis(t)
			defer mr.Close()

			tracker := NewTracker(client)
			ctx := context.Background()

			// Set current usage
			key := tracker.periodKey(tt.tenantID, tt.period, "workflow")
			err := client.Set(ctx, key, tt.current, time.Hour).Err()
			require.NoError(t, err)

			exceeded, remaining, err := tracker.CheckQuota(ctx, tt.tenantID, tt.period, tt.quota)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantExceed, exceeded)

			if tt.quota == -1 {
				assert.Equal(t, int64(-1), remaining)
			} else if !tt.wantExceed {
				assert.Equal(t, tt.quota-tt.current, remaining)
			} else {
				assert.Equal(t, int64(0), remaining)
			}
		})
	}
}

func TestTracker_Reset(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	tracker := NewTracker(client)
	ctx := context.Background()

	tenantID := "tenant-1"

	// Create some usage data
	err := tracker.IncrementWorkflowExecutions(ctx, tenantID, PeriodDaily)
	require.NoError(t, err)
	err = tracker.IncrementWorkflowExecutions(ctx, tenantID, PeriodMonthly)
	require.NoError(t, err)
	err = tracker.IncrementStepExecutions(ctx, tenantID, PeriodDaily)
	require.NoError(t, err)

	// Verify data exists
	count, err := tracker.GetWorkflowExecutions(ctx, tenantID, PeriodDaily)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Reset
	err = tracker.Reset(ctx, tenantID)
	assert.NoError(t, err)

	// Verify all data is cleared
	count, err = tracker.GetWorkflowExecutions(ctx, tenantID, PeriodDaily)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	count, err = tracker.GetWorkflowExecutions(ctx, tenantID, PeriodMonthly)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	count, err = tracker.GetStepExecutions(ctx, tenantID, PeriodDaily)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestTracker_IncrementAndDecrement(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	tracker := NewTracker(client)
	ctx := context.Background()

	tenantID := "tenant-1"

	// Increment
	err := tracker.IncrementWorkflowExecutions(ctx, tenantID, PeriodDaily)
	require.NoError(t, err)

	count, err := tracker.GetWorkflowExecutions(ctx, tenantID, PeriodDaily)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Decrement (for cancellation handling)
	err = tracker.DecrementWorkflowExecutions(ctx, tenantID, PeriodDaily)
	assert.NoError(t, err)

	count, err = tracker.GetWorkflowExecutions(ctx, tenantID, PeriodDaily)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestTracker_ConcurrentIncrements(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	tracker := NewTracker(client)
	ctx := context.Background()

	tenantID := "tenant-1"
	goroutines := 100

	// Launch concurrent increments
	done := make(chan bool)
	for i := 0; i < goroutines; i++ {
		go func() {
			err := tracker.IncrementWorkflowExecutions(ctx, tenantID, PeriodDaily)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Verify count is accurate
	count, err := tracker.GetWorkflowExecutions(ctx, tenantID, PeriodDaily)
	assert.NoError(t, err)
	assert.Equal(t, int64(goroutines), count)
}

func TestTracker_KeyExpiration(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	tracker := NewTracker(client)
	ctx := context.Background()

	tenantID := "tenant-1"

	// Increment with daily period
	err := tracker.IncrementWorkflowExecutions(ctx, tenantID, PeriodDaily)
	require.NoError(t, err)

	// Get key
	key := tracker.periodKey(tenantID, PeriodDaily, "workflow")

	// Check TTL exists
	ttl, err := client.TTL(ctx, key).Result()
	assert.NoError(t, err)
	assert.True(t, ttl > 0, "Key should have TTL set")

	// For daily, TTL should be around 48 hours
	assert.True(t, ttl <= 48*time.Hour, "Daily key TTL should be <= 48 hours")
	assert.True(t, ttl > 47*time.Hour, "Daily key TTL should be > 47 hours")
}
