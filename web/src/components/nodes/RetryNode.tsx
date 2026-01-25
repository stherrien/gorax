import { Handle, Position } from '@xyflow/react'
import { useState } from 'react'

export interface RetryNodeData {
  label: string
  config: {
    strategy: 'fixed' | 'exponential' | 'exponential_jitter'
    maxAttempts: number
    initialDelayMs: number
    maxDelayMs?: number
    multiplier?: number
    jitter?: boolean
    retryableErrors?: string[]
    nonRetryableErrors?: string[]
    retryableStatusCodes?: number[]
  }
  onConfigChange?: (config: RetryNodeData['config']) => void
}

export interface RetryNodeProps {
  id: string
  data: RetryNodeData
  selected?: boolean
}

export default function RetryNode({ id, data, selected }: RetryNodeProps) {
  const [isExpanded, setIsExpanded] = useState(false)

  const getStrategyIcon = () => {
    switch (data.config.strategy) {
      case 'fixed':
        return '⏱️'
      case 'exponential':
        return '📈'
      case 'exponential_jitter':
        return '🎲'
      default:
        return '🔄'
    }
  }

  const getStrategyLabel = () => {
    switch (data.config.strategy) {
      case 'fixed':
        return 'Fixed Delay'
      case 'exponential':
        return 'Exponential'
      case 'exponential_jitter':
        return 'Exponential + Jitter'
      default:
        return 'Unknown'
    }
  }

  const formatDelay = (ms: number) => {
    if (ms < 1000) return `${ms}ms`
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
    return `${(ms / 60000).toFixed(1)}m`
  }

  return (
    <div
      className={`
        px-4 py-3 rounded-lg shadow-lg min-w-[180px]
        bg-gradient-to-br from-blue-500 to-indigo-600
        ${selected ? 'ring-2 ring-white ring-offset-2 ring-offset-gray-900' : ''}
      `}
    >
      {/* Input handle */}
      <Handle
        type="target"
        position={Position.Top}
        className="w-3 h-3 bg-white border-2 border-blue-500"
      />

      {/* Header */}
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center space-x-2">
          <span className="text-lg">{getStrategyIcon()}</span>
          <div>
            <p className="text-white font-medium text-sm">{data.label}</p>
            <p className="text-white/70 text-xs">Retry: {getStrategyLabel()}</p>
          </div>
        </div>
        <button
          onClick={() => setIsExpanded(!isExpanded)}
          className="text-white hover:text-white/70 transition-colors"
        >
          {isExpanded ? '▼' : '▶'}
        </button>
      </div>

      {/* Quick info */}
      <div className="flex items-center justify-between text-xs text-white/80 mb-1">
        <span>Max Attempts: {data.config.maxAttempts}</span>
        <span>Delay: {formatDelay(data.config.initialDelayMs)}</span>
      </div>

      {/* Expanded config */}
      {isExpanded && (
        <div className="mt-2 space-y-2 text-xs text-white/80">
          {data.config.maxDelayMs && (
            <div>
              <strong>Max Delay:</strong> {formatDelay(data.config.maxDelayMs)}
            </div>
          )}
          {data.config.multiplier && (
            <div>
              <strong>Multiplier:</strong> {data.config.multiplier}x
            </div>
          )}
          {data.config.jitter && (
            <div>
              <strong>Jitter:</strong> Enabled
            </div>
          )}
          {data.config.retryableErrors && data.config.retryableErrors.length > 0 && (
            <div>
              <strong>Retry On:</strong> {data.config.retryableErrors.join(', ')}
            </div>
          )}
          {data.config.nonRetryableErrors && data.config.nonRetryableErrors.length > 0 && (
            <div>
              <strong>Don't Retry:</strong> {data.config.nonRetryableErrors.join(', ')}
            </div>
          )}
          {data.config.retryableStatusCodes && data.config.retryableStatusCodes.length > 0 && (
            <div>
              <strong>Status Codes:</strong> {data.config.retryableStatusCodes.join(', ')}
            </div>
          )}
        </div>
      )}

      {/* Output handles */}
      <div className="mt-3 flex justify-around">
        <div className="flex flex-col items-center">
          <span className="text-white/70 text-xs mb-1">Success</span>
          <Handle
            type="source"
            position={Position.Bottom}
            id={`${id}-success`}
            className="w-3 h-3 bg-white border-2 border-green-500 relative"
            style={{ position: 'relative', transform: 'none', top: 0, left: 0 }}
          />
        </div>
        <div className="flex flex-col items-center">
          <span className="text-white/70 text-xs mb-1">Failed</span>
          <Handle
            type="source"
            position={Position.Bottom}
            id={`${id}-failed`}
            className="w-3 h-3 bg-white border-2 border-red-500 relative"
            style={{ position: 'relative', transform: 'none', top: 0, left: 0 }}
          />
        </div>
      </div>
    </div>
  )
}
