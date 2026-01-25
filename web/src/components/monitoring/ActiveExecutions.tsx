import { Link } from 'react-router-dom'

interface ActiveExecution {
  id: string
  workflowId: string
  workflowName: string
  status: string
  startedAt: string
  currentNode?: string
  progress: number
}

interface ActiveExecutionsProps {
  executions: ActiveExecution[]
  loading: boolean
  error: Error | null
}

const statusColors: Record<string, string> = {
  running: 'bg-blue-500',
  queued: 'bg-yellow-500',
  pending: 'bg-gray-500',
}

export function ActiveExecutions({ executions, loading, error }: ActiveExecutionsProps) {
  const getTimeRunning = (startedAt: string) => {
    const start = new Date(startedAt)
    const now = new Date()
    const diffMs = now.getTime() - start.getTime()
    const diffSecs = Math.floor(diffMs / 1000)
    const diffMins = Math.floor(diffSecs / 60)
    const diffHours = Math.floor(diffMins / 60)

    if (diffHours > 0) {
      return `${diffHours}h ${diffMins % 60}m`
    } else if (diffMins > 0) {
      return `${diffMins}m ${diffSecs % 60}s`
    } else {
      return `${diffSecs}s`
    }
  }

  return (
    <div className="bg-gray-800 rounded-lg">
      <div className="px-4 py-3 border-b border-gray-700 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h2 className="text-lg font-semibold text-white">Active Executions</h2>
          {executions.length > 0 && (
            <span className="px-2 py-0.5 bg-blue-500/20 text-blue-400 text-xs font-medium rounded-full">
              {executions.length}
            </span>
          )}
        </div>
        {loading && (
          <div className="flex items-center gap-2 text-gray-400 text-sm">
            <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
            Live
          </div>
        )}
      </div>

      <div className="p-4">
        {error && (
          <div className="text-center py-8 text-red-400">
            <svg className="w-8 h-8 mx-auto mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <p>Failed to load active executions</p>
          </div>
        )}

        {!error && loading && executions.length === 0 && (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="bg-gray-700/50 rounded-lg p-4 animate-pulse">
                <div className="h-4 w-48 bg-gray-600 rounded mb-3" />
                <div className="h-2 w-full bg-gray-600 rounded" />
              </div>
            ))}
          </div>
        )}

        {!error && !loading && executions.length === 0 && (
          <div className="text-center py-12 text-gray-400">
            <svg
              className="w-12 h-12 mx-auto mb-3 text-gray-600"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <p className="text-lg font-medium">No active executions</p>
            <p className="text-sm mt-1">All workflows are idle</p>
          </div>
        )}

        {!error && executions.length > 0 && (
          <div className="space-y-3">
            {executions.map((execution) => (
              <Link
                key={execution.id}
                to={`/executions/${execution.id}`}
                className="block bg-gray-700/50 rounded-lg p-4 hover:bg-gray-700 transition-colors"
              >
                <div className="flex items-start justify-between gap-4 mb-3">
                  <div>
                    <p className="text-white font-medium">{execution.workflowName}</p>
                    <p className="text-gray-400 text-sm">
                      {execution.currentNode
                        ? `Running: ${execution.currentNode}`
                        : 'Starting...'}
                    </p>
                  </div>
                  <div className="text-right">
                    <span
                      className={`inline-flex items-center gap-1 px-2 py-0.5 text-xs font-medium rounded-full ${
                        execution.status === 'running'
                          ? 'bg-blue-500/20 text-blue-400'
                          : 'bg-yellow-500/20 text-yellow-400'
                      }`}
                    >
                      <span className="w-1.5 h-1.5 rounded-full bg-current animate-pulse" />
                      {execution.status}
                    </span>
                    <p className="text-gray-500 text-xs mt-1">
                      {getTimeRunning(execution.startedAt)}
                    </p>
                  </div>
                </div>

                {/* Progress bar */}
                <div className="relative">
                  <div className="h-2 bg-gray-600 rounded-full overflow-hidden">
                    <div
                      className={`h-full ${statusColors[execution.status] || 'bg-blue-500'} transition-all duration-300`}
                      style={{ width: `${execution.progress}%` }}
                    />
                  </div>
                  <p className="text-xs text-gray-500 mt-1 text-right">
                    {execution.progress}% complete
                  </p>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
