import React, { useState } from 'react'
import { CronBuilder } from './CronBuilder'
import { TimezoneSelector } from './TimezoneSelector'
import { SchedulePreview } from './SchedulePreview'

export interface ScheduleFormData {
  name: string
  cronExpression: string
  timezone: string
  enabled: boolean
}

export interface ScheduleFormProps {
  initialData?: Partial<ScheduleFormData>
  onSubmit: (data: ScheduleFormData) => void | Promise<void>
  onCancel: () => void
  submitLabel?: string
}

interface FormErrors {
  name?: string
  cronExpression?: string
}

const validateCronExpression = (cron: string): boolean => {
  if (!cron || cron.trim() === '') return false
  const parts = cron.split(' ')
  return parts.length === 5
}

export const ScheduleForm: React.FC<ScheduleFormProps> = ({
  initialData,
  onSubmit,
  onCancel,
  submitLabel,
}) => {
  const isEditMode = !!initialData

  const [formData, setFormData] = useState<ScheduleFormData>({
    name: initialData?.name || '',
    cronExpression: initialData?.cronExpression || '0 9 * * *',
    timezone: initialData?.timezone || 'UTC',
    enabled: initialData?.enabled !== undefined ? initialData.enabled : true,
  })

  const [errors, setErrors] = useState<FormErrors>({})
  const [isSubmitting, setIsSubmitting] = useState(false)

  const validateForm = (): boolean => {
    const newErrors: FormErrors = {}

    if (!formData.name || formData.name.trim() === '') {
      newErrors.name = 'Schedule name is required'
    }

    if (!validateCronExpression(formData.cronExpression)) {
      newErrors.cronExpression = 'Invalid cron expression'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) {
      return
    }

    try {
      setIsSubmitting(true)
      await onSubmit(formData)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({ ...formData, name: e.target.value })
    if (errors.name) {
      setErrors({ ...errors, name: undefined })
    }
  }

  const handleCronChange = (value: string) => {
    setFormData({ ...formData, cronExpression: value })
    if (errors.cronExpression) {
      setErrors({ ...errors, cronExpression: undefined })
    }
  }

  const handleTimezoneChange = (timezone: string) => {
    setFormData({ ...formData, timezone })
  }

  const handleEnabledChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({ ...formData, enabled: e.target.checked })
  }

  const buttonText = isSubmitting
    ? 'Creating...'
    : submitLabel || (isEditMode ? 'Update Schedule' : 'Create Schedule')

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div>
        <label
          htmlFor="schedule-name"
          className="block text-sm font-medium text-gray-700 mb-2"
        >
          Schedule Name
        </label>
        <input
          id="schedule-name"
          type="text"
          value={formData.name}
          onChange={handleNameChange}
          placeholder="e.g., Daily Data Sync"
          className={`w-full px-3 py-2 border rounded-md ${
            errors.name ? 'border-red-500' : 'border-gray-300'
          }`}
          disabled={isSubmitting}
          aria-label="Schedule Name"
        />
        {errors.name && <div className="mt-1 text-sm text-red-600">{errors.name}</div>}
      </div>

      <div>
        <CronBuilder
          value={formData.cronExpression}
          onChange={handleCronChange}
          disabled={isSubmitting}
        />
        {errors.cronExpression && (
          <div className="mt-1 text-sm text-red-600">{errors.cronExpression}</div>
        )}
      </div>

      <div>
        <TimezoneSelector
          value={formData.timezone}
          onChange={handleTimezoneChange}
          disabled={isSubmitting}
          searchable
          showCurrentTime
        />
      </div>

      <div className="flex items-center">
        <input
          id="schedule-enabled"
          type="checkbox"
          checked={formData.enabled}
          onChange={handleEnabledChange}
          disabled={isSubmitting}
          className="h-4 w-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
          aria-label="Enabled"
        />
        <label htmlFor="schedule-enabled" className="ml-2 text-sm text-gray-700">
          Enabled
        </label>
      </div>

      {formData.cronExpression && (
        <div className="pt-4 border-t border-gray-200">
          <SchedulePreview
            cronExpression={formData.cronExpression}
            timezone={formData.timezone}
            count={10}
          />
        </div>
      )}

      <div className="flex justify-end gap-3 pt-4">
        <button
          type="button"
          onClick={onCancel}
          disabled={isSubmitting}
          className="px-4 py-2 text-sm border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={isSubmitting}
          className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {buttonText}
        </button>
      </div>
    </form>
  )
}
