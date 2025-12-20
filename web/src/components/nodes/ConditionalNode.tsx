import { Handle, Position } from '@xyflow/react'

export interface ConditionalNodeData {
  label: string
  condition?: string
  description?: string
}

export interface ConditionalNodeProps {
  id: string
  data: ConditionalNodeData
  selected?: boolean
}

export default function ConditionalNode({ data, selected }: ConditionalNodeProps) {
  return (
    <div
      className={`
        px-4 py-3 rounded-lg shadow-lg min-w-[180px]
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

      <div className="space-y-1">
        <div className="flex items-center space-x-2">
          <span className="text-lg">🔀</span>
          <div className="flex-1">
            <p className="text-white font-medium text-sm">{data.label}</p>
            <p className="text-white/70 text-xs">Conditional</p>
          </div>
        </div>

        {data.condition && (
          <div className="mt-2 p-2 bg-white/10 rounded text-xs text-white/90 font-mono break-all">
            {data.condition}
          </div>
        )}

        {data.description && (
          <p className="text-white/60 text-xs mt-1">{data.description}</p>
        )}
      </div>

      {/* Output handles - True (right) and False (left) */}
      <Handle
        type="source"
        position={Position.Right}
        id="true"
        className="w-3 h-3 bg-green-400 border-2 border-white"
        style={{ top: '50%' }}
      />
      <Handle
        type="source"
        position={Position.Left}
        id="false"
        className="w-3 h-3 bg-red-400 border-2 border-white"
        style={{ top: '50%' }}
      />

      {/* Labels for true/false outputs */}
      <div className="absolute -right-12 top-1/2 transform -translate-y-1/2 text-xs text-green-400 font-medium">
        True
      </div>
      <div className="absolute -left-12 top-1/2 transform -translate-y-1/2 text-xs text-red-400 font-medium">
        False
      </div>
    </div>
  )
}
