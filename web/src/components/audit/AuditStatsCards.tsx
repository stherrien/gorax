import { AuditStats } from '../../types/audit'
import {
  DocumentTextIcon,
  ExclamationTriangleIcon,
  XCircleIcon,
  UserGroupIcon,
} from '@heroicons/react/24/outline'

interface AuditStatsCardsProps {
  stats: AuditStats
  isLoading?: boolean
}

interface StatCardProps {
  label: string
  value: number | string
  icon: React.ComponentType<any>
  color: string
}

function StatCard({ label, value, icon: Icon, color }: StatCardProps) {
  const colorClasses: Record<string, string> = {
    blue: 'bg-blue-100 text-blue-600',
    yellow: 'bg-yellow-100 text-yellow-600',
    red: 'bg-red-100 text-red-600',
    purple: 'bg-purple-100 text-purple-600',
  }

  return (
    <div className="overflow-hidden rounded-lg bg-white shadow">
      <div className="p-5">
        <div className="flex items-center">
          <div className={`flex-shrink-0 rounded-md p-3 ${colorClasses[color]}`}>
            <Icon className="h-6 w-6" aria-hidden="true" />
          </div>
          <div className="ml-5 w-0 flex-1">
            <dl>
              <dt className="truncate text-sm font-medium text-gray-500">{label}</dt>
              <dd className="mt-1 text-3xl font-semibold tracking-tight text-gray-900">
                {value}
              </dd>
            </dl>
          </div>
        </div>
      </div>
    </div>
  )
}

export function AuditStatsCards({ stats, isLoading = false }: AuditStatsCardsProps) {
  if (isLoading) {
    return (
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
        {[...Array(4)].map((_, i) => (
          <div
            key={i}
            className="h-32 animate-pulse rounded-lg bg-gray-200 shadow"
          />
        ))}
      </div>
    )
  }

  const getUniqueUsers = () => {
    return stats.topUsers?.length || 0
  }

  return (
    <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
      <StatCard
        label="Total Events"
        value={stats.totalEvents.toLocaleString()}
        icon={DocumentTextIcon}
        color="blue"
      />
      <StatCard
        label="Critical Events"
        value={stats.criticalEvents}
        icon={ExclamationTriangleIcon}
        color="yellow"
      />
      <StatCard
        label="Failed Events"
        value={stats.failedEvents}
        icon={XCircleIcon}
        color="red"
      />
      <StatCard
        label="Active Users"
        value={getUniqueUsers()}
        icon={UserGroupIcon}
        color="purple"
      />
    </div>
  )
}
