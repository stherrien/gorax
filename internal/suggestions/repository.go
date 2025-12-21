package suggestions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	// ErrSuggestionNotFound is returned when a suggestion is not found
	ErrSuggestionNotFound = errors.New("suggestion not found")
)

// PostgresRepository implements Repository for PostgreSQL
type PostgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// suggestionRow is the database representation of a suggestion
type suggestionRow struct {
	ID          string    `db:"id"`
	TenantID    string    `db:"tenant_id"`
	ExecutionID string    `db:"execution_id"`
	NodeID      string    `db:"node_id"`
	Category    string    `db:"category"`
	Type        string    `db:"type"`
	Confidence  string    `db:"confidence"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	Details     *string   `db:"details"`
	Fix         []byte    `db:"fix"`
	Source      string    `db:"source"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	AppliedAt   *time.Time `db:"applied_at"`
	DismissedAt *time.Time `db:"dismissed_at"`
}

func (r *suggestionRow) toSuggestion() (*Suggestion, error) {
	s := &Suggestion{
		ID:          r.ID,
		TenantID:    r.TenantID,
		ExecutionID: r.ExecutionID,
		NodeID:      r.NodeID,
		Category:    ErrorCategory(r.Category),
		Type:        SuggestionType(r.Type),
		Confidence:  SuggestionConfidence(r.Confidence),
		Title:       r.Title,
		Description: r.Description,
		Source:      SuggestionSource(r.Source),
		Status:      SuggestionStatus(r.Status),
		CreatedAt:   r.CreatedAt,
		AppliedAt:   r.AppliedAt,
		DismissedAt: r.DismissedAt,
	}

	if r.Details != nil {
		s.Details = *r.Details
	}

	if len(r.Fix) > 0 {
		var fix SuggestionFix
		if err := json.Unmarshal(r.Fix, &fix); err != nil {
			return nil, err
		}
		s.Fix = &fix
	}

	return s, nil
}

// Create creates a new suggestion
func (r *PostgresRepository) Create(ctx context.Context, suggestion *Suggestion) error {
	var fixJSON []byte
	var err error
	if suggestion.Fix != nil {
		fixJSON, err = json.Marshal(suggestion.Fix)
		if err != nil {
			return err
		}
	}

	query := `
		INSERT INTO execution_suggestions (
			id, tenant_id, execution_id, node_id,
			category, type, confidence,
			title, description, details, fix,
			source, status, created_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9, $10, $11,
			$12, $13, $14
		)
	`

	var details *string
	if suggestion.Details != "" {
		details = &suggestion.Details
	}

	_, err = r.db.ExecContext(ctx, query,
		suggestion.ID, suggestion.TenantID, suggestion.ExecutionID, suggestion.NodeID,
		string(suggestion.Category), string(suggestion.Type), string(suggestion.Confidence),
		suggestion.Title, suggestion.Description, details, fixJSON,
		string(suggestion.Source), string(suggestion.Status), suggestion.CreatedAt,
	)

	return err
}

// CreateBatch creates multiple suggestions in a single transaction
func (r *PostgresRepository) CreateBatch(ctx context.Context, suggestions []*Suggestion) error {
	if len(suggestions) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO execution_suggestions (
			id, tenant_id, execution_id, node_id,
			category, type, confidence,
			title, description, details, fix,
			source, status, created_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9, $10, $11,
			$12, $13, $14
		)
	`

	stmt, err := tx.PreparexContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, suggestion := range suggestions {
		var fixJSON []byte
		if suggestion.Fix != nil {
			fixJSON, err = json.Marshal(suggestion.Fix)
			if err != nil {
				return err
			}
		}

		var details *string
		if suggestion.Details != "" {
			details = &suggestion.Details
		}

		_, err = stmt.ExecContext(ctx,
			suggestion.ID, suggestion.TenantID, suggestion.ExecutionID, suggestion.NodeID,
			string(suggestion.Category), string(suggestion.Type), string(suggestion.Confidence),
			suggestion.Title, suggestion.Description, details, fixJSON,
			string(suggestion.Source), string(suggestion.Status), suggestion.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetByID retrieves a suggestion by ID
func (r *PostgresRepository) GetByID(ctx context.Context, tenantID, id string) (*Suggestion, error) {
	query := `
		SELECT id, tenant_id, execution_id, node_id,
			   category, type, confidence,
			   title, description, details, fix,
			   source, status, created_at, applied_at, dismissed_at
		FROM execution_suggestions
		WHERE id = $1 AND tenant_id = $2
	`

	var row suggestionRow
	err := r.db.GetContext(ctx, &row, query, id, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSuggestionNotFound
		}
		return nil, err
	}

	return row.toSuggestion()
}

// GetByExecutionID retrieves all suggestions for an execution
func (r *PostgresRepository) GetByExecutionID(ctx context.Context, tenantID, executionID string) ([]*Suggestion, error) {
	query := `
		SELECT id, tenant_id, execution_id, node_id,
			   category, type, confidence,
			   title, description, details, fix,
			   source, status, created_at, applied_at, dismissed_at
		FROM execution_suggestions
		WHERE tenant_id = $1 AND execution_id = $2
		ORDER BY
			CASE confidence
				WHEN 'high' THEN 1
				WHEN 'medium' THEN 2
				WHEN 'low' THEN 3
			END,
			created_at DESC
	`

	var rows []suggestionRow
	err := r.db.SelectContext(ctx, &rows, query, tenantID, executionID)
	if err != nil {
		return nil, err
	}

	suggestions := make([]*Suggestion, 0, len(rows))
	for _, row := range rows {
		s, err := row.toSuggestion()
		if err != nil {
			return nil, err
		}
		suggestions = append(suggestions, s)
	}

	return suggestions, nil
}

// UpdateStatus updates the status of a suggestion
func (r *PostgresRepository) UpdateStatus(ctx context.Context, tenantID, id string, status SuggestionStatus) error {
	var query string
	var args []interface{}

	now := time.Now()

	switch status {
	case StatusApplied:
		query = `
			UPDATE execution_suggestions
			SET status = $3, applied_at = $4
			WHERE id = $1 AND tenant_id = $2
		`
		args = []interface{}{id, tenantID, string(status), now}
	case StatusDismissed:
		query = `
			UPDATE execution_suggestions
			SET status = $3, dismissed_at = $4
			WHERE id = $1 AND tenant_id = $2
		`
		args = []interface{}{id, tenantID, string(status), now}
	default:
		query = `
			UPDATE execution_suggestions
			SET status = $3
			WHERE id = $1 AND tenant_id = $2
		`
		args = []interface{}{id, tenantID, string(status)}
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrSuggestionNotFound
	}

	return nil
}

// Delete deletes a suggestion
func (r *PostgresRepository) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM execution_suggestions WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, tenantID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrSuggestionNotFound
	}

	return nil
}

// DeleteByExecutionID deletes all suggestions for an execution
func (r *PostgresRepository) DeleteByExecutionID(ctx context.Context, tenantID, executionID string) error {
	query := `DELETE FROM execution_suggestions WHERE tenant_id = $1 AND execution_id = $2`
	_, err := r.db.ExecContext(ctx, query, tenantID, executionID)
	return err
}

// ErrorPatternRow is the database representation of an error pattern
type ErrorPatternRow struct {
	ID                    string    `db:"id"`
	TenantID              *string   `db:"tenant_id"`
	Name                  string    `db:"name"`
	Description           *string   `db:"description"`
	Category              string    `db:"category"`
	Patterns              []byte    `db:"patterns"`
	HTTPCodes             []byte    `db:"http_codes"`
	NodeTypes             []byte    `db:"node_types"`
	SuggestionType        string    `db:"suggestion_type"`
	SuggestionTitle       string    `db:"suggestion_title"`
	SuggestionDescription string    `db:"suggestion_description"`
	SuggestionConfidence  string    `db:"suggestion_confidence"`
	FixTemplate           []byte    `db:"fix_template"`
	Priority              int       `db:"priority"`
	IsActive              bool      `db:"is_active"`
	CreatedAt             time.Time `db:"created_at"`
	UpdatedAt             time.Time `db:"updated_at"`
}

// ErrorPattern represents an error pattern for matching
type ErrorPattern struct {
	ID                    string
	TenantID              *string
	Name                  string
	Description           string
	Category              ErrorCategory
	Patterns              []string
	HTTPCodes             []int
	NodeTypes             []string
	SuggestionType        SuggestionType
	SuggestionTitle       string
	SuggestionDescription string
	SuggestionConfidence  SuggestionConfidence
	FixTemplate           *SuggestionFix
	Priority              int
	IsActive              bool
}

// GetActivePatterns retrieves all active error patterns
func (r *PostgresRepository) GetActivePatterns(ctx context.Context, tenantID string) ([]*ErrorPattern, error) {
	query := `
		SELECT id, tenant_id, name, description, category,
			   patterns, http_codes, node_types,
			   suggestion_type, suggestion_title, suggestion_description, suggestion_confidence,
			   fix_template, priority, is_active, created_at, updated_at
		FROM error_patterns
		WHERE is_active = true AND (tenant_id IS NULL OR tenant_id = $1)
		ORDER BY priority DESC, created_at ASC
	`

	var rows []ErrorPatternRow
	err := r.db.SelectContext(ctx, &rows, query, tenantID)
	if err != nil {
		return nil, err
	}

	patterns := make([]*ErrorPattern, 0, len(rows))
	for _, row := range rows {
		p, err := rowToErrorPattern(&row)
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, p)
	}

	return patterns, nil
}

func rowToErrorPattern(row *ErrorPatternRow) (*ErrorPattern, error) {
	p := &ErrorPattern{
		ID:                    row.ID,
		TenantID:              row.TenantID,
		Name:                  row.Name,
		Category:              ErrorCategory(row.Category),
		SuggestionType:        SuggestionType(row.SuggestionType),
		SuggestionTitle:       row.SuggestionTitle,
		SuggestionDescription: row.SuggestionDescription,
		SuggestionConfidence:  SuggestionConfidence(row.SuggestionConfidence),
		Priority:              row.Priority,
		IsActive:              row.IsActive,
	}

	if row.Description != nil {
		p.Description = *row.Description
	}

	if len(row.Patterns) > 0 {
		if err := json.Unmarshal(row.Patterns, &p.Patterns); err != nil {
			return nil, err
		}
	}

	if len(row.HTTPCodes) > 0 {
		if err := json.Unmarshal(row.HTTPCodes, &p.HTTPCodes); err != nil {
			return nil, err
		}
	}

	if len(row.NodeTypes) > 0 {
		if err := json.Unmarshal(row.NodeTypes, &p.NodeTypes); err != nil {
			return nil, err
		}
	}

	if len(row.FixTemplate) > 0 {
		var fix SuggestionFix
		if err := json.Unmarshal(row.FixTemplate, &fix); err != nil {
			return nil, err
		}
		p.FixTemplate = &fix
	}

	return p, nil
}
