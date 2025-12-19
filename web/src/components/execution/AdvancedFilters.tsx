import { useState, useEffect, useCallback } from 'react'
import DateRangePicker, { DateRange } from '../ui/DateRangePicker'
import type { ExecutionListParams, ExecutionStatus } from '../../api/executions'

export interface AdvancedFiltersProps {
  filters: ExecutionListParams
  onChange: (filters: ExecutionListParams) => void
  onApply?: () => void
  autoApply?: boolean
  className?: string
}

const STATUS_OPTIONS: { value: ExecutionStatus; label: string }[] = [
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
  { value: 'running', label: 'Running' },
  { value: 'cancelled', label: 'Cancelled' },
]

const TRIGGER_OPTIONS: { value: string; label: string }[] = [
  { value: 'manual', label: 'Manual' },
  { value: 'scheduled', label: 'Scheduled' },
  { value: 'webhook', label: 'Webhook' },
  { value: 'webhook_replay', label: 'Webhook Replay' },
]

export default function AdvancedFilters({
  filters,
  onChange,
  onApply,
  autoApply = true,
  className = '',
}: AdvancedFiltersProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const [errorSearchInput, setErrorSearchInput] = useState(filters.errorSearch || '')
  const [executionIdInput, setExecutionIdInput] = useState(filters.executionIdPrefix || '')
  const [minDurationInput, setMinDurationInput] = useState(
    filters.minDurationMs?.toString() || ''
  )
  const [maxDurationInput, setMaxDurationInput] = useState(
    filters.maxDurationMs?.toString() || ''
  )

  const updateFilter = useCallback(
    (updates: Partial<ExecutionListParams>) => {
      onChange({ ...filters, ...updates })
    },
    [filters, onChange]
  )

  useEffect(() => {
    const timer = setTimeout(() => {
      if (errorSearchInput !== (filters.errorSearch || '')) {
        updateFilter({ errorSearch: errorSearchInput || undefined })
      }
    }, 300)

    return () => clearTimeout(timer)
  }, [errorSearchInput, filters.errorSearch, updateFilter])

  useEffect(() => {
    const timer = setTimeout(() => {
      if (executionIdInput !== (filters.executionIdPrefix || '')) {
        updateFilter({ executionIdPrefix: executionIdInput || undefined })
      }
    }, 300)

    return () => clearTimeout(timer)
  }, [executionIdInput, filters.executionIdPrefix, updateFilter])

  useEffect(() => {
    const timer = setTimeout(() => {
      const minDuration = minDurationInput ? parseInt(minDurationInput, 10) : undefined
      const maxDuration = maxDurationInput ? parseInt(maxDurationInput, 10) : undefined

      if (
        minDuration !== filters.minDurationMs ||
        maxDuration !== filters.maxDurationMs
      ) {
        updateFilter({
          minDurationMs: minDuration,
          maxDurationMs: maxDuration,
        })
      }
    }, 300)

    return () => clearTimeout(timer)
  }, [minDurationInput, maxDurationInput, filters.minDurationMs, filters.maxDurationMs, updateFilter])

  const toggleStatus = (status: ExecutionStatus) => {
    const currentStatus = Array.isArray(filters.status) ? filters.status : []
    const newStatus = currentStatus.includes(status)
      ? currentStatus.filter((s) => s !== status)
      : [...currentStatus, status]

    updateFilter({ status: newStatus.length > 0 ? newStatus : undefined })
  }

  const toggleTriggerType = (triggerType: string) => {
    const currentTriggerType = Array.isArray(filters.triggerType)
      ? filters.triggerType
      : []
    const newTriggerType = currentTriggerType.includes(triggerType)
      ? currentTriggerType.filter((t) => t !== triggerType)
      : [...currentTriggerType, triggerType]

    updateFilter({
      triggerType: newTriggerType.length > 0 ? newTriggerType : undefined,
    })
  }

  const handleDateRangeChange = (range: DateRange | null) => {
    if (range) {
      updateFilter({
        startDate: range.startDate.toISOString(),
        endDate: range.endDate.toISOString(),
      })
    } else {
      updateFilter({ startDate: undefined, endDate: undefined })
    }
  }

  const clearAllFilters = () => {
    setErrorSearchInput('')
    setExecutionIdInput('')
    setMinDurationInput('')
    setMaxDurationInput('')
    onChange({})
  }

  const hasActiveFilters = () => {
    return (
      (filters.status && filters.status.length > 0) ||
      (filters.triggerType && filters.triggerType.length > 0) ||
      filters.startDate ||
      filters.endDate ||
      filters.errorSearch ||
      filters.executionIdPrefix ||
      filters.minDurationMs !== undefined ||
      filters.maxDurationMs !== undefined
    )
  }

  const dateRange: DateRange | null =
    filters.startDate && filters.endDate
      ? {
          startDate: new Date(filters.startDate),
          endDate: new Date(filters.endDate),
        }
      : null

  return (
    <div className={`bg-gray-800 rounded-lg ${className}`}>
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full px-4 py-3 flex items-center justify-between text-white hover:bg-gray-700/50 transition-colors rounded-lg"
      >
        <span className="font-medium">Advanced Filters</span>
        <svg
          className={`w-5 h-5 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M19 9l-7 7-7-7"
          />
        </svg>
      </button>

      {isExpanded && (
        <div className="px-4 pb-4 space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Status
              </label>
              <div className="space-y-2">
                {STATUS_OPTIONS.map((option) => (
                  <label key={option.value} className="flex items-center">
                    <input
                      type="checkbox"
                      checked={
                        Array.isArray(filters.status) &&
                        filters.status.includes(option.value)
                      }
                      onChange={() => toggleStatus(option.value)}
                      className="w-4 h-4 text-primary-600 bg-gray-700 border-gray-600 rounded focus:ring-primary-500"
                    />
                    <span className="ml-2 text-sm text-gray-300">
                      {option.label}
                    </span>
                  </label>
                ))}
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Trigger Type
              </label>
              <div className="space-y-2">
                {TRIGGER_OPTIONS.map((option) => (
                  <label key={option.value} className="flex items-center">
                    <input
                      type="checkbox"
                      checked={
                        Array.isArray(filters.triggerType) &&
                        filters.triggerType.includes(option.value)
                      }
                      onChange={() => toggleTriggerType(option.value)}
                      className="w-4 h-4 text-primary-600 bg-gray-700 border-gray-600 rounded focus:ring-primary-500"
                    />
                    <span className="ml-2 text-sm text-gray-300">
                      {option.label}
                    </span>
                  </label>
                ))}
              </div>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Date Range
            </label>
            <DateRangePicker value={dateRange} onChange={handleDateRangeChange} />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Error Message Search
            </label>
            <input
              type="text"
              value={errorSearchInput}
              onChange={(e) => setErrorSearchInput(e.target.value)}
              placeholder="Search error messages..."
              className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg border border-gray-600 focus:outline-none focus:ring-2 focus:ring-primary-500 placeholder-gray-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Execution ID
            </label>
            <input
              type="text"
              value={executionIdInput}
              onChange={(e) => setExecutionIdInput(e.target.value)}
              placeholder="Search by execution ID..."
              className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg border border-gray-600 focus:outline-none focus:ring-2 focus:ring-primary-500 placeholder-gray-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Duration Range (milliseconds)
            </label>
            <div className="grid grid-cols-2 gap-4">
              <input
                type="number"
                value={minDurationInput}
                onChange={(e) => setMinDurationInput(e.target.value)}
                placeholder="Min"
                min="0"
                className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg border border-gray-600 focus:outline-none focus:ring-2 focus:ring-primary-500 placeholder-gray-500"
              />
              <input
                type="number"
                value={maxDurationInput}
                onChange={(e) => setMaxDurationInput(e.target.value)}
                placeholder="Max"
                min="0"
                className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg border border-gray-600 focus:outline-none focus:ring-2 focus:ring-primary-500 placeholder-gray-500"
              />
            </div>
          </div>

          <div className="flex gap-2 pt-2">
            {hasActiveFilters() && (
              <button
                onClick={clearAllFilters}
                className="px-4 py-2 text-sm bg-gray-700 hover:bg-gray-600 text-white rounded-lg transition-colors"
              >
                Clear All
              </button>
            )}
            {!autoApply && onApply && (
              <button
                onClick={onApply}
                className="px-4 py-2 text-sm bg-primary-600 hover:bg-primary-700 text-white rounded-lg transition-colors"
              >
                Apply
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
