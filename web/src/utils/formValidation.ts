/**
 * Form Validation Utilities
 *
 * Comprehensive validation functions for form fields including cron expressions,
 * timezone validation, and async validation support.
 */

import type { ValidationResult, ValidationError } from './validation'
import { validateName, validateDescription } from './validation'

export type { ValidationResult, ValidationError }
export { validateUUID } from './validation'

// --- Cron Expression Validation ---

/**
 * Cron field constraints
 */
const CRON_FIELDS = {
  minute: { min: 0, max: 59, name: 'minute' },
  hour: { min: 0, max: 23, name: 'hour' },
  dayOfMonth: { min: 1, max: 31, name: 'day of month' },
  month: { min: 1, max: 12, name: 'month' },
  dayOfWeek: { min: 0, max: 6, name: 'day of week' },
  second: { min: 0, max: 59, name: 'second' }, // Optional 6th field
} as const

/**
 * Month name mappings
 */
const MONTH_NAMES: Record<string, number> = {
  jan: 1, feb: 2, mar: 3, apr: 4, may: 5, jun: 6,
  jul: 7, aug: 8, sep: 9, oct: 10, nov: 11, dec: 12,
}

/**
 * Day of week name mappings
 */
const DAY_NAMES: Record<string, number> = {
  sun: 0, mon: 1, tue: 2, wed: 3, thu: 4, fri: 5, sat: 6,
}

/**
 * Validates a single cron field value
 */
function validateCronFieldValue(
  value: string,
  min: number,
  max: number,
  fieldName: string,
  nameMapping?: Record<string, number>
): ValidationError | null {
  // Handle wildcard
  if (value === '*') {
    return null
  }

  // Handle step values (e.g., */5, 0-59/10)
  if (value.includes('/')) {
    const [rangeOrWildcard, stepStr] = value.split('/')
    const step = parseInt(stepStr, 10)

    if (isNaN(step) || step < 1) {
      return {
        field: 'cron_expression',
        message: `Invalid step value '${stepStr}' in ${fieldName}`,
        code: 'invalid_cron_step',
      }
    }

    // Validate the range part
    if (rangeOrWildcard !== '*') {
      const rangeError = validateCronFieldValue(rangeOrWildcard, min, max, fieldName, nameMapping)
      if (rangeError) return rangeError
    }

    return null
  }

  // Handle ranges (e.g., 1-5)
  if (value.includes('-')) {
    const [startStr, endStr] = value.split('-')
    const start = parseCronValue(startStr, nameMapping)
    const end = parseCronValue(endStr, nameMapping)

    if (start === null || end === null) {
      return {
        field: 'cron_expression',
        message: `Invalid range '${value}' in ${fieldName}`,
        code: 'invalid_cron_range',
      }
    }

    if (start < min || start > max || end < min || end > max) {
      return {
        field: 'cron_expression',
        message: `Range '${value}' out of bounds for ${fieldName} (${min}-${max})`,
        code: 'invalid_cron_range',
      }
    }

    if (start > end) {
      return {
        field: 'cron_expression',
        message: `Invalid range '${value}' in ${fieldName}: start must be <= end`,
        code: 'invalid_cron_range',
      }
    }

    return null
  }

  // Handle lists (e.g., 1,3,5)
  if (value.includes(',')) {
    const parts = value.split(',')
    for (const part of parts) {
      const error = validateCronFieldValue(part.trim(), min, max, fieldName, nameMapping)
      if (error) return error
    }
    return null
  }

  // Handle single values
  const num = parseCronValue(value, nameMapping)
  if (num === null || num < min || num > max) {
    return {
      field: 'cron_expression',
      message: `Invalid value '${value}' for ${fieldName} (must be ${min}-${max})`,
      code: 'invalid_cron_value',
    }
  }

  return null
}

/**
 * Parses a cron value (number or name)
 */
function parseCronValue(value: string, nameMapping?: Record<string, number>): number | null {
  // Try numeric first
  const num = parseInt(value, 10)
  if (!isNaN(num)) {
    return num
  }

  // Try name mapping
  if (nameMapping) {
    const lower = value.toLowerCase()
    if (lower in nameMapping) {
      return nameMapping[lower]
    }
  }

  return null
}

/**
 * Validates a complete cron expression
 *
 * Supports both 5-part (standard) and 6-part (with seconds) cron expressions.
 * Format: minute hour day-of-month month day-of-week [second]
 * Or: second minute hour day-of-month month day-of-week
 *
 * Examples:
 * - "0 9 * * *"     → Every day at 9:00 AM
 * - "0 0 * * 0"     → Every Sunday at midnight
 * - "* /15 * * * *"  → Every 15 minutes
 * - "0 9 * * MON-FRI" → Weekdays at 9:00 AM
 */
export function validateCronExpression(
  cron: string,
  fieldName = 'cron_expression'
): ValidationResult {
  const errors: ValidationError[] = []

  if (!cron || cron.trim() === '') {
    errors.push({
      field: fieldName,
      message: 'Cron expression is required',
      code: 'required',
    })
    return { valid: false, errors }
  }

  const parts = cron.trim().split(/\s+/)

  // Must have 5 or 6 parts
  if (parts.length < 5 || parts.length > 6) {
    errors.push({
      field: fieldName,
      message: `Cron expression must have 5 or 6 parts, got ${parts.length}`,
      code: 'invalid_cron_format',
    })
    return { valid: false, errors }
  }

  // Determine if 5-part (no seconds) or 6-part (with seconds)
  const hasSeconds = parts.length === 6
  let fieldIndex = 0

  // If 6 parts, first field is seconds
  if (hasSeconds) {
    const secondError = validateCronFieldValue(
      parts[fieldIndex++],
      CRON_FIELDS.second.min,
      CRON_FIELDS.second.max,
      CRON_FIELDS.second.name
    )
    if (secondError) errors.push(secondError)
  }

  // Minute
  const minuteError = validateCronFieldValue(
    parts[fieldIndex++],
    CRON_FIELDS.minute.min,
    CRON_FIELDS.minute.max,
    CRON_FIELDS.minute.name
  )
  if (minuteError) errors.push(minuteError)

  // Hour
  const hourError = validateCronFieldValue(
    parts[fieldIndex++],
    CRON_FIELDS.hour.min,
    CRON_FIELDS.hour.max,
    CRON_FIELDS.hour.name
  )
  if (hourError) errors.push(hourError)

  // Day of month
  const dayOfMonthError = validateCronFieldValue(
    parts[fieldIndex++],
    CRON_FIELDS.dayOfMonth.min,
    CRON_FIELDS.dayOfMonth.max,
    CRON_FIELDS.dayOfMonth.name
  )
  if (dayOfMonthError) errors.push(dayOfMonthError)

  // Month
  const monthError = validateCronFieldValue(
    parts[fieldIndex++],
    CRON_FIELDS.month.min,
    CRON_FIELDS.month.max,
    CRON_FIELDS.month.name,
    MONTH_NAMES
  )
  if (monthError) errors.push(monthError)

  // Day of week
  const dayOfWeekError = validateCronFieldValue(
    parts[fieldIndex],
    CRON_FIELDS.dayOfWeek.min,
    CRON_FIELDS.dayOfWeek.max,
    CRON_FIELDS.dayOfWeek.name,
    DAY_NAMES
  )
  if (dayOfWeekError) errors.push(dayOfWeekError)

  return { valid: errors.length === 0, errors }
}

// --- Timezone Validation ---

/**
 * Common timezone identifiers (subset for quick validation)
 */
const COMMON_TIMEZONES = new Set([
  'UTC',
  'America/New_York',
  'America/Chicago',
  'America/Denver',
  'America/Los_Angeles',
  'America/Phoenix',
  'America/Anchorage',
  'America/Toronto',
  'America/Vancouver',
  'America/Mexico_City',
  'America/Sao_Paulo',
  'Europe/London',
  'Europe/Paris',
  'Europe/Berlin',
  'Europe/Amsterdam',
  'Europe/Madrid',
  'Europe/Rome',
  'Europe/Moscow',
  'Asia/Tokyo',
  'Asia/Shanghai',
  'Asia/Hong_Kong',
  'Asia/Singapore',
  'Asia/Mumbai',
  'Asia/Dubai',
  'Asia/Seoul',
  'Australia/Sydney',
  'Australia/Melbourne',
  'Pacific/Auckland',
  'Pacific/Honolulu',
])

/**
 * Validates a timezone string
 */
export function validateTimezone(
  timezone: string,
  fieldName = 'timezone'
): ValidationResult {
  const errors: ValidationError[] = []

  if (!timezone || timezone.trim() === '') {
    // Timezone can be optional, default to UTC
    return { valid: true, errors }
  }

  // Quick check against common timezones
  if (COMMON_TIMEZONES.has(timezone)) {
    return { valid: true, errors }
  }

  // Try to validate using Intl API
  try {
    Intl.DateTimeFormat(undefined, { timeZone: timezone })
    return { valid: true, errors }
  } catch {
    errors.push({
      field: fieldName,
      message: `Invalid timezone: ${timezone}`,
      code: 'invalid_timezone',
    })
    return { valid: false, errors }
  }
}

// --- Required Field Validation ---

/**
 * Validates that a field has a non-empty value
 */
export function validateRequired(
  value: unknown,
  fieldName: string
): ValidationResult {
  const errors: ValidationError[] = []

  if (value === null || value === undefined) {
    errors.push({
      field: fieldName,
      message: `${fieldName} is required`,
      code: 'required',
    })
    return { valid: false, errors }
  }

  if (typeof value === 'string' && value.trim() === '') {
    errors.push({
      field: fieldName,
      message: `${fieldName} is required`,
      code: 'required',
    })
    return { valid: false, errors }
  }

  return { valid: true, errors }
}

/**
 * Validates schedule form data
 */
export interface ScheduleFormData {
  name: string
  cron_expression: string
  timezone?: string
  enabled?: boolean
}

export function validateScheduleForm(data: ScheduleFormData): ValidationResult {
  const errors: ValidationError[] = []

  // Name validation
  const nameResult = validateName(data.name, 'name', 100)
  if (!nameResult.valid) {
    errors.push(...nameResult.errors)
  }

  // Cron expression validation
  const cronResult = validateCronExpression(data.cron_expression)
  if (!cronResult.valid) {
    errors.push(...cronResult.errors)
  }

  // Timezone validation (optional)
  if (data.timezone) {
    const tzResult = validateTimezone(data.timezone)
    if (!tzResult.valid) {
      errors.push(...tzResult.errors)
    }
  }

  return { valid: errors.length === 0, errors }
}

/**
 * Validates workflow form data
 */
export interface WorkflowFormData {
  name: string
  description?: string
}

export function validateWorkflowForm(data: WorkflowFormData): ValidationResult {
  const errors: ValidationError[] = []

  // Name validation
  const nameResult = validateName(data.name, 'name', 100)
  if (!nameResult.valid) {
    errors.push(...nameResult.errors)
  }

  // Description validation (optional)
  if (data.description) {
    const descResult = validateDescription(data.description)
    if (!descResult.valid) {
      errors.push(...descResult.errors)
    }
  }

  return { valid: errors.length === 0, errors }
}

// --- Async Validation ---

export interface AsyncValidationResult extends ValidationResult {
  pending?: boolean
}

/**
 * Creates an async validator that can be debounced
 */
export function createAsyncValidator<T>(
  validateFn: (value: T) => Promise<ValidationResult>,
  debounceMs = 300
): (value: T) => Promise<AsyncValidationResult> {
  let timeoutId: ReturnType<typeof setTimeout> | null = null
  let lastValue: T | undefined

  return (value: T): Promise<AsyncValidationResult> => {
    return new Promise((resolve) => {
      if (timeoutId) {
        clearTimeout(timeoutId)
      }

      lastValue = value

      timeoutId = setTimeout(async () => {
        try {
          const result = await validateFn(value)
          // Only resolve if this is still the latest value
          if (value === lastValue) {
            resolve(result)
          }
        } catch (error) {
          resolve({
            valid: false,
            errors: [{
              field: 'unknown',
              message: error instanceof Error ? error.message : 'Validation failed',
              code: 'async_error',
            }],
          })
        }
      }, debounceMs)
    })
  }
}

/**
 * Validates cron expression via API (for server-side validation)
 */
export async function validateCronExpressionAsync(
  cron: string,
  timezone: string = 'UTC',
  apiEndpoint: string = '/api/v1/schedules/parse-cron'
): Promise<ValidationResult> {
  // First, do client-side validation
  const clientResult = validateCronExpression(cron)
  if (!clientResult.valid) {
    return clientResult
  }

  try {
    const response = await fetch(apiEndpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ cron_expression: cron, timezone }),
    })

    if (!response.ok) {
      const data = await response.json()
      return {
        valid: false,
        errors: [{
          field: 'cron_expression',
          message: data.error || 'Invalid cron expression',
          code: 'server_validation_error',
        }],
      }
    }

    return { valid: true, errors: [] }
  } catch {
    // If server is unreachable, rely on client-side validation
    return clientResult
  }
}
