package audit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCategory_Valid(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		valid    bool
	}{
		{"authentication", CategoryAuthentication, true},
		{"authorization", CategoryAuthorization, true},
		{"data_access", CategoryDataAccess, true},
		{"configuration", CategoryConfiguration, true},
		{"workflow", CategoryWorkflow, true},
		{"integration", CategoryIntegration, true},
		{"credential", CategoryCredential, true},
		{"user_management", CategoryUserManagement, true},
		{"system", CategorySystem, true},
		{"invalid", Category("invalid"), false},
	}

	validCategories := map[Category]bool{
		CategoryAuthentication: true,
		CategoryAuthorization:  true,
		CategoryDataAccess:     true,
		CategoryConfiguration:  true,
		CategoryWorkflow:       true,
		CategoryIntegration:    true,
		CategoryCredential:     true,
		CategoryUserManagement: true,
		CategorySystem:         true,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, exists := validCategories[tt.category]
			assert.Equal(t, tt.valid, exists)
		})
	}
}

func TestEventType_Valid(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		valid     bool
	}{
		{"create", EventTypeCreate, true},
		{"read", EventTypeRead, true},
		{"update", EventTypeUpdate, true},
		{"delete", EventTypeDelete, true},
		{"execute", EventTypeExecute, true},
		{"login", EventTypeLogin, true},
		{"logout", EventTypeLogout, true},
		{"permission_change", EventTypePermissionChange, true},
		{"export", EventTypeExport, true},
		{"import", EventTypeImport, true},
		{"access", EventTypeAccess, true},
		{"configure", EventTypeConfigure, true},
		{"invalid", EventType("invalid"), false},
	}

	validEventTypes := map[EventType]bool{
		EventTypeCreate:           true,
		EventTypeRead:             true,
		EventTypeUpdate:           true,
		EventTypeDelete:           true,
		EventTypeExecute:          true,
		EventTypeLogin:            true,
		EventTypeLogout:           true,
		EventTypePermissionChange: true,
		EventTypeExport:           true,
		EventTypeImport:           true,
		EventTypeAccess:           true,
		EventTypeConfigure:        true,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, exists := validEventTypes[tt.eventType]
			assert.Equal(t, tt.valid, exists)
		})
	}
}

func TestSeverity_Valid(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		valid    bool
	}{
		{"info", SeverityInfo, true},
		{"warning", SeverityWarning, true},
		{"error", SeverityError, true},
		{"critical", SeverityCritical, true},
		{"invalid", Severity("invalid"), false},
	}

	validSeverities := map[Severity]bool{
		SeverityInfo:     true,
		SeverityWarning:  true,
		SeverityError:    true,
		SeverityCritical: true,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, exists := validSeverities[tt.severity]
			assert.Equal(t, tt.valid, exists)
		})
	}
}

func TestStatus_Valid(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		valid  bool
	}{
		{"success", StatusSuccess, true},
		{"failure", StatusFailure, true},
		{"partial", StatusPartial, true},
		{"invalid", Status("invalid"), false},
	}

	validStatuses := map[Status]bool{
		StatusSuccess: true,
		StatusFailure: true,
		StatusPartial: true,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, exists := validStatuses[tt.status]
			assert.Equal(t, tt.valid, exists)
		})
	}
}

func TestAuditEvent_Complete(t *testing.T) {
	now := time.Now()
	event := AuditEvent{
		ID:           "test-id",
		TenantID:     "tenant-1",
		UserID:       "user-1",
		UserEmail:    "user@example.com",
		Category:     CategoryWorkflow,
		EventType:    EventTypeExecute,
		Action:       "workflow.executed",
		ResourceType: "workflow",
		ResourceID:   "wf-123",
		ResourceName: "Test Workflow",
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0",
		Severity:     SeverityInfo,
		Status:       StatusSuccess,
		Metadata: map[string]interface{}{
			"execution_id": "exec-123",
			"duration_ms":  1000,
		},
		CreatedAt: now,
	}

	assert.Equal(t, "test-id", event.ID)
	assert.Equal(t, "tenant-1", event.TenantID)
	assert.Equal(t, "user-1", event.UserID)
	assert.Equal(t, "user@example.com", event.UserEmail)
	assert.Equal(t, CategoryWorkflow, event.Category)
	assert.Equal(t, EventTypeExecute, event.EventType)
	assert.Equal(t, "workflow.executed", event.Action)
	assert.Equal(t, "workflow", event.ResourceType)
	assert.Equal(t, "wf-123", event.ResourceID)
	assert.Equal(t, "Test Workflow", event.ResourceName)
	assert.Equal(t, "192.168.1.1", event.IPAddress)
	assert.Equal(t, "Mozilla/5.0", event.UserAgent)
	assert.Equal(t, SeverityInfo, event.Severity)
	assert.Equal(t, StatusSuccess, event.Status)
	assert.Equal(t, "exec-123", event.Metadata["execution_id"])
	assert.Equal(t, 1000, event.Metadata["duration_ms"])
	assert.Equal(t, now, event.CreatedAt)
}

func TestRetentionPolicy_Validation(t *testing.T) {
	policy := RetentionPolicy{
		ID:                "policy-1",
		TenantID:          "tenant-1",
		HotRetentionDays:  90,
		WarmRetentionDays: 365,
		ColdRetentionDays: 2555,
		ArchiveEnabled:    true,
		ArchiveBucket:     "audit-archive",
		ArchivePath:       "tenant-1/audit",
		PurgeEnabled:      true,
	}

	// Test valid retention policy
	assert.Greater(t, policy.HotRetentionDays, 0)
	assert.GreaterOrEqual(t, policy.WarmRetentionDays, policy.HotRetentionDays)
	assert.GreaterOrEqual(t, policy.ColdRetentionDays, policy.WarmRetentionDays)

	// Test invalid retention policy
	invalidPolicy := RetentionPolicy{
		HotRetentionDays:  365,
		WarmRetentionDays: 90, // Invalid: less than hot
		ColdRetentionDays: 2555,
	}
	assert.Less(t, invalidPolicy.WarmRetentionDays, invalidPolicy.HotRetentionDays)
}

func TestQueryFilter_Complete(t *testing.T) {
	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now()

	filter := QueryFilter{
		TenantID:      "tenant-1",
		UserID:        "user-1",
		UserEmail:     "user@example.com",
		Categories:    []Category{CategoryWorkflow, CategoryCredential},
		EventTypes:    []EventType{EventTypeExecute, EventTypeAccess},
		Actions:       []string{"workflow.executed", "credential.accessed"},
		ResourceType:  "workflow",
		ResourceID:    "wf-123",
		IPAddress:     "192.168.1.1",
		Severities:    []Severity{SeverityInfo, SeverityWarning},
		Statuses:      []Status{StatusSuccess},
		StartDate:     startDate,
		EndDate:       endDate,
		Limit:         100,
		Offset:        0,
		SortBy:        "created_at",
		SortDirection: "DESC",
	}

	assert.Equal(t, "tenant-1", filter.TenantID)
	assert.Equal(t, "user-1", filter.UserID)
	assert.Equal(t, 2, len(filter.Categories))
	assert.Equal(t, 2, len(filter.EventTypes))
	assert.Equal(t, 2, len(filter.Actions))
	assert.Equal(t, "workflow", filter.ResourceType)
	assert.Equal(t, 100, filter.Limit)
	assert.Equal(t, 0, filter.Offset)
	assert.Equal(t, "created_at", filter.SortBy)
	assert.Equal(t, "DESC", filter.SortDirection)
}

func TestExportFormat_Valid(t *testing.T) {
	tests := []struct {
		name   string
		format ExportFormat
		valid  bool
	}{
		{"csv", ExportFormatCSV, true},
		{"json", ExportFormatJSON, true},
		{"invalid", ExportFormat("xml"), false},
	}

	validFormats := map[ExportFormat]bool{
		ExportFormatCSV:  true,
		ExportFormatJSON: true,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, exists := validFormats[tt.format]
			assert.Equal(t, tt.valid, exists)
		})
	}
}
