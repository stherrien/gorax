import { useState, useEffect } from 'react'
import type { Node } from '@xyflow/react'
import LoopConfigPanel from './LoopConfigPanel'
import ParallelConfigPanel from './ParallelConfigPanel'
import SlackConfigPanel from './SlackConfigPanel'
import PrioritySelector from '../webhooks/PrioritySelector'
import { AIConfigPanel } from '../ai/AIConfigPanel'
import type { AIConfigData } from '../ai/AIConfigPanel'
import type { AIAction } from '../../types/ai'
import { AI_MODELS } from '../../types/ai'

// Mock credentials for now - in production this would come from useCredentials hook
const mockCredentials = [
  { id: 'cred-openai', tenantId: 't1', name: 'OpenAI API Key', type: 'api_key' as const, createdAt: '', updatedAt: '' },
  { id: 'cred-anthropic', tenantId: 't1', name: 'Anthropic API Key', type: 'api_key' as const, createdAt: '', updatedAt: '' },
]

// Map node types to AI actions
const AI_NODE_TYPE_MAP: Record<string, AIAction> = {
  ai_chat: 'chat_completion',
  ai_summarize: 'summarization',
  ai_classify: 'classification',
  ai_extract: 'entity_extraction',
  ai_embed: 'embedding',
}

const isAINodeType = (nodeType: string): boolean => {
  return nodeType in AI_NODE_TYPE_MAP
}

interface PropertyPanelProps {
  node: Node | null
  onUpdate: (nodeId: string, data: any) => void
  onClose: () => void
  onSave?: () => Promise<void>
  isSaving?: boolean
}

export default function PropertyPanel({ node, onUpdate, onClose, onSave, isSaving }: PropertyPanelProps) {
  const [formData, setFormData] = useState<any>({})
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [successMessage, setSuccessMessage] = useState<string | null>(null)
  const [originalData, setOriginalData] = useState<any>({})

  // Reset form when node changes
  useEffect(() => {
    if (node) {
      const data = node.data || {}
      setFormData(data)
      setOriginalData(data)
      setErrors({})
      setSuccessMessage(null)
    }
  }, [node])

  if (!node) {
    return (
      <div className="w-80 bg-gray-800 border-l border-gray-700 p-6 flex items-center justify-center">
        <div className="text-center text-gray-400">No node selected</div>
      </div>
    )
  }

  const nodeType: string = typeof node.data?.nodeType === 'string' ? node.data.nodeType : 'unknown'

  const handleChange = (field: string, value: any) => {
    setFormData((prev: any) => ({
      ...prev,
      [field]: value,
    }))
    // Clear error for this field
    if (errors[field]) {
      setErrors((prev) => {
        const newErrors = { ...prev }
        delete newErrors[field]
        return newErrors
      })
    }
    setSuccessMessage(null)
  }

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {}

    // Name is required for all nodes
    if (!formData.label || formData.label.trim() === '') {
      newErrors.label = 'Name is required'
    }

    // Node-specific validation
    if (nodeType === 'http') {
      if (formData.url && !isValidUrl(formData.url)) {
        newErrors.url = 'Invalid URL format'
      }
    }

    // Slack action validation
    if (nodeType === 'slack_send_message') {
      if (!formData.channel || formData.channel.trim() === '') {
        newErrors.channel = 'Channel ID is required'
      }
      if (formData.blocks && formData.blocks.trim() !== '') {
        try {
          JSON.parse(formData.blocks)
        } catch {
          newErrors.blocks = 'Invalid JSON format'
        }
      }
    }

    if (nodeType === 'slack_send_dm') {
      if (!formData.user || formData.user.trim() === '') {
        newErrors.user = 'User email or ID is required'
      }
      if (formData.blocks && formData.blocks.trim() !== '') {
        try {
          JSON.parse(formData.blocks)
        } catch {
          newErrors.blocks = 'Invalid JSON format'
        }
      }
    }

    if (nodeType === 'slack_update_message') {
      if (!formData.channel || formData.channel.trim() === '') {
        newErrors.channel = 'Channel ID is required'
      }
      if (!formData.ts || formData.ts.trim() === '') {
        newErrors.ts = 'Message timestamp is required'
      }
      if (formData.blocks && formData.blocks.trim() !== '') {
        try {
          JSON.parse(formData.blocks)
        } catch {
          newErrors.blocks = 'Invalid JSON format'
        }
      }
    }

    if (nodeType === 'slack_add_reaction') {
      if (!formData.channel || formData.channel.trim() === '') {
        newErrors.channel = 'Channel ID is required'
      }
      if (!formData.timestamp || formData.timestamp.trim() === '') {
        newErrors.timestamp = 'Message timestamp is required'
      }
      if (!formData.emoji || formData.emoji.trim() === '') {
        newErrors.emoji = 'Emoji name is required'
      }
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const isValidUrl = (url: string): boolean => {
    try {
      new URL(url)
      return true
    } catch {
      return false
    }
  }

  const handleSave = async () => {
    if (!validate()) {
      return
    }

    // Update local state first
    onUpdate(node.id, formData)
    setOriginalData(formData)

    // Persist to backend if onSave is provided
    if (onSave) {
      try {
        await onSave()
        setSuccessMessage('Saved successfully')
        setTimeout(() => setSuccessMessage(null), 3000)
      } catch {
        // Error is handled by parent component
      }
    } else {
      setSuccessMessage('Saved successfully')
      setTimeout(() => setSuccessMessage(null), 3000)
    }
  }

  const handleReset = () => {
    setFormData(originalData)
    setErrors({})
    setSuccessMessage(null)
  }

  return (
    <div className="w-80 bg-gray-800 border-l border-gray-700 flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b border-gray-700">
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-white font-semibold text-lg">Node Properties</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-white transition-colors"
            aria-label="Close"
          >
            ✕
          </button>
        </div>
        <div className="text-gray-400 text-sm">{formData.label || 'Unnamed Node'}</div>
      </div>

      {/* Success Message */}
      {successMessage && (
        <div className="mx-4 mt-4 p-3 bg-green-900/20 border border-green-500/30 text-green-400 text-sm rounded">
          {successMessage}
        </div>
      )}

      {/* Form */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {/* Common Fields */}
        <div>
          <label htmlFor="node-name" className="block text-sm font-medium text-gray-300 mb-2">
            Name *
          </label>
          <input
            id="node-name"
            type="text"
            value={formData.label || ''}
            onChange={(e) => handleChange('label', e.target.value)}
            className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
          />
          {errors.label && <div className="mt-1 text-xs text-red-400">{errors.label}</div>}
        </div>

        {/* Webhook Trigger Fields */}
        {nodeType === 'webhook' && <WebhookFields formData={formData} onChange={handleChange} />}

        {/* HTTP Action Fields */}
        {nodeType === 'http' && <HttpFields formData={formData} onChange={handleChange} errors={errors} />}

        {/* Transform Action Fields */}
        {nodeType === 'transform' && <TransformFields formData={formData} onChange={handleChange} />}

        {/* Conditional Control Fields */}
        {nodeType === 'conditional' && <ConditionalFields formData={formData} onChange={handleChange} />}

        {/* Loop Control Fields */}
        {nodeType === 'loop' && (
          <LoopConfigPanel
            config={{
              source: formData.source || '',
              itemVariable: formData.itemVariable || '',
              indexVariable: formData.indexVariable || '',
              maxIterations: formData.maxIterations || 1000,
              onError: formData.onError || 'stop',
            }}
            onChange={(loopConfig) => {
              // Update all loop fields at once
              Object.entries(loopConfig).forEach(([key, value]) => {
                handleChange(key, value)
              })
            }}
          />
        )}

        {/* Parallel Control Fields */}
        {nodeType === 'parallel' && (
          <ParallelConfigPanel
            config={{
              errorStrategy: formData.errorStrategy || 'fail_fast',
              maxConcurrency: formData.maxConcurrency ?? 0,
            }}
            onChange={(parallelConfig) => {
              // Update all parallel fields at once
              Object.entries(parallelConfig).forEach(([key, value]) => {
                handleChange(key, value)
              })
            }}
          />
        )}

        {/* Slack Action Fields */}
        {(nodeType === 'slack_send_message' ||
          nodeType === 'slack_send_dm' ||
          nodeType === 'slack_update_message' ||
          nodeType === 'slack_add_reaction') && (
          <SlackConfigPanel
            nodeType={nodeType}
            formData={formData}
            onChange={handleChange}
            errors={errors}
          />
        )}

        {/* AI Action Fields */}
        {isAINodeType(nodeType) && (
          <AIConfigPanel
            action={AI_NODE_TYPE_MAP[nodeType]}
            config={formData.aiConfig || {}}
            models={AI_MODELS}
            credentials={mockCredentials}
            onChange={(config: AIConfigData) => handleChange('aiConfig', config)}
            variables={['trigger.data', 'steps.previous.output', 'env.api_url']}
          />
        )}
      </div>

      {/* Actions */}
      <div className="p-4 border-t border-gray-700 flex space-x-2">
        <button
          onClick={handleReset}
          disabled={isSaving}
          className="flex-1 px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          Reset
        </button>
        <button
          onClick={handleSave}
          disabled={isSaving}
          className="flex-1 px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {isSaving ? 'Saving...' : 'Save'}
        </button>
      </div>
    </div>
  )
}

// Webhook trigger fields
function WebhookFields({ formData, onChange }: any) {
  return (
    <>
      <div>
        <label htmlFor="webhook-path" className="block text-sm font-medium text-gray-300 mb-2">
          Path
        </label>
        <input
          id="webhook-path"
          type="text"
          value={formData.path || ''}
          onChange={(e) => onChange('path', e.target.value)}
          placeholder="/webhook"
          className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
        />
      </div>

      <div>
        <label htmlFor="webhook-method" className="block text-sm font-medium text-gray-300 mb-2">
          Method
        </label>
        <select
          id="webhook-method"
          value={formData.method || 'POST'}
          onChange={(e) => onChange('method', e.target.value)}
          className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
        >
          <option value="GET">GET</option>
          <option value="POST">POST</option>
          <option value="PUT">PUT</option>
          <option value="DELETE">DELETE</option>
          <option value="PATCH">PATCH</option>
        </select>
      </div>

      <PrioritySelector
        value={formData.priority ?? 1}
        onChange={(priority) => onChange('priority', priority)}
        id="webhook-priority"
      />
    </>
  )
}

// HTTP action fields
function HttpFields({ formData, onChange, errors }: any) {
  return (
    <>
      <div>
        <label htmlFor="http-url" className="block text-sm font-medium text-gray-300 mb-2">
          URL
        </label>
        <input
          id="http-url"
          type="text"
          value={formData.url || ''}
          onChange={(e) => onChange('url', e.target.value)}
          placeholder="https://api.example.com"
          className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
        />
        {errors.url && <div className="mt-1 text-xs text-red-400">{errors.url}</div>}
      </div>

      <div>
        <label htmlFor="http-method" className="block text-sm font-medium text-gray-300 mb-2">
          Method
        </label>
        <select
          id="http-method"
          value={formData.method || 'GET'}
          onChange={(e) => onChange('method', e.target.value)}
          className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
        >
          <option value="GET">GET</option>
          <option value="POST">POST</option>
          <option value="PUT">PUT</option>
          <option value="DELETE">DELETE</option>
          <option value="PATCH">PATCH</option>
        </select>
      </div>

      <div>
        <label htmlFor="http-headers" className="block text-sm font-medium text-gray-300 mb-2">
          Headers (JSON)
        </label>
        <textarea
          id="http-headers"
          value={formData.headers || ''}
          onChange={(e) => onChange('headers', e.target.value)}
          placeholder='{"Content-Type": "application/json"}'
          rows={3}
          className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 font-mono"
        />
      </div>

      <div>
        <label htmlFor="http-body" className="block text-sm font-medium text-gray-300 mb-2">
          Body
        </label>
        <textarea
          id="http-body"
          value={formData.body || ''}
          onChange={(e) => onChange('body', e.target.value)}
          placeholder='{"key": "value"}'
          rows={4}
          className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 font-mono"
        />
      </div>
    </>
  )
}

// Transform action fields
function TransformFields({ formData, onChange }: any) {
  return (
    <div>
      <label htmlFor="transform-mapping" className="block text-sm font-medium text-gray-300 mb-2">
        Mapping (JSON)
      </label>
      <textarea
        id="transform-mapping"
        value={formData.mapping || ''}
        onChange={(e) => onChange('mapping', e.target.value)}
        placeholder='{"output": "$.input.data"}'
        rows={6}
        className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 font-mono"
      />
    </div>
  )
}

// Conditional control fields
function ConditionalFields({ formData, onChange }: any) {
  return (
    <div>
      <label htmlFor="conditional-condition" className="block text-sm font-medium text-gray-300 mb-2">
        Condition
      </label>
      <textarea
        id="conditional-condition"
        value={formData.condition || ''}
        onChange={(e) => onChange('condition', e.target.value)}
        placeholder='$.status === "success"'
        rows={4}
        className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 font-mono"
      />
    </div>
  )
}
