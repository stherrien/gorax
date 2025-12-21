package retention

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// CleanupService defines the interface for cleanup operations
type CleanupService interface {
	CleanupAllTenants(ctx context.Context) (*CleanupResult, error)
}

// Scheduler manages scheduled retention cleanup operations
type Scheduler struct {
	service  CleanupService
	logger   *slog.Logger
	interval time.Duration

	// Running state
	running bool
	mu      sync.Mutex
	wg      sync.WaitGroup
	stopCh  chan struct{}
}

// NewScheduler creates a new retention cleanup scheduler
func NewScheduler(service CleanupService, logger *slog.Logger, interval time.Duration) *Scheduler {
	return &Scheduler{
		service:  service,
		logger:   logger,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start starts the cleanup scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	s.logger.Info("retention cleanup scheduler started", "interval", s.interval)

	s.wg.Add(1)
	go s.run(ctx)

	return nil
}

// Stop stops the scheduler gracefully
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	s.logger.Info("stopping retention cleanup scheduler...")
	close(s.stopCh)
	s.wg.Wait()
	s.logger.Info("retention cleanup scheduler stopped")
}

// Wait waits for the scheduler to finish
func (s *Scheduler) Wait() {
	s.wg.Wait()
}

// IsRunning returns whether the scheduler is currently running
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// SetInterval sets the cleanup interval
func (s *Scheduler) SetInterval(interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.interval = interval
}

// RunOnce executes cleanup once without starting the scheduler
func (s *Scheduler) RunOnce(ctx context.Context) (*CleanupResult, error) {
	s.logger.Info("executing one-time retention cleanup")
	return s.executeCleanup(ctx)
}

// run is the main scheduler loop
func (s *Scheduler) run(ctx context.Context) {
	defer s.wg.Done()
	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run initial cleanup immediately
	s.executeCleanup(ctx)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("retention cleanup scheduler context cancelled")
			return
		case <-s.stopCh:
			s.logger.Info("retention cleanup scheduler stop signal received")
			return
		case <-ticker.C:
			s.executeCleanup(ctx)
		}
	}
}

// executeCleanup runs the cleanup operation
func (s *Scheduler) executeCleanup(ctx context.Context) (*CleanupResult, error) {
	s.logger.Info("starting scheduled retention cleanup")
	startTime := time.Now()

	result, err := s.service.CleanupAllTenants(ctx)
	if err != nil {
		s.logger.Error("retention cleanup failed",
			"error", err,
			"duration_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, err
	}

	s.logger.Info("retention cleanup completed",
		"executions_deleted", result.ExecutionsDeleted,
		"step_executions_deleted", result.StepExecutionsDeleted,
		"batches_processed", result.BatchesProcessed,
		"duration_ms", time.Since(startTime).Milliseconds(),
	)

	return result, nil
}
