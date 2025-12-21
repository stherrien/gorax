import { FC } from 'react'
import type { ConversationMessage, GeneratedWorkflow } from '../../types/aibuilder'
import { getRoleLabel } from '../../types/aibuilder'

interface ChatMessageProps {
  message: ConversationMessage
  onPreviewWorkflow?: (workflow: GeneratedWorkflow) => void
}

export const ChatMessage: FC<ChatMessageProps> = ({
  message,
  onPreviewWorkflow,
}) => {
  const isUser = message.role === 'user'
  const isAssistant = message.role === 'assistant'

  return (
    <div
      className={`flex ${isUser ? 'justify-end' : 'justify-start'} mb-4`}
    >
      <div
        className={`max-w-[80%] rounded-lg p-4 ${
          isUser
            ? 'bg-blue-600 text-white'
            : 'bg-gray-100 text-gray-900'
        }`}
      >
        {/* Role indicator for assistant */}
        {isAssistant && (
          <div className="mb-2 flex items-center text-xs text-gray-500">
            <span className="mr-1">🤖</span>
            <span>{getRoleLabel(message.role)}</span>
          </div>
        )}

        {/* Message content */}
        <p className="whitespace-pre-wrap text-sm">{message.content}</p>

        {/* Workflow preview button */}
        {message.workflow && onPreviewWorkflow && (
          <button
            onClick={() => onPreviewWorkflow(message.workflow!)}
            className={`mt-3 flex items-center rounded px-3 py-1.5 text-xs font-medium ${
              isUser
                ? 'bg-blue-500 text-white hover:bg-blue-400'
                : 'bg-blue-100 text-blue-700 hover:bg-blue-200'
            }`}
          >
            <span className="mr-1">📋</span>
            View Workflow ({message.workflow.definition.nodes.length} nodes)
          </button>
        )}

        {/* Timestamp */}
        <div
          className={`mt-2 text-xs ${
            isUser ? 'text-blue-200' : 'text-gray-400'
          }`}
        >
          {new Date(message.created_at).toLocaleTimeString()}
        </div>
      </div>
    </div>
  )
}

export default ChatMessage
