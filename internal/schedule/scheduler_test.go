package schedule

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"
)

// MockService for testing scheduler
type MockService struct {
	getDueSchedulesFunc func(ctx context.Context) ([]*Schedule, error)
	markScheduleRunFunc func(ctx context.Context, scheduleID, executionID string) error
	mu                  sync.Mutex
	callCount           int
}

func (m *MockService) GetDueSchedules(ctx context.Context) ([]*Schedule, error) {
	m.mu.Lock()
	m.callCount++
	m.mu.Unlock()
	if m.getDueSchedulesFunc != nil {
		return m.getDueSchedulesFunc(ctx)
	}
	return []*Schedule{}, nil
}

func (m *MockService) MarkScheduleRun(ctx context.Context, scheduleID, executionID string) error {
	if m.markScheduleRunFunc != nil {
		return m.markScheduleRunFunc(ctx, scheduleID, executionID)
	}
	return nil
}

func (m *MockService) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

// MockExecutor for testing
type MockExecutor struct {
	executedSchedules []string
	mu                sync.Mutex
	executeFunc       func(ctx context.Context, tenantID, workflowID, scheduleID string) (string, error)
}

func (m *MockExecutor) ExecuteScheduled(ctx context.Context, tenantID, workflowID, scheduleID string) (string, error) {
	m.mu.Lock()
	m.executedSchedules = append(m.executedSchedules, scheduleID)
	m.mu.Unlock()

	if m.executeFunc != nil {
		return m.executeFunc(ctx, tenantID, workflowID, scheduleID)
	}
	return "execution-123", nil
}

func (m *MockExecutor) GetExecutedSchedules() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string{}, m.executedSchedules...)
}

func TestSchedulerStartStop(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Reduce noise in tests
	}))

	mockService := &MockService{}
	mockExecutor := &MockExecutor{}

	scheduler := NewScheduler(mockService, mockExecutor, logger)
	scheduler.SetCheckInterval(100 * time.Millisecond) // Fast interval for testing

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduler
	err := scheduler.Start(ctx)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Verify it's running
	if !scheduler.IsRunning() {
		t.Error("IsRunning() should return true after Start()")
	}

	// Let it run for a bit
	time.Sleep(250 * time.Millisecond)

	// Stop scheduler
	scheduler.Stop()

	// Verify it stopped
	time.Sleep(100 * time.Millisecond)
	if scheduler.IsRunning() {
		t.Error("IsRunning() should return false after Stop()")
	}

	// Verify service was called at least once
	callCount := mockService.GetCallCount()
	if callCount < 1 {
		t.Errorf("GetDueSchedules() should be called at least once, got %d calls", callCount)
	}
}

func TestSchedulerExecutesDueSchedules(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	now := time.Now()
	dueSchedule := &Schedule{
		ID:          "schedule-1",
		TenantID:    "tenant-1",
		WorkflowID:  "workflow-1",
		Name:        "Test Schedule",
		Enabled:     true,
		NextRunAt:   &now,
		CronExpression: "0 12 * * *",
		Timezone:    "UTC",
	}

	mockService := &MockService{
		getDueSchedulesFunc: func(ctx context.Context) ([]*Schedule, error) {
			return []*Schedule{dueSchedule}, nil
		},
		markScheduleRunFunc: func(ctx context.Context, scheduleID, executionID string) error {
			return nil
		},
	}

	mockExecutor := &MockExecutor{
		executeFunc: func(ctx context.Context, tenantID, workflowID, scheduleID string) (string, error) {
			// Verify correct parameters
			if tenantID != "tenant-1" {
				t.Errorf("ExecuteScheduled() tenantID = %v, want %v", tenantID, "tenant-1")
			}
			if workflowID != "workflow-1" {
				t.Errorf("ExecuteScheduled() workflowID = %v, want %v", workflowID, "workflow-1")
			}
			if scheduleID != "schedule-1" {
				t.Errorf("ExecuteScheduled() scheduleID = %v, want %v", scheduleID, "schedule-1")
			}
			return "execution-123", nil
		},
	}

	scheduler := NewScheduler(mockService, mockExecutor, logger)
	scheduler.SetCheckInterval(100 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduler
	scheduler.Start(ctx)

	// Wait for execution
	time.Sleep(250 * time.Millisecond)

	// Stop scheduler
	scheduler.Stop()
	scheduler.Wait()

	// Verify schedule was executed
	executedSchedules := mockExecutor.GetExecutedSchedules()
	if len(executedSchedules) == 0 {
		t.Error("No schedules were executed")
	}
	if len(executedSchedules) > 0 && executedSchedules[0] != "schedule-1" {
		t.Errorf("Executed schedule ID = %v, want %v", executedSchedules[0], "schedule-1")
	}
}

func TestSchedulerIgnoresDisabledSchedules(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	now := time.Now()
	disabledSchedule := &Schedule{
		ID:          "schedule-disabled",
		TenantID:    "tenant-1",
		WorkflowID:  "workflow-1",
		Name:        "Disabled Schedule",
		Enabled:     false,
		NextRunAt:   &now,
		CronExpression: "0 12 * * *",
		Timezone:    "UTC",
	}

	mockService := &MockService{
		getDueSchedulesFunc: func(ctx context.Context) ([]*Schedule, error) {
			return []*Schedule{disabledSchedule}, nil
		},
	}

	mockExecutor := &MockExecutor{}

	scheduler := NewScheduler(mockService, mockExecutor, logger)
	scheduler.SetCheckInterval(100 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduler
	scheduler.Start(ctx)

	// Wait for check
	time.Sleep(250 * time.Millisecond)

	// Stop scheduler
	scheduler.Stop()
	scheduler.Wait()

	// Verify schedule was NOT executed
	executedSchedules := mockExecutor.GetExecutedSchedules()
	if len(executedSchedules) != 0 {
		t.Errorf("Disabled schedule should not be executed, but got %d executions", len(executedSchedules))
	}
}

func TestSchedulerMultipleSchedules(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	now := time.Now()
	schedules := []*Schedule{
		{
			ID:             "schedule-1",
			TenantID:       "tenant-1",
			WorkflowID:     "workflow-1",
			Name:           "Schedule 1",
			Enabled:        true,
			NextRunAt:      &now,
			CronExpression: "0 12 * * *",
			Timezone:       "UTC",
		},
		{
			ID:             "schedule-2",
			TenantID:       "tenant-1",
			WorkflowID:     "workflow-2",
			Name:           "Schedule 2",
			Enabled:        true,
			NextRunAt:      &now,
			CronExpression: "0 13 * * *",
			Timezone:       "UTC",
		},
		{
			ID:             "schedule-3",
			TenantID:       "tenant-2",
			WorkflowID:     "workflow-3",
			Name:           "Schedule 3",
			Enabled:        true,
			NextRunAt:      &now,
			CronExpression: "0 14 * * *",
			Timezone:       "UTC",
		},
	}

	var executedOnce sync.Once
	mockService := &MockService{
		getDueSchedulesFunc: func(ctx context.Context) ([]*Schedule, error) {
			// Return schedules only on first call to avoid multiple executions
			var result []*Schedule
			executedOnce.Do(func() {
				result = schedules
			})
			return result, nil
		},
		markScheduleRunFunc: func(ctx context.Context, scheduleID, executionID string) error {
			return nil
		},
	}

	mockExecutor := &MockExecutor{}

	scheduler := NewScheduler(mockService, mockExecutor, logger)
	scheduler.SetCheckInterval(100 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduler
	scheduler.Start(ctx)

	// Wait for executions
	time.Sleep(250 * time.Millisecond)

	// Stop scheduler
	scheduler.Stop()
	scheduler.Wait()

	// Verify all schedules were executed at least once
	executedSchedules := mockExecutor.GetExecutedSchedules()
	if len(executedSchedules) < 3 {
		t.Errorf("Expected at least 3 schedules to be executed, got %d", len(executedSchedules))
	}

	// Verify all schedule IDs are present
	scheduleIDs := make(map[string]bool)
	for _, id := range executedSchedules {
		scheduleIDs[id] = true
	}

	for _, schedule := range schedules {
		if !scheduleIDs[schedule.ID] {
			t.Errorf("Schedule %s was not executed", schedule.ID)
		}
	}
}
