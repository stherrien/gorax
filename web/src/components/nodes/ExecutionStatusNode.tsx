import { ComponentType } from 'react'
import type { NodeProps } from '@xyflow/react'
import { useExecutionTraceStore } from '../../stores/executionTraceStore'
import type { NodeStatus } from '../../stores/executionTraceStore'
import '../../styles/executionAnimations.css'

/**
 * Status indicator icons for each execution state
 */
function StatusIndicator({ status }: { status: NodeStatus }) {
  switch (status) {
    case 'pending':
      return (
        <div
          className="absolute -top-2 -right-2 w-6 h-6 bg-gray-500 rounded-full flex items-center justify-center text-white text-xs z-10"
          data-testid="status-indicator-pending"
          title="Pending"
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        </div>
      )

    case 'running':
      return (
        <div
          className="absolute -top-2 -right-2 w-6 h-6 bg-blue-500 rounded-full flex items-center justify-center text-white text-xs z-10 animate-spin-slow"
          data-testid="status-indicator-running"
          title="Running"
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24">
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
        </div>
      )

    case 'completed':
      return (
        <div
          className="absolute -top-2 -right-2 w-6 h-6 bg-green-500 rounded-full flex items-center justify-center text-white text-xs z-10"
          data-testid="status-indicator-completed"
          title="Completed"
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
          </svg>
        </div>
      )

    case 'failed':
      return (
        <div
          className="absolute -top-2 -right-2 w-6 h-6 bg-red-500 rounded-full flex items-center justify-center text-white text-xs z-10"
          data-testid="status-indicator-failed"
          title="Failed"
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </div>
      )
  }
}

/**
 * Get animation class based on node status
 */
function getAnimationClass(status: NodeStatus | undefined): string {
  switch (status) {
    case 'running':
      return 'animate-pulse-glow'
    case 'completed':
      return 'animate-checkmark'
    case 'failed':
      return 'animate-shake'
    default:
      return ''
  }
}

/**
 * Get status class based on node status
 */
function getStatusClass(status: NodeStatus | undefined): string {
  if (!status) return ''
  return `execution-status-${status}`
}

/**
 * Higher-Order Component that wraps any node component with execution status visualization
 *
 * @example
 * ```tsx
 * const ActionNodeWithStatus = ExecutionStatusNode(ActionNode)
 * const TriggerNodeWithStatus = ExecutionStatusNode(TriggerNode)
 * ```
 */
export function ExecutionStatusNode<T extends Record<string, unknown>>(
  BaseNode: ComponentType<NodeProps<T>>
) {
  return function WrappedNode(props: NodeProps<T>) {
    // Get node status from store
    const { nodeStatuses } = useExecutionTraceStore()
    const status = nodeStatuses[props.id]

    // If no status, render base node without wrapper
    if (!status) {
      return <BaseNode {...props} />
    }

    const animationClass = getAnimationClass(status)
    const statusClass = getStatusClass(status)

    return (
      <div
        className={`relative ${statusClass} ${animationClass}`}
        data-testid="execution-status-wrapper"
      >
        <StatusIndicator status={status} />
        <BaseNode {...props} />
      </div>
    )
  }
}
