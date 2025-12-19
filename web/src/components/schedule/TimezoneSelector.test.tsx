import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { TimezoneSelector } from './TimezoneSelector'

describe('TimezoneSelector', () => {
  const mockOnChange = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders with default UTC selection', () => {
    render(<TimezoneSelector value="UTC" onChange={mockOnChange} />)

    expect(screen.getByRole('combobox')).toHaveValue('UTC')
  })

  it('displays current value', () => {
    render(<TimezoneSelector value="America/New_York" onChange={mockOnChange} />)

    expect(screen.getByRole('combobox')).toHaveValue('America/New_York')
  })

  it('shows popular timezones at the top', () => {
    render(<TimezoneSelector value="UTC" onChange={mockOnChange} />)

    const select = screen.getByRole('combobox')
    const options = Array.from(select.querySelectorAll('option'))
    const optionValues = options.map((opt) => opt.value)

    expect(optionValues.slice(0, 4)).toContain('UTC')
    expect(optionValues.slice(0, 4)).toContain('America/New_York')
    expect(optionValues.slice(0, 4)).toContain('America/Los_Angeles')
    expect(optionValues.slice(0, 4)).toContain('Europe/London')
  })

  it('calls onChange when timezone is selected', () => {
    render(<TimezoneSelector value="UTC" onChange={mockOnChange} />)

    const select = screen.getByRole('combobox')
    fireEvent.change(select, { target: { value: 'America/New_York' } })

    expect(mockOnChange).toHaveBeenCalledWith('America/New_York')
  })

  it('groups timezones by region', () => {
    render(<TimezoneSelector value="UTC" onChange={mockOnChange} />)

    const select = screen.getByRole('combobox')
    const optgroups = select.querySelectorAll('optgroup')

    expect(optgroups.length).toBeGreaterThan(0)

    const labels = Array.from(optgroups).map((og) => og.getAttribute('label'))
    expect(labels).toContain('America')
    expect(labels).toContain('Europe')
    expect(labels).toContain('Asia')
  })

  it('can be disabled', () => {
    render(<TimezoneSelector value="UTC" onChange={mockOnChange} disabled />)

    expect(screen.getByRole('combobox')).toBeDisabled()
  })

  it('displays timezone offset hint', () => {
    render(<TimezoneSelector value="America/New_York" onChange={mockOnChange} />)

    // Should show the timezone label with offset
    const label = screen.getByText(/Timezone/i)
    expect(label).toBeInTheDocument()
    expect(label.textContent).toContain('UTC-5')
  })

  it('supports search filtering', () => {
    render(<TimezoneSelector value="UTC" onChange={mockOnChange} searchable />)

    const searchInput = screen.getByPlaceholderText(/Search timezone/i)
    expect(searchInput).toBeInTheDocument()

    fireEvent.change(searchInput, { target: { value: 'London' } })

    // After search, only matching timezones should be visible
    const options = screen.getAllByRole('option')
    const visibleOptions = options.filter(
      (opt) => opt.textContent && opt.textContent.includes('London')
    )
    expect(visibleOptions.length).toBeGreaterThan(0)
  })

  it('shows current time in selected timezone', () => {
    render(<TimezoneSelector value="America/New_York" onChange={mockOnChange} showCurrentTime />)

    // Should display current time in the selected timezone
    expect(screen.getByText(/Current time:/i)).toBeInTheDocument()
  })
})
