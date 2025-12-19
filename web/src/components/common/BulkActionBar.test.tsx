import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { BulkActionBar } from './BulkActionBar'

describe('BulkActionBar', () => {
  it('should not render when count is 0', () => {
    const { container } = render(
      <BulkActionBar count={0} onClear={vi.fn()}>
        <button>Action</button>
      </BulkActionBar>
    )

    expect(container.firstChild).toBeNull()
  })

  it('should render when count is greater than 0', () => {
    render(
      <BulkActionBar count={5} onClear={vi.fn()}>
        <button>Action</button>
      </BulkActionBar>
    )

    expect(screen.getByText('5 items selected')).toBeInTheDocument()
  })

  it('should show singular text for 1 item', () => {
    render(
      <BulkActionBar count={1} onClear={vi.fn()}>
        <button>Action</button>
      </BulkActionBar>
    )

    expect(screen.getByText('1 item selected')).toBeInTheDocument()
  })

  it('should render action buttons', () => {
    render(
      <BulkActionBar count={3} onClear={vi.fn()}>
        <button>Delete</button>
        <button>Export</button>
      </BulkActionBar>
    )

    expect(screen.getByText('Delete')).toBeInTheDocument()
    expect(screen.getByText('Export')).toBeInTheDocument()
  })

  it('should call onClear when clear button is clicked', () => {
    const onClear = vi.fn()

    render(
      <BulkActionBar count={5} onClear={onClear}>
        <button>Action</button>
      </BulkActionBar>
    )

    const clearButton = screen.getByText('Clear selection')
    fireEvent.click(clearButton)

    expect(onClear).toHaveBeenCalledTimes(1)
  })

  it('should render in sticky position', () => {
    const { container } = render(
      <BulkActionBar count={5} onClear={vi.fn()}>
        <button>Action</button>
      </BulkActionBar>
    )

    const bar = container.firstChild as HTMLElement
    expect(bar).toHaveClass('sticky')
  })

  it('should have proper z-index for layering', () => {
    const { container } = render(
      <BulkActionBar count={5} onClear={vi.fn()}>
        <button>Action</button>
      </BulkActionBar>
    )

    const bar = container.firstChild as HTMLElement
    expect(bar).toHaveClass('z-10')
  })
})
