import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import Schedules from './Schedules'
import * as useSchedulesHook from '../hooks/useSchedules'
import * as useWorkflowsHook from '../hooks/useWorkflows'

vi.mock('../hooks/useSchedules')
vi.mock('../hooks/useWorkflows')

function renderWithRouter(component: React.ReactElement) {
  return render(<BrowserRouter>{component}</BrowserRouter>)
}

describe('Schedules Page', () => {
  beforeEach(() => {
    vi.clearAllMocks()

    // Always mock useScheduleMutations
    vi.mocked(useSchedulesHook.useScheduleMutations).mockReturnValue({
      toggleSchedule: vi.fn(),
      createSchedule: vi.fn(),
      updateSchedule: vi.fn(),
      deleteSchedule: vi.fn(),
      creating: false,
      updating: false,
      deleting: false,
    })
  })

  const mockSchedules = [
    {
      id: 'sched-1',
      tenantId: 'tenant-1',
      workflowId: 'wf-1',
      name: 'Daily Backup',
      cronExpression: '0 0 * * *',
      timezone: 'UTC',
      enabled: true,
      nextRunAt: '2025-01-20T00:00:00Z',
      createdBy: 'user-1',
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
    },
  ]

  const mockWorkflows = [
    {
      id: 'wf-1',
      tenantId: 'tenant-1',
      name: 'Test Workflow',
      status: 'active' as const,
      definition: { nodes: [], edges: [] },
      version: 1,
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
    },
  ]

  it('should render page header', () => {
    vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
      schedules: [],
      total: 0,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
      workflows: [],
      total: 0,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    renderWithRouter(<Schedules />)

    expect(screen.getByText('Schedules')).toBeInTheDocument()
    expect(screen.getByText('Create Schedule')).toBeInTheDocument()
  })

  it('should render tab navigation', () => {
    vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
      schedules: [],
      total: 0,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
      workflows: [],
      total: 0,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    renderWithRouter(<Schedules />)

    expect(screen.getByText('List View')).toBeInTheDocument()
    expect(screen.getByText('Calendar View')).toBeInTheDocument()
    expect(screen.getByText('Timeline View')).toBeInTheDocument()
  })

  it('should show loading state', () => {
    vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
      schedules: [],
      total: 0,
      loading: true,
      error: null,
      refetch: vi.fn(),
    })

    vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
      workflows: [],
      loading: true,
      error: null,
      refetch: vi.fn(),
    })

    renderWithRouter(<Schedules />)

    expect(screen.getByText(/Loading schedules/i)).toBeInTheDocument()
  })

  it('should show error state', () => {
    const error = new Error('Failed to fetch')

    vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
      schedules: [],
      total: 0,
      loading: false,
      error,
      refetch: vi.fn(),
    })

    vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
      workflows: [],
      total: 0,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    renderWithRouter(<Schedules />)

    expect(screen.getByText(/Failed to fetch schedules/i)).toBeInTheDocument()
  })

  it('should render schedules in list view', () => {
    vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
      schedules: mockSchedules,
      total: 1,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
      workflows: mockWorkflows,
      total: 1,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    renderWithRouter(<Schedules />)

    expect(screen.getByText('Daily Backup')).toBeInTheDocument()
  })

  it('should switch to calendar view', () => {
    vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
      schedules: mockSchedules,
      total: 1,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
      workflows: mockWorkflows,
      total: 1,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    renderWithRouter(<Schedules />)

    const calendarTab = screen.getByText('Calendar View')
    fireEvent.click(calendarTab)

    // Calendar should be visible
    expect(screen.getByText(/January|February|March|April|May|June|July|August|September|October|November|December/)).toBeInTheDocument()
  })

  it('should switch to timeline view', () => {
    vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
      schedules: mockSchedules,
      total: 1,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
      workflows: mockWorkflows,
      total: 1,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    renderWithRouter(<Schedules />)

    const timelineTab = screen.getByText('Timeline View')
    fireEvent.click(timelineTab)

    // Timeline should be visible
    expect(screen.getByText(/Next 24 Hours/i)).toBeInTheDocument()
  })

  it('should filter by status', () => {
    const refetch = vi.fn()

    vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
      schedules: mockSchedules,
      total: 1,
      loading: false,
      error: null,
      refetch,
    })

    vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
      workflows: mockWorkflows,
      total: 1,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    renderWithRouter(<Schedules />)

    const statusFilter = screen.getByLabelText(/status/i)
    fireEvent.change(statusFilter, { target: { value: 'enabled' } })

    // Should trigger refetch with filter
    waitFor(() => {
      expect(refetch).toHaveBeenCalled()
    })
  })

  it('should handle toggle schedule', async () => {
    const refetch = vi.fn()
    const toggleSchedule = vi.fn().mockResolvedValue({})

    vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
      schedules: mockSchedules,
      total: 1,
      loading: false,
      error: null,
      refetch,
    })

    // Override the default mock for this test
    vi.mocked(useSchedulesHook.useScheduleMutations).mockReturnValue({
      toggleSchedule,
      createSchedule: vi.fn(),
      updateSchedule: vi.fn(),
      deleteSchedule: vi.fn(),
      creating: false,
      updating: false,
      deleting: false,
    })

    vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
      workflows: mockWorkflows,
      total: 1,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    renderWithRouter(<Schedules />)

    const toggleButton = screen.getByRole('switch')
    fireEvent.click(toggleButton)

    await waitFor(() => {
      expect(toggleSchedule).toHaveBeenCalledWith('sched-1', false)
    })
  })

  it('should show total count', () => {
    vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
      schedules: mockSchedules,
      total: 5,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
      workflows: mockWorkflows,
      total: 1,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    renderWithRouter(<Schedules />)

    expect(screen.getByText(/5 total/i)).toBeInTheDocument()
  })

  it('should show empty state', () => {
    vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
      schedules: [],
      total: 0,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
      workflows: [],
      total: 0,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    renderWithRouter(<Schedules />)

    expect(screen.getByText('No schedules found')).toBeInTheDocument()
  })

  it('should show edit button for schedules', async () => {
    vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
      schedules: mockSchedules,
      total: 1,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
      workflows: mockWorkflows,
      total: 1,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    renderWithRouter(<Schedules />)

    const editButton = screen.getByRole('button', { name: /edit/i })
    expect(editButton).toBeInTheDocument()

    // Clicking edit button should not throw errors
    fireEvent.click(editButton)
    expect(editButton).toBeInTheDocument()
  })

  describe('Delete confirmation', () => {
    it('should show delete confirmation modal when delete clicked', () => {
      vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
        schedules: mockSchedules,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
        workflows: mockWorkflows,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter(<Schedules />)

      const deleteButton = screen.getByRole('button', { name: /delete/i })
      fireEvent.click(deleteButton)

      expect(screen.getByText('Delete Schedule')).toBeInTheDocument()
      expect(screen.getByText(/are you sure you want to delete/i)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /confirm/i })).toBeInTheDocument()
    })

    it('should close modal when cancel clicked', () => {
      vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
        schedules: mockSchedules,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
        workflows: mockWorkflows,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter(<Schedules />)

      // Open modal
      const deleteButton = screen.getByRole('button', { name: /delete/i })
      fireEvent.click(deleteButton)
      expect(screen.getByText('Delete Schedule')).toBeInTheDocument()

      // Click cancel
      const cancelButton = screen.getByRole('button', { name: /cancel/i })
      fireEvent.click(cancelButton)

      expect(screen.queryByText('Delete Schedule')).not.toBeInTheDocument()
    })

    it('should call deleteSchedule and refetch on confirm', async () => {
      const refetch = vi.fn()
      const deleteSchedule = vi.fn().mockResolvedValue({})

      vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
        schedules: mockSchedules,
        total: 1,
        loading: false,
        error: null,
        refetch,
      })

      vi.mocked(useSchedulesHook.useScheduleMutations).mockReturnValue({
        toggleSchedule: vi.fn(),
        createSchedule: vi.fn(),
        updateSchedule: vi.fn(),
        deleteSchedule,
        creating: false,
        updating: false,
        deleting: false,
      })

      vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
        workflows: mockWorkflows,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter(<Schedules />)

      // Open modal and confirm
      fireEvent.click(screen.getByRole('button', { name: /delete/i }))
      fireEvent.click(screen.getByRole('button', { name: /confirm/i }))

      await waitFor(() => {
        expect(deleteSchedule).toHaveBeenCalledWith('sched-1')
      })

      await waitFor(() => {
        expect(refetch).toHaveBeenCalled()
      })
    })

    it('should show delete error in modal', async () => {
      const deleteSchedule = vi.fn().mockRejectedValue(new Error('Delete failed'))

      vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
        schedules: mockSchedules,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      vi.mocked(useSchedulesHook.useScheduleMutations).mockReturnValue({
        toggleSchedule: vi.fn(),
        createSchedule: vi.fn(),
        updateSchedule: vi.fn(),
        deleteSchedule,
        creating: false,
        updating: false,
        deleting: false,
      })

      vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
        workflows: mockWorkflows,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter(<Schedules />)

      fireEvent.click(screen.getByRole('button', { name: /delete/i }))
      fireEvent.click(screen.getByRole('button', { name: /confirm/i }))

      await waitFor(() => {
        expect(screen.getByText('Delete failed')).toBeInTheDocument()
      })
    })

    it('should disable buttons during delete operation', () => {
      vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
        schedules: mockSchedules,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      vi.mocked(useSchedulesHook.useScheduleMutations).mockReturnValue({
        toggleSchedule: vi.fn(),
        createSchedule: vi.fn(),
        updateSchedule: vi.fn(),
        deleteSchedule: vi.fn(),
        creating: false,
        updating: false,
        deleting: true,
      })

      vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
        workflows: mockWorkflows,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter(<Schedules />)

      // All schedule actions should be disabled during delete operation
      const toggleButton = screen.getByRole('switch')
      expect(toggleButton).toBeDisabled()
    })
  })

  describe('Toggle error handling', () => {
    it('should show toggle error message', async () => {
      const toggleSchedule = vi.fn().mockRejectedValue(new Error('Toggle failed'))

      vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
        schedules: mockSchedules,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      vi.mocked(useSchedulesHook.useScheduleMutations).mockReturnValue({
        toggleSchedule,
        createSchedule: vi.fn(),
        updateSchedule: vi.fn(),
        deleteSchedule: vi.fn(),
        creating: false,
        updating: false,
        deleting: false,
      })

      vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
        workflows: mockWorkflows,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter(<Schedules />)

      const toggleButton = screen.getByRole('switch')
      fireEvent.click(toggleButton)

      await waitFor(() => {
        expect(screen.getByText('Toggle failed')).toBeInTheDocument()
      })
    })

    it('should show generic toggle error for non-Error objects', async () => {
      const toggleSchedule = vi.fn().mockRejectedValue('Unknown error')

      vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
        schedules: mockSchedules,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      vi.mocked(useSchedulesHook.useScheduleMutations).mockReturnValue({
        toggleSchedule,
        createSchedule: vi.fn(),
        updateSchedule: vi.fn(),
        deleteSchedule: vi.fn(),
        creating: false,
        updating: false,
        deleting: false,
      })

      vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
        workflows: mockWorkflows,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter(<Schedules />)

      fireEvent.click(screen.getByRole('switch'))

      await waitFor(() => {
        expect(screen.getByText('Toggle failed')).toBeInTheDocument()
      })
    })
  })

  describe('Search functionality', () => {
    it('should update search query when typing', () => {
      vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
        schedules: mockSchedules,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
        workflows: mockWorkflows,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter(<Schedules />)

      const searchInput = screen.getByPlaceholderText(/search schedules/i)
      fireEvent.change(searchInput, { target: { value: 'backup' } })

      expect(searchInput).toHaveValue('backup')
    })
  })

  describe('Sort functionality', () => {
    it('should show sort dropdown in list view', () => {
      vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
        schedules: mockSchedules,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
        workflows: mockWorkflows,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter(<Schedules />)

      const sortDropdown = screen.getByLabelText(/sort by/i)
      expect(sortDropdown).toBeInTheDocument()
    })

    it('should hide sort dropdown in calendar view', () => {
      vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
        schedules: mockSchedules,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
        workflows: mockWorkflows,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter(<Schedules />)

      // Switch to calendar view
      fireEvent.click(screen.getByText('Calendar View'))

      expect(screen.queryByLabelText(/sort by/i)).not.toBeInTheDocument()
    })

    it('should change sort option', () => {
      vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
        schedules: mockSchedules,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
        workflows: mockWorkflows,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter(<Schedules />)

      const sortDropdown = screen.getByLabelText(/sort by/i)
      fireEvent.change(sortDropdown, { target: { value: 'name' } })

      expect(sortDropdown).toHaveValue('name')
    })
  })

  describe('Error display', () => {
    it('should show detailed error message', () => {
      const error = new Error('Network connection timeout')

      vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
        schedules: [],
        total: 0,
        loading: false,
        error,
        refetch: vi.fn(),
      })

      vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
        workflows: [],
        total: 0,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter(<Schedules />)

      expect(screen.getByText('Network connection timeout')).toBeInTheDocument()
    })
  })

  describe('Filter parameters', () => {
    it('should pass enabled filter to hook when enabled is selected', async () => {
      vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
        schedules: mockSchedules,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
        workflows: mockWorkflows,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter(<Schedules />)

      const statusFilter = screen.getByLabelText(/status/i)
      fireEvent.change(statusFilter, { target: { value: 'enabled' } })

      await waitFor(() => {
        expect(useSchedulesHook.useSchedules).toHaveBeenCalledWith(
          expect.objectContaining({ enabled: true })
        )
      })
    })

    it('should pass disabled filter to hook when disabled is selected', async () => {
      vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
        schedules: mockSchedules,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
        workflows: mockWorkflows,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter(<Schedules />)

      const statusFilter = screen.getByLabelText(/status/i)
      fireEvent.change(statusFilter, { target: { value: 'disabled' } })

      await waitFor(() => {
        expect(useSchedulesHook.useSchedules).toHaveBeenCalledWith(
          expect.objectContaining({ enabled: false })
        )
      })
    })

    it('should pass search param to hook when search query entered', async () => {
      vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
        schedules: mockSchedules,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
        workflows: mockWorkflows,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter(<Schedules />)

      const searchInput = screen.getByPlaceholderText(/search schedules/i)
      fireEvent.change(searchInput, { target: { value: 'test' } })

      await waitFor(() => {
        expect(useSchedulesHook.useSchedules).toHaveBeenCalledWith(
          expect.objectContaining({ search: 'test' })
        )
      })
    })
  })

  describe('Disabled state during operations', () => {
    it('should disable schedule list during update', () => {
      vi.mocked(useSchedulesHook.useSchedules).mockReturnValue({
        schedules: mockSchedules,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      vi.mocked(useSchedulesHook.useScheduleMutations).mockReturnValue({
        toggleSchedule: vi.fn(),
        createSchedule: vi.fn(),
        updateSchedule: vi.fn(),
        deleteSchedule: vi.fn(),
        creating: false,
        updating: true,
        deleting: false,
      })

      vi.mocked(useWorkflowsHook.useWorkflows).mockReturnValue({
        workflows: mockWorkflows,
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter(<Schedules />)

      // Toggle should be disabled
      const toggleButton = screen.getByRole('switch')
      expect(toggleButton).toBeDisabled()
    })
  })
})
