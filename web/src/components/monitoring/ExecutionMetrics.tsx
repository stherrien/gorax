import type { MonitoringStats } from '../../types/management'

interface ExecutionMetricsProps {
  stats: MonitoringStats | null
  loading: boolean
}

interface MetricCardProps {
  title: string
  value: string | number
  subtitle?: string
  trend?: 'up' | 'down' | 'neutral'
  trendValue?: string
  icon: React.ReactNode
  color: 'green' | 'blue' | 'yellow' | 'red' | 'purple'
  loading?: boolean
}

const colorClasses = {
  green: 'text-green-400 bg-green-500/10',
  blue: 'text-blue-400 bg-blue-500/10',
  yellow: 'text-yellow-400 bg-yellow-500/10',
  red: 'text-red-400 bg-red-500/10',
  purple: 'text-purple-400 bg-purple-500/10',
}

function MetricCard({
  title,
  value,
  subtitle,
  trend,
  trendValue,
  icon,
  color,
  loading,
}: MetricCardProps) {
  return (
    <div className="bg-gray-800 rounded-lg p-4">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <p className="text-gray-400 text-sm">{title}</p>
          {loading ? (
            <div className="h-8 w-20 bg-gray-700 animate-pulse rounded mt-1" />
          ) : (
            <p className="text-2xl font-bold text-white mt-1">{value}</p>
          )}
          {subtitle && (
            <p className="text-xs text-gray-500 mt-1">{subtitle}</p>
          )}
          {trend && trendValue && (
            <div className="flex items-center gap-1 mt-2">
              {trend === 'up' && (
                <svg className="w-4 h-4 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                </svg>
              )}
              {trend === 'down' && (
                <svg className="w-4 h-4 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 17h8m0 0V9m0 8l-8-8-4 4-6-6" />
                </svg>
              )}
              <span
                className={`text-xs ${
                  trend === 'up' ? 'text-green-400' : trend === 'down' ? 'text-red-400' : 'text-gray-400'
                }`}
              >
                {trendValue}
              </span>
            </div>
          )}
        </div>
        <div className={`p-2 rounded-lg ${colorClasses[color]}`}>
          {icon}
        </div>
      </div>
    </div>
  )
}

export function ExecutionMetrics({ stats, loading }: ExecutionMetricsProps) {
  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
    return `${(ms / 60000).toFixed(1)}m`
  }

  const formatRate = (rate: number) => {
    return `${rate.toFixed(1)}%`
  }

  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
      <MetricCard
        title="Active Executions"
        value={stats?.activeExecutions ?? 0}
        subtitle="Currently running"
        icon={
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M13 10V3L4 14h7v7l9-11h-7z"
            />
          </svg>
        }
        color="blue"
        loading={loading}
      />

      <MetricCard
        title="Queued"
        value={stats?.queuedExecutions ?? 0}
        subtitle="Waiting to start"
        icon={
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        }
        color="yellow"
        loading={loading}
      />

      <MetricCard
        title="Exec/min"
        value={stats?.executionsPerMinute?.toFixed(1) ?? '0'}
        subtitle="Throughput"
        icon={
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
            />
          </svg>
        }
        color="purple"
        loading={loading}
      />

      <MetricCard
        title="Avg Duration"
        value={formatDuration(stats?.averageExecutionTime ?? 0)}
        subtitle="Per execution"
        icon={
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        }
        color="blue"
        loading={loading}
      />

      <MetricCard
        title="Success Rate"
        value={formatRate(stats?.successRate ?? 0)}
        subtitle="Last hour"
        trend={(stats?.successRate ?? 0) >= 95 ? 'up' : 'down'}
        trendValue={(stats?.successRate ?? 0) >= 95 ? 'Healthy' : 'Below target'}
        icon={
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        }
        color="green"
        loading={loading}
      />

      <MetricCard
        title="Error Rate"
        value={formatRate(stats?.errorRate ?? 0)}
        subtitle="Last hour"
        trend={(stats?.errorRate ?? 0) <= 5 ? 'down' : 'up'}
        trendValue={(stats?.errorRate ?? 0) <= 5 ? 'Normal' : 'Elevated'}
        icon={
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        }
        color={(stats?.errorRate ?? 0) <= 5 ? 'green' : 'red'}
        loading={loading}
      />
    </div>
  )
}
