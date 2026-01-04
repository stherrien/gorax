import { useState, useEffect } from 'react'
import { useParams, Link, Navigate } from 'react-router-dom'
import { useExecution } from '../hooks/useExecutions'
import { executionAPI } from '../api/executions'
import type { ExecutionStep, ExecutionStatus } from '../api/executions'
import { isValidResourceId } from '../utils/routing'

export default function ExecutionDetail() {
  const { id } = useParams()

  // Guard against invalid IDs
  if (!isValidResourceId(id)) {
    return <Navigate to="/executions" replace />
  }

  const { execution, loading, error, refetch } = useExecution(id)

  const [steps, setSteps] = useState<ExecutionStep[]>([])
  const [stepsLoading, setStepsLoading] = useState(false)
  const [actionMessage, setActionMessage] = useState<string | null>(null)
  const [actionLoading, setActionLoading] = useState(false)

  // Load steps when execution loads
  useEffect(() => {
    const loadSteps = async () => {
      if (!execution?.id) return

      try {
        setStepsLoading(true)
        const response = await executionAPI.getSteps(execution.id)
        setSteps(response.steps)
      } catch (err) {
        console.error('Failed to load steps:', err)
      } finally {
        setStepsLoading(false)
      }
    }

    loadSteps()
  }, [execution?.id])

  const handleCancel = async () => {
    if (!execution?.id) return

    try {
      setActionLoading(true)
      setActionMessage(null)
      await executionAPI.cancel(execution.id)
      setActionMessage('Execution cancelled successfully')
      refetch()
    } catch (err: any) {
      setActionMessage(`Failed to cancel: ${err.message}`)
    } finally {
      setActionLoading(false)
    }
  }

  const handleRetry = async () => {
    if (!execution?.id) return

    try {
      setActionLoading(true)
      setActionMessage(null)
      const newExecution = await executionAPI.retry(execution.id)
      setActionMessage(`Execution started: ${newExecution.id}`)
    } catch (err: any) {
      setActionMessage(`Failed to retry: ${err.message}`)
    } finally {
      setActionLoading(false)
    }
  }

  const formatDuration = (ms?: number) => {
    if (!ms) return 'N/A'
    if (ms < 1000) return `${ms}ms`
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
    if (ms < 3600000) return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`
    return `${Math.floor(ms / 3600000)}h ${Math.floor((ms % 3600000) / 60000)}m`
  }

  // Loading state
  if (loading) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-white text-lg">Loading execution...</div>
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-400 text-lg mb-4">Failed to load execution</div>
          <div className="text-gray-400 text-sm">{error.message}</div>
        </div>
      </div>
    )
  }

  if (!execution) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-gray-400 text-lg">Execution not found</div>
      </div>
    )
  }

  const showCancelButton = execution.status === 'running' || execution.status === 'queued'
  const showRetryButton = execution.status === 'failed'

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <Link to="/executions" className="text-gray-400 hover:text-white text-sm mb-2 block">
            ← Back to Executions
          </Link>
          <h1 className="text-2xl font-bold text-white">
            Execution {execution.id.substring(0, 8)}...
          </h1>
          <p className="text-gray-400">{execution.workflowName}</p>
        </div>
        <div className="flex items-center space-x-3">
          {showCancelButton && (
            <button
              onClick={handleCancel}
              disabled={actionLoading}
              className="px-4 py-2 bg-red-600 text-white rounded-lg text-sm font-medium hover:bg-red-700 transition-colors disabled:opacity-50"
            >
              {actionLoading ? 'Cancelling...' : 'Cancel'}
            </button>
          )}
          {showRetryButton && (
            <button
              onClick={handleRetry}
              disabled={actionLoading}
              className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors disabled:opacity-50"
            >
              {actionLoading ? 'Retrying...' : 'Retry'}
            </button>
          )}
          <StatusBadge status={execution.status} />
        </div>
      </div>

      {/* Action Message */}
      {actionMessage && (
        <div className="mb-4 p-3 bg-gray-800 rounded-lg text-white text-sm">
          {actionMessage}
        </div>
      )}

      {/* Summary */}
      <div className="grid grid-cols-4 gap-4 mb-8">
        <div className="bg-gray-800 rounded-lg p-4">
          <p className="text-gray-400 text-sm">Trigger</p>
          <p className="text-white font-medium capitalize">{execution.trigger.type}</p>
          {execution.trigger.source && (
            <p className="text-gray-500 text-xs">{execution.trigger.source}</p>
          )}
        </div>
        <div className="bg-gray-800 rounded-lg p-4">
          <p className="text-gray-400 text-sm">Started</p>
          <p className="text-white font-medium">
            {new Date(execution.startedAt).toLocaleString()}
          </p>
        </div>
        <div className="bg-gray-800 rounded-lg p-4">
          <p className="text-gray-400 text-sm">Completed</p>
          <p className="text-white font-medium">
            {execution.completedAt
              ? new Date(execution.completedAt).toLocaleString()
              : 'In progress'}
          </p>
        </div>
        <div className="bg-gray-800 rounded-lg p-4">
          <p className="text-gray-400 text-sm">Duration</p>
          <p className="text-white font-medium">{formatDuration(execution.duration)}</p>
        </div>
      </div>

      {/* Steps */}
      <div className="bg-gray-800 rounded-lg p-6">
        <h2 className="text-lg font-semibold text-white mb-4">Execution Steps</h2>

        {stepsLoading ? (
          <div className="text-center py-8 text-gray-400">Loading steps...</div>
        ) : steps.length === 0 ? (
          <div className="text-center py-8 text-gray-400">No steps found</div>
        ) : (
          <div className="space-y-4">
            {steps.map((step, index) => (
              <StepItem key={step.id} step={step} index={index} />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

interface StepItemProps {
  step: ExecutionStep
  index: number
}

function StepItem({ step, index }: StepItemProps) {
  const [expanded, setExpanded] = useState(false)

  const formatDuration = (ms?: number) => {
    if (!ms) return 'N/A'
    if (ms < 1000) return `${ms}ms`
    return `${(ms / 1000).toFixed(1)}s`
  }

  const getNodeType = (nodeId: string) => {
    if (nodeId.startsWith('trigger-')) return 'trigger'
    if (nodeId.startsWith('action-')) return 'action'
    if (nodeId.startsWith('control-')) return 'control'
    return 'unknown'
  }

  return (
    <div className="border border-gray-700 rounded-lg overflow-hidden">
      <div
        className="flex items-center justify-between p-4 bg-gray-700/50 cursor-pointer hover:bg-gray-700"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-center space-x-4">
          <span className="w-6 h-6 rounded-full bg-gray-600 flex items-center justify-center text-sm text-white flex-shrink-0">
            {index + 1}
          </span>
          <div>
            <p className="text-white font-medium">{step.nodeName}</p>
            <p className="text-gray-400 text-sm">{getNodeType(step.nodeId)}</p>
          </div>
        </div>
        <div className="flex items-center space-x-4">
          <span className="text-gray-400 text-sm">{formatDuration(step.duration)}</span>
          <StepStatusBadge status={step.status} />
          <span className="text-gray-400 text-sm">{expanded ? '▼' : '▶'}</span>
        </div>
      </div>

      {expanded && (
        <div className="p-4 bg-gray-900 space-y-4">
          {step.input && (
            <div>
              <p className="text-gray-400 text-xs mb-2">Input</p>
              <pre className="text-sm text-gray-300 overflow-x-auto bg-gray-800 p-3 rounded">
                {JSON.stringify(step.input, null, 2)}
              </pre>
            </div>
          )}

          {step.output && (
            <div>
              <p className="text-gray-400 text-xs mb-2">Output</p>
              <pre className="text-sm text-gray-300 overflow-x-auto bg-gray-800 p-3 rounded">
                {JSON.stringify(step.output, null, 2)}
              </pre>
            </div>
          )}

          {step.error && (
            <div>
              <p className="text-red-400 text-xs mb-2">Error</p>
              <pre className="text-sm text-red-300 overflow-x-auto bg-red-900/20 p-3 rounded border border-red-500/30">
                {step.error}
              </pre>
            </div>
          )}

          <div className="text-xs text-gray-500 space-y-1">
            <p>Started: {new Date(step.startedAt).toLocaleString()}</p>
            {step.completedAt && (
              <p>Completed: {new Date(step.completedAt).toLocaleString()}</p>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

function StatusBadge({ status }: { status: ExecutionStatus }) {
  const getStatusColor = (status: ExecutionStatus) => {
    switch (status) {
      case 'completed':
        return 'bg-green-500/20 text-green-400'
      case 'failed':
        return 'bg-red-500/20 text-red-400'
      case 'running':
        return 'bg-blue-500/20 text-blue-400'
      case 'queued':
        return 'bg-yellow-500/20 text-yellow-400'
      case 'cancelled':
        return 'bg-gray-500/20 text-gray-400'
      case 'timeout':
        return 'bg-orange-500/20 text-orange-400'
      default:
        return 'bg-gray-500/20 text-gray-400'
    }
  }

  const getStatusLabel = (status: ExecutionStatus) => {
    return status.charAt(0).toUpperCase() + status.slice(1)
  }

  return (
    <span className={`inline-flex px-3 py-1 text-sm font-medium rounded-full ${getStatusColor(status)}`}>
      {getStatusLabel(status)}
    </span>
  )
}

function StepStatusBadge({ status }: { status: ExecutionStatus }) {
  const getStatusColor = (status: ExecutionStatus) => {
    switch (status) {
      case 'completed':
        return 'bg-green-500/20 text-green-400'
      case 'failed':
        return 'bg-red-500/20 text-red-400'
      case 'running':
        return 'bg-blue-500/20 text-blue-400'
      case 'queued':
        return 'bg-yellow-500/20 text-yellow-400'
      case 'cancelled':
        return 'bg-gray-500/20 text-gray-400'
      case 'timeout':
        return 'bg-orange-500/20 text-orange-400'
      default:
        return 'bg-gray-500/20 text-gray-400'
    }
  }

  const getStatusLabel = (status: ExecutionStatus) => {
    return status.charAt(0).toUpperCase() + status.slice(1)
  }

  return (
    <span className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(status)}`}>
      {getStatusLabel(status)}
    </span>
  )
}
