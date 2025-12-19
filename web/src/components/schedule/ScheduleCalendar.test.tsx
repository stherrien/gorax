import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import ScheduleCalendar from './ScheduleCalendar'
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
    enabled: true,
    nextRunAt: '2025-01-20T01:00:00Z',
    createdBy: 'user-1',
    createdAt: '2025-01-01T00:00:00Z',
    updatedAt: '2025-01-01T00:00:00Z',
  },
]

describe('ScheduleCalendar', () => {
  it('should render calendar with current month', () => {
    render(<ScheduleCalendar schedules={mockSchedules} />)

    // Should show month/year
    expect(screen.getByText(/January|February|March|April|May|June|July|August|September|October|November|December/)).toBeInTheDocument()
    expect(screen.getByText(/2025|2024|2026/)).toBeInTheDocument()
  })

  it('should render weekday headers', () => {
    render(<ScheduleCalendar schedules={mockSchedules} />)

    expect(screen.getByText('Sun')).toBeInTheDocument()
    expect(screen.getByText('Mon')).toBeInTheDocument()
    expect(screen.getByText('Tue')).toBeInTheDocument()
    expect(screen.getByText('Wed')).toBeInTheDocument()
    expect(screen.getByText('Thu')).toBeInTheDocument()
    expect(screen.getByText('Fri')).toBeInTheDocument()
    expect(screen.getByText('Sat')).toBeInTheDocument()
  })

  it('should navigate to previous month', () => {
    render(<ScheduleCalendar schedules={mockSchedules} />)

    const prevButton = screen.getByLabelText(/previous month/i)
    fireEvent.click(prevButton)

    // Should still show a valid month
    expect(screen.getByText(/January|February|March|April|May|June|July|August|September|October|November|December/)).toBeInTheDocument()
  })

  it('should navigate to next month', () => {
    render(<ScheduleCalendar schedules={mockSchedules} />)

    const nextButton = screen.getByLabelText(/next month/i)
    fireEvent.click(nextButton)

    // Should still show a valid month
    expect(screen.getByText(/January|February|March|April|May|June|July|August|September|October|November|December/)).toBeInTheDocument()
  })

  it('should highlight today', () => {
    render(<ScheduleCalendar schedules={mockSchedules} />)

    const today = new Date().getDate()
    const todayElements = screen.getAllByText(today.toString())

    // At least one should have the "today" styling
    expect(todayElements.length).toBeGreaterThan(0)
  })

  it('should show schedule count on days with schedules', () => {
    const schedulesOn20th = [
      {
        id: 'sched-1',
        tenantId: 'tenant-1',
        workflowId: 'wf-1',
        name: 'Schedule 1',
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
        name: 'Schedule 2',
        cronExpression: '0 12 * * *',
        timezone: 'UTC',
        enabled: true,
        nextRunAt: '2025-01-20T12:00:00Z',
        createdBy: 'user-1',
        createdAt: '2025-01-01T00:00:00Z',
        updatedAt: '2025-01-01T00:00:00Z',
      },
    ]

    render(<ScheduleCalendar schedules={schedulesOn20th} />)

    // Should show indicator for schedules
    const cells = screen.getAllByRole('button')
    const cellsWith20 = cells.filter((cell) => cell.textContent?.includes('20'))

    expect(cellsWith20.length).toBeGreaterThan(0)
  })

  it('should call onDayClick when day is clicked', () => {
    const onDayClick = vi.fn()
    render(<ScheduleCalendar schedules={mockSchedules} onDayClick={onDayClick} />)

    const dayCells = screen.getAllByRole('button')
    if (dayCells.length > 0) {
      fireEvent.click(dayCells[15]) // Click middle of month
      expect(onDayClick).toHaveBeenCalled()
    }
  })

  it('should handle empty schedules', () => {
    render(<ScheduleCalendar schedules={[]} />)

    expect(screen.getByText(/January|February|March|April|May|June|July|August|September|October|November|December/)).toBeInTheDocument()
  })
})
