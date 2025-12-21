package aibuilder

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository implements Repository for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateConversation(ctx context.Context, conv *Conversation) error {
	args := m.Called(ctx, conv)
	return args.Error(0)
}

func (m *MockRepository) GetConversation(ctx context.Context, tenantID, conversationID string) (*Conversation, error) {
	args := m.Called(ctx, tenantID, conversationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Conversation), args.Error(1)
}

func (m *MockRepository) UpdateConversation(ctx context.Context, conv *Conversation) error {
	args := m.Called(ctx, conv)
	return args.Error(0)
}

func (m *MockRepository) ListConversations(ctx context.Context, tenantID, userID string, limit, offset int) ([]*Conversation, error) {
	args := m.Called(ctx, tenantID, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Conversation), args.Error(1)
}

func (m *MockRepository) CreateMessage(ctx context.Context, msg *ConversationMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockRepository) GetMessages(ctx context.Context, conversationID string) ([]ConversationMessage, error) {
	args := m.Called(ctx, conversationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ConversationMessage), args.Error(1)
}

func (m *MockRepository) UpdateConversationWorkflow(ctx context.Context, conversationID string, workflow *GeneratedWorkflow) error {
	args := m.Called(ctx, conversationID, workflow)
	return args.Error(0)
}

// MockGenerator implements Generator interface for testing
type MockGenerator struct {
	mock.Mock
}

func (m *MockGenerator) Generate(ctx context.Context, request *BuildRequest, history []ConversationMessage) (*GeneratedWorkflow, string, error) {
	args := m.Called(ctx, request, history)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*GeneratedWorkflow), args.String(1), args.Error(2)
}

func (m *MockGenerator) Refine(ctx context.Context, workflow *GeneratedWorkflow, feedback string, history []ConversationMessage) (*GeneratedWorkflow, string, error) {
	args := m.Called(ctx, workflow, feedback, history)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*GeneratedWorkflow), args.String(1), args.Error(2)
}

// MockWorkflowCreator implements WorkflowCreator interface for testing
type MockWorkflowCreator struct {
	mock.Mock
}

func (m *MockWorkflowCreator) CreateWorkflow(ctx context.Context, tenantID, userID string, workflow *GeneratedWorkflow) (string, error) {
	args := m.Called(ctx, tenantID, userID, workflow)
	return args.String(0), args.Error(1)
}

func TestNewAIBuilderService(t *testing.T) {
	repo := &MockRepository{}
	gen := &MockGenerator{}
	creator := &MockWorkflowCreator{}

	svc := NewAIBuilderService(repo, gen, creator)

	assert.NotNil(t, svc)
}

func TestAIBuilderService_Generate(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		repo := &MockRepository{}
		gen := &MockGenerator{}
		creator := &MockWorkflowCreator{}
		svc := NewAIBuilderService(repo, gen, creator)

		workflow := &GeneratedWorkflow{
			Name: "Test Workflow",
			Definition: &WorkflowDefinition{
				Nodes: []GeneratedNode{
					{ID: "n1", Type: "trigger:webhook", Name: "Trigger"},
				},
			},
		}

		repo.On("CreateConversation", mock.Anything, mock.Anything).Return(nil)
		repo.On("CreateMessage", mock.Anything, mock.Anything).Return(nil)
		repo.On("UpdateConversationWorkflow", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		gen.On("Generate", mock.Anything, mock.Anything, mock.Anything).Return(workflow, "Here's your workflow", nil)

		request := &BuildRequest{
			Description: "Create a webhook workflow",
		}

		result, err := svc.Generate(context.Background(), "tenant-1", "user-1", request)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.ConversationID)
		assert.Equal(t, workflow, result.Workflow)
		assert.Equal(t, "Here's your workflow", result.Explanation)

		repo.AssertExpectations(t)
		gen.AssertExpectations(t)
	})

	t.Run("invalid request", func(t *testing.T) {
		repo := &MockRepository{}
		gen := &MockGenerator{}
		creator := &MockWorkflowCreator{}
		svc := NewAIBuilderService(repo, gen, creator)

		request := &BuildRequest{
			Description: "", // Invalid - empty
		}

		result, err := svc.Generate(context.Background(), "tenant-1", "user-1", request)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("generator error", func(t *testing.T) {
		repo := &MockRepository{}
		gen := &MockGenerator{}
		creator := &MockWorkflowCreator{}
		svc := NewAIBuilderService(repo, gen, creator)

		repo.On("CreateConversation", mock.Anything, mock.Anything).Return(nil)
		repo.On("CreateMessage", mock.Anything, mock.Anything).Return(nil)
		gen.On("Generate", mock.Anything, mock.Anything, mock.Anything).Return(nil, "", errors.New("LLM error"))

		request := &BuildRequest{
			Description: "Create a webhook workflow",
		}

		result, err := svc.Generate(context.Background(), "tenant-1", "user-1", request)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "LLM error")
	})
}

func TestAIBuilderService_Refine(t *testing.T) {
	t.Run("successful refinement", func(t *testing.T) {
		repo := &MockRepository{}
		gen := &MockGenerator{}
		creator := &MockWorkflowCreator{}
		svc := NewAIBuilderService(repo, gen, creator)

		existingWorkflow := &GeneratedWorkflow{
			Name: "Original",
			Definition: &WorkflowDefinition{
				Nodes: []GeneratedNode{
					{ID: "n1", Type: "trigger:webhook", Name: "Trigger"},
				},
			},
		}

		refinedWorkflow := &GeneratedWorkflow{
			Name: "Refined",
			Definition: &WorkflowDefinition{
				Nodes: []GeneratedNode{
					{ID: "n1", Type: "trigger:webhook", Name: "Trigger"},
					{ID: "n2", Type: "action:http", Name: "HTTP"},
				},
			},
		}

		conv := &Conversation{
			ID:              "conv-123",
			TenantID:        "tenant-1",
			UserID:          "user-1",
			Status:          ConversationStatusActive,
			CurrentWorkflow: existingWorkflow,
			Messages:        []ConversationMessage{},
		}

		repo.On("GetConversation", mock.Anything, "tenant-1", "conv-123").Return(conv, nil)
		repo.On("GetMessages", mock.Anything, "conv-123").Return([]ConversationMessage{}, nil)
		repo.On("CreateMessage", mock.Anything, mock.Anything).Return(nil)
		repo.On("UpdateConversationWorkflow", mock.Anything, "conv-123", mock.Anything).Return(nil)
		gen.On("Refine", mock.Anything, existingWorkflow, "Add an HTTP action", mock.Anything).Return(refinedWorkflow, "Added HTTP action", nil)

		request := &RefineRequest{
			ConversationID: "conv-123",
			Message:        "Add an HTTP action",
		}

		result, err := svc.Refine(context.Background(), "tenant-1", request)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "conv-123", result.ConversationID)
		assert.Equal(t, refinedWorkflow, result.Workflow)

		repo.AssertExpectations(t)
		gen.AssertExpectations(t)
	})

	t.Run("conversation not found", func(t *testing.T) {
		repo := &MockRepository{}
		gen := &MockGenerator{}
		creator := &MockWorkflowCreator{}
		svc := NewAIBuilderService(repo, gen, creator)

		repo.On("GetConversation", mock.Anything, "tenant-1", "conv-123").Return(nil, errors.New("not found"))

		request := &RefineRequest{
			ConversationID: "conv-123",
			Message:        "Add an HTTP action",
		}

		result, err := svc.Refine(context.Background(), "tenant-1", request)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("conversation not active", func(t *testing.T) {
		repo := &MockRepository{}
		gen := &MockGenerator{}
		creator := &MockWorkflowCreator{}
		svc := NewAIBuilderService(repo, gen, creator)

		conv := &Conversation{
			ID:       "conv-123",
			TenantID: "tenant-1",
			Status:   ConversationStatusCompleted,
		}

		repo.On("GetConversation", mock.Anything, "tenant-1", "conv-123").Return(conv, nil)

		request := &RefineRequest{
			ConversationID: "conv-123",
			Message:        "Add an HTTP action",
		}

		result, err := svc.Refine(context.Background(), "tenant-1", request)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not active")
	})
}

func TestAIBuilderService_GetConversation(t *testing.T) {
	t.Run("existing conversation", func(t *testing.T) {
		repo := &MockRepository{}
		gen := &MockGenerator{}
		creator := &MockWorkflowCreator{}
		svc := NewAIBuilderService(repo, gen, creator)

		conv := &Conversation{
			ID:       "conv-123",
			TenantID: "tenant-1",
			UserID:   "user-1",
			Status:   ConversationStatusActive,
		}

		repo.On("GetConversation", mock.Anything, "tenant-1", "conv-123").Return(conv, nil)
		repo.On("GetMessages", mock.Anything, "conv-123").Return([]ConversationMessage{}, nil)

		result, err := svc.GetConversation(context.Background(), "tenant-1", "conv-123")

		require.NoError(t, err)
		assert.Equal(t, "conv-123", result.ID)

		repo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		repo := &MockRepository{}
		gen := &MockGenerator{}
		creator := &MockWorkflowCreator{}
		svc := NewAIBuilderService(repo, gen, creator)

		repo.On("GetConversation", mock.Anything, "tenant-1", "conv-123").Return(nil, errors.New("not found"))

		result, err := svc.GetConversation(context.Background(), "tenant-1", "conv-123")

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestAIBuilderService_ListConversations(t *testing.T) {
	repo := &MockRepository{}
	gen := &MockGenerator{}
	creator := &MockWorkflowCreator{}
	svc := NewAIBuilderService(repo, gen, creator)

	convs := []*Conversation{
		{ID: "conv-1", TenantID: "tenant-1", UserID: "user-1"},
		{ID: "conv-2", TenantID: "tenant-1", UserID: "user-1"},
	}

	repo.On("ListConversations", mock.Anything, "tenant-1", "user-1", 20, 0).Return(convs, nil)

	result, err := svc.ListConversations(context.Background(), "tenant-1", "user-1")

	require.NoError(t, err)
	assert.Len(t, result, 2)

	repo.AssertExpectations(t)
}

func TestAIBuilderService_Apply(t *testing.T) {
	t.Run("successful apply", func(t *testing.T) {
		repo := &MockRepository{}
		gen := &MockGenerator{}
		creator := &MockWorkflowCreator{}
		svc := NewAIBuilderService(repo, gen, creator)

		workflow := &GeneratedWorkflow{
			Name: "Test Workflow",
			Definition: &WorkflowDefinition{
				Nodes: []GeneratedNode{
					{ID: "n1", Type: "trigger:webhook", Name: "Trigger"},
				},
			},
		}

		conv := &Conversation{
			ID:              "conv-123",
			TenantID:        "tenant-1",
			UserID:          "user-1",
			Status:          ConversationStatusActive,
			CurrentWorkflow: workflow,
		}

		repo.On("GetConversation", mock.Anything, "tenant-1", "conv-123").Return(conv, nil)
		creator.On("CreateWorkflow", mock.Anything, "tenant-1", "user-1", workflow).Return("workflow-456", nil)
		repo.On("UpdateConversation", mock.Anything, mock.Anything).Return(nil)

		request := &ApplyRequest{
			ConversationID: "conv-123",
		}

		workflowID, err := svc.Apply(context.Background(), "tenant-1", "user-1", request)

		require.NoError(t, err)
		assert.Equal(t, "workflow-456", workflowID)

		repo.AssertExpectations(t)
		creator.AssertExpectations(t)
	})

	t.Run("no workflow to apply", func(t *testing.T) {
		repo := &MockRepository{}
		gen := &MockGenerator{}
		creator := &MockWorkflowCreator{}
		svc := NewAIBuilderService(repo, gen, creator)

		conv := &Conversation{
			ID:              "conv-123",
			TenantID:        "tenant-1",
			Status:          ConversationStatusActive,
			CurrentWorkflow: nil, // No workflow
		}

		repo.On("GetConversation", mock.Anything, "tenant-1", "conv-123").Return(conv, nil)

		request := &ApplyRequest{
			ConversationID: "conv-123",
		}

		workflowID, err := svc.Apply(context.Background(), "tenant-1", "user-1", request)

		require.Error(t, err)
		assert.Empty(t, workflowID)
		assert.Contains(t, err.Error(), "no workflow")
	})
}

func TestAIBuilderService_AbandonConversation(t *testing.T) {
	t.Run("successful abandon", func(t *testing.T) {
		repo := &MockRepository{}
		gen := &MockGenerator{}
		creator := &MockWorkflowCreator{}
		svc := NewAIBuilderService(repo, gen, creator)

		conv := &Conversation{
			ID:       "conv-123",
			TenantID: "tenant-1",
			Status:   ConversationStatusActive,
		}

		repo.On("GetConversation", mock.Anything, "tenant-1", "conv-123").Return(conv, nil)
		repo.On("UpdateConversation", mock.Anything, mock.MatchedBy(func(c *Conversation) bool {
			return c.Status == ConversationStatusAbandoned
		})).Return(nil)

		err := svc.AbandonConversation(context.Background(), "tenant-1", "conv-123")

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("already completed", func(t *testing.T) {
		repo := &MockRepository{}
		gen := &MockGenerator{}
		creator := &MockWorkflowCreator{}
		svc := NewAIBuilderService(repo, gen, creator)

		conv := &Conversation{
			ID:       "conv-123",
			TenantID: "tenant-1",
			Status:   ConversationStatusCompleted,
		}

		repo.On("GetConversation", mock.Anything, "tenant-1", "conv-123").Return(conv, nil)

		err := svc.AbandonConversation(context.Background(), "tenant-1", "conv-123")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not active")
	})
}

// Integration helper tests

func TestContextRepositoryAdapter(t *testing.T) {
	// Test that context.Context can be used as the ctx parameter
	repo := &MockRepository{}

	ctx := context.Background()
	repo.On("GetConversation", ctx, "tenant", "conv").Return(nil, errors.New("test"))

	_, err := repo.GetConversation(ctx, "tenant", "conv")
	assert.Error(t, err)
}
