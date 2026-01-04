package audit

import "time"

// Category represents the category of an audit event
type Category string

const (
	CategoryAuthentication Category = "authentication"
	CategoryAuthorization  Category = "authorization"
	CategoryDataAccess     Category = "data_access"
	CategoryConfiguration  Category = "configuration"
	CategoryWorkflow       Category = "workflow"
	CategoryIntegration    Category = "integration"
	CategoryCredential     Category = "credential"
	CategoryUserManagement Category = "user_management"
	CategorySystem         Category = "system"
)

// EventType represents the type of audit event
type EventType string

const (
	EventTypeCreate           EventType = "create"
	EventTypeRead             EventType = "read"
	EventTypeUpdate           EventType = "update"
	EventTypeDelete           EventType = "delete"
	EventTypeExecute          EventType = "execute"
	EventTypeLogin            EventType = "login"
	EventTypeLogout           EventType = "logout"
	EventTypePermissionChange EventType = "permission_change"
	EventTypeExport           EventType = "export"
	EventTypeImport           EventType = "import"
	EventTypeAccess           EventType = "access"
	EventTypeConfigure        EventType = "configure"
)

// Severity represents the severity level of an audit event
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// Status represents the outcome of an audit event
type Status string

const (
	StatusSuccess Status = "success"
	StatusFailure Status = "failure"
	StatusPartial Status = "partial"
)

// AuditEvent represents a single audit log entry
type AuditEvent struct {
	ID           string                 `db:"id" json:"id"`
	TenantID     string                 `db:"tenant_id" json:"tenantId"`
	UserID       string                 `db:"user_id" json:"userId"`
	UserEmail    string                 `db:"user_email" json:"userEmail"`
	Category     Category               `db:"category" json:"category"`
	EventType    EventType              `db:"event_type" json:"eventType"`
	Action       string                 `db:"action" json:"action"`
	ResourceType string                 `db:"resource_type" json:"resourceType"`
	ResourceID   string                 `db:"resource_id" json:"resourceId"`
	ResourceName string                 `db:"resource_name" json:"resourceName"`
	IPAddress    string                 `db:"ip_address" json:"ipAddress"`
	UserAgent    string                 `db:"user_agent" json:"userAgent"`
	Severity     Severity               `db:"severity" json:"severity"`
	Status       Status                 `db:"status" json:"status"`
	ErrorMessage string                 `db:"error_message" json:"errorMessage,omitempty"`
	Metadata     map[string]interface{} `db:"metadata" json:"metadata"`
	CreatedAt    time.Time              `db:"created_at" json:"createdAt"`
}

// RetentionPolicy represents audit log retention configuration
type RetentionPolicy struct {
	ID                string    `db:"id" json:"id"`
	TenantID          string    `db:"tenant_id" json:"tenantId"`
	HotRetentionDays  int       `db:"hot_retention_days" json:"hotRetentionDays"`
	WarmRetentionDays int       `db:"warm_retention_days" json:"warmRetentionDays"`
	ColdRetentionDays int       `db:"cold_retention_days" json:"coldRetentionDays"`
	ArchiveEnabled    bool      `db:"archive_enabled" json:"archiveEnabled"`
	ArchiveBucket     string    `db:"archive_bucket" json:"archiveBucket"`
	ArchivePath       string    `db:"archive_path" json:"archivePath"`
	PurgeEnabled      bool      `db:"purge_enabled" json:"purgeEnabled"`
	LastArchiveAt     time.Time `db:"last_archive_at" json:"lastArchiveAt"`
	LastPurgeAt       time.Time `db:"last_purge_at" json:"lastPurgeAt"`
	CreatedAt         time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt         time.Time `db:"updated_at" json:"updatedAt"`
}

// IntegrityRecord represents a daily integrity hash for tamper detection
type IntegrityRecord struct {
	ID         string    `db:"id" json:"id"`
	TenantID   string    `db:"tenant_id" json:"tenantId"`
	Date       time.Time `db:"date" json:"date"`
	EventCount int       `db:"event_count" json:"eventCount"`
	Hash       string    `db:"hash" json:"hash"`
	CreatedAt  time.Time `db:"created_at" json:"createdAt"`
}

// QueryFilter represents filters for querying audit logs
type QueryFilter struct {
	TenantID      string
	UserID        string
	UserEmail     string
	Categories    []Category
	EventTypes    []EventType
	Actions       []string
	ResourceType  string
	ResourceID    string
	IPAddress     string
	Severities    []Severity
	Statuses      []Status
	StartDate     time.Time
	EndDate       time.Time
	Limit         int
	Offset        int
	SortBy        string
	SortDirection string
}

// AuditStats represents aggregate statistics for audit logs
type AuditStats struct {
	TotalEvents      int              `json:"totalEvents"`
	EventsByCategory map[Category]int `json:"eventsByCategory"`
	EventsBySeverity map[Severity]int `json:"eventsBySeverity"`
	EventsByStatus   map[Status]int   `json:"eventsByStatus"`
	TopUsers         []UserActivity   `json:"topUsers"`
	TopActions       []ActionCount    `json:"topActions"`
	CriticalEvents   int              `json:"criticalEvents"`
	FailedEvents     int              `json:"failedEvents"`
	RecentCritical   []AuditEvent     `json:"recentCritical"`
	TimeRange        TimeRange        `json:"timeRange"`
}

// UserActivity represents activity summary for a user
type UserActivity struct {
	UserID     string `db:"user_id" json:"userId"`
	UserEmail  string `db:"user_email" json:"userEmail"`
	EventCount int    `db:"event_count" json:"eventCount"`
}

// ActionCount represents count for a specific action
type ActionCount struct {
	Action string `db:"action" json:"action"`
	Count  int    `db:"count" json:"count"`
}

// TimeRange represents a time range for queries
type TimeRange struct {
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

// ExportFormat represents the format for exporting audit logs
type ExportFormat string

const (
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatJSON ExportFormat = "json"
)

// ComplianceReport represents a compliance report
type ComplianceReport struct {
	ReportType      string         `json:"reportType"`
	TenantID        string         `json:"tenantId"`
	TimeRange       TimeRange      `json:"timeRange"`
	TotalEvents     int            `json:"totalEvents"`
	DataAccessLogs  []AuditEvent   `json:"dataAccessLogs"`
	UserActivityLog []UserActivity `json:"userActivityLog"`
	SecurityEvents  []AuditEvent   `json:"securityEvents"`
	GeneratedAt     time.Time      `json:"generatedAt"`
}
