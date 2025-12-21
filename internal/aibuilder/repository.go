package aibuilder

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	// ErrConversationNotFound is returned when a conversation is not found
	ErrConversationNotFound = errors.New("conversation not found")
)

// PostgresRepository implements the repository interface for PostgreSQL
type PostgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// conversationRow is the database representation of a conversation
type conversationRow struct {
	ID              string    `db:"id"`
	TenantID        string    `db:"tenant_id"`
	UserID          string    `db:"user_id"`
	Status          string    `db:"status"`
	CurrentWorkflow []byte    `db:"current_workflow"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// messageRow is the database representation of a message
type messageRow struct {
	ID               string    `db:"id"`
	ConversationID   string    `db:"conversation_id"`
	Role             string    `db:"role"`
	Content          string    `db:"content"`
	Workflow         []byte    `db:"workflow"`
	PromptTokens     *int      `db:"prompt_tokens"`
	CompletionTokens *int      `db:"completion_tokens"`
	CreatedAt        time.Time `db:"created_at"`
}

func (r *conversationRow) toConversation() (*Conversation, error) {
	conv := &Conversation{
		ID:        r.ID,
		TenantID:  r.TenantID,
		UserID:    r.UserID,
		Status:    ConversationStatus(r.Status),
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}

	if len(r.CurrentWorkflow) > 0 {
		var workflow GeneratedWorkflow
		if err := json.Unmarshal(r.CurrentWorkflow, &workflow); err != nil {
			return nil, err
		}
		conv.CurrentWorkflow = &workflow
	}

	return conv, nil
}

func (r *messageRow) toMessage() (*ConversationMessage, error) {
	msg := &ConversationMessage{
		ID:             r.ID,
		ConversationID: r.ConversationID,
		Role:           MessageRole(r.Role),
		Content:        r.Content,
		CreatedAt:      r.CreatedAt,
	}

	if r.PromptTokens != nil {
		msg.PromptTokens = *r.PromptTokens
	}
	if r.CompletionTokens != nil {
		msg.CompletionTokens = *r.CompletionTokens
	}

	if len(r.Workflow) > 0 {
		var workflow GeneratedWorkflow
		if err := json.Unmarshal(r.Workflow, &workflow); err != nil {
			return nil, err
		}
		msg.Workflow = &workflow
	}

	return msg, nil
}

// CreateConversation creates a new conversation
func (r *PostgresRepository) CreateConversation(ctx context.Context, conv *Conversation) error {
	var workflowJSON []byte
	var err error
	if conv.CurrentWorkflow != nil {
		workflowJSON, err = json.Marshal(conv.CurrentWorkflow)
		if err != nil {
			return err
		}
	}

	query := `
		INSERT INTO aibuilder_conversations (
			id, tenant_id, user_id, status, current_workflow, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.db.ExecContext(ctx, query,
		conv.ID, conv.TenantID, conv.UserID, string(conv.Status),
		workflowJSON, conv.CreatedAt, conv.UpdatedAt,
	)

	return err
}

// GetConversation retrieves a conversation by ID
func (r *PostgresRepository) GetConversation(ctx context.Context, tenantID, conversationID string) (*Conversation, error) {
	query := `
		SELECT id, tenant_id, user_id, status, current_workflow, created_at, updated_at
		FROM aibuilder_conversations
		WHERE id = $1 AND tenant_id = $2
	`

	var row conversationRow
	err := r.db.GetContext(ctx, &row, query, conversationID, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrConversationNotFound
		}
		return nil, err
	}

	conv, err := row.toConversation()
	if err != nil {
		return nil, err
	}

	// Load messages
	messages, err := r.GetMessages(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	conv.Messages = messages

	return conv, nil
}

// UpdateConversation updates a conversation
func (r *PostgresRepository) UpdateConversation(ctx context.Context, conv *Conversation) error {
	var workflowJSON []byte
	var err error
	if conv.CurrentWorkflow != nil {
		workflowJSON, err = json.Marshal(conv.CurrentWorkflow)
		if err != nil {
			return err
		}
	}

	query := `
		UPDATE aibuilder_conversations
		SET status = $3, current_workflow = $4, updated_at = $5
		WHERE id = $1 AND tenant_id = $2
	`

	result, err := r.db.ExecContext(ctx, query,
		conv.ID, conv.TenantID, string(conv.Status), workflowJSON, time.Now(),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrConversationNotFound
	}

	return nil
}

// ListConversations lists conversations for a user
func (r *PostgresRepository) ListConversations(ctx context.Context, tenantID, userID string, limit, offset int) ([]*Conversation, error) {
	query := `
		SELECT id, tenant_id, user_id, status, current_workflow, created_at, updated_at
		FROM aibuilder_conversations
		WHERE tenant_id = $1 AND user_id = $2
		ORDER BY updated_at DESC
		LIMIT $3 OFFSET $4
	`

	var rows []conversationRow
	err := r.db.SelectContext(ctx, &rows, query, tenantID, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	conversations := make([]*Conversation, 0, len(rows))
	for _, row := range rows {
		conv, err := row.toConversation()
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// CreateMessage creates a new message
func (r *PostgresRepository) CreateMessage(ctx context.Context, msg *ConversationMessage) error {
	var workflowJSON []byte
	var err error
	if msg.Workflow != nil {
		workflowJSON, err = json.Marshal(msg.Workflow)
		if err != nil {
			return err
		}
	}

	var promptTokens, completionTokens *int
	if msg.PromptTokens > 0 {
		promptTokens = &msg.PromptTokens
	}
	if msg.CompletionTokens > 0 {
		completionTokens = &msg.CompletionTokens
	}

	query := `
		INSERT INTO aibuilder_messages (
			id, conversation_id, role, content, workflow, prompt_tokens, completion_tokens, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.ExecContext(ctx, query,
		msg.ID, msg.ConversationID, string(msg.Role), msg.Content,
		workflowJSON, promptTokens, completionTokens, msg.CreatedAt,
	)

	return err
}

// GetMessages retrieves all messages for a conversation
func (r *PostgresRepository) GetMessages(ctx context.Context, conversationID string) ([]ConversationMessage, error) {
	query := `
		SELECT id, conversation_id, role, content, workflow, prompt_tokens, completion_tokens, created_at
		FROM aibuilder_messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
	`

	var rows []messageRow
	err := r.db.SelectContext(ctx, &rows, query, conversationID)
	if err != nil {
		return nil, err
	}

	messages := make([]ConversationMessage, 0, len(rows))
	for _, row := range rows {
		msg, err := row.toMessage()
		if err != nil {
			return nil, err
		}
		messages = append(messages, *msg)
	}

	return messages, nil
}

// UpdateConversationWorkflow updates only the current workflow
func (r *PostgresRepository) UpdateConversationWorkflow(ctx context.Context, conversationID string, workflow *GeneratedWorkflow) error {
	var workflowJSON []byte
	var err error
	if workflow != nil {
		workflowJSON, err = json.Marshal(workflow)
		if err != nil {
			return err
		}
	}

	query := `
		UPDATE aibuilder_conversations
		SET current_workflow = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, conversationID, workflowJSON, time.Now())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrConversationNotFound
	}

	return nil
}

// ContextRepositoryAdapter adapts PostgresRepository to ContextRepository interface
type ContextRepositoryAdapter struct {
	repo *PostgresRepository
}

// NewContextRepositoryAdapter creates a new adapter
func NewContextRepositoryAdapter(repo *PostgresRepository) *ContextRepositoryAdapter {
	return &ContextRepositoryAdapter{repo: repo}
}

func (a *ContextRepositoryAdapter) CreateConversation(ctx context.Context, conv *Conversation) error {
	return a.repo.CreateConversation(ctx, conv)
}

func (a *ContextRepositoryAdapter) GetConversation(ctx context.Context, tenantID, conversationID string) (*Conversation, error) {
	return a.repo.GetConversation(ctx, tenantID, conversationID)
}

func (a *ContextRepositoryAdapter) UpdateConversation(ctx context.Context, conv *Conversation) error {
	return a.repo.UpdateConversation(ctx, conv)
}

func (a *ContextRepositoryAdapter) ListConversations(ctx context.Context, tenantID, userID string, limit, offset int) ([]*Conversation, error) {
	return a.repo.ListConversations(ctx, tenantID, userID, limit, offset)
}

func (a *ContextRepositoryAdapter) CreateMessage(ctx context.Context, msg *ConversationMessage) error {
	return a.repo.CreateMessage(ctx, msg)
}

func (a *ContextRepositoryAdapter) GetMessages(ctx context.Context, conversationID string) ([]ConversationMessage, error) {
	return a.repo.GetMessages(ctx, conversationID)
}

func (a *ContextRepositoryAdapter) UpdateConversationWorkflow(ctx context.Context, conversationID string, workflow *GeneratedWorkflow) error {
	return a.repo.UpdateConversationWorkflow(ctx, conversationID, workflow)
}
