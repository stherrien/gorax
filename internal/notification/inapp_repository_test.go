package notification

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	// Use test database connection
	// In real tests, you would set up a test database
	db, err := sqlx.Connect("postgres", "host=localhost port=5433 user=postgres password=postgres dbname=gorax_test sslmode=disable")
	if err != nil {
		t.Skip("Database not available:", err)
	}

	// Clean up notifications table
	db.Exec("DELETE FROM notifications")

	return db
}

func setTestTenant(ctx context.Context, db *sqlx.DB, tenantID uuid.UUID, userID string) context.Context {
	db.Exec("SELECT set_config('app.current_tenant_id', $1, false)", tenantID.String())
	db.Exec("SELECT set_config('app.current_user_id', $1, false)", userID)
	return ctx
}

func TestInAppRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewInAppRepository(db)

	tenantID := uuid.New()
	userID := "user-123"
	ctx := setTestTenant(context.Background(), db, tenantID, userID)

	notif := &InAppNotification{
		TenantID: tenantID,
		UserID:   userID,
		Title:    "Test Notification",
		Message:  "This is a test notification",
		Type:     NotificationTypeInfo,
		Link:     "https://example.com/tasks/123",
		Metadata: map[string]interface{}{
			"task_id": "123",
		},
	}

	err := repo.Create(ctx, notif)
	require.NoError(t, err)

	assert.NotEqual(t, uuid.Nil, notif.ID)
	assert.False(t, notif.IsRead)
	assert.NotZero(t, notif.CreatedAt)
	assert.NotZero(t, notif.UpdatedAt)
}

func TestInAppRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewInAppRepository(db)

	tenantID := uuid.New()
	userID := "user-123"
	ctx := setTestTenant(context.Background(), db, tenantID, userID)

	// Create notification
	notif := &InAppNotification{
		TenantID: tenantID,
		UserID:   userID,
		Title:    "Test Notification",
		Message:  "This is a test notification",
		Type:     NotificationTypeInfo,
	}

	err := repo.Create(ctx, notif)
	require.NoError(t, err)

	// Get notification
	retrieved, err := repo.GetByID(ctx, notif.ID)
	require.NoError(t, err)

	assert.Equal(t, notif.ID, retrieved.ID)
	assert.Equal(t, notif.Title, retrieved.Title)
	assert.Equal(t, notif.Message, retrieved.Message)
	assert.Equal(t, notif.Type, retrieved.Type)
}

func TestInAppRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewInAppRepository(db)

	tenantID := uuid.New()
	ctx := setTestTenant(context.Background(), db, tenantID, "user-123")

	_, err := repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
}

func TestInAppRepository_ListByUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewInAppRepository(db)

	tenantID := uuid.New()
	userID := "user-123"
	ctx := setTestTenant(context.Background(), db, tenantID, userID)

	// Create multiple notifications
	for i := 0; i < 5; i++ {
		notif := &InAppNotification{
			TenantID: tenantID,
			UserID:   userID,
			Title:    "Test Notification",
			Message:  "This is a test notification",
			Type:     NotificationTypeInfo,
		}
		err := repo.Create(ctx, notif)
		require.NoError(t, err)
	}

	// List notifications
	notifications, err := repo.ListByUser(ctx, userID, 10, 0)
	require.NoError(t, err)

	assert.Len(t, notifications, 5)
}

func TestInAppRepository_ListByUser_Pagination(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewInAppRepository(db)

	tenantID := uuid.New()
	userID := "user-123"
	ctx := setTestTenant(context.Background(), db, tenantID, userID)

	// Create 10 notifications
	for i := 0; i < 10; i++ {
		notif := &InAppNotification{
			TenantID: tenantID,
			UserID:   userID,
			Title:    "Test Notification",
			Message:  "This is a test notification",
			Type:     NotificationTypeInfo,
		}
		err := repo.Create(ctx, notif)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Get first page
	page1, err := repo.ListByUser(ctx, userID, 5, 0)
	require.NoError(t, err)
	assert.Len(t, page1, 5)

	// Get second page
	page2, err := repo.ListByUser(ctx, userID, 5, 5)
	require.NoError(t, err)
	assert.Len(t, page2, 5)

	// Ensure different notifications
	assert.NotEqual(t, page1[0].ID, page2[0].ID)
}

func TestInAppRepository_ListUnread(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewInAppRepository(db)

	tenantID := uuid.New()
	userID := "user-123"
	ctx := setTestTenant(context.Background(), db, tenantID, userID)

	// Create 3 unread notifications
	for i := 0; i < 3; i++ {
		notif := &InAppNotification{
			TenantID: tenantID,
			UserID:   userID,
			Title:    "Test Notification",
			Message:  "This is a test notification",
			Type:     NotificationTypeInfo,
		}
		err := repo.Create(ctx, notif)
		require.NoError(t, err)
	}

	// Create 2 read notifications
	for i := 0; i < 2; i++ {
		notif := &InAppNotification{
			TenantID: tenantID,
			UserID:   userID,
			Title:    "Test Notification",
			Message:  "This is a test notification",
			Type:     NotificationTypeInfo,
			IsRead:   true,
		}
		err := repo.Create(ctx, notif)
		require.NoError(t, err)
	}

	// List unread
	unread, err := repo.ListUnread(ctx, userID, 10, 0)
	require.NoError(t, err)

	assert.Len(t, unread, 3)
	for _, notif := range unread {
		assert.False(t, notif.IsRead)
	}
}

func TestInAppRepository_CountUnread(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewInAppRepository(db)

	tenantID := uuid.New()
	userID := "user-123"
	ctx := setTestTenant(context.Background(), db, tenantID, userID)

	// Create 5 unread notifications
	for i := 0; i < 5; i++ {
		notif := &InAppNotification{
			TenantID: tenantID,
			UserID:   userID,
			Title:    "Test Notification",
			Message:  "This is a test notification",
			Type:     NotificationTypeInfo,
		}
		err := repo.Create(ctx, notif)
		require.NoError(t, err)
	}

	// Count unread
	count, err := repo.CountUnread(ctx, userID)
	require.NoError(t, err)

	assert.Equal(t, 5, count)
}

func TestInAppRepository_MarkAsRead(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewInAppRepository(db)

	tenantID := uuid.New()
	userID := "user-123"
	ctx := setTestTenant(context.Background(), db, tenantID, userID)

	// Create notification
	notif := &InAppNotification{
		TenantID: tenantID,
		UserID:   userID,
		Title:    "Test Notification",
		Message:  "This is a test notification",
		Type:     NotificationTypeInfo,
	}

	err := repo.Create(ctx, notif)
	require.NoError(t, err)
	assert.False(t, notif.IsRead)

	// Mark as read
	err = repo.MarkAsRead(ctx, notif.ID)
	require.NoError(t, err)

	// Verify
	retrieved, err := repo.GetByID(ctx, notif.ID)
	require.NoError(t, err)

	assert.True(t, retrieved.IsRead)
	assert.NotNil(t, retrieved.ReadAt)
}

func TestInAppRepository_MarkAllAsRead(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewInAppRepository(db)

	tenantID := uuid.New()
	userID := "user-123"
	ctx := setTestTenant(context.Background(), db, tenantID, userID)

	// Create multiple unread notifications
	for i := 0; i < 5; i++ {
		notif := &InAppNotification{
			TenantID: tenantID,
			UserID:   userID,
			Title:    "Test Notification",
			Message:  "This is a test notification",
			Type:     NotificationTypeInfo,
		}
		err := repo.Create(ctx, notif)
		require.NoError(t, err)
	}

	// Mark all as read
	err := repo.MarkAllAsRead(ctx, userID)
	require.NoError(t, err)

	// Verify
	count, err := repo.CountUnread(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestInAppRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewInAppRepository(db)

	tenantID := uuid.New()
	userID := "user-123"
	ctx := setTestTenant(context.Background(), db, tenantID, userID)

	// Create notification
	notif := &InAppNotification{
		TenantID: tenantID,
		UserID:   userID,
		Title:    "Test Notification",
		Message:  "This is a test notification",
		Type:     NotificationTypeInfo,
	}

	err := repo.Create(ctx, notif)
	require.NoError(t, err)

	// Delete
	err = repo.Delete(ctx, notif.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(ctx, notif.ID)
	assert.Error(t, err)
}

func TestInAppRepository_DeleteOlderThan(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewInAppRepository(db)

	tenantID := uuid.New()
	userID := "user-123"
	ctx := setTestTenant(context.Background(), db, tenantID, userID)

	// Create old notification (simulate by updating created_at)
	notif := &InAppNotification{
		TenantID: tenantID,
		UserID:   userID,
		Title:    "Old Notification",
		Message:  "This is an old notification",
		Type:     NotificationTypeInfo,
	}

	err := repo.Create(ctx, notif)
	require.NoError(t, err)

	// Manually update created_at to 31 days ago
	_, err = db.Exec("UPDATE notifications SET created_at = $1 WHERE id = $2",
		time.Now().Add(-31*24*time.Hour), notif.ID)
	require.NoError(t, err)

	// Create new notification
	newNotif := &InAppNotification{
		TenantID: tenantID,
		UserID:   userID,
		Title:    "New Notification",
		Message:  "This is a new notification",
		Type:     NotificationTypeInfo,
	}

	err = repo.Create(ctx, newNotif)
	require.NoError(t, err)

	// Delete notifications older than 30 days
	deleted, err := repo.DeleteOlderThan(ctx, 30*24*time.Hour)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)

	// Verify old notification is deleted
	_, err = repo.GetByID(ctx, notif.ID)
	assert.Error(t, err)

	// Verify new notification still exists
	_, err = repo.GetByID(ctx, newNotif.ID)
	require.NoError(t, err)
}

func TestInAppRepository_TenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewInAppRepository(db)

	tenant1ID := uuid.New()
	tenant2ID := uuid.New()
	userID := "user-123"

	// Create notification for tenant 1
	ctx1 := setTestTenant(context.Background(), db, tenant1ID, userID)
	notif1 := &InAppNotification{
		TenantID: tenant1ID,
		UserID:   userID,
		Title:    "Tenant 1 Notification",
		Message:  "This is a tenant 1 notification",
		Type:     NotificationTypeInfo,
	}
	err := repo.Create(ctx1, notif1)
	require.NoError(t, err)

	// Try to access from tenant 2 context
	ctx2 := setTestTenant(context.Background(), db, tenant2ID, userID)
	_, err = repo.GetByID(ctx2, notif1.ID)
	assert.Error(t, err, "Should not be able to access notification from different tenant")

	// List should be empty for tenant 2
	notifications, err := repo.ListByUser(ctx2, userID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, notifications, 0)
}
