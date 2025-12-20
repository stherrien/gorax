import { Handle, Position } from '@xyflow/react'

interface AINodeData {
  label: string
  nodeType: string
  aiConfig?: {
    model?: string
    provider?: string
  }
}

export interface AINodeProps {
  id: string
  data: AINodeData
  selected?: boolean
}

const AI_ACTION_CONFIG: Record<
  string,
  { label: string; icon: string; description: string }
> = {
  ai_chat: {
    label: 'Chat Completion',
    icon: '🤖',
    description: 'Generate AI responses',
  },
  ai_summarize: {
    label: 'Summarize',
    icon: '📝',
    description: 'Summarize text',
  },
  ai_classify: {
    label: 'Classify',
    icon: '🏷️',
    description: 'Classify into categories',
  },
  ai_extract: {
    label: 'Extract Entities',
    icon: '🔍',
    description: 'Extract named entities',
  },
  ai_embed: {
    label: 'Embeddings',
    icon: '📊',
    description: 'Generate vector embeddings',
  },
}

const DEFAULT_CONFIG = {
  label: 'AI Action',
  icon: '🧠',
  description: 'AI-powered action',
}

export default function AINode({ data, selected }: AINodeProps) {
  const config = AI_ACTION_CONFIG[data.nodeType] || DEFAULT_CONFIG
  const modelName = data.aiConfig?.model || 'No model selected'

  return (
    <div
      className={`
        px-4 py-3 rounded-lg shadow-lg min-w-[160px]
        bg-gradient-to-br from-violet-500 to-purple-600
        ${selected ? 'ring-2 ring-white ring-offset-2 ring-offset-gray-900' : ''}
      `}
    >
      {/* Input handle */}
      <Handle
        type="target"
        position={Position.Top}
        className="w-3 h-3 bg-white border-2 border-violet-500"
      />

      <div className="flex items-center space-x-2">
        <span className="text-xl">{config.icon}</span>
        <div className="flex-1 min-w-0">
          <p className="text-white font-medium text-sm truncate">{data.label}</p>
          <p className="text-white/80 text-xs">{config.label}</p>
          <p className="text-white/60 text-xs truncate">{modelName}</p>
        </div>
      </div>

      {/* Output handle */}
      <Handle
        type="source"
        position={Position.Bottom}
        className="w-3 h-3 bg-white border-2 border-violet-500"
      />
    </div>
  )
}
