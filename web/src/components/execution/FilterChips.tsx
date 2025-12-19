import { format } from 'date-fns'

export interface FilterChipsProps {
  filters: {
    status?: string[]
    workflowId?: string
    workflowName?: string
    triggerType?: string[]
    startDate?: Date
    endDate?: Date
    errorSearch?: string
    executionIdPrefix?: string
    minDurationMs?: number
    maxDurationMs?: number
  }
  onRemove: (filterKey: string) => void
  resultCount?: number
  className?: string
}

interface FilterChip {
  key: string
  label: string
  value: string
}

function formatDuration(ms: number): string {
  if (ms < 1000) {
    return `${ms}ms`
  }
  return `${(ms / 1000).toFixed(1)}s`
}

function getChips(filters: FilterChipsProps['filters']): FilterChip[] {
  const chips: FilterChip[] = []

  if (filters.status && filters.status.length > 0) {
    chips.push({
      key: 'status',
      label: 'Status',
      value: filters.status.join(', '),
    })
  }

  if (filters.workflowId && filters.workflowName) {
    chips.push({
      key: 'workflowId',
      label: 'Workflow',
      value: filters.workflowName,
    })
  }

  if (filters.triggerType && filters.triggerType.length > 0) {
    chips.push({
      key: 'triggerType',
      label: 'Trigger',
      value: filters.triggerType.join(', '),
    })
  }

  if (filters.startDate && filters.endDate) {
    chips.push({
      key: 'dateRange',
      label: 'Date Range',
      value: `${format(filters.startDate, 'MMM d')} - ${format(filters.endDate, 'MMM d')}`,
    })
  }

  if (filters.errorSearch) {
    chips.push({
      key: 'errorSearch',
      label: 'Error',
      value: filters.errorSearch,
    })
  }

  if (filters.executionIdPrefix) {
    chips.push({
      key: 'executionIdPrefix',
      label: 'ID',
      value: filters.executionIdPrefix,
    })
  }

  if (filters.minDurationMs !== undefined || filters.maxDurationMs !== undefined) {
    let value: string
    if (filters.minDurationMs !== undefined && filters.maxDurationMs !== undefined) {
      value = `${formatDuration(filters.minDurationMs)} - ${formatDuration(filters.maxDurationMs)}`
    } else if (filters.minDurationMs !== undefined) {
      value = `>= ${formatDuration(filters.minDurationMs)}`
    } else {
      value = `<= ${formatDuration(filters.maxDurationMs!)}`
    }

    chips.push({
      key: 'duration',
      label: 'Duration',
      value,
    })
  }

  return chips
}

export default function FilterChips({
  filters,
  onRemove,
  resultCount,
  className = '',
}: FilterChipsProps) {
  const chips = getChips(filters)

  if (chips.length === 0) {
    return null
  }

  return (
    <div className={`flex flex-wrap items-center gap-2 ${className}`}>
      {chips.map((chip) => (
        <button
          key={chip.key}
          onClick={() => onRemove(chip.key)}
          className="inline-flex items-center gap-2 px-3 py-1.5 bg-primary-500/20 text-primary-300 rounded-full text-sm hover:bg-primary-500/30 transition-colors"
        >
          <span>
            <span className="font-medium">{chip.label}:</span> {chip.value}
          </span>
          <svg
            className="w-4 h-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      ))}

      {resultCount !== undefined && (
        <span className="text-sm text-gray-400 ml-2">
          {resultCount} {resultCount === 1 ? 'result' : 'results'}
        </span>
      )}
    </div>
  )
}
