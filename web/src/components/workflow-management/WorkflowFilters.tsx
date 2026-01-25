import { useState } from 'react'
import type { WorkflowStatus } from '../../api/workflows'

export interface WorkflowFilterState {
  search: string
  status: WorkflowStatus | 'all'
  sortBy: 'name' | 'updatedAt' | 'createdAt' | 'status'
  sortDirection: 'asc' | 'desc'
}

interface WorkflowFiltersProps {
  filters: WorkflowFilterState
  onFiltersChange: (filters: WorkflowFilterState) => void
  onViewModeChange?: (mode: 'grid' | 'list') => void
  viewMode?: 'grid' | 'list'
}

const statusOptions: Array<{ value: WorkflowStatus | 'all'; label: string }> = [
  { value: 'all', label: 'All Statuses' },
  { value: 'active', label: 'Active' },
  { value: 'draft', label: 'Draft' },
  { value: 'inactive', label: 'Inactive' },
]

const sortOptions: Array<{ value: WorkflowFilterState['sortBy']; label: string }> = [
  { value: 'updatedAt', label: 'Last Updated' },
  { value: 'createdAt', label: 'Created Date' },
  { value: 'name', label: 'Name' },
  { value: 'status', label: 'Status' },
]

export function WorkflowFilters({
  filters,
  onFiltersChange,
  onViewModeChange,
  viewMode = 'list',
}: WorkflowFiltersProps) {
  const [isExpanded, setIsExpanded] = useState(false)

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onFiltersChange({ ...filters, search: e.target.value })
  }

  const handleStatusChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    onFiltersChange({ ...filters, status: e.target.value as WorkflowStatus | 'all' })
  }

  const handleSortChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    onFiltersChange({ ...filters, sortBy: e.target.value as WorkflowFilterState['sortBy'] })
  }

  const handleSortDirectionToggle = () => {
    onFiltersChange({
      ...filters,
      sortDirection: filters.sortDirection === 'asc' ? 'desc' : 'asc',
    })
  }

  const handleReset = () => {
    onFiltersChange({
      search: '',
      status: 'all',
      sortBy: 'updatedAt',
      sortDirection: 'desc',
    })
  }

  const hasActiveFilters = filters.search !== '' || filters.status !== 'all'

  return (
    <div className="bg-gray-800 rounded-lg p-4 mb-4">
      {/* Main filter row */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        {/* Search */}
        <div className="relative flex-1 max-w-md">
          <input
            type="text"
            placeholder="Search workflows..."
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

        {/* Quick filters and view toggle */}
        <div className="flex items-center gap-3">
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

          {/* Sort dropdown */}
          <div className="flex items-center gap-1">
            <select
              value={filters.sortBy}
              onChange={handleSortChange}
              className="px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              {sortOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
            <button
              onClick={handleSortDirectionToggle}
              className="p-2 bg-gray-700 border border-gray-600 rounded-lg text-gray-400 hover:text-white transition-colors"
              title={filters.sortDirection === 'asc' ? 'Ascending' : 'Descending'}
            >
              {filters.sortDirection === 'asc' ? (
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 15l7-7 7 7" />
                </svg>
              ) : (
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              )}
            </button>
          </div>

          {/* View mode toggle */}
          {onViewModeChange && (
            <div className="flex rounded-lg overflow-hidden border border-gray-600">
              <button
                onClick={() => onViewModeChange('list')}
                className={`p-2 ${
                  viewMode === 'list'
                    ? 'bg-primary-600 text-white'
                    : 'bg-gray-700 text-gray-400 hover:text-white'
                } transition-colors`}
                title="List view"
              >
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 10h16M4 14h16M4 18h16" />
                </svg>
              </button>
              <button
                onClick={() => onViewModeChange('grid')}
                className={`p-2 ${
                  viewMode === 'grid'
                    ? 'bg-primary-600 text-white'
                    : 'bg-gray-700 text-gray-400 hover:text-white'
                } transition-colors`}
                title="Grid view"
              >
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
                </svg>
              </button>
            </div>
          )}

          {/* Advanced filters toggle */}
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="p-2 bg-gray-700 border border-gray-600 rounded-lg text-gray-400 hover:text-white transition-colors"
            title="Advanced filters"
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4" />
            </svg>
          </button>
        </div>
      </div>

      {/* Active filters indicator */}
      {hasActiveFilters && (
        <div className="mt-3 flex items-center gap-2 flex-wrap">
          <span className="text-xs text-gray-400">Active filters:</span>
          {filters.search && (
            <span className="px-2 py-1 bg-gray-700 rounded text-xs text-white flex items-center gap-1">
              Search: &quot;{filters.search}&quot;
              <button
                onClick={() => onFiltersChange({ ...filters, search: '' })}
                className="text-gray-400 hover:text-white"
              >
                <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </span>
          )}
          {filters.status !== 'all' && (
            <span className="px-2 py-1 bg-gray-700 rounded text-xs text-white flex items-center gap-1">
              Status: {filters.status}
              <button
                onClick={() => onFiltersChange({ ...filters, status: 'all' })}
                className="text-gray-400 hover:text-white"
              >
                <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </span>
          )}
          <button
            onClick={handleReset}
            className="px-2 py-1 text-xs text-primary-400 hover:text-primary-300"
          >
            Clear all
          </button>
        </div>
      )}

      {/* Advanced filters panel */}
      {isExpanded && (
        <div className="mt-4 pt-4 border-t border-gray-700">
          <p className="text-sm text-gray-400 mb-2">
            Advanced filtering options coming soon...
          </p>
        </div>
      )}
    </div>
  )
}

export const defaultFilters: WorkflowFilterState = {
  search: '',
  status: 'all',
  sortBy: 'updatedAt',
  sortDirection: 'desc',
}
