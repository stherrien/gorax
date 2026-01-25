/**
 * Management type definitions for workflow management, monitoring, and admin interfaces
 */

// ============================================================================
// User Types
// ============================================================================

export type UserRole = 'admin' | 'operator' | 'viewer'
export type UserStatus = 'active' | 'inactive' | 'pending' | 'suspended'

export interface User {
  id: string
  tenantId: string
  email: string
  name: string
  role: UserRole
  status: UserStatus
  avatar?: string
  lastLoginAt?: string
  createdAt: string
  updatedAt: string
}

export interface UserCreateInput {
  email: string
  name: string
  role: UserRole
  sendInvite?: boolean
}

export interface UserUpdateInput {
  name?: string
  role?: UserRole
  status?: UserStatus
}

export interface UserListParams {
  page?: number
  limit?: number
  role?: UserRole
  status?: UserStatus
  search?: string
}

export interface UserListResponse {
  users: User[]
  total: number
}

// ============================================================================
// Tenant Types
// ============================================================================

export type TenantStatus = 'active' | 'suspended' | 'trial' | 'cancelled'
export type TenantPlan = 'free' | 'starter' | 'professional' | 'enterprise'

export interface TenantLimits {
  maxWorkflows: number
  maxExecutionsPerMonth: number
  maxUsers: number
  maxCredentials: number
  retentionDays: number
}

export interface TenantUsage {
  workflowCount: number
  executionsThisMonth: number
  userCount: number
  credentialCount: number
  storageBytes: number
}

export interface Tenant {
  id: string
  name: string
  slug: string
  status: TenantStatus
  plan: TenantPlan
  limits: TenantLimits
  usage: TenantUsage
  ownerId: string
  createdAt: string
  updatedAt: string
}

export interface TenantCreateInput {
  name: string
  slug: string
  plan: TenantPlan
  ownerEmail: string
}

export interface TenantUpdateInput {
  name?: string
  status?: TenantStatus
  plan?: TenantPlan
  limits?: Partial<TenantLimits>
}

export interface TenantListParams {
  page?: number
  limit?: number
  status?: TenantStatus
  plan?: TenantPlan
  search?: string
}

export interface TenantListResponse {
  tenants: Tenant[]
  total: number
}

// ============================================================================
// Execution Log Types
// ============================================================================

export type LogLevel = 'debug' | 'info' | 'warn' | 'error'

export interface ExecutionLog {
  id: string
  executionId: string
  nodeId?: string
  nodeName?: string
  level: LogLevel
  message: string
  data?: Record<string, unknown>
  timestamp: string
}

export interface ExecutionLogListParams {
  executionId: string
  nodeId?: string
  level?: LogLevel | LogLevel[]
  search?: string
  limit?: number
  offset?: number
}

export interface ExecutionLogListResponse {
  logs: ExecutionLog[]
  total: number
}

// ============================================================================
// System Health Types
// ============================================================================

export type HealthStatus = 'healthy' | 'degraded' | 'unhealthy'

export interface ServiceHealth {
  name: string
  status: HealthStatus
  responseTime: number
  lastCheck: string
  message?: string
}

export interface SystemHealth {
  overall: HealthStatus
  services: ServiceHealth[]
  uptime: number
  lastUpdated: string
}

// ============================================================================
// Monitoring Stats Types
// ============================================================================

export interface MonitoringStats {
  activeExecutions: number
  queuedExecutions: number
  executionsPerMinute: number
  averageExecutionTime: number
  successRate: number
  errorRate: number
  lastHourExecutions: number
  lastHourFailures: number
}

export interface WorkflowStats {
  workflowId: string
  workflowName: string
  totalExecutions: number
  successfulExecutions: number
  failedExecutions: number
  averageDuration: number
  lastExecutedAt?: string
}

// ============================================================================
// Filter Types
// ============================================================================

export interface DateRange {
  startDate: string
  endDate: string
}

export interface PaginationParams {
  page: number
  limit: number
}

export interface SortParams {
  sortBy: string
  sortDirection: 'asc' | 'desc'
}

// ============================================================================
// Bulk Operation Types
// ============================================================================

export interface BulkOperationResult {
  successCount: number
  failureCount: number
  failures: Array<{
    id: string
    error: string
  }>
}
