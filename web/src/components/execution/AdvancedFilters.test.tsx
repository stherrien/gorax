import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import AdvancedFilters from './AdvancedFilters'
import type { ExecutionListParams } from '../../api/executions'

describe('AdvancedFilters', () => {
  const defaultFilters: ExecutionListParams = {}

  it('renders collapsed by default', () => {
    const onChange = vi.fn()
    render(<AdvancedFilters filters={defaultFilters} onChange={onChange} />)

    expect(screen.getByText(/advanced filters/i)).toBeInTheDocument()
    expect(screen.queryByLabelText(/status/i)).not.toBeInTheDocument()
  })

  it('expands when toggle button is clicked', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<AdvancedFilters filters={defaultFilters} onChange={onChange} />)

    await user.click(screen.getByText(/advanced filters/i))

    expect(screen.getByLabelText(/status/i)).toBeInTheDocument()
  })

  it('collapses when toggle button is clicked again', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<AdvancedFilters filters={defaultFilters} onChange={onChange} />)

    await user.click(screen.getByText(/advanced filters/i))
    expect(screen.getByLabelText(/status/i)).toBeInTheDocument()

    await user.click(screen.getByText(/advanced filters/i))
    await waitFor(() => {
      expect(screen.queryByLabelText(/status/i)).not.toBeInTheDocument()
    })
  })

  it('shows status checkboxes when expanded', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<AdvancedFilters filters={defaultFilters} onChange={onChange} />)

    await user.click(screen.getByText(/advanced filters/i))

    expect(screen.getByLabelText(/completed/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/failed/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/running/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/cancelled/i)).toBeInTheDocument()
  })

  it('updates status filter when checkbox is clicked', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<AdvancedFilters filters={defaultFilters} onChange={onChange} />)

    await user.click(screen.getByText(/advanced filters/i))
    await user.click(screen.getByLabelText(/completed/i))

    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({
        status: ['completed'],
      })
    )
  })

  it('allows multiple status selections', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<AdvancedFilters filters={defaultFilters} onChange={onChange} />)

    await user.click(screen.getByText(/advanced filters/i))
    await user.click(screen.getByLabelText(/completed/i))
    await user.click(screen.getByLabelText(/failed/i))

    expect(onChange).toHaveBeenLastCalledWith(
      expect.objectContaining({
        status: expect.arrayContaining(['completed', 'failed']),
      })
    )
  })

  it('shows trigger type checkboxes', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<AdvancedFilters filters={defaultFilters} onChange={onChange} />)

    await user.click(screen.getByText(/advanced filters/i))

    expect(screen.getByLabelText(/manual/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/scheduled/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/webhook/i)).toBeInTheDocument()
  })

  it('updates trigger type filter when checkbox is clicked', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<AdvancedFilters filters={defaultFilters} onChange={onChange} />)

    await user.click(screen.getByText(/advanced filters/i))
    await user.click(screen.getByLabelText(/webhook/i))

    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({
        triggerType: ['webhook'],
      })
    )
  })

  it('shows date range picker', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<AdvancedFilters filters={defaultFilters} onChange={onChange} />)

    await user.click(screen.getByText(/advanced filters/i))

    expect(screen.getByText(/date range/i)).toBeInTheDocument()
  })

  it('shows error search input', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<AdvancedFilters filters={defaultFilters} onChange={onChange} />)

    await user.click(screen.getByText(/advanced filters/i))

    expect(screen.getByPlaceholderText(/search error messages/i)).toBeInTheDocument()
  })

  it('debounces error search input', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<AdvancedFilters filters={defaultFilters} onChange={onChange} />)

    await user.click(screen.getByText(/advanced filters/i))

    const input = screen.getByPlaceholderText(/search error messages/i)
    await user.type(input, 'timeout')

    expect(onChange).not.toHaveBeenCalled()

    await waitFor(
      () => {
        expect(onChange).toHaveBeenCalledWith(
          expect.objectContaining({
            errorSearch: 'timeout',
          })
        )
      },
      { timeout: 500 }
    )
  })

  it('shows execution ID search input', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<AdvancedFilters filters={defaultFilters} onChange={onChange} />)

    await user.click(screen.getByText(/advanced filters/i))

    expect(screen.getByPlaceholderText(/search by execution id/i)).toBeInTheDocument()
  })

  it('updates execution ID filter with debounce', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<AdvancedFilters filters={defaultFilters} onChange={onChange} />)

    await user.click(screen.getByText(/advanced filters/i))

    const input = screen.getByPlaceholderText(/search by execution id/i)
    await user.type(input, 'abc123')

    await waitFor(
      () => {
        expect(onChange).toHaveBeenCalledWith(
          expect.objectContaining({
            executionIdPrefix: 'abc123',
          })
        )
      },
      { timeout: 500 }
    )
  })

  it('shows duration range inputs', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<AdvancedFilters filters={defaultFilters} onChange={onChange} />)

    await user.click(screen.getByText(/advanced filters/i))

    expect(screen.getByPlaceholderText(/min/i)).toBeInTheDocument()
    expect(screen.getByPlaceholderText(/max/i)).toBeInTheDocument()
  })

  it('updates duration range filters', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(<AdvancedFilters filters={defaultFilters} onChange={onChange} />)

    await user.click(screen.getByText(/advanced filters/i))

    const minInput = screen.getByPlaceholderText(/min/i)
    const maxInput = screen.getByPlaceholderText(/max/i)

    await user.type(minInput, '1000')
    await user.type(maxInput, '5000')

    await waitFor(() => {
      expect(onChange).toHaveBeenCalledWith(
        expect.objectContaining({
          minDurationMs: 1000,
          maxDurationMs: 5000,
        })
      )
    })
  })

  it('shows clear all button when filters are active', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(
      <AdvancedFilters
        filters={{ status: ['completed'] }}
        onChange={onChange}
      />
    )

    await user.click(screen.getByText(/advanced filters/i))

    expect(screen.getByText(/clear all/i)).toBeInTheDocument()
  })

  it('clears all filters when clear all button is clicked', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(
      <AdvancedFilters
        filters={{ status: ['completed'], errorSearch: 'error' }}
        onChange={onChange}
      />
    )

    await user.click(screen.getByText(/advanced filters/i))
    await user.click(screen.getByText(/clear all/i))

    expect(onChange).toHaveBeenCalledWith({})
  })

  it('applies filters when apply button is clicked with auto-apply disabled', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    const onApply = vi.fn()

    render(
      <AdvancedFilters
        filters={defaultFilters}
        onChange={onChange}
        onApply={onApply}
        autoApply={false}
      />
    )

    await user.click(screen.getByText(/advanced filters/i))
    await user.click(screen.getByLabelText(/completed/i))

    expect(screen.getByText(/apply/i)).toBeInTheDocument()

    await user.click(screen.getByText(/apply/i))

    expect(onApply).toHaveBeenCalled()
  })

  it('preserves existing filters when adding new ones', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()

    render(
      <AdvancedFilters
        filters={{ status: ['completed'] }}
        onChange={onChange}
      />
    )

    await user.click(screen.getByText(/advanced filters/i))
    await user.click(screen.getByLabelText(/failed/i))

    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({
        status: expect.arrayContaining(['completed', 'failed']),
      })
    )
  })
})
