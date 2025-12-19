package eventtypes

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
)

// EventType represents an event type from the registry
type EventType struct {
	ID          string          `db:"id" json:"id"`
	Name        string          `db:"name" json:"name"`
	Description string          `db:"description" json:"description"`
	Schema      json.RawMessage `db:"schema" json:"schema"`
	Version     int             `db:"version" json:"version"`
	CreatedAt   time.Time       `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updatedAt"`
}

// Repository handles event type data access
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new event type repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// ListAll returns all event types
func (r *Repository) ListAll(ctx context.Context) ([]EventType, error) {
	query := `
		SELECT id, name, description, schema, version, created_at, updated_at
		FROM event_types
		ORDER BY name ASC
	`

	var eventTypes []EventType
	err := r.db.SelectContext(ctx, &eventTypes, query)
	if err != nil {
		return nil, err
	}

	return eventTypes, nil
}
