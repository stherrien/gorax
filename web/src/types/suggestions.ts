// Error categories for suggestions
export type ErrorCategory =
  | 'network'
  | 'auth'
  | 'data'
  | 'rate_limit'
  | 'timeout'
  | 'config'
  | 'external_service'
  | 'unknown'

// Suggestion types
export type SuggestionType =
  | 'retry'
  | 'config_change'
  | 'credential_update'
  | 'data_fix'
  | 'workflow_modification'
  | 'manual_intervention'

// Confidence levels
export type SuggestionConfidence = 'high' | 'medium' | 'low'

// Suggestion status
export type SuggestionStatus = 'pending' | 'applied' | 'dismissed'

// Suggestion source
export type SuggestionSource = 'pattern' | 'llm'

// Retry configuration
export interface RetryConfig {
  max_retries: number
  backoff_ms: number
  backoff_factor: number
}

// Suggestion fix data
export interface SuggestionFix {
  action_type: string
  config_path?: string
  old_value?: unknown
  new_value?: unknown
  retry_config?: RetryConfig
  action_data?: unknown
}

// Main Suggestion interface
export interface Suggestion {
  id: string
  tenant_id: string
  execution_id: string
  node_id: string
  category: ErrorCategory
  type: SuggestionType
  confidence: SuggestionConfidence
  title: string
  description: string
  details?: string
  fix?: SuggestionFix
  source: SuggestionSource
  status: SuggestionStatus
  created_at: string
  applied_at?: string
  dismissed_at?: string
}

// Error context for analysis
export interface ErrorContext {
  execution_id: string
  workflow_id: string
  node_id: string
  node_type: string
  error_message: string
  error_code?: string
  http_status?: number
  retry_count?: number
  input_data?: Record<string, unknown>
  node_config?: Record<string, unknown>
}

// Analyze request body
export interface AnalyzeRequest {
  workflow_id: string
  node_id: string
  node_type: string
  error_message: string
  error_code?: string
  http_status?: number
  retry_count?: number
  input_data?: Record<string, unknown>
  node_config?: Record<string, unknown>
}

// API response types
export interface SuggestionsListResponse {
  data: Suggestion[]
}

export interface SuggestionResponse {
  data: Suggestion
}

export interface AnalyzeResponse {
  data: Suggestion[]
}

export interface SuggestionActionResponse {
  message: string
}

// Stats for suggestions
export interface SuggestionStats {
  total: number
  pending: number
  applied: number
  dismissed: number
  by_source: Record<string, number>
  by_confidence: Record<string, number>
}

// Helper function to get category display label
export function getCategoryLabel(category: ErrorCategory): string {
  const labels: Record<ErrorCategory, string> = {
    network: 'Network',
    auth: 'Authentication',
    data: 'Data',
    rate_limit: 'Rate Limit',
    timeout: 'Timeout',
    config: 'Configuration',
    external_service: 'External Service',
    unknown: 'Unknown',
  }
  return labels[category] || category
}

// Helper function to get type display label
export function getTypeLabel(type: SuggestionType): string {
  const labels: Record<SuggestionType, string> = {
    retry: 'Retry',
    config_change: 'Configuration Change',
    credential_update: 'Update Credentials',
    data_fix: 'Fix Data',
    workflow_modification: 'Modify Workflow',
    manual_intervention: 'Manual Intervention',
  }
  return labels[type] || type
}

// Helper function to get confidence display label
export function getConfidenceLabel(confidence: SuggestionConfidence): string {
  const labels: Record<SuggestionConfidence, string> = {
    high: 'High',
    medium: 'Medium',
    low: 'Low',
  }
  return labels[confidence] || confidence
}

// Helper function to get status display label
export function getStatusLabel(status: SuggestionStatus): string {
  const labels: Record<SuggestionStatus, string> = {
    pending: 'Pending',
    applied: 'Applied',
    dismissed: 'Dismissed',
  }
  return labels[status] || status
}

// Helper function to get category color
export function getCategoryColor(category: ErrorCategory): string {
  const colors: Record<ErrorCategory, string> = {
    network: 'blue',
    auth: 'red',
    data: 'orange',
    rate_limit: 'yellow',
    timeout: 'purple',
    config: 'cyan',
    external_service: 'pink',
    unknown: 'gray',
  }
  return colors[category] || 'gray'
}

// Helper function to get confidence color
export function getConfidenceColor(confidence: SuggestionConfidence): string {
  const colors: Record<SuggestionConfidence, string> = {
    high: 'green',
    medium: 'yellow',
    low: 'red',
  }
  return colors[confidence] || 'gray'
}

// Helper function to get status color
export function getStatusColor(status: SuggestionStatus): string {
  const colors: Record<SuggestionStatus, string> = {
    pending: 'blue',
    applied: 'green',
    dismissed: 'gray',
  }
  return colors[status] || 'gray'
}
