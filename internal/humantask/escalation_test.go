package humantask

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockEscalationRepository extends MockRepository with escalation methods
type MockEscalationRepository struct {
	mock.Mock
}

func (m *MockEscalationRepository) Create(ctx context.Context, task *HumanTask) error {
	args := m.Called(ctx, task)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	task.ID = uuid.New()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	return nil
}

func (m *MockEscalationRepository) GetByID(ctx context.Context, id uuid.UUID) (*HumanTask, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*HumanTask), args.Error(1)
}

func (m *MockEscalationRepository) List(ctx context.Context, filter TaskFilter) ([]*HumanTask, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*HumanTask), args.Error(1)
}

func (m *MockEscalationRepository) Update(ctx context.Context, task *HumanTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockEscalationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEscalationRepository) GetOverdueTasks(ctx context.Context, tenantID uuid.UUID) ([]*HumanTask, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*HumanTask), args.Error(1)
}

func (m *MockEscalationRepository) CountPendingByAssignee(ctx context.Context, tenantID uuid.UUID, assignee string) (int, error) {
	args := m.Called(ctx, tenantID, assignee)
	return args.Int(0), args.Error(1)
}

func (m *MockEscalationRepository) CreateEscalation(ctx context.Context, escalation *TaskEscalation) error {
	args := m.Called(ctx, escalation)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	escalation.ID = uuid.New()
	escalation.EscalatedAt = time.Now()
	escalation.CreatedAt = time.Now()
	return nil
}

func (m *MockEscalationRepository) GetEscalationsByTaskID(ctx context.Context, taskID uuid.UUID) ([]*TaskEscalation, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*TaskEscalation), args.Error(1)
}

func (m *MockEscalationRepository) GetActiveEscalation(ctx context.Context, taskID uuid.UUID) (*TaskEscalation, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TaskEscalation), args.Error(1)
}

func (m *MockEscalationRepository) UpdateEscalation(ctx context.Context, escalation *TaskEscalation) error {
	args := m.Called(ctx, escalation)
	return args.Error(0)
}

func (m *MockEscalationRepository) CompleteEscalationsByTaskID(ctx context.Context, taskID uuid.UUID, completedBy *uuid.UUID) error {
	args := m.Called(ctx, taskID, completedBy)
	return args.Error(0)
}

// MockEscalationNotificationService includes escalation notification method
type MockEscalationNotificationService struct {
	mock.Mock
}

func (m *MockEscalationNotificationService) NotifyTaskAssigned(ctx context.Context, task *HumanTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockEscalationNotificationService) NotifyTaskCompleted(ctx context.Context, task *HumanTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockEscalationNotificationService) NotifyTaskOverdue(ctx context.Context, task *HumanTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockEscalationNotificationService) NotifyTaskEscalated(ctx context.Context, task *HumanTask, escalation *TaskEscalation) error {
	args := m.Called(ctx, task, escalation)
	return args.Error(0)
}

func TestEscalationConfig_GetMaxLevel(t *testing.T) {
	tests := []struct {
		name   string
		config EscalationConfig
		want   int
	}{
		{
			name: "single level",
			config: EscalationConfig{
				Enabled: true,
				Levels: []EscalationLevel{
					{Level: 1, TimeoutMinutes: 30, BackupApprovers: []string{"manager1"}},
				},
			},
			want: 1,
		},
		{
			name: "multiple levels",
			config: EscalationConfig{
				Enabled: true,
				Levels: []EscalationLevel{
					{Level: 1, TimeoutMinutes: 30, BackupApprovers: []string{"manager1"}},
					{Level: 2, TimeoutMinutes: 60, BackupApprovers: []string{"director1"}},
					{Level: 3, TimeoutMinutes: 120, BackupApprovers: []string{"vp1"}},
				},
			},
			want: 3,
		},
		{
			name:   "empty levels",
			config: EscalationConfig{Enabled: true, Levels: []EscalationLevel{}},
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetMaxLevel()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEscalationConfig_GetLevelConfig(t *testing.T) {
	config := EscalationConfig{
		Enabled: true,
		Levels: []EscalationLevel{
			{Level: 1, TimeoutMinutes: 30, BackupApprovers: []string{"manager1"}},
			{Level: 2, TimeoutMinutes: 60, BackupApprovers: []string{"director1"}},
		},
	}

	// Level 1 exists
	level1 := config.GetLevelConfig(1)
	require.NotNil(t, level1)
	assert.Equal(t, 30, level1.TimeoutMinutes)
	assert.Equal(t, []string{"manager1"}, level1.BackupApprovers)

	// Level 2 exists
	level2 := config.GetLevelConfig(2)
	require.NotNil(t, level2)
	assert.Equal(t, 60, level2.TimeoutMinutes)

	// Level 3 doesn't exist
	level3 := config.GetLevelConfig(3)
	assert.Nil(t, level3)
}

func TestHumanTask_Escalate(t *testing.T) {
	task := &HumanTask{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		Status:    StatusPending,
		Assignees: json.RawMessage(`["user1", "user2"]`),
	}

	newAssignees := []string{"manager1", "manager2"}
	newDueDate := time.Now().Add(30 * time.Minute)

	err := task.Escalate(newAssignees, &newDueDate)
	require.NoError(t, err)

	assert.Equal(t, 1, task.EscalationLevel)
	assert.NotNil(t, task.LastEscalatedAt)
	assert.Equal(t, &newDueDate, task.DueDate)

	var assignees []string
	err = json.Unmarshal(task.Assignees, &assignees)
	require.NoError(t, err)
	assert.Equal(t, newAssignees, assignees)
}

func TestHumanTask_EscalateFailsWhenNotPending(t *testing.T) {
	task := &HumanTask{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		Status:    StatusApproved,
		Assignees: json.RawMessage(`["user1"]`),
	}

	err := task.Escalate([]string{"manager1"}, nil)
	assert.ErrorIs(t, err, ErrTaskNotPending)
}

func TestHumanTask_CanEscalate(t *testing.T) {
	tests := []struct {
		name               string
		status             string
		escalationLevel    int
		maxEscalationLevel int
		want               bool
	}{
		{
			name:               "can escalate - pending with room",
			status:             StatusPending,
			escalationLevel:    0,
			maxEscalationLevel: 2,
			want:               true,
		},
		{
			name:               "cannot escalate - at max",
			status:             StatusPending,
			escalationLevel:    2,
			maxEscalationLevel: 2,
			want:               false,
		},
		{
			name:               "cannot escalate - not pending",
			status:             StatusApproved,
			escalationLevel:    0,
			maxEscalationLevel: 2,
			want:               false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &HumanTask{
				Status:             tt.status,
				EscalationLevel:    tt.escalationLevel,
				MaxEscalationLevel: tt.maxEscalationLevel,
			}
			assert.Equal(t, tt.want, task.CanEscalate())
		})
	}
}

func TestService_GetEscalationHistory(t *testing.T) {
	tenantID := uuid.New()
	taskID := uuid.New()

	repo := new(MockEscalationRepository)
	notif := new(MockEscalationNotificationService)
	svc := NewService(repo, notif)

	task := &HumanTask{
		ID:       taskID,
		TenantID: tenantID,
		Status:   StatusPending,
	}

	escalations := []*TaskEscalation{
		{
			ID:               uuid.New(),
			TaskID:           taskID,
			EscalationLevel:  1,
			EscalatedFrom:    json.RawMessage(`["user1"]`),
			EscalatedTo:      json.RawMessage(`["manager1"]`),
			EscalationReason: EscalationReasonTimeout,
			Status:           EscalationStatusSuperseded,
		},
		{
			ID:               uuid.New(),
			TaskID:           taskID,
			EscalationLevel:  2,
			EscalatedFrom:    json.RawMessage(`["manager1"]`),
			EscalatedTo:      json.RawMessage(`["director1"]`),
			EscalationReason: EscalationReasonTimeout,
			Status:           EscalationStatusActive,
		},
	}

	repo.On("GetByID", mock.Anything, taskID).Return(task, nil)
	repo.On("GetEscalationsByTaskID", mock.Anything, taskID).Return(escalations, nil)

	history, err := svc.GetEscalationHistory(context.Background(), tenantID, taskID)
	require.NoError(t, err)
	require.NotNil(t, history)

	assert.Equal(t, taskID, history.TaskID)
	assert.Len(t, history.Escalations, 2)
	assert.Equal(t, 1, history.Escalations[0].Level)
	assert.Equal(t, 2, history.Escalations[1].Level)
}

func TestService_GetEscalationHistory_TaskNotFound(t *testing.T) {
	tenantID := uuid.New()
	taskID := uuid.New()

	repo := new(MockEscalationRepository)
	notif := new(MockEscalationNotificationService)
	svc := NewService(repo, notif)

	repo.On("GetByID", mock.Anything, taskID).Return(nil, ErrTaskNotFound)

	_, err := svc.GetEscalationHistory(context.Background(), tenantID, taskID)
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestService_UpdateEscalationConfig(t *testing.T) {
	tenantID := uuid.New()
	taskID := uuid.New()

	repo := new(MockEscalationRepository)
	notif := new(MockEscalationNotificationService)
	svc := NewService(repo, notif)

	task := &HumanTask{
		ID:       taskID,
		TenantID: tenantID,
		Status:   StatusPending,
		Config:   json.RawMessage(`{}`),
	}

	repo.On("GetByID", mock.Anything, taskID).Return(task, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	req := UpdateEscalationRequest{
		Config: EscalationConfig{
			Enabled: true,
			Levels: []EscalationLevel{
				{Level: 1, TimeoutMinutes: 30, BackupApprovers: []string{"manager1"}},
				{Level: 2, TimeoutMinutes: 60, BackupApprovers: []string{"director1"}},
			},
			FinalAction:      TimeoutActionAutoApprove,
			NotifyOnEscalate: true,
		},
	}

	err := svc.UpdateEscalationConfig(context.Background(), tenantID, taskID, req)
	require.NoError(t, err)

	// Verify task was updated with correct max level
	repo.AssertCalled(t, "Update", mock.Anything, mock.MatchedBy(func(t *HumanTask) bool {
		return t.MaxEscalationLevel == 2
	}))
}

func TestService_UpdateEscalationConfig_TaskNotPending(t *testing.T) {
	tenantID := uuid.New()
	taskID := uuid.New()

	repo := new(MockEscalationRepository)
	notif := new(MockEscalationNotificationService)
	svc := NewService(repo, notif)

	task := &HumanTask{
		ID:       taskID,
		TenantID: tenantID,
		Status:   StatusApproved, // Already completed
		Config:   json.RawMessage(`{}`),
	}

	repo.On("GetByID", mock.Anything, taskID).Return(task, nil)

	req := UpdateEscalationRequest{
		Config: EscalationConfig{
			Enabled: true,
			Levels: []EscalationLevel{
				{Level: 1, TimeoutMinutes: 30, BackupApprovers: []string{"manager1"}},
			},
		},
	}

	err := svc.UpdateEscalationConfig(context.Background(), tenantID, taskID, req)
	assert.ErrorIs(t, err, ErrTaskNotPending)
}

func TestTaskEscalation_ToSummary(t *testing.T) {
	now := time.Now()
	escalation := &TaskEscalation{
		ID:               uuid.New(),
		TaskID:           uuid.New(),
		EscalationLevel:  2,
		EscalatedAt:      now,
		EscalatedFrom:    json.RawMessage(`["user1", "user2"]`),
		EscalatedTo:      json.RawMessage(`["manager1"]`),
		EscalationReason: EscalationReasonTimeout,
		Status:           EscalationStatusActive,
	}

	summary := escalation.ToSummary()

	assert.Equal(t, escalation.ID, summary.ID)
	assert.Equal(t, 2, summary.Level)
	assert.Equal(t, []string{"user1", "user2"}, summary.FromAssignees)
	assert.Equal(t, []string{"manager1"}, summary.ToAssignees)
	assert.Equal(t, EscalationReasonTimeout, summary.Reason)
	assert.Equal(t, EscalationStatusActive, summary.Status)
}

func TestParseEscalationConfig(t *testing.T) {
	tests := []struct {
		name       string
		configJSON json.RawMessage
		wantNil    bool
		wantErr    bool
	}{
		{
			name:       "empty config",
			configJSON: json.RawMessage(`{}`),
			wantNil:    true,
			wantErr:    false,
		},
		{
			name: "valid escalation config",
			configJSON: json.RawMessage(`{
				"escalation": {
					"enabled": true,
					"levels": [
						{"level": 1, "timeout_minutes": 30, "backup_approvers": ["manager1"]}
					],
					"final_action": "auto_approve"
				}
			}`),
			wantNil: false,
			wantErr: false,
		},
		{
			name:       "null config",
			configJSON: nil,
			wantNil:    true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseEscalationConfig(tt.configJSON)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantNil {
				assert.Nil(t, config)
			} else {
				assert.NotNil(t, config)
			}
		})
	}
}
