/**
 * Standardized API response types.
 *
 * These types match the backend response format from internal/api/response/response.go
 */

// --- Error Response Types ---

/**
 * Error codes matching backend response.ErrorCode
 */
export type APIErrorCode =
  | 'validation_error'
  | 'not_found'
  | 'bad_request'
  | 'unauthorized'
  | 'forbidden'
  | 'internal_error'
  | 'conflict'
  | 'rate_limit_exceeded'

/**
 * Standard API error response
 */
export interface APIErrorResponse {
  error: string
  code: APIErrorCode
  details?: Record<string, string>
}

/**
 * Type guard to check if response is an error
 */
export function isAPIError(response: unknown): response is APIErrorResponse {
  return (
    typeof response === 'object' &&
    response !== null &&
    'error' in response &&
    'code' in response
  )
}

// --- Success Response Types ---

/**
 * Data wrapper response (single item)
 */
export interface DataResponse<T> {
  data: T
}

/**
 * Paginated response with metadata
 */
export interface PaginatedResponse<T> {
  data: T[]
  limit: number
  offset: number
  total?: number
}

/**
 * Type guard to check if response has data wrapper
 */
export function hasDataWrapper<T>(response: unknown): response is DataResponse<T> {
  return (
    typeof response === 'object' &&
    response !== null &&
    'data' in response
  )
}

/**
 * Type guard to check if response is paginated
 */
export function isPaginatedResponse<T>(response: unknown): response is PaginatedResponse<T> {
  return (
    typeof response === 'object' &&
    response !== null &&
    'data' in response &&
    Array.isArray((response as PaginatedResponse<T>).data) &&
    'limit' in response &&
    'offset' in response
  )
}

/**
 * Unwrap data from response
 * Handles both wrapped {data: T} and direct T responses
 */
export function unwrapData<T>(response: T | DataResponse<T>): T {
  if (hasDataWrapper<T>(response)) {
    return response.data
  }
  return response
}

/**
 * Unwrap paginated response
 * Returns the array of items from the data field
 */
export function unwrapPaginated<T>(response: PaginatedResponse<T>): T[] {
  return response.data
}

// --- Request Types ---

/**
 * Standard pagination parameters
 */
export interface PaginationParams {
  limit?: number
  offset?: number
}

/**
 * Standard sort parameters
 */
export interface SortParams {
  sort?: string
  order?: 'asc' | 'desc'
}

/**
 * Combined list parameters
 */
export interface ListParams extends PaginationParams, SortParams {
  search?: string
}

// --- Common Entity Types ---

/**
 * Base entity with ID and timestamps
 */
export interface BaseEntity {
  id: string
  created_at: string
  updated_at: string
}

/**
 * Entity with tenant isolation
 */
export interface TenantEntity extends BaseEntity {
  tenant_id: string
}

// --- Schedule Types ---

export interface Schedule extends TenantEntity {
  workflow_id: string
  name: string
  cron_expression: string
  timezone: string
  enabled: boolean
  next_run_at: string | null
  last_run_at: string | null
}

export interface CreateScheduleInput {
  name: string
  cron_expression: string
  timezone?: string
  enabled?: boolean
}

export interface UpdateScheduleInput {
  name?: string
  cron_expression?: string
  timezone?: string
  enabled?: boolean
}

export interface ParseCronResponse {
  valid: boolean
  next_run: string
}

export interface PreviewScheduleResponse {
  valid: boolean
  next_runs: string[]
  count: number
  timezone: string
}

// --- Workflow Types ---

export type WorkflowStatus = 'draft' | 'published' | 'archived'

export interface Workflow extends TenantEntity {
  name: string
  description: string
  version: number
  status: WorkflowStatus
}

export interface CreateWorkflowInput {
  name: string
  description?: string
}

export interface UpdateWorkflowInput {
  name?: string
  description?: string
  status?: WorkflowStatus
}

// --- Execution Types ---

export type ExecutionStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'

export interface Execution extends TenantEntity {
  workflow_id: string
  schedule_id?: string
  status: ExecutionStatus
  started_at: string
  completed_at: string | null
  error: string | null
  trigger_type: 'manual' | 'scheduled' | 'webhook'
}

// --- User Types ---

export type UserRole = 'admin' | 'user' | 'viewer'

export interface User extends TenantEntity {
  email: string
  name: string
  role: UserRole
}

// --- Credential Types ---

export interface Credential extends TenantEntity {
  name: string
  type: string
  description: string
  last_used_at: string | null
}

export interface CreateCredentialInput {
  name: string
  type: string
  description?: string
  value: Record<string, string>
}

// --- Webhook Types ---

export interface WebhookEndpoint extends TenantEntity {
  workflow_id: string
  name: string
  path: string
  method: 'GET' | 'POST' | 'PUT' | 'DELETE'
  enabled: boolean
  secret_token?: string
}

// --- CSRF Token ---

export interface CSRFTokenResponse {
  token: string
}

// --- Health Check ---

export interface HealthResponse {
  status: 'healthy' | 'degraded' | 'unhealthy'
  version?: string
  timestamp: string
}
