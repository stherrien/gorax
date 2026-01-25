import type { SystemHealth, ServiceHealth, HealthStatus } from '../../types/management'

interface SystemHealthPanelProps {
  health: SystemHealth | null
  loading: boolean
  error: Error | null
}

const statusConfig: Record<HealthStatus, { color: string; bgColor: string; label: string }> = {
  healthy: {
    color: 'text-green-400',
    bgColor: 'bg-green-500/20',
    label: 'Healthy',
  },
  degraded: {
    color: 'text-yellow-400',
    bgColor: 'bg-yellow-500/20',
    label: 'Degraded',
  },
  unhealthy: {
    color: 'text-red-400',
    bgColor: 'bg-red-500/20',
    label: 'Unhealthy',
  },
}

function ServiceHealthRow({ service }: { service: ServiceHealth }) {
  const config = statusConfig[service.status] || statusConfig.unhealthy

  return (
    <div className="flex items-center justify-between py-2">
      <div className="flex items-center gap-3">
        <div
          className={`w-2 h-2 rounded-full ${
            service.status === 'healthy'
              ? 'bg-green-500'
              : service.status === 'degraded'
              ? 'bg-yellow-500'
              : 'bg-red-500'
          }`}
        />
        <span className="text-gray-300">{service.name}</span>
      </div>
      <div className="flex items-center gap-3">
        <span className="text-gray-500 text-sm">{service.responseTime}ms</span>
        <span className={`text-xs font-medium ${config.color}`}>{config.label}</span>
      </div>
    </div>
  )
}

export function SystemHealthPanel({ health, loading, error }: SystemHealthPanelProps) {
  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400)
    const hours = Math.floor((seconds % 86400) / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)

    if (days > 0) {
      return `${days}d ${hours}h`
    } else if (hours > 0) {
      return `${hours}h ${minutes}m`
    } else {
      return `${minutes}m`
    }
  }

  if (error) {
    return (
      <div className="bg-gray-800 rounded-lg p-4">
        <h2 className="text-lg font-semibold text-white mb-4">System Health</h2>
        <div className="text-center py-8 text-red-400">
          <svg className="w-8 h-8 mx-auto mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <p>Failed to load health status</p>
        </div>
      </div>
    )
  }

  if (loading && !health) {
    return (
      <div className="bg-gray-800 rounded-lg p-4">
        <h2 className="text-lg font-semibold text-white mb-4">System Health</h2>
        <div className="space-y-3">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="h-8 bg-gray-700 rounded animate-pulse" />
          ))}
        </div>
      </div>
    )
  }

  const overallConfig = health
    ? statusConfig[health.overall] || statusConfig.unhealthy
    : statusConfig.healthy

  return (
    <div className="bg-gray-800 rounded-lg">
      <div className="px-4 py-3 border-b border-gray-700">
        <h2 className="text-lg font-semibold text-white">System Health</h2>
      </div>

      <div className="p-4">
        {/* Overall Status */}
        <div className={`rounded-lg p-4 ${overallConfig.bgColor} mb-4`}>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div
                className={`w-3 h-3 rounded-full ${
                  health?.overall === 'healthy'
                    ? 'bg-green-500'
                    : health?.overall === 'degraded'
                    ? 'bg-yellow-500'
                    : 'bg-red-500'
                }`}
              />
              <span className={`font-semibold ${overallConfig.color}`}>
                {overallConfig.label}
              </span>
            </div>
            {health?.uptime !== undefined && (
              <span className="text-gray-400 text-sm">
                Uptime: {formatUptime(health.uptime)}
              </span>
            )}
          </div>
        </div>

        {/* Services List */}
        <div className="divide-y divide-gray-700">
          {health?.services?.map((service) => (
            <ServiceHealthRow key={service.name} service={service} />
          ))}

          {(!health?.services || health.services.length === 0) && (
            <div className="text-center py-4 text-gray-500 text-sm">
              No services to display
            </div>
          )}
        </div>

        {/* Last Updated */}
        {health?.lastUpdated && (
          <p className="text-xs text-gray-500 mt-4 text-right">
            Last updated: {new Date(health.lastUpdated).toLocaleTimeString()}
          </p>
        )}
      </div>
    </div>
  )
}
