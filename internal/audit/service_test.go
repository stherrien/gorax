package audit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of the audit repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateAuditEvent(ctx context.Context, event *AuditEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockRepository) CreateAuditEventBatch(ctx context.Context, events []*AuditEvent) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *MockRepository) GetAuditEvent(ctx context.Context, tenantID, eventID string) (*AuditEvent, error) {
	args := m.Called(ctx, tenantID, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuditEvent), args.Error(1)
}

func (m *MockRepository) QueryAuditEvents(ctx context.Context, filter QueryFilter) ([]AuditEvent, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]AuditEvent), args.Int(1), args.Error(2)
}

func (m *MockRepository) GetAuditStats(ctx context.Context, tenantID string, timeRange TimeRange) (*AuditStats, error) {
	args := m.Called(ctx, tenantID, timeRange)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuditStats), args.Error(1)
}

func (m *MockRepository) GetRetentionPolicy(ctx context.Context, tenantID string) (*RetentionPolicy, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RetentionPolicy), args.Error(1)
}

func (m *MockRepository) UpdateRetentionPolicy(ctx context.Context, policy *RetentionPolicy) error {
	args := m.Called(ctx, policy)
	return args.Error(0)
}

func (m *MockRepository) DeleteOldAuditEvents(ctx context.Context, tenantID string, cutoffDate time.Time) (int64, error) {
	args := m.Called(ctx, tenantID, cutoffDate)
	return args.Get(0).(int64), args.Error(1)
}

func TestService_LogEvent(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, 100, 5*time.Second)
	defer svc.Close()

	ctx := context.Background()
	event := &AuditEvent{
		TenantID:  "tenant-1",
		UserID:    "user-1",
		Category:  CategoryWorkflow,
		EventType: EventTypeExecute,
		Action:    "workflow.executed",
		Severity:  SeverityInfo,
		Status:    StatusSuccess,
	}

	// Expect batch operation for async processing
	mockRepo.On("CreateAuditEventBatch", mock.Anything, mock.AnythingOfType("[]*audit.AuditEvent")).Return(nil)

	err := svc.LogEvent(ctx, event)
	require.NoError(t, err)

	// Wait for async processing and flush
	svc.Flush()

	mockRepo.AssertExpectations(t)
}

func TestService_LogEventSync(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, 100, 5*time.Second)
	defer svc.Close()

	ctx := context.Background()
	event := &AuditEvent{
		TenantID:  "tenant-1",
		Category:  CategoryAuthentication,
		EventType: EventTypeLogin,
		Action:    "user.login",
		Severity:  SeverityInfo,
		Status:    StatusSuccess,
	}

	mockRepo.On("CreateAuditEvent", ctx, event).Return(nil)

	err := svc.LogEventSync(ctx, event)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestService_LogEventBatch(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, 100, 5*time.Second)
	defer svc.Close()

	ctx := context.Background()
	events := []*AuditEvent{
		{
			TenantID:  "tenant-1",
			Category:  CategoryWorkflow,
			EventType: EventTypeCreate,
			Action:    "workflow.created",
			Severity:  SeverityInfo,
			Status:    StatusSuccess,
		},
		{
			TenantID:  "tenant-1",
			Category:  CategoryWorkflow,
			EventType: EventTypeUpdate,
			Action:    "workflow.updated",
			Severity:  SeverityInfo,
			Status:    StatusSuccess,
		},
	}

	mockRepo.On("CreateAuditEventBatch", ctx, events).Return(nil)

	err := svc.LogEventBatch(ctx, events)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestService_GetAuditEvent(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, 100, 5*time.Second)
	defer svc.Close()

	ctx := context.Background()
	tenantID := "tenant-1"
	eventID := "event-1"

	expectedEvent := &AuditEvent{
		ID:        eventID,
		TenantID:  tenantID,
		Category:  CategoryWorkflow,
		EventType: EventTypeExecute,
		Action:    "workflow.executed",
		Severity:  SeverityInfo,
		Status:    StatusSuccess,
	}

	mockRepo.On("GetAuditEvent", ctx, tenantID, eventID).Return(expectedEvent, nil)

	event, err := svc.GetAuditEvent(ctx, tenantID, eventID)
	require.NoError(t, err)
	assert.Equal(t, expectedEvent.ID, event.ID)
	assert.Equal(t, expectedEvent.Action, event.Action)

	mockRepo.AssertExpectations(t)
}

func TestService_QueryAuditEvents(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, 100, 5*time.Second)
	defer svc.Close()

	ctx := context.Background()
	filter := QueryFilter{
		TenantID:   "tenant-1",
		Categories: []Category{CategoryWorkflow},
		Limit:      10,
	}

	expectedEvents := []AuditEvent{
		{
			ID:        "event-1",
			TenantID:  "tenant-1",
			Category:  CategoryWorkflow,
			EventType: EventTypeExecute,
			Action:    "workflow.executed",
			Severity:  SeverityInfo,
			Status:    StatusSuccess,
		},
	}

	mockRepo.On("QueryAuditEvents", ctx, filter).Return(expectedEvents, 1, nil)

	events, total, err := svc.QueryAuditEvents(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, events, 1)
	assert.Equal(t, "event-1", events[0].ID)

	mockRepo.AssertExpectations(t)
}

func TestService_GetAuditStats(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, 100, 5*time.Second)
	defer svc.Close()

	ctx := context.Background()
	tenantID := "tenant-1"
	timeRange := TimeRange{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now(),
	}

	expectedStats := &AuditStats{
		TotalEvents: 100,
		EventsByCategory: map[Category]int{
			CategoryWorkflow:   50,
			CategoryCredential: 30,
		},
		EventsBySeverity: map[Severity]int{
			SeverityInfo:     80,
			SeverityWarning:  15,
			SeverityCritical: 5,
		},
		CriticalEvents: 5,
		FailedEvents:   10,
	}

	mockRepo.On("GetAuditStats", ctx, tenantID, timeRange).Return(expectedStats, nil)

	stats, err := svc.GetAuditStats(ctx, tenantID, timeRange)
	require.NoError(t, err)
	assert.Equal(t, 100, stats.TotalEvents)
	assert.Equal(t, 5, stats.CriticalEvents)
	assert.Equal(t, 10, stats.FailedEvents)

	mockRepo.AssertExpectations(t)
}

func TestService_GetRetentionPolicy(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, 100, 5*time.Second)
	defer svc.Close()

	ctx := context.Background()
	tenantID := "tenant-1"

	expectedPolicy := &RetentionPolicy{
		TenantID:          tenantID,
		HotRetentionDays:  90,
		WarmRetentionDays: 365,
		ColdRetentionDays: 2555,
		ArchiveEnabled:    true,
	}

	mockRepo.On("GetRetentionPolicy", ctx, tenantID).Return(expectedPolicy, nil)

	policy, err := svc.GetRetentionPolicy(ctx, tenantID)
	require.NoError(t, err)
	assert.Equal(t, 90, policy.HotRetentionDays)
	assert.True(t, policy.ArchiveEnabled)

	mockRepo.AssertExpectations(t)
}

func TestService_UpdateRetentionPolicy(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, 100, 5*time.Second)
	defer svc.Close()

	ctx := context.Background()
	policy := &RetentionPolicy{
		TenantID:          "tenant-1",
		HotRetentionDays:  180,
		WarmRetentionDays: 365,
		ColdRetentionDays: 2555,
		ArchiveEnabled:    false,
	}

	mockRepo.On("UpdateRetentionPolicy", ctx, policy).Return(nil)

	err := svc.UpdateRetentionPolicy(ctx, policy)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestService_ValidateEvent(t *testing.T) {
	tests := []struct {
		name    string
		event   *AuditEvent
		wantErr bool
	}{
		{
			name: "valid event",
			event: &AuditEvent{
				TenantID:  "tenant-1",
				Category:  CategoryWorkflow,
				EventType: EventTypeExecute,
				Action:    "workflow.executed",
				Severity:  SeverityInfo,
				Status:    StatusSuccess,
			},
			wantErr: false,
		},
		{
			name: "missing tenant_id",
			event: &AuditEvent{
				Category:  CategoryWorkflow,
				EventType: EventTypeExecute,
				Action:    "workflow.executed",
				Severity:  SeverityInfo,
				Status:    StatusSuccess,
			},
			wantErr: true,
		},
		{
			name: "missing category",
			event: &AuditEvent{
				TenantID:  "tenant-1",
				EventType: EventTypeExecute,
				Action:    "workflow.executed",
				Severity:  SeverityInfo,
				Status:    StatusSuccess,
			},
			wantErr: true,
		},
		{
			name: "missing action",
			event: &AuditEvent{
				TenantID:  "tenant-1",
				Category:  CategoryWorkflow,
				EventType: EventTypeExecute,
				Severity:  SeverityInfo,
				Status:    StatusSuccess,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEvent(tt.event)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_AsyncBuffering(t *testing.T) {
	mockRepo := new(MockRepository)
	bufferSize := 3
	svc := NewService(mockRepo, bufferSize, 1*time.Second)
	defer svc.Close()

	ctx := context.Background()

	// Create events to fill the buffer
	events := make([]*AuditEvent, bufferSize)
	for i := 0; i < bufferSize; i++ {
		events[i] = &AuditEvent{
			TenantID:  "tenant-1",
			Category:  CategoryWorkflow,
			EventType: EventTypeExecute,
			Action:    "workflow.executed",
			Severity:  SeverityInfo,
			Status:    StatusSuccess,
		}
	}

	// Expect batch insert when buffer is full
	mockRepo.On("CreateAuditEventBatch", mock.Anything, mock.MatchedBy(func(e []*AuditEvent) bool {
		return len(e) == bufferSize
	})).Return(nil).Once()

	// Log events
	for _, event := range events {
		err := svc.LogEvent(ctx, event)
		require.NoError(t, err)
	}

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}

func TestService_ErrorHandling(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewService(mockRepo, 100, 5*time.Second)
	defer svc.Close()

	ctx := context.Background()
	event := &AuditEvent{
		TenantID:  "tenant-1",
		Category:  CategoryWorkflow,
		EventType: EventTypeExecute,
		Action:    "workflow.executed",
		Severity:  SeverityInfo,
		Status:    StatusSuccess,
	}

	expectedErr := errors.New("database error")
	mockRepo.On("CreateAuditEvent", ctx, event).Return(expectedErr)

	err := svc.LogEventSync(ctx, event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")

	mockRepo.AssertExpectations(t)
}
