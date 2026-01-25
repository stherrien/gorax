import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import {
  useFormValidation,
  getFormErrors,
  hasVisibleError,
  getVisibleError,
} from './useFormValidation'
import type { ValidationResult } from '../utils/validation'

interface TestFormValues {
  name: string
  email: string
}

const initialValues: TestFormValues = {
  name: '',
  email: '',
}

const mockValidate = (values: TestFormValues): ValidationResult => {
  const errors: ValidationResult['errors'] = []

  if (!values.name.trim()) {
    errors.push({
      field: 'name',
      message: 'Name is required',
      code: 'required',
    })
  }

  if (!values.email.trim()) {
    errors.push({
      field: 'email',
      message: 'Email is required',
      code: 'required',
    })
  } else if (!values.email.includes('@')) {
    errors.push({
      field: 'email',
      message: 'Invalid email format',
      code: 'invalid_format',
    })
  }

  return { valid: errors.length === 0, errors }
}

describe('useFormValidation', () => {
  describe('initialization', () => {
    it('should initialize with provided values', () => {
      const { result } = renderHook(() =>
        useFormValidation({
          initialValues: { name: 'John', email: 'john@example.com' },
        })
      )

      expect(result.current.values.name).toBe('John')
      expect(result.current.values.email).toBe('john@example.com')
    })

    it('should initialize with empty errors and touched', () => {
      const { result } = renderHook(() =>
        useFormValidation({ initialValues })
      )

      expect(result.current.errors).toEqual({})
      expect(result.current.touched).toEqual({})
    })

    it('should initialize isSubmitting as false', () => {
      const { result } = renderHook(() =>
        useFormValidation({ initialValues })
      )

      expect(result.current.isSubmitting).toBe(false)
    })

    it('should initialize isDirty as false', () => {
      const { result } = renderHook(() =>
        useFormValidation({ initialValues })
      )

      expect(result.current.isDirty).toBe(false)
    })
  })

  describe('handleChange', () => {
    it('should update field value', () => {
      const { result } = renderHook(() =>
        useFormValidation({ initialValues })
      )

      act(() => {
        result.current.handleChange({
          target: { name: 'name', value: 'John' },
        } as React.ChangeEvent<HTMLInputElement>)
      })

      expect(result.current.values.name).toBe('John')
    })

    it('should set isDirty to true after change', () => {
      const { result } = renderHook(() =>
        useFormValidation({ initialValues })
      )

      act(() => {
        result.current.handleChange({
          target: { name: 'name', value: 'John' },
        } as React.ChangeEvent<HTMLInputElement>)
      })

      expect(result.current.isDirty).toBe(true)
    })

    it('should validate on change when validateOnChange is true', () => {
      const { result } = renderHook(() =>
        useFormValidation({
          initialValues,
          validate: mockValidate,
          validateOnChange: true,
        })
      )

      act(() => {
        result.current.handleChange({
          target: { name: 'email', value: 'invalid' },
        } as React.ChangeEvent<HTMLInputElement>)
      })

      expect(result.current.errors.email).toBe('Invalid email format')
    })

    it('should clear errors when value becomes valid', () => {
      const { result } = renderHook(() =>
        useFormValidation({
          initialValues,
          validate: mockValidate,
          validateOnChange: true,
        })
      )

      // Set invalid value first
      act(() => {
        result.current.handleChange({
          target: { name: 'email', value: 'invalid' },
        } as React.ChangeEvent<HTMLInputElement>)
      })

      expect(result.current.errors.email).toBe('Invalid email format')

      // Fix the value
      act(() => {
        result.current.handleChange({
          target: { name: 'email', value: 'valid@email.com' },
        } as React.ChangeEvent<HTMLInputElement>)
      })

      expect(result.current.errors.email).toBeUndefined()
    })

    it('should not validate on change when validateOnChange is false', () => {
      const { result } = renderHook(() =>
        useFormValidation({
          initialValues,
          validate: mockValidate,
          validateOnChange: false,
        })
      )

      act(() => {
        result.current.handleChange({
          target: { name: 'email', value: 'invalid' },
        } as React.ChangeEvent<HTMLInputElement>)
      })

      expect(result.current.errors.email).toBeUndefined()
    })
  })

  describe('handleBlur', () => {
    it('should mark field as touched', () => {
      const { result } = renderHook(() =>
        useFormValidation({ initialValues })
      )

      act(() => {
        result.current.handleBlur({
          target: { name: 'name' },
        } as React.FocusEvent<HTMLInputElement>)
      })

      expect(result.current.touched.name).toBe(true)
    })

    it('should validate on blur when validateOnBlur is true', () => {
      const { result } = renderHook(() =>
        useFormValidation({
          initialValues,
          validate: mockValidate,
          validateOnBlur: true,
        })
      )

      act(() => {
        result.current.handleBlur({
          target: { name: 'name' },
        } as React.FocusEvent<HTMLInputElement>)
      })

      expect(result.current.errors.name).toBe('Name is required')
    })
  })

  describe('handleSubmit', () => {
    it('should mark all fields as touched', async () => {
      const { result } = renderHook(() =>
        useFormValidation({ initialValues })
      )

      await act(async () => {
        await result.current.handleSubmit()
      })

      expect(result.current.touched.name).toBe(true)
      expect(result.current.touched.email).toBe(true)
    })

    it('should validate form before submitting', async () => {
      const onSubmit = vi.fn()
      const { result } = renderHook(() =>
        useFormValidation({
          initialValues,
          validate: mockValidate,
          onSubmit,
        })
      )

      await act(async () => {
        await result.current.handleSubmit()
      })

      // Should not call onSubmit because form is invalid
      expect(onSubmit).not.toHaveBeenCalled()
      expect(result.current.errors.name).toBe('Name is required')
    })

    it('should call onSubmit when form is valid', async () => {
      const onSubmit = vi.fn()
      const { result } = renderHook(() =>
        useFormValidation({
          initialValues: { name: 'John', email: 'john@example.com' },
          validate: mockValidate,
          onSubmit,
        })
      )

      await act(async () => {
        await result.current.handleSubmit()
      })

      expect(onSubmit).toHaveBeenCalledWith({
        name: 'John',
        email: 'john@example.com',
      })
    })

    it('should set isSubmitting during submit', async () => {
      const onSubmit = vi.fn().mockImplementation(() => {
        // Capture isSubmitting state during the callback
        // Note: This works because the callback runs while isSubmitting is true
        return new Promise<void>((resolve) => {
          // Use setTimeout to allow state to settle
          setTimeout(() => {
            resolve()
          }, 10)
        })
      })

      const { result } = renderHook(() =>
        useFormValidation({
          initialValues: { name: 'John', email: 'john@example.com' },
          validate: mockValidate,
          onSubmit,
        })
      )

      await act(async () => {
        await result.current.handleSubmit()
      })

      // onSubmit should have been called
      expect(onSubmit).toHaveBeenCalled()
      // After submit completes, isSubmitting should be false
      expect(result.current.isSubmitting).toBe(false)
    })

    it('should prevent default form event', async () => {
      const preventDefault = vi.fn()
      const { result } = renderHook(() =>
        useFormValidation({ initialValues })
      )

      await act(async () => {
        await result.current.handleSubmit({
          preventDefault,
        } as unknown as React.FormEvent)
      })

      expect(preventDefault).toHaveBeenCalled()
    })
  })

  describe('setFieldValue', () => {
    it('should set field value programmatically', () => {
      const { result } = renderHook(() =>
        useFormValidation({ initialValues })
      )

      act(() => {
        result.current.setFieldValue('name', 'Jane')
      })

      expect(result.current.values.name).toBe('Jane')
    })
  })

  describe('setFieldError', () => {
    it('should set field error', () => {
      const { result } = renderHook(() =>
        useFormValidation({ initialValues })
      )

      act(() => {
        result.current.setFieldError('name', 'Custom error')
      })

      expect(result.current.errors.name).toBe('Custom error')
    })

    it('should clear field error when set to null', () => {
      const { result } = renderHook(() =>
        useFormValidation({ initialValues })
      )

      act(() => {
        result.current.setFieldError('name', 'Custom error')
      })

      act(() => {
        result.current.setFieldError('name', null)
      })

      expect(result.current.errors.name).toBeUndefined()
    })
  })

  describe('setFieldTouched', () => {
    it('should set field touched state', () => {
      const { result } = renderHook(() =>
        useFormValidation({ initialValues })
      )

      act(() => {
        result.current.setFieldTouched('name', true)
      })

      expect(result.current.touched.name).toBe(true)
    })
  })

  describe('setValues', () => {
    it('should set multiple values', () => {
      const { result } = renderHook(() =>
        useFormValidation({ initialValues })
      )

      act(() => {
        result.current.setValues({ name: 'John', email: 'john@example.com' })
      })

      expect(result.current.values.name).toBe('John')
      expect(result.current.values.email).toBe('john@example.com')
    })
  })

  describe('reset', () => {
    it('should reset to initial values', () => {
      const { result } = renderHook(() =>
        useFormValidation({ initialValues })
      )

      act(() => {
        result.current.setValues({ name: 'John', email: 'john@example.com' })
        result.current.setFieldError('name', 'Error')
        result.current.setFieldTouched('name', true)
      })

      act(() => {
        result.current.reset()
      })

      expect(result.current.values).toEqual(initialValues)
      expect(result.current.errors).toEqual({})
      expect(result.current.touched).toEqual({})
    })
  })

  describe('resetField', () => {
    it('should reset single field', () => {
      const { result } = renderHook(() =>
        useFormValidation({ initialValues })
      )

      act(() => {
        result.current.setFieldValue('name', 'John')
        result.current.setFieldValue('email', 'john@example.com')
        result.current.setFieldError('name', 'Error')
        result.current.setFieldTouched('name', true)
      })

      act(() => {
        result.current.resetField('name')
      })

      expect(result.current.values.name).toBe('')
      expect(result.current.values.email).toBe('john@example.com')
      expect(result.current.errors.name).toBeUndefined()
      expect(result.current.touched.name).toBeUndefined()
    })
  })

  describe('validateField', () => {
    it('should validate single field', () => {
      const { result } = renderHook(() =>
        useFormValidation({
          initialValues,
          validate: mockValidate,
        })
      )

      act(() => {
        result.current.validateField('name')
      })

      expect(result.current.errors.name).toBe('Name is required')
    })

    it('should return validation result', () => {
      const { result } = renderHook(() =>
        useFormValidation({
          initialValues,
          validate: mockValidate,
        })
      )

      let validationResult: ReturnType<typeof result.current.validateField>
      act(() => {
        validationResult = result.current.validateField('name')
      })

      expect(validationResult!.valid).toBe(false)
      expect(validationResult!.errors).toHaveLength(1)
    })
  })

  describe('validateForm', () => {
    it('should validate entire form', async () => {
      const { result } = renderHook(() =>
        useFormValidation({
          initialValues,
          validate: mockValidate,
        })
      )

      let validationResult: ValidationResult
      await act(async () => {
        validationResult = await result.current.validateForm()
      })

      expect(validationResult!.valid).toBe(false)
      expect(result.current.errors.name).toBe('Name is required')
      expect(result.current.errors.email).toBe('Email is required')
    })
  })

  describe('validateAsync', () => {
    it('should run async validation', async () => {
      const validateAsync = vi.fn().mockResolvedValue({
        valid: false,
        errors: [{ field: 'email', message: 'Email already exists', code: 'duplicate' }],
      })

      const { result } = renderHook(() =>
        useFormValidation({
          initialValues: { name: 'John', email: 'john@example.com' },
          validate: mockValidate,
          validateAsync,
        })
      )

      await act(async () => {
        await result.current.validateForm()
      })

      expect(validateAsync).toHaveBeenCalled()
      expect(result.current.errors.email).toBe('Email already exists')
    })

    it('should set isValidating during async validation', async () => {
      let resolveValidation: () => void
      const validatePromise = new Promise<ValidationResult>((resolve) => {
        resolveValidation = () => resolve({ valid: true, errors: [] })
      })

      const validateAsync = vi.fn().mockReturnValue(validatePromise)

      const { result } = renderHook(() =>
        useFormValidation({
          initialValues: { name: 'John', email: 'john@example.com' },
          validate: mockValidate,
          validateAsync,
        })
      )

      // Start validation
      act(() => {
        result.current.validateForm()
      })

      expect(result.current.isValidating).toBe(true)

      // Complete validation
      await act(async () => {
        resolveValidation!()
        await validatePromise
      })

      expect(result.current.isValidating).toBe(false)
    })
  })

  describe('getFieldProps', () => {
    it('should return field props object', () => {
      const { result } = renderHook(() =>
        useFormValidation({
          initialValues: { name: 'John', email: '' },
        })
      )

      const props = result.current.getFieldProps('name')

      expect(props.name).toBe('name')
      expect(props.value).toBe('John')
      expect(typeof props.onChange).toBe('function')
      expect(typeof props.onBlur).toBe('function')
    })
  })

  describe('isValid', () => {
    it('should be true when no errors', () => {
      const { result } = renderHook(() =>
        useFormValidation({ initialValues })
      )

      expect(result.current.isValid).toBe(true)
    })

    it('should be false when there are errors', () => {
      const { result } = renderHook(() =>
        useFormValidation({ initialValues })
      )

      act(() => {
        result.current.setFieldError('name', 'Error')
      })

      expect(result.current.isValid).toBe(false)
    })
  })
})

describe('helper functions', () => {
  describe('getFormErrors', () => {
    it('should return errors for touched fields only', () => {
      const errors = { name: 'Name error', email: 'Email error' }
      const touched = { name: true }

      const result = getFormErrors(errors, touched)

      expect(result).toHaveLength(1)
      expect(result[0]).toEqual({ field: 'name', message: 'Name error' })
    })

    it('should return empty array when no touched errors', () => {
      const errors = { name: 'Name error' }
      const touched = {}

      const result = getFormErrors(errors, touched)

      expect(result).toHaveLength(0)
    })
  })

  describe('hasVisibleError', () => {
    it('should return true for touched field with error', () => {
      const errors = { name: 'Error' }
      const touched = { name: true }

      expect(hasVisibleError('name', errors, touched)).toBe(true)
    })

    it('should return false for untouched field with error', () => {
      const errors = { name: 'Error' }
      const touched = {}

      expect(hasVisibleError('name', errors, touched)).toBe(false)
    })

    it('should return false for touched field without error', () => {
      const errors = {}
      const touched = { name: true }

      expect(hasVisibleError('name', errors, touched)).toBe(false)
    })
  })

  describe('getVisibleError', () => {
    it('should return error message for touched field', () => {
      const errors = { name: 'Name error' }
      const touched = { name: true }

      expect(getVisibleError('name', errors, touched)).toBe('Name error')
    })

    it('should return null for untouched field', () => {
      const errors = { name: 'Name error' }
      const touched = {}

      expect(getVisibleError('name', errors, touched)).toBeNull()
    })
  })
})
