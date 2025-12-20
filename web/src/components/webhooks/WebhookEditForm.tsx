import { useState } from 'react'
import type { Webhook, WebhookUpdateInput } from '../../api/webhooks'
import PrioritySelector from './PrioritySelector'

export interface WebhookEditFormProps {
  webhook: Webhook
  onSave: (updates: WebhookUpdateInput) => Promise<void>
  onCancel: () => void
}

export default function WebhookEditForm({ webhook, onSave, onCancel }: WebhookEditFormProps) {
  const [formData, setFormData] = useState({
    name: webhook.name,
    path: webhook.path,
    enabled: webhook.enabled,
    priority: webhook.priority,
    description: '',
  })

  const [errors, setErrors] = useState<Record<string, string>>({})
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)

  const handleChange = (field: string, value: any) => {
    setFormData((prev) => ({ ...prev, [field]: value }))
    if (errors[field]) {
      setErrors((prev) => {
        const newErrors = { ...prev }
        delete newErrors[field]
        return newErrors
      })
    }
  }

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {}

    if (!formData.name.trim()) {
      newErrors.name = 'Name is required'
    }

    if (!formData.path.trim()) {
      newErrors.path = 'Path is required'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validate()) {
      return
    }

    setSaving(true)
    setSaveError(null)

    try {
      await onSave({
        name: formData.name,
        path: formData.path,
        enabled: formData.enabled,
        priority: formData.priority,
      })
    } catch (error: any) {
      setSaveError(error.message || 'Save failed')
      setSaving(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {saveError && (
        <div className="p-3 bg-red-800/50 rounded-lg text-red-200 text-sm">{saveError}</div>
      )}

      <div>
        <label htmlFor="webhook-name" className="block text-sm font-medium text-gray-300 mb-2">
          Name
        </label>
        <input
          id="webhook-name"
          type="text"
          value={formData.name}
          onChange={(e) => handleChange('name', e.target.value)}
          disabled={saving}
          className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50"
        />
        {errors.name && <div className="mt-1 text-xs text-red-400">{errors.name}</div>}
      </div>

      <div>
        <label htmlFor="webhook-path" className="block text-sm font-medium text-gray-300 mb-2">
          Path
        </label>
        <input
          id="webhook-path"
          type="text"
          value={formData.path}
          onChange={(e) => handleChange('path', e.target.value)}
          disabled={saving}
          className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50"
        />
        {errors.path && <div className="mt-1 text-xs text-red-400">{errors.path}</div>}
      </div>

      <PrioritySelector
        value={formData.priority}
        onChange={(priority) => handleChange('priority', priority)}
        disabled={saving}
        id="webhook-priority"
      />

      <div>
        <label htmlFor="webhook-enabled" className="block text-sm font-medium text-gray-300 mb-2">
          Enabled
        </label>
        <button
          type="button"
          role="switch"
          aria-checked={formData.enabled}
          disabled={saving}
          onClick={() => handleChange('enabled', !formData.enabled)}
          className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 focus:ring-offset-gray-800 disabled:opacity-50 disabled:cursor-not-allowed ${
            formData.enabled ? 'bg-primary-600' : 'bg-gray-600'
          }`}
        >
          <span
            className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
              formData.enabled ? 'translate-x-6' : 'translate-x-1'
            }`}
          />
        </button>
      </div>

      <div className="flex space-x-2 pt-4">
        <button
          type="button"
          onClick={onCancel}
          disabled={saving}
          className="flex-1 px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={saving}
          className="flex-1 px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {saving ? 'Saving...' : 'Save'}
        </button>
      </div>
    </form>
  )
}
