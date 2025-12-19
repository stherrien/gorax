import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import ScheduleList from './ScheduleList'
import type { Schedule } from '../../api/schedules'

const mockSchedules: Schedule[] = [
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
  {
    id: 'sched-2',
    tenantId: 'tenant-1',
    workflowId: 'wf-2',
    name: 'Hourly Sync',
    cronExpression: '0 * * * *',
    timezone: 'UTC',
    enabled: false,
    nextRunAt: '2025-01-20T01:00:00Z',
    createdBy: 'user-1',
    createdAt: '2025-01-01T00:00:00Z',
    updatedAt: '2025-01-01T00:00:00Z',
  },
]

function renderWithRouter(component: React.ReactElement) {
  return render(<BrowserRouter>{component}</BrowserRouter>)
}

describe('ScheduleList', () => {
  it('should render list of schedules', () => {
    renderWithRouter(
      <ScheduleList
        schedules={mockSchedules}
        workflows={[
          { id: 'wf-1', name: 'Workflow 1' },
          { id: 'wf-2', name: 'Workflow 2' },
        ]}
        onToggle={() => {}}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    )

    expect(screen.getByText('Daily Backup')).toBeInTheDocument()
    expect(screen.getByText('Hourly Sync')).toBeInTheDocument()
  })

  it('should show empty state when no schedules', () => {
    renderWithRouter(
      <ScheduleList
        schedules={[]}
        workflows={[]}
        onToggle={() => {}}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    )

    expect(screen.getByText('No schedules found')).toBeInTheDocument()
  })

  it('should display workflow names correctly', () => {
    renderWithRouter(
      <ScheduleList
        schedules={mockSchedules}
        workflows={[
          { id: 'wf-1', name: 'Workflow 1' },
          { id: 'wf-2', name: 'Workflow 2' },
        ]}
        onToggle={() => {}}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    )

    expect(screen.getByText('Workflow 1')).toBeInTheDocument()
    expect(screen.getByText('Workflow 2')).toBeInTheDocument()
  })

  it('should show "Unknown Workflow" for missing workflow', () => {
    renderWithRouter(
      <ScheduleList
        schedules={mockSchedules}
        workflows={[]}
        onToggle={() => {}}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    )

    const unknownWorkflows = screen.getAllByText('Unknown Workflow')
    expect(unknownWorkflows).toHaveLength(2)
  })

  it('should call onToggle when toggle clicked', () => {
    const onToggle = vi.fn()

    renderWithRouter(
      <ScheduleList
        schedules={mockSchedules}
        workflows={[]}
        onToggle={onToggle}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    )

    const toggleButtons = screen.getAllByRole('switch')
    fireEvent.click(toggleButtons[0])

    expect(onToggle).toHaveBeenCalledWith('sched-1', true)
  })

  it('should call onEdit when edit clicked', () => {
    const onEdit = vi.fn()

    renderWithRouter(
      <ScheduleList
        schedules={mockSchedules}
        workflows={[]}
        onToggle={() => {}}
        onEdit={onEdit}
        onDelete={() => {}}
      />
    )

    const editButtons = screen.getAllByText('Edit')
    fireEvent.click(editButtons[0])

    expect(onEdit).toHaveBeenCalledWith('sched-1')
  })

  it('should call onDelete when delete clicked', () => {
    const onDelete = vi.fn()

    renderWithRouter(
      <ScheduleList
        schedules={mockSchedules}
        workflows={[]}
        onToggle={() => {}}
        onEdit={() => {}}
        onDelete={onDelete}
      />
    )

    const deleteButtons = screen.getAllByText('Delete')
    fireEvent.click(deleteButtons[0])

    expect(onDelete).toHaveBeenCalledWith('sched-1')
  })

  it('should sort schedules by next run time', () => {
    const unsortedSchedules = [...mockSchedules].reverse()

    renderWithRouter(
      <ScheduleList
        schedules={unsortedSchedules}
        workflows={[]}
        sortBy="nextRun"
        onToggle={() => {}}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    )

    const scheduleNames = screen
      .getAllByRole('heading', { level: 3 })
      .map((h) => h.textContent)

    expect(scheduleNames[0]).toBe('Daily Backup')
    expect(scheduleNames[1]).toBe('Hourly Sync')
  })

  it('should sort schedules by name', () => {
    renderWithRouter(
      <ScheduleList
        schedules={mockSchedules}
        workflows={[]}
        sortBy="name"
        onToggle={() => {}}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    )

    const scheduleNames = screen
      .getAllByRole('heading', { level: 3 })
      .map((h) => h.textContent)

    expect(scheduleNames[0]).toBe('Daily Backup')
    expect(scheduleNames[1]).toBe('Hourly Sync')
  })
})
