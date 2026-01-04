package audit

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

func setupTestDB(t *testing.T) (*sqlx.DB, func()) {
	t.Helper()

	// This will be replaced with actual test DB setup
	// For now, return nil to make the test compile
	return nil, func() {}
}

func TestRepository_CreateAuditEvent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		t.Skip("Database not configured for testing")
	}
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	tenantID := uuid.New().String()
	userID := uuid.New().String()

	event := &AuditEvent{
		TenantID:     tenantID,
		UserID:       userID,
		UserEmail:    "test@example.com",
		Category:     CategoryWorkflow,
		EventType:    EventTypeExecute,
		Action:       "workflow.executed",
		ResourceType: "workflow",
		ResourceID:   "wf-123",
		ResourceName: "Test Workflow",
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0",
		Severity:     SeverityInfo,
		Status:       StatusSuccess,
		Metadata: map[string]interface{}{
			"execution_id": "exec-123",
			"duration_ms":  1000,
		},
	}

	err := repo.CreateAuditEvent(ctx, event)
	require.NoError(t, err)
	assert.NotEmpty(t, event.ID)
	assert.NotZero(t, event.CreatedAt)
}

func TestRepository_CreateAuditEventBatch(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		t.Skip("Database not configured for testing")
	}
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	tenantID := uuid.New().String()
	events := []*AuditEvent{
		{
			TenantID:  tenantID,
			Category:  CategoryWorkflow,
			EventType: EventTypeCreate,
			Action:    "workflow.created",
			Severity:  SeverityInfo,
			Status:    StatusSuccess,
		},
		{
			TenantID:  tenantID,
			Category:  CategoryWorkflow,
			EventType: EventTypeUpdate,
			Action:    "workflow.updated",
			Severity:  SeverityInfo,
			Status:    StatusSuccess,
		},
	}

	err := repo.CreateAuditEventBatch(ctx, events)
	require.NoError(t, err)

	for _, event := range events {
		assert.NotEmpty(t, event.ID)
		assert.NotZero(t, event.CreatedAt)
	}
}

func TestRepository_GetAuditEvent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		t.Skip("Database not configured for testing")
	}
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	tenantID := uuid.New().String()
	event := &AuditEvent{
		TenantID:  tenantID,
		Category:  CategoryWorkflow,
		EventType: EventTypeExecute,
		Action:    "workflow.executed",
		Severity:  SeverityInfo,
		Status:    StatusSuccess,
	}

	err := repo.CreateAuditEvent(ctx, event)
	require.NoError(t, err)

	retrieved, err := repo.GetAuditEvent(ctx, tenantID, event.ID)
	require.NoError(t, err)
	assert.Equal(t, event.ID, retrieved.ID)
	assert.Equal(t, event.TenantID, retrieved.TenantID)
	assert.Equal(t, event.Action, retrieved.Action)
}

func TestRepository_QueryAuditEvents(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		t.Skip("Database not configured for testing")
	}
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	tenantID := uuid.New().String()
	userID := uuid.New().String()

	// Create test events
	events := []*AuditEvent{
		{
			TenantID:  tenantID,
			UserID:    userID,
			Category:  CategoryWorkflow,
			EventType: EventTypeCreate,
			Action:    "workflow.created",
			Severity:  SeverityInfo,
			Status:    StatusSuccess,
		},
		{
			TenantID:  tenantID,
			UserID:    userID,
			Category:  CategoryWorkflow,
			EventType: EventTypeExecute,
			Action:    "workflow.executed",
			Severity:  SeverityInfo,
			Status:    StatusSuccess,
		},
		{
			TenantID:  tenantID,
			UserID:    userID,
			Category:  CategoryCredential,
			EventType: EventTypeAccess,
			Action:    "credential.accessed",
			Severity:  SeverityWarning,
			Status:    StatusSuccess,
		},
	}

	for _, event := range events {
		err := repo.CreateAuditEvent(ctx, event)
		require.NoError(t, err)
	}

	t.Run("filter by category", func(t *testing.T) {
		filter := QueryFilter{
			TenantID:   tenantID,
			Categories: []Category{CategoryWorkflow},
			Limit:      10,
		}

		results, total, err := repo.QueryAuditEvents(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, results, 2)

		for _, result := range results {
			assert.Equal(t, CategoryWorkflow, result.Category)
		}
	})

	t.Run("filter by severity", func(t *testing.T) {
		filter := QueryFilter{
			TenantID:   tenantID,
			Severities: []Severity{SeverityWarning},
			Limit:      10,
		}

		results, total, err := repo.QueryAuditEvents(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, results, 1)
		assert.Equal(t, SeverityWarning, results[0].Severity)
	})

	t.Run("filter by user", func(t *testing.T) {
		filter := QueryFilter{
			TenantID: tenantID,
			UserID:   userID,
			Limit:    10,
		}

		results, total, err := repo.QueryAuditEvents(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, results, 3)

		for _, result := range results {
			assert.Equal(t, userID, result.UserID)
		}
	})
}

func TestRepository_GetAuditStats(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		t.Skip("Database not configured for testing")
	}
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	tenantID := uuid.New().String()
	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now()

	// Create test events
	events := []*AuditEvent{
		{
			TenantID:  tenantID,
			Category:  CategoryWorkflow,
			EventType: EventTypeExecute,
			Action:    "workflow.executed",
			Severity:  SeverityInfo,
			Status:    StatusSuccess,
		},
		{
			TenantID:  tenantID,
			Category:  CategoryWorkflow,
			EventType: EventTypeExecute,
			Action:    "workflow.executed",
			Severity:  SeverityError,
			Status:    StatusFailure,
		},
		{
			TenantID:  tenantID,
			Category:  CategoryCredential,
			EventType: EventTypeAccess,
			Action:    "credential.accessed",
			Severity:  SeverityCritical,
			Status:    StatusSuccess,
		},
	}

	for _, event := range events {
		err := repo.CreateAuditEvent(ctx, event)
		require.NoError(t, err)
	}

	stats, err := repo.GetAuditStats(ctx, tenantID, TimeRange{
		StartDate: startDate,
		EndDate:   endDate,
	})
	require.NoError(t, err)
	assert.Equal(t, 3, stats.TotalEvents)
	assert.Equal(t, 2, stats.EventsByCategory[CategoryWorkflow])
	assert.Equal(t, 1, stats.EventsByCategory[CategoryCredential])
	assert.Equal(t, 1, stats.FailedEvents)
	assert.Equal(t, 1, stats.CriticalEvents)
}

func TestRepository_GetRetentionPolicy(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		t.Skip("Database not configured for testing")
	}
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	tenantID := uuid.New().String()

	policy, err := repo.GetRetentionPolicy(ctx, tenantID)
	require.NoError(t, err)
	assert.NotNil(t, policy)
	assert.Equal(t, tenantID, policy.TenantID)
}

func TestRepository_UpdateRetentionPolicy(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		t.Skip("Database not configured for testing")
	}
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	tenantID := uuid.New().String()

	policy, err := repo.GetRetentionPolicy(ctx, tenantID)
	require.NoError(t, err)

	policy.HotRetentionDays = 180
	policy.ArchiveEnabled = false

	err = repo.UpdateRetentionPolicy(ctx, policy)
	require.NoError(t, err)

	updated, err := repo.GetRetentionPolicy(ctx, tenantID)
	require.NoError(t, err)
	assert.Equal(t, 180, updated.HotRetentionDays)
	assert.False(t, updated.ArchiveEnabled)
}

func TestRepository_DeleteOldAuditEvents(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		t.Skip("Database not configured for testing")
	}
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	tenantID := uuid.New().String()

	// Create an old event
	oldEvent := &AuditEvent{
		TenantID:  tenantID,
		Category:  CategoryWorkflow,
		EventType: EventTypeExecute,
		Action:    "workflow.executed",
		Severity:  SeverityInfo,
		Status:    StatusSuccess,
		CreatedAt: time.Now().Add(-100 * 24 * time.Hour),
	}

	err := repo.CreateAuditEvent(ctx, oldEvent)
	require.NoError(t, err)

	// Delete events older than 90 days
	cutoffDate := time.Now().Add(-90 * 24 * time.Hour)
	deletedCount, err := repo.DeleteOldAuditEvents(ctx, tenantID, cutoffDate)
	require.NoError(t, err)
	assert.Greater(t, deletedCount, int64(0))
}
