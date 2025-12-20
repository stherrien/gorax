package retention

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockService is a mock implementation of the Service
type MockService struct {
	mock.Mock
}

func (m *MockService) CleanupAllTenants(ctx context.Context) (*CleanupResult, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CleanupResult), args.Error(1)
}

func (m *MockService) CleanupOldExecutions(ctx context.Context, tenantID string) (*CleanupResult, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CleanupResult), args.Error(1)
}

func (m *MockService) GetRetentionPolicy(ctx context.Context, tenantID string) (*RetentionPolicy, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RetentionPolicy), args.Error(1)
}

func TestNewScheduler(t *testing.T) {
	service := new(MockService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	scheduler := NewScheduler(service, logger, 24*time.Hour)

	assert.NotNil(t, scheduler)
	assert.Equal(t, 24*time.Hour, scheduler.interval)
	assert.False(t, scheduler.IsRunning())
}

func TestScheduler_StartAndStop(t *testing.T) {
	service := new(MockService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Mock the initial cleanup call that happens on start
	result := &CleanupResult{ExecutionsDeleted: 0}
	service.On("CleanupAllTenants", mock.Anything).Return(result, nil)

	scheduler := NewScheduler(service, logger, 1*time.Hour)

	// Start scheduler
	err := scheduler.Start(context.Background())
	assert.NoError(t, err)
	assert.True(t, scheduler.IsRunning())

	// Starting again should be no-op
	err = scheduler.Start(context.Background())
	assert.NoError(t, err)
	assert.True(t, scheduler.IsRunning())

	// Stop scheduler
	scheduler.Stop()
	assert.False(t, scheduler.IsRunning())

	// Stopping again should be no-op
	scheduler.Stop()
	assert.False(t, scheduler.IsRunning())
}

func TestScheduler_RunsCleanup(t *testing.T) {
	service := new(MockService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Setup mock to track cleanup calls
	callCount := 0
	var mu sync.Mutex

	result := &CleanupResult{
		ExecutionsDeleted:     100,
		StepExecutionsDeleted: 300,
		BatchesProcessed:      2,
	}

	service.On("CleanupAllTenants", mock.Anything).Run(func(args mock.Arguments) {
		mu.Lock()
		callCount++
		mu.Unlock()
	}).Return(result, nil)

	// Use short interval for testing
	scheduler := NewScheduler(service, logger, 100*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()

	err := scheduler.Start(ctx)
	assert.NoError(t, err)

	// Wait for at least 2 cleanup cycles
	time.Sleep(350 * time.Millisecond)

	scheduler.Stop()

	mu.Lock()
	calls := callCount
	mu.Unlock()

	// Should have run at least 2 times (initial + 2 intervals)
	assert.GreaterOrEqual(t, calls, 2, "cleanup should run multiple times")
}

func TestScheduler_HandlesErrors(t *testing.T) {
	service := new(MockService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Mock service to return error
	service.On("CleanupAllTenants", mock.Anything).Return(nil, assert.AnError)

	scheduler := NewScheduler(service, logger, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := scheduler.Start(ctx)
	assert.NoError(t, err)

	// Wait for cleanup attempts
	time.Sleep(200 * time.Millisecond)

	scheduler.Stop()

	// Scheduler should continue running despite errors
	service.AssertCalled(t, "CleanupAllTenants", mock.Anything)
}

func TestScheduler_ContextCancellation(t *testing.T) {
	service := new(MockService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	result := &CleanupResult{ExecutionsDeleted: 10}
	service.On("CleanupAllTenants", mock.Anything).Return(result, nil)

	scheduler := NewScheduler(service, logger, 1*time.Second)

	ctx, cancel := context.WithCancel(context.Background())

	err := scheduler.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, scheduler.IsRunning())

	// Cancel context
	cancel()

	// Wait for scheduler to stop
	scheduler.Wait()

	assert.False(t, scheduler.IsRunning())
}

func TestScheduler_SetInterval(t *testing.T) {
	service := new(MockService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	scheduler := NewScheduler(service, logger, 1*time.Hour)
	assert.Equal(t, 1*time.Hour, scheduler.interval)

	scheduler.SetInterval(30 * time.Minute)
	assert.Equal(t, 30*time.Minute, scheduler.interval)
}

func TestScheduler_ImmediateExecution(t *testing.T) {
	service := new(MockService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	executed := false
	var mu sync.Mutex

	result := &CleanupResult{ExecutionsDeleted: 10}
	service.On("CleanupAllTenants", mock.Anything).Run(func(args mock.Arguments) {
		mu.Lock()
		executed = true
		mu.Unlock()
	}).Return(result, nil)

	// Use long interval - should still execute immediately on start
	scheduler := NewScheduler(service, logger, 10*time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := scheduler.Start(ctx)
	assert.NoError(t, err)

	// Wait a bit for initial execution
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	wasExecuted := executed
	mu.Unlock()

	scheduler.Stop()

	assert.True(t, wasExecuted, "cleanup should execute immediately on start")
}

func TestScheduler_RunOnce(t *testing.T) {
	service := new(MockService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	result := &CleanupResult{
		ExecutionsDeleted:     50,
		StepExecutionsDeleted: 150,
		BatchesProcessed:      1,
	}
	service.On("CleanupAllTenants", mock.Anything).Return(result, nil)

	scheduler := NewScheduler(service, logger, 1*time.Hour)

	// Test manual one-time execution
	cleanupResult, err := scheduler.RunOnce(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 50, cleanupResult.ExecutionsDeleted)
	assert.Equal(t, 150, cleanupResult.StepExecutionsDeleted)

	service.AssertCalled(t, "CleanupAllTenants", mock.Anything)
}

func TestScheduler_ConcurrentStopCalls(t *testing.T) {
	service := new(MockService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	result := &CleanupResult{ExecutionsDeleted: 10}
	service.On("CleanupAllTenants", mock.Anything).Return(result, nil)

	scheduler := NewScheduler(service, logger, 1*time.Second)

	err := scheduler.Start(context.Background())
	assert.NoError(t, err)

	// Call Stop concurrently multiple times
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scheduler.Stop()
		}()
	}

	wg.Wait()
	assert.False(t, scheduler.IsRunning())
}
