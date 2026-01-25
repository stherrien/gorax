import { Handle, Position } from '@xyflow/react'
import { useState } from 'react'

export interface TryNodeData {
  label: string
  config: {
    tryNodes: string[]
    catchNodes?: string[]
    finallyNodes?: string[]
    errorBinding?: string
    retryConfig?: {
      strategy: string
      maxAttempts: number
      initialDelayMs: number
      maxDelayMs?: number
      multiplier?: number
      jitter?: boolean
      retryableErrors?: string[]
      nonRetryableErrors?: string[]
    }
  }
  onConfigChange?: (config: TryNodeData['config']) => void
}

export interface TryNodeProps {
  id: string
  data: TryNodeData
  selected?: boolean
}

export default function TryNode({ id, data, selected }: TryNodeProps) {
  const [isExpanded, setIsExpanded] = useState(false)

  const hasCatch = data.config.catchNodes && data.config.catchNodes.length > 0
  const hasFinally = data.config.finallyNodes && data.config.finallyNodes.length > 0

  return (
    <div
      className={`
        px-4 py-3 rounded-lg shadow-lg min-w-[200px]
        bg-gradient-to-br from-amber-500 to-orange-600
        ${selected ? 'ring-2 ring-white ring-offset-2 ring-offset-gray-900' : ''}
      `}
    >
      {/* Input handle */}
      <Handle
        type="target"
        position={Position.Top}
        className="w-3 h-3 bg-white border-2 border-amber-500"
      />

      {/* Header */}
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center space-x-2">
          <span className="text-lg">🛡️</span>
          <div>
            <p className="text-white font-medium text-sm">{data.label}</p>
            <p className="text-white/70 text-xs">Try/Catch/Finally</p>
          </div>
        </div>
        <button
          onClick={() => setIsExpanded(!isExpanded)}
          className="text-white hover:text-white/70 transition-colors"
        >
          {isExpanded ? '▼' : '▶'}
        </button>
      </div>

      {/* Expanded config */}
      {isExpanded && (
        <div className="mt-2 space-y-2 text-xs text-white/80">
          <div>
            <strong>Try Nodes:</strong> {data.config.tryNodes.length || 0}
          </div>
          {hasCatch && (
            <div>
              <strong>Catch Nodes:</strong> {data.config.catchNodes?.length || 0}
            </div>
          )}
          {hasFinally && (
            <div>
              <strong>Finally Nodes:</strong> {data.config.finallyNodes?.length || 0}
            </div>
          )}
          {data.config.errorBinding && (
            <div>
              <strong>Error Var:</strong> {data.config.errorBinding}
            </div>
          )}
          {data.config.retryConfig && (
            <div>
              <strong>Retry:</strong> {data.config.retryConfig.strategy} (max: {data.config.retryConfig.maxAttempts})
            </div>
          )}
        </div>
      )}

      {/* Branch indicators */}
      <div className="mt-2 flex justify-around text-xs">
        <div className="flex flex-col items-center">
          <span className="text-white/70">Try</span>
          <Handle
            type="source"
            position={Position.Bottom}
            id={`${id}-try`}
            className="w-3 h-3 bg-white border-2 border-green-500 relative"
            style={{ position: 'relative', transform: 'none', top: '4px', left: 0 }}
          />
        </div>
        {hasCatch && (
          <div className="flex flex-col items-center">
            <span className="text-white/70">Catch</span>
            <Handle
              type="source"
              position={Position.Bottom}
              id={`${id}-catch`}
              className="w-3 h-3 bg-white border-2 border-red-500 relative"
              style={{ position: 'relative', transform: 'none', top: '4px', left: 0 }}
            />
          </div>
        )}
        {hasFinally && (
          <div className="flex flex-col items-center">
            <span className="text-white/70">Finally</span>
            <Handle
              type="source"
              position={Position.Bottom}
              id={`${id}-finally`}
              className="w-3 h-3 bg-white border-2 border-blue-500 relative"
              style={{ position: 'relative', transform: 'none', top: '4px', left: 0 }}
            />
          </div>
        )}
      </div>

      {/* Main output handle */}
      <Handle
        type="source"
        position={Position.Bottom}
        id={`${id}-output`}
        className="w-3 h-3 bg-white border-2 border-amber-500"
      />
    </div>
  )
}
