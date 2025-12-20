import { Handle, Position } from '@xyflow/react'
import type { TriggerNodeData } from '../../stores/workflowStore'

export interface TriggerNodeProps {
  id: string
  data: TriggerNodeData
  selected?: boolean
}

export default function TriggerNode({ data, selected }: TriggerNodeProps) {
  const icons: Record<string, string> = {
    webhook: '🔗',
    schedule: '⏰',
  }

  return (
    <div
      className={`
        px-4 py-3 rounded-lg shadow-lg min-w-[150px]
        bg-gradient-to-br from-indigo-500 to-purple-600
        ${selected ? 'ring-2 ring-white ring-offset-2 ring-offset-gray-900' : ''}
      `}
    >
      <div className="flex items-center space-x-2">
        <span className="text-lg">{icons[data.triggerType] || '📥'}</span>
        <div>
          <p className="text-white font-medium text-sm">{data.label}</p>
          <p className="text-white/70 text-xs capitalize">{data.triggerType}</p>
        </div>
      </div>

      {/* Output handle */}
      <Handle
        type="source"
        position={Position.Bottom}
        className="w-3 h-3 bg-white border-2 border-indigo-500"
      />
    </div>
  )
}
