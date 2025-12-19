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
})
