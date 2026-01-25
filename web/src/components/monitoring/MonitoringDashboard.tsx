import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useMonitoringStats, useActiveExecutions, useSystemHealth } from '../../hooks/useMonitoring'
import { ExecutionMetrics } from './ExecutionMetrics'
import { ActiveExecutions } from './ActiveExecutions'
import { SystemHealthPanel } from './SystemHealthPanel'

interface MonitoringDashboardProps {
  autoRefreshInterval?: number
}

export function MonitoringDashboard({ autoRefreshInterval = 10000 }: MonitoringDashboardProps) {
  const [autoRefresh, setAutoRefresh] = useState(true)
  const refreshInterval = autoRefresh ? autoRefreshInterval : undefined

  const {
    stats,
    loading: statsLoading,
    error: statsError,
    refetch: refetchStats,
  } = useMonitoringStats({ refetchInterval: refreshInterval })

  const {
    activeExecutions,
    loading: executionsLoading,
    error: executionsError,
    refetch: refetchExecutions,
  } = useActiveExecutions({ refetchInterval: refreshInterval })

  const {
    health,
    loading: healthLoading,
    error: healthError,
    refetch: refetchHealth,
  } = useSystemHealth({ refetchInterval: refreshInterval ? refreshInterval * 6 : undefined })

  const handleRefresh = () => {
    refetchStats()
    refetchExecutions()
    refetchHealth()
  }

  const hasError = statsError || executionsError || healthError
  const isLoading = statsLoading && executionsLoading && healthLoading

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Monitoring</h1>
          <p className="text-gray-400 text-sm mt-1">
            Real-time execution monitoring and system health
          </p>
        </div>
        <div className="flex items-center gap-3">
          <label className="flex items-center gap-2 text-sm text-gray-300">
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
              className="w-4 h-4 bg-gray-700 border-gray-600 rounded text-primary-600 focus:ring-primary-500 focus:ring-2"
            />
            Auto-refresh
          </label>
          <button
            onClick={handleRefresh}
            disabled={isLoading}
            className="flex items-center gap-2 px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 transition-colors disabled:opacity-50"
          >
            <svg
              className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
            Refresh
          </button>
          <Link
            to="/executions"
            className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors"
          >
            View All Executions
          </Link>
        </div>
      </div>

      {/* Error Banner */}
      {hasError && (
        <div className="bg-red-500/10 border border-red-500/30 rounded-lg p-4">
          <div className="flex items-center gap-2 text-red-400">
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span className="font-medium">Failed to load some monitoring data</span>
          </div>
          <p className="text-red-400/70 text-sm mt-1">
            {statsError?.message || executionsError?.message || healthError?.message}
          </p>
        </div>
      )}

      {/* Metrics Cards */}
      <ExecutionMetrics stats={stats} loading={statsLoading} />

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Active Executions - Takes 2 columns */}
        <div className="lg:col-span-2">
          <ActiveExecutions
            executions={activeExecutions}
            loading={executionsLoading}
            error={executionsError}
          />
        </div>

        {/* System Health - Takes 1 column */}
        <div>
          <SystemHealthPanel
            health={health}
            loading={healthLoading}
            error={healthError}
          />
        </div>
      </div>
    </div>
  )
}
