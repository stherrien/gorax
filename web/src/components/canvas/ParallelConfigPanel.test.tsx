import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import ParallelConfigPanel from './ParallelConfigPanel'

describe('ParallelConfigPanel', () => {
  const defaultConfig = {
    errorStrategy: 'fail_fast' as const,
    maxConcurrency: 0,
  }

  it('renders parallel configuration form', () => {
    const onChange = vi.fn()
    render(<ParallelConfigPanel config={defaultConfig} onChange={onChange} />)

    expect(screen.getByText('Parallel Configuration')).toBeInTheDocument()
    expect(screen.getByLabelText(/Error Strategy/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Max Concurrency/i)).toBeInTheDocument()
  })

  it('displays current error strategy', () => {
    const onChange = vi.fn()
    render(<ParallelConfigPanel config={defaultConfig} onChange={onChange} />)

    const select = screen.getByLabelText(/Error Strategy/i) as HTMLSelectElement
    expect(select.value).toBe('fail_fast')
  })

  it('displays current max concurrency', () => {
    const onChange = vi.fn()
    const config = { ...defaultConfig, maxConcurrency: 5 }
    render(<ParallelConfigPanel config={config} onChange={onChange} />)

    const input = screen.getByLabelText(/Max Concurrency/i) as HTMLInputElement
    expect(input.value).toBe('5')
  })

  it('calls onChange when error strategy changes', () => {
    const onChange = vi.fn()
    render(<ParallelConfigPanel config={defaultConfig} onChange={onChange} />)

    const select = screen.getByLabelText(/Error Strategy/i)
    fireEvent.change(select, { target: { value: 'wait_all' } })

    expect(onChange).toHaveBeenCalledWith({
      ...defaultConfig,
      errorStrategy: 'wait_all',
    })
  })

  it('calls onChange when max concurrency changes', () => {
    const onChange = vi.fn()
    render(<ParallelConfigPanel config={defaultConfig} onChange={onChange} />)

    const input = screen.getByLabelText(/Max Concurrency/i)
    fireEvent.change(input, { target: { value: '10' } })

    expect(onChange).toHaveBeenCalledWith({
      ...defaultConfig,
      maxConcurrency: 10,
    })
  })

  it('displays wait_all strategy correctly', () => {
    const onChange = vi.fn()
    const config = { ...defaultConfig, errorStrategy: 'wait_all' as const }
    render(<ParallelConfigPanel config={config} onChange={onChange} />)

    const select = screen.getByLabelText(/Error Strategy/i) as HTMLSelectElement
    expect(select.value).toBe('wait_all')
  })

  it('displays 0 max concurrency as unlimited', () => {
    const onChange = vi.fn()
    render(<ParallelConfigPanel config={defaultConfig} onChange={onChange} />)

    // Should show 0 as the value (means unlimited)
    const input = screen.getByLabelText(/Max Concurrency/i) as HTMLInputElement
    expect(input.value).toBe('0')
    // Help text should explain 0 means unlimited
    expect(screen.getByText(/0 = unlimited/i)).toBeInTheDocument()
  })

  it('has accessible form elements', () => {
    const onChange = vi.fn()
    render(<ParallelConfigPanel config={defaultConfig} onChange={onChange} />)

    // All form elements should have proper labels
    expect(screen.getByRole('combobox', { name: /Error Strategy/i })).toBeInTheDocument()
    expect(screen.getByRole('spinbutton', { name: /Max Concurrency/i })).toBeInTheDocument()
  })
})
