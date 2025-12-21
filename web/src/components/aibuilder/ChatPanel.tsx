import { FC, useState, useRef, useEffect, FormEvent } from 'react'
import { useAIBuilderChat } from '../../hooks/useAIBuilder'
import { ChatMessage } from './ChatMessage'
import type { GeneratedWorkflow } from '../../types/aibuilder'

interface ChatPanelProps {
  onWorkflowGenerated?: (workflow: GeneratedWorkflow) => void
  onApply?: (workflowId: string) => void
}

export const ChatPanel: FC<ChatPanelProps> = ({
  onWorkflowGenerated,
  onApply,
}) => {
  const [input, setInput] = useState('')
  const [error, setError] = useState<string | null>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLTextAreaElement>(null)

  const {
    messages,
    currentWorkflow,
    isGenerating,
    isLoading,
    startConversation,
    sendMessage,
    applyWorkflow,
    abandonConversation,
    reset,
    hasConversation,
    hasWorkflow,
  } = useAIBuilderChat()

  // Scroll to bottom when messages change
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  // Focus input on mount
  useEffect(() => {
    inputRef.current?.focus()
  }, [])

  // Notify parent when workflow is generated
  useEffect(() => {
    if (currentWorkflow && onWorkflowGenerated) {
      onWorkflowGenerated(currentWorkflow)
    }
  }, [currentWorkflow, onWorkflowGenerated])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!input.trim() || isLoading) return

    const message = input.trim()
    setInput('')
    setError(null)

    try {
      if (hasConversation) {
        await sendMessage(message)
      } else {
        await startConversation(message)
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to send message'
      setError(errorMessage)
      console.error('Failed to send message:', err)
    }
  }

  const handleApply = async () => {
    if (!hasWorkflow) return
    setError(null)

    try {
      const workflowId = await applyWorkflow()
      onApply?.(workflowId)
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to create workflow'
      setError(errorMessage)
      console.error('Failed to apply workflow:', err)
    }
  }

  const handleAbandon = async () => {
    setError(null)
    try {
      await abandonConversation()
    } catch (err) {
      console.error('Failed to abandon conversation:', err)
    }
  }

  const handleReset = () => {
    reset()
    setInput('')
    setError(null)
    inputRef.current?.focus()
  }

  const handleExampleClick = (text: string) => {
    setInput(text)
    setError(null)
    inputRef.current?.focus()
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit(e)
    }
  }

  const handlePreviewWorkflow = (workflow: GeneratedWorkflow) => {
    onWorkflowGenerated?.(workflow)
  }

  return (
    <div className="flex h-full flex-col bg-white">
      {/* Header */}
      <div className="border-b border-gray-200 px-4 py-3">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">
              AI Workflow Builder
            </h2>
            <p className="text-sm text-gray-500">
              Describe your workflow in natural language
            </p>
          </div>
          {hasConversation && (
            <div className="flex items-center space-x-2">
              <button
                onClick={handleReset}
                className="rounded px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100"
              >
                New Chat
              </button>
              <button
                onClick={handleAbandon}
                className="rounded px-3 py-1.5 text-sm text-red-600 hover:bg-red-50"
              >
                Abandon
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4">
        {messages.length === 0 ? (
          <div className="flex h-full flex-col items-center justify-center text-center">
            <div className="mb-4 text-4xl">🤖</div>
            <h3 className="mb-2 text-lg font-medium text-gray-900">
              Build Workflows with AI
            </h3>
            <p className="max-w-md text-sm text-gray-500">
              Describe what you want your workflow to do, and I&apos;ll generate
              it for you. You can refine it through conversation.
            </p>
            <div className="mt-6 space-y-2 text-left">
              <p className="text-xs font-medium text-gray-400">Try saying:</p>
              <ExamplePrompt text="Send a Slack message when a GitHub PR is opened" onClick={handleExampleClick} />
              <ExamplePrompt text="Every day at 9am, fetch data from an API and email me a summary" onClick={handleExampleClick} />
              <ExamplePrompt text="When a webhook is received, transform the data and POST to another API" onClick={handleExampleClick} />
            </div>
          </div>
        ) : (
          <div className="space-y-4">
            {messages.map((message) => (
              <ChatMessage
                key={message.id}
                message={message}
                onPreviewWorkflow={handlePreviewWorkflow}
              />
            ))}
            {isLoading && (
              <div className="flex justify-start">
                <div className="rounded-lg bg-gray-100 p-4">
                  <div className="flex items-center space-x-2">
                    <LoadingDots />
                    <span className="text-sm text-gray-500">
                      {isGenerating ? 'Generating workflow...' : 'Refining workflow...'}
                    </span>
                  </div>
                </div>
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>
        )}
      </div>

      {/* Workflow Actions */}
      {hasWorkflow && !isLoading && (
        <div className="border-t border-gray-200 bg-gray-50 px-4 py-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <span className="text-sm font-medium text-gray-700">
                {currentWorkflow?.name}
              </span>
              <span className="rounded bg-green-100 px-2 py-0.5 text-xs text-green-700">
                {currentWorkflow?.definition.nodes.length} nodes
              </span>
            </div>
            <button
              onClick={handleApply}
              className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
            >
              Create Workflow
            </button>
          </div>
        </div>
      )}

      {/* Error Display */}
      {error && (
        <div className="border-t border-red-200 bg-red-50 px-4 py-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <span className="text-red-500">⚠️</span>
              <span className="text-sm text-red-700">{error}</span>
            </div>
            <button
              onClick={() => setError(null)}
              className="text-red-400 hover:text-red-600"
            >
              ×
            </button>
          </div>
        </div>
      )}

      {/* Input */}
      <div className="border-t border-gray-200 p-4">
        <form onSubmit={handleSubmit}>
          <div className="flex items-end space-x-3">
            <div className="flex-1">
              <textarea
                ref={inputRef}
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder={
                  hasConversation
                    ? 'Describe changes or refinements...'
                    : 'Describe your workflow...'
                }
                rows={2}
                className="w-full resize-none rounded-lg border border-gray-300 px-4 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                disabled={isLoading}
              />
            </div>
            <button
              type="submit"
              disabled={!input.trim() || isLoading}
              className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {isLoading ? 'Sending...' : 'Send'}
            </button>
          </div>
          <p className="mt-2 text-xs text-gray-400">
            Press Enter to send, Shift+Enter for new line
          </p>
        </form>
      </div>
    </div>
  )
}

// Helper components

interface ExamplePromptProps {
  text: string
  onClick: (text: string) => void
}

const ExamplePrompt: FC<ExamplePromptProps> = ({ text, onClick }) => (
  <button
    className="block w-full rounded border border-gray-200 bg-white px-3 py-2 text-left text-sm text-gray-600 hover:border-blue-300 hover:bg-blue-50 transition-colors"
    onClick={() => onClick(text)}
  >
    &quot;{text}&quot;
  </button>
)

const LoadingDots: FC = () => (
  <div className="flex space-x-1">
    <div className="h-2 w-2 animate-bounce rounded-full bg-gray-400" style={{ animationDelay: '0ms' }} />
    <div className="h-2 w-2 animate-bounce rounded-full bg-gray-400" style={{ animationDelay: '150ms' }} />
    <div className="h-2 w-2 animate-bounce rounded-full bg-gray-400" style={{ animationDelay: '300ms' }} />
  </div>
)

export default ChatPanel
