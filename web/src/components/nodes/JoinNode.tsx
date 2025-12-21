import { Handle, Position } from '@xyflow/react'

export interface JoinNodeData {
  label: string
  joinStrategy?: 'wait_all' | 'wait_n'
  requiredCount?: number
  timeoutMs?: number
  onTimeout?: 'fail' | 'continue'
}

export interface JoinNodeProps {
  id: string
  data: JoinNodeData
  selected?: boolean
}

export default function JoinNode({ data, selected }: JoinNodeProps) {
  return (
    <div
      className={`
        px-4 py-3 rounded-lg shadow-lg min-w-[200px]
        bg-gradient-to-br from-orange-500 to-amber-600
        ${selected ? 'ring-2 ring-white ring-offset-2 ring-offset-gray-900' : ''}
      `}
    >
      {/* Input handle */}
      <Handle
        type="target"
        position={Position.Top}
        className="w-3 h-3 bg-white border-2 border-orange-500"
      />

      <div className="space-y-2">
        <div className="flex items-center space-x-2">
          <span className="text-lg">⚡</span>
          <div className="flex-1">
            <p className="text-white font-medium text-sm">{data.label}</p>
            <p className="text-white/70 text-xs">Join</p>
          </div>
        </div>

        {/* Display configuration details */}
        {data.joinStrategy && (
          <div className="mt-2 p-2 bg-white/10 rounded text-xs text-white/90">
            {data.joinStrategy === 'wait_all'
              ? 'Wait All'
              : `Wait ${data.requiredCount || 'N'}`}
          </div>
        )}

        {data.timeoutMs !== undefined && (
          <p className="text-white/80 text-xs">
            Timeout: {data.timeoutMs}ms
          </p>
        )}

        {data.onTimeout && data.timeoutMs !== undefined && (
          <p className="text-white/80 text-xs">
            {data.onTimeout === 'continue' ? 'Continue on timeout' : 'Fail on timeout'}
          </p>
        )}
      </div>

      {/* Output handle */}
      <Handle
        type="source"
        position={Position.Bottom}
        className="w-3 h-3 bg-white border-2 border-orange-500"
      />
    </div>
  )
}
