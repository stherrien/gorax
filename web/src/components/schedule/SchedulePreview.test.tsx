import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { SchedulePreview } from './SchedulePreview'
import { scheduleAPI } from '../../api/schedules'

vi.mock('../../api/schedules')

describe('SchedulePreview', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading state while fetching', () => {
    vi.mocked(scheduleAPI.preview).mockImplementation(
      () => new Promise(() => {}) // Never resolves
    )

    render(<SchedulePreview cronExpression="0 9 * * *" timezone="UTC" />)

    expect(screen.getByText(/Loading/i)).toBeInTheDocument()
  })

  it('displays next run times', async () => {
    vi.mocked(scheduleAPI.preview).mockResolvedValue({
      valid: true,
      next_runs: [
        '2025-12-20T09:00:00Z',
        '2025-12-21T09:00:00Z',
        '2025-12-22T09:00:00Z',
      ],
      count: 3,
      timezone: 'UTC',
    })

    render(<SchedulePreview cronExpression="0 9 * * *" timezone="UTC" />)

    await waitFor(() => {
      expect(screen.getByText(/Next Run Times/i)).toBeInTheDocument()
    })

    expect(screen.getByText(/Dec 20, 2025/)).toBeInTheDocument()
    expect(screen.getByText(/Dec 21, 2025/)).toBeInTheDocument()
    expect(screen.getByText(/Dec 22, 2025/)).toBeInTheDocument()
  })

  it('shows error state for invalid expression', async () => {
    vi.mocked(scheduleAPI.preview).mockRejectedValue(
      new Error('invalid cron expression')
    )

    render(<SchedulePreview cronExpression="invalid" timezone="UTC" />)

    await waitFor(() => {
      expect(screen.getByText(/Error/i)).toBeInTheDocument()
    })
  })

  it('displays custom count of run times', async () => {
    vi.mocked(scheduleAPI.preview).mockResolvedValue({
      valid: true,
      next_runs: ['2025-12-20T09:00:00Z'],
      count: 1,
      timezone: 'UTC',
    })

    render(<SchedulePreview cronExpression="0 9 * * *" timezone="UTC" count={1} />)

    await waitFor(() => {
      expect(screen.getByText(/Dec 20, 2025/)).toBeInTheDocument()
    })

    expect(screen.queryByText(/Dec 21, 2025/)).not.toBeInTheDocument()
  })

  it('shows timezone in display', async () => {
    vi.mocked(scheduleAPI.preview).mockResolvedValue({
      valid: true,
      next_runs: ['2025-12-20T14:00:00Z'],
      count: 1,
      timezone: 'America/New_York',
    })

    render(
      <SchedulePreview cronExpression="0 9 * * *" timezone="America/New_York" />
    )

    await waitFor(() => {
      expect(screen.getByText(/America\/New_York/i)).toBeInTheDocument()
    })
  })

  it('refetches when cron expression changes', async () => {
    vi.mocked(scheduleAPI.preview).mockResolvedValue({
      valid: true,
      next_runs: ['2025-12-20T09:00:00Z'],
      count: 1,
      timezone: 'UTC',
    })

    const { rerender } = render(
      <SchedulePreview cronExpression="0 9 * * *" timezone="UTC" />
    )

    await waitFor(() => {
      expect(scheduleAPI.preview).toHaveBeenCalledWith('0 9 * * *', 'UTC', 10)
    })

    vi.mocked(scheduleAPI.preview).mockClear()

    rerender(<SchedulePreview cronExpression="0 12 * * *" timezone="UTC" />)

    await waitFor(() => {
      expect(scheduleAPI.preview).toHaveBeenCalledWith('0 12 * * *', 'UTC', 10)
    })
  })

  it('refetches when timezone changes', async () => {
    vi.mocked(scheduleAPI.preview).mockResolvedValue({
      valid: true,
      next_runs: ['2025-12-20T09:00:00Z'],
      count: 1,
      timezone: 'UTC',
    })

    const { rerender } = render(
      <SchedulePreview cronExpression="0 9 * * *" timezone="UTC" />
    )

    await waitFor(() => {
      expect(scheduleAPI.preview).toHaveBeenCalledWith('0 9 * * *', 'UTC', 10)
    })

    vi.mocked(scheduleAPI.preview).mockClear()

    rerender(<SchedulePreview cronExpression="0 9 * * *" timezone="America/New_York" />)

    await waitFor(() => {
      expect(scheduleAPI.preview).toHaveBeenCalledWith(
        '0 9 * * *',
        'America/New_York',
        10
      )
    })
  })

  it('does not fetch for empty cron expression', () => {
    render(<SchedulePreview cronExpression="" timezone="UTC" />)

    expect(scheduleAPI.preview).not.toHaveBeenCalled()
    expect(screen.getByText(/No schedule set/i)).toBeInTheDocument()
  })

  it('displays relative time hints', async () => {
    const tomorrow = new Date()
    tomorrow.setDate(tomorrow.getDate() + 1)
    const tomorrowStr = tomorrow.toISOString()

    vi.mocked(scheduleAPI.preview).mockResolvedValue({
      valid: true,
      next_runs: [tomorrowStr],
      count: 1,
      timezone: 'UTC',
    })

    render(<SchedulePreview cronExpression="0 9 * * *" timezone="UTC" />)

    await waitFor(() => {
      expect(screen.getByText(/tomorrow/i)).toBeInTheDocument()
    })
  })
})
