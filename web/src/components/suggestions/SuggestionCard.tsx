import { FC } from 'react'
import type { Suggestion } from '../../types/suggestions'
import {
  getCategoryLabel,
  getTypeLabel,
  getConfidenceLabel,
  getCategoryColor,
  getConfidenceColor,
} from '../../types/suggestions'

interface SuggestionCardProps {
  suggestion: Suggestion
  onApply?: () => void
  onDismiss?: () => void
  onSelect?: () => void
  isSelected?: boolean
  isApplying?: boolean
  isDismissing?: boolean
  showActions?: boolean
}

const categoryIcons: Record<string, string> = {
  network: '🌐',
  auth: '🔐',
  data: '📊',
  rate_limit: '⏱️',
  timeout: '⏰',
  config: '⚙️',
  external_service: '🔗',
  unknown: '❓',
}

const typeIcons: Record<string, string> = {
  retry: '🔄',
  config_change: '⚙️',
  credential_update: '🔑',
  data_fix: '🔧',
  workflow_modification: '📝',
  manual_intervention: '👤',
}

export const SuggestionCard: FC<SuggestionCardProps> = ({
  suggestion,
  onApply,
  onDismiss,
  onSelect,
  isSelected = false,
  isApplying = false,
  isDismissing = false,
  showActions = true,
}) => {
  const categoryColor = getCategoryColor(suggestion.category)
  const confidenceColor = getConfidenceColor(suggestion.confidence)

  const handleClick = () => {
    if (onSelect) {
      onSelect()
    }
  }

  return (
    <div
      className={`rounded-lg border p-4 transition-all ${
        isSelected
          ? 'border-blue-500 bg-blue-50'
          : 'border-gray-200 bg-white hover:border-gray-300'
      } ${onSelect ? 'cursor-pointer' : ''} ${
        suggestion.status !== 'pending' ? 'opacity-60' : ''
      }`}
      onClick={handleClick}
      role={onSelect ? 'button' : undefined}
      tabIndex={onSelect ? 0 : undefined}
    >
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-center space-x-2">
          <span className="text-lg" title={getCategoryLabel(suggestion.category)}>
            {categoryIcons[suggestion.category] || '❓'}
          </span>
          <div>
            <h4 className="font-medium text-gray-900">{suggestion.title}</h4>
            <div className="mt-1 flex items-center space-x-2">
              <span
                className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-${categoryColor}-100 text-${categoryColor}-800`}
              >
                {getCategoryLabel(suggestion.category)}
              </span>
              <span
                className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-${confidenceColor}-100 text-${confidenceColor}-800`}
              >
                {getConfidenceLabel(suggestion.confidence)} confidence
              </span>
            </div>
          </div>
        </div>
        <span className="text-sm text-gray-500" title={`Source: ${suggestion.source}`}>
          {suggestion.source === 'llm' ? '🤖' : '📋'}
        </span>
      </div>

      {/* Description */}
      <p className="mt-3 text-sm text-gray-600">{suggestion.description}</p>

      {/* Details */}
      {suggestion.details && (
        <p className="mt-2 text-xs text-gray-500 italic">{suggestion.details}</p>
      )}

      {/* Fix preview */}
      {suggestion.fix && (
        <div className="mt-3 rounded bg-gray-50 p-2">
          <div className="flex items-center space-x-1 text-xs text-gray-700">
            <span>{typeIcons[suggestion.type] || '🔧'}</span>
            <span className="font-medium">{getTypeLabel(suggestion.type)}</span>
          </div>
          {suggestion.fix.config_path && (
            <p className="mt-1 text-xs text-gray-500">
              Config: <code className="rounded bg-gray-200 px-1">{suggestion.fix.config_path}</code>
              {suggestion.fix.new_value !== undefined && (
                <> → <code className="rounded bg-green-100 px-1">{JSON.stringify(suggestion.fix.new_value)}</code></>
              )}
            </p>
          )}
          {suggestion.fix.retry_config && (
            <p className="mt-1 text-xs text-gray-500">
              Max retries: {suggestion.fix.retry_config.max_retries},
              Backoff: {suggestion.fix.retry_config.backoff_ms}ms × {suggestion.fix.retry_config.backoff_factor}
            </p>
          )}
        </div>
      )}

      {/* Actions */}
      {showActions && suggestion.status === 'pending' && (onApply || onDismiss) && (
        <div className="mt-4 flex items-center justify-end space-x-2">
          {onDismiss && (
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation()
                onDismiss()
              }}
              disabled={isDismissing || isApplying}
              className="rounded border border-gray-300 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
            >
              {isDismissing ? 'Dismissing...' : 'Dismiss'}
            </button>
          )}
          {onApply && (
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation()
                onApply()
              }}
              disabled={isApplying || isDismissing}
              className="rounded bg-blue-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {isApplying ? 'Applying...' : 'Apply Fix'}
            </button>
          )}
        </div>
      )}

      {/* Status indicator for non-pending */}
      {suggestion.status !== 'pending' && (
        <div className="mt-3 flex items-center text-sm">
          {suggestion.status === 'applied' && (
            <span className="text-green-600">✓ Applied</span>
          )}
          {suggestion.status === 'dismissed' && (
            <span className="text-gray-500">✗ Dismissed</span>
          )}
        </div>
      )}
    </div>
  )
}

export default SuggestionCard
