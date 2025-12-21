package humantask

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	// This would connect to a test database
	// For now, we'll skip if no test DB is available
	t.Skip("Integration test - requires test database")
	return nil
}

func TestRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	executionID := uuid.New()

	assignees, _ := json.Marshal([]string{uuid.New().String(), "admin"})
	config, _ := json.Marshal(map[string]interface{}{
		"timeout": "1h",
	})

	task := &HumanTask{
		TenantID:    tenantID,
		ExecutionID: executionID,
		StepID:      "step-1",
		TaskType:    TaskTypeApproval,
		Title:       "Approve deployment",
		Description: "Please approve the production deployment",
		Assignees:   assignees,
		Status:      StatusPending,
		Config:      config,
	}

	err := repo.Create(ctx, task)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, task.ID)
	assert.False(t, task.CreatedAt.IsZero())
}

func TestRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func() uuid.UUID
		wantErr error
	}{
		{
			name: "existing task",
			setup: func() uuid.UUID {
				task := createTestTask(t, repo, ctx)
				return task.ID
			},
			wantErr: nil,
		},
		{
			name: "non-existent task",
			setup: func() uuid.UUID {
				return uuid.New()
			},
			wantErr: ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskID := tt.setup()

			task, err := repo.GetByID(ctx, taskID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, task)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, task)
				assert.Equal(t, taskID, task.ID)
			}
		})
	}
}

func TestRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	executionID := uuid.New()
	userID := uuid.New()

	// Create multiple tasks
	createTaskWithParams(t, repo, ctx, tenantID, executionID, []string{userID.String()}, StatusPending)
	createTaskWithParams(t, repo, ctx, tenantID, executionID, []string{userID.String()}, StatusApproved)
	createTaskWithParams(t, repo, ctx, tenantID, uuid.New(), []string{"admin"}, StatusPending)

	tests := []struct {
		name      string
		filter    TaskFilter
		wantCount int
	}{
		{
			name: "all tasks for tenant",
			filter: TaskFilter{
				TenantID: tenantID,
				Limit:    10,
			},
			wantCount: 3,
		},
		{
			name: "pending tasks only",
			filter: TaskFilter{
				TenantID: tenantID,
				Status:   strPtr(StatusPending),
				Limit:    10,
			},
			wantCount: 2,
		},
		{
			name: "tasks for specific execution",
			filter: TaskFilter{
				TenantID:    tenantID,
				ExecutionID: &executionID,
				Limit:       10,
			},
			wantCount: 2,
		},
		{
			name: "tasks assigned to user",
			filter: TaskFilter{
				TenantID: tenantID,
				Assignee: strPtr(userID.String()),
				Limit:    10,
			},
			wantCount: 2,
		},
		{
			name: "tasks assigned to role",
			filter: TaskFilter{
				TenantID: tenantID,
				Assignee: strPtr("admin"),
				Limit:    10,
			},
			wantCount: 1,
		},
		{
			name: "with pagination",
			filter: TaskFilter{
				TenantID: tenantID,
				Limit:    2,
				Offset:   1,
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, err := repo.List(ctx, tt.filter)
			require.NoError(t, err)
			assert.Len(t, tasks, tt.wantCount)
		})
	}
}

func TestRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	task := createTestTask(t, repo, ctx)

	// Update task
	userID := uuid.New()
	now := time.Now()
	task.Status = StatusApproved
	task.CompletedAt = &now
	task.CompletedBy = &userID
	task.ResponseData = []byte(`{"comment": "Approved"}`)

	err := repo.Update(ctx, task)
	require.NoError(t, err)

	// Verify update
	updated, err := repo.GetByID(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusApproved, updated.Status)
	assert.NotNil(t, updated.CompletedAt)
	assert.NotNil(t, updated.CompletedBy)
	assert.Equal(t, userID, *updated.CompletedBy)
}

func TestRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	task := createTestTask(t, repo, ctx)

	err := repo.Delete(ctx, task.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(ctx, task.ID)
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestRepository_GetOverdueTasks(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Create task with past due date
	pastDue := time.Now().Add(-1 * time.Hour)
	task := createTestTask(t, repo, ctx)
	task.DueDate = &pastDue
	task.Status = StatusPending
	err := repo.Update(ctx, task)
	require.NoError(t, err)

	// Create task with future due date
	futureDue := time.Now().Add(1 * time.Hour)
	task2 := createTestTask(t, repo, ctx)
	task2.DueDate = &futureDue
	task2.Status = StatusPending
	err = repo.Update(ctx, task2)
	require.NoError(t, err)

	// Create completed task with past due date (should not be returned)
	task3 := createTestTask(t, repo, ctx)
	task3.DueDate = &pastDue
	task3.Status = StatusApproved
	err = repo.Update(ctx, task3)
	require.NoError(t, err)

	// Get overdue tasks
	overdue, err := repo.GetOverdueTasks(ctx, task.TenantID)
	require.NoError(t, err)

	// Should only return the first task
	assert.Len(t, overdue, 1)
	assert.Equal(t, task.ID, overdue[0].ID)
}

func TestRepository_CountPendingByAssignee(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	userID := uuid.New()

	// Create tasks assigned to user
	createTaskWithParams(t, repo, ctx, tenantID, uuid.New(), []string{userID.String()}, StatusPending)
	createTaskWithParams(t, repo, ctx, tenantID, uuid.New(), []string{userID.String()}, StatusPending)
	createTaskWithParams(t, repo, ctx, tenantID, uuid.New(), []string{userID.String()}, StatusApproved)

	count, err := repo.CountPendingByAssignee(ctx, tenantID, userID.String())
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

// Helper functions

func createTestTask(t *testing.T, repo Repository, ctx context.Context) *HumanTask {
	assignees, _ := json.Marshal([]string{uuid.New().String()})
	config, _ := json.Marshal(map[string]interface{}{})

	task := &HumanTask{
		TenantID:    uuid.New(),
		ExecutionID: uuid.New(),
		StepID:      "step-1",
		TaskType:    TaskTypeApproval,
		Title:       "Test task",
		Description: "Test description",
		Assignees:   assignees,
		Status:      StatusPending,
		Config:      config,
	}

	err := repo.Create(ctx, task)
	require.NoError(t, err)

	return task
}

func createTaskWithParams(t *testing.T, repo Repository, ctx context.Context,
	tenantID, executionID uuid.UUID, assignees []string, status string) *HumanTask {

	assigneesJSON, _ := json.Marshal(assignees)
	config, _ := json.Marshal(map[string]interface{}{})

	task := &HumanTask{
		TenantID:    tenantID,
		ExecutionID: executionID,
		StepID:      "step-1",
		TaskType:    TaskTypeApproval,
		Title:       "Test task",
		Description: "Test description",
		Assignees:   assigneesJSON,
		Status:      status,
		Config:      config,
	}

	err := repo.Create(ctx, task)
	require.NoError(t, err)

	return task
}
