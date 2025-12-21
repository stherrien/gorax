import React, { useState, useMemo } from 'react'
import type { AIAction, AIModel } from '../../types/ai'
import type { Credential } from '../../api/credentials'
import { ModelSelector } from './ModelSelector'
import { PromptEditor } from './PromptEditor'

export interface AIConfigData {
  action?: AIAction
  credentialId?: string
  model?: string
  systemPrompt?: string
  text?: string
  maxLength?: number
  format?: 'paragraph' | 'bullets'
  focus?: string
  categories?: string[]
  multiLabel?: boolean
  description?: string
  entityTypes?: string[]
  customEntities?: Record<string, string>
  texts?: string[]
  maxTokens?: number
  temperature?: number
  topP?: number
}

export interface AIConfigPanelProps {
  action: AIAction
  config: AIConfigData
  models: AIModel[]
  credentials: Credential[]
  onChange: (config: AIConfigData) => void
  variables?: string[]
  error?: string
  disabled?: boolean
}

const ACTION_OPTIONS: { value: AIAction; label: string }[] = [
  { value: 'chat_completion', label: 'Chat Completion' },
  { value: 'summarization', label: 'Summarization' },
  { value: 'classification', label: 'Classification' },
  { value: 'entity_extraction', label: 'Entity Extraction' },
  { value: 'embedding', label: 'Embedding' },
]

const FORMAT_OPTIONS = [
  { value: 'paragraph', label: 'Paragraph' },
  { value: 'bullets', label: 'Bullet Points' },
]

const COMMON_ENTITY_TYPES = [
  'person',
  'organization',
  'location',
  'date',
  'email',
  'phone',
  'money',
  'product',
]

export const AIConfigPanel: React.FC<AIConfigPanelProps> = ({
  action,
  config,
  models,
  credentials,
  onChange,
  variables = [],
  error,
  disabled = false,
}) => {
  const [showAdvanced, setShowAdvanced] = useState(false)

  // Filter models based on action capability
  const filteredModels = useMemo(() => {
    if (action === 'embedding') {
      return models.filter((m) => m.capabilities.includes('embedding'))
    }
    return models.filter((m) => m.capabilities.includes('chat'))
  }, [models, action])

  const handleActionChange = (newAction: AIAction) => {
    onChange({ ...config, action: newAction })
  }

  const handleChange = (field: keyof AIConfigData, value: unknown) => {
    onChange({ ...config, [field]: value })
  }

  const handleCategoriesChange = (value: string) => {
    const categories = value
      .split(',')
      .map((c) => c.trim())
      .filter((c) => c)
    handleChange('categories', categories)
  }

  const handleEntityTypesChange = (value: string) => {
    const types = value
      .split(',')
      .map((t) => t.trim())
      .filter((t) => t)
    handleChange('entityTypes', types)
  }

  const handleTextsChange = (value: string) => {
    const texts = value.split('\n').filter((t) => t.trim())
    handleChange('texts', texts)
  }

  return (
    <div className="space-y-4">
      {error && (
        <div className="p-3 bg-red-50 border border-red-200 text-red-700 rounded-md text-sm">
          {error}
        </div>
      )}

      {/* Action Type */}
      <div>
        <label htmlFor="ai-action" className="block text-sm font-medium text-gray-700 mb-1">
          Action Type
        </label>
        <select
          id="ai-action"
          value={action}
          onChange={(e) => handleActionChange(e.target.value as AIAction)}
          disabled={disabled}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:outline-none disabled:bg-gray-100"
        >
          {ACTION_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
      </div>

      {/* Credential Selector */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Credential</label>
        <select
          value={config.credentialId || ''}
          onChange={(e) => handleChange('credentialId', e.target.value)}
          disabled={disabled}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:outline-none disabled:bg-gray-100"
        >
          <option value="">Select Credential</option>
          {credentials.map((cred) => (
            <option key={cred.id} value={cred.id}>
              {cred.name}
            </option>
          ))}
        </select>
      </div>

      {/* Model Selector */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Model</label>
        <ModelSelector
          models={filteredModels}
          value={config.model || ''}
          onChange={(model) => handleChange('model', model)}
          disabled={disabled}
          showPricing
          groupByProvider
        />
      </div>

      {/* Action-specific fields */}
      {action === 'chat_completion' && (
        <ChatCompletionFields
          config={config}
          onChange={handleChange}
          variables={variables}
          disabled={disabled}
        />
      )}

      {action === 'summarization' && (
        <SummarizationFields
          config={config}
          onChange={handleChange}
          onFormatChange={(format) => handleChange('format', format)}
          variables={variables}
          disabled={disabled}
        />
      )}

      {action === 'classification' && (
        <ClassificationFields
          config={config}
          onChange={handleChange}
          onCategoriesChange={handleCategoriesChange}
          variables={variables}
          disabled={disabled}
        />
      )}

      {action === 'entity_extraction' && (
        <EntityExtractionFields
          config={config}
          onChange={handleChange}
          onEntityTypesChange={handleEntityTypesChange}
          disabled={disabled}
          variables={variables}
        />
      )}

      {action === 'embedding' && (
        <EmbeddingFields
          config={config}
          onTextsChange={handleTextsChange}
          variables={variables}
          disabled={disabled}
        />
      )}

      {/* Advanced Options */}
      {action !== 'embedding' && (
        <div className="border-t border-gray-200 pt-4">
          <button
            type="button"
            onClick={() => setShowAdvanced(!showAdvanced)}
            className="flex items-center text-sm font-medium text-gray-700 hover:text-gray-900"
          >
            <svg
              className={`w-4 h-4 mr-1 transition-transform ${showAdvanced ? 'rotate-90' : ''}`}
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 5l7 7-7 7"
              />
            </svg>
            Advanced Options
          </button>

          {showAdvanced && (
            <div className="mt-4 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Temperature
                  <span className="text-gray-500 font-normal ml-1">(0-2)</span>
                </label>
                <input
                  type="number"
                  min="0"
                  max="2"
                  step="0.1"
                  value={config.temperature ?? 0.7}
                  onChange={(e) => handleChange('temperature', parseFloat(e.target.value))}
                  disabled={disabled}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:outline-none disabled:bg-gray-100"
                />
                <p className="text-xs text-gray-500 mt-1">
                  Higher values make output more random, lower values more focused.
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Max Tokens</label>
                <input
                  type="number"
                  min="1"
                  value={config.maxTokens ?? ''}
                  onChange={(e) =>
                    handleChange(
                      'maxTokens',
                      e.target.value ? parseInt(e.target.value, 10) : undefined
                    )
                  }
                  placeholder="Default (model-specific)"
                  disabled={disabled}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:outline-none disabled:bg-gray-100"
                />
                <p className="text-xs text-gray-500 mt-1">
                  Maximum number of tokens in the response.
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Top P
                  <span className="text-gray-500 font-normal ml-1">(0-1)</span>
                </label>
                <input
                  type="number"
                  min="0"
                  max="1"
                  step="0.1"
                  value={config.topP ?? ''}
                  onChange={(e) =>
                    handleChange('topP', e.target.value ? parseFloat(e.target.value) : undefined)
                  }
                  placeholder="Default (1.0)"
                  disabled={disabled}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:outline-none disabled:bg-gray-100"
                />
                <p className="text-xs text-gray-500 mt-1">
                  Nucleus sampling: only consider tokens with top_p probability mass.
                </p>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// Chat Completion Fields
function ChatCompletionFields({
  config,
  onChange,
  variables,
  disabled,
}: {
  config: AIConfigData
  onChange: (field: keyof AIConfigData, value: unknown) => void
  variables: string[]
  disabled: boolean
}) {
  return (
    <>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">System Prompt</label>
        <PromptEditor
          value={config.systemPrompt || ''}
          onChange={(value) => onChange('systemPrompt', value)}
          placeholder="You are a helpful assistant..."
          rows={3}
          showVariables
          variables={variables}
          disabled={disabled}
          helperText="Optional system message to set the AI's behavior."
        />
      </div>
    </>
  )
}

// Summarization Fields
function SummarizationFields({
  config,
  onChange,
  onFormatChange,
  variables,
  disabled,
}: {
  config: AIConfigData
  onChange: (field: keyof AIConfigData, value: unknown) => void
  onFormatChange: (format: 'paragraph' | 'bullets') => void
  variables: string[]
  disabled: boolean
}) {
  return (
    <>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Text to Summarize</label>
        <PromptEditor
          value={config.text || ''}
          onChange={(value) => onChange('text', value)}
          placeholder="Enter text to summarize or use {{variable}}..."
          rows={5}
          showVariables
          variables={variables}
          disabled={disabled}
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Summary Format</label>
        <select
          value={config.format || 'paragraph'}
          onChange={(e) => onFormatChange(e.target.value as 'paragraph' | 'bullets')}
          disabled={disabled}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:outline-none disabled:bg-gray-100"
        >
          {FORMAT_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Max Length
          <span className="text-gray-500 font-normal ml-1">(words)</span>
        </label>
        <input
          type="number"
          min="1"
          value={config.maxLength ?? ''}
          onChange={(e) =>
            onChange('maxLength', e.target.value ? parseInt(e.target.value, 10) : undefined)
          }
          placeholder="Default (automatic)"
          disabled={disabled}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:outline-none disabled:bg-gray-100"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Focus Area</label>
        <input
          type="text"
          value={config.focus || ''}
          onChange={(e) => onChange('focus', e.target.value)}
          placeholder="e.g., key findings, action items..."
          disabled={disabled}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:outline-none disabled:bg-gray-100"
        />
        <p className="text-xs text-gray-500 mt-1">
          Optional focus area for the summary.
        </p>
      </div>
    </>
  )
}

// Classification Fields
function ClassificationFields({
  config,
  onChange,
  onCategoriesChange,
  variables,
  disabled,
}: {
  config: AIConfigData
  onChange: (field: keyof AIConfigData, value: unknown) => void
  onCategoriesChange: (value: string) => void
  variables: string[]
  disabled: boolean
}) {
  return (
    <>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Text to Classify</label>
        <PromptEditor
          value={config.text || ''}
          onChange={(value) => onChange('text', value)}
          placeholder="Enter text to classify or use {{variable}}..."
          rows={4}
          showVariables
          variables={variables}
          disabled={disabled}
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Categories</label>
        <input
          type="text"
          value={config.categories?.join(', ') || ''}
          onChange={(e) => onCategoriesChange(e.target.value)}
          placeholder="positive, negative, neutral"
          disabled={disabled}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:outline-none disabled:bg-gray-100"
        />
        <p className="text-xs text-gray-500 mt-1">Comma-separated list of categories.</p>
      </div>

      <div className="flex items-center">
        <input
          type="checkbox"
          id="multi-label"
          checked={config.multiLabel || false}
          onChange={(e) => onChange('multiLabel', e.target.checked)}
          disabled={disabled}
          className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
        />
        <label htmlFor="multi-label" className="ml-2 text-sm text-gray-700">
          Allow multiple categories
        </label>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Category Description
        </label>
        <textarea
          value={config.description || ''}
          onChange={(e) => onChange('description', e.target.value)}
          placeholder="Describe what each category means..."
          rows={3}
          disabled={disabled}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:outline-none disabled:bg-gray-100"
        />
        <p className="text-xs text-gray-500 mt-1">Optional context about the categories.</p>
      </div>
    </>
  )
}

// Entity Extraction Fields
function EntityExtractionFields({
  config,
  onChange,
  onEntityTypesChange,
  variables,
  disabled,
}: {
  config: AIConfigData
  onChange: (field: keyof AIConfigData, value: unknown) => void
  onEntityTypesChange: (value: string) => void
  variables: string[]
  disabled: boolean
}) {
  return (
    <>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Text to Analyze</label>
        <PromptEditor
          value={config.text || ''}
          onChange={(value) => onChange('text', value)}
          placeholder="Enter text to extract entities from..."
          rows={5}
          showVariables
          variables={variables}
          disabled={disabled}
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Entity Types</label>
        <input
          type="text"
          value={config.entityTypes?.join(', ') || ''}
          onChange={(e) => onEntityTypesChange(e.target.value)}
          placeholder="person, organization, location..."
          disabled={disabled}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:outline-none disabled:bg-gray-100"
        />
        <p className="text-xs text-gray-500 mt-1">
          Common types: {COMMON_ENTITY_TYPES.join(', ')}
        </p>
      </div>
    </>
  )
}

// Embedding Fields
function EmbeddingFields({
  config,
  onTextsChange,
  variables,
  disabled,
}: {
  config: AIConfigData
  onTextsChange: (value: string) => void
  variables: string[]
  disabled: boolean
}) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1">Texts to Embed</label>
      <PromptEditor
        value={config.texts?.join('\n') || ''}
        onChange={onTextsChange}
        placeholder="Enter one text per line to generate embeddings..."
        rows={6}
        showVariables
        variables={variables}
        disabled={disabled}
        helperText="Enter one text per line. Each text will get its own embedding vector."
      />
    </div>
  )
}
