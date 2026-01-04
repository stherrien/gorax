import { useState } from 'react'
import {
  QueryFilter,
  AuditCategory,
  AuditEventType,
  AuditSeverity,
  AuditStatus,
  TimeRangePreset,
  getTimeRangeFromPreset,
  CATEGORY_LABELS,
  EVENT_TYPE_LABELS,
  SEVERITY_LABELS,
  STATUS_LABELS,
} from '../../types/audit'

interface AuditFilterPanelProps {
  filter: QueryFilter
  onFilterChange: (filter: QueryFilter) => void
  onApply: () => void
  onReset: () => void
}

export function AuditFilterPanel({
  filter,
  onFilterChange,
  onApply,
  onReset,
}: AuditFilterPanelProps) {
  const [timePreset, setTimePreset] = useState<TimeRangePreset>(
    TimeRangePreset.Last24Hours
  )

  const handleTimePresetChange = (preset: TimeRangePreset) => {
    setTimePreset(preset)
    const range = getTimeRangeFromPreset(preset)
    if (range) {
      onFilterChange({
        ...filter,
        startDate: range.startDate,
        endDate: range.endDate,
      })
    }
  }

  const handleCategoryToggle = (category: AuditCategory) => {
    const categories = filter.categories || []
    const newCategories = categories.includes(category)
      ? categories.filter((c) => c !== category)
      : [...categories, category]

    onFilterChange({ ...filter, categories: newCategories })
  }

  const handleEventTypeToggle = (eventType: AuditEventType) => {
    const eventTypes = filter.eventTypes || []
    const newEventTypes = eventTypes.includes(eventType)
      ? eventTypes.filter((t) => t !== eventType)
      : [...eventTypes, eventType]

    onFilterChange({ ...filter, eventTypes: newEventTypes })
  }

  const handleSeverityToggle = (severity: AuditSeverity) => {
    const severities = filter.severities || []
    const newSeverities = severities.includes(severity)
      ? severities.filter((s) => s !== severity)
      : [...severities, severity]

    onFilterChange({ ...filter, severities: newSeverities })
  }

  const handleStatusToggle = (status: AuditStatus) => {
    const statuses = filter.statuses || []
    const newStatuses = statuses.includes(status)
      ? statuses.filter((s) => s !== status)
      : [...statuses, status]

    onFilterChange({ ...filter, statuses: newStatuses })
  }

  return (
    <div className="w-80 space-y-6 border-r border-gray-200 bg-white p-4">
      <div>
        <h3 className="text-sm font-semibold text-gray-900">Filters</h3>
      </div>

      {/* Time Range */}
      <div>
        <label className="block text-sm font-medium text-gray-700">
          Time Range
        </label>
        <select
          value={timePreset}
          onChange={(e) =>
            handleTimePresetChange(e.target.value as TimeRangePreset)
          }
          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
        >
          <option value={TimeRangePreset.Last24Hours}>Last 24 Hours</option>
          <option value={TimeRangePreset.Last7Days}>Last 7 Days</option>
          <option value={TimeRangePreset.Last30Days}>Last 30 Days</option>
          <option value={TimeRangePreset.Last90Days}>Last 90 Days</option>
        </select>
      </div>

      {/* Category Filter */}
      <div>
        <label className="block text-sm font-medium text-gray-700">
          Categories
        </label>
        <div className="mt-2 space-y-2">
          {Object.values(AuditCategory).map((category) => (
            <label key={category} className="flex items-center">
              <input
                type="checkbox"
                checked={filter.categories?.includes(category) || false}
                onChange={() => handleCategoryToggle(category)}
                className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
              />
              <span className="ml-2 text-sm text-gray-700">
                {CATEGORY_LABELS[category]}
              </span>
            </label>
          ))}
        </div>
      </div>

      {/* Event Type Filter */}
      <div>
        <label className="block text-sm font-medium text-gray-700">
          Event Types
        </label>
        <div className="mt-2 space-y-2">
          {Object.values(AuditEventType).map((eventType) => (
            <label key={eventType} className="flex items-center">
              <input
                type="checkbox"
                checked={filter.eventTypes?.includes(eventType) || false}
                onChange={() => handleEventTypeToggle(eventType)}
                className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
              />
              <span className="ml-2 text-sm text-gray-700">
                {EVENT_TYPE_LABELS[eventType]}
              </span>
            </label>
          ))}
        </div>
      </div>

      {/* Severity Filter */}
      <div>
        <label className="block text-sm font-medium text-gray-700">
          Severities
        </label>
        <div className="mt-2 space-y-2">
          {Object.values(AuditSeverity).map((severity) => (
            <label key={severity} className="flex items-center">
              <input
                type="checkbox"
                checked={filter.severities?.includes(severity) || false}
                onChange={() => handleSeverityToggle(severity)}
                className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
              />
              <span className="ml-2 text-sm text-gray-700">
                {SEVERITY_LABELS[severity]}
              </span>
            </label>
          ))}
        </div>
      </div>

      {/* Status Filter */}
      <div>
        <label className="block text-sm font-medium text-gray-700">Status</label>
        <div className="mt-2 space-y-2">
          {Object.values(AuditStatus).map((status) => (
            <label key={status} className="flex items-center">
              <input
                type="checkbox"
                checked={filter.statuses?.includes(status) || false}
                onChange={() => handleStatusToggle(status)}
                className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
              />
              <span className="ml-2 text-sm text-gray-700">
                {STATUS_LABELS[status]}
              </span>
            </label>
          ))}
        </div>
      </div>

      {/* User Filter */}
      <div>
        <label className="block text-sm font-medium text-gray-700">
          User Email
        </label>
        <input
          type="text"
          value={filter.userEmail || ''}
          onChange={(e) =>
            onFilterChange({ ...filter, userEmail: e.target.value })
          }
          placeholder="Filter by user email"
          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
        />
      </div>

      {/* IP Address Filter */}
      <div>
        <label className="block text-sm font-medium text-gray-700">
          IP Address
        </label>
        <input
          type="text"
          value={filter.ipAddress || ''}
          onChange={(e) =>
            onFilterChange({ ...filter, ipAddress: e.target.value })
          }
          placeholder="Filter by IP address"
          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
        />
      </div>

      {/* Resource Type Filter */}
      <div>
        <label className="block text-sm font-medium text-gray-700">
          Resource Type
        </label>
        <input
          type="text"
          value={filter.resourceType || ''}
          onChange={(e) =>
            onFilterChange({ ...filter, resourceType: e.target.value })
          }
          placeholder="Filter by resource type"
          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
        />
      </div>

      {/* Action Buttons */}
      <div className="space-y-2">
        <button
          onClick={onApply}
          className="w-full rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700"
        >
          Apply Filters
        </button>
        <button
          onClick={onReset}
          className="w-full rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50"
        >
          Reset
        </button>
      </div>
    </div>
  )
}
