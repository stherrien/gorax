import React, { useState, useCallback } from 'react'
import { CronBuilder } from './CronBuilder'
import { TimezoneSelector } from './TimezoneSelector'
import { SchedulePreview } from './SchedulePreview'
import {
  validateCronExpression,
  validateTimezone,
  validateRequired,
  type ValidationResult,
} from '../../utils/formValidation'
import {
  useFormValidation,
  hasVisibleError,
  getVisibleError,
  getFormErrors,
} from '../../hooks/useFormValidation'
import { FormErrorSummary } from '../common/FormErrorSummary'

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

interface FormValues {
  name: string
  cronExpression: string
  timezone: string
}

/**
 * Validates the schedule form fields
 */
function validateScheduleFormValues(values: FormValues): ValidationResult {
  const errors: ValidationResult['errors'] = []

  // Validate name
  const nameResult = validateRequired(values.name, 'name')
  if (!nameResult.valid) {
    errors.push({
      field: 'name',
      message: 'Schedule name is required',
      code: 'required',
    })
  }

  // Validate cron expression
  const cronResult = validateCronExpression(values.cronExpression, 'cronExpression')
  if (!cronResult.valid) {
    // Use the first error message from cron validation
    const cronError = cronResult.errors[0]
    errors.push({
      field: 'cronExpression',
      message: cronError?.message || 'Invalid cron expression',
      code: cronError?.code || 'invalid_cron',
    })
  }

  // Validate timezone
  const tzResult = validateTimezone(values.timezone, 'timezone')
  if (!tzResult.valid) {
    errors.push(...tzResult.errors)
  }

  return { valid: errors.length === 0, errors }
}

export const ScheduleForm: React.FC<ScheduleFormProps> = ({
  initialData,
  onSubmit,
  onCancel,
  submitLabel,
}) => {
  const isEditMode = !!initialData
  const [enabled, setEnabled] = useState(
    initialData?.enabled !== undefined ? initialData.enabled : true
  )
  const [isSubmitting, setIsSubmitting] = useState(false)

  const {
    values,
    errors,
    touched,
    setFieldValue,
    setFieldTouched,
    validateForm,
    getFieldProps,
  } = useFormValidation({
    initialValues: {
      name: initialData?.name || '',
      cronExpression: initialData?.cronExpression || '0 9 * * *',
      timezone: initialData?.timezone || 'UTC',
    },
    validate: validateScheduleFormValues,
    validateOnChange: true,
    validateOnBlur: true,
  })

  const handleSubmit = useCallback(async (e: React.FormEvent) => {
    e.preventDefault()

    // Touch all fields to show any validation errors
    setFieldTouched('name', true)
    setFieldTouched('cronExpression', true)
    setFieldTouched('timezone', true)

    const result = await validateForm()
    if (!result.valid) {
      return
    }

    try {
      setIsSubmitting(true)
      await onSubmit({
        name: values.name,
        cronExpression: values.cronExpression,
        timezone: values.timezone,
        enabled,
      })
    } finally {
      setIsSubmitting(false)
    }
  }, [values, enabled, onSubmit, validateForm, setFieldTouched])

  const handleCronChange = useCallback((value: string) => {
    setFieldValue('cronExpression', value)
  }, [setFieldValue])

  const handleCronBlur = useCallback(() => {
    setFieldTouched('cronExpression', true)
  }, [setFieldTouched])

  const handleTimezoneChange = useCallback((timezone: string) => {
    setFieldValue('timezone', timezone)
    setFieldTouched('timezone', true)
  }, [setFieldValue, setFieldTouched])

  const handleEnabledChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setEnabled(e.target.checked)
  }, [])

  const buttonText = isSubmitting
    ? 'Creating...'
    : submitLabel || (isEditMode ? 'Update Schedule' : 'Create Schedule')

  const nameError = getVisibleError('name', errors, touched)
  const cronError = getVisibleError('cronExpression', errors, touched)
  const hasNameError = hasVisibleError('name', errors, touched)
  const hasCronError = hasVisibleError('cronExpression', errors, touched)

  return (
    <form onSubmit={handleSubmit} className="space-y-6" noValidate>
      {/* Form Error Summary */}
      <FormErrorSummary errors={getFormErrors(errors, touched)} />

      <div>
        <label
          htmlFor="schedule-name"
          className="block text-sm font-medium text-gray-700 mb-2"
        >
          Schedule Name
        </label>
        <input
          {...getFieldProps('name')}
          id="schedule-name"
          type="text"
          placeholder="e.g., Daily Data Sync"
          className={`w-full px-3 py-2 border rounded-md ${
            hasNameError ? 'border-red-500' : 'border-gray-300'
          } focus:outline-none focus:ring-2 focus:ring-blue-500`}
          disabled={isSubmitting}
          aria-label="Schedule Name"
          aria-invalid={hasNameError}
          aria-describedby={hasNameError ? 'name-error' : undefined}
        />
        {nameError && (
          <div id="name-error" className="mt-1 text-sm text-red-600" role="alert">
            {nameError}
          </div>
        )}
      </div>

      <div>
        <CronBuilder
          value={values.cronExpression}
          onChange={handleCronChange}
          onBlur={handleCronBlur}
          disabled={isSubmitting}
          error={hasCronError}
        />
        {cronError && (
          <div id="cron-error" className="mt-1 text-sm text-red-600" role="alert">
            {cronError}
          </div>
        )}
      </div>

      <div>
        <TimezoneSelector
          value={values.timezone}
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
          checked={enabled}
          onChange={handleEnabledChange}
          disabled={isSubmitting}
          className="h-4 w-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
          aria-label="Enabled"
        />
        <label htmlFor="schedule-enabled" className="ml-2 text-sm text-gray-700">
          Enabled
        </label>
      </div>

      {values.cronExpression && !hasCronError && (
        <div className="pt-4 border-t border-gray-200">
          <SchedulePreview
            cronExpression={values.cronExpression}
            timezone={values.timezone}
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
