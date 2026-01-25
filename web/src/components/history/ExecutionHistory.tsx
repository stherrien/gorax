import { useState, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { useExecutions } from '../../hooks/useExecutions'
import { useWorkflows } from '../../hooks/useWorkflows'
import { ExecutionFilters, defaultExecutionFilters, type ExecutionFilterState } from './ExecutionFilters'
import { Pagination } from '../ui/Pagination'
import type { Execution, ExecutionStatus } from '../../api/executions'

const PAGE_SIZE = 20

const statusColors: Record<ExecutionStatus, string> = {
  completed: 'bg-green-500/20 text-green-400',
  failed: 'bg-red-500/20 text-red-400',
  running: 'bg-blue-500/20 text-blue-400',
  queued: 'bg-yellow-500/20 text-yellow-400',
  cancelled: 'bg-gray-500/20 text-gray-400',
  timeout: 'bg-orange-500/20 text-orange-400',
}

const statusIcons: Record<ExecutionStatus, React.ReactNode> = {
  completed: (
    <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
    </svg>
  ),
  failed: (
    <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
    </svg>
  ),
  running: (
    <svg className="w-3 h-3 animate-spin" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
    </svg>
  ),
  queued: (
    <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
  cancelled: (
    <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
    </svg>
  ),
  timeout: (
    <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
}

export function ExecutionHistory() {
  const [page, setPage] = useState(1)
  const [filters, setFilters] = useState<ExecutionFilterState>(defaultExecutionFilters)

  const { workflows } = useWorkflows()

  const queryParams = useMemo(() => ({
    page,
    limit: PAGE_SIZE,
    ...(filters.status !== 'all' && { status: filters.status }),
    ...(filters.triggerType !== 'all' && { triggerType: filters.triggerType }),
    ...(filters.workflowId !== 'all' && { workflowId: filters.workflowId }),
    ...(filters.startDate && { startDate: filters.startDate }),
    ...(filters.endDate && { endDate: filters.endDate }),
    ...(filters.search && { executionIdPrefix: filters.search }),
  }), [page, filters])

  const { executions, total, loading, error, refetch } = useExecutions(queryParams)

  const totalPages = Math.ceil(total / PAGE_SIZE)

  const handleFiltersChange = (newFilters: ExecutionFilterState) => {
    setFilters(newFilters)
    setPage(1) // Reset to first page when filters change
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString()
  }

  const formatDuration = (ms?: number) => {
    if (!ms) return '-'
    if (ms < 1000) return `${ms}ms`
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
    return `${(ms / 60000).toFixed(1)}m`
  }

  const workflowOptions = workflows.map((w) => ({ id: w.id, name: w.name }))

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Execution History</h1>
          <p className="text-gray-400 text-sm mt-1">
            View and analyze past workflow executions
          </p>
        </div>
        <button
          onClick={() => refetch()}
          disabled={loading}
          className="flex items-center gap-2 px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 transition-colors disabled:opacity-50"
        >
          <svg
            className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          Refresh
        </button>
      </div>

      {/* Filters */}
      <ExecutionFilters
        filters={filters}
        onFiltersChange={handleFiltersChange}
        workflows={workflowOptions}
      />

      {/* Error state */}
      {error && (
        <div className="bg-red-500/10 border border-red-500/30 rounded-lg p-4 text-red-400">
          Failed to load executions: {error.message}
        </div>
      )}

      {/* Table */}
      <div className="bg-gray-800 rounded-lg overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-700">
                <th className="text-left px-4 py-3 text-sm font-medium text-gray-400">
                  Execution ID
                </th>
                <th className="text-left px-4 py-3 text-sm font-medium text-gray-400">
                  Workflow
                </th>
                <th className="text-left px-4 py-3 text-sm font-medium text-gray-400">
                  Status
                </th>
                <th className="text-left px-4 py-3 text-sm font-medium text-gray-400 hidden sm:table-cell">
                  Trigger
                </th>
                <th className="text-left px-4 py-3 text-sm font-medium text-gray-400 hidden md:table-cell">
                  Duration
                </th>
                <th className="text-left px-4 py-3 text-sm font-medium text-gray-400 hidden lg:table-cell">
                  Started
                </th>
                <th className="text-right px-4 py-3 text-sm font-medium text-gray-400">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody>
              {loading && executions.length === 0 ? (
                // Loading skeleton
                [...Array(5)].map((_, i) => (
                  <tr key={i} className="border-b border-gray-700">
                    <td colSpan={7} className="px-4 py-4">
                      <div className="h-6 bg-gray-700 rounded animate-pulse" />
                    </td>
                  </tr>
                ))
              ) : executions.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-4 py-12 text-center text-gray-400">
                    No executions found
                  </td>
                </tr>
              ) : (
                executions.map((execution: Execution) => (
                  <tr
                    key={execution.id}
                    className="border-b border-gray-700 hover:bg-gray-700/50 transition-colors"
                  >
                    <td className="px-4 py-3">
                      <Link
                        to={`/executions/${execution.id}`}
                        className="text-primary-400 hover:text-primary-300 font-mono text-sm"
                      >
                        {execution.id.substring(0, 8)}...
                      </Link>
                    </td>
                    <td className="px-4 py-3">
                      <Link
                        to={`/workflows/${execution.workflowId}`}
                        className="text-white hover:text-primary-400"
                      >
                        {execution.workflowName}
                      </Link>
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-flex items-center gap-1 px-2 py-1 text-xs font-medium rounded-full ${statusColors[execution.status]}`}
                      >
                        {statusIcons[execution.status]}
                        {execution.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-gray-300 hidden sm:table-cell capitalize">
                      {execution.trigger.type}
                    </td>
                    <td className="px-4 py-3 text-gray-300 hidden md:table-cell">
                      {formatDuration(execution.duration)}
                    </td>
                    <td className="px-4 py-3 text-gray-300 text-sm hidden lg:table-cell">
                      {formatDate(execution.startedAt)}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <Link
                        to={`/executions/${execution.id}`}
                        className="px-3 py-1 text-sm text-gray-300 hover:text-white transition-colors"
                      >
                        View
                      </Link>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <Pagination
          currentPage={page}
          totalPages={totalPages}
          onPageChange={setPage}
          totalItems={total}
          pageSize={PAGE_SIZE}
        />
      )}
    </div>
  )
}
