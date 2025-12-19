import { useState, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { useExecutions } from '../hooks/useExecutions'
import { useBulkSelection } from '../hooks/useBulkSelection'
import { SelectableTable } from '../components/common/SelectableTable'
import { BulkActionBar } from '../components/common/BulkActionBar'
import { BulkExecutionActions } from '../components/execution/BulkExecutionActions'
import type { Execution, ExecutionStatus } from '../api/executions'

export default function Executions() {
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState<ExecutionStatus | ''>('')
  const [workflowId, setWorkflowId] = useState('')
  const [search, setSearch] = useState('')
  const [toast, setToast] = useState<{ type: 'success' | 'error'; message: string } | null>(null)

  const params = useMemo(() => {
    const p: any = { page, limit: 20 }
    if (status) p.status = status
    if (workflowId) p.workflowId = workflowId
    if (search) p.search = search
    return p
  }, [page, status, workflowId, search])

  const { executions, total, loading, error, refetch } = useExecutions(params)
  const bulkSelection = useBulkSelection<Execution>()

  const totalPages = Math.ceil(total / 20)

  const handleBulkSuccess = () => {
    setToast({ type: 'success', message: 'Bulk operation completed successfully' })
    bulkSelection.clear()
    refetch()
    setTimeout(() => setToast(null), 3000)
  }

  const handleBulkError = (message: string) => {
    setToast({ type: 'error', message })
    setTimeout(() => setToast(null), 5000)
  }

  // Loading state
  if (loading) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-white text-lg">Loading executions...</div>
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-400 text-lg mb-4">Failed to load executions</div>
          <div className="text-gray-400 text-sm">{error.message}</div>
        </div>
      </div>
    )
  }

  return (
    <div>
      {toast && (
        <div
          className={`fixed top-4 right-4 px-4 py-3 rounded-lg shadow-lg z-50 ${
            toast.type === 'success' ? 'bg-green-600' : 'bg-red-600'
          } text-white`}
        >
          {toast.message}
        </div>
      )}

      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-white">Executions</h1>
        <div className="text-gray-400 text-sm">
          {total} total execution{total !== 1 ? 's' : ''}
        </div>
      </div>

      {bulkSelection.count > 0 && (
        <BulkActionBar count={bulkSelection.count} onClear={bulkSelection.clear}>
          <BulkExecutionActions
            selectedIds={Array.from(bulkSelection.selected)}
            onSuccess={handleBulkSuccess}
            onError={handleBulkError}
          />
        </BulkActionBar>
      )}

      {/* Filters */}
      <div className="bg-gray-800 rounded-lg p-4 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label htmlFor="status-filter" className="block text-sm text-gray-400 mb-2">
              Status
            </label>
            <select
              id="status-filter"
              value={status}
              onChange={(e) => {
                setStatus(e.target.value as ExecutionStatus | '')
                setPage(1)
              }}
              className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="">All Statuses</option>
              <option value="completed">Completed</option>
              <option value="failed">Failed</option>
              <option value="running">Running</option>
              <option value="queued">Queued</option>
              <option value="cancelled">Cancelled</option>
              <option value="timeout">Timeout</option>
            </select>
          </div>

          <div>
            <label htmlFor="workflow-filter" className="block text-sm text-gray-400 mb-2">
              Workflow
            </label>
            <select
              id="workflow-filter"
              value={workflowId}
              onChange={(e) => {
                setWorkflowId(e.target.value)
                setPage(1)
              }}
              className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="">All Workflows</option>
            </select>
          </div>

          <div>
            <label htmlFor="search-input" className="block text-sm text-gray-400 mb-2">
              Search
            </label>
            <input
              id="search-input"
              type="text"
              value={search}
              onChange={(e) => {
                setSearch(e.target.value)
                setPage(1)
              }}
              placeholder="Search workflows..."
              className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 placeholder-gray-500"
            />
          </div>
        </div>
      </div>

      {/* Executions List */}
      <div className="bg-gray-800 rounded-lg overflow-hidden">
        {executions.length === 0 ? (
          <div className="text-center py-12 text-gray-400">No executions found</div>
        ) : (
          <>
            <SelectableTable
              items={executions}
              selectedIds={bulkSelection.selected}
              onSelectionChange={bulkSelection.toggle}
              onSelectAll={() => bulkSelection.toggleAll(executions)}
              onRangeSelect={(startId, endId) => bulkSelection.selectRange(startId, endId, executions)}
              columns={[
                {
                  header: 'Workflow',
                  accessor: (execution) => (
                    <Link
                      to={`/executions/${execution.id}`}
                      className="text-white hover:text-primary-400 font-medium"
                    >
                      {execution.workflowName}
                    </Link>
                  ),
                },
                {
                  header: 'Status',
                  accessor: (execution) => (
                    <span
                      className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(execution.status)}`}
                    >
                      {getStatusLabel(execution.status)}
                    </span>
                  ),
                },
                {
                  header: 'Trigger',
                  accessor: (execution) => (
                    <span className="text-gray-400">{getTriggerLabel(execution.trigger)}</span>
                  ),
                },
                {
                  header: 'Progress',
                  accessor: (execution) => (
                    <span className="text-gray-400">
                      {execution.completedSteps}/{execution.stepCount}
                    </span>
                  ),
                },
                {
                  header: 'Started',
                  accessor: (execution) => (
                    <span className="text-gray-400">{getTimeAgo(execution.startedAt)}</span>
                  ),
                },
              ]}
            />

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="px-4 py-3 bg-gray-700 flex items-center justify-between">
                <div className="text-sm text-gray-400">
                  Page {page} of {totalPages}
                </div>
                <div className="flex space-x-2">
                  <button
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                    disabled={page === 1}
                    className="px-3 py-1 bg-gray-600 text-white rounded text-sm hover:bg-gray-500 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Previous
                  </button>
                  <button
                    onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                    disabled={page === totalPages}
                    className="px-3 py-1 bg-gray-600 text-white rounded text-sm hover:bg-gray-500 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Next
                  </button>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}

// Helper functions
const getStatusColor = (status: ExecutionStatus) => {
  switch (status) {
    case 'completed':
      return 'bg-green-500/20 text-green-400'
    case 'failed':
      return 'bg-red-500/20 text-red-400'
    case 'running':
      return 'bg-blue-500/20 text-blue-400'
    case 'queued':
      return 'bg-yellow-500/20 text-yellow-400'
    case 'cancelled':
      return 'bg-gray-500/20 text-gray-400'
    case 'timeout':
      return 'bg-orange-500/20 text-orange-400'
    default:
      return 'bg-gray-500/20 text-gray-400'
  }
}

const getStatusLabel = (status: ExecutionStatus) => {
  return status.charAt(0).toUpperCase() + status.slice(1)
}

const getTriggerLabel = (trigger: Execution['trigger']) => {
  if (trigger.type === 'webhook') {
    return trigger.source ? `webhook (${trigger.source})` : 'webhook'
  } else if (trigger.type === 'schedule') {
    return 'schedule'
  } else {
    return 'manual'
  }
}

const getTimeAgo = (dateString: string) => {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins} min ago`
  if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`
  return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`
}
