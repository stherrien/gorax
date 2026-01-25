import { type ReactNode } from 'react'

/**
 * Error severity levels for different error boundary contexts
 */
export type ErrorSeverity = 'critical' | 'warning' | 'info'

/**
 * Props for error fallback components
 */
export interface ErrorFallbackProps {
  error: Error
  errorInfo?: React.ErrorInfo
  componentStack?: string
  onReset?: () => void
  onRetry?: () => void
  severity?: ErrorSeverity
  title?: string
  description?: string
  showDetails?: boolean
  children?: ReactNode
}

/**
 * Icon component for error display
 */
function ErrorIcon({ severity }: { severity: ErrorSeverity }) {
  const colors = {
    critical: 'text-red-500',
    warning: 'text-yellow-500',
    info: 'text-blue-500',
  }

  const bgColors = {
    critical: 'bg-red-100',
    warning: 'bg-yellow-100',
    info: 'bg-blue-100',
  }

  return (
    <div className={`flex items-center justify-center w-12 h-12 mx-auto ${bgColors[severity]} rounded-full`}>
      <svg
        className={`w-6 h-6 ${colors[severity]}`}
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
        />
      </svg>
    </div>
  )
}

/**
 * Generic error fallback component for full-page errors
 */
export function ErrorFallback({
  error,
  errorInfo,
  onReset,
  onRetry,
  severity = 'critical',
  title = 'Something went wrong',
  description = 'We apologize for the inconvenience. Please try again or contact support if the problem persists.',
  showDetails = process.env.NODE_ENV === 'development',
}: ErrorFallbackProps) {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div className="max-w-md w-full">
        <div className="bg-white rounded-lg shadow-lg p-6">
          <ErrorIcon severity={severity} />

          <h2 className="mt-4 text-xl font-semibold text-gray-900 text-center">
            {title}
          </h2>

          <p className="mt-2 text-sm text-gray-600 text-center">
            {description}
          </p>

          {showDetails && error && (
            <details className="mt-4 text-xs">
              <summary className="cursor-pointer text-gray-700 font-medium">
                Error details (development only)
              </summary>
              <pre className="mt-2 p-2 bg-gray-100 rounded text-red-600 overflow-auto max-h-40">
                {error.toString()}
                {errorInfo?.componentStack}
              </pre>
            </details>
          )}

          <div className="mt-6 flex flex-col gap-2">
            {onRetry && (
              <button
                onClick={onRetry}
                className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
              >
                Try again
              </button>
            )}

            {onReset && (
              <button
                onClick={onReset}
                className="w-full px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition-colors"
              >
                Reset
              </button>
            )}

            <button
              onClick={() => (window.location.href = '/')}
              className="w-full px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition-colors"
            >
              Go to homepage
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

/**
 * Inline error fallback for component-level errors (panels, nodes, etc.)
 */
export function InlineErrorFallback({
  error,
  onReset,
  onRetry,
  severity = 'warning',
  title = 'Component Error',
  description,
  showDetails = process.env.NODE_ENV === 'development',
}: ErrorFallbackProps) {
  const bgColors = {
    critical: 'bg-red-900/20 border-red-500/30',
    warning: 'bg-yellow-900/20 border-yellow-500/30',
    info: 'bg-blue-900/20 border-blue-500/30',
  }

  const textColors = {
    critical: 'text-red-400',
    warning: 'text-yellow-400',
    info: 'text-blue-400',
  }

  return (
    <div className={`p-4 rounded-lg border ${bgColors[severity]}`}>
      <div className="flex items-start space-x-3">
        <svg
          className={`w-5 h-5 ${textColors[severity]} flex-shrink-0 mt-0.5`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
        <div className="flex-1 min-w-0">
          <h4 className={`text-sm font-medium ${textColors[severity]}`}>{title}</h4>
          {description && (
            <p className="mt-1 text-sm text-gray-400">{description}</p>
          )}

          {showDetails && error && (
            <details className="mt-2 text-xs">
              <summary className="cursor-pointer text-gray-500 hover:text-gray-400">
                Show details
              </summary>
              <pre className="mt-1 p-2 bg-gray-900/50 rounded text-red-400 overflow-auto max-h-24 text-xs">
                {error.message}
              </pre>
            </details>
          )}

          <div className="mt-3 flex space-x-2">
            {onRetry && (
              <button
                onClick={onRetry}
                className="px-3 py-1 text-xs bg-gray-700 text-white rounded hover:bg-gray-600 transition-colors"
              >
                Retry
              </button>
            )}
            {onReset && (
              <button
                onClick={onReset}
                className="px-3 py-1 text-xs bg-gray-700 text-white rounded hover:bg-gray-600 transition-colors"
              >
                Reset
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

/**
 * Canvas-specific error fallback with workflow recovery options
 */
export function CanvasErrorFallback({
  error,
  onReset,
  onRetry,
  showDetails = process.env.NODE_ENV === 'development',
}: ErrorFallbackProps & {
  onRecoverFromBackup?: () => void
}) {
  return (
    <div className="w-full h-full flex items-center justify-center bg-gray-900">
      <div className="max-w-lg w-full mx-4">
        <div className="bg-gray-800 rounded-lg shadow-xl p-6 border border-gray-700">
          <div className="flex items-center justify-center w-16 h-16 mx-auto bg-red-900/30 rounded-full">
            <svg
              className="w-8 h-8 text-red-500"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
          </div>

          <h2 className="mt-4 text-xl font-semibold text-white text-center">
            Canvas Error
          </h2>

          <p className="mt-2 text-sm text-gray-400 text-center">
            The workflow canvas encountered an error. Your workflow data may still be saved.
            Try refreshing or recovering from a backup.
          </p>

          {showDetails && error && (
            <details className="mt-4 text-xs">
              <summary className="cursor-pointer text-gray-500 hover:text-gray-400 font-medium">
                Error details (development only)
              </summary>
              <pre className="mt-2 p-3 bg-gray-900 rounded text-red-400 overflow-auto max-h-32 text-xs">
                {error.message}
                {'\n\nStack:\n'}
                {error.stack}
              </pre>
            </details>
          )}

          <div className="mt-6 flex flex-col gap-2">
            {onRetry && (
              <button
                onClick={onRetry}
                className="w-full px-4 py-2 bg-primary-600 text-white rounded-lg font-medium hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 focus:ring-offset-gray-800 transition-colors"
              >
                Retry
              </button>
            )}

            {onReset && (
              <button
                onClick={onReset}
                className="w-full px-4 py-2 bg-gray-700 text-white rounded-lg font-medium hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 focus:ring-offset-gray-800 transition-colors"
              >
                Reset Canvas
              </button>
            )}

            <button
              onClick={() => window.location.reload()}
              className="w-full px-4 py-2 bg-gray-700 text-gray-300 rounded-lg font-medium hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 focus:ring-offset-gray-800 transition-colors"
            >
              Refresh Page
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

/**
 * Panel-specific error fallback (property panel, node palette, etc.)
 */
export function PanelErrorFallback({
  error,
  onReset,
  onRetry,
  title = 'Panel Error',
  showDetails = process.env.NODE_ENV === 'development',
}: ErrorFallbackProps) {
  return (
    <div className="w-full h-full bg-gray-800 border-l border-gray-700 flex flex-col">
      <div className="flex-1 flex items-center justify-center p-4">
        <div className="text-center">
          <div className="flex items-center justify-center w-12 h-12 mx-auto bg-yellow-900/30 rounded-full mb-4">
            <svg
              className="w-6 h-6 text-yellow-500"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
          </div>

          <h3 className="text-lg font-medium text-white mb-2">{title}</h3>
          <p className="text-sm text-gray-400 mb-4">
            This panel encountered an error.
          </p>

          {showDetails && error && (
            <details className="mb-4 text-xs text-left">
              <summary className="cursor-pointer text-gray-500 hover:text-gray-400">
                Show details
              </summary>
              <pre className="mt-2 p-2 bg-gray-900 rounded text-red-400 overflow-auto max-h-24">
                {error.message}
              </pre>
            </details>
          )}

          <div className="flex justify-center space-x-2">
            {onRetry && (
              <button
                onClick={onRetry}
                className="px-4 py-2 text-sm bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
              >
                Retry
              </button>
            )}
            {onReset && (
              <button
                onClick={onReset}
                className="px-4 py-2 text-sm bg-gray-700 text-white rounded-lg hover:bg-gray-600 transition-colors"
              >
                Reset
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default ErrorFallback
