import { useState } from 'react'
import type { Tenant, TenantCreateInput, TenantUpdateInput, TenantPlan, TenantStatus } from '../../types/management'

interface TenantFormCreateProps {
  mode: 'create'
  tenant?: never
  onSubmit: (data: TenantCreateInput) => Promise<void>
  onCancel: () => void
  submitting: boolean
}

interface TenantFormEditProps {
  mode: 'edit'
  tenant: Tenant
  onSubmit: (data: TenantUpdateInput) => Promise<void>
  onCancel: () => void
  submitting: boolean
}

type TenantFormProps = TenantFormCreateProps | TenantFormEditProps

export function TenantForm({ mode, tenant, onSubmit, onCancel, submitting }: TenantFormProps) {
  const [formData, setFormData] = useState({
    name: tenant?.name || '',
    slug: tenant?.slug || '',
    plan: tenant?.plan || 'free' as TenantPlan,
    status: tenant?.status || 'active' as TenantStatus,
    ownerEmail: '',
    maxWorkflows: tenant?.limits?.maxWorkflows || 10,
    maxExecutionsPerMonth: tenant?.limits?.maxExecutionsPerMonth || 1000,
    maxUsers: tenant?.limits?.maxUsers || 5,
    maxCredentials: tenant?.limits?.maxCredentials || 10,
    retentionDays: tenant?.limits?.retentionDays || 30,
  })

  const [errors, setErrors] = useState<Record<string, string>>({})

  const validateForm = () => {
    const newErrors: Record<string, string> = {}

    if (!formData.name.trim()) {
      newErrors.name = 'Name is required'
    }

    if (mode === 'create') {
      if (!formData.slug.trim()) {
        newErrors.slug = 'Slug is required'
      } else if (!/^[a-z0-9-]+$/.test(formData.slug)) {
        newErrors.slug = 'Slug must contain only lowercase letters, numbers, and hyphens'
      }

      if (!formData.ownerEmail) {
        newErrors.ownerEmail = 'Owner email is required'
      } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.ownerEmail)) {
        newErrors.ownerEmail = 'Invalid email format'
      }
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) return

    if (mode === 'create') {
      await onSubmit({
        name: formData.name,
        slug: formData.slug,
        plan: formData.plan,
        ownerEmail: formData.ownerEmail,
      })
    } else {
      await onSubmit({
        name: formData.name,
        status: formData.status,
        plan: formData.plan,
        limits: {
          maxWorkflows: formData.maxWorkflows,
          maxExecutionsPerMonth: formData.maxExecutionsPerMonth,
          maxUsers: formData.maxUsers,
          maxCredentials: formData.maxCredentials,
          retentionDays: formData.retentionDays,
        },
      })
    }
  }

  const handleSlugGenerate = () => {
    const slug = formData.name
      .toLowerCase()
      .replace(/[^a-z0-9\s-]/g, '')
      .replace(/\s+/g, '-')
      .replace(/-+/g, '-')
      .trim()
    setFormData({ ...formData, slug })
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-gray-800 rounded-lg w-full max-w-lg max-h-[90vh] overflow-y-auto">
        <div className="px-6 py-4 border-b border-gray-700 sticky top-0 bg-gray-800">
          <h2 className="text-lg font-semibold text-white">
            {mode === 'create' ? 'Add New Tenant' : 'Edit Tenant'}
          </h2>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {/* Name */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1">
              Name <span className="text-red-400">*</span>
            </label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className={`w-full px-4 py-2 bg-gray-700 border rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500 ${
                errors.name ? 'border-red-500' : 'border-gray-600'
              }`}
              placeholder="Acme Corporation"
            />
            {errors.name && <p className="mt-1 text-sm text-red-400">{errors.name}</p>}
          </div>

          {/* Slug (only for create) */}
          {mode === 'create' && (
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-1">
                Slug <span className="text-red-400">*</span>
              </label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={formData.slug}
                  onChange={(e) => setFormData({ ...formData, slug: e.target.value })}
                  className={`flex-1 px-4 py-2 bg-gray-700 border rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500 ${
                    errors.slug ? 'border-red-500' : 'border-gray-600'
                  }`}
                  placeholder="acme-corp"
                />
                <button
                  type="button"
                  onClick={handleSlugGenerate}
                  className="px-3 py-2 bg-gray-700 text-gray-300 rounded-lg hover:bg-gray-600 transition-colors text-sm"
                >
                  Generate
                </button>
              </div>
              {errors.slug && <p className="mt-1 text-sm text-red-400">{errors.slug}</p>}
              <p className="mt-1 text-xs text-gray-500">
                URL-friendly identifier (lowercase, no spaces)
              </p>
            </div>
          )}

          {/* Owner Email (only for create) */}
          {mode === 'create' && (
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-1">
                Owner Email <span className="text-red-400">*</span>
              </label>
              <input
                type="email"
                value={formData.ownerEmail}
                onChange={(e) => setFormData({ ...formData, ownerEmail: e.target.value })}
                className={`w-full px-4 py-2 bg-gray-700 border rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500 ${
                  errors.ownerEmail ? 'border-red-500' : 'border-gray-600'
                }`}
                placeholder="admin@acme.com"
              />
              {errors.ownerEmail && <p className="mt-1 text-sm text-red-400">{errors.ownerEmail}</p>}
            </div>
          )}

          {/* Plan */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1">Plan</label>
            <select
              value={formData.plan}
              onChange={(e) => setFormData({ ...formData, plan: e.target.value as TenantPlan })}
              className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="free">Free</option>
              <option value="starter">Starter</option>
              <option value="professional">Professional</option>
              <option value="enterprise">Enterprise</option>
            </select>
          </div>

          {/* Status (only for edit) */}
          {mode === 'edit' && (
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-1">Status</label>
              <select
                value={formData.status}
                onChange={(e) => setFormData({ ...formData, status: e.target.value as TenantStatus })}
                className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              >
                <option value="active">Active</option>
                <option value="trial">Trial</option>
                <option value="suspended">Suspended</option>
                <option value="cancelled">Cancelled</option>
              </select>
            </div>
          )}

          {/* Limits (only for edit) */}
          {mode === 'edit' && (
            <div className="space-y-4 pt-4 border-t border-gray-700">
              <h3 className="text-sm font-medium text-gray-300">Resource Limits</h3>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs text-gray-400 mb-1">Max Workflows</label>
                  <input
                    type="number"
                    min="1"
                    value={formData.maxWorkflows}
                    onChange={(e) =>
                      setFormData({ ...formData, maxWorkflows: parseInt(e.target.value) || 0 })
                    }
                    className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
                  />
                </div>
                <div>
                  <label className="block text-xs text-gray-400 mb-1">Max Users</label>
                  <input
                    type="number"
                    min="1"
                    value={formData.maxUsers}
                    onChange={(e) =>
                      setFormData({ ...formData, maxUsers: parseInt(e.target.value) || 0 })
                    }
                    className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
                  />
                </div>
                <div>
                  <label className="block text-xs text-gray-400 mb-1">Max Executions/Month</label>
                  <input
                    type="number"
                    min="1"
                    value={formData.maxExecutionsPerMonth}
                    onChange={(e) =>
                      setFormData({ ...formData, maxExecutionsPerMonth: parseInt(e.target.value) || 0 })
                    }
                    className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
                  />
                </div>
                <div>
                  <label className="block text-xs text-gray-400 mb-1">Max Credentials</label>
                  <input
                    type="number"
                    min="1"
                    value={formData.maxCredentials}
                    onChange={(e) =>
                      setFormData({ ...formData, maxCredentials: parseInt(e.target.value) || 0 })
                    }
                    className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs text-gray-400 mb-1">Log Retention (days)</label>
                <input
                  type="number"
                  min="1"
                  value={formData.retentionDays}
                  onChange={(e) =>
                    setFormData({ ...formData, retentionDays: parseInt(e.target.value) || 0 })
                  }
                  className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
                />
              </div>
            </div>
          )}

          {/* Actions */}
          <div className="flex justify-end gap-3 pt-4">
            <button
              type="button"
              onClick={onCancel}
              disabled={submitting}
              className="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 transition-colors disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors disabled:opacity-50"
            >
              {submitting
                ? mode === 'create'
                  ? 'Creating...'
                  : 'Saving...'
                : mode === 'create'
                ? 'Create Tenant'
                : 'Save Changes'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
