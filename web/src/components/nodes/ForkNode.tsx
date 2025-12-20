import { Handle, Position } from '@xyflow/react'

interface ForkNodeData {
  label: string
  branchCount?: number
}

interface ForkNodeProps {
  data: ForkNodeData
  selected?: boolean
}

export default function ForkNode({ data, selected }: ForkNodeProps) {
  return (
    <div
      className={`
        px-4 py-3 rounded-lg shadow-lg min-w-[200px]
        bg-gradient-to-br from-green-500 to-emerald-600
        ${selected ? 'ring-2 ring-white ring-offset-2 ring-offset-gray-900' : ''}
      `}
    >
      {/* Input handle */}
      <Handle
        type="target"
        position={Position.Top}
        className="w-3 h-3 bg-white border-2 border-green-500"
      />

      <div className="space-y-2">
        <div className="flex items-center space-x-2">
          <span className="text-lg">🔱</span>
          <div className="flex-1">
            <p className="text-white font-medium text-sm">{data.label}</p>
            <p className="text-white/70 text-xs">Fork</p>
          </div>
        </div>

        {/* Display configuration details */}
        {data.branchCount !== undefined && (
          <div className="mt-2 p-2 bg-white/10 rounded text-xs text-white/90">
            {data.branchCount} branches
          </div>
        )}
      </div>

      {/* Output handle */}
      <Handle
        type="source"
        position={Position.Bottom}
        className="w-3 h-3 bg-white border-2 border-green-500"
      />
    </div>
  )
}
