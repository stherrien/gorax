import type { Tenant } from '../../types/management'

interface TenantTableProps {
  tenants: Tenant[]
  loading: boolean
  onEdit: (tenant: Tenant) => void
  onDelete: (tenant: Tenant) => void
  onSuspend: (tenantId: string) => void
  onReactivate: (tenantId: string) => void
  suspending: boolean
  reactivating: boolean
}

const statusColors: Record<string, string> = {
  active: 'bg-green-500/20 text-green-400',
  trial: 'bg-blue-500/20 text-blue-400',
  suspended: 'bg-red-500/20 text-red-400',
  cancelled: 'bg-gray-500/20 text-gray-400',
}

const planColors: Record<string, string> = {
  free: 'bg-gray-500/20 text-gray-400',
  starter: 'bg-blue-500/20 text-blue-400',
  professional: 'bg-purple-500/20 text-purple-400',
  enterprise: 'bg-yellow-500/20 text-yellow-400',
}

export function TenantTable({
  tenants,
  loading,
  onEdit,
  onDelete,
  onSuspend,
  onReactivate,
  suspending,
  reactivating,
}: TenantTableProps) {
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    })
  }

  const formatUsage = (usage: number, limit: number) => {
    const percentage = limit > 0 ? (usage / limit) * 100 : 0
    return {
      text: `${usage} / ${limit}`,
      percentage,
      color:
        percentage >= 90
          ? 'bg-red-500'
          : percentage >= 70
          ? 'bg-yellow-500'
          : 'bg-green-500',
    }
  }

  return (
    <div className="bg-gray-800 rounded-lg overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-700">
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-400">Tenant</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-400">Plan</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-400">Status</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-400 hidden lg:table-cell">
                Usage
              </th>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-400 hidden md:table-cell">
                Created
              </th>
              <th className="text-right px-4 py-3 text-sm font-medium text-gray-400">Actions</th>
            </tr>
          </thead>
          <tbody>
            {loading && tenants.length === 0 ? (
              // Loading skeleton
              [...Array(5)].map((_, i) => (
                <tr key={i} className="border-b border-gray-700">
                  <td colSpan={6} className="px-4 py-4">
                    <div className="h-10 bg-gray-700 rounded animate-pulse" />
                  </td>
                </tr>
              ))
            ) : tenants.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-4 py-12 text-center text-gray-400">
                  No tenants found
                </td>
              </tr>
            ) : (
              tenants.map((tenant) => {
                const workflowUsage = formatUsage(
                  tenant.usage.workflowCount,
                  tenant.limits.maxWorkflows
                )
                const userUsage = formatUsage(
                  tenant.usage.userCount,
                  tenant.limits.maxUsers
                )

                return (
                  <tr
                    key={tenant.id}
                    className="border-b border-gray-700 hover:bg-gray-700/50 transition-colors"
                  >
                    <td className="px-4 py-3">
                      <div>
                        <p className="text-white font-medium">{tenant.name}</p>
                        <p className="text-gray-400 text-sm">{tenant.slug}</p>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-flex px-2 py-1 text-xs font-medium rounded-full capitalize ${
                          planColors[tenant.plan] || planColors.free
                        }`}
                      >
                        {tenant.plan}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-flex px-2 py-1 text-xs font-medium rounded-full capitalize ${
                          statusColors[tenant.status] || statusColors.cancelled
                        }`}
                      >
                        {tenant.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 hidden lg:table-cell">
                      <div className="space-y-2">
                        <div>
                          <div className="flex justify-between text-xs text-gray-400 mb-1">
                            <span>Workflows</span>
                            <span>{workflowUsage.text}</span>
                          </div>
                          <div className="h-1.5 bg-gray-600 rounded-full overflow-hidden">
                            <div
                              className={`h-full ${workflowUsage.color} transition-all`}
                              style={{ width: `${Math.min(workflowUsage.percentage, 100)}%` }}
                            />
                          </div>
                        </div>
                        <div>
                          <div className="flex justify-between text-xs text-gray-400 mb-1">
                            <span>Users</span>
                            <span>{userUsage.text}</span>
                          </div>
                          <div className="h-1.5 bg-gray-600 rounded-full overflow-hidden">
                            <div
                              className={`h-full ${userUsage.color} transition-all`}
                              style={{ width: `${Math.min(userUsage.percentage, 100)}%` }}
                            />
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-gray-300 text-sm hidden md:table-cell">
                      {formatDate(tenant.createdAt)}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex justify-end items-center gap-2">
                        {tenant.status === 'active' ? (
                          <button
                            onClick={() => onSuspend(tenant.id)}
                            disabled={suspending}
                            className="px-3 py-1 text-sm text-yellow-400 hover:text-yellow-300 transition-colors disabled:opacity-50"
                          >
                            Suspend
                          </button>
                        ) : tenant.status === 'suspended' ? (
                          <button
                            onClick={() => onReactivate(tenant.id)}
                            disabled={reactivating}
                            className="px-3 py-1 text-sm text-green-400 hover:text-green-300 transition-colors disabled:opacity-50"
                          >
                            Reactivate
                          </button>
                        ) : null}
                        <button
                          onClick={() => onEdit(tenant)}
                          className="px-3 py-1 text-sm text-gray-300 hover:text-white transition-colors"
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => onDelete(tenant)}
                          className="px-3 py-1 text-sm text-red-400 hover:text-red-300 transition-colors"
                        >
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                )
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
