package retention

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// Repository defines the interface for retention data operations
type Repository interface {
	GetRetentionPolicy(ctx context.Context, tenantID string) (*RetentionPolicy, error)
	DeleteOldExecutions(ctx context.Context, tenantID string, cutoffDate time.Time, batchSize int) (*CleanupResult, error)
	ArchiveAndDeleteOldExecutions(ctx context.Context, tenantID string, cutoffDate time.Time, batchSize int) (*CleanupResult, error)
	GetTenantsWithRetention(ctx context.Context) ([]string, error)
	LogCleanup(ctx context.Context, log *CleanupLog) error
}

// Service handles retention policy operations
type Service struct {
	repo   Repository
	logger *slog.Logger
	config Config
}

// NewService creates a new retention service
func NewService(repo Repository, logger *slog.Logger, config Config) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
		config: config,
	}
}

// GetRetentionPolicy retrieves the retention policy for a tenant
// Returns default policy if tenant doesn't have one configured
func (s *Service) GetRetentionPolicy(ctx context.Context, tenantID string) (*RetentionPolicy, error) {
	policy, err := s.repo.GetRetentionPolicy(ctx, tenantID)
	if err != nil {
		if err == ErrNotFound {
			// Return default policy if not configured
			return &RetentionPolicy{
				TenantID:      tenantID,
				RetentionDays: s.config.DefaultRetentionDays,
				Enabled:       true,
			}, nil
		}
		return nil, fmt.Errorf("failed to get retention policy: %w", err)
	}
	return policy, nil
}

// CleanupOldExecutions deletes old executions for a tenant based on retention policy
func (s *Service) CleanupOldExecutions(ctx context.Context, tenantID string) (*CleanupResult, error) {
	startTime := time.Now()

	// Get retention policy
	policy, err := s.GetRetentionPolicy(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get retention policy: %w", err)
	}

	// Skip if retention is disabled
	if !policy.Enabled {
		s.logger.Info("retention disabled for tenant, skipping cleanup",
			"tenant_id", tenantID,
		)
		return &CleanupResult{
			ExecutionsDeleted:     0,
			StepExecutionsDeleted: 0,
			ExecutionsArchived:    0,
			BatchesProcessed:      0,
		}, nil
	}

	// Calculate cutoff date
	cutoffDate := s.calculateCutoffDate(time.Now(), policy.RetentionDays)

	s.logger.Info("starting cleanup for tenant",
		"tenant_id", tenantID,
		"retention_days", policy.RetentionDays,
		"cutoff_date", cutoffDate,
		"archive_enabled", s.config.ArchiveBeforeDelete,
	)

	// Archive and/or delete old executions based on configuration
	var result *CleanupResult
	if s.config.ArchiveBeforeDelete {
		result, err = s.repo.ArchiveAndDeleteOldExecutions(ctx, tenantID, cutoffDate, s.config.BatchSize)
	} else {
		result, err = s.repo.DeleteOldExecutions(ctx, tenantID, cutoffDate, s.config.BatchSize)
	}
	if err != nil {
		// Log failure
		if s.config.EnableAuditLog {
			errorMsg := err.Error()
			logEntry := &CleanupLog{
				ID:                    uuid.New().String(),
				TenantID:              tenantID,
				ExecutionsDeleted:     0,
				ExecutionsArchived:    0,
				StepExecutionsDeleted: 0,
				RetentionDays:         policy.RetentionDays,
				CutoffDate:            cutoffDate,
				DurationMs:            int(time.Since(startTime).Milliseconds()),
				Status:                "failed",
				ErrorMessage:          &errorMsg,
				CreatedAt:             time.Now(),
			}
			if logErr := s.repo.LogCleanup(ctx, logEntry); logErr != nil {
				s.logger.Error("failed to log cleanup failure", "error", logErr)
			}
		}
		return nil, fmt.Errorf("failed to delete old executions: %w", err)
	}

	duration := time.Since(startTime)

	s.logger.Info("cleanup completed",
		"tenant_id", tenantID,
		"executions_deleted", result.ExecutionsDeleted,
		"executions_archived", result.ExecutionsArchived,
		"step_executions_deleted", result.StepExecutionsDeleted,
		"batches_processed", result.BatchesProcessed,
		"duration_ms", duration.Milliseconds(),
	)

	// Log success
	if s.config.EnableAuditLog {
		logEntry := &CleanupLog{
			ID:                    uuid.New().String(),
			TenantID:              tenantID,
			ExecutionsDeleted:     result.ExecutionsDeleted,
			ExecutionsArchived:    result.ExecutionsArchived,
			StepExecutionsDeleted: result.StepExecutionsDeleted,
			RetentionDays:         policy.RetentionDays,
			CutoffDate:            cutoffDate,
			DurationMs:            int(duration.Milliseconds()),
			Status:                "completed",
			CreatedAt:             time.Now(),
		}
		if err := s.repo.LogCleanup(ctx, logEntry); err != nil {
			s.logger.Error("failed to log cleanup success", "error", err)
		}
	}

	return result, nil
}

// CleanupAllTenants runs cleanup for all tenants with retention enabled
func (s *Service) CleanupAllTenants(ctx context.Context) (*CleanupResult, error) {
	s.logger.Info("starting cleanup for all tenants")

	// Get all tenants
	tenants, err := s.repo.GetTenantsWithRetention(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenants: %w", err)
	}

	s.logger.Info("found tenants for cleanup", "count", len(tenants))

	// Aggregate results
	totalResult := &CleanupResult{
		ExecutionsDeleted:     0,
		StepExecutionsDeleted: 0,
		ExecutionsArchived:    0,
		BatchesProcessed:      0,
	}

	// Process each tenant
	for _, tenantID := range tenants {
		result, err := s.CleanupOldExecutions(ctx, tenantID)
		if err != nil {
			s.logger.Error("failed to cleanup tenant",
				"tenant_id", tenantID,
				"error", err,
			)
			// Continue with other tenants
			continue
		}

		// Aggregate results
		totalResult.ExecutionsDeleted += result.ExecutionsDeleted
		totalResult.StepExecutionsDeleted += result.StepExecutionsDeleted
		totalResult.ExecutionsArchived += result.ExecutionsArchived
		totalResult.BatchesProcessed += result.BatchesProcessed
	}

	s.logger.Info("completed cleanup for all tenants",
		"total_executions_deleted", totalResult.ExecutionsDeleted,
		"total_executions_archived", totalResult.ExecutionsArchived,
		"total_step_executions_deleted", totalResult.StepExecutionsDeleted,
		"total_batches", totalResult.BatchesProcessed,
	)

	return totalResult, nil
}

// calculateCutoffDate calculates the cutoff date based on retention days
func (s *Service) calculateCutoffDate(now time.Time, retentionDays int) time.Time {
	return now.Add(-time.Duration(retentionDays) * 24 * time.Hour)
}
