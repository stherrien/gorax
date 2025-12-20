package webhook

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCleanupRepository is a mock implementation of the repository for testing
type MockCleanupRepository struct {
	mock.Mock
}

func (m *MockCleanupRepository) DeleteOldEvents(ctx context.Context, retentionPeriod time.Duration, batchSize int) (int, error) {
	args := m.Called(ctx, retentionPeriod, batchSize)
	return args.Int(0), args.Error(1)
}

func TestCleanupService_Run_Success(t *testing.T) {
	mockRepo := new(MockCleanupRepository)
	ctx := context.Background()
	retentionPeriod := 30 * 24 * time.Hour
	batchSize := 1000

	// Mock expects batch deletion calls
	// First batch: 1000 deleted
	mockRepo.On("DeleteOldEvents", ctx, retentionPeriod, batchSize).Return(1000, nil).Once()
	// Second batch: 500 deleted
	mockRepo.On("DeleteOldEvents", ctx, retentionPeriod, batchSize).Return(500, nil).Once()
	// Third batch: 0 deleted (no more old events)
	mockRepo.On("DeleteOldEvents", ctx, retentionPeriod, batchSize).Return(0, nil).Once()

	service := NewCleanupService(mockRepo, batchSize, retentionPeriod)

	result, err := service.Run(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 1500, result.TotalDeleted)
	assert.Equal(t, 2, result.BatchesProcessed)
	assert.GreaterOrEqual(t, result.DurationMs, int64(0))
	mockRepo.AssertExpectations(t)
}

func TestCleanupService_Run_EmptyDatabase(t *testing.T) {
	mockRepo := new(MockCleanupRepository)
	ctx := context.Background()
	retentionPeriod := 30 * 24 * time.Hour
	batchSize := 1000

	// First call returns 0 (no old events)
	mockRepo.On("DeleteOldEvents", ctx, retentionPeriod, batchSize).Return(0, nil).Once()

	service := NewCleanupService(mockRepo, batchSize, retentionPeriod)

	result, err := service.Run(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.TotalDeleted)
	assert.Equal(t, 0, result.BatchesProcessed)
	assert.GreaterOrEqual(t, result.DurationMs, int64(0))
	mockRepo.AssertExpectations(t)
}

func TestCleanupService_Run_ContextCancellation(t *testing.T) {
	mockRepo := new(MockCleanupRepository)
	ctx, cancel := context.WithCancel(context.Background())
	retentionPeriod := 30 * 24 * time.Hour
	batchSize := 1000

	// First batch succeeds, then cancel context
	mockRepo.On("DeleteOldEvents", mock.Anything, retentionPeriod, batchSize).Return(1000, nil).Run(func(args mock.Arguments) {
		// Cancel context after first batch
		cancel()
	}).Once()

	service := NewCleanupService(mockRepo, batchSize, retentionPeriod)

	result, err := service.Run(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	assert.Equal(t, 1000, result.TotalDeleted)
	assert.Equal(t, 1, result.BatchesProcessed)
	mockRepo.AssertExpectations(t)
}

func TestCleanupService_Run_WithError(t *testing.T) {
	mockRepo := new(MockCleanupRepository)
	ctx := context.Background()
	retentionPeriod := 30 * 24 * time.Hour
	batchSize := 1000

	// First batch succeeds
	mockRepo.On("DeleteOldEvents", ctx, retentionPeriod, batchSize).Return(1000, nil).Once()
	// Second batch fails
	mockRepo.On("DeleteOldEvents", ctx, retentionPeriod, batchSize).Return(0, assert.AnError).Once()

	service := NewCleanupService(mockRepo, batchSize, retentionPeriod)

	result, err := service.Run(ctx)

	assert.Error(t, err)
	assert.Equal(t, 1000, result.TotalDeleted)
	assert.Equal(t, 1, result.BatchesProcessed)
	mockRepo.AssertExpectations(t)
}

func TestCleanupService_Run_WithDefaultValues(t *testing.T) {
	mockRepo := new(MockCleanupRepository)
	ctx := context.Background()

	// Use default values
	defaultRetention := 30 * 24 * time.Hour
	defaultBatchSize := 1000

	mockRepo.On("DeleteOldEvents", ctx, defaultRetention, defaultBatchSize).Return(0, nil).Once()

	service := NewCleanupService(mockRepo, defaultBatchSize, defaultRetention)

	result, err := service.Run(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.TotalDeleted)
	mockRepo.AssertExpectations(t)
}

func TestCleanupService_Run_CustomRetentionPeriod(t *testing.T) {
	mockRepo := new(MockCleanupRepository)
	ctx := context.Background()

	// Custom 7-day retention
	customRetention := 7 * 24 * time.Hour
	batchSize := 500

	mockRepo.On("DeleteOldEvents", ctx, customRetention, batchSize).Return(100, nil).Once()
	mockRepo.On("DeleteOldEvents", ctx, customRetention, batchSize).Return(0, nil).Once()

	service := NewCleanupService(mockRepo, batchSize, customRetention)

	result, err := service.Run(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 100, result.TotalDeleted)
	assert.Equal(t, 1, result.BatchesProcessed)
	mockRepo.AssertExpectations(t)
}

func TestCleanupService_Run_LargeBatch(t *testing.T) {
	mockRepo := new(MockCleanupRepository)
	ctx := context.Background()
	retentionPeriod := 30 * 24 * time.Hour
	batchSize := 5000

	// Process multiple large batches
	mockRepo.On("DeleteOldEvents", ctx, retentionPeriod, batchSize).Return(5000, nil).Once()
	mockRepo.On("DeleteOldEvents", ctx, retentionPeriod, batchSize).Return(5000, nil).Once()
	mockRepo.On("DeleteOldEvents", ctx, retentionPeriod, batchSize).Return(3000, nil).Once()
	mockRepo.On("DeleteOldEvents", ctx, retentionPeriod, batchSize).Return(0, nil).Once()

	service := NewCleanupService(mockRepo, batchSize, retentionPeriod)

	result, err := service.Run(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 13000, result.TotalDeleted)
	assert.Equal(t, 3, result.BatchesProcessed)
	mockRepo.AssertExpectations(t)
}
