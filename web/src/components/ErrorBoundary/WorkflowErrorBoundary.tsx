import { Component, type ErrorInfo, type ReactNode } from 'react'
import { errorLogger, type ErrorContext } from '../../services/errorLogger'
import { CanvasErrorFallback, PanelErrorFallback, InlineErrorFallback } from './ErrorFallback'

/**
 * Props for WorkflowErrorBoundary
 */
interface WorkflowErrorBoundaryProps {
  children: ReactNode
  workflowId?: string
  componentName?: string
  fallbackType?: 'canvas' | 'panel' | 'inline' | 'custom'
  fallback?: ReactNode
  onError?: (error: Error, errorInfo: ErrorInfo) => void
  onReset?: () => void
}

/**
 * State for WorkflowErrorBoundary
 */
interface WorkflowErrorBoundaryState {
  hasError: boolean
  error?: Error
  errorInfo?: ErrorInfo
}

/**
 * WorkflowErrorBoundary
 *
 * Specialized error boundary for workflow-related components.
 * Provides workflow-specific error logging and recovery options.
 */
class WorkflowErrorBoundary extends Component<
  WorkflowErrorBoundaryProps,
  WorkflowErrorBoundaryState
> {
  constructor(props: WorkflowErrorBoundaryProps) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): WorkflowErrorBoundaryState {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    const { workflowId, componentName, onError } = this.props

    // Log the error with workflow context
    const context: ErrorContext = {
      componentName: componentName ?? 'WorkflowComponent',
      componentStack: errorInfo.componentStack ?? undefined,
      workflowId,
    }

    errorLogger.logBoundaryError(
      error,
      errorInfo,
      componentName ?? 'WorkflowComponent',
      context
    )

    // Update state with error info
    this.setState({ error, errorInfo })

    // Call custom error handler if provided
    onError?.(error, errorInfo)
  }

  handleReset = (): void => {
    this.setState({ hasError: false, error: undefined, errorInfo: undefined })
    this.props.onReset?.()
  }

  handleRetry = (): void => {
    this.setState({ hasError: false, error: undefined, errorInfo: undefined })
  }

  render(): ReactNode {
    const { hasError, error, errorInfo } = this.state
    const { children, fallbackType = 'inline', fallback } = this.props

    if (!hasError) {
      return children
    }

    // Use custom fallback if provided
    if (fallback) {
      return fallback
    }

    const errorObj = error ?? new Error('Unknown error')

    // Render appropriate fallback based on type
    switch (fallbackType) {
      case 'canvas':
        return (
          <CanvasErrorFallback
            error={errorObj}
            errorInfo={errorInfo}
            onReset={this.handleReset}
            onRetry={this.handleRetry}
          />
        )

      case 'panel':
        return (
          <PanelErrorFallback
            error={errorObj}
            errorInfo={errorInfo}
            onReset={this.handleReset}
            onRetry={this.handleRetry}
          />
        )

      case 'inline':
      default:
        return (
          <InlineErrorFallback
            error={errorObj}
            errorInfo={errorInfo}
            onReset={this.handleReset}
            onRetry={this.handleRetry}
            title="Component Error"
            description="This component encountered an error."
          />
        )
    }
  }
}

/**
 * CanvasErrorBoundary
 *
 * Specialized error boundary for the workflow canvas.
 * Pre-configured with canvas-specific fallback and recovery options.
 */
class CanvasErrorBoundary extends Component<
  Omit<WorkflowErrorBoundaryProps, 'fallbackType'>,
  WorkflowErrorBoundaryState
> {
  constructor(props: Omit<WorkflowErrorBoundaryProps, 'fallbackType'>) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): WorkflowErrorBoundaryState {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    const { workflowId, onError } = this.props

    errorLogger.logBoundaryError(
      error,
      errorInfo,
      'WorkflowCanvas',
      {
        workflowId,
        componentStack: errorInfo.componentStack ?? undefined,
      }
    )

    this.setState({ error, errorInfo })
    onError?.(error, errorInfo)
  }

  handleReset = (): void => {
    this.setState({ hasError: false, error: undefined, errorInfo: undefined })
    this.props.onReset?.()
  }

  handleRetry = (): void => {
    this.setState({ hasError: false, error: undefined, errorInfo: undefined })
  }

  render(): ReactNode {
    const { hasError, error, errorInfo } = this.state
    const { children, fallback } = this.props

    if (!hasError) {
      return children
    }

    if (fallback) {
      return fallback
    }

    return (
      <CanvasErrorFallback
        error={error ?? new Error('Unknown canvas error')}
        errorInfo={errorInfo}
        onReset={this.handleReset}
        onRetry={this.handleRetry}
      />
    )
  }
}

/**
 * PanelErrorBoundary
 *
 * Specialized error boundary for panels (property panel, node palette, etc.).
 * Pre-configured with panel-specific fallback.
 */
class PanelErrorBoundary extends Component<
  Omit<WorkflowErrorBoundaryProps, 'fallbackType'> & { title?: string },
  WorkflowErrorBoundaryState
> {
  constructor(props: Omit<WorkflowErrorBoundaryProps, 'fallbackType'> & { title?: string }) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): WorkflowErrorBoundaryState {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    const { componentName, onError } = this.props

    errorLogger.logBoundaryError(
      error,
      errorInfo,
      componentName ?? 'Panel',
      {
        componentStack: errorInfo.componentStack ?? undefined,
      }
    )

    this.setState({ error, errorInfo })
    onError?.(error, errorInfo)
  }

  handleReset = (): void => {
    this.setState({ hasError: false, error: undefined, errorInfo: undefined })
    this.props.onReset?.()
  }

  handleRetry = (): void => {
    this.setState({ hasError: false, error: undefined, errorInfo: undefined })
  }

  render(): ReactNode {
    const { hasError, error, errorInfo } = this.state
    const { children, fallback, title } = this.props

    if (!hasError) {
      return children
    }

    if (fallback) {
      return fallback
    }

    return (
      <PanelErrorFallback
        error={error ?? new Error('Unknown panel error')}
        errorInfo={errorInfo}
        onReset={this.handleReset}
        onRetry={this.handleRetry}
        title={title}
      />
    )
  }
}

/**
 * NodeErrorBoundary
 *
 * Specialized error boundary for individual nodes.
 * Shows a minimal inline error to avoid disrupting the canvas layout.
 */
class NodeErrorBoundary extends Component<
  Omit<WorkflowErrorBoundaryProps, 'fallbackType'> & { nodeId?: string },
  WorkflowErrorBoundaryState
> {
  constructor(props: Omit<WorkflowErrorBoundaryProps, 'fallbackType'> & { nodeId?: string }) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): WorkflowErrorBoundaryState {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    const { nodeId, workflowId, onError } = this.props

    errorLogger.logBoundaryError(
      error,
      errorInfo,
      'Node',
      {
        nodeId,
        workflowId,
        componentStack: errorInfo.componentStack ?? undefined,
      }
    )

    this.setState({ error, errorInfo })
    onError?.(error, errorInfo)
  }

  handleRetry = (): void => {
    this.setState({ hasError: false, error: undefined, errorInfo: undefined })
  }

  render(): ReactNode {
    const { hasError, error } = this.state
    const { children, fallback } = this.props

    if (!hasError) {
      return children
    }

    if (fallback) {
      return fallback
    }

    // Minimal node error display to maintain canvas layout
    return (
      <div className="p-2 bg-red-900/30 border border-red-500/50 rounded text-center min-w-[100px]">
        <div className="text-red-400 text-xs font-medium mb-1">Node Error</div>
        <button
          onClick={this.handleRetry}
          className="text-xs text-gray-400 hover:text-white underline"
        >
          Retry
        </button>
        {process.env.NODE_ENV === 'development' && error && (
          <div className="mt-1 text-xs text-red-300 truncate" title={error.message}>
            {error.message.slice(0, 50)}
          </div>
        )}
      </div>
    )
  }
}

export {
  WorkflowErrorBoundary,
  CanvasErrorBoundary,
  PanelErrorBoundary,
  NodeErrorBoundary,
}

export default WorkflowErrorBoundary
