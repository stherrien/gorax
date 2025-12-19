import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { CronBuilder } from './CronBuilder'

describe('CronBuilder', () => {
  const mockOnChange = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders with default empty state', () => {
    render(<CronBuilder value="" onChange={mockOnChange} />)

    expect(screen.getByText(/Cron Expression/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Minute/i)).toBeInTheDocument()
  })

  it('displays current cron expression in description', () => {
    render(<CronBuilder value="0 9 * * *" onChange={mockOnChange} />)

    // Check that description is not "No schedule set" or "Invalid"
    expect(screen.queryByText(/No schedule set/i)).not.toBeInTheDocument()
    expect(screen.queryByText(/Invalid cron expression/i)).not.toBeInTheDocument()

    // Description should contain something about 9 AM
    const text = screen.getByText(/9/)
    expect(text).toBeInTheDocument()
  })

  it('shows preset options', () => {
    render(<CronBuilder value="" onChange={mockOnChange} />)

    const presetButton = screen.getByRole('button', { name: /Presets/i })
    fireEvent.click(presetButton)

    expect(screen.getByText(/Every minute/i)).toBeInTheDocument()
    expect(screen.getByText(/Every hour/i)).toBeInTheDocument()
    expect(screen.getByText(/Every day at midnight/i)).toBeInTheDocument()
    expect(screen.getByText(/Every day at 9 AM/i)).toBeInTheDocument()
    expect(screen.getByText(/Weekly on Monday at 9 AM/i)).toBeInTheDocument()
    expect(screen.getByText(/Monthly on the 1st/i)).toBeInTheDocument()
    expect(screen.getByText(/Weekdays at 9 AM/i)).toBeInTheDocument()
  })

  it('applies preset when selected', () => {
    render(<CronBuilder value="" onChange={mockOnChange} />)

    const presetButton = screen.getByRole('button', { name: /Presets/i })
    fireEvent.click(presetButton)

    const everyHourOption = screen.getByText(/Every hour/i)
    fireEvent.click(everyHourOption)

    expect(mockOnChange).toHaveBeenCalledWith('0 * * * *')
  })

  it('calls onChange when cron expression is manually edited in advanced mode', () => {
    render(<CronBuilder value="" onChange={mockOnChange} />)

    // Switch to advanced mode first
    const advancedButton = screen.getByRole('button', { name: /Advanced/i })
    fireEvent.click(advancedButton)

    const input = screen.getByLabelText(/Cron expression input/i)
    fireEvent.change(input, { target: { value: '*/5 * * * *' } })

    expect(mockOnChange).toHaveBeenCalledWith('*/5 * * * *')
  })

  it('shows human-readable description', () => {
    render(<CronBuilder value="0 9 * * 1" onChange={mockOnChange} />)

    // Should show a description mentioning Monday and 9 AM
    expect(screen.getByText(/Monday/)).toBeInTheDocument()
  })

  it('shows error state for invalid cron expression', () => {
    render(<CronBuilder value="invalid" onChange={mockOnChange} />)

    expect(screen.getByText(/Invalid cron expression/i)).toBeInTheDocument()
  })

  it('allows switching between simple and advanced mode', () => {
    render(<CronBuilder value="" onChange={mockOnChange} />)

    // Should default to simple mode with dropdowns
    expect(screen.getByLabelText(/Minute/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Hour/i)).toBeInTheDocument()

    // Switch to advanced mode
    const advancedButton = screen.getByRole('button', { name: /Advanced/i })
    fireEvent.click(advancedButton)

    // Should show raw cron input (only one textbox in advanced mode)
    expect(screen.getByLabelText(/Cron expression input/i)).toBeInTheDocument()
    expect(screen.queryByLabelText(/Minute/i)).not.toBeInTheDocument()
  })

  it('updates description when expression changes', () => {
    const { rerender } = render(<CronBuilder value="0 9 * * *" onChange={mockOnChange} />)

    // Initially shows 9
    expect(screen.getByText(/9/)).toBeInTheDocument()

    rerender(<CronBuilder value="0 12 * * *" onChange={mockOnChange} />)

    // After update shows 12
    expect(screen.getByText(/12/)).toBeInTheDocument()
  })

  it('handles preset for "Every minute"', () => {
    render(<CronBuilder value="" onChange={mockOnChange} />)

    const presetButton = screen.getByRole('button', { name: /Presets/i })
    fireEvent.click(presetButton)

    const everyMinuteOption = screen.getByText(/Every minute/i)
    fireEvent.click(everyMinuteOption)

    expect(mockOnChange).toHaveBeenCalledWith('* * * * *')
  })

  it('handles preset for "Weekdays at 9 AM"', () => {
    render(<CronBuilder value="" onChange={mockOnChange} />)

    const presetButton = screen.getByRole('button', { name: /Presets/i })
    fireEvent.click(presetButton)

    const weekdaysOption = screen.getByText(/Weekdays at 9 AM/i)
    fireEvent.click(weekdaysOption)

    expect(mockOnChange).toHaveBeenCalledWith('0 9 * * 1-5')
  })
})
