package webhook

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// CleanupScheduler manages scheduled webhook event cleanup
type CleanupScheduler struct {
	service  *CleanupService
	logger   *slog.Logger
	schedule string
	cron     *cron.Cron

	// Running state
	running bool
	mu      sync.Mutex
	wg      sync.WaitGroup
	stopCh  chan struct{}
}

// NewCleanupScheduler creates a new cleanup scheduler
func NewCleanupScheduler(service *CleanupService, schedule string, logger *slog.Logger) *CleanupScheduler {
	return &CleanupScheduler{
		service:  service,
		logger:   logger,
		schedule: schedule,
		stopCh:   make(chan struct{}),
	}
}

// Start starts the cleanup scheduler
func (s *CleanupScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	s.logger.Info("cleanup scheduler started", "schedule", s.schedule)

	// Create cron scheduler
	s.cron = cron.New()

	// Add cleanup job to cron
	_, err := s.cron.AddFunc(s.schedule, func() {
		s.runCleanup(ctx)
	})
	if err != nil {
		s.logger.Error("failed to add cleanup job to cron", "error", err)
		return err
	}

	// Start cron
	s.cron.Start()

	// Run initial cleanup immediately
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.runCleanup(ctx)
	}()

	// Wait for stop signal
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		<-s.stopCh
		s.cron.Stop()
	}()

	return nil
}

// Stop stops the scheduler gracefully
func (s *CleanupScheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	s.logger.Info("stopping cleanup scheduler...")
	close(s.stopCh)
	s.wg.Wait()
	s.logger.Info("cleanup scheduler stopped")
}

// Wait waits for the scheduler to finish
func (s *CleanupScheduler) Wait() {
	s.wg.Wait()
}

// runCleanup executes the cleanup process
func (s *CleanupScheduler) runCleanup(ctx context.Context) {
	s.logger.Info("starting webhook event cleanup")
	startTime := time.Now()

	result, err := s.service.Run(ctx)
	if err != nil {
		s.logger.Error("cleanup failed",
			"error", err,
			"total_deleted", result.TotalDeleted,
			"batches_processed", result.BatchesProcessed,
			"duration_ms", result.DurationMs,
		)
		return
	}

	s.logger.Info("cleanup completed",
		"total_deleted", result.TotalDeleted,
		"batches_processed", result.BatchesProcessed,
		"duration_ms", result.DurationMs,
		"retention_period", s.service.GetRetentionPeriod().String(),
	)

	// Log warning if cleanup took too long
	if time.Since(startTime) > 5*time.Minute {
		s.logger.Warn("cleanup took longer than expected",
			"duration", time.Since(startTime).String(),
		)
	}
}
