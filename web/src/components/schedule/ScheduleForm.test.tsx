import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ScheduleForm, ScheduleFormProps } from './ScheduleForm'

// Mock the scheduleAPI to avoid actual API calls
vi.mock('../../api/schedules', () => ({
  scheduleAPI: {
    preview: vi.fn().mockResolvedValue({
      valid: true,
      next_runs: [
        new Date().toISOString(),
        new Date(Date.now() + 86400000).toISOString(),
      ],
      count: 2,
      timezone: 'UTC',
    }),
  },
}))

describe('ScheduleForm', () => {
  const mockOnSubmit = vi.fn()
  const mockOnCancel = vi.fn()

  const defaultProps: ScheduleFormProps = {
    onSubmit: mockOnSubmit,
    onCancel: mockOnCancel,
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Initial Render', () => {
    it('should render form with all required fields', () => {
      render(<ScheduleForm {...defaultProps} />)

      expect(screen.getByLabelText(/schedule name/i)).toBeInTheDocument()
      expect(screen.getByText(/cron expression/i)).toBeInTheDocument()
      expect(screen.getByText(/timezone/i)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /create schedule/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
    })

    it('should render CronBuilder component', () => {
      render(<ScheduleForm {...defaultProps} />)

      // CronBuilder has presets button
      expect(screen.getByRole('button', { name: /presets/i })).toBeInTheDocument()
    })

    it('should render TimezoneSelector component', () => {
      render(<ScheduleForm {...defaultProps} />)

      expect(screen.getByText(/timezone/i)).toBeInTheDocument()
    })

    it('should render SchedulePreview component', async () => {
      render(<ScheduleForm {...defaultProps} />)

      // SchedulePreview shows when cron expression is set
      await waitFor(() => {
        expect(screen.getByText(/next run times/i)).toBeInTheDocument()
      })
    })

    it('should render enabled checkbox', () => {
      render(<ScheduleForm {...defaultProps} />)

      expect(screen.getByLabelText(/enabled/i)).toBeInTheDocument()
    })
  })

  describe('Form Interaction', () => {
    it('should update schedule name when user types', async () => {
      const user = userEvent.setup()
      render(<ScheduleForm {...defaultProps} />)

      const nameInput = screen.getByLabelText(/schedule name/i)
      await user.type(nameInput, 'Daily Backup')

      expect(nameInput).toHaveValue('Daily Backup')
    })

    it('should update cron expression when user selects preset', async () => {
      const user = userEvent.setup()
      render(<ScheduleForm {...defaultProps} />)

      const presetsButton = screen.getByRole('button', { name: /presets/i })
      await user.click(presetsButton)

      // Select "Every day at midnight" preset
      const presetOption = screen.getByText(/every day at midnight/i)
      await user.click(presetOption)

      // Cron expression should be updated
      expect(screen.getByText(/runs at 00:00 every day/i)).toBeInTheDocument()
    })

    it('should update timezone when user selects from dropdown', async () => {
      const user = userEvent.setup()
      render(<ScheduleForm {...defaultProps} />)

      const timezoneSelect = screen.getByRole('combobox')
      await user.selectOptions(timezoneSelect, 'America/New_York')

      expect(timezoneSelect).toHaveValue('America/New_York')
    })

    it('should toggle enabled checkbox', async () => {
      const user = userEvent.setup()
      render(<ScheduleForm {...defaultProps} />)

      const enabledCheckbox = screen.getByLabelText(/enabled/i) as HTMLInputElement
      expect(enabledCheckbox.checked).toBe(true) // Default enabled

      await user.click(enabledCheckbox)
      expect(enabledCheckbox.checked).toBe(false)

      await user.click(enabledCheckbox)
      expect(enabledCheckbox.checked).toBe(true)
    })
  })

  describe('Form Validation', () => {
    it('should show error when submitting without schedule name', async () => {
      const user = userEvent.setup()
      render(<ScheduleForm {...defaultProps} />)

      const submitButton = screen.getByRole('button', { name: /create schedule/i })
      await user.click(submitButton)

      // Error appears in both FormErrorSummary and individual field error div
      const errors = await screen.findAllByText(/schedule name is required/i)
      expect(errors.length).toBeGreaterThanOrEqual(1)
      expect(mockOnSubmit).not.toHaveBeenCalled()
    })

    it('should show error when cron expression is invalid', async () => {
      const user = userEvent.setup()
      render(<ScheduleForm {...defaultProps} />)

      const nameInput = screen.getByLabelText(/schedule name/i)
      await user.type(nameInput, 'Test Schedule')

      // Switch to advanced mode
      const advancedButton = screen.getByRole('button', { name: /advanced/i })
      await user.click(advancedButton)

      // Enter invalid cron expression
      const cronInput = screen.getByPlaceholderText('* * * * *')
      await user.clear(cronInput)
      await user.type(cronInput, 'invalid cron')

      const submitButton = screen.getByRole('button', { name: /create schedule/i })
      await user.click(submitButton)

      const errors = await screen.findAllByText(/invalid cron expression/i)
      expect(errors.length).toBeGreaterThan(0)
      expect(mockOnSubmit).not.toHaveBeenCalled()
    })

    it('should disable submit button during submission', async () => {
      const user = userEvent.setup()
      const mockSubmitSlow = vi.fn(
        () => new Promise((resolve) => setTimeout(resolve, 1000))
      )

      render(<ScheduleForm {...defaultProps} onSubmit={mockSubmitSlow} />)

      const nameInput = screen.getByLabelText(/schedule name/i)
      await user.type(nameInput, 'Test Schedule')

      const submitButton = screen.getByRole('button', { name: /create schedule/i })
      await user.click(submitButton)

      expect(submitButton).toBeDisabled()
      expect(screen.getByText(/creating.../i)).toBeInTheDocument()
    })
  })

  describe('Form Submission', () => {
    it('should submit form with correct data', async () => {
      const user = userEvent.setup()
      render(<ScheduleForm {...defaultProps} />)

      const nameInput = screen.getByLabelText(/schedule name/i)
      await user.type(nameInput, 'Daily Backup')

      // Select preset
      const presetsButton = screen.getByRole('button', { name: /presets/i })
      await user.click(presetsButton)
      const presetOption = screen.getByText(/every day at midnight/i)
      await user.click(presetOption)

      // Select timezone
      const timezoneSelect = screen.getByRole('combobox')
      await user.selectOptions(timezoneSelect, 'America/New_York')

      const submitButton = screen.getByRole('button', { name: /create schedule/i })
      await user.click(submitButton)

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalledWith({
          name: 'Daily Backup',
          cronExpression: '0 0 * * *',
          timezone: 'America/New_York',
          enabled: true,
        })
      })
    })

    it('should call onCancel when cancel button is clicked', async () => {
      const user = userEvent.setup()
      render(<ScheduleForm {...defaultProps} />)

      const cancelButton = screen.getByRole('button', { name: /cancel/i })
      await user.click(cancelButton)

      expect(mockOnCancel).toHaveBeenCalled()
      expect(mockOnSubmit).not.toHaveBeenCalled()
    })
  })

  describe('Edit Mode', () => {
    const existingSchedule = {
      name: 'Existing Schedule',
      cronExpression: '0 9 * * 1-5',
      timezone: 'Europe/London',
      enabled: false,
    }

    it('should populate form with existing schedule data', () => {
      render(<ScheduleForm {...defaultProps} initialData={existingSchedule} />)

      const nameInput = screen.getByLabelText(/schedule name/i) as HTMLInputElement
      expect(nameInput.value).toBe('Existing Schedule')

      const timezoneSelect = screen.getByRole('combobox') as HTMLSelectElement
      expect(timezoneSelect.value).toBe('Europe/London')

      const enabledCheckbox = screen.getByLabelText(/enabled/i) as HTMLInputElement
      expect(enabledCheckbox.checked).toBe(false)
    })

    it('should show update button text in edit mode', () => {
      render(<ScheduleForm {...defaultProps} initialData={existingSchedule} />)

      expect(screen.getByRole('button', { name: /update schedule/i })).toBeInTheDocument()
      expect(
        screen.queryByRole('button', { name: /create schedule/i })
      ).not.toBeInTheDocument()
    })

    it('should submit updated data', async () => {
      const user = userEvent.setup()
      render(<ScheduleForm {...defaultProps} initialData={existingSchedule} />)

      const nameInput = screen.getByLabelText(/schedule name/i)
      await user.clear(nameInput)
      await user.type(nameInput, 'Updated Schedule')

      const submitButton = screen.getByRole('button', { name: /update schedule/i })
      await user.click(submitButton)

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalledWith({
          name: 'Updated Schedule',
          cronExpression: '0 9 * * 1-5',
          timezone: 'Europe/London',
          enabled: false,
        })
      })
    })
  })

  describe('Integration with CronBuilder', () => {
    it('should update SchedulePreview when cron expression changes', async () => {
      const user = userEvent.setup()
      render(<ScheduleForm {...defaultProps} />)

      // Select a preset
      const presetsButton = screen.getByRole('button', { name: /presets/i })
      await user.click(presetsButton)
      const presetOption = screen.getByText(/every hour/i)
      await user.click(presetOption)

      // SchedulePreview should update (check for description)
      expect(screen.getByText(/runs at minute 0 of every hour/i)).toBeInTheDocument()
    })

    it('should validate cron expression in advanced mode', async () => {
      const user = userEvent.setup()
      render(<ScheduleForm {...defaultProps} />)

      const nameInput = screen.getByLabelText(/schedule name/i)
      await user.type(nameInput, 'Test')

      // Switch to advanced mode
      const advancedButton = screen.getByRole('button', { name: /advanced/i })
      await user.click(advancedButton)

      // Enter valid cron expression
      const cronInput = screen.getByPlaceholderText('* * * * *')
      await user.clear(cronInput)
      await user.type(cronInput, '0 12 * * *')

      expect(screen.getByText(/every day at 12:00 pm/i)).toBeInTheDocument()
    })
  })
})
