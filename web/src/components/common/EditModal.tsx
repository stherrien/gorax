import { useState, useEffect, useCallback, FormEvent, KeyboardEvent } from 'react'

export interface EditField {
  name: string
  label: string
  type: 'text' | 'textarea' | 'select' | 'checkbox' | 'number'
  required?: boolean
  placeholder?: string
  options?: { value: string; label: string }[]
  maxLength?: number
  min?: number
  max?: number
}

export interface EditModalProps<T extends Record<string, unknown>> {
  isOpen: boolean
  title: string
  fields: EditField[]
  initialValues: T
  onSave: (values: T) => Promise<void>
  onCancel: () => void
  itemId?: string
}

type FormErrors = Record<string, string>

export function EditModal<T extends Record<string, unknown>>({
  isOpen,
  title,
  fields,
  initialValues,
  onSave,
  onCancel,
  itemId,
}: EditModalProps<T>) {
  const [formData, setFormData] = useState<T>(initialValues)
  const [errors, setErrors] = useState<FormErrors>({})
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)

  // Reset form when modal opens with new initial values
  useEffect(() => {
    if (isOpen) {
      setFormData(initialValues)
      setErrors({})
      setSaveError(null)
    }
  }, [isOpen, initialValues])

  const handleBackdropClick = useCallback(
    (e: React.MouseEvent) => {
      if (e.target === e.currentTarget && !saving) {
        onCancel()
      }
    },
    [onCancel, saving]
  )

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === 'Escape' && !saving) {
        onCancel()
      }
    },
    [onCancel, saving]
  )

  const handleChange = useCallback((fieldName: string, value: unknown) => {
    setFormData((prev) => ({ ...prev, [fieldName]: value }))
    setErrors((prev) => {
      if (prev[fieldName]) {
        const newErrors = { ...prev }
        delete newErrors[fieldName]
        return newErrors
      }
      return prev
    })
  }, [])

  const validate = useCallback((): boolean => {
    const newErrors: FormErrors = {}

    for (const field of fields) {
      const value = formData[field.name]

      if (field.required) {
        if (value === undefined || value === null || value === '') {
          newErrors[field.name] = `${field.label} is required`
          continue
        }
      }

      if (field.maxLength && typeof value === 'string' && value.length > field.maxLength) {
        newErrors[field.name] = `${field.label} must be ${field.maxLength} characters or less`
      }

      if (field.type === 'number' && value !== undefined && value !== null && value !== '') {
        const numValue = typeof value === 'number' ? value : Number(value)
        if (!Number.isNaN(numValue)) {
          if (field.min !== undefined && numValue < field.min) {
            newErrors[field.name] = `${field.label} must be at least ${field.min}`
          }
          if (field.max !== undefined && numValue > field.max) {
            newErrors[field.name] = `${field.label} must be at most ${field.max}`
          }
        }
      }
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }, [fields, formData])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()

    if (!validate()) {
      return
    }

    setSaving(true)
    setSaveError(null)

    try {
      await onSave(formData)
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Save failed'
      setSaveError(errorMessage)
    } finally {
      setSaving(false)
    }
  }

  if (!isOpen) {
    return null
  }

  return (
    <div
      data-testid="edit-modal-backdrop"
      onClick={handleBackdropClick}
      onKeyDown={handleKeyDown}
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
      role="presentation"
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="edit-modal-title"
        className="bg-gray-800 rounded-lg p-6 max-w-lg w-full max-h-[90vh] overflow-y-auto"
      >
        {/* Header */}
        <div className="flex justify-between items-start mb-6">
          <div>
            <h2 id="edit-modal-title" className="text-xl font-bold text-white">
              {title}
            </h2>
            {itemId && <p className="text-sm text-gray-400 mt-1">ID: {itemId}</p>}
          </div>
          <button
            type="button"
            onClick={onCancel}
            disabled={saving}
            className="text-gray-400 hover:text-white transition-colors disabled:opacity-50"
            aria-label="Close"
          >
            <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        {/* Error Message */}
        {saveError && (
          <div className="mb-4 p-3 bg-red-900/30 border border-red-600/50 rounded-lg">
            <p className="text-red-300 text-sm">{saveError}</p>
          </div>
        )}

        {/* Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          {fields.map((field) => (
            <div key={field.name}>
              <label
                htmlFor={`edit-modal-${field.name}`}
                className="block text-sm font-medium text-gray-300 mb-2"
              >
                {field.label}
                {field.required && <span className="text-red-400 ml-1">*</span>}
              </label>

              {field.type === 'textarea' ? (
                <textarea
                  id={`edit-modal-${field.name}`}
                  value={(formData[field.name] as string) || ''}
                  onChange={(e) => handleChange(field.name, e.target.value)}
                  disabled={saving}
                  placeholder={field.placeholder}
                  maxLength={field.maxLength}
                  rows={4}
                  className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
                />
              ) : field.type === 'select' ? (
                <select
                  id={`edit-modal-${field.name}`}
                  value={(formData[field.name] as string) || ''}
                  onChange={(e) => handleChange(field.name, e.target.value)}
                  disabled={saving}
                  className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <option value="">Select {field.label}</option>
                  {field.options?.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              ) : field.type === 'checkbox' ? (
                <button
                  type="button"
                  id={`edit-modal-${field.name}`}
                  role="switch"
                  aria-checked={Boolean(formData[field.name])}
                  disabled={saving}
                  onClick={() => handleChange(field.name, !formData[field.name])}
                  className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed ${
                    formData[field.name] ? 'bg-primary-600' : 'bg-gray-600'
                  }`}
                >
                  <span
                    className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                      formData[field.name] ? 'translate-x-6' : 'translate-x-1'
                    }`}
                  />
                </button>
              ) : (
                <input
                  id={`edit-modal-${field.name}`}
                  type={field.type}
                  value={(formData[field.name] as string | number) ?? ''}
                  onChange={(e) =>
                    handleChange(
                      field.name,
                      field.type === 'number' ? Number(e.target.value) : e.target.value
                    )
                  }
                  disabled={saving}
                  placeholder={field.placeholder}
                  maxLength={field.maxLength}
                  min={field.min}
                  max={field.max}
                  className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
                />
              )}

              {errors[field.name] && (
                <p className="mt-1 text-xs text-red-400">{errors[field.name]}</p>
              )}

              {field.maxLength && field.type !== 'number' && (
                <p className="mt-1 text-xs text-gray-500">
                  {((formData[field.name] as string) || '').length}/{field.maxLength}
                </p>
              )}
            </div>
          ))}

          {/* Actions */}
          <div className="flex justify-end gap-3 pt-4">
            <button
              type="button"
              onClick={onCancel}
              disabled={saving}
              className="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={saving}
              className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {saving ? 'Saving...' : 'Save'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
