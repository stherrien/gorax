import { useState } from 'react'
import type { ExecutionStatus } from '../../api/executions'

export interface ExecutionFilterState {
  search: string
  status: ExecutionStatus | 'all'
  triggerType: string | 'all'
  startDate: string
  endDate: string
  workflowId: string | 'all'
}

interface ExecutionFiltersProps {
  filters: ExecutionFilterState
  onFiltersChange: (filters: ExecutionFilterState) => void
  workflows?: Array<{ id: string; name: string }>
}

const statusOptions: Array<{ value: ExecutionStatus | 'all'; label: string }> = [
  { value: 'all', label: 'All Statuses' },
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
  { value: 'running', label: 'Running' },
  { value: 'queued', label: 'Queued' },
  { value: 'cancelled', label: 'Cancelled' },
  { value: 'timeout', label: 'Timeout' },
]

const triggerOptions: Array<{ value: string; label: string }> = [
  { value: 'all', label: 'All Triggers' },
  { value: 'webhook', label: 'Webhook' },
  { value: 'schedule', label: 'Schedule' },
  { value: 'manual', label: 'Manual' },
]

export function ExecutionFilters({
  filters,
  onFiltersChange,
  workflows = [],
}: ExecutionFiltersProps) {
  const [isExpanded, setIsExpanded] = useState(false)

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onFiltersChange({ ...filters, search: e.target.value })
  }

  const handleStatusChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    onFiltersChange({ ...filters, status: e.target.value as ExecutionStatus | 'all' })
  }

  const handleTriggerChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    onFiltersChange({ ...filters, triggerType: e.target.value })
  }

  const handleWorkflowChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    onFiltersChange({ ...filters, workflowId: e.target.value })
  }

  const handleDateChange = (field: 'startDate' | 'endDate', value: string) => {
    onFiltersChange({ ...filters, [field]: value })
  }

  const handleReset = () => {
    onFiltersChange({
      search: '',
      status: 'all',
      triggerType: 'all',
      startDate: '',
      endDate: '',
      workflowId: 'all',
    })
  }

  const hasActiveFilters =
    filters.search !== '' ||
    filters.status !== 'all' ||
    filters.triggerType !== 'all' ||
    filters.startDate !== '' ||
    filters.endDate !== '' ||
    filters.workflowId !== 'all'

  return (
    <div className="bg-gray-800 rounded-lg p-4 mb-4">
      {/* Main filter row */}
      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        {/* Search */}
        <div className="relative flex-1 max-w-md">
          <input
            type="text"
            placeholder="Search executions..."
            value={filters.search}
            onChange={handleSearchChange}
            className="w-full px-4 py-2 pl-10 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          />
          <svg
            className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
          </svg>
        </div>

        {/* Quick filters */}
        <div className="flex flex-wrap items-center gap-3">
          {/* Status filter */}
          <select
            value={filters.status}
            onChange={handleStatusChange}
            className="px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            {statusOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>

          {/* Trigger type filter */}
          <select
            value={filters.triggerType}
            onChange={handleTriggerChange}
            className="px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            {triggerOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>

          {/* Workflow filter */}
          {workflows.length > 0 && (
            <select
              value={filters.workflowId}
              onChange={handleWorkflowChange}
              className="px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 max-w-[200px]"
            >
              <option value="all">All Workflows</option>
              {workflows.map((wf) => (
                <option key={wf.id} value={wf.id}>
                  {wf.name}
                </option>
              ))}
            </select>
          )}

          {/* Advanced filters toggle */}
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className={`p-2 border rounded-lg transition-colors ${
              isExpanded
                ? 'bg-primary-600 border-primary-500 text-white'
                : 'bg-gray-700 border-gray-600 text-gray-400 hover:text-white'
            }`}
            title="Advanced filters"
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
              />
            </svg>
          </button>
        </div>
      </div>

      {/* Advanced filters panel */}
      {isExpanded && (
        <div className="mt-4 pt-4 border-t border-gray-700">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Date range */}
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">
                Start Date
              </label>
              <input
                type="date"
                value={filters.startDate}
                onChange={(e) => handleDateChange('startDate', e.target.value)}
                className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">
                End Date
              </label>
              <input
                type="date"
                value={filters.endDate}
                onChange={(e) => handleDateChange('endDate', e.target.value)}
                className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              />
            </div>
          </div>
        </div>
      )}

      {/* Active filters indicator */}
      {hasActiveFilters && (
        <div className="mt-3 flex items-center gap-2 flex-wrap">
          <span className="text-xs text-gray-400">Active filters:</span>
          {filters.search && (
            <FilterTag
              label={`Search: "${filters.search}"`}
              onRemove={() => onFiltersChange({ ...filters, search: '' })}
            />
          )}
          {filters.status !== 'all' && (
            <FilterTag
              label={`Status: ${filters.status}`}
              onRemove={() => onFiltersChange({ ...filters, status: 'all' })}
            />
          )}
          {filters.triggerType !== 'all' && (
            <FilterTag
              label={`Trigger: ${filters.triggerType}`}
              onRemove={() => onFiltersChange({ ...filters, triggerType: 'all' })}
            />
          )}
          {filters.workflowId !== 'all' && (
            <FilterTag
              label="Workflow filter active"
              onRemove={() => onFiltersChange({ ...filters, workflowId: 'all' })}
            />
          )}
          {(filters.startDate || filters.endDate) && (
            <FilterTag
              label="Date range active"
              onRemove={() => onFiltersChange({ ...filters, startDate: '', endDate: '' })}
            />
          )}
          <button
            onClick={handleReset}
            className="px-2 py-1 text-xs text-primary-400 hover:text-primary-300"
          >
            Clear all
          </button>
        </div>
      )}
    </div>
  )
}

function FilterTag({ label, onRemove }: { label: string; onRemove: () => void }) {
  return (
    <span className="px-2 py-1 bg-gray-700 rounded text-xs text-white flex items-center gap-1">
      {label}
      <button onClick={onRemove} className="text-gray-400 hover:text-white">
        <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M6 18L18 6M6 6l12 12"
          />
        </svg>
      </button>
    </span>
  )
}

export const defaultExecutionFilters: ExecutionFilterState = {
  search: '',
  status: 'all',
  triggerType: 'all',
  startDate: '',
  endDate: '',
  workflowId: 'all',
}
