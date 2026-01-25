/**
 * Error Boundary Components
 *
 * Provides comprehensive error boundary handling for the workflow builder UI.
 * Includes specialized boundaries for different component contexts and
 * fallback UIs tailored to each context.
 */

// Main error boundary (re-export from the ErrorBoundary.tsx file)
export { default as ErrorBoundary, SentryErrorBoundary } from '../ErrorBoundary.tsx'

// Specialized error boundaries
export {
  WorkflowErrorBoundary,
  CanvasErrorBoundary,
  PanelErrorBoundary,
  NodeErrorBoundary,
} from './WorkflowErrorBoundary'

// Error fallback components
export {
  ErrorFallback,
  InlineErrorFallback,
  CanvasErrorFallback,
  PanelErrorFallback,
  type ErrorFallbackProps,
  type ErrorSeverity,
} from './ErrorFallback'

// Re-export types
export type { ErrorContext, ErrorLogEntry } from '../../services/errorLogger'
