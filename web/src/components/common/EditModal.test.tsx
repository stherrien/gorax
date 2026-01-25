import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { EditModal, EditField } from './EditModal'

describe('EditModal', () => {
  const mockOnSave = vi.fn()
  const mockOnCancel = vi.fn()

  const defaultFields: EditField[] = [
    { name: 'title', label: 'Title', type: 'text', required: true },
    { name: 'description', label: 'Description', type: 'textarea' },
    { name: 'status', label: 'Status', type: 'select', options: [
      { value: 'active', label: 'Active' },
      { value: 'inactive', label: 'Inactive' },
    ]},
    { name: 'enabled', label: 'Enabled', type: 'checkbox' },
  ]

  const defaultInitialValues = {
    title: 'Test Task',
    description: 'This is a test description',
    status: 'active',
    enabled: true,
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Rendering', () => {
    it('should render modal when isOpen is true', () => {
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      expect(screen.getByRole('dialog')).toBeInTheDocument()
      expect(screen.getByText('Edit Task')).toBeInTheDocument()
    })

    it('should not render modal when isOpen is false', () => {
      render(
        <EditModal
          isOpen={false}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })

    it('should display item ID when provided', () => {
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
          itemId="task-123"
        />
      )

      expect(screen.getByText(/ID: task-123/)).toBeInTheDocument()
    })

    it('should render all form fields with correct labels', () => {
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      expect(screen.getByLabelText(/Title/)).toBeInTheDocument()
      expect(screen.getByLabelText(/Description/)).toBeInTheDocument()
      expect(screen.getByLabelText(/Status/)).toBeInTheDocument()
      expect(screen.getByRole('switch', { name: /Enabled/i })).toBeInTheDocument()
    })

    it('should pre-populate fields with initial values', () => {
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      expect(screen.getByLabelText(/Title/)).toHaveValue('Test Task')
      expect(screen.getByLabelText(/Description/)).toHaveValue('This is a test description')
      expect(screen.getByLabelText(/Status/)).toHaveValue('active')
      expect(screen.getByRole('switch', { name: /Enabled/i })).toBeChecked()
    })

    it('should display required field indicator', () => {
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const titleLabel = screen.getByText('Title')
      expect(titleLabel.parentElement).toHaveTextContent('*')
    })

    it('should render cancel and save buttons', () => {
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument()
    })

    it('should render close button in header', () => {
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      expect(screen.getByLabelText(/close/i)).toBeInTheDocument()
    })
  })

  describe('Field Types', () => {
    it('should render text input field', () => {
      const fields: EditField[] = [
        { name: 'name', label: 'Name', type: 'text', placeholder: 'Enter name' },
      ]

      render(
        <EditModal
          isOpen={true}
          title="Edit"
          fields={fields}
          initialValues={{ name: 'Test' }}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const input = screen.getByLabelText(/Name/)
      expect(input).toHaveAttribute('type', 'text')
      expect(input).toHaveAttribute('placeholder', 'Enter name')
    })

    it('should render textarea field', () => {
      const fields: EditField[] = [
        { name: 'description', label: 'Description', type: 'textarea', placeholder: 'Enter description' },
      ]

      render(
        <EditModal
          isOpen={true}
          title="Edit"
          fields={fields}
          initialValues={{ description: 'Test desc' }}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const textarea = screen.getByLabelText(/Description/)
      expect(textarea.tagName).toBe('TEXTAREA')
      expect(textarea).toHaveAttribute('placeholder', 'Enter description')
    })

    it('should render select field with options', () => {
      const fields: EditField[] = [
        { name: 'priority', label: 'Priority', type: 'select', options: [
          { value: 'low', label: 'Low' },
          { value: 'high', label: 'High' },
        ]},
      ]

      render(
        <EditModal
          isOpen={true}
          title="Edit"
          fields={fields}
          initialValues={{ priority: 'low' }}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const select = screen.getByLabelText(/Priority/)
      expect(select.tagName).toBe('SELECT')
      expect(screen.getByText('Low')).toBeInTheDocument()
      expect(screen.getByText('High')).toBeInTheDocument()
    })

    it('should render checkbox (switch) field', () => {
      const fields: EditField[] = [
        { name: 'active', label: 'Active', type: 'checkbox' },
      ]

      render(
        <EditModal
          isOpen={true}
          title="Edit"
          fields={fields}
          initialValues={{ active: true }}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const switchBtn = screen.getByRole('switch')
      expect(switchBtn).toBeChecked()
    })

    it('should render number input field', () => {
      const fields: EditField[] = [
        { name: 'quantity', label: 'Quantity', type: 'number', min: 1, max: 100 },
      ]

      render(
        <EditModal
          isOpen={true}
          title="Edit"
          fields={fields}
          initialValues={{ quantity: 50 }}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const input = screen.getByLabelText(/Quantity/)
      expect(input).toHaveAttribute('type', 'number')
      expect(input).toHaveAttribute('min', '1')
      expect(input).toHaveAttribute('max', '100')
    })
  })

  describe('Form Editing', () => {
    it('should allow editing text field', async () => {
      const user = userEvent.setup()
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const titleInput = screen.getByLabelText(/Title/)
      await user.clear(titleInput)
      await user.type(titleInput, 'New Title')

      expect(titleInput).toHaveValue('New Title')
    })

    it('should allow editing textarea field', async () => {
      const user = userEvent.setup()
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const descInput = screen.getByLabelText(/Description/)
      await user.clear(descInput)
      await user.type(descInput, 'New description')

      expect(descInput).toHaveValue('New description')
    })

    it('should allow changing select value', async () => {
      const user = userEvent.setup()
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const selectInput = screen.getByLabelText(/Status/)
      await user.selectOptions(selectInput, 'inactive')

      expect(selectInput).toHaveValue('inactive')
    })

    it('should allow toggling checkbox', async () => {
      const user = userEvent.setup()
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const switchBtn = screen.getByRole('switch')
      expect(switchBtn).toBeChecked()

      await user.click(switchBtn)
      expect(switchBtn).not.toBeChecked()

      await user.click(switchBtn)
      expect(switchBtn).toBeChecked()
    })

    it('should display character count for fields with maxLength', () => {
      const fields: EditField[] = [
        { name: 'title', label: 'Title', type: 'text', maxLength: 50 },
      ]

      render(
        <EditModal
          isOpen={true}
          title="Edit"
          fields={fields}
          initialValues={{ title: 'Test' }}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      expect(screen.getByText('4/50')).toBeInTheDocument()
    })
  })

  describe('Validation', () => {
    it('should show error for empty required field', async () => {
      const user = userEvent.setup()
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const titleInput = screen.getByLabelText(/Title/)
      await user.clear(titleInput)

      const saveButton = screen.getByRole('button', { name: /save/i })
      await user.click(saveButton)

      expect(screen.getByText('Title is required')).toBeInTheDocument()
      expect(mockOnSave).not.toHaveBeenCalled()
    })

    it('should enforce maxLength by displaying character count', () => {
      const fields: EditField[] = [
        { name: 'title', label: 'Title', type: 'text', maxLength: 10 },
      ]

      render(
        <EditModal
          isOpen={true}
          title="Edit"
          fields={fields}
          initialValues={{ title: 'Test' }}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      // Character count should be displayed
      expect(screen.getByText('4/10')).toBeInTheDocument()
    })

    it('should pass validation when number is within min/max constraints', async () => {
      const user = userEvent.setup()
      mockOnSave.mockResolvedValueOnce(undefined)
      const fields: EditField[] = [
        { name: 'quantity', label: 'Quantity', type: 'number', min: 1, max: 10, required: true },
      ]

      render(
        <EditModal
          isOpen={true}
          title="Edit"
          fields={fields}
          initialValues={{ quantity: 5 }}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const saveButton = screen.getByRole('button', { name: /save/i })
      await user.click(saveButton)

      // Valid value should allow save
      await waitFor(() => {
        expect(mockOnSave).toHaveBeenCalledWith({ quantity: 5 })
      })
    })

    it('should render number input with min and max attributes for browser validation', () => {
      const fields: EditField[] = [
        { name: 'quantity', label: 'Quantity', type: 'number', min: 1, max: 100 },
      ]

      render(
        <EditModal
          isOpen={true}
          title="Edit"
          fields={fields}
          initialValues={{ quantity: 50 }}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const input = screen.getByLabelText(/Quantity/)
      expect(input).toHaveAttribute('min', '1')
      expect(input).toHaveAttribute('max', '100')
      expect(input).toHaveAttribute('type', 'number')
    })

    it('should clear field error when field is edited', async () => {
      const user = userEvent.setup()
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const titleInput = screen.getByLabelText(/Title/)
      await user.clear(titleInput)

      const saveButton = screen.getByRole('button', { name: /save/i })
      await user.click(saveButton)

      expect(screen.getByText('Title is required')).toBeInTheDocument()

      await user.type(titleInput, 'New Title')

      expect(screen.queryByText('Title is required')).not.toBeInTheDocument()
    })
  })

  describe('Save Functionality', () => {
    it('should call onSave with form data when valid', async () => {
      const user = userEvent.setup()
      mockOnSave.mockResolvedValueOnce(undefined)

      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const saveButton = screen.getByRole('button', { name: /save/i })
      await user.click(saveButton)

      await waitFor(() => {
        expect(mockOnSave).toHaveBeenCalledWith(defaultInitialValues)
      })
    })

    it('should call onSave with updated values after editing', async () => {
      const user = userEvent.setup()
      mockOnSave.mockResolvedValueOnce(undefined)

      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const titleInput = screen.getByLabelText(/Title/)
      await user.clear(titleInput)
      await user.type(titleInput, 'Updated Title')

      const saveButton = screen.getByRole('button', { name: /save/i })
      await user.click(saveButton)

      await waitFor(() => {
        expect(mockOnSave).toHaveBeenCalledWith(expect.objectContaining({
          title: 'Updated Title',
        }))
      })
    })

    it('should show loading state during save', async () => {
      const user = userEvent.setup()
      mockOnSave.mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      )

      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const saveButton = screen.getByRole('button', { name: /save/i })
      await user.click(saveButton)

      expect(screen.getByRole('button', { name: /saving/i })).toBeInTheDocument()
      expect(saveButton).toBeDisabled()

      await waitFor(() => {
        expect(screen.queryByRole('button', { name: /saving/i })).not.toBeInTheDocument()
      })
    })

    it('should disable all inputs during save', async () => {
      const user = userEvent.setup()
      mockOnSave.mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      )

      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const saveButton = screen.getByRole('button', { name: /save/i })
      await user.click(saveButton)

      expect(screen.getByLabelText(/Title/)).toBeDisabled()
      expect(screen.getByLabelText(/Description/)).toBeDisabled()
      expect(screen.getByRole('button', { name: /cancel/i })).toBeDisabled()
    })

    it('should display error message on save failure', async () => {
      const user = userEvent.setup()
      mockOnSave.mockRejectedValueOnce(new Error('Network error'))

      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const saveButton = screen.getByRole('button', { name: /save/i })
      await user.click(saveButton)

      await waitFor(() => {
        expect(screen.getByText('Network error')).toBeInTheDocument()
      })
    })

    it('should display generic error message when error has no message', async () => {
      const user = userEvent.setup()
      mockOnSave.mockRejectedValueOnce('Unknown error')

      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const saveButton = screen.getByRole('button', { name: /save/i })
      await user.click(saveButton)

      await waitFor(() => {
        expect(screen.getByText('Save failed')).toBeInTheDocument()
      })
    })

    it('should re-enable buttons after save failure', async () => {
      const user = userEvent.setup()
      mockOnSave.mockRejectedValueOnce(new Error('Save failed'))

      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const saveButton = screen.getByRole('button', { name: /save/i })
      await user.click(saveButton)

      await waitFor(() => {
        expect(screen.getByText('Save failed')).toBeInTheDocument()
      })

      expect(screen.getByRole('button', { name: /save/i })).not.toBeDisabled()
      expect(screen.getByRole('button', { name: /cancel/i })).not.toBeDisabled()
    })
  })

  describe('Modal Close Behavior', () => {
    it('should call onCancel when clicking cancel button', async () => {
      const user = userEvent.setup()
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const cancelButton = screen.getByRole('button', { name: /cancel/i })
      await user.click(cancelButton)

      expect(mockOnCancel).toHaveBeenCalled()
    })

    it('should call onCancel when clicking close button', async () => {
      const user = userEvent.setup()
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const closeButton = screen.getByLabelText(/close/i)
      await user.click(closeButton)

      expect(mockOnCancel).toHaveBeenCalled()
    })

    it('should call onCancel when clicking backdrop', async () => {
      const user = userEvent.setup()
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const backdrop = screen.getByTestId('edit-modal-backdrop')
      await user.click(backdrop)

      expect(mockOnCancel).toHaveBeenCalled()
    })

    it('should not close modal when clicking inside modal content', async () => {
      const user = userEvent.setup()
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const modal = screen.getByRole('dialog')
      await user.click(modal)

      expect(mockOnCancel).not.toHaveBeenCalled()
    })

    it('should not close modal when clicking backdrop during save', async () => {
      const user = userEvent.setup()
      mockOnSave.mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      )

      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const saveButton = screen.getByRole('button', { name: /save/i })
      await user.click(saveButton)

      const backdrop = screen.getByTestId('edit-modal-backdrop')
      await user.click(backdrop)

      expect(mockOnCancel).not.toHaveBeenCalled()
    })

    it('should call onCancel when pressing Escape key', async () => {
      const user = userEvent.setup()
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const backdrop = screen.getByTestId('edit-modal-backdrop')
      backdrop.focus()
      await user.keyboard('{Escape}')

      expect(mockOnCancel).toHaveBeenCalled()
    })

    it('should not close when pressing Escape during save', async () => {
      const user = userEvent.setup()
      mockOnSave.mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      )

      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const saveButton = screen.getByRole('button', { name: /save/i })
      await user.click(saveButton)

      const backdrop = screen.getByTestId('edit-modal-backdrop')
      backdrop.focus()
      await user.keyboard('{Escape}')

      expect(mockOnCancel).not.toHaveBeenCalled()
    })
  })

  describe('Form Reset Behavior', () => {
    it('should reset form when modal reopens with new values', async () => {
      const { rerender } = render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      expect(screen.getByLabelText(/Title/)).toHaveValue('Test Task')

      // Close modal
      rerender(
        <EditModal
          isOpen={false}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      // Reopen with new values
      rerender(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={{ ...defaultInitialValues, title: 'New Task' }}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      expect(screen.getByLabelText(/Title/)).toHaveValue('New Task')
    })

    it('should clear errors when modal reopens', async () => {
      const user = userEvent.setup()
      const { rerender } = render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      // Trigger validation error
      const titleInput = screen.getByLabelText(/Title/)
      await user.clear(titleInput)
      const saveButton = screen.getByRole('button', { name: /save/i })
      await user.click(saveButton)

      expect(screen.getByText('Title is required')).toBeInTheDocument()

      // Close and reopen modal
      rerender(
        <EditModal
          isOpen={false}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      rerender(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      expect(screen.queryByText('Title is required')).not.toBeInTheDocument()
    })
  })

  describe('Accessibility', () => {
    it('should have proper ARIA attributes', () => {
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const dialog = screen.getByRole('dialog')
      expect(dialog).toHaveAttribute('aria-modal', 'true')
      expect(dialog).toHaveAttribute('aria-labelledby')
    })

    it('should have proper labels for all form fields', () => {
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      expect(screen.getByLabelText(/Title/)).toBeInTheDocument()
      expect(screen.getByLabelText(/Description/)).toBeInTheDocument()
      expect(screen.getByLabelText(/Status/)).toBeInTheDocument()
    })

    it('should have close button with aria-label', () => {
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const closeButton = screen.getByLabelText(/close/i)
      expect(closeButton).toHaveAttribute('aria-label')
    })

    it('should have proper switch role for checkbox fields', () => {
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const switchElement = screen.getByRole('switch')
      expect(switchElement).toHaveAttribute('aria-checked', 'true')
    })
  })

  describe('Dark Theme Styling', () => {
    it('should use dark background colors', () => {
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const modal = screen.getByRole('dialog')
      expect(modal).toHaveClass('bg-gray-800')
    })

    it('should use white text for title', () => {
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const title = screen.getByText('Edit Task')
      expect(title).toHaveClass('text-white')
    })

    it('should use dark theme for input fields', () => {
      render(
        <EditModal
          isOpen={true}
          title="Edit Task"
          fields={defaultFields}
          initialValues={defaultInitialValues}
          onSave={mockOnSave}
          onCancel={mockOnCancel}
        />
      )

      const titleInput = screen.getByLabelText(/Title/)
      expect(titleInput).toHaveClass('bg-gray-700')
      expect(titleInput).toHaveClass('text-white')
    })
  })
})
