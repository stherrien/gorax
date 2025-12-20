import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import userEvent from '@testing-library/user-event'
import CreateSchedule from './CreateSchedule'
import { scheduleAPI } from '../api/schedules'

const mockNavigate = vi.fn()

vi.mock('../api/schedules', () => ({
  scheduleAPI: {
    create: vi.fn(),
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

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useSearchParams: () => [new URLSearchParams({ workflowId: 'workflow-123' })],
  }
})

describe('CreateSchedule', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  const renderWithRouter = (component: React.ReactElement) => {
    return render(<BrowserRouter>{component}</BrowserRouter>)
  }

  describe('Page Render', () => {
    it('should render page title', () => {
      renderWithRouter(<CreateSchedule />)
      expect(screen.getByRole('heading', { name: /create schedule/i })).toBeInTheDocument()
    })

    it('should render ScheduleForm component', () => {
      renderWithRouter(<CreateSchedule />)
      expect(screen.getByLabelText(/schedule name/i)).toBeInTheDocument()
      expect(screen.getByText(/cron expression/i)).toBeInTheDocument()
    })
  })

  describe('Schedule Creation', () => {
    it('should create schedule and navigate on success', async () => {
      const user = userEvent.setup()
      const mockSchedule = {
        id: 'schedule-123',
        name: 'Test Schedule',
        cronExpression: '0 9 * * *',
        timezone: 'UTC',
        enabled: true,
      }

      vi.mocked(scheduleAPI.create).mockResolvedValue(mockSchedule as any)

      renderWithRouter(<CreateSchedule />)

      const nameInput = screen.getByLabelText(/schedule name/i)
      await user.type(nameInput, 'Test Schedule')

      const submitButton = screen.getByRole('button', { name: /create schedule/i })
      await user.click(submitButton)

      await waitFor(() => {
        expect(scheduleAPI.create).toHaveBeenCalledWith('workflow-123', {
          name: 'Test Schedule',
          cronExpression: '0 9 * * *',
          timezone: 'UTC',
          enabled: true,
        })
      })

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/schedules')
      })
    })

    it('should show error message on creation failure', async () => {
      const user = userEvent.setup()
      const errorMessage = 'Failed to create schedule'

      vi.mocked(scheduleAPI.create).mockRejectedValue(new Error(errorMessage))

      renderWithRouter(<CreateSchedule />)

      const nameInput = screen.getByLabelText(/schedule name/i)
      await user.type(nameInput, 'Test Schedule')

      const submitButton = screen.getByRole('button', { name: /create schedule/i })
      await user.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText(errorMessage)).toBeInTheDocument()
      })

      expect(mockNavigate).not.toHaveBeenCalled()
    })
  })

  describe('Navigation', () => {
    it('should navigate back to schedules on cancel', async () => {
      const user = userEvent.setup()

      renderWithRouter(<CreateSchedule />)

      const cancelButton = screen.getByRole('button', { name: /cancel/i })
      await user.click(cancelButton)

      expect(mockNavigate).toHaveBeenCalledWith('/schedules')
    })
  })
})
