import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import FilterChips from './FilterChips'

describe('FilterChips', () => {
  it('renders nothing when no filters are active', () => {
    const onRemove = vi.fn()
    const { container } = render(<FilterChips filters={{}} onRemove={onRemove} />)

    // Component returns null when no filters, so firstChild will be null
    expect(container.firstChild).toBeNull()
  })

  it('renders status filter chip', () => {
    const onRemove = vi.fn()
    render(<FilterChips filters={{ status: ['completed'] }} onRemove={onRemove} />)

    expect(screen.getByText('Status:')).toBeInTheDocument()
    expect(screen.getByText('completed')).toBeInTheDocument()
  })

  it('renders multiple status filter chips', () => {
    const onRemove = vi.fn()
    render(
      <FilterChips
        filters={{ status: ['completed', 'failed'] }}
        onRemove={onRemove}
      />
    )

    expect(screen.getByText('Status:')).toBeInTheDocument()
    expect(screen.getByText('completed, failed')).toBeInTheDocument()
  })

  it('renders workflow filter chip', () => {
    const onRemove = vi.fn()
    render(
      <FilterChips
        filters={{ workflowId: 'workflow-123', workflowName: 'My Workflow' }}
        onRemove={onRemove}
      />
    )

    expect(screen.getByText('Workflow:')).toBeInTheDocument()
    expect(screen.getByText('My Workflow')).toBeInTheDocument()
  })

  it('renders trigger type filter chip', () => {
    const onRemove = vi.fn()
    render(
      <FilterChips
        filters={{ triggerType: ['webhook', 'manual'] }}
        onRemove={onRemove}
      />
    )

    expect(screen.getByText('Trigger:')).toBeInTheDocument()
    expect(screen.getByText('webhook, manual')).toBeInTheDocument()
  })

  it('renders date range filter chip', () => {
    const onRemove = vi.fn()
    // Use explicit time to avoid timezone issues
    const startDate = new Date('2025-01-01T12:00:00Z')
    const endDate = new Date('2025-01-31T12:00:00Z')

    render(
      <FilterChips
        filters={{ startDate, endDate }}
        onRemove={onRemove}
      />
    )

    expect(screen.getByText('Date Range:')).toBeInTheDocument()
    // The format depends on the actual rendering - let's just check it contains Jan
    expect(screen.getByText(/Jan/)).toBeInTheDocument()
  })

  it('renders error search filter chip', () => {
    const onRemove = vi.fn()
    render(
      <FilterChips
        filters={{ errorSearch: 'timeout' }}
        onRemove={onRemove}
      />
    )

    expect(screen.getByText('Error:')).toBeInTheDocument()
    expect(screen.getByText('timeout')).toBeInTheDocument()
  })

  it('renders execution ID filter chip', () => {
    const onRemove = vi.fn()
    render(
      <FilterChips
        filters={{ executionIdPrefix: 'abc123' }}
        onRemove={onRemove}
      />
    )

    expect(screen.getByText('ID:')).toBeInTheDocument()
    expect(screen.getByText('abc123')).toBeInTheDocument()
  })

  it('renders duration range filter chip (min only)', () => {
    const onRemove = vi.fn()
    render(
      <FilterChips
        filters={{ minDurationMs: 1000 }}
        onRemove={onRemove}
      />
    )

    expect(screen.getByText('Duration:')).toBeInTheDocument()
    expect(screen.getByText('>= 1.0s')).toBeInTheDocument()
  })

  it('renders duration range filter chip (max only)', () => {
    const onRemove = vi.fn()
    render(
      <FilterChips
        filters={{ maxDurationMs: 5000 }}
        onRemove={onRemove}
      />
    )

    expect(screen.getByText('Duration:')).toBeInTheDocument()
    expect(screen.getByText('<= 5.0s')).toBeInTheDocument()
  })

  it('renders duration range filter chip (min and max)', () => {
    const onRemove = vi.fn()
    render(
      <FilterChips
        filters={{ minDurationMs: 1000, maxDurationMs: 5000 }}
        onRemove={onRemove}
      />
    )

    expect(screen.getByText('Duration:')).toBeInTheDocument()
    expect(screen.getByText('1.0s - 5.0s')).toBeInTheDocument()
  })

  it('calls onRemove with correct key when chip is clicked', async () => {
    const user = userEvent.setup()
    const onRemove = vi.fn()

    render(
      <FilterChips
        filters={{ status: ['completed'] }}
        onRemove={onRemove}
      />
    )

    await user.click(screen.getByRole('button'))

    expect(onRemove).toHaveBeenCalledWith('status')
  })

  it('calls onRemove for each filter chip independently', async () => {
    const user = userEvent.setup()
    const onRemove = vi.fn()

    render(
      <FilterChips
        filters={{
          status: ['completed'],
          triggerType: ['webhook'],
          errorSearch: 'timeout',
        }}
        onRemove={onRemove}
      />
    )

    const buttons = screen.getAllByRole('button')
    expect(buttons).toHaveLength(3)

    await user.click(buttons[0])
    expect(onRemove).toHaveBeenCalledWith('status')

    await user.click(buttons[1])
    expect(onRemove).toHaveBeenCalledWith('triggerType')

    await user.click(buttons[2])
    expect(onRemove).toHaveBeenCalledWith('errorSearch')
  })

  it('displays result count when provided', () => {
    const onRemove = vi.fn()
    render(
      <FilterChips
        filters={{ status: ['completed'] }}
        onRemove={onRemove}
        resultCount={42}
      />
    )

    expect(screen.getByText(/42 results/i)).toBeInTheDocument()
  })

  it('displays singular "result" when count is 1', () => {
    const onRemove = vi.fn()
    render(
      <FilterChips
        filters={{ status: ['completed'] }}
        onRemove={onRemove}
        resultCount={1}
      />
    )

    expect(screen.getByText(/1 result$/i)).toBeInTheDocument()
  })

  it('renders all filters at once', () => {
    const onRemove = vi.fn()
    render(
      <FilterChips
        filters={{
          status: ['completed', 'failed'],
          workflowId: 'wf-1',
          workflowName: 'Test Workflow',
          triggerType: ['webhook'],
          startDate: new Date('2025-01-01T12:00:00Z'),
          endDate: new Date('2025-01-31T12:00:00Z'),
          errorSearch: 'error text',
          executionIdPrefix: 'exec-123',
          minDurationMs: 1000,
          maxDurationMs: 5000,
        }}
        onRemove={onRemove}
        resultCount={10}
      />
    )

    expect(screen.getByText('Status:')).toBeInTheDocument()
    expect(screen.getByText('Workflow:')).toBeInTheDocument()
    expect(screen.getByText('Trigger:')).toBeInTheDocument()
    expect(screen.getByText('Date Range:')).toBeInTheDocument()
    expect(screen.getByText('Error:')).toBeInTheDocument()
    expect(screen.getByText('ID:')).toBeInTheDocument()
    expect(screen.getByText('Duration:')).toBeInTheDocument()
    expect(screen.getByText('10 results')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const onRemove = vi.fn()
    const { container } = render(
      <FilterChips
        filters={{ status: ['completed'] }}
        onRemove={onRemove}
        className="custom-class"
      />
    )

    expect(container.firstChild).toHaveClass('custom-class')
  })
})
