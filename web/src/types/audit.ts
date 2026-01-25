/**
 * Audit log types for compliance and security monitoring
 */

/**
 * Category represents the category of an audit event
 */
export enum AuditCategory {
  Authentication = 'authentication',
  Authorization = 'authorization',
  DataAccess = 'data_access',
  Configuration = 'configuration',
  Workflow = 'workflow',
  Integration = 'integration',
  Credential = 'credential',
  UserManagement = 'user_management',
  System = 'system',
}

/**
 * EventType represents the type of audit event
 */
export enum AuditEventType {
  Create = 'create',
  Read = 'read',
  Update = 'update',
  Delete = 'delete',
  Execute = 'execute',
  Login = 'login',
  Logout = 'logout',
  PermissionChange = 'permission_change',
  Export = 'export',
  Import = 'import',
  Access = 'access',
  Configure = 'configure',
}

/**
 * Severity represents the severity level of an audit event
 */
export enum AuditSeverity {
  Info = 'info',
  Warning = 'warning',
  Error = 'error',
  Critical = 'critical',
}

/**
 * Status represents the outcome of an audit event
 */
export enum AuditStatus {
  Success = 'success',
  Failure = 'failure',
  Partial = 'partial',
}

/**
 * AuditEvent represents a single audit log entry
 */
export interface AuditEvent {
  id: string
  tenantId: string
  userId: string
  userEmail: string
  category: AuditCategory
  eventType: AuditEventType
  action: string
  resourceType: string
  resourceId: string
  resourceName: string
  ipAddress: string
  userAgent: string
  severity: AuditSeverity
  status: AuditStatus
  errorMessage?: string
  metadata: Record<string, any>
  createdAt: string
}

/**
 * TimeRange represents a time range for queries
 */
export interface TimeRange {
  startDate: string
  endDate: string
}

/**
 * QueryFilter represents filters for querying audit logs
 */
export interface QueryFilter {
  tenantId?: string
  userId?: string
  userEmail?: string
  categories?: AuditCategory[]
  eventTypes?: AuditEventType[]
  actions?: string[]
  resourceType?: string
  resourceId?: string
  ipAddress?: string
  severities?: AuditSeverity[]
  statuses?: AuditStatus[]
  startDate?: string
  endDate?: string
  limit?: number
  offset?: number
  sortBy?: string
  sortDirection?: 'ASC' | 'DESC'
}

/**
 * UserActivity represents activity summary for a user
 */
export interface UserActivity {
  userId: string
  userEmail: string
  eventCount: number
}

/**
 * ActionCount represents count for a specific action
 */
export interface ActionCount {
  action: string
  count: number
}

/**
 * AuditStats represents aggregate statistics for audit logs
 */
export interface AuditStats {
  totalEvents: number
  eventsByCategory: Record<AuditCategory, number>
  eventsBySeverity: Record<AuditSeverity, number>
  eventsByStatus: Record<AuditStatus, number>
  topUsers: UserActivity[]
  topActions: ActionCount[]
  criticalEvents: number
  failedEvents: number
  recentCritical: AuditEvent[]
  timeRange: TimeRange
}

/**
 * QueryResponse represents paginated query response
 */
export interface QueryResponse {
  events: AuditEvent[]
  total: number
  page: number
  limit: number
}

/**
 * ExportFormat represents the format for exporting audit logs
 */
export enum ExportFormat {
  CSV = 'csv',
  JSON = 'json',
}

/**
 * ExportRequest represents a request to export audit logs
 */
export interface ExportRequest {
  filter: QueryFilter
  format: ExportFormat
}

/**
 * Predefined time range options
 */
export enum TimeRangePreset {
  Last24Hours = 'last_24_hours',
  Last7Days = 'last_7_days',
  Last30Days = 'last_30_days',
  Last90Days = 'last_90_days',
  Custom = 'custom',
}

/**
 * Helper to get time range from preset
 */
export function getTimeRangeFromPreset(preset: TimeRangePreset): TimeRange | null {
  const now = new Date()
  const endDate = now.toISOString()

  switch (preset) {
    case TimeRangePreset.Last24Hours: {
      const startDate = new Date(now.getTime() - 24 * 60 * 60 * 1000).toISOString()
      return { startDate, endDate }
    }
    case TimeRangePreset.Last7Days: {
      const startDate = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000).toISOString()
      return { startDate, endDate }
    }
    case TimeRangePreset.Last30Days: {
      const startDate = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000).toISOString()
      return { startDate, endDate }
    }
    case TimeRangePreset.Last90Days: {
      const startDate = new Date(now.getTime() - 90 * 24 * 60 * 60 * 1000).toISOString()
      return { startDate, endDate }
    }
    case TimeRangePreset.Custom:
      return null
    default:
      return null
  }
}

/**
 * Category display labels
 */
export const CATEGORY_LABELS: Record<AuditCategory, string> = {
  [AuditCategory.Authentication]: 'Authentication',
  [AuditCategory.Authorization]: 'Authorization',
  [AuditCategory.DataAccess]: 'Data Access',
  [AuditCategory.Configuration]: 'Configuration',
  [AuditCategory.Workflow]: 'Workflow',
  [AuditCategory.Integration]: 'Integration',
  [AuditCategory.Credential]: 'Credential',
  [AuditCategory.UserManagement]: 'User Management',
  [AuditCategory.System]: 'System',
}

/**
 * Event type display labels
 */
export const EVENT_TYPE_LABELS: Record<AuditEventType, string> = {
  [AuditEventType.Create]: 'Create',
  [AuditEventType.Read]: 'Read',
  [AuditEventType.Update]: 'Update',
  [AuditEventType.Delete]: 'Delete',
  [AuditEventType.Execute]: 'Execute',
  [AuditEventType.Login]: 'Login',
  [AuditEventType.Logout]: 'Logout',
  [AuditEventType.PermissionChange]: 'Permission Change',
  [AuditEventType.Export]: 'Export',
  [AuditEventType.Import]: 'Import',
  [AuditEventType.Access]: 'Access',
  [AuditEventType.Configure]: 'Configure',
}

/**
 * Severity display labels
 */
export const SEVERITY_LABELS: Record<AuditSeverity, string> = {
  [AuditSeverity.Info]: 'Info',
  [AuditSeverity.Warning]: 'Warning',
  [AuditSeverity.Error]: 'Error',
  [AuditSeverity.Critical]: 'Critical',
}

/**
 * Status display labels
 */
export const STATUS_LABELS: Record<AuditStatus, string> = {
  [AuditStatus.Success]: 'Success',
  [AuditStatus.Failure]: 'Failure',
  [AuditStatus.Partial]: 'Partial',
}

/**
 * Severity color mapping for UI
 */
export const SEVERITY_COLORS: Record<AuditSeverity, string> = {
  [AuditSeverity.Info]: 'blue',
  [AuditSeverity.Warning]: 'yellow',
  [AuditSeverity.Error]: 'orange',
  [AuditSeverity.Critical]: 'red',
}

/**
 * Status color mapping for UI
 */
export const STATUS_COLORS: Record<AuditStatus, string> = {
  [AuditStatus.Success]: 'green',
  [AuditStatus.Failure]: 'red',
  [AuditStatus.Partial]: 'yellow',
}
