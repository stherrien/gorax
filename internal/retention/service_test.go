package retention

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetRetentionPolicy(ctx context.Context, tenantID string) (*RetentionPolicy, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RetentionPolicy), args.Error(1)
}

func (m *MockRepository) DeleteOldExecutions(ctx context.Context, tenantID string, cutoffDate time.Time, batchSize int) (*CleanupResult, error) {
	args := m.Called(ctx, tenantID, cutoffDate, batchSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CleanupResult), args.Error(1)
}

func (m *MockRepository) ArchiveAndDeleteOldExecutions(ctx context.Context, tenantID string, cutoffDate time.Time, batchSize int) (*CleanupResult, error) {
	args := m.Called(ctx, tenantID, cutoffDate, batchSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CleanupResult), args.Error(1)
}

func (m *MockRepository) GetTenantsWithRetention(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRepository) LogCleanup(ctx context.Context, log *CleanupLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func TestNewService(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := DefaultConfig()

	service := NewService(repo, logger, config)

	assert.NotNil(t, service)
	assert.Equal(t, repo, service.repo)
	assert.Equal(t, logger, service.logger)
	assert.Equal(t, config, service.config)
}

func TestService_GetRetentionPolicy(t *testing.T) {
	tests := []struct {
		name      string
		tenantID  string
		mockSetup func(*MockRepository)
		want      *RetentionPolicy
		wantErr   bool
	}{
		{
			name:     "successful retrieval",
			tenantID: "tenant-1",
			mockSetup: func(m *MockRepository) {
				policy := &RetentionPolicy{
					TenantID:      "tenant-1",
					RetentionDays: 30,
					Enabled:       true,
				}
				m.On("GetRetentionPolicy", mock.Anything, "tenant-1").Return(policy, nil)
			},
			want: &RetentionPolicy{
				TenantID:      "tenant-1",
				RetentionDays: 30,
				Enabled:       true,
			},
			wantErr: false,
		},
		{
			name:     "tenant not found - returns default policy",
			tenantID: "tenant-2",
			mockSetup: func(m *MockRepository) {
				m.On("GetRetentionPolicy", mock.Anything, "tenant-2").Return(nil, ErrNotFound)
			},
			want: &RetentionPolicy{
				TenantID:      "tenant-2",
				RetentionDays: 90, // Default from config
				Enabled:       true,
			},
			wantErr: false,
		},
		{
			name:     "database error",
			tenantID: "tenant-3",
			mockSetup: func(m *MockRepository) {
				m.On("GetRetentionPolicy", mock.Anything, "tenant-3").Return(nil, errors.New("db error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			tt.mockSetup(repo)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			config := DefaultConfig()
			service := NewService(repo, logger, config)

			got, err := service.GetRetentionPolicy(context.Background(), tt.tenantID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_CleanupOldExecutions(t *testing.T) {
	tests := []struct {
		name      string
		tenantID  string
		mockSetup func(*MockRepository)
		want      *CleanupResult
		wantErr   bool
	}{
		{
			name:     "successful cleanup with executions to delete",
			tenantID: "tenant-1",
			mockSetup: func(m *MockRepository) {
				policy := &RetentionPolicy{
					TenantID:      "tenant-1",
					RetentionDays: 30,
					Enabled:       true,
				}
				m.On("GetRetentionPolicy", mock.Anything, "tenant-1").Return(policy, nil)

				result := &CleanupResult{
					ExecutionsDeleted:     150,
					StepExecutionsDeleted: 450,
					BatchesProcessed:      2,
				}
				m.On("DeleteOldExecutions", mock.Anything, "tenant-1", mock.AnythingOfType("time.Time"), 1000).Return(result, nil)
				m.On("LogCleanup", mock.Anything, mock.AnythingOfType("*retention.CleanupLog")).Return(nil)
			},
			want: &CleanupResult{
				ExecutionsDeleted:     150,
				StepExecutionsDeleted: 450,
				BatchesProcessed:      2,
			},
			wantErr: false,
		},
		{
			name:     "retention disabled for tenant",
			tenantID: "tenant-2",
			mockSetup: func(m *MockRepository) {
				policy := &RetentionPolicy{
					TenantID:      "tenant-2",
					RetentionDays: 30,
					Enabled:       false,
				}
				m.On("GetRetentionPolicy", mock.Anything, "tenant-2").Return(policy, nil)
			},
			want: &CleanupResult{
				ExecutionsDeleted:     0,
				StepExecutionsDeleted: 0,
				BatchesProcessed:      0,
			},
			wantErr: false,
		},
		{
			name:     "no executions to delete",
			tenantID: "tenant-3",
			mockSetup: func(m *MockRepository) {
				policy := &RetentionPolicy{
					TenantID:      "tenant-3",
					RetentionDays: 30,
					Enabled:       true,
				}
				m.On("GetRetentionPolicy", mock.Anything, "tenant-3").Return(policy, nil)

				result := &CleanupResult{
					ExecutionsDeleted:     0,
					StepExecutionsDeleted: 0,
					BatchesProcessed:      0,
				}
				m.On("DeleteOldExecutions", mock.Anything, "tenant-3", mock.AnythingOfType("time.Time"), 1000).Return(result, nil)
				m.On("LogCleanup", mock.Anything, mock.AnythingOfType("*retention.CleanupLog")).Return(nil)
			},
			want: &CleanupResult{
				ExecutionsDeleted:     0,
				StepExecutionsDeleted: 0,
				BatchesProcessed:      0,
			},
			wantErr: false,
		},
		{
			name:     "deletion error",
			tenantID: "tenant-4",
			mockSetup: func(m *MockRepository) {
				policy := &RetentionPolicy{
					TenantID:      "tenant-4",
					RetentionDays: 30,
					Enabled:       true,
				}
				m.On("GetRetentionPolicy", mock.Anything, "tenant-4").Return(policy, nil)
				m.On("DeleteOldExecutions", mock.Anything, "tenant-4", mock.AnythingOfType("time.Time"), 1000).Return(nil, errors.New("deletion error"))
				// Expect log cleanup call for the error
				m.On("LogCleanup", mock.Anything, mock.AnythingOfType("*retention.CleanupLog")).Return(nil)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			tt.mockSetup(repo)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			config := DefaultConfig()
			// Disable archival for these tests to use DeleteOldExecutions
			config.ArchiveBeforeDelete = false
			service := NewService(repo, logger, config)

			got, err := service.CleanupOldExecutions(context.Background(), tt.tenantID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_CleanupAllTenants(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(*MockRepository)
		wantTotal *CleanupResult
		wantErr   bool
	}{
		{
			name: "successful cleanup for multiple tenants",
			mockSetup: func(m *MockRepository) {
				m.On("GetTenantsWithRetention", mock.Anything).Return([]string{"tenant-1", "tenant-2"}, nil)

				// Tenant 1
				policy1 := &RetentionPolicy{TenantID: "tenant-1", RetentionDays: 30, Enabled: true}
				m.On("GetRetentionPolicy", mock.Anything, "tenant-1").Return(policy1, nil)
				result1 := &CleanupResult{ExecutionsDeleted: 100, StepExecutionsDeleted: 300, BatchesProcessed: 1}
				m.On("DeleteOldExecutions", mock.Anything, "tenant-1", mock.AnythingOfType("time.Time"), 1000).Return(result1, nil)
				m.On("LogCleanup", mock.Anything, mock.MatchedBy(func(log *CleanupLog) bool {
					return log.TenantID == "tenant-1"
				})).Return(nil)

				// Tenant 2
				policy2 := &RetentionPolicy{TenantID: "tenant-2", RetentionDays: 60, Enabled: true}
				m.On("GetRetentionPolicy", mock.Anything, "tenant-2").Return(policy2, nil)
				result2 := &CleanupResult{ExecutionsDeleted: 50, StepExecutionsDeleted: 150, BatchesProcessed: 1}
				m.On("DeleteOldExecutions", mock.Anything, "tenant-2", mock.AnythingOfType("time.Time"), 1000).Return(result2, nil)
				m.On("LogCleanup", mock.Anything, mock.MatchedBy(func(log *CleanupLog) bool {
					return log.TenantID == "tenant-2"
				})).Return(nil)
			},
			wantTotal: &CleanupResult{
				ExecutionsDeleted:     150,
				StepExecutionsDeleted: 450,
				BatchesProcessed:      2,
			},
			wantErr: false,
		},
		{
			name: "skip disabled tenant",
			mockSetup: func(m *MockRepository) {
				m.On("GetTenantsWithRetention", mock.Anything).Return([]string{"tenant-1", "tenant-2"}, nil)

				// Tenant 1 - enabled
				policy1 := &RetentionPolicy{TenantID: "tenant-1", RetentionDays: 30, Enabled: true}
				m.On("GetRetentionPolicy", mock.Anything, "tenant-1").Return(policy1, nil)
				result1 := &CleanupResult{ExecutionsDeleted: 100, StepExecutionsDeleted: 300, BatchesProcessed: 1}
				m.On("DeleteOldExecutions", mock.Anything, "tenant-1", mock.AnythingOfType("time.Time"), 1000).Return(result1, nil)
				m.On("LogCleanup", mock.Anything, mock.MatchedBy(func(log *CleanupLog) bool {
					return log.TenantID == "tenant-1"
				})).Return(nil)

				// Tenant 2 - disabled
				policy2 := &RetentionPolicy{TenantID: "tenant-2", RetentionDays: 60, Enabled: false}
				m.On("GetRetentionPolicy", mock.Anything, "tenant-2").Return(policy2, nil)
			},
			wantTotal: &CleanupResult{
				ExecutionsDeleted:     100,
				StepExecutionsDeleted: 300,
				BatchesProcessed:      1,
			},
			wantErr: false,
		},
		{
			name: "error getting tenants",
			mockSetup: func(m *MockRepository) {
				m.On("GetTenantsWithRetention", mock.Anything).Return(nil, errors.New("db error"))
			},
			wantTotal: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			tt.mockSetup(repo)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			config := DefaultConfig()
			// Disable archival for these tests to use DeleteOldExecutions
			config.ArchiveBeforeDelete = false
			service := NewService(repo, logger, config)

			got, err := service.CleanupAllTenants(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTotal, got)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_CalculateCutoffDate(t *testing.T) {
	tests := []struct {
		name          string
		retentionDays int
		baseTime      time.Time
		want          time.Time
	}{
		{
			name:          "30 days retention",
			retentionDays: 30,
			baseTime:      time.Date(2024, 1, 31, 12, 0, 0, 0, time.UTC),
			want:          time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name:          "90 days retention",
			retentionDays: 90,
			baseTime:      time.Date(2024, 4, 1, 12, 0, 0, 0, time.UTC),
			want:          time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
		},
		{
			name:          "7 days retention",
			retentionDays: 7,
			baseTime:      time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			want:          time.Date(2024, 1, 8, 12, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			config := DefaultConfig()
			service := NewService(repo, logger, config)

			got := service.calculateCutoffDate(tt.baseTime, tt.retentionDays)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, 90, config.DefaultRetentionDays)
	assert.Equal(t, 1000, config.BatchSize)
	assert.True(t, config.EnableAuditLog)
	assert.True(t, config.ArchiveBeforeDelete)
}

func TestService_CleanupOldExecutions_WithArchival(t *testing.T) {
	tests := []struct {
		name      string
		tenantID  string
		mockSetup func(*MockRepository)
		want      *CleanupResult
		wantErr   bool
	}{
		{
			name:     "successful archival and cleanup",
			tenantID: "tenant-1",
			mockSetup: func(m *MockRepository) {
				policy := &RetentionPolicy{
					TenantID:      "tenant-1",
					RetentionDays: 30,
					Enabled:       true,
				}
				m.On("GetRetentionPolicy", mock.Anything, "tenant-1").Return(policy, nil)

				result := &CleanupResult{
					ExecutionsDeleted:     150,
					ExecutionsArchived:    150,
					StepExecutionsDeleted: 450,
					BatchesProcessed:      2,
				}
				m.On("ArchiveAndDeleteOldExecutions", mock.Anything, "tenant-1", mock.AnythingOfType("time.Time"), 1000).Return(result, nil)
				m.On("LogCleanup", mock.Anything, mock.MatchedBy(func(log *CleanupLog) bool {
					return log.TenantID == "tenant-1" && log.ExecutionsArchived == 150
				})).Return(nil)
			},
			want: &CleanupResult{
				ExecutionsDeleted:     150,
				ExecutionsArchived:    150,
				StepExecutionsDeleted: 450,
				BatchesProcessed:      2,
			},
			wantErr: false,
		},
		{
			name:     "archival error",
			tenantID: "tenant-2",
			mockSetup: func(m *MockRepository) {
				policy := &RetentionPolicy{
					TenantID:      "tenant-2",
					RetentionDays: 30,
					Enabled:       true,
				}
				m.On("GetRetentionPolicy", mock.Anything, "tenant-2").Return(policy, nil)
				m.On("ArchiveAndDeleteOldExecutions", mock.Anything, "tenant-2", mock.AnythingOfType("time.Time"), 1000).Return(nil, errors.New("archival error"))
				m.On("LogCleanup", mock.Anything, mock.AnythingOfType("*retention.CleanupLog")).Return(nil)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			tt.mockSetup(repo)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			config := DefaultConfig()
			// Enable archival (default behavior)
			config.ArchiveBeforeDelete = true
			service := NewService(repo, logger, config)

			got, err := service.CleanupOldExecutions(context.Background(), tt.tenantID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			repo.AssertExpectations(t)
		})
	}
}
