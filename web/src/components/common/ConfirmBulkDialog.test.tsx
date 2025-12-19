import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { ConfirmBulkDialog } from './ConfirmBulkDialog'

describe('ConfirmBulkDialog', () => {
  it('should not render when not open', () => {
    render(
      <ConfirmBulkDialog
        open={false}
        action="delete"
        count={5}
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />
    )

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('should render dialog when open', () => {
    render(
      <ConfirmBulkDialog
        open={true}
        action="delete"
        count={5}
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />
    )

    expect(screen.getByRole('dialog')).toBeInTheDocument()
  })

  it('should show action and count in title', () => {
    render(
      <ConfirmBulkDialog
        open={true}
        action="delete"
        count={5}
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />
    )

    expect(screen.getByText(/Delete 5 items/i)).toBeInTheDocument()
  })

  it('should show warning message for destructive actions', () => {
    render(
      <ConfirmBulkDialog
        open={true}
        action="delete"
        count={3}
        destructive={true}
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />
    )

    expect(screen.getByText(/This action cannot be undone/i)).toBeInTheDocument()
  })

  it('should require confirmation text for destructive actions', () => {
    render(
      <ConfirmBulkDialog
        open={true}
        action="delete"
        count={3}
        destructive={true}
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />
    )

    expect(screen.getByPlaceholderText('Type DELETE to confirm')).toBeInTheDocument()
  })

  it('should disable confirm button until correct text is entered', async () => {
    const onConfirm = vi.fn()

    render(
      <ConfirmBulkDialog
        open={true}
        action="delete"
        count={3}
        destructive={true}
        onConfirm={onConfirm}
        onCancel={vi.fn()}
      />
    )

    const confirmButton = screen.getByText('Delete')
    expect(confirmButton).toBeDisabled()

    const input = screen.getByPlaceholderText('Type DELETE to confirm')
    fireEvent.change(input, { target: { value: 'DELETE' } })

    await waitFor(() => {
      expect(confirmButton).not.toBeDisabled()
    })
  })

  it('should call onConfirm when confirmed', () => {
    const onConfirm = vi.fn()

    render(
      <ConfirmBulkDialog
        open={true}
        action="enable"
        count={3}
        onConfirm={onConfirm}
        onCancel={vi.fn()}
      />
    )

    const confirmButton = screen.getByText('Enable')
    fireEvent.click(confirmButton)

    expect(onConfirm).toHaveBeenCalledTimes(1)
  })

  it('should call onCancel when cancelled', () => {
    const onCancel = vi.fn()

    render(
      <ConfirmBulkDialog
        open={true}
        action="delete"
        count={3}
        onConfirm={vi.fn()}
        onCancel={onCancel}
      />
    )

    const cancelButton = screen.getByText('Cancel')
    fireEvent.click(cancelButton)

    expect(onCancel).toHaveBeenCalledTimes(1)
  })

  it('should show custom message when provided', () => {
    render(
      <ConfirmBulkDialog
        open={true}
        action="delete"
        count={3}
        message="Custom warning message"
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />
    )

    expect(screen.getByText('Custom warning message')).toBeInTheDocument()
  })

  it('should capitalize action in button text', () => {
    render(
      <ConfirmBulkDialog
        open={true}
        action="enable"
        count={3}
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />
    )

    expect(screen.getByText('Enable')).toBeInTheDocument()
  })

  it('should not require confirmation text for non-destructive actions', () => {
    render(
      <ConfirmBulkDialog
        open={true}
        action="enable"
        count={3}
        destructive={false}
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />
    )

    expect(screen.queryByPlaceholderText('Type DELETE to confirm')).not.toBeInTheDocument()
  })
})
