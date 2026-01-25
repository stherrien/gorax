import { useState } from 'react'
import { useTenants, useTenantMutations } from '../../hooks/useTenants'
import { TenantTable } from './TenantTable'
import { TenantForm } from './TenantForm'
import type { Tenant, TenantCreateInput, TenantUpdateInput, TenantStatus, TenantPlan } from '../../types/management'

export function TenantManagement() {
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [editingTenant, setEditingTenant] = useState<Tenant | null>(null)
  const [deleteConfirm, setDeleteConfirm] = useState<Tenant | null>(null)
  const [statusFilter, setStatusFilter] = useState<TenantStatus | 'all'>('all')
  const [planFilter, setPlanFilter] = useState<TenantPlan | 'all'>('all')
  const [searchQuery, setSearchQuery] = useState('')

  const { tenants, total, loading, error, refetch } = useTenants({
    ...(statusFilter !== 'all' && { status: statusFilter }),
    ...(planFilter !== 'all' && { plan: planFilter }),
    ...(searchQuery && { search: searchQuery }),
  })

  const {
    createTenant,
    updateTenant,
    deleteTenant,
    suspendTenant,
    reactivateTenant,
    creating,
    updating,
    deleting,
    suspending,
    reactivating,
  } = useTenantMutations()

  const handleCreate = async (input: TenantCreateInput) => {
    await createTenant(input)
    setShowCreateForm(false)
    refetch()
  }

  const handleUpdate = async (updates: TenantUpdateInput) => {
    if (!editingTenant) return
    await updateTenant(editingTenant.id, updates)
    setEditingTenant(null)
    refetch()
  }

  const handleDelete = async () => {
    if (!deleteConfirm) return
    await deleteTenant(deleteConfirm.id)
    setDeleteConfirm(null)
    refetch()
  }

  const handleSuspend = async (tenantId: string) => {
    await suspendTenant(tenantId)
    refetch()
  }

  const handleReactivate = async (tenantId: string) => {
    await reactivateTenant(tenantId)
    refetch()
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Tenant Management</h1>
          <p className="text-gray-400 text-sm mt-1">
            Manage tenants, plans, and resource limits
          </p>
        </div>
        <button
          onClick={() => setShowCreateForm(true)}
          className="flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors"
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Add Tenant
        </button>
      </div>

      {/* Filters */}
      <div className="bg-gray-800 rounded-lg p-4 flex flex-col gap-4 sm:flex-row sm:items-center">
        {/* Search */}
        <div className="relative flex-1 max-w-md">
          <input
            type="text"
            placeholder="Search tenants..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full px-4 py-2 pl-10 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500"
          />
          <svg
            className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>

        <div className="flex items-center gap-3">
          {/* Status filter */}
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as TenantStatus | 'all')}
            className="px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <option value="all">All Statuses</option>
            <option value="active">Active</option>
            <option value="trial">Trial</option>
            <option value="suspended">Suspended</option>
            <option value="cancelled">Cancelled</option>
          </select>

          {/* Plan filter */}
          <select
            value={planFilter}
            onChange={(e) => setPlanFilter(e.target.value as TenantPlan | 'all')}
            className="px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <option value="all">All Plans</option>
            <option value="free">Free</option>
            <option value="starter">Starter</option>
            <option value="professional">Professional</option>
            <option value="enterprise">Enterprise</option>
          </select>

          {/* Refresh */}
          <button
            onClick={() => refetch()}
            disabled={loading}
            className="p-2 bg-gray-700 text-gray-400 hover:text-white rounded-lg transition-colors disabled:opacity-50"
          >
            <svg
              className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </button>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard label="Total Tenants" value={total} />
        <StatCard label="Active" value={tenants.filter((t) => t.status === 'active').length} color="green" />
        <StatCard label="Trial" value={tenants.filter((t) => t.status === 'trial').length} color="blue" />
        <StatCard label="Suspended" value={tenants.filter((t) => t.status === 'suspended').length} color="red" />
      </div>

      {/* Error state */}
      {error && (
        <div className="bg-red-500/10 border border-red-500/30 rounded-lg p-4 text-red-400">
          Failed to load tenants: {error.message}
        </div>
      )}

      {/* Tenant Table */}
      <TenantTable
        tenants={tenants}
        loading={loading}
        onEdit={setEditingTenant}
        onDelete={setDeleteConfirm}
        onSuspend={handleSuspend}
        onReactivate={handleReactivate}
        suspending={suspending}
        reactivating={reactivating}
      />

      {/* Create Tenant Modal */}
      {showCreateForm && (
        <TenantForm
          mode="create"
          onSubmit={handleCreate}
          onCancel={() => setShowCreateForm(false)}
          submitting={creating}
        />
      )}

      {/* Edit Tenant Modal */}
      {editingTenant && (
        <TenantForm
          mode="edit"
          tenant={editingTenant}
          onSubmit={handleUpdate}
          onCancel={() => setEditingTenant(null)}
          submitting={updating}
        />
      )}

      {/* Delete Confirmation Modal */}
      {deleteConfirm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-lg p-6 w-96 max-w-[90vw]">
            <h3 className="text-white text-lg font-semibold mb-4">Delete Tenant</h3>
            <p className="text-gray-400 mb-4">
              Are you sure you want to delete <span className="text-white font-medium">{deleteConfirm.name}</span>?
              This will permanently delete all associated data including workflows, executions, and users.
            </p>
            <div className="flex justify-end gap-2">
              <button
                onClick={() => setDeleteConfirm(null)}
                disabled={deleting}
                className="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 transition-colors disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                onClick={handleDelete}
                disabled={deleting}
                className="px-4 py-2 bg-red-600 text-white rounded-lg text-sm font-medium hover:bg-red-700 transition-colors disabled:opacity-50"
              >
                {deleting ? 'Deleting...' : 'Delete Tenant'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function StatCard({
  label,
  value,
  color = 'gray',
}: {
  label: string
  value: number
  color?: 'gray' | 'green' | 'blue' | 'red'
}) {
  const colorClasses = {
    gray: 'text-gray-400',
    green: 'text-green-400',
    blue: 'text-blue-400',
    red: 'text-red-400',
  }

  return (
    <div className="bg-gray-800 rounded-lg p-4">
      <p className="text-gray-400 text-sm">{label}</p>
      <p className={`text-2xl font-bold ${colorClasses[color]}`}>{value}</p>
    </div>
  )
}
