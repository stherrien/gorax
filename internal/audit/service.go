package audit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AuditRepository defines the interface for audit data access
type AuditRepository interface {
	CreateAuditEvent(ctx context.Context, event *AuditEvent) error
	CreateAuditEventBatch(ctx context.Context, events []*AuditEvent) error
	GetAuditEvent(ctx context.Context, tenantID, eventID string) (*AuditEvent, error)
	QueryAuditEvents(ctx context.Context, filter QueryFilter) ([]AuditEvent, int, error)
	GetAuditStats(ctx context.Context, tenantID string, timeRange TimeRange) (*AuditStats, error)
	GetRetentionPolicy(ctx context.Context, tenantID string) (*RetentionPolicy, error)
	UpdateRetentionPolicy(ctx context.Context, policy *RetentionPolicy) error
	DeleteOldAuditEvents(ctx context.Context, tenantID string, cutoffDate time.Time) (int64, error)
}

// Service handles audit logging business logic with async batching
type Service struct {
	repo       AuditRepository
	buffer     chan *AuditEvent
	bufferSize int
	flushTimer time.Duration
	done       chan struct{}
	wg         sync.WaitGroup
}

// NewService creates a new audit service with async batching
func NewService(repo AuditRepository, bufferSize int, flushTimer time.Duration) *Service {
	if bufferSize <= 0 {
		bufferSize = 100
	}
	if flushTimer <= 0 {
		flushTimer = 5 * time.Second
	}

	svc := &Service{
		repo:       repo,
		buffer:     make(chan *AuditEvent, bufferSize),
		bufferSize: bufferSize,
		flushTimer: flushTimer,
		done:       make(chan struct{}),
	}

	svc.wg.Add(1)
	go svc.processBatch()

	return svc
}

// LogEvent logs an audit event asynchronously
func (s *Service) LogEvent(ctx context.Context, event *AuditEvent) error {
	if err := validateEvent(event); err != nil {
		return err
	}

	select {
	case s.buffer <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Buffer is full, log synchronously
		return s.LogEventSync(ctx, event)
	}
}

// LogEventSync logs an audit event synchronously
func (s *Service) LogEventSync(ctx context.Context, event *AuditEvent) error {
	if err := validateEvent(event); err != nil {
		return err
	}

	if err := s.repo.CreateAuditEvent(ctx, event); err != nil {
		return fmt.Errorf("create audit event: %w", err)
	}

	return nil
}

// LogEventBatch logs multiple audit events in a batch
func (s *Service) LogEventBatch(ctx context.Context, events []*AuditEvent) error {
	for _, event := range events {
		if err := validateEvent(event); err != nil {
			return err
		}
	}

	if err := s.repo.CreateAuditEventBatch(ctx, events); err != nil {
		return fmt.Errorf("create audit event batch: %w", err)
	}

	return nil
}

// GetAuditEvent retrieves a single audit event
func (s *Service) GetAuditEvent(ctx context.Context, tenantID, eventID string) (*AuditEvent, error) {
	event, err := s.repo.GetAuditEvent(ctx, tenantID, eventID)
	if err != nil {
		return nil, fmt.Errorf("get audit event: %w", err)
	}

	return event, nil
}

// QueryAuditEvents queries audit events with filters
func (s *Service) QueryAuditEvents(ctx context.Context, filter QueryFilter) ([]AuditEvent, int, error) {
	events, total, err := s.repo.QueryAuditEvents(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("query audit events: %w", err)
	}

	return events, total, nil
}

// GetAuditStats retrieves aggregate statistics for audit logs
func (s *Service) GetAuditStats(ctx context.Context, tenantID string, timeRange TimeRange) (*AuditStats, error) {
	if err := validateTimeRange(timeRange); err != nil {
		return nil, err
	}

	stats, err := s.repo.GetAuditStats(ctx, tenantID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("get audit stats: %w", err)
	}

	return stats, nil
}

// GetRetentionPolicy retrieves the retention policy for a tenant
func (s *Service) GetRetentionPolicy(ctx context.Context, tenantID string) (*RetentionPolicy, error) {
	policy, err := s.repo.GetRetentionPolicy(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get retention policy: %w", err)
	}

	return policy, nil
}

// UpdateRetentionPolicy updates the retention policy for a tenant
func (s *Service) UpdateRetentionPolicy(ctx context.Context, policy *RetentionPolicy) error {
	if err := validateRetentionPolicy(policy); err != nil {
		return err
	}

	if err := s.repo.UpdateRetentionPolicy(ctx, policy); err != nil {
		return fmt.Errorf("update retention policy: %w", err)
	}

	return nil
}

// CleanupOldLogs deletes audit logs older than the retention policy
func (s *Service) CleanupOldLogs(ctx context.Context, tenantID string) (int64, error) {
	policy, err := s.repo.GetRetentionPolicy(ctx, tenantID)
	if err != nil {
		return 0, fmt.Errorf("get retention policy: %w", err)
	}

	if !policy.PurgeEnabled {
		return 0, nil
	}

	cutoffDate := time.Now().AddDate(0, 0, -policy.ColdRetentionDays)
	deletedCount, err := s.repo.DeleteOldAuditEvents(ctx, tenantID, cutoffDate)
	if err != nil {
		return 0, fmt.Errorf("delete old audit events: %w", err)
	}

	return deletedCount, nil
}

// Flush flushes all pending audit events
func (s *Service) Flush() {
	close(s.buffer)
	s.wg.Wait()

	// Recreate buffer for continued use
	s.buffer = make(chan *AuditEvent, s.bufferSize)
	s.wg.Add(1)
	go s.processBatch()
}

// Close gracefully shuts down the service
func (s *Service) Close() {
	close(s.done)
	close(s.buffer)
	s.wg.Wait()
}

// processBatch processes buffered audit events in batches
func (s *Service) processBatch() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.flushTimer)
	defer ticker.Stop()

	var batch []*AuditEvent

	flush := func() {
		if len(batch) == 0 {
			return
		}

		ctx := context.Background()
		if err := s.repo.CreateAuditEventBatch(ctx, batch); err != nil {
			// Log error but don't fail (audit logging should not break main flow)
			// In production, this would go to a dedicated error logging system
			fmt.Printf("ERROR: failed to insert audit event batch: %v\n", err)
		}

		batch = batch[:0]
	}

	for {
		select {
		case event, ok := <-s.buffer:
			if !ok {
				// Buffer closed, flush remaining events
				flush()
				return
			}

			batch = append(batch, event)

			// Flush when batch is full
			if len(batch) >= s.bufferSize {
				flush()
			}

		case <-ticker.C:
			// Flush on timer
			flush()

		case <-s.done:
			// Shutting down, flush remaining events
			flush()
			return
		}
	}
}

// validateEvent validates an audit event
func validateEvent(event *AuditEvent) error {
	if event.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	if event.Category == "" {
		return fmt.Errorf("category is required")
	}

	if event.EventType == "" {
		return fmt.Errorf("event_type is required")
	}

	if event.Action == "" {
		return fmt.Errorf("action is required")
	}

	if event.Severity == "" {
		return fmt.Errorf("severity is required")
	}

	if event.Status == "" {
		return fmt.Errorf("status is required")
	}

	return nil
}

// validateTimeRange validates a time range
func validateTimeRange(timeRange TimeRange) error {
	if timeRange.StartDate.IsZero() || timeRange.EndDate.IsZero() {
		return fmt.Errorf("start date and end date are required")
	}

	if timeRange.EndDate.Before(timeRange.StartDate) {
		return fmt.Errorf("end date must be after start date")
	}

	return nil
}

// validateRetentionPolicy validates a retention policy
func validateRetentionPolicy(policy *RetentionPolicy) error {
	if policy.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	if policy.HotRetentionDays <= 0 {
		return fmt.Errorf("hot_retention_days must be positive")
	}

	if policy.WarmRetentionDays < policy.HotRetentionDays {
		return fmt.Errorf("warm_retention_days must be >= hot_retention_days")
	}

	if policy.ColdRetentionDays < policy.WarmRetentionDays {
		return fmt.Errorf("cold_retention_days must be >= warm_retention_days")
	}

	return nil
}
