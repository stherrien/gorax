import { Handle, Position } from '@xyflow/react'
import type { ActionNodeData } from '../../stores/workflowStore'

interface ActionNodeProps {
  data: ActionNodeData
  selected?: boolean
}

export default function ActionNode({ data, selected }: ActionNodeProps) {
  const icons: Record<string, string> = {
    http: '🌐',
    transform: '🔄',
    formula: '🔢',
    code: '💻',
    script: '📜',
    email: '📧',
    slack_send_message: '💬',
    slack_send_dm: '✉️',
    slack_update_message: '✏️',
    slack_add_reaction: '👍',
  }

  return (
    <div
      className={`
        px-4 py-3 rounded-lg shadow-lg min-w-[150px]
        bg-gradient-to-br from-emerald-500 to-teal-600
        ${selected ? 'ring-2 ring-white ring-offset-2 ring-offset-gray-900' : ''}
      `}
    >
      {/* Input handle */}
      <Handle
        type="target"
        position={Position.Top}
        className="w-3 h-3 bg-white border-2 border-emerald-500"
      />

      <div className="flex items-center space-x-2">
        <span className="text-lg">{icons[data.actionType] || '⚙️'}</span>
        <div>
          <p className="text-white font-medium text-sm">{data.label}</p>
          <p className="text-white/70 text-xs capitalize">{data.actionType}</p>
        </div>
      </div>

      {/* Output handle */}
      <Handle
        type="source"
        position={Position.Bottom}
        className="w-3 h-3 bg-white border-2 border-emerald-500"
      />
    </div>
  )
}
