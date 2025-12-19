import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import ScheduleCard from './ScheduleCard'
import type { Schedule } from '../../api/schedules'

const mockSchedule: Schedule = {
  id: 'sched-1',
  tenantId: 'tenant-1',
  workflowId: 'wf-1',
  name: 'Daily Backup',
  cronExpression: '0 0 * * *',
  timezone: 'UTC',
  enabled: true,
  nextRunAt: '2025-01-20T00:00:00Z',
  lastRunAt: '2025-01-19T00:00:00Z',
  createdBy: 'user-1',
  createdAt: '2025-01-01T00:00:00Z',
  updatedAt: '2025-01-01T00:00:00Z',
}

function renderWithRouter(component: React.ReactElement) {
  return render(<BrowserRouter>{component}</BrowserRouter>)
}

describe('ScheduleCard', () => {
  it('should render schedule information', () => {
    renderWithRouter(
      <ScheduleCard
        schedule={mockSchedule}
        workflowName="Test Workflow"
        onToggle={() => {}}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    )

    expect(screen.getByText('Daily Backup')).toBeInTheDocument()
    expect(screen.getByText('0 0 * * *')).toBeInTheDocument()
    expect(screen.getByText('Test Workflow')).toBeInTheDocument()
  })

  it('should show enabled badge when enabled', () => {
    renderWithRouter(
      <ScheduleCard
        schedule={mockSchedule}
        workflowName="Test Workflow"
        onToggle={() => {}}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    )

    expect(screen.getByText('Enabled')).toBeInTheDocument()
  })

  it('should show disabled badge when disabled', () => {
    const disabledSchedule = { ...mockSchedule, enabled: false }

    renderWithRouter(
      <ScheduleCard
        schedule={disabledSchedule}
        workflowName="Test Workflow"
        onToggle={() => {}}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    )

    expect(screen.getByText('Disabled')).toBeInTheDocument()
  })

  it('should call onToggle when toggle button clicked', () => {
    const onToggle = vi.fn()

    renderWithRouter(
      <ScheduleCard
        schedule={mockSchedule}
        workflowName="Test Workflow"
        onToggle={onToggle}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    )

    const toggleButton = screen.getByRole('switch')
    fireEvent.click(toggleButton)

    expect(onToggle).toHaveBeenCalledWith(mockSchedule.id, mockSchedule.enabled)
  })

  it('should call onEdit when edit button clicked', () => {
    const onEdit = vi.fn()

    renderWithRouter(
      <ScheduleCard
        schedule={mockSchedule}
        workflowName="Test Workflow"
        onToggle={() => {}}
        onEdit={onEdit}
        onDelete={() => {}}
      />
    )

    const editButton = screen.getByText('Edit')
    fireEvent.click(editButton)

    expect(onEdit).toHaveBeenCalledWith(mockSchedule.id)
  })

  it('should call onDelete when delete button clicked', () => {
    const onDelete = vi.fn()

    renderWithRouter(
      <ScheduleCard
        schedule={mockSchedule}
        workflowName="Test Workflow"
        onToggle={() => {}}
        onEdit={() => {}}
        onDelete={onDelete}
      />
    )

    const deleteButton = screen.getByText('Delete')
    fireEvent.click(deleteButton)

    expect(onDelete).toHaveBeenCalledWith(mockSchedule.id)
  })

  it('should display next run time in relative format', () => {
    renderWithRouter(
      <ScheduleCard
        schedule={mockSchedule}
        workflowName="Test Workflow"
        onToggle={() => {}}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    )

    // Should show some form of time display
    expect(screen.getByText(/Next run:/)).toBeInTheDocument()
  })

  it('should handle schedule without next run time', () => {
    const scheduleWithoutNextRun = { ...mockSchedule, nextRunAt: undefined }

    renderWithRouter(
      <ScheduleCard
        schedule={scheduleWithoutNextRun}
        workflowName="Test Workflow"
        onToggle={() => {}}
        onEdit={() => {}}
        onDelete={() => {}}
      />
    )

    expect(screen.getByText(/Next run:/)).toBeInTheDocument()
  })
})
