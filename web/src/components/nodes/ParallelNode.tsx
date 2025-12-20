import { Handle, Position } from '@xyflow/react'

export interface ParallelNodeData {
  label: string
  errorStrategy?: 'fail_fast' | 'wait_all'
  maxConcurrency?: number
  branchCount?: number
}

export interface ParallelNodeProps {
  id: string
  data: ParallelNodeData
  selected?: boolean
}

export default function ParallelNode({ data, selected }: ParallelNodeProps) {
  return (
    <div
      className={`
        px-4 py-3 rounded-lg shadow-lg min-w-[200px]
        bg-gradient-to-br from-blue-500 to-cyan-600
        ${selected ? 'ring-2 ring-white ring-offset-2 ring-offset-gray-900' : ''}
      `}
    >
      {/* Input handle */}
      <Handle
        type="target"
        position={Position.Top}
        className="w-3 h-3 bg-white border-2 border-blue-500"
      />

      <div className="space-y-2">
        <div className="flex items-center space-x-2">
          <span className="text-lg">⚡</span>
          <div className="flex-1">
            <p className="text-white font-medium text-sm">{data.label}</p>
            <p className="text-white/70 text-xs">Parallel</p>
          </div>
        </div>

        {/* Display configuration details */}
        {data.branchCount !== undefined && (
          <div className="mt-2 p-2 bg-white/10 rounded text-xs text-white/90">
            {data.branchCount} branches
          </div>
        )}

        {data.errorStrategy && (
          <p className="text-white/80 text-xs">
            {data.errorStrategy === 'fail_fast' ? 'Fail Fast' : 'Wait All'}
          </p>
        )}

        {data.maxConcurrency !== undefined && (
          <p className="text-white/80 text-xs">
            {data.maxConcurrency === 0 ? 'Unlimited' : `Max: ${data.maxConcurrency}`}
          </p>
        )}
      </div>

      {/* Output handle */}
      <Handle
        type="source"
        position={Position.Bottom}
        className="w-3 h-3 bg-white border-2 border-blue-500"
      />
    </div>
  )
}
