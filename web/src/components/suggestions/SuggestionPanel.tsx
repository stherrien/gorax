import { FC, useState } from 'react'
import type { Suggestion } from '../../types/suggestions'
import { SuggestionCard } from './SuggestionCard'
import { sortSuggestionsByConfidence, groupSuggestionsByCategory } from '../../hooks/useSuggestions'
import { getCategoryLabel } from '../../types/suggestions'

interface SuggestionPanelProps {
  suggestions: Suggestion[]
  isLoading?: boolean
  error?: Error | null
  onApply?: (suggestionId: string) => Promise<void>
  onDismiss?: (suggestionId: string) => Promise<void>
  onRefresh?: () => void
  groupByCategory?: boolean
  showHeader?: boolean
  emptyMessage?: string
}

type ViewMode = 'all' | 'pending' | 'applied' | 'dismissed'

export const SuggestionPanel: FC<SuggestionPanelProps> = ({
  suggestions,
  isLoading = false,
  error = null,
  onApply,
  onDismiss,
  onRefresh,
  groupByCategory = false,
  showHeader = true,
  emptyMessage = 'No suggestions available',
}) => {
  const [viewMode, setViewMode] = useState<ViewMode>('pending')
  const [applyingId, setApplyingId] = useState<string | null>(null)
  const [dismissingId, setDismissingId] = useState<string | null>(null)
  const [selectedId, setSelectedId] = useState<string | null>(null)

  // Filter suggestions by view mode
  const filteredSuggestions = suggestions.filter((s) => {
    if (viewMode === 'all') return true
    return s.status === viewMode
  })

  // Sort by confidence
  const sortedSuggestions = sortSuggestionsByConfidence(filteredSuggestions)

  // Optionally group by category
  const groupedSuggestions = groupByCategory
    ? groupSuggestionsByCategory(sortedSuggestions)
    : { all: sortedSuggestions }

  const handleApply = async (suggestionId: string) => {
    if (!onApply) return
    setApplyingId(suggestionId)
    try {
      await onApply(suggestionId)
    } finally {
      setApplyingId(null)
    }
  }

  const handleDismiss = async (suggestionId: string) => {
    if (!onDismiss) return
    setDismissingId(suggestionId)
    try {
      await onDismiss(suggestionId)
    } finally {
      setDismissingId(null)
    }
  }

  // Count by status
  const pendingCount = suggestions.filter((s) => s.status === 'pending').length
  const appliedCount = suggestions.filter((s) => s.status === 'applied').length
  const dismissedCount = suggestions.filter((s) => s.status === 'dismissed').length

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-500 border-t-transparent" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-lg bg-red-50 p-4 text-red-700">
        <p className="font-medium">Error loading suggestions</p>
        <p className="mt-1 text-sm">{error.message}</p>
        {onRefresh && (
          <button
            onClick={onRefresh}
            className="mt-2 text-sm text-red-600 underline hover:text-red-800"
          >
            Try again
          </button>
        )}
      </div>
    )
  }

  return (
    <div className="rounded-lg border border-gray-200 bg-white">
      {showHeader && (
        <div className="border-b border-gray-200 px-4 py-3">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-medium text-gray-900">
              Smart Suggestions
              {suggestions.length > 0 && (
                <span className="ml-2 text-sm text-gray-500">({suggestions.length})</span>
              )}
            </h3>
            {onRefresh && (
              <button
                onClick={onRefresh}
                className="text-sm text-blue-600 hover:text-blue-800"
                title="Refresh suggestions"
              >
                🔄 Refresh
              </button>
            )}
          </div>

          {/* View mode tabs */}
          <div className="mt-3 flex space-x-1">
            {[
              { mode: 'pending' as ViewMode, label: 'Pending', count: pendingCount },
              { mode: 'applied' as ViewMode, label: 'Applied', count: appliedCount },
              { mode: 'dismissed' as ViewMode, label: 'Dismissed', count: dismissedCount },
              { mode: 'all' as ViewMode, label: 'All', count: suggestions.length },
            ].map(({ mode, label, count }) => (
              <button
                key={mode}
                onClick={() => setViewMode(mode)}
                className={`rounded-full px-3 py-1 text-xs font-medium ${
                  viewMode === mode
                    ? 'bg-blue-100 text-blue-800'
                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                }`}
              >
                {label} ({count})
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Suggestions list */}
      <div className="divide-y divide-gray-100">
        {Object.entries(groupedSuggestions).map(([category, categorySuggestions]) => (
          <div key={category}>
            {groupByCategory && category !== 'all' && (
              <div className="bg-gray-50 px-4 py-2 text-sm font-medium text-gray-700">
                {getCategoryLabel(category as any)}
              </div>
            )}
            <div className="space-y-3 p-4">
              {categorySuggestions.length === 0 ? (
                <p className="text-center text-sm text-gray-500">{emptyMessage}</p>
              ) : (
                categorySuggestions.map((suggestion) => (
                  <SuggestionCard
                    key={suggestion.id}
                    suggestion={suggestion}
                    onApply={onApply ? () => handleApply(suggestion.id) : undefined}
                    onDismiss={onDismiss ? () => handleDismiss(suggestion.id) : undefined}
                    onSelect={() => setSelectedId(selectedId === suggestion.id ? null : suggestion.id)}
                    isSelected={selectedId === suggestion.id}
                    isApplying={applyingId === suggestion.id}
                    isDismissing={dismissingId === suggestion.id}
                    showActions={suggestion.status === 'pending'}
                  />
                ))
              )}
            </div>
          </div>
        ))}

        {filteredSuggestions.length === 0 && (
          <div className="p-8 text-center text-sm text-gray-500">
            {viewMode === 'pending' && pendingCount === 0
              ? 'No pending suggestions'
              : emptyMessage}
          </div>
        )}
      </div>
    </div>
  )
}

export default SuggestionPanel
