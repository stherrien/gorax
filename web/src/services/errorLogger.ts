import * as Sentry from '@sentry/react'

/**
 * Error categories for classification and routing
 */
export type ErrorCategory =
  | 'render'      // React component render errors
  | 'network'     // API/network errors
  | 'validation'  // Form/data validation errors
  | 'state'       // State management errors
  | 'canvas'      // Workflow canvas specific errors
  | 'node'        // Node-related errors
  | 'unknown'     // Uncategorized errors

/**
 * Error severity levels
 */
export type ErrorSeverity = 'fatal' | 'error' | 'warning' | 'info'

/**
 * Context information for error logging
 */
export interface ErrorContext {
  componentName?: string
  componentStack?: string
  workflowId?: string
  nodeId?: string
  userId?: string
  tenantId?: string
  action?: string
  additionalData?: Record<string, unknown>
}

/**
 * Structured error log entry
 */
export interface ErrorLogEntry {
  id: string
  timestamp: string
  category: ErrorCategory
  severity: ErrorSeverity
  message: string
  stack?: string
  context: ErrorContext
  browserInfo: BrowserInfo
}

/**
 * Browser information for debugging
 */
interface BrowserInfo {
  userAgent: string
  url: string
  referrer: string
  screenSize: string
  language: string
}

/**
 * Get browser information for error context
 */
function getBrowserInfo(): BrowserInfo {
  return {
    userAgent: navigator.userAgent,
    url: window.location.href,
    referrer: document.referrer,
    screenSize: `${window.screen.width}x${window.screen.height}`,
    language: navigator.language,
  }
}

/**
 * Generate a unique error ID
 */
function generateErrorId(): string {
  return `err_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`
}

/**
 * Categorize an error based on its characteristics
 */
function categorizeError(error: Error, context?: ErrorContext): ErrorCategory {
  const message = error.message.toLowerCase()
  const stack = error.stack?.toLowerCase() ?? ''

  // Network errors
  if (
    message.includes('network') ||
    message.includes('fetch') ||
    message.includes('xhr') ||
    message.includes('timeout') ||
    error.name === 'NetworkError' ||
    error.name === 'AbortError'
  ) {
    return 'network'
  }

  // Validation errors
  if (
    message.includes('validation') ||
    message.includes('invalid') ||
    message.includes('required')
  ) {
    return 'validation'
  }

  // Canvas errors
  if (
    context?.componentName?.toLowerCase().includes('canvas') ||
    stack.includes('reactflow') ||
    stack.includes('xyflow') ||
    message.includes('canvas')
  ) {
    return 'canvas'
  }

  // Node errors
  if (
    context?.nodeId ||
    message.includes('node') ||
    stack.includes('node')
  ) {
    return 'node'
  }

  // State errors
  if (
    message.includes('state') ||
    stack.includes('zustand') ||
    stack.includes('redux')
  ) {
    return 'state'
  }

  // Render errors (React component errors)
  if (
    stack.includes('react') ||
    context?.componentStack
  ) {
    return 'render'
  }

  return 'unknown'
}

/**
 * Determine error severity based on error characteristics
 */
function determineSeverity(error: Error, category: ErrorCategory): ErrorSeverity {
  // Fatal errors that prevent app from functioning
  if (
    category === 'render' ||
    category === 'state' ||
    error.message.includes('fatal') ||
    error.message.includes('crash')
  ) {
    return 'fatal'
  }

  // Errors that affect user experience
  if (
    category === 'canvas' ||
    category === 'network'
  ) {
    return 'error'
  }

  // Validation and node errors are usually recoverable
  if (
    category === 'validation' ||
    category === 'node'
  ) {
    return 'warning'
  }

  return 'error'
}

/**
 * Error Logger Service
 *
 * Centralized error logging with Sentry integration and local storage backup
 */
class ErrorLoggerService {
  private readonly MAX_LOCAL_LOGS = 50
  private readonly LOCAL_STORAGE_KEY = 'gorax_error_logs'

  /**
   * Log an error with full context
   */
  log(error: Error, context?: ErrorContext): ErrorLogEntry {
    const category = categorizeError(error, context)
    const severity = determineSeverity(error, category)

    const entry: ErrorLogEntry = {
      id: generateErrorId(),
      timestamp: new Date().toISOString(),
      category,
      severity,
      message: error.message,
      stack: error.stack,
      context: context ?? {},
      browserInfo: getBrowserInfo(),
    }

    // Log to console in development
    if (process.env.NODE_ENV === 'development') {
      console.group(`[ErrorLogger] ${severity.toUpperCase()}: ${category}`)
      console.error('Error:', error)
      console.log('Context:', context)
      console.log('Entry:', entry)
      console.groupEnd()
    }

    // Report to Sentry
    this.reportToSentry(error, entry)

    // Store locally for debugging
    this.storeLocally(entry)

    return entry
  }

  /**
   * Log a React error boundary catch
   */
  logBoundaryError(
    error: Error,
    errorInfo: React.ErrorInfo,
    componentName: string,
    additionalContext?: Partial<ErrorContext>
  ): ErrorLogEntry {
    return this.log(error, {
      componentName,
      componentStack: errorInfo.componentStack ?? undefined,
      ...additionalContext,
    })
  }

  /**
   * Log a network/API error
   */
  logNetworkError(
    error: Error,
    endpoint: string,
    method: string,
    additionalContext?: Partial<ErrorContext>
  ): ErrorLogEntry {
    return this.log(error, {
      action: `${method} ${endpoint}`,
      additionalData: {
        endpoint,
        method,
      },
      ...additionalContext,
    })
  }

  /**
   * Log a workflow-specific error
   */
  logWorkflowError(
    error: Error,
    workflowId: string,
    nodeId?: string,
    action?: string
  ): ErrorLogEntry {
    return this.log(error, {
      workflowId,
      nodeId,
      action,
    })
  }

  /**
   * Report error to Sentry with context
   */
  private reportToSentry(error: Error, entry: ErrorLogEntry): void {
    Sentry.withScope((scope) => {
      scope.setLevel(this.mapSeverityToSentryLevel(entry.severity))
      scope.setTag('error_category', entry.category)
      scope.setTag('error_id', entry.id)

      if (entry.context.componentName) {
        scope.setTag('component', entry.context.componentName)
      }

      if (entry.context.workflowId) {
        scope.setTag('workflow_id', entry.context.workflowId)
      }

      if (entry.context.nodeId) {
        scope.setTag('node_id', entry.context.nodeId)
      }

      scope.setContext('errorContext', { ...entry.context })
      scope.setContext('browserInfo', { ...entry.browserInfo })

      if (entry.context.componentStack) {
        scope.setContext('componentStack', {
          stack: entry.context.componentStack,
        })
      }

      Sentry.captureException(error)
    })
  }

  /**
   * Map internal severity to Sentry severity level
   */
  private mapSeverityToSentryLevel(severity: ErrorSeverity): Sentry.SeverityLevel {
    const mapping: Record<ErrorSeverity, Sentry.SeverityLevel> = {
      fatal: 'fatal',
      error: 'error',
      warning: 'warning',
      info: 'info',
    }
    return mapping[severity]
  }

  /**
   * Store error log locally for debugging
   */
  private storeLocally(entry: ErrorLogEntry): void {
    try {
      const logs = this.getLocalLogs()
      logs.unshift(entry)

      // Keep only the most recent logs
      const trimmedLogs = logs.slice(0, this.MAX_LOCAL_LOGS)

      localStorage.setItem(this.LOCAL_STORAGE_KEY, JSON.stringify(trimmedLogs))
    } catch {
      // Silently fail if localStorage is unavailable
    }
  }

  /**
   * Get locally stored error logs
   */
  getLocalLogs(): ErrorLogEntry[] {
    try {
      const stored = localStorage.getItem(this.LOCAL_STORAGE_KEY)
      return stored ? JSON.parse(stored) : []
    } catch {
      return []
    }
  }

  /**
   * Clear locally stored error logs
   */
  clearLocalLogs(): void {
    try {
      localStorage.removeItem(this.LOCAL_STORAGE_KEY)
    } catch {
      // Silently fail
    }
  }

  /**
   * Get error logs filtered by category
   */
  getLogsByCategory(category: ErrorCategory): ErrorLogEntry[] {
    return this.getLocalLogs().filter((log) => log.category === category)
  }

  /**
   * Get error logs filtered by severity
   */
  getLogsBySeverity(severity: ErrorSeverity): ErrorLogEntry[] {
    return this.getLocalLogs().filter((log) => log.severity === severity)
  }
}

// Export singleton instance
export const errorLogger = new ErrorLoggerService()

// Export class for testing
export { ErrorLoggerService }
