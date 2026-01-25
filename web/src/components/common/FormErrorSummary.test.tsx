/**
 * FormErrorSummary Component Tests
 */

import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { FormErrorSummary, FieldError } from './FormErrorSummary'
import type { FormFieldError } from '../../hooks/useFormValidation'

describe('FormErrorSummary', () => {
  const mockErrors: FormFieldError[] = [
    { field: 'name', message: 'Name is required' },
    { field: 'email', message: 'Invalid email format' },
    { field: 'cron_expression', message: 'Invalid cron expression' },
  ]

  describe('rendering', () => {
    it('should render nothing when errors array is empty', () => {
      const { container } = render(<FormErrorSummary errors={[]} />)
      expect(container.firstChild).toBeNull()
    })

    it('should render error summary when errors exist', () => {
      render(<FormErrorSummary errors={mockErrors} />)

      expect(screen.getByTestId('form-error-summary')).toBeInTheDocument()
      expect(screen.getByText('Please fix the following errors:')).toBeInTheDocument()
    })

    it('should render all error messages', () => {
      render(<FormErrorSummary errors={mockErrors} />)

      expect(screen.getByText(/Name is required/)).toBeInTheDocument()
      expect(screen.getByText(/Invalid email format/)).toBeInTheDocument()
      expect(screen.getByText(/Invalid cron expression/)).toBeInTheDocument()
    })

    it('should render error items with data-testid', () => {
      render(<FormErrorSummary errors={mockErrors} />)

      expect(screen.getByTestId('error-item-name')).toBeInTheDocument()
      expect(screen.getByTestId('error-item-email')).toBeInTheDocument()
      expect(screen.getByTestId('error-item-cron_expression')).toBeInTheDocument()
    })

    it('should use custom title when provided', () => {
      render(<FormErrorSummary errors={mockErrors} title="Validation errors:" />)

      expect(screen.getByText('Validation errors:')).toBeInTheDocument()
      expect(screen.queryByText('Please fix the following errors:')).not.toBeInTheDocument()
    })

    it('should apply custom className', () => {
      render(<FormErrorSummary errors={mockErrors} className="custom-class" />)

      const summary = screen.getByTestId('form-error-summary')
      expect(summary).toHaveClass('custom-class')
    })
  })

  describe('field name formatting', () => {
    it('should format snake_case field names to Title Case', () => {
      const errors: FormFieldError[] = [
        { field: 'cron_expression', message: 'Invalid' },
      ]
      render(<FormErrorSummary errors={errors} />)

      expect(screen.getByText('Cron Expression:')).toBeInTheDocument()
    })

    it('should format camelCase field names to Title Case', () => {
      const errors: FormFieldError[] = [
        { field: 'cronExpression', message: 'Invalid' },
      ]
      render(<FormErrorSummary errors={errors} />)

      expect(screen.getByText('Cron Expression:')).toBeInTheDocument()
    })

    it('should capitalize simple field names', () => {
      const errors: FormFieldError[] = [
        { field: 'email', message: 'Invalid' },
      ]
      render(<FormErrorSummary errors={errors} />)

      expect(screen.getByText('Email:')).toBeInTheDocument()
    })
  })

  describe('accessibility', () => {
    it('should have role="alert"', () => {
      render(<FormErrorSummary errors={mockErrors} />)

      const summary = screen.getByTestId('form-error-summary')
      expect(summary).toHaveAttribute('role', 'alert')
    })

    it('should have aria-live="polite"', () => {
      render(<FormErrorSummary errors={mockErrors} />)

      const summary = screen.getByTestId('form-error-summary')
      expect(summary).toHaveAttribute('aria-live', 'polite')
    })

    it('should have aria-labelledby pointing to title', () => {
      render(<FormErrorSummary errors={mockErrors} />)

      const summary = screen.getByTestId('form-error-summary')
      expect(summary).toHaveAttribute('aria-labelledby', 'error-summary-title')

      const title = screen.getByText('Please fix the following errors:')
      expect(title).toHaveAttribute('id', 'error-summary-title')
    })
  })

  describe('error click handling', () => {
    it('should call onErrorClick when error item is clicked', () => {
      const onErrorClick = vi.fn()
      render(<FormErrorSummary errors={mockErrors} onErrorClick={onErrorClick} />)

      fireEvent.click(screen.getByTestId('error-item-name'))

      expect(onErrorClick).toHaveBeenCalledWith('name')
    })

    it('should make items focusable when onErrorClick is provided', () => {
      const onErrorClick = vi.fn()
      render(<FormErrorSummary errors={mockErrors} onErrorClick={onErrorClick} />)

      const item = screen.getByTestId('error-item-name')
      expect(item).toHaveAttribute('tabIndex', '0')
      expect(item).toHaveAttribute('role', 'button')
    })

    it('should not make items focusable when onErrorClick is not provided', () => {
      render(<FormErrorSummary errors={mockErrors} />)

      const item = screen.getByTestId('error-item-name')
      expect(item).not.toHaveAttribute('tabIndex')
      expect(item).not.toHaveAttribute('role', 'button')
    })

    it('should call onErrorClick on Enter key press', () => {
      const onErrorClick = vi.fn()
      render(<FormErrorSummary errors={mockErrors} onErrorClick={onErrorClick} />)

      const item = screen.getByTestId('error-item-email')
      fireEvent.keyDown(item, { key: 'Enter' })

      expect(onErrorClick).toHaveBeenCalledWith('email')
    })

    it('should call onErrorClick on Space key press', () => {
      const onErrorClick = vi.fn()
      render(<FormErrorSummary errors={mockErrors} onErrorClick={onErrorClick} />)

      const item = screen.getByTestId('error-item-email')
      fireEvent.keyDown(item, { key: ' ' })

      expect(onErrorClick).toHaveBeenCalledWith('email')
    })

    it('should have aria-label for clickable items', () => {
      const onErrorClick = vi.fn()
      render(<FormErrorSummary errors={mockErrors} onErrorClick={onErrorClick} />)

      const item = screen.getByTestId('error-item-name')
      expect(item).toHaveAttribute('aria-label', 'Go to name: Name is required')
    })
  })
})

describe('FieldError', () => {
  describe('rendering', () => {
    it('should render nothing when error is null', () => {
      const { container } = render(<FieldError error={null} />)
      expect(container.firstChild).toBeNull()
    })

    it('should render nothing when error is undefined', () => {
      const { container } = render(<FieldError error={undefined} />)
      expect(container.firstChild).toBeNull()
    })

    it('should render nothing when error is empty string', () => {
      const { container } = render(<FieldError error="" />)
      expect(container.firstChild).toBeNull()
    })

    it('should render error message when provided', () => {
      render(<FieldError error="This field is required" />)

      expect(screen.getByText('This field is required')).toBeInTheDocument()
    })

    it('should apply custom className', () => {
      render(<FieldError error="Error" className="custom-error" />)

      const error = screen.getByText('Error')
      expect(error).toHaveClass('custom-error')
    })

    it('should have role="alert"', () => {
      render(<FieldError error="Error" />)

      const error = screen.getByText('Error')
      expect(error).toHaveAttribute('role', 'alert')
    })
  })

  describe('accessibility', () => {
    it('should set id when fieldId is provided', () => {
      render(<FieldError error="Error" fieldId="email" />)

      const error = screen.getByText('Error')
      expect(error).toHaveAttribute('id', 'email-error')
    })

    it('should set data-testid with fieldId when provided', () => {
      render(<FieldError error="Error" fieldId="email" />)

      expect(screen.getByTestId('field-error-email')).toBeInTheDocument()
    })

    it('should use generic data-testid when fieldId not provided', () => {
      render(<FieldError error="Error" />)

      expect(screen.getByTestId('field-error')).toBeInTheDocument()
    })
  })
})
