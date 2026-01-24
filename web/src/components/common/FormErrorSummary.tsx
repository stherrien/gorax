/**
 * FormErrorSummary Component
 *
 * Displays a summary of form validation errors at the top of a form.
 * Shows only errors for fields that have been touched (interacted with).
 */

import type { FormFieldError } from '../../hooks/useFormValidation'

export interface FormErrorSummaryProps {
  /**
   * Array of errors to display
   */
  errors: FormFieldError[]
  /**
   * Optional title for the error summary
   * @default "Please fix the following errors:"
   */
  title?: string
  /**
   * Optional CSS class for the container
   */
  className?: string
  /**
   * Optional callback when an error item is clicked
   * Can be used to focus the corresponding field
   */
  onErrorClick?: (field: string) => void
}

/**
 * FormErrorSummary displays validation errors in an accessible alert box.
 *
 * @example
 * ```tsx
 * const errors = getFormErrors(form.errors, form.touched)
 *
 * return (
 *   <form>
 *     <FormErrorSummary errors={errors} />
 *     // ... form fields
 *   </form>
 * )
 * ```
 */
export function FormErrorSummary({
  errors,
  title = 'Please fix the following errors:',
  className = '',
  onErrorClick,
}: FormErrorSummaryProps): JSX.Element | null {
  if (errors.length === 0) {
    return null
  }

  const handleClick = (field: string) => {
    if (onErrorClick) {
      onErrorClick(field)
    }
  }

  const handleKeyDown = (field: string, event: React.KeyboardEvent) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      handleClick(field)
    }
  }

  return (
    <div
      className={`form-error-summary bg-red-50 border border-red-200 rounded-md p-4 mb-4 ${className}`}
      role="alert"
      aria-live="polite"
      aria-labelledby="error-summary-title"
      data-testid="form-error-summary"
    >
      <h2
        id="error-summary-title"
        className="text-sm font-medium text-red-800 mb-2"
      >
        {title}
      </h2>
      <ul className="list-disc list-inside space-y-1">
        {errors.map(({ field, message }) => (
          <li
            key={field}
            className={`text-sm text-red-700 ${
              onErrorClick ? 'cursor-pointer hover:underline' : ''
            }`}
            onClick={() => handleClick(field)}
            onKeyDown={(e) => handleKeyDown(field, e)}
            tabIndex={onErrorClick ? 0 : undefined}
            role={onErrorClick ? 'button' : undefined}
            aria-label={onErrorClick ? `Go to ${field}: ${message}` : undefined}
            data-testid={`error-item-${field}`}
          >
            <span className="font-medium">{formatFieldName(field)}:</span>{' '}
            {message}
          </li>
        ))}
      </ul>
    </div>
  )
}

/**
 * Formats a field name for display.
 * Converts snake_case or camelCase to Title Case.
 */
function formatFieldName(field: string): string {
  return field
    .replace(/_/g, ' ')
    .replace(/([a-z])([A-Z])/g, '$1 $2')
    .replace(/\b\w/g, (char) => char.toUpperCase())
}

/**
 * Inline form field error component.
 * Use this below individual form fields for field-level errors.
 */
export interface FieldErrorProps {
  /**
   * Error message to display
   */
  error: string | null | undefined
  /**
   * Optional CSS class
   */
  className?: string
  /**
   * Field ID for accessibility
   */
  fieldId?: string
}

export function FieldError({
  error,
  className = '',
  fieldId,
}: FieldErrorProps): JSX.Element | null {
  if (!error) {
    return null
  }

  return (
    <p
      className={`text-sm text-red-600 mt-1 ${className}`}
      role="alert"
      id={fieldId ? `${fieldId}-error` : undefined}
      data-testid={fieldId ? `field-error-${fieldId}` : 'field-error'}
    >
      {error}
    </p>
  )
}

export default FormErrorSummary
