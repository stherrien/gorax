import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import DateRangePicker from './DateRangePicker'

describe('DateRangePicker', () => {
  it('renders with default state (no dates selected)', () => {
    const onChange = vi.fn()
    render(<DateRangePicker value={null} onChange={onChange} />)

    expect(screen.getByText(/select date range/i)).toBeInTheDocument()
  })

  it('displays selected date range', () => {
    const onChange = vi.fn()
    const startDate = new Date('2025-01-01')
    const endDate = new Date('2025-01-31')

    render(
      <DateRangePicker
        value={{ startDate, endDate }}
        onChange={onChange}
      />
    )

    expect(screen.getByText(/jan 1, 2025 - jan 31, 2025/i)).toBeInTheDocument()
  })

  it('opens calendar popover when clicked', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<DateRangePicker value={null} onChange={onChange} />)

    const button = screen.getByRole('button')
    await user.click(button)

    expect(screen.getByRole('dialog')).toBeInTheDocument()
  })

  it('applies preset range when preset button clicked', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<DateRangePicker value={null} onChange={onChange} />)

    await user.click(screen.getByRole('button'))
    await user.click(screen.getByText(/last 7 days/i))

    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({
        startDate: expect.any(Date),
        endDate: expect.any(Date),
      })
    )
  })

  it('applies "Today" preset correctly', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<DateRangePicker value={null} onChange={onChange} />)

    await user.click(screen.getByRole('button'))
    await user.click(screen.getByText(/^today$/i))

    expect(onChange).toHaveBeenCalled()
    const call = onChange.mock.calls[0][0]

    const today = new Date()
    today.setHours(0, 0, 0, 0)
    const todayEnd = new Date()
    todayEnd.setHours(23, 59, 59, 999)

    expect(call.startDate.getDate()).toBe(today.getDate())
    expect(call.endDate.getDate()).toBe(todayEnd.getDate())
  })

  it('applies "Yesterday" preset correctly', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<DateRangePicker value={null} onChange={onChange} />)

    await user.click(screen.getByRole('button'))
    await user.click(screen.getByText(/yesterday/i))

    expect(onChange).toHaveBeenCalled()
    const call = onChange.mock.calls[0][0]

    const yesterday = new Date()
    yesterday.setDate(yesterday.getDate() - 1)
    yesterday.setHours(0, 0, 0, 0)

    expect(call.startDate.getDate()).toBe(yesterday.getDate())
  })

  it('applies "Last 30 days" preset correctly', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<DateRangePicker value={null} onChange={onChange} />)

    await user.click(screen.getByRole('button'))
    await user.click(screen.getByText(/last 30 days/i))

    expect(onChange).toHaveBeenCalled()
    const call = onChange.mock.calls[0][0]

    const daysDiff = Math.floor(
      (call.endDate.getTime() - call.startDate.getTime()) / (1000 * 60 * 60 * 24)
    )

    expect(daysDiff).toBeGreaterThanOrEqual(29)
    expect(daysDiff).toBeLessThanOrEqual(30)
  })

  it('applies "This month" preset correctly', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<DateRangePicker value={null} onChange={onChange} />)

    await user.click(screen.getByRole('button'))
    await user.click(screen.getByText(/this month/i))

    expect(onChange).toHaveBeenCalled()
    const call = onChange.mock.calls[0][0]

    const now = new Date()
    expect(call.startDate.getMonth()).toBe(now.getMonth())
    expect(call.startDate.getDate()).toBe(1)
  })

  it('clears date range when clear button clicked', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    const startDate = new Date('2025-01-01')
    const endDate = new Date('2025-01-31')

    render(
      <DateRangePicker
        value={{ startDate, endDate }}
        onChange={onChange}
      />
    )

    await user.click(screen.getByRole('button'))
    await user.click(screen.getByText(/clear/i))

    expect(onChange).toHaveBeenCalledWith(null)
  })

  it('closes popover when "Apply" button clicked', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<DateRangePicker value={null} onChange={onChange} />)

    await user.click(screen.getByRole('button'))
    expect(screen.getByRole('dialog')).toBeInTheDocument()

    await user.click(screen.getByText(/today/i))
    await user.click(screen.getByText(/apply/i))

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('shows placeholder when no date selected', () => {
    const onChange = vi.fn()
    render(
      <DateRangePicker
        value={null}
        onChange={onChange}
        placeholder="Choose dates..."
      />
    )

    expect(screen.getByText(/choose dates.../i)).toBeInTheDocument()
  })

  it('formats date range according to locale', () => {
    const onChange = vi.fn()
    const startDate = new Date('2025-12-01')
    const endDate = new Date('2025-12-15')

    render(
      <DateRangePicker
        value={{ startDate, endDate }}
        onChange={onChange}
      />
    )

    expect(screen.getByText(/dec 1, 2025 - dec 15, 2025/i)).toBeInTheDocument()
  })

  it('disables the component when disabled prop is true', () => {
    const onChange = vi.fn()
    render(<DateRangePicker value={null} onChange={onChange} disabled />)

    const button = screen.getByRole('button')
    expect(button).toBeDisabled()
  })

  it('applies custom className to root element', () => {
    const onChange = vi.fn()
    const { container } = render(
      <DateRangePicker
        value={null}
        onChange={onChange}
        className="custom-class"
      />
    )

    expect(container.firstChild).toHaveClass('custom-class')
  })
})
