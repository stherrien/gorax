package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBulkRepository is a mock implementation of RepositoryInterface for testing
type MockBulkRepository struct {
	mock.Mock
}

func (m *MockBulkRepository) GetByID(ctx context.Context, tenantID, id string) (*Workflow, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Workflow), args.Error(1)
}

func (m *MockBulkRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *MockBulkRepository) Update(ctx context.Context, tenantID, id string, input UpdateWorkflowInput) (*Workflow, error) {
	args := m.Called(ctx, tenantID, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Workflow), args.Error(1)
}

func (m *MockBulkRepository) Create(ctx context.Context, tenantID, createdBy string, input CreateWorkflowInput) (*Workflow, error) {
	args := m.Called(ctx, tenantID, createdBy, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Workflow), args.Error(1)
}

func (m *MockBulkRepository) List(ctx context.Context, tenantID string, limit, offset int) ([]*Workflow, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Workflow), args.Error(1)
}

func (m *MockBulkRepository) CreateExecution(ctx context.Context, tenantID, workflowID string, workflowVersion int, triggerType string, triggerData []byte) (*Execution, error) {
	args := m.Called(ctx, tenantID, workflowID, workflowVersion, triggerType, triggerData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Execution), args.Error(1)
}

func (m *MockBulkRepository) GetExecutionByID(ctx context.Context, tenantID, id string) (*Execution, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Execution), args.Error(1)
}

func (m *MockBulkRepository) UpdateExecutionStatus(ctx context.Context, id string, status ExecutionStatus, outputData []byte, errorMessage *string) error {
	args := m.Called(ctx, id, status, outputData, errorMessage)
	return args.Error(0)
}

func (m *MockBulkRepository) GetStepExecutionsByExecutionID(ctx context.Context, executionID string) ([]*StepExecution, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*StepExecution), args.Error(1)
}

func (m *MockBulkRepository) ListExecutions(ctx context.Context, tenantID string, workflowID string, limit, offset int) ([]*Execution, error) {
	args := m.Called(ctx, tenantID, workflowID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Execution), args.Error(1)
}

func (m *MockBulkRepository) ListExecutionsAdvanced(ctx context.Context, tenantID string, filter ExecutionFilter, cursor string, limit int) (*ExecutionListResult, error) {
	args := m.Called(ctx, tenantID, filter, cursor, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ExecutionListResult), args.Error(1)
}

func (m *MockBulkRepository) GetExecutionWithSteps(ctx context.Context, tenantID, executionID string) (*ExecutionWithSteps, error) {
	args := m.Called(ctx, tenantID, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ExecutionWithSteps), args.Error(1)
}

func (m *MockBulkRepository) CountExecutions(ctx context.Context, tenantID string, filter ExecutionFilter) (int, error) {
	args := m.Called(ctx, tenantID, filter)
	return args.Int(0), args.Error(1)
}

func (m *MockBulkRepository) CreateWorkflowVersion(ctx context.Context, workflowID string, version int, definition json.RawMessage, createdBy string) (*WorkflowVersion, error) {
	args := m.Called(ctx, workflowID, version, definition, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*WorkflowVersion), args.Error(1)
}

func (m *MockBulkRepository) ListWorkflowVersions(ctx context.Context, workflowID string) ([]*WorkflowVersion, error) {
	args := m.Called(ctx, workflowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*WorkflowVersion), args.Error(1)
}

func (m *MockBulkRepository) GetWorkflowVersion(ctx context.Context, workflowID string, version int) (*WorkflowVersion, error) {
	args := m.Called(ctx, workflowID, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*WorkflowVersion), args.Error(1)
}

func (m *MockBulkRepository) RestoreWorkflowVersion(ctx context.Context, tenantID, workflowID string, version int) (*Workflow, error) {
	args := m.Called(ctx, tenantID, workflowID, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Workflow), args.Error(1)
}

// MockWebhookService for testing
type MockWebhookService struct {
	mock.Mock
}

func (m *MockWebhookService) SyncWorkflowWebhooks(ctx context.Context, tenantID, workflowID string, nodes []WebhookNodeConfig) error {
	args := m.Called(ctx, tenantID, workflowID, nodes)
	return args.Error(0)
}

func (m *MockWebhookService) DeleteByWorkflowID(ctx context.Context, workflowID string) error {
	args := m.Called(ctx, workflowID)
	return args.Error(0)
}

func (m *MockWebhookService) GetByWorkflowID(ctx context.Context, workflowID string) ([]*WebhookInfo, error) {
	args := m.Called(ctx, workflowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*WebhookInfo), args.Error(1)
}

// Test BulkDelete
func TestBulkService_BulkDelete(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("successfully deletes multiple workflows", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{"workflow-1", "workflow-2", "workflow-3"}

		// Mock successful deletions
		for _, id := range workflowIDs {
			mockWebhook.On("DeleteByWorkflowID", ctx, id).Return(nil)
			mockRepo.On("Delete", ctx, tenantID, id).Return(nil)
		}

		result := bulkService.BulkDelete(ctx, tenantID, workflowIDs)

		assert.Equal(t, 3, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
		mockRepo.AssertExpectations(t)
		mockWebhook.AssertExpectations(t)
	})

	t.Run("handles partial failures", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{"workflow-1", "workflow-2", "workflow-3"}

		// First workflow succeeds
		mockWebhook.On("DeleteByWorkflowID", ctx, "workflow-1").Return(nil)
		mockRepo.On("Delete", ctx, tenantID, "workflow-1").Return(nil)

		// Second workflow fails
		mockWebhook.On("DeleteByWorkflowID", ctx, "workflow-2").Return(nil)
		mockRepo.On("Delete", ctx, tenantID, "workflow-2").Return(errors.New("database error"))

		// Third workflow succeeds
		mockWebhook.On("DeleteByWorkflowID", ctx, "workflow-3").Return(nil)
		mockRepo.On("Delete", ctx, tenantID, "workflow-3").Return(nil)

		result := bulkService.BulkDelete(ctx, tenantID, workflowIDs)

		assert.Equal(t, 2, result.SuccessCount)
		assert.Equal(t, 1, len(result.Failures))
		assert.Equal(t, "workflow-2", result.Failures[0].WorkflowID)
		assert.Contains(t, result.Failures[0].Error, "database error")
		mockRepo.AssertExpectations(t)
		mockWebhook.AssertExpectations(t)
	})

	t.Run("handles empty workflow list", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{}

		result := bulkService.BulkDelete(ctx, tenantID, workflowIDs)

		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
	})

	t.Run("continues on webhook deletion failure", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{"workflow-1"}

		// Webhook deletion fails but workflow deletion succeeds
		mockWebhook.On("DeleteByWorkflowID", ctx, "workflow-1").Return(errors.New("webhook error"))
		mockRepo.On("Delete", ctx, tenantID, "workflow-1").Return(nil)

		result := bulkService.BulkDelete(ctx, tenantID, workflowIDs)

		assert.Equal(t, 1, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
		mockRepo.AssertExpectations(t)
		mockWebhook.AssertExpectations(t)
	})
}

// Test BulkEnable
func TestBulkService_BulkEnable(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("successfully enables multiple workflows", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{"workflow-1", "workflow-2"}

		for _, id := range workflowIDs {
			workflow := &Workflow{
				ID:     id,
				Status: "inactive",
			}
			mockRepo.On("Update", ctx, tenantID, id, mock.MatchedBy(func(input UpdateWorkflowInput) bool {
				return input.Status == "active"
			})).Return(workflow, nil)
		}

		result := bulkService.BulkEnable(ctx, tenantID, workflowIDs)

		assert.Equal(t, 2, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
		mockRepo.AssertExpectations(t)
	})

	t.Run("handles update failures", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{"workflow-1", "workflow-2"}

		workflow := &Workflow{ID: "workflow-1", Status: "active"}
		mockRepo.On("Update", ctx, tenantID, "workflow-1", mock.Anything).Return(workflow, nil)
		mockRepo.On("Update", ctx, tenantID, "workflow-2", mock.Anything).Return(nil, errors.New("not found"))

		result := bulkService.BulkEnable(ctx, tenantID, workflowIDs)

		assert.Equal(t, 1, result.SuccessCount)
		assert.Equal(t, 1, len(result.Failures))
		assert.Equal(t, "workflow-2", result.Failures[0].WorkflowID)
		mockRepo.AssertExpectations(t)
	})
}

// Test BulkDisable
func TestBulkService_BulkDisable(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("successfully disables multiple workflows", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{"workflow-1", "workflow-2"}

		for _, id := range workflowIDs {
			workflow := &Workflow{
				ID:     id,
				Status: "inactive",
			}
			mockRepo.On("Update", ctx, tenantID, id, mock.MatchedBy(func(input UpdateWorkflowInput) bool {
				return input.Status == "inactive"
			})).Return(workflow, nil)
		}

		result := bulkService.BulkDisable(ctx, tenantID, workflowIDs)

		assert.Equal(t, 2, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
		mockRepo.AssertExpectations(t)
	})
}

// Test BulkExport
func TestBulkService_BulkExport(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("successfully exports multiple workflows", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{"workflow-1", "workflow-2"}

		definition1 := json.RawMessage(`{"nodes":[],"edges":[]}`)
		definition2 := json.RawMessage(`{"nodes":[],"edges":[]}`)

		workflow1 := &Workflow{
			ID:          "workflow-1",
			Name:        "Test Workflow 1",
			Description: "Description 1",
			Definition:  definition1,
			Status:      "active",
			Version:     1,
		}

		workflow2 := &Workflow{
			ID:          "workflow-2",
			Name:        "Test Workflow 2",
			Description: "Description 2",
			Definition:  definition2,
			Status:      "active",
			Version:     2,
		}

		mockRepo.On("GetByID", ctx, tenantID, "workflow-1").Return(workflow1, nil)
		mockRepo.On("GetByID", ctx, tenantID, "workflow-2").Return(workflow2, nil)

		export, result := bulkService.BulkExport(ctx, tenantID, workflowIDs)

		assert.Equal(t, 2, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
		assert.Equal(t, 2, len(export.Workflows))
		assert.Equal(t, "workflow-1", export.Workflows[0].ID)
		assert.Equal(t, "workflow-2", export.Workflows[1].ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("handles missing workflows", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{"workflow-1", "workflow-2"}

		definition1 := json.RawMessage(`{"nodes":[],"edges":[]}`)
		workflow1 := &Workflow{
			ID:         "workflow-1",
			Name:       "Test Workflow 1",
			Definition: definition1,
			Status:     "active",
			Version:    1,
		}

		mockRepo.On("GetByID", ctx, tenantID, "workflow-1").Return(workflow1, nil)
		mockRepo.On("GetByID", ctx, tenantID, "workflow-2").Return(nil, ErrNotFound)

		export, result := bulkService.BulkExport(ctx, tenantID, workflowIDs)

		assert.Equal(t, 1, result.SuccessCount)
		assert.Equal(t, 1, len(result.Failures))
		assert.Equal(t, 1, len(export.Workflows))
		assert.Equal(t, "workflow-2", result.Failures[0].WorkflowID)
		mockRepo.AssertExpectations(t)
	})
}

// Test BulkClone
func TestBulkService_BulkClone(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("successfully clones multiple workflows", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		userID := "user-1"
		workflowIDs := []string{"workflow-1", "workflow-2"}

		definition1 := json.RawMessage(`{"nodes":[],"edges":[]}`)
		definition2 := json.RawMessage(`{"nodes":[],"edges":[]}`)

		workflow1 := &Workflow{
			ID:          "workflow-1",
			Name:        "Test Workflow 1",
			Description: "Description 1",
			Definition:  definition1,
			Status:      "active",
		}

		workflow2 := &Workflow{
			ID:          "workflow-2",
			Name:        "Test Workflow 2",
			Description: "Description 2",
			Definition:  definition2,
			Status:      "active",
		}

		cloned1 := &Workflow{
			ID:          "cloned-1",
			Name:        "Test Workflow 1 (Copy)",
			Description: "Description 1",
			Definition:  definition1,
			Status:      "draft",
		}

		cloned2 := &Workflow{
			ID:          "cloned-2",
			Name:        "Test Workflow 2 (Copy)",
			Description: "Description 2",
			Definition:  definition2,
			Status:      "draft",
		}

		mockRepo.On("GetByID", ctx, tenantID, "workflow-1").Return(workflow1, nil)
		mockRepo.On("GetByID", ctx, tenantID, "workflow-2").Return(workflow2, nil)

		mockRepo.On("Create", ctx, tenantID, userID, mock.MatchedBy(func(input CreateWorkflowInput) bool {
			return input.Name == "Test Workflow 1 (Copy)"
		})).Return(cloned1, nil)

		mockRepo.On("Create", ctx, tenantID, userID, mock.MatchedBy(func(input CreateWorkflowInput) bool {
			return input.Name == "Test Workflow 2 (Copy)"
		})).Return(cloned2, nil)

		// Webhooks sync is not expected for workflows with no webhook nodes in empty definitions
		// mockWebhook.On("SyncWorkflowWebhooks", ctx, tenantID, "cloned-1", mock.Anything).Return(nil)
		// mockWebhook.On("SyncWorkflowWebhooks", ctx, tenantID, "cloned-2", mock.Anything).Return(nil)

		clones, result := bulkService.BulkClone(ctx, tenantID, userID, workflowIDs)

		assert.Equal(t, 2, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
		assert.Equal(t, 2, len(clones))
		assert.Equal(t, "cloned-1", clones[0].ID)
		assert.Equal(t, "cloned-2", clones[1].ID)
		mockRepo.AssertExpectations(t)
		mockWebhook.AssertExpectations(t)
	})

	t.Run("handles clone failures", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		userID := "user-1"
		workflowIDs := []string{"workflow-1", "workflow-2"}

		definition1 := json.RawMessage(`{"nodes":[],"edges":[]}`)
		workflow1 := &Workflow{
			ID:         "workflow-1",
			Name:       "Test Workflow 1",
			Definition: definition1,
			Status:     "active",
		}

		mockRepo.On("GetByID", ctx, tenantID, "workflow-1").Return(workflow1, nil)
		mockRepo.On("GetByID", ctx, tenantID, "workflow-2").Return(nil, ErrNotFound)

		cloned1 := &Workflow{
			ID:         "cloned-1",
			Name:       "Test Workflow 1 (Copy)",
			Definition: definition1,
			Status:     "draft",
		}

		mockRepo.On("Create", ctx, tenantID, userID, mock.Anything).Return(cloned1, nil)
		// Webhook sync is not expected for workflows with no webhook nodes in empty definitions

		clones, result := bulkService.BulkClone(ctx, tenantID, userID, workflowIDs)

		assert.Equal(t, 1, result.SuccessCount)
		assert.Equal(t, 1, len(result.Failures))
		assert.Equal(t, 1, len(clones))
		assert.Equal(t, "workflow-2", result.Failures[0].WorkflowID)
		mockRepo.AssertExpectations(t)
	})
}

// ============================================================
// Edge Case Tests
// ============================================================

// Test BulkDelete edge cases
func TestBulkService_BulkDelete_EdgeCases(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("nil workflow IDs slice", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"

		result := bulkService.BulkDelete(ctx, tenantID, nil)

		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
	})

	t.Run("nil webhook service", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		bulkService := NewBulkService(mockRepo, nil, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{"workflow-1"}

		mockRepo.On("Delete", ctx, tenantID, "workflow-1").Return(nil)

		result := bulkService.BulkDelete(ctx, tenantID, workflowIDs)

		assert.Equal(t, 1, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
		mockRepo.AssertExpectations(t)
	})

	t.Run("all workflows fail deletion", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{"workflow-1", "workflow-2", "workflow-3"}

		for _, id := range workflowIDs {
			mockWebhook.On("DeleteByWorkflowID", ctx, id).Return(nil)
			mockRepo.On("Delete", ctx, tenantID, id).Return(errors.New("database error"))
		}

		result := bulkService.BulkDelete(ctx, tenantID, workflowIDs)

		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 3, len(result.Failures))
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty tenant ID", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		workflowIDs := []string{"workflow-1"}

		mockWebhook.On("DeleteByWorkflowID", ctx, "workflow-1").Return(nil)
		mockRepo.On("Delete", ctx, "", "workflow-1").Return(nil)

		result := bulkService.BulkDelete(ctx, "", workflowIDs)

		assert.Equal(t, 1, result.SuccessCount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("duplicate workflow IDs", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{"workflow-1", "workflow-1", "workflow-1"}

		// Each duplicate will be processed separately
		mockWebhook.On("DeleteByWorkflowID", ctx, "workflow-1").Return(nil).Times(3)
		mockRepo.On("Delete", ctx, tenantID, "workflow-1").Return(nil).Once()
		mockRepo.On("Delete", ctx, tenantID, "workflow-1").Return(ErrNotFound).Twice()

		result := bulkService.BulkDelete(ctx, tenantID, workflowIDs)

		assert.Equal(t, 1, result.SuccessCount)
		assert.Equal(t, 2, len(result.Failures))
		mockRepo.AssertExpectations(t)
	})

	t.Run("large workflow list", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"

		// Create 100 workflow IDs
		workflowIDs := make([]string, 100)
		for i := 0; i < 100; i++ {
			id := "workflow-" + string(rune('0'+i/10)) + string(rune('0'+i%10))
			workflowIDs[i] = id
			mockWebhook.On("DeleteByWorkflowID", ctx, id).Return(nil)
			mockRepo.On("Delete", ctx, tenantID, id).Return(nil)
		}

		result := bulkService.BulkDelete(ctx, tenantID, workflowIDs)

		assert.Equal(t, 100, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
	})
}

// Test BulkEnable edge cases
func TestBulkService_BulkEnable_EdgeCases(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("nil workflow IDs slice", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"

		result := bulkService.BulkEnable(ctx, tenantID, nil)

		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
	})

	t.Run("empty workflow IDs slice", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"

		result := bulkService.BulkEnable(ctx, tenantID, []string{})

		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
	})

	t.Run("all workflows fail to enable", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{"workflow-1", "workflow-2"}

		mockRepo.On("Update", ctx, tenantID, "workflow-1", mock.Anything).Return(nil, errors.New("not found"))
		mockRepo.On("Update", ctx, tenantID, "workflow-2", mock.Anything).Return(nil, errors.New("database error"))

		result := bulkService.BulkEnable(ctx, tenantID, workflowIDs)

		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 2, len(result.Failures))
		mockRepo.AssertExpectations(t)
	})

	t.Run("already enabled workflows", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{"workflow-1"}

		// Already active workflow
		workflow := &Workflow{ID: "workflow-1", Status: "active"}
		mockRepo.On("Update", ctx, tenantID, "workflow-1", mock.Anything).Return(workflow, nil)

		result := bulkService.BulkEnable(ctx, tenantID, workflowIDs)

		assert.Equal(t, 1, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
		mockRepo.AssertExpectations(t)
	})
}

// Test BulkDisable edge cases
func TestBulkService_BulkDisable_EdgeCases(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("nil workflow IDs slice", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"

		result := bulkService.BulkDisable(ctx, tenantID, nil)

		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
	})

	t.Run("all workflows fail to disable", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{"workflow-1", "workflow-2"}

		mockRepo.On("Update", ctx, tenantID, "workflow-1", mock.Anything).Return(nil, errors.New("not found"))
		mockRepo.On("Update", ctx, tenantID, "workflow-2", mock.Anything).Return(nil, errors.New("database error"))

		result := bulkService.BulkDisable(ctx, tenantID, workflowIDs)

		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 2, len(result.Failures))
		mockRepo.AssertExpectations(t)
	})

	t.Run("already disabled workflows", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{"workflow-1"}

		// Already inactive workflow
		workflow := &Workflow{ID: "workflow-1", Status: "inactive"}
		mockRepo.On("Update", ctx, tenantID, "workflow-1", mock.Anything).Return(workflow, nil)

		result := bulkService.BulkDisable(ctx, tenantID, workflowIDs)

		assert.Equal(t, 1, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
		mockRepo.AssertExpectations(t)
	})
}

// Test BulkExport edge cases
func TestBulkService_BulkExport_EdgeCases(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("nil workflow IDs slice", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"

		export, result := bulkService.BulkExport(ctx, tenantID, nil)

		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
		assert.Equal(t, 0, len(export.Workflows))
		assert.Equal(t, "1.0", export.Version)
	})

	t.Run("empty workflow IDs slice", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"

		export, result := bulkService.BulkExport(ctx, tenantID, []string{})

		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
		assert.Equal(t, 0, len(export.Workflows))
	})

	t.Run("all workflows fail to export", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{"workflow-1", "workflow-2"}

		mockRepo.On("GetByID", ctx, tenantID, "workflow-1").Return(nil, ErrNotFound)
		mockRepo.On("GetByID", ctx, tenantID, "workflow-2").Return(nil, errors.New("database error"))

		export, result := bulkService.BulkExport(ctx, tenantID, workflowIDs)

		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 2, len(result.Failures))
		assert.Equal(t, 0, len(export.Workflows))
		mockRepo.AssertExpectations(t)
	})

	t.Run("workflow with nil definition", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{"workflow-1"}

		workflow := &Workflow{
			ID:          "workflow-1",
			Name:        "Test Workflow",
			Description: "Description",
			Definition:  nil,
			Status:      "draft",
			Version:     1,
		}

		mockRepo.On("GetByID", ctx, tenantID, "workflow-1").Return(workflow, nil)

		export, result := bulkService.BulkExport(ctx, tenantID, workflowIDs)

		assert.Equal(t, 1, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
		assert.Equal(t, 1, len(export.Workflows))
		assert.Nil(t, export.Workflows[0].Definition)
		mockRepo.AssertExpectations(t)
	})

	t.Run("export preserves all workflow data", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		workflowIDs := []string{"workflow-1"}

		definition := json.RawMessage(`{"nodes":[{"id":"node-1"}],"edges":[]}`)
		workflow := &Workflow{
			ID:          "workflow-1",
			Name:        "Test Workflow",
			Description: "Test Description",
			Definition:  definition,
			Status:      "active",
			Version:     5,
		}

		mockRepo.On("GetByID", ctx, tenantID, "workflow-1").Return(workflow, nil)

		export, result := bulkService.BulkExport(ctx, tenantID, workflowIDs)

		assert.Equal(t, 1, result.SuccessCount)
		assert.Equal(t, "workflow-1", export.Workflows[0].ID)
		assert.Equal(t, "Test Workflow", export.Workflows[0].Name)
		assert.Equal(t, "Test Description", export.Workflows[0].Description)
		assert.Equal(t, "active", export.Workflows[0].Status)
		assert.Equal(t, 5, export.Workflows[0].Version)
		assert.Equal(t, definition, export.Workflows[0].Definition)
		mockRepo.AssertExpectations(t)
	})
}

// Test BulkClone edge cases
func TestBulkService_BulkClone_EdgeCases(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("nil workflow IDs slice", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		userID := "user-1"

		clones, result := bulkService.BulkClone(ctx, tenantID, userID, nil)

		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
		assert.Equal(t, 0, len(clones))
	})

	t.Run("empty workflow IDs slice", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		userID := "user-1"

		clones, result := bulkService.BulkClone(ctx, tenantID, userID, []string{})

		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
		assert.Equal(t, 0, len(clones))
	})

	t.Run("nil webhook service", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		bulkService := NewBulkService(mockRepo, nil, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		userID := "user-1"
		workflowIDs := []string{"workflow-1"}

		definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
		original := &Workflow{
			ID:          "workflow-1",
			Name:        "Test Workflow",
			Description: "Description",
			Definition:  definition,
			Status:      "active",
		}

		cloned := &Workflow{
			ID:          "cloned-1",
			Name:        "Test Workflow (Copy)",
			Description: "Description",
			Definition:  definition,
			Status:      "draft",
		}

		mockRepo.On("GetByID", ctx, tenantID, "workflow-1").Return(original, nil)
		mockRepo.On("Create", ctx, tenantID, userID, mock.Anything).Return(cloned, nil)

		clones, result := bulkService.BulkClone(ctx, tenantID, userID, workflowIDs)

		assert.Equal(t, 1, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
		assert.Equal(t, 1, len(clones))
		mockRepo.AssertExpectations(t)
	})

	t.Run("clone creation fails", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		userID := "user-1"
		workflowIDs := []string{"workflow-1"}

		definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
		original := &Workflow{
			ID:          "workflow-1",
			Name:        "Test Workflow",
			Description: "Description",
			Definition:  definition,
			Status:      "active",
		}

		mockRepo.On("GetByID", ctx, tenantID, "workflow-1").Return(original, nil)
		mockRepo.On("Create", ctx, tenantID, userID, mock.Anything).Return(nil, errors.New("database error"))

		clones, result := bulkService.BulkClone(ctx, tenantID, userID, workflowIDs)

		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 1, len(result.Failures))
		assert.Equal(t, 0, len(clones))
		assert.Contains(t, result.Failures[0].Error, "create clone")
		mockRepo.AssertExpectations(t)
	})

	t.Run("webhook sync fails but clone succeeds", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		userID := "user-1"
		workflowIDs := []string{"workflow-1"}

		// Workflow with webhook trigger node
		definition := json.RawMessage(`{"nodes":[{"id":"node-1","type":"trigger:webhook","data":{"config":{}}}],"edges":[]}`)
		original := &Workflow{
			ID:          "workflow-1",
			Name:        "Test Workflow",
			Description: "Description",
			Definition:  definition,
			Status:      "active",
		}

		cloned := &Workflow{
			ID:          "cloned-1",
			Name:        "Test Workflow (Copy)",
			Description: "Description",
			Definition:  definition,
			Status:      "draft",
		}

		mockRepo.On("GetByID", ctx, tenantID, "workflow-1").Return(original, nil)
		mockRepo.On("Create", ctx, tenantID, userID, mock.Anything).Return(cloned, nil)
		mockWebhook.On("SyncWorkflowWebhooks", ctx, tenantID, "cloned-1", mock.Anything).Return(errors.New("webhook sync error"))

		clones, result := bulkService.BulkClone(ctx, tenantID, userID, workflowIDs)

		// Clone should succeed even if webhook sync fails
		assert.Equal(t, 1, result.SuccessCount)
		assert.Equal(t, 0, len(result.Failures))
		assert.Equal(t, 1, len(clones))
		mockRepo.AssertExpectations(t)
		mockWebhook.AssertExpectations(t)
	})

	t.Run("all workflows fail to clone", func(t *testing.T) {
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)
		bulkService := NewBulkService(mockRepo, mockWebhook, logger)

		ctx := context.Background()
		tenantID := "tenant-1"
		userID := "user-1"
		workflowIDs := []string{"workflow-1", "workflow-2"}

		mockRepo.On("GetByID", ctx, tenantID, "workflow-1").Return(nil, ErrNotFound)
		mockRepo.On("GetByID", ctx, tenantID, "workflow-2").Return(nil, errors.New("database error"))

		clones, result := bulkService.BulkClone(ctx, tenantID, userID, workflowIDs)

		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 2, len(result.Failures))
		assert.Equal(t, 0, len(clones))
		mockRepo.AssertExpectations(t)
	})
}

// Test extractWebhookNodes edge cases
func TestBulkService_ExtractWebhookNodes(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockRepo := new(MockBulkRepository)
	bulkService := NewBulkService(mockRepo, nil, logger)

	t.Run("nil definition", func(t *testing.T) {
		nodes := bulkService.extractWebhookNodes(nil)
		assert.Nil(t, nodes)
	})

	t.Run("empty definition", func(t *testing.T) {
		nodes := bulkService.extractWebhookNodes(json.RawMessage(`{}`))
		assert.Nil(t, nodes)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		nodes := bulkService.extractWebhookNodes(json.RawMessage(`invalid json`))
		assert.Nil(t, nodes)
	})

	t.Run("no webhook nodes", func(t *testing.T) {
		definition := json.RawMessage(`{"nodes":[{"id":"node-1","type":"action:http","data":{}}],"edges":[]}`)
		nodes := bulkService.extractWebhookNodes(definition)
		assert.Nil(t, nodes)
	})

	t.Run("single webhook node with default auth", func(t *testing.T) {
		definition := json.RawMessage(`{"nodes":[{"id":"node-1","type":"trigger:webhook","data":{"config":null}}],"edges":[]}`)
		nodes := bulkService.extractWebhookNodes(definition)
		assert.Equal(t, 1, len(nodes))
		assert.Equal(t, "node-1", nodes[0].NodeID)
		assert.Equal(t, AuthTypeSignature, nodes[0].AuthType)
	})

	t.Run("webhook node with custom auth type", func(t *testing.T) {
		definition := json.RawMessage(`{"nodes":[{"id":"node-1","type":"trigger:webhook","data":{"config":{"auth_type":"basic"}}}],"edges":[]}`)
		nodes := bulkService.extractWebhookNodes(definition)
		assert.Equal(t, 1, len(nodes))
		assert.Equal(t, "node-1", nodes[0].NodeID)
		assert.Equal(t, "basic", nodes[0].AuthType)
	})

	t.Run("multiple webhook nodes", func(t *testing.T) {
		definition := json.RawMessage(`{"nodes":[
			{"id":"node-1","type":"trigger:webhook","data":{"config":{"auth_type":"signature"}}},
			{"id":"node-2","type":"action:http","data":{}},
			{"id":"node-3","type":"trigger:webhook","data":{"config":{"auth_type":"api_key"}}}
		],"edges":[]}`)
		nodes := bulkService.extractWebhookNodes(definition)
		assert.Equal(t, 2, len(nodes))
		assert.Equal(t, "node-1", nodes[0].NodeID)
		assert.Equal(t, "node-3", nodes[1].NodeID)
	})

	t.Run("webhook node with invalid config JSON", func(t *testing.T) {
		// Config is present but contains invalid JSON - should use default
		definition := json.RawMessage(`{"nodes":[{"id":"node-1","type":"trigger:webhook","data":{"config":123}}],"edges":[]}`)
		nodes := bulkService.extractWebhookNodes(definition)
		assert.Equal(t, 1, len(nodes))
		assert.Equal(t, AuthTypeSignature, nodes[0].AuthType)
	})

	t.Run("empty nodes array", func(t *testing.T) {
		definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
		nodes := bulkService.extractWebhookNodes(definition)
		assert.Nil(t, nodes)
	})
}

// Test NewBulkService
func TestNewBulkService(t *testing.T) {
	t.Run("creates service with all dependencies", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		mockRepo := new(MockBulkRepository)
		mockWebhook := new(MockWebhookService)

		service := NewBulkService(mockRepo, mockWebhook, logger)

		assert.NotNil(t, service)
		assert.NotNil(t, service.repo)
		assert.NotNil(t, service.webhookService)
		assert.NotNil(t, service.logger)
	})

	t.Run("creates service with nil webhook service", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		mockRepo := new(MockBulkRepository)

		service := NewBulkService(mockRepo, nil, logger)

		assert.NotNil(t, service)
		assert.NotNil(t, service.repo)
		assert.Nil(t, service.webhookService)
	})
}
