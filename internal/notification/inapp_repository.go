package notification

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeInfo    NotificationType = "info"
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeError   NotificationType = "error"
	NotificationTypeSuccess NotificationType = "success"
)

// InAppNotification represents an in-app notification
type InAppNotification struct {
	ID        uuid.UUID              `db:"id" json:"id"`
	TenantID  uuid.UUID              `db:"tenant_id" json:"tenant_id"`
	UserID    string                 `db:"user_id" json:"user_id"`
	Title     string                 `db:"title" json:"title"`
	Message   string                 `db:"message" json:"message"`
	Type      NotificationType       `db:"type" json:"type"`
	Link      string                 `db:"link" json:"link,omitempty"`
	Metadata  map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
	IsRead    bool                   `db:"is_read" json:"is_read"`
	ReadAt    *time.Time             `db:"read_at" json:"read_at,omitempty"`
	CreatedAt time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt time.Time              `db:"updated_at" json:"updated_at"`
}

// InAppRepository handles in-app notification persistence
type InAppRepository struct {
	db *sqlx.DB
}

// NewInAppRepository creates a new in-app notification repository
func NewInAppRepository(db *sqlx.DB) *InAppRepository {
	return &InAppRepository{db: db}
}

// Create creates a new in-app notification
func (r *InAppRepository) Create(ctx context.Context, notif *InAppNotification) error {
	if notif.ID == uuid.Nil {
		notif.ID = uuid.New()
	}

	// Marshal metadata to JSON
	metadataJSON, err := json.Marshal(notif.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO notifications (
			id, tenant_id, user_id, title, message, type, link, metadata, is_read, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
		)
		RETURNING created_at, updated_at
	`

	err = r.db.QueryRowContext(
		ctx, query,
		notif.ID, notif.TenantID, notif.UserID, notif.Title, notif.Message,
		notif.Type, notif.Link, metadataJSON, notif.IsRead,
	).Scan(&notif.CreatedAt, &notif.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// GetByID retrieves a notification by ID
func (r *InAppRepository) GetByID(ctx context.Context, id uuid.UUID) (*InAppNotification, error) {
	query := `
		SELECT id, tenant_id, user_id, title, message, type, link, metadata,
		       is_read, read_at, created_at, updated_at
		FROM notifications
		WHERE id = $1
	`

	var notif InAppNotification
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&notif.ID, &notif.TenantID, &notif.UserID, &notif.Title, &notif.Message,
		&notif.Type, &notif.Link, &metadataJSON, &notif.IsRead, &notif.ReadAt,
		&notif.CreatedAt, &notif.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("notification not found: %s", id)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	// Unmarshal metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &notif.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &notif, nil
}

// ListByUser lists notifications for a user with pagination
func (r *InAppRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*InAppNotification, error) {
	query := `
		SELECT id, tenant_id, user_id, title, message, type, link, metadata,
		       is_read, read_at, created_at, updated_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	return r.scanNotifications(rows)
}

// ListUnread lists unread notifications for a user
func (r *InAppRepository) ListUnread(ctx context.Context, userID string, limit, offset int) ([]*InAppNotification, error) {
	query := `
		SELECT id, tenant_id, user_id, title, message, type, link, metadata,
		       is_read, read_at, created_at, updated_at
		FROM notifications
		WHERE user_id = $1 AND is_read = false
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list unread notifications: %w", err)
	}
	defer rows.Close()

	return r.scanNotifications(rows)
}

// CountUnread counts unread notifications for a user
func (r *InAppRepository) CountUnread(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1 AND is_read = false
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return count, nil
}

// MarkAsRead marks a notification as read
func (r *InAppRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE notifications
		SET is_read = true, read_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND is_read = false
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("notification not found or already read: %s", id)
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (r *InAppRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	query := `
		UPDATE notifications
		SET is_read = true, read_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND is_read = false
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// Delete deletes a notification
func (r *InAppRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("notification not found: %s", id)
	}

	return nil
}

// DeleteOlderThan deletes notifications older than the specified duration
func (r *InAppRepository) DeleteOlderThan(ctx context.Context, duration time.Duration) (int, error) {
	query := `
		DELETE FROM notifications
		WHERE created_at < $1
	`

	cutoffTime := time.Now().Add(-duration)

	result, err := r.db.ExecContext(ctx, query, cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old notifications: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rows), nil
}

// scanNotifications scans multiple notifications from rows
func (r *InAppRepository) scanNotifications(rows *sql.Rows) ([]*InAppNotification, error) {
	notifications := make([]*InAppNotification, 0)

	for rows.Next() {
		var notif InAppNotification
		var metadataJSON []byte

		err := rows.Scan(
			&notif.ID, &notif.TenantID, &notif.UserID, &notif.Title, &notif.Message,
			&notif.Type, &notif.Link, &metadataJSON, &notif.IsRead, &notif.ReadAt,
			&notif.CreatedAt, &notif.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		// Unmarshal metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &notif.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		notifications = append(notifications, &notif)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return notifications, nil
}

// ListByType lists notifications by type for a user
func (r *InAppRepository) ListByType(ctx context.Context, userID string, notifType NotificationType, limit, offset int) ([]*InAppNotification, error) {
	query := `
		SELECT id, tenant_id, user_id, title, message, type, link, metadata,
		       is_read, read_at, created_at, updated_at
		FROM notifications
		WHERE user_id = $1 AND type = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, userID, notifType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications by type: %w", err)
	}
	defer rows.Close()

	return r.scanNotifications(rows)
}

// BulkCreate creates multiple notifications in a single transaction
func (r *InAppRepository) BulkCreate(ctx context.Context, notifications []*InAppNotification) error {
	if len(notifications) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO notifications (
			id, tenant_id, user_id, title, message, type, link, metadata, is_read, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, notif := range notifications {
		if notif.ID == uuid.Nil {
			notif.ID = uuid.New()
		}

		metadataJSON, err := json.Marshal(notif.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		_, err = stmt.ExecContext(
			ctx,
			notif.ID, notif.TenantID, notif.UserID, notif.Title, notif.Message,
			notif.Type, notif.Link, metadataJSON, notif.IsRead,
		)

		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				return fmt.Errorf("database error: %s", pqErr.Message)
			}
			return fmt.Errorf("failed to insert notification: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
