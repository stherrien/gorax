import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { BulkActionsToolbar } from './BulkActionsToolbar'
import * as useBulkWorkflowsModule from '../../hooks/useBulkWorkflows'

// Mock the useBulkWorkflows hook
vi.mock('../../hooks/useBulkWorkflows', () => ({
  useBulkWorkflows: vi.fn(),
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

describe('BulkActionsToolbar', () => {
  const mockBulkDelete = vi.fn()
  const mockBulkEnable = vi.fn()
  const mockBulkDisable = vi.fn()
  const mockBulkExport = vi.fn()
  const mockBulkClone = vi.fn()
  const mockOnClearSelection = vi.fn()
  const mockOnOperationComplete = vi.fn()

  const defaultProps = {
    selectedCount: 3,
    selectedWorkflowIds: ['wf-1', 'wf-2', 'wf-3'],
    onClearSelection: mockOnClearSelection,
    onOperationComplete: mockOnOperationComplete,
  }

  beforeEach(() => {
    vi.clearAllMocks()

    vi.mocked(useBulkWorkflowsModule.useBulkWorkflows).mockReturnValue({
      bulkDelete: mockBulkDelete,
      bulkEnable: mockBulkEnable,
      bulkDisable: mockBulkDisable,
      bulkExport: mockBulkExport,
      bulkClone: mockBulkClone,
      bulkDeleting: false,
      bulkEnabling: false,
      bulkDisabling: false,
      bulkExporting: false,
      bulkCloning: false,
    })
  })

  describe('rendering', () => {
    it('should render nothing when no items selected', () => {
      const { container } = render(
        <BulkActionsToolbar {...defaultProps} selectedCount={0} selectedWorkflowIds={[]} />
      )

      expect(container.firstChild).toBeNull()
    })

    it('should render toolbar when items are selected', () => {
      render(<BulkActionsToolbar {...defaultProps} />)

      expect(screen.getByText('3')).toBeInTheDocument()
      expect(screen.getByText('workflows selected')).toBeInTheDocument()
    })

    it('should show singular "workflow" when one item selected', () => {
      render(
        <BulkActionsToolbar
          {...defaultProps}
          selectedCount={1}
          selectedWorkflowIds={['wf-1']}
        />
      )

      expect(screen.getByText('1')).toBeInTheDocument()
      expect(screen.getByText('workflow selected')).toBeInTheDocument()
    })

    it('should render all action buttons', () => {
      render(<BulkActionsToolbar {...defaultProps} />)

      expect(screen.getByTitle('Delete selected workflows')).toBeInTheDocument()
      expect(screen.getByTitle('Enable selected workflows')).toBeInTheDocument()
      expect(screen.getByTitle('Disable selected workflows')).toBeInTheDocument()
      expect(screen.getByTitle('Export selected workflows')).toBeInTheDocument()
      expect(screen.getByTitle('Clone selected workflows')).toBeInTheDocument()
    })

    it('should render clear selection button', () => {
      render(<BulkActionsToolbar {...defaultProps} />)

      expect(screen.getByText('Clear selection')).toBeInTheDocument()
    })
  })

  describe('loading states', () => {
    it('should disable all buttons when deleting', () => {
      vi.mocked(useBulkWorkflowsModule.useBulkWorkflows).mockReturnValue({
        bulkDelete: mockBulkDelete,
        bulkEnable: mockBulkEnable,
        bulkDisable: mockBulkDisable,
        bulkExport: mockBulkExport,
        bulkClone: mockBulkClone,
        bulkDeleting: true,
        bulkEnabling: false,
        bulkDisabling: false,
        bulkExporting: false,
        bulkCloning: false,
      })

      render(<BulkActionsToolbar {...defaultProps} />)

      expect(screen.getByTitle('Delete selected workflows')).toBeDisabled()
      expect(screen.getByTitle('Enable selected workflows')).toBeDisabled()
      expect(screen.getByTitle('Disable selected workflows')).toBeDisabled()
      expect(screen.getByTitle('Export selected workflows')).toBeDisabled()
      expect(screen.getByTitle('Clone selected workflows')).toBeDisabled()
    })

    it('should disable all buttons when enabling', () => {
      vi.mocked(useBulkWorkflowsModule.useBulkWorkflows).mockReturnValue({
        bulkDelete: mockBulkDelete,
        bulkEnable: mockBulkEnable,
        bulkDisable: mockBulkDisable,
        bulkExport: mockBulkExport,
        bulkClone: mockBulkClone,
        bulkDeleting: false,
        bulkEnabling: true,
        bulkDisabling: false,
        bulkExporting: false,
        bulkCloning: false,
      })

      render(<BulkActionsToolbar {...defaultProps} />)

      expect(screen.getByTitle('Delete selected workflows')).toBeDisabled()
      expect(screen.getByTitle('Enable selected workflows')).toBeDisabled()
    })

    it('should disable all buttons when disabling', () => {
      vi.mocked(useBulkWorkflowsModule.useBulkWorkflows).mockReturnValue({
        bulkDelete: mockBulkDelete,
        bulkEnable: mockBulkEnable,
        bulkDisable: mockBulkDisable,
        bulkExport: mockBulkExport,
        bulkClone: mockBulkClone,
        bulkDeleting: false,
        bulkEnabling: false,
        bulkDisabling: true,
        bulkExporting: false,
        bulkCloning: false,
      })

      render(<BulkActionsToolbar {...defaultProps} />)

      expect(screen.getByTitle('Delete selected workflows')).toBeDisabled()
      expect(screen.getByTitle('Disable selected workflows')).toBeDisabled()
    })

    it('should disable all buttons when exporting', () => {
      vi.mocked(useBulkWorkflowsModule.useBulkWorkflows).mockReturnValue({
        bulkDelete: mockBulkDelete,
        bulkEnable: mockBulkEnable,
        bulkDisable: mockBulkDisable,
        bulkExport: mockBulkExport,
        bulkClone: mockBulkClone,
        bulkDeleting: false,
        bulkEnabling: false,
        bulkDisabling: false,
        bulkExporting: true,
        bulkCloning: false,
      })

      render(<BulkActionsToolbar {...defaultProps} />)

      expect(screen.getByTitle('Delete selected workflows')).toBeDisabled()
      expect(screen.getByTitle('Export selected workflows')).toBeDisabled()
    })

    it('should disable all buttons when cloning', () => {
      vi.mocked(useBulkWorkflowsModule.useBulkWorkflows).mockReturnValue({
        bulkDelete: mockBulkDelete,
        bulkEnable: mockBulkEnable,
        bulkDisable: mockBulkDisable,
        bulkExport: mockBulkExport,
        bulkClone: mockBulkClone,
        bulkDeleting: false,
        bulkEnabling: false,
        bulkDisabling: false,
        bulkExporting: false,
        bulkCloning: true,
      })

      render(<BulkActionsToolbar {...defaultProps} />)

      expect(screen.getByTitle('Delete selected workflows')).toBeDisabled()
      expect(screen.getByTitle('Clone selected workflows')).toBeDisabled()
    })
  })

  describe('confirmation dialogs', () => {
    it('should open delete confirmation dialog', async () => {
      const user = userEvent.setup()
      render(<BulkActionsToolbar {...defaultProps} />)

      await user.click(screen.getByTitle('Delete selected workflows'))

      expect(screen.getByTestId('confirm-dialog')).toBeInTheDocument()
      expect(screen.getByTestId('dialog-action')).toHaveTextContent('delete')
    })

    it('should open enable confirmation dialog', async () => {
      const user = userEvent.setup()
      render(<BulkActionsToolbar {...defaultProps} />)

      await user.click(screen.getByTitle('Enable selected workflows'))

      expect(screen.getByTestId('confirm-dialog')).toBeInTheDocument()
      expect(screen.getByTestId('dialog-action')).toHaveTextContent('enable')
    })

    it('should open disable confirmation dialog', async () => {
      const user = userEvent.setup()
      render(<BulkActionsToolbar {...defaultProps} />)

      await user.click(screen.getByTitle('Disable selected workflows'))

      expect(screen.getByTestId('confirm-dialog')).toBeInTheDocument()
      expect(screen.getByTestId('dialog-action')).toHaveTextContent('disable')
    })

    it('should open clone confirmation dialog', async () => {
      const user = userEvent.setup()
      render(<BulkActionsToolbar {...defaultProps} />)

      await user.click(screen.getByTitle('Clone selected workflows'))

      expect(screen.getByTestId('confirm-dialog')).toBeInTheDocument()
      expect(screen.getByTestId('dialog-action')).toHaveTextContent('clone')
    })

    it('should close dialog on cancel', async () => {
      const user = userEvent.setup()
      render(<BulkActionsToolbar {...defaultProps} />)

      await user.click(screen.getByTitle('Delete selected workflows'))
      expect(screen.getByTestId('confirm-dialog')).toBeInTheDocument()

      await user.click(screen.getByTestId('cancel-button'))
      expect(screen.queryByTestId('confirm-dialog')).not.toBeInTheDocument()
    })
  })

  describe('bulk operations', () => {
    it('should call bulkDelete when confirmed', async () => {
      const user = userEvent.setup()
      mockBulkDelete.mockResolvedValueOnce({
        success_count: 3,
        failures: [],
      })

      render(<BulkActionsToolbar {...defaultProps} />)

      await user.click(screen.getByTitle('Delete selected workflows'))
      await user.click(screen.getByTestId('confirm-button'))

      await waitFor(() => {
        expect(mockBulkDelete).toHaveBeenCalledWith(['wf-1', 'wf-2', 'wf-3'])
      })
    })

    it('should call bulkEnable when confirmed', async () => {
      const user = userEvent.setup()
      mockBulkEnable.mockResolvedValueOnce({
        success_count: 3,
        failures: [],
      })

      render(<BulkActionsToolbar {...defaultProps} />)

      await user.click(screen.getByTitle('Enable selected workflows'))
      await user.click(screen.getByTestId('confirm-button'))

      await waitFor(() => {
        expect(mockBulkEnable).toHaveBeenCalledWith(['wf-1', 'wf-2', 'wf-3'])
      })
    })

    it('should call bulkDisable when confirmed', async () => {
      const user = userEvent.setup()
      mockBulkDisable.mockResolvedValueOnce({
        success_count: 3,
        failures: [],
      })

      render(<BulkActionsToolbar {...defaultProps} />)

      await user.click(screen.getByTitle('Disable selected workflows'))
      await user.click(screen.getByTestId('confirm-button'))

      await waitFor(() => {
        expect(mockBulkDisable).toHaveBeenCalledWith(['wf-1', 'wf-2', 'wf-3'])
      })
    })

    it('should call bulkExport directly without confirmation', async () => {
      const user = userEvent.setup()
      mockBulkExport.mockResolvedValueOnce({
        result: {
          success_count: 3,
          failures: [],
        },
      })

      render(<BulkActionsToolbar {...defaultProps} />)

      await user.click(screen.getByTitle('Export selected workflows'))

      await waitFor(() => {
        expect(mockBulkExport).toHaveBeenCalledWith(['wf-1', 'wf-2', 'wf-3'])
      })
    })

    it('should call bulkClone when confirmed', async () => {
      const user = userEvent.setup()
      mockBulkClone.mockResolvedValueOnce({
        result: {
          success_count: 3,
          failures: [],
        },
      })

      render(<BulkActionsToolbar {...defaultProps} />)

      await user.click(screen.getByTitle('Clone selected workflows'))
      await user.click(screen.getByTestId('confirm-button'))

      await waitFor(() => {
        expect(mockBulkClone).toHaveBeenCalledWith(['wf-1', 'wf-2', 'wf-3'])
      })
    })
  })

  describe('clear selection', () => {
    it('should call onClearSelection when clear button clicked', async () => {
      const user = userEvent.setup()
      render(<BulkActionsToolbar {...defaultProps} />)

      await user.click(screen.getByText('Clear selection'))

      expect(mockOnClearSelection).toHaveBeenCalled()
    })
  })
})
