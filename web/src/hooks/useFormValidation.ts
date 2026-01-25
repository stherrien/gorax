/**
 * Form Validation Hook
 *
 * A standardized hook for form state management with validation support.
 * Provides consistent form handling across all components.
 */

import { useState, useCallback, useMemo, useRef, useEffect } from 'react'
import type { ValidationResult } from '../utils/validation'

// --- Types ---

export interface FieldState {
  value: string
  touched: boolean
  error: string | null
}

export interface FormState<T extends Record<string, string>> {
  values: T
  errors: Partial<Record<keyof T, string>>
  touched: Partial<Record<keyof T, boolean>>
  isSubmitting: boolean
  isValid: boolean
  isDirty: boolean
}

export interface UseFormValidationOptions<T extends Record<string, string>> {
  initialValues: T
  validate?: (values: T) => ValidationResult
  validateAsync?: (values: T) => Promise<ValidationResult>
  onSubmit?: (values: T) => Promise<void> | void
  validateOnChange?: boolean
  validateOnBlur?: boolean
}

export interface UseFormValidationReturn<T extends Record<string, string>> {
  // State
  values: T
  errors: Partial<Record<keyof T, string>>
  touched: Partial<Record<keyof T, boolean>>
  isSubmitting: boolean
  isValid: boolean
  isDirty: boolean
  isValidating: boolean

  // Actions
  handleChange: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => void
  handleBlur: (e: React.FocusEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => void
  handleSubmit: (e?: React.FormEvent) => Promise<void>
  setFieldValue: (field: keyof T, value: string) => void
  setFieldError: (field: keyof T, error: string | null) => void
  setFieldTouched: (field: keyof T, touched: boolean) => void
  setValues: (values: Partial<T>) => void
  setErrors: (errors: Partial<Record<keyof T, string>>) => void
  reset: () => void
  resetField: (field: keyof T) => void
  validateField: (field: keyof T) => ValidationResult
  validateForm: () => Promise<ValidationResult>

  // Field props helper
  getFieldProps: (field: keyof T) => {
    name: string
    value: string
    onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => void
    onBlur: (e: React.FocusEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => void
  }
}

// --- Hook Implementation ---

export function useFormValidation<T extends Record<string, string>>(
  options: UseFormValidationOptions<T>
): UseFormValidationReturn<T> {
  const {
    initialValues,
    validate,
    validateAsync,
    onSubmit,
    validateOnChange = true,
    validateOnBlur = true,
  } = options

  // Store initial values for reset
  const initialValuesRef = useRef(initialValues)

  // State
  const [values, setValuesState] = useState<T>(initialValues)
  const [errors, setErrorsState] = useState<Partial<Record<keyof T, string>>>({})
  const [touched, setTouchedState] = useState<Partial<Record<keyof T, boolean>>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isValidating, setIsValidating] = useState(false)

  // Update initial values ref when they change
  useEffect(() => {
    initialValuesRef.current = initialValues
  }, [initialValues])

  // Derived state
  const isValid = useMemo(() => {
    return Object.keys(errors).length === 0 && Object.values(errors).every((e) => !e)
  }, [errors])

  const isDirty = useMemo(() => {
    return Object.keys(values).some(
      (key) => values[key as keyof T] !== initialValuesRef.current[key as keyof T]
    )
  }, [values])

  // Run validation and update errors
  const runValidation = useCallback((valuesToValidate: T): ValidationResult => {
    if (!validate) {
      return { valid: true, errors: [] }
    }

    const result = validate(valuesToValidate)
    const newErrors: Partial<Record<keyof T, string>> = {}

    result.errors.forEach((error) => {
      newErrors[error.field as keyof T] = error.message
    })

    setErrorsState(newErrors)
    return result
  }, [validate])

  // Run async validation
  const runAsyncValidation = useCallback(async (valuesToValidate: T): Promise<ValidationResult> => {
    // First run sync validation
    const syncResult = runValidation(valuesToValidate)
    if (!syncResult.valid) {
      return syncResult
    }

    // Then run async validation if provided
    if (!validateAsync) {
      return syncResult
    }

    setIsValidating(true)
    try {
      const asyncResult = await validateAsync(valuesToValidate)
      const newErrors: Partial<Record<keyof T, string>> = {}

      asyncResult.errors.forEach((error) => {
        newErrors[error.field as keyof T] = error.message
      })

      setErrorsState((prev) => ({ ...prev, ...newErrors }))
      return asyncResult
    } finally {
      setIsValidating(false)
    }
  }, [runValidation, validateAsync])

  // Field-level validation
  const validateField = useCallback((field: keyof T): ValidationResult => {
    if (!validate) {
      return { valid: true, errors: [] }
    }

    const result = validate(values)
    const fieldErrors = result.errors.filter((e) => e.field === field)

    if (fieldErrors.length > 0) {
      setErrorsState((prev) => ({
        ...prev,
        [field]: fieldErrors[0].message,
      }))
      return { valid: false, errors: fieldErrors }
    }

    setErrorsState((prev) => {
      const newErrors = { ...prev }
      delete newErrors[field]
      return newErrors
    })

    return { valid: true, errors: [] }
  }, [validate, values])

  // Full form validation
  const validateForm = useCallback(async (): Promise<ValidationResult> => {
    return runAsyncValidation(values)
  }, [runAsyncValidation, values])

  // Handle input change
  const handleChange = useCallback((
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  ) => {
    const { name, value } = e.target
    const field = name as keyof T

    setValuesState((prev) => ({
      ...prev,
      [field]: value,
    }))

    if (validateOnChange) {
      // Validate with the new value
      if (validate) {
        const newValues = { ...values, [field]: value }
        const result = validate(newValues)
        const fieldErrors = result.errors.filter((e) => e.field === field)

        if (fieldErrors.length > 0) {
          setErrorsState((prev) => ({
            ...prev,
            [field]: fieldErrors[0].message,
          }))
        } else {
          setErrorsState((prev) => {
            const newErrors = { ...prev }
            delete newErrors[field]
            return newErrors
          })
        }
      }
    }
  }, [validate, validateOnChange, values])

  // Handle input blur
  const handleBlur = useCallback((
    e: React.FocusEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  ) => {
    const { name } = e.target
    const field = name as keyof T

    setTouchedState((prev) => ({
      ...prev,
      [field]: true,
    }))

    if (validateOnBlur) {
      validateField(field)
    }
  }, [validateField, validateOnBlur])

  // Handle form submit
  const handleSubmit = useCallback(async (e?: React.FormEvent) => {
    if (e) {
      e.preventDefault()
    }

    // Mark all fields as touched
    const allTouched: Partial<Record<keyof T, boolean>> = {}
    Object.keys(values).forEach((key) => {
      allTouched[key as keyof T] = true
    })
    setTouchedState(allTouched)

    // Validate form
    const result = await runAsyncValidation(values)

    if (!result.valid) {
      return
    }

    // Call onSubmit handler
    if (onSubmit) {
      setIsSubmitting(true)
      try {
        await onSubmit(values)
      } finally {
        setIsSubmitting(false)
      }
    }
  }, [values, runAsyncValidation, onSubmit])

  // Set field value programmatically
  const setFieldValue = useCallback((field: keyof T, value: string) => {
    setValuesState((prev) => ({
      ...prev,
      [field]: value,
    }))
  }, [])

  // Set field error programmatically
  const setFieldError = useCallback((field: keyof T, error: string | null) => {
    if (error) {
      setErrorsState((prev) => ({
        ...prev,
        [field]: error,
      }))
    } else {
      setErrorsState((prev) => {
        const newErrors = { ...prev }
        delete newErrors[field]
        return newErrors
      })
    }
  }, [])

  // Set field touched programmatically
  const setFieldTouched = useCallback((field: keyof T, touched: boolean) => {
    setTouchedState((prev) => ({
      ...prev,
      [field]: touched,
    }))
  }, [])

  // Set multiple values
  const setValues = useCallback((newValues: Partial<T>) => {
    setValuesState((prev) => ({
      ...prev,
      ...newValues,
    }))
  }, [])

  // Set multiple errors
  const setErrors = useCallback((newErrors: Partial<Record<keyof T, string>>) => {
    setErrorsState(newErrors)
  }, [])

  // Reset form to initial values
  const reset = useCallback(() => {
    setValuesState(initialValuesRef.current)
    setErrorsState({})
    setTouchedState({})
  }, [])

  // Reset single field
  const resetField = useCallback((field: keyof T) => {
    setValuesState((prev) => ({
      ...prev,
      [field]: initialValuesRef.current[field],
    }))
    setErrorsState((prev) => {
      const newErrors = { ...prev }
      delete newErrors[field]
      return newErrors
    })
    setTouchedState((prev) => {
      const newTouched = { ...prev }
      delete newTouched[field]
      return newTouched
    })
  }, [])

  // Get field props helper
  const getFieldProps = useCallback((field: keyof T) => ({
    name: field as string,
    value: values[field],
    onChange: handleChange,
    onBlur: handleBlur,
  }), [values, handleChange, handleBlur])

  return {
    values,
    errors,
    touched,
    isSubmitting,
    isValid,
    isDirty,
    isValidating,
    handleChange,
    handleBlur,
    handleSubmit,
    setFieldValue,
    setFieldError,
    setFieldTouched,
    setValues,
    setErrors,
    reset,
    resetField,
    validateField,
    validateForm,
    getFieldProps,
  }
}

// --- Field Error Component Helper Types ---

export interface FormFieldError {
  field: string
  message: string
}

/**
 * Get all form errors as an array
 */
export function getFormErrors<T extends Record<string, string>>(
  errors: Partial<Record<keyof T, string>>,
  touched: Partial<Record<keyof T, boolean>>
): FormFieldError[] {
  return Object.entries(errors)
    .filter(([field, message]) => message && touched[field as keyof T])
    .map(([field, message]) => ({
      field,
      message: message as string,
    }))
}

/**
 * Check if a field has an error that should be displayed
 */
export function hasVisibleError<T extends Record<string, string>>(
  field: keyof T,
  errors: Partial<Record<keyof T, string>>,
  touched: Partial<Record<keyof T, boolean>>
): boolean {
  return Boolean(errors[field] && touched[field])
}

/**
 * Get the error message for a field (only if touched)
 */
export function getVisibleError<T extends Record<string, string>>(
  field: keyof T,
  errors: Partial<Record<keyof T, string>>,
  touched: Partial<Record<keyof T, boolean>>
): string | null {
  if (errors[field] && touched[field]) {
    return errors[field] as string
  }
  return null
}
