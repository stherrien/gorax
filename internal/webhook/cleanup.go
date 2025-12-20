package webhook

import (
	"context"
	"fmt"
	"time"
)

// CleanupRepository defines the interface for cleanup operations
type CleanupRepository interface {
	DeleteOldEvents(ctx context.Context, retentionPeriod time.Duration, batchSize int) (int, error)
}

// CleanupService handles webhook event cleanup operations
type CleanupService struct {
	repo            CleanupRepository
	batchSize       int
	retentionPeriod time.Duration
}

// CleanupResult contains statistics from the cleanup operation
type CleanupResult struct {
	TotalDeleted     int
	BatchesProcessed int
	DurationMs       int64
	StartTime        time.Time
	EndTime          time.Time
}

// NewCleanupService creates a new cleanup service
func NewCleanupService(repo CleanupRepository, batchSize int, retentionPeriod time.Duration) *CleanupService {
	// Set defaults if not provided
	if batchSize <= 0 {
		batchSize = 1000
	}
	if retentionPeriod <= 0 {
		retentionPeriod = 30 * 24 * time.Hour // 30 days
	}

	return &CleanupService{
		repo:            repo,
		batchSize:       batchSize,
		retentionPeriod: retentionPeriod,
	}
}

// Run executes the cleanup process
func (s *CleanupService) Run(ctx context.Context) (*CleanupResult, error) {
	startTime := time.Now()

	result := &CleanupResult{
		TotalDeleted:     0,
		BatchesProcessed: 0,
		StartTime:        startTime,
	}

	for {
		// Check context cancellation
		if err := ctx.Err(); err != nil {
			result.EndTime = time.Now()
			result.DurationMs = time.Since(startTime).Milliseconds()
			return result, fmt.Errorf("context canceled: %w", err)
		}

		// Delete a batch of old events
		deleted, err := s.repo.DeleteOldEvents(ctx, s.retentionPeriod, s.batchSize)
		if err != nil {
			result.EndTime = time.Now()
			result.DurationMs = time.Since(startTime).Milliseconds()
			return result, fmt.Errorf("delete batch failed: %w", err)
		}

		// If no events were deleted, we're done
		if deleted == 0 {
			break
		}

		result.TotalDeleted += deleted
		result.BatchesProcessed++
	}

	result.EndTime = time.Now()
	result.DurationMs = time.Since(startTime).Milliseconds()

	return result, nil
}

// SetRetentionPeriod updates the retention period
func (s *CleanupService) SetRetentionPeriod(period time.Duration) {
	s.retentionPeriod = period
}

// SetBatchSize updates the batch size
func (s *CleanupService) SetBatchSize(size int) {
	if size > 0 {
		s.batchSize = size
	}
}

// GetRetentionPeriod returns the current retention period
func (s *CleanupService) GetRetentionPeriod() time.Duration {
	return s.retentionPeriod
}

// GetBatchSize returns the current batch size
func (s *CleanupService) GetBatchSize() int {
	return s.batchSize
}
