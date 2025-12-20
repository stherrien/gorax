import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import PrioritySelector from './PrioritySelector'

describe('PrioritySelector', () => {
  it('renders select element with label', () => {
    const onChange = vi.fn()
    render(<PrioritySelector value={1} onChange={onChange} />)

    expect(screen.getByLabelText(/priority/i)).toBeInTheDocument()
  })

  it('renders all priority options', () => {
    const onChange = vi.fn()
    render(<PrioritySelector value={1} onChange={onChange} />)

    const select = screen.getByLabelText(/priority/i) as HTMLSelectElement

    expect(select).toBeInTheDocument()
    const options = Array.from(select.options).map((opt) => opt.textContent)

    expect(options).toContain('Low (0)')
    expect(options).toContain('Normal (1)')
    expect(options).toContain('High (2)')
    expect(options).toContain('Critical (3)')
  })

  it('displays selected priority value', () => {
    const onChange = vi.fn()
    render(<PrioritySelector value={2} onChange={onChange} />)

    const select = screen.getByLabelText(/priority/i) as HTMLSelectElement
    expect(select.value).toBe('2')
  })

  it('calls onChange when priority is changed', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<PrioritySelector value={1} onChange={onChange} />)

    const select = screen.getByLabelText(/priority/i)
    await user.selectOptions(select, '3')

    expect(onChange).toHaveBeenCalledWith(3)
  })

  it('calls onChange with number type, not string', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<PrioritySelector value={0} onChange={onChange} />)

    const select = screen.getByLabelText(/priority/i)
    await user.selectOptions(select, '2')

    expect(onChange).toHaveBeenCalledWith(2)
    expect(typeof onChange.mock.calls[0][0]).toBe('number')
  })

  it('renders with default value when no value prop provided', () => {
    const onChange = vi.fn()
    render(<PrioritySelector onChange={onChange} />)

    const select = screen.getByLabelText(/priority/i) as HTMLSelectElement
    expect(select.value).toBe('1')
  })

  it('applies disabled prop correctly', () => {
    const onChange = vi.fn()
    render(<PrioritySelector value={1} onChange={onChange} disabled />)

    const select = screen.getByLabelText(/priority/i)
    expect(select).toBeDisabled()
  })

  it('does not call onChange when disabled', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<PrioritySelector value={1} onChange={onChange} disabled />)

    const select = screen.getByLabelText(/priority/i)
    await user.selectOptions(select, '3')

    expect(onChange).not.toHaveBeenCalled()
  })

  it('renders with custom id prop', () => {
    const onChange = vi.fn()
    render(<PrioritySelector value={1} onChange={onChange} id="custom-priority" />)

    const select = screen.getByLabelText(/priority/i)
    expect(select).toHaveAttribute('id', 'custom-priority')
  })

  it('renders with default id when no id prop provided', () => {
    const onChange = vi.fn()
    render(<PrioritySelector value={1} onChange={onChange} />)

    const select = screen.getByLabelText(/priority/i)
    expect(select).toHaveAttribute('id', 'priority-selector')
  })

  it('applies correct styling classes', () => {
    const onChange = vi.fn()
    render(<PrioritySelector value={1} onChange={onChange} />)

    const select = screen.getByLabelText(/priority/i)
    expect(select).toHaveClass('w-full')
    expect(select).toHaveClass('px-3')
    expect(select).toHaveClass('py-2')
    expect(select).toHaveClass('bg-gray-700')
    expect(select).toHaveClass('text-white')
    expect(select).toHaveClass('rounded-lg')
  })

  it('shows priority descriptions in option text', () => {
    const onChange = vi.fn()
    render(<PrioritySelector value={1} onChange={onChange} />)

    const select = screen.getByLabelText(/priority/i) as HTMLSelectElement
    const options = Array.from(select.options)

    expect(options[0].textContent).toBe('Low (0)')
    expect(options[1].textContent).toBe('Normal (1)')
    expect(options[2].textContent).toBe('High (2)')
    expect(options[3].textContent).toBe('Critical (3)')
  })
})
