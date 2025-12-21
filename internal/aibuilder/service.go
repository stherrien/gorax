package aibuilder

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

// ContextRepository adapts Repository interface to use context.Context
type ContextRepository interface {
	CreateConversation(ctx context.Context, conv *Conversation) error
	GetConversation(ctx context.Context, tenantID, conversationID string) (*Conversation, error)
	UpdateConversation(ctx context.Context, conv *Conversation) error
	ListConversations(ctx context.Context, tenantID, userID string, limit, offset int) ([]*Conversation, error)
	CreateMessage(ctx context.Context, msg *ConversationMessage) error
	GetMessages(ctx context.Context, conversationID string) ([]ConversationMessage, error)
	UpdateConversationWorkflow(ctx context.Context, conversationID string, workflow *GeneratedWorkflow) error
}

// ContextGenerator adapts Generator interface to use context.Context
type ContextGenerator interface {
	Generate(ctx context.Context, request *BuildRequest, history []ConversationMessage) (*GeneratedWorkflow, string, error)
	Refine(ctx context.Context, workflow *GeneratedWorkflow, feedback string, history []ConversationMessage) (*GeneratedWorkflow, string, error)
}

// WorkflowCreator creates actual workflows from generated workflows
type WorkflowCreator interface {
	CreateWorkflow(ctx context.Context, tenantID, userID string, workflow *GeneratedWorkflow) (string, error)
}

// AIBuilderService implements the Service interface for AI workflow building
type AIBuilderService struct {
	repo    ContextRepository
	gen     ContextGenerator
	creator WorkflowCreator
	logger  *slog.Logger
}

// NewAIBuilderService creates a new AI builder service
func NewAIBuilderService(repo ContextRepository, gen ContextGenerator, creator WorkflowCreator) *AIBuilderService {
	return &AIBuilderService{
		repo:    repo,
		gen:     gen,
		creator: creator,
		logger:  slog.Default(),
	}
}

// Generate creates a new workflow from a description
func (s *AIBuilderService) Generate(ctx context.Context, tenantID, userID string, request *BuildRequest) (*BuildResult, error) {
	// Validate request
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Create new conversation
	conv := NewConversation(tenantID, userID)
	if err := s.repo.CreateConversation(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Add user message
	userMsg := conv.AddMessage(MessageRoleUser, request.Description)
	userMsg.ConversationID = conv.ID
	if err := s.repo.CreateMessage(ctx, userMsg); err != nil {
		s.logger.Warn("failed to save user message", "error", err)
	}

	// Generate workflow
	workflow, explanation, err := s.gen.Generate(ctx, request, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate workflow: %w", err)
	}

	// Add assistant message
	assistantMsg := conv.AddMessage(MessageRoleAssistant, explanation)
	assistantMsg.ConversationID = conv.ID
	assistantMsg.Workflow = workflow
	if err := s.repo.CreateMessage(ctx, assistantMsg); err != nil {
		s.logger.Warn("failed to save assistant message", "error", err)
	}

	// Update conversation with workflow
	conv.SetCurrentWorkflow(workflow)
	if err := s.repo.UpdateConversationWorkflow(ctx, conv.ID, workflow); err != nil {
		s.logger.Warn("failed to update conversation workflow", "error", err)
	}

	return &BuildResult{
		ConversationID: conv.ID,
		Workflow:       workflow,
		Explanation:    explanation,
		Warnings:       []string{},
		Suggestions:    []string{},
	}, nil
}

// Refine modifies an existing workflow based on feedback
func (s *AIBuilderService) Refine(ctx context.Context, tenantID string, request *RefineRequest) (*BuildResult, error) {
	// Validate request
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Get conversation
	conv, err := s.repo.GetConversation(ctx, tenantID, request.ConversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Check conversation is active
	if !conv.IsActive() {
		return nil, errors.New("conversation is not active")
	}

	// Check there's a workflow to refine
	if conv.CurrentWorkflow == nil {
		return nil, errors.New("no workflow to refine")
	}

	// Get message history
	history, err := s.repo.GetMessages(ctx, conv.ID)
	if err != nil {
		s.logger.Warn("failed to get message history", "error", err)
		history = []ConversationMessage{}
	}

	// Add user message
	userMsg := conv.AddMessage(MessageRoleUser, request.Message)
	userMsg.ConversationID = conv.ID
	if err := s.repo.CreateMessage(ctx, userMsg); err != nil {
		s.logger.Warn("failed to save user message", "error", err)
	}

	// Refine workflow
	refined, explanation, err := s.gen.Refine(ctx, conv.CurrentWorkflow, request.Message, history)
	if err != nil {
		return nil, fmt.Errorf("failed to refine workflow: %w", err)
	}

	// Add assistant message
	assistantMsg := conv.AddMessage(MessageRoleAssistant, explanation)
	assistantMsg.ConversationID = conv.ID
	assistantMsg.Workflow = refined
	if err := s.repo.CreateMessage(ctx, assistantMsg); err != nil {
		s.logger.Warn("failed to save assistant message", "error", err)
	}

	// Update conversation with refined workflow
	conv.SetCurrentWorkflow(refined)
	if err := s.repo.UpdateConversationWorkflow(ctx, conv.ID, refined); err != nil {
		s.logger.Warn("failed to update conversation workflow", "error", err)
	}

	return &BuildResult{
		ConversationID: conv.ID,
		Workflow:       refined,
		Explanation:    explanation,
		Warnings:       []string{},
		Suggestions:    []string{},
	}, nil
}

// GetConversation retrieves a conversation by ID
func (s *AIBuilderService) GetConversation(ctx context.Context, tenantID, conversationID string) (*Conversation, error) {
	conv, err := s.repo.GetConversation(ctx, tenantID, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Load messages
	messages, err := s.repo.GetMessages(ctx, conversationID)
	if err != nil {
		s.logger.Warn("failed to get messages", "error", err)
	} else {
		conv.Messages = messages
	}

	return conv, nil
}

// ListConversations lists all conversations for a tenant/user
func (s *AIBuilderService) ListConversations(ctx context.Context, tenantID, userID string) ([]*Conversation, error) {
	const defaultLimit = 20
	const defaultOffset = 0

	convs, err := s.repo.ListConversations(ctx, tenantID, userID, defaultLimit, defaultOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}

	return convs, nil
}

// Apply creates a real workflow from a generated workflow
func (s *AIBuilderService) Apply(ctx context.Context, tenantID, userID string, request *ApplyRequest) (string, error) {
	// Validate request
	if err := request.Validate(); err != nil {
		return "", fmt.Errorf("invalid request: %w", err)
	}

	// Get conversation
	conv, err := s.repo.GetConversation(ctx, tenantID, request.ConversationID)
	if err != nil {
		return "", fmt.Errorf("failed to get conversation: %w", err)
	}

	// Check there's a workflow to apply
	if conv.CurrentWorkflow == nil {
		return "", errors.New("no workflow to apply")
	}

	// Override workflow name if provided
	workflow := conv.CurrentWorkflow
	if request.WorkflowName != "" {
		workflow.Name = request.WorkflowName
	}

	// Create actual workflow
	workflowID, err := s.creator.CreateWorkflow(ctx, tenantID, userID, workflow)
	if err != nil {
		return "", fmt.Errorf("failed to create workflow: %w", err)
	}

	// Mark conversation as completed
	conv.Complete()
	if err := s.repo.UpdateConversation(ctx, conv); err != nil {
		s.logger.Warn("failed to update conversation status", "error", err)
	}

	return workflowID, nil
}

// AbandonConversation marks a conversation as abandoned
func (s *AIBuilderService) AbandonConversation(ctx context.Context, tenantID, conversationID string) error {
	conv, err := s.repo.GetConversation(ctx, tenantID, conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	if !conv.IsActive() {
		return errors.New("conversation is not active")
	}

	conv.Abandon()
	if err := s.repo.UpdateConversation(ctx, conv); err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	return nil
}
