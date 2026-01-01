package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of RepositoryInterface for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, tenantID, createdBy string, input CreateWorkflowInput) (*Workflow, error) {
	args := m.Called(ctx, tenantID, createdBy, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Workflow), args.Error(1)
}

func (m *MockRepository) GetByID(ctx context.Context, tenantID, id string) (*Workflow, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Workflow), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, tenantID, id string, input UpdateWorkflowInput) (*Workflow, error) {
	args := m.Called(ctx, tenantID, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Workflow), args.Error(1)
}

func (m *MockRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *MockRepository) List(ctx context.Context, tenantID string, limit, offset int) ([]*Workflow, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Workflow), args.Error(1)
}

func (m *MockRepository) CreateExecution(ctx context.Context, tenantID, workflowID string, workflowVersion int, triggerType string, triggerData []byte) (*Execution, error) {
	args := m.Called(ctx, tenantID, workflowID, workflowVersion, triggerType, triggerData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Execution), args.Error(1)
}

func (m *MockRepository) GetExecutionByID(ctx context.Context, tenantID, id string) (*Execution, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Execution), args.Error(1)
}

func (m *MockRepository) UpdateExecutionStatus(ctx context.Context, id string, status ExecutionStatus, outputData []byte, errorMessage *string) error {
	args := m.Called(ctx, id, status, outputData, errorMessage)
	return args.Error(0)
}

func (m *MockRepository) GetStepExecutionsByExecutionID(ctx context.Context, executionID string) ([]*StepExecution, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*StepExecution), args.Error(1)
}

func (m *MockRepository) ListExecutions(ctx context.Context, tenantID string, workflowID string, limit, offset int) ([]*Execution, error) {
	args := m.Called(ctx, tenantID, workflowID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Execution), args.Error(1)
}

func (m *MockRepository) ListExecutionsAdvanced(ctx context.Context, tenantID string, filter ExecutionFilter, cursor string, limit int) (*ExecutionListResult, error) {
	args := m.Called(ctx, tenantID, filter, cursor, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ExecutionListResult), args.Error(1)
}

func (m *MockRepository) GetExecutionWithSteps(ctx context.Context, tenantID, executionID string) (*ExecutionWithSteps, error) {
	args := m.Called(ctx, tenantID, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ExecutionWithSteps), args.Error(1)
}

func (m *MockRepository) CountExecutions(ctx context.Context, tenantID string, filter ExecutionFilter) (int, error) {
	args := m.Called(ctx, tenantID, filter)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) CreateWorkflowVersion(ctx context.Context, workflowID string, version int, definition json.RawMessage, createdBy string) (*WorkflowVersion, error) {
	args := m.Called(ctx, workflowID, version, definition, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*WorkflowVersion), args.Error(1)
}

func (m *MockRepository) ListWorkflowVersions(ctx context.Context, workflowID string) ([]*WorkflowVersion, error) {
	args := m.Called(ctx, workflowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*WorkflowVersion), args.Error(1)
}

func (m *MockRepository) GetWorkflowVersion(ctx context.Context, workflowID string, version int) (*WorkflowVersion, error) {
	args := m.Called(ctx, workflowID, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*WorkflowVersion), args.Error(1)
}

func (m *MockRepository) RestoreWorkflowVersion(ctx context.Context, tenantID, workflowID string, version int) (*Workflow, error) {
	args := m.Called(ctx, tenantID, workflowID, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Workflow), args.Error(1)
}

func newTestService() (*Service, *MockRepository) {
	mockRepo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	service := &Service{
		repo:   mockRepo,
		logger: logger,
	}

	return service, mockRepo
}

// TestListExecutionsAdvanced_Success tests successful execution listing with filters
func TestListExecutionsAdvanced_Success(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"

	now := time.Now()
	executions := []*Execution{
		{
			ID:          "exec-1",
			TenantID:    tenantID,
			WorkflowID:  "workflow-1",
			Status:      "completed",
			TriggerType: "webhook",
			CreatedAt:   now,
		},
		{
			ID:          "exec-2",
			TenantID:    tenantID,
			WorkflowID:  "workflow-1",
			Status:      "running",
			TriggerType: "manual",
			CreatedAt:   now.Add(-1 * time.Hour),
		},
	}

	expectedResult := &ExecutionListResult{
		Data:       executions,
		Cursor:     "next-cursor",
		HasMore:    true,
		TotalCount: 10,
	}

	tests := []struct {
		name   string
		filter ExecutionFilter
		cursor string
		limit  int
	}{
		{
			name:   "list all with default limit",
			filter: ExecutionFilter{},
			cursor: "",
			limit:  20,
		},
		{
			name: "filter by workflow_id",
			filter: ExecutionFilter{
				WorkflowID: "workflow-1",
			},
			cursor: "",
			limit:  10,
		},
		{
			name: "filter by status",
			filter: ExecutionFilter{
				Status: "completed",
			},
			cursor: "",
			limit:  50,
		},
		{
			name: "filter by trigger_type",
			filter: ExecutionFilter{
				TriggerType: "webhook",
			},
			cursor: "",
			limit:  25,
		},
		{
			name: "with cursor pagination",
			filter: ExecutionFilter{
				WorkflowID: "workflow-1",
			},
			cursor: "previous-cursor",
			limit:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.On("ListExecutionsAdvanced", ctx, tenantID, tt.filter, tt.cursor, tt.limit).
				Return(expectedResult, nil).Once()

			result, err := service.ListExecutionsAdvanced(ctx, tenantID, tt.filter, tt.cursor, tt.limit)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, expectedResult, result)
			assert.Equal(t, len(executions), len(result.Data))
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestListExecutionsAdvanced_DefaultLimit tests default limit application
func TestListExecutionsAdvanced_DefaultLimit(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"
	filter := ExecutionFilter{}

	expectedResult := &ExecutionListResult{
		Data:       []*Execution{},
		Cursor:     "",
		HasMore:    false,
		TotalCount: 0,
	}

	tests := []struct {
		name          string
		inputLimit    int
		expectedLimit int
	}{
		{
			name:          "zero limit uses default",
			inputLimit:    0,
			expectedLimit: 20,
		},
		{
			name:          "negative limit uses default",
			inputLimit:    -1,
			expectedLimit: 20,
		},
		{
			name:          "valid limit is preserved",
			inputLimit:    50,
			expectedLimit: 50,
		},
		{
			name:          "max limit is capped at 100",
			inputLimit:    150,
			expectedLimit: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.On("ListExecutionsAdvanced", ctx, tenantID, filter, "", tt.expectedLimit).
				Return(expectedResult, nil).Once()

			result, err := service.ListExecutionsAdvanced(ctx, tenantID, filter, "", tt.inputLimit)

			require.NoError(t, err)
			assert.NotNil(t, result)
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestListExecutionsAdvanced_FilterValidation tests filter validation
func TestListExecutionsAdvanced_FilterValidation(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"

	startDate := time.Now()
	endDate := startDate.Add(-1 * time.Hour) // End before start (invalid)

	invalidFilter := ExecutionFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	mockRepo.On("ListExecutionsAdvanced", ctx, tenantID, invalidFilter, "", 20).
		Return(nil, errors.New("invalid filter: end_date must be after start_date"))

	result, err := service.ListExecutionsAdvanced(ctx, tenantID, invalidFilter, "", 20)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid filter")
}

// TestListExecutionsAdvanced_RepositoryError tests repository error handling
func TestListExecutionsAdvanced_RepositoryError(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"
	filter := ExecutionFilter{}

	mockRepo.On("ListExecutionsAdvanced", ctx, tenantID, filter, "", 20).
		Return(nil, errors.New("database connection error"))

	result, err := service.ListExecutionsAdvanced(ctx, tenantID, filter, "", 20)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database connection error")
}

// TestGetExecutionWithSteps_Success tests successful retrieval of execution with steps
func TestGetExecutionWithSteps_Success(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"
	executionID := "exec-123"

	now := time.Now()
	expectedResult := &ExecutionWithSteps{
		Execution: &Execution{
			ID:          executionID,
			TenantID:    tenantID,
			WorkflowID:  "workflow-1",
			Status:      "completed",
			TriggerType: "webhook",
			CreatedAt:   now,
		},
		Steps: []*StepExecution{
			{
				ID:          "step-1",
				ExecutionID: executionID,
				NodeID:      "node-1",
				NodeType:    "action:http",
				Status:      "completed",
				StartedAt:   &now,
			},
			{
				ID:          "step-2",
				ExecutionID: executionID,
				NodeID:      "node-2",
				NodeType:    "action:transform",
				Status:      "completed",
				StartedAt:   &now,
			},
		},
	}

	mockRepo.On("GetExecutionWithSteps", ctx, tenantID, executionID).
		Return(expectedResult, nil)

	result, err := service.GetExecutionWithSteps(ctx, tenantID, executionID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, executionID, result.Execution.ID)
	assert.Equal(t, 2, len(result.Steps))
	assert.Equal(t, "step-1", result.Steps[0].ID)
	assert.Equal(t, "step-2", result.Steps[1].ID)
	mockRepo.AssertExpectations(t)
}

// TestGetExecutionWithSteps_NotFound tests execution not found error
func TestGetExecutionWithSteps_NotFound(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"
	executionID := "non-existent"

	mockRepo.On("GetExecutionWithSteps", ctx, tenantID, executionID).
		Return(nil, ErrNotFound)

	result, err := service.GetExecutionWithSteps(ctx, tenantID, executionID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrNotFound, err)
	mockRepo.AssertExpectations(t)
}

// TestGetExecutionWithSteps_TenantIsolation tests that tenant isolation is enforced
func TestGetExecutionWithSteps_TenantIsolation(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	otherTenantID := "tenant-456"
	executionID := "exec-123"

	// Mock returns not found when accessing with wrong tenant
	mockRepo.On("GetExecutionWithSteps", ctx, otherTenantID, executionID).
		Return(nil, ErrNotFound)

	result, err := service.GetExecutionWithSteps(ctx, otherTenantID, executionID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrNotFound, err)
	mockRepo.AssertExpectations(t)
}

// TestGetExecutionStats_Success tests successful stats retrieval
func TestGetExecutionStats_Success(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"

	tests := []struct {
		name          string
		filter        ExecutionFilter
		mockCounts    map[string]int
		expectedStats ExecutionStats
	}{
		{
			name:   "all executions stats",
			filter: ExecutionFilter{},
			mockCounts: map[string]int{
				"pending":   5,
				"running":   3,
				"completed": 20,
				"failed":    2,
				"cancelled": 1,
			},
			expectedStats: ExecutionStats{
				TotalCount: 31,
				StatusCounts: map[string]int{
					"pending":   5,
					"running":   3,
					"completed": 20,
					"failed":    2,
					"cancelled": 1,
				},
			},
		},
		{
			name: "workflow-specific stats",
			filter: ExecutionFilter{
				WorkflowID: "workflow-1",
			},
			mockCounts: map[string]int{
				"pending":   2,
				"running":   1,
				"completed": 10,
				"failed":    1,
				"cancelled": 0,
			},
			expectedStats: ExecutionStats{
				TotalCount: 14,
				StatusCounts: map[string]int{
					"pending":   2,
					"running":   1,
					"completed": 10,
					"failed":    1,
					"cancelled": 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock count for each status
			for status, count := range tt.mockCounts {
				statusFilter := ExecutionFilter{
					WorkflowID:  tt.filter.WorkflowID,
					Status:      status,
					TriggerType: tt.filter.TriggerType,
					StartDate:   tt.filter.StartDate,
					EndDate:     tt.filter.EndDate,
				}
				mockRepo.On("CountExecutions", ctx, tenantID, statusFilter).
					Return(count, nil).Once()
			}

			result, err := service.GetExecutionStats(ctx, tenantID, tt.filter)

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedStats.TotalCount, result.TotalCount)
			assert.Equal(t, tt.expectedStats.StatusCounts, result.StatusCounts)
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestGetExecutionStats_FilterValidation tests filter validation for stats
func TestGetExecutionStats_FilterValidation(t *testing.T) {
	service, _ := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"

	startDate := time.Now()
	endDate := startDate.Add(-1 * time.Hour) // End before start (invalid)

	invalidFilter := ExecutionFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	result, err := service.GetExecutionStats(ctx, tenantID, invalidFilter)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid filter")
}

// TestGetExecutionStats_RepositoryError tests repository error handling for stats
func TestGetExecutionStats_RepositoryError(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"
	filter := ExecutionFilter{}

	// Mock error for first status count
	mockRepo.On("CountExecutions", ctx, tenantID, mock.Anything).
		Return(0, errors.New("database error")).Once()

	result, err := service.GetExecutionStats(ctx, tenantID, filter)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
}

// TestGetExecutionStats_ZeroCounts tests stats with zero executions
func TestGetExecutionStats_ZeroCounts(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"
	filter := ExecutionFilter{}

	// Mock zero count for all statuses
	statuses := []string{"pending", "running", "completed", "failed", "cancelled"}
	for _, status := range statuses {
		statusFilter := ExecutionFilter{Status: status}
		mockRepo.On("CountExecutions", ctx, tenantID, statusFilter).
			Return(0, nil).Once()
	}

	result, err := service.GetExecutionStats(ctx, tenantID, filter)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.TotalCount)
	assert.Equal(t, 0, result.StatusCounts["pending"])
	assert.Equal(t, 0, result.StatusCounts["running"])
	assert.Equal(t, 0, result.StatusCounts["completed"])
	assert.Equal(t, 0, result.StatusCounts["failed"])
	assert.Equal(t, 0, result.StatusCounts["cancelled"])
	mockRepo.AssertExpectations(t)
}
