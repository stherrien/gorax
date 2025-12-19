import { Link } from 'react-router-dom'
import { useDashboardStats, useRecentExecutions } from '../hooks/useExecutions'
import type { Execution, ExecutionStatus } from '../api/executions'

export default function Dashboard() {
  const { stats, loading: statsLoading, error: statsError } = useDashboardStats()
  const {
    executions: recentExecutions,
    loading: executionsLoading,
    error: executionsError,
  } = useRecentExecutions(5)

  const loading = statsLoading || executionsLoading
  const error = statsError || executionsError

  // Loading state
  if (loading) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-white text-lg">Loading dashboard...</div>
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-400 text-lg mb-4">Failed to load dashboard</div>
          <div className="text-gray-400 text-sm">{error.message}</div>
        </div>
      </div>
    )
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-white mb-6">Dashboard</h1>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <StatCard
          title="Active Workflows"
          value={stats?.activeWorkflows.toString() || '0'}
          change={`${stats?.activeWorkflows || 0} total`}
          positive
        />
        <StatCard
          title="Executions Today"
          value={stats?.executionsToday.toString() || '0'}
          change={`${stats?.successRateToday.toFixed(1)}% success rate`}
          positive={stats ? stats.successRateToday > 95 : true}
        />
        <StatCard
          title="Failed Executions"
          value={stats?.failedToday.toString() || '0'}
          change={`${stats?.totalExecutions || 0} total executions`}
          positive={stats ? stats.failedToday === 0 : true}
        />
      </div>

      {/* Quick Actions */}
      <div className="bg-gray-800 rounded-lg p-6 mb-8">
        <h2 className="text-lg font-semibold text-white mb-4">Quick Actions</h2>
        <div className="flex space-x-4">
          <Link
            to="/workflows/new"
            className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors"
          >
            Create Workflow
          </Link>
          <Link
            to="/workflows"
            className="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 transition-colors"
          >
            View All Workflows
          </Link>
          <Link
            to="/executions"
            className="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 transition-colors"
          >
            View Executions
          </Link>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="bg-gray-800 rounded-lg p-6">
        <h2 className="text-lg font-semibold text-white mb-4">Recent Executions</h2>

        {recentExecutions.length === 0 ? (
          <div className="text-center py-8 text-gray-400">
            No recent executions found
          </div>
        ) : (
          <div className="space-y-3">
            {recentExecutions.map((execution) => (
              <ExecutionItem key={execution.id} execution={execution} />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

interface StatCardProps {
  title: string
  value: string
  change: string
  positive: boolean
}

function StatCard({ title, value, change, positive }: StatCardProps) {
  return (
    <div className="bg-gray-800 rounded-lg p-6">
      <p className="text-gray-400 text-sm">{title}</p>
      <p className="text-3xl font-bold text-white mt-2">{value}</p>
      <p className={`text-sm mt-2 ${positive ? 'text-green-400' : 'text-red-400'}`}>
        {change}
      </p>
    </div>
  )
}

interface ExecutionItemProps {
  execution: Execution
}

function ExecutionItem({ execution }: ExecutionItemProps) {
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
      return `Triggered by webhook${trigger.source ? ` (${trigger.source})` : ''}`
    } else if (trigger.type === 'schedule') {
      return 'Triggered by schedule'
    } else {
      return 'Triggered manually'
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

  return (
    <div className="flex items-center justify-between p-3 bg-gray-700 rounded-lg">
      <div>
        <p className="text-white font-medium">{execution.workflowName}</p>
        <p className="text-gray-400 text-sm">{getTriggerLabel(execution.trigger)}</p>
      </div>
      <div className="text-right">
        <span className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(execution.status)}`}>
          {getStatusLabel(execution.status)}
        </span>
        <p className="text-gray-400 text-sm mt-1">{getTimeAgo(execution.startedAt)}</p>
      </div>
    </div>
  )
}
