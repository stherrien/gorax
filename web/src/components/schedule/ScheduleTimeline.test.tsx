import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import ScheduleTimeline from './ScheduleTimeline'
import type { Schedule } from '../../api/schedules'

function renderWithRouter(component: React.ReactElement) {
  return render(<BrowserRouter>{component}</BrowserRouter>)
}

describe('ScheduleTimeline', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2025-01-20T12:00:00Z'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  const mockSchedules: Schedule[] = [
    {
      id: 'sched-1',
      tenantId: 'tenant-1',
      workflowId: 'wf-1',
      name: 'Schedule 1',
      cronExpression: '0 14 * * *',
      timezone: 'UTC',
      enabled: true,
      nextRunAt: '2025-01-20T14:00:00Z',
      createdBy: 'user-1',
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
    },
    {
      id: 'sched-2',
      tenantId: 'tenant-1',
      workflowId: 'wf-2',
      name: 'Schedule 2',
      cronExpression: '0 16 * * *',
      timezone: 'UTC',
      enabled: true,
      nextRunAt: '2025-01-20T16:00:00Z',
      createdBy: 'user-1',
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
    },
  ]

  it('should render timeline', () => {
    renderWithRouter(<ScheduleTimeline schedules={mockSchedules} />)

    expect(screen.getByText(/Next 24 Hours/i)).toBeInTheDocument()
  })

  it('should show now marker', () => {
    renderWithRouter(<ScheduleTimeline schedules={mockSchedules} />)

    expect(screen.getByText(/Now/i)).toBeInTheDocument()
  })

  it('should display schedules in timeline', () => {
    renderWithRouter(
      <ScheduleTimeline
        schedules={mockSchedules}
        workflows={[
          { id: 'wf-1', name: 'Workflow 1' },
          { id: 'wf-2', name: 'Workflow 2' },
        ]}
      />
    )

    expect(screen.getByText('Schedule 1')).toBeInTheDocument()
    expect(screen.getByText('Schedule 2')).toBeInTheDocument()
  })

  it('should group schedules by hour', () => {
    renderWithRouter(<ScheduleTimeline schedules={mockSchedules} />)

    // Should show hour markers
    expect(screen.getByText(/14:00/)).toBeInTheDocument()
    expect(screen.getByText(/16:00/)).toBeInTheDocument()
  })

  it('should handle empty schedules', () => {
    renderWithRouter(<ScheduleTimeline schedules={[]} />)

    expect(screen.getByText('No scheduled runs in the next 24 hours')).toBeInTheDocument()
  })

  it('should show schedule tooltip on hover', async () => {
    renderWithRouter(
      <ScheduleTimeline
        schedules={mockSchedules}
        workflows={[{ id: 'wf-1', name: 'Workflow 1' }]}
      />
    )

    const scheduleItem = screen.getByText('Schedule 1')
    fireEvent.mouseEnter(scheduleItem)

    // Tooltip should show workflow name
    expect(screen.getByText('Workflow 1')).toBeInTheDocument()
  })

  it('should handle schedules in different timezones', () => {
    const scheduleWithTimezone: Schedule = {
      id: 'sched-3',
      tenantId: 'tenant-1',
      workflowId: 'wf-3',
      name: 'Schedule 3',
      cronExpression: '0 9 * * *',
      timezone: 'America/New_York',
      enabled: true,
      nextRunAt: '2025-01-20T14:00:00Z', // 9am EST = 2pm UTC
      createdBy: 'user-1',
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
    }

    renderWithRouter(<ScheduleTimeline schedules={[scheduleWithTimezone]} />)

    expect(screen.getByText('Schedule 3')).toBeInTheDocument()
  })

  it('should filter out schedules beyond 24 hours', () => {
    const futureSchedule: Schedule = {
      id: 'sched-future',
      tenantId: 'tenant-1',
      workflowId: 'wf-1',
      name: 'Future Schedule',
      cronExpression: '0 0 * * *',
      timezone: 'UTC',
      enabled: true,
      nextRunAt: '2025-01-22T00:00:00Z', // 36 hours in future
      createdBy: 'user-1',
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
    }

    renderWithRouter(
      <ScheduleTimeline schedules={[...mockSchedules, futureSchedule]} />
    )

    expect(screen.getByText('Schedule 1')).toBeInTheDocument()
    expect(screen.queryByText('Future Schedule')).not.toBeInTheDocument()
  })

  it('should not show past schedules', () => {
    const pastSchedule: Schedule = {
      id: 'sched-past',
      tenantId: 'tenant-1',
      workflowId: 'wf-1',
      name: 'Past Schedule',
      cronExpression: '0 10 * * *',
      timezone: 'UTC',
      enabled: true,
      nextRunAt: '2025-01-20T10:00:00Z', // 2 hours ago (before current time)
      createdBy: 'user-1',
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
    }

    renderWithRouter(<ScheduleTimeline schedules={[pastSchedule]} />)

    // Past schedules should not show (they're in the past)
    expect(screen.queryByText('Past Schedule')).not.toBeInTheDocument()
    expect(screen.getByText('No scheduled runs in the next 24 hours')).toBeInTheDocument()
  })
})
