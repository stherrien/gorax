import React, { useState, useRef, useEffect, useMemo } from 'react'
import type { AIModel, AIProvider } from '../../types/ai'

export interface ModelSelectorProps {
  models: AIModel[]
  value: string
  onChange: (modelId: string) => void
  filterProvider?: AIProvider
  filterCapability?: string
  placeholder?: string
  disabled?: boolean
  loading?: boolean
  showPricing?: boolean
  showContext?: boolean
  showProviderBadge?: boolean
  groupByProvider?: boolean
  searchable?: boolean
}

const formatCost = (costPer1M: number): string => {
  const dollars = costPer1M / 100
  return `$${dollars.toFixed(2)}`
}

const formatContextWindow = (tokens: number): string => {
  if (tokens >= 1000000) {
    return `${(tokens / 1000000).toFixed(1)}M`
  }
  if (tokens >= 1000) {
    return `${Math.round(tokens / 1000)}K`
  }
  return tokens.toString()
}

const providerLabels: Record<AIProvider, string> = {
  openai: 'OpenAI',
  anthropic: 'Anthropic',
  bedrock: 'AWS Bedrock',
}

export const ModelSelector: React.FC<ModelSelectorProps> = ({
  models,
  value,
  onChange,
  filterProvider,
  filterCapability,
  placeholder = 'Select Model',
  disabled = false,
  loading = false,
  showPricing = false,
  showContext = false,
  showProviderBadge = false,
  groupByProvider = false,
  searchable = false,
}) => {
  const [isOpen, setIsOpen] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')
  const dropdownRef = useRef<HTMLDivElement>(null)

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false)
        setSearchTerm('')
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside)
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [isOpen])

  // Filter models
  const filteredModels = useMemo(() => {
    let result = models

    // Filter by provider
    if (filterProvider) {
      result = result.filter((m) => m.provider === filterProvider)
    }

    // Filter by capability
    if (filterCapability) {
      result = result.filter((m) => m.capabilities.includes(filterCapability))
    }

    // Filter by search term
    if (searchTerm) {
      const search = searchTerm.toLowerCase()
      result = result.filter(
        (m) =>
          m.name.toLowerCase().includes(search) ||
          m.id.toLowerCase().includes(search) ||
          m.provider.toLowerCase().includes(search)
      )
    }

    return result
  }, [models, filterProvider, filterCapability, searchTerm])

  // Group models by provider
  const groupedModels = useMemo(() => {
    if (!groupByProvider) return null

    const groups: Record<AIProvider, AIModel[]> = {
      openai: [],
      anthropic: [],
      bedrock: [],
    }

    filteredModels.forEach((model) => {
      groups[model.provider].push(model)
    })

    return groups
  }, [filteredModels, groupByProvider])

  const selectedModel = models.find((m) => m.id === value)

  const handleSelect = (modelId: string) => {
    onChange(modelId)
    setIsOpen(false)
    setSearchTerm('')
  }

  const handleToggle = () => {
    if (!disabled && !loading) {
      setIsOpen(!isOpen)
    }
  }

  const renderModelOption = (model: AIModel) => (
    <button
      key={model.id}
      type="button"
      onClick={() => handleSelect(model.id)}
      className="w-full px-4 py-2 text-left hover:bg-gray-100 focus:bg-gray-100 focus:outline-none"
    >
      <div className="flex items-center justify-between">
        <div className="flex-1 min-w-0">
          <div className="text-sm font-medium text-gray-900">{model.name}</div>
          <div className="flex items-center gap-2 text-xs text-gray-500">
            {showPricing && (
              <span>
                {formatCost(model.inputCostPer1M)} input / {formatCost(model.outputCostPer1M)} output
              </span>
            )}
            {showContext && <span>{formatContextWindow(model.contextWindow)} context</span>}
          </div>
        </div>
        {!groupByProvider && (
          <span className="ml-2 px-2 py-1 text-xs font-medium bg-gray-100 text-gray-600 rounded">
            {model.provider}
          </span>
        )}
      </div>
    </button>
  )

  return (
    <div ref={dropdownRef} className="relative">
      <button
        type="button"
        onClick={handleToggle}
        disabled={disabled || loading}
        className="w-full px-3 py-2 text-left bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed flex items-center justify-between"
      >
        <div className="flex items-center gap-2 flex-1 min-w-0">
          <span className={selectedModel ? 'text-gray-900' : 'text-gray-500'}>
            {loading ? 'Loading...' : selectedModel ? selectedModel.name : placeholder}
          </span>
          {showProviderBadge && selectedModel && (
            <span className="px-2 py-0.5 text-xs font-medium bg-gray-100 text-gray-600 rounded">
              {selectedModel.provider}
            </span>
          )}
        </div>
        <svg
          className={`w-5 h-5 text-gray-400 transition-transform ${isOpen ? 'rotate-180' : ''}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {isOpen && (
        <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-80 overflow-auto">
          {searchable && (
            <div className="sticky top-0 bg-white border-b border-gray-200 p-2">
              <input
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="Search models..."
                className="w-full px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                onClick={(e) => e.stopPropagation()}
                autoFocus
              />
            </div>
          )}

          {filteredModels.length === 0 ? (
            <div className="px-4 py-6 text-center text-gray-500 text-sm">No models available</div>
          ) : groupByProvider && groupedModels ? (
            <div>
              {(Object.keys(groupedModels) as AIProvider[]).map((provider) => {
                const providerModels = groupedModels[provider]
                if (providerModels.length === 0) return null

                return (
                  <div key={provider}>
                    <div className="px-4 py-2 text-xs font-semibold text-gray-500 bg-gray-50 border-b border-gray-200">
                      {providerLabels[provider]}
                    </div>
                    <div>{providerModels.map(renderModelOption)}</div>
                  </div>
                )
              })}
            </div>
          ) : (
            <div className="py-1">{filteredModels.map(renderModelOption)}</div>
          )}
        </div>
      )}
    </div>
  )
}
