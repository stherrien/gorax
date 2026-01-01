import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { BulkExecutionActions } from './BulkExecutionActions'
import * as executionsModule from '../../api/executions'

// Mock the executions API
vi.mock('../../api/executions', () => ({
  executionAPI: {
    bulkDelete: vi.fn(),
    bulkRetry: vi.fn(),
  },
}))

// Mock the ConfirmBulkDialog component
vi.mock('../common/ConfirmBulkDialog', () => ({
  ConfirmBulkDialog: ({
    open,
    action,
    onConfirm,
    onCancel,
  }: {
    open: boolean
    action: string
    count: number
    destructive?: boolean
    message?: string
    onConfirm: () => void
    onCancel: () => void
  }) =>
    open ? (
      <div data-testid="confirm-dialog">
        <span data-testid="dialog-action">{action}</span>
        <button data-testid="confirm-button" onClick={onConfirm}>
          Confirm
        </button>
        <button data-testid="cancel-button" onClick={onCancel}>
          Cancel
        </button>
      </div>
    ) : null,
}))

describe('BulkExecutionActions', () => {
  const mockOnSuccess = vi.fn()
  const mockOnError = vi.fn()

  const defaultProps = {
    selectedIds: ['exec-1', 'exec-2', 'exec-3'],
    onSuccess: mockOnSuccess,
    onError: mockOnError,
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('rendering', () => {
    it('should render Retry Failed button', () => {
      render(<BulkExecutionActions {...defaultProps} />)

      expect(screen.getByText('Retry Failed')).toBeInTheDocument()
    })

    it('should render Delete button', () => {
      render(<BulkExecutionActions {...defaultProps} />)

      expect(screen.getByText('Delete')).toBeInTheDocument()
    })

    it('should not show any dialog initially', () => {
      render(<BulkExecutionActions {...defaultProps} />)

      expect(screen.queryByTestId('confirm-dialog')).not.toBeInTheDocument()
    })
  })

  describe('delete dialog', () => {
    it('should open delete dialog when Delete button is clicked', async () => {
      const user = userEvent.setup()
      render(<BulkExecutionActions {...defaultProps} />)

      await user.click(screen.getByText('Delete'))

      expect(screen.getByTestId('confirm-dialog')).toBeInTheDocument()
      expect(screen.getByTestId('dialog-action')).toHaveTextContent('delete')
    })

    it('should close delete dialog when cancel is clicked', async () => {
      const user = userEvent.setup()
      render(<BulkExecutionActions {...defaultProps} />)

      await user.click(screen.getByText('Delete'))
      expect(screen.getByTestId('confirm-dialog')).toBeInTheDocument()

      await user.click(screen.getByTestId('cancel-button'))
      expect(screen.queryByTestId('confirm-dialog')).not.toBeInTheDocument()
    })

    it('should call bulkDelete and onSuccess when all deletions succeed', async () => {
      const user = userEvent.setup()
      vi.mocked(executionsModule.executionAPI.bulkDelete).mockResolvedValueOnce({
        success: ['exec-1', 'exec-2', 'exec-3'],
        failed: [],
      })

      render(<BulkExecutionActions {...defaultProps} />)

      await user.click(screen.getByText('Delete'))
      await user.click(screen.getByTestId('confirm-button'))

      await waitFor(() => {
        expect(executionsModule.executionAPI.bulkDelete).toHaveBeenCalledWith([
          'exec-1',
          'exec-2',
          'exec-3',
        ])
      })

      await waitFor(() => {
        expect(mockOnSuccess).toHaveBeenCalled()
      })

      expect(mockOnError).not.toHaveBeenCalled()
    })

    it('should call onError when some deletions fail', async () => {
      const user = userEvent.setup()
      vi.mocked(executionsModule.executionAPI.bulkDelete).mockResolvedValueOnce({
        success: ['exec-1', 'exec-2'],
        failed: ['exec-3'],
      })

      render(<BulkExecutionActions {...defaultProps} />)

      await user.click(screen.getByText('Delete'))
      await user.click(screen.getByTestId('confirm-button'))

      await waitFor(() => {
        expect(mockOnError).toHaveBeenCalledWith(
          'Deleted 2 executions. Failed to delete 1 executions.'
        )
      })

      expect(mockOnSuccess).not.toHaveBeenCalled()
    })

    it('should call onError when API throws an error', async () => {
      const user = userEvent.setup()
      vi.mocked(executionsModule.executionAPI.bulkDelete).mockRejectedValueOnce(
        new Error('Network error')
      )

      render(<BulkExecutionActions {...defaultProps} />)

      await user.click(screen.getByText('Delete'))
      await user.click(screen.getByTestId('confirm-button'))

      await waitFor(() => {
        expect(mockOnError).toHaveBeenCalledWith('Network error')
      })
    })

    it('should call onError with generic message for non-Error objects', async () => {
      const user = userEvent.setup()
      vi.mocked(executionsModule.executionAPI.bulkDelete).mockRejectedValueOnce(
        'Unknown error'
      )

      render(<BulkExecutionActions {...defaultProps} />)

      await user.click(screen.getByText('Delete'))
      await user.click(screen.getByTestId('confirm-button'))

      await waitFor(() => {
        expect(mockOnError).toHaveBeenCalledWith('Failed to delete executions')
      })
    })
  })

  describe('retry dialog', () => {
    it('should open retry dialog when Retry Failed button is clicked', async () => {
      const user = userEvent.setup()
      render(<BulkExecutionActions {...defaultProps} />)

      await user.click(screen.getByText('Retry Failed'))

      expect(screen.getByTestId('confirm-dialog')).toBeInTheDocument()
      expect(screen.getByTestId('dialog-action')).toHaveTextContent('retry')
    })

    it('should close retry dialog when cancel is clicked', async () => {
      const user = userEvent.setup()
      render(<BulkExecutionActions {...defaultProps} />)

      await user.click(screen.getByText('Retry Failed'))
      expect(screen.getByTestId('confirm-dialog')).toBeInTheDocument()

      await user.click(screen.getByTestId('cancel-button'))
      expect(screen.queryByTestId('confirm-dialog')).not.toBeInTheDocument()
    })

    it('should call bulkRetry and onSuccess when all retries succeed', async () => {
      const user = userEvent.setup()
      vi.mocked(executionsModule.executionAPI.bulkRetry).mockResolvedValueOnce({
        success: ['exec-1', 'exec-2', 'exec-3'],
        failed: [],
      })

      render(<BulkExecutionActions {...defaultProps} />)

      await user.click(screen.getByText('Retry Failed'))
      await user.click(screen.getByTestId('confirm-button'))

      await waitFor(() => {
        expect(executionsModule.executionAPI.bulkRetry).toHaveBeenCalledWith([
          'exec-1',
          'exec-2',
          'exec-3',
        ])
      })

      await waitFor(() => {
        expect(mockOnSuccess).toHaveBeenCalled()
      })

      expect(mockOnError).not.toHaveBeenCalled()
    })

    it('should call onError when some retries fail', async () => {
      const user = userEvent.setup()
      vi.mocked(executionsModule.executionAPI.bulkRetry).mockResolvedValueOnce({
        success: ['exec-1'],
        failed: ['exec-2', 'exec-3'],
      })

      render(<BulkExecutionActions {...defaultProps} />)

      await user.click(screen.getByText('Retry Failed'))
      await user.click(screen.getByTestId('confirm-button'))

      await waitFor(() => {
        expect(mockOnError).toHaveBeenCalledWith(
          'Retried 1 executions. Failed to retry 2 executions.'
        )
      })

      expect(mockOnSuccess).not.toHaveBeenCalled()
    })

    it('should call onError when API throws an error', async () => {
      const user = userEvent.setup()
      vi.mocked(executionsModule.executionAPI.bulkRetry).mockRejectedValueOnce(
        new Error('Server error')
      )

      render(<BulkExecutionActions {...defaultProps} />)

      await user.click(screen.getByText('Retry Failed'))
      await user.click(screen.getByTestId('confirm-button'))

      await waitFor(() => {
        expect(mockOnError).toHaveBeenCalledWith('Server error')
      })
    })

    it('should call onError with generic message for non-Error objects', async () => {
      const user = userEvent.setup()
      vi.mocked(executionsModule.executionAPI.bulkRetry).mockRejectedValueOnce(
        'Unknown error'
      )

      render(<BulkExecutionActions {...defaultProps} />)

      await user.click(screen.getByText('Retry Failed'))
      await user.click(screen.getByTestId('confirm-button'))

      await waitFor(() => {
        expect(mockOnError).toHaveBeenCalledWith('Failed to retry executions')
      })
    })
  })

  describe('disabled state', () => {
    it('should disable buttons during delete processing', async () => {
      const user = userEvent.setup()
      // Use a promise that we control to keep the loading state
      let resolveDelete: (value: { success: string[]; failed: string[] }) => void
      vi.mocked(executionsModule.executionAPI.bulkDelete).mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveDelete = resolve
          })
      )

      render(<BulkExecutionActions {...defaultProps} />)

      await user.click(screen.getByText('Delete'))
      await user.click(screen.getByTestId('confirm-button'))

      // While processing, buttons should be disabled
      await waitFor(() => {
        expect(screen.getByText('Retry Failed')).toBeDisabled()
        expect(screen.getByText('Delete')).toBeDisabled()
      })

      // Resolve the promise
      resolveDelete!({ success: ['exec-1', 'exec-2', 'exec-3'], failed: [] })

      // After processing, buttons should be enabled again
      await waitFor(() => {
        expect(screen.getByText('Retry Failed')).not.toBeDisabled()
        expect(screen.getByText('Delete')).not.toBeDisabled()
      })
    })
  })
})
