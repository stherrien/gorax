import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { LogExportButton } from './LogExportButton'
import * as workflowsModule from '../../api/workflows'

// Mock the workflows API
vi.mock('../../api/workflows', () => ({
  workflowAPI: {
    exportLogs: vi.fn(),
  },
}))

describe('LogExportButton', () => {
  const defaultProps = {
    executionId: 'exec-123',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('rendering', () => {
    it('should render export button', () => {
      render(<LogExportButton {...defaultProps} />)

      expect(screen.getByRole('button', { name: /export logs/i })).toBeInTheDocument()
      expect(screen.getByText('Export')).toBeInTheDocument()
    })

    it('should not show dropdown initially', () => {
      render(<LogExportButton {...defaultProps} />)

      expect(screen.queryByRole('menu')).not.toBeInTheDocument()
    })

    it('should apply custom className', () => {
      const { container } = render(
        <LogExportButton {...defaultProps} className="custom-class" />
      )

      const wrapper = container.querySelector('.log-export-button-container')
      expect(wrapper?.className).toContain('custom-class')
    })
  })

  describe('dropdown behavior', () => {
    it('should open dropdown when button is clicked', async () => {
      const user = userEvent.setup()
      render(<LogExportButton {...defaultProps} />)

      await user.click(screen.getByRole('button', { name: /export logs/i }))

      expect(screen.getByRole('menu')).toBeInTheDocument()
    })

    it('should close dropdown when button is clicked again', async () => {
      const user = userEvent.setup()
      render(<LogExportButton {...defaultProps} />)

      // Open dropdown
      await user.click(screen.getByRole('button', { name: /export logs/i }))
      expect(screen.getByRole('menu')).toBeInTheDocument()

      // Close dropdown
      await user.click(screen.getByRole('button', { name: /export logs/i }))
      expect(screen.queryByRole('menu')).not.toBeInTheDocument()
    })

    it('should show all export format options', async () => {
      const user = userEvent.setup()
      render(<LogExportButton {...defaultProps} />)

      await user.click(screen.getByRole('button', { name: /export logs/i }))

      expect(screen.getByText('Text (.txt)')).toBeInTheDocument()
      expect(screen.getByText('JSON (.json)')).toBeInTheDocument()
      expect(screen.getByText('CSV (.csv)')).toBeInTheDocument()
    })

    it('should show descriptions for each format', async () => {
      const user = userEvent.setup()
      render(<LogExportButton {...defaultProps} />)

      await user.click(screen.getByRole('button', { name: /export logs/i }))

      expect(screen.getByText('Human-readable format')).toBeInTheDocument()
      expect(screen.getByText('Structured data format')).toBeInTheDocument()
      expect(screen.getByText('Spreadsheet format')).toBeInTheDocument()
    })

    it('should have proper aria attributes', async () => {
      const user = userEvent.setup()
      render(<LogExportButton {...defaultProps} />)

      const button = screen.getByRole('button', { name: /export logs/i })
      expect(button).toHaveAttribute('aria-haspopup', 'true')
      expect(button).toHaveAttribute('aria-expanded', 'false')

      await user.click(button)

      expect(button).toHaveAttribute('aria-expanded', 'true')
    })
  })

  describe('export actions', () => {
    it('should call exportLogs with txt format', async () => {
      const user = userEvent.setup()
      vi.mocked(workflowsModule.workflowAPI.exportLogs).mockResolvedValueOnce(undefined)

      render(<LogExportButton {...defaultProps} />)

      await user.click(screen.getByRole('button', { name: /export logs/i }))
      await user.click(screen.getByText('Text (.txt)'))

      await waitFor(() => {
        expect(workflowsModule.workflowAPI.exportLogs).toHaveBeenCalledWith(
          'exec-123',
          'txt'
        )
      })
    })

    it('should call exportLogs with json format', async () => {
      const user = userEvent.setup()
      vi.mocked(workflowsModule.workflowAPI.exportLogs).mockResolvedValueOnce(undefined)

      render(<LogExportButton {...defaultProps} />)

      await user.click(screen.getByRole('button', { name: /export logs/i }))
      await user.click(screen.getByText('JSON (.json)'))

      await waitFor(() => {
        expect(workflowsModule.workflowAPI.exportLogs).toHaveBeenCalledWith(
          'exec-123',
          'json'
        )
      })
    })

    it('should call exportLogs with csv format', async () => {
      const user = userEvent.setup()
      vi.mocked(workflowsModule.workflowAPI.exportLogs).mockResolvedValueOnce(undefined)

      render(<LogExportButton {...defaultProps} />)

      await user.click(screen.getByRole('button', { name: /export logs/i }))
      await user.click(screen.getByText('CSV (.csv)'))

      await waitFor(() => {
        expect(workflowsModule.workflowAPI.exportLogs).toHaveBeenCalledWith(
          'exec-123',
          'csv'
        )
      })
    })

    it('should close dropdown when format is selected', async () => {
      const user = userEvent.setup()
      vi.mocked(workflowsModule.workflowAPI.exportLogs).mockResolvedValueOnce(undefined)

      render(<LogExportButton {...defaultProps} />)

      await user.click(screen.getByRole('button', { name: /export logs/i }))
      expect(screen.getByRole('menu')).toBeInTheDocument()

      await user.click(screen.getByText('Text (.txt)'))

      await waitFor(() => {
        expect(screen.queryByRole('menu')).not.toBeInTheDocument()
      })
    })
  })

  describe('loading state', () => {
    it('should show loading state during export', async () => {
      const user = userEvent.setup()
      let resolveExport: () => void
      vi.mocked(workflowsModule.workflowAPI.exportLogs).mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveExport = resolve
          })
      )

      render(<LogExportButton {...defaultProps} />)

      await user.click(screen.getByRole('button', { name: /export logs/i }))
      await user.click(screen.getByText('Text (.txt)'))

      // While loading, should show "Exporting..." text
      await waitFor(() => {
        expect(screen.getByText('Exporting...')).toBeInTheDocument()
      })

      // Button should be disabled
      expect(screen.getByRole('button', { name: /export logs/i })).toBeDisabled()

      // Dropdown should not be visible
      expect(screen.queryByRole('menu')).not.toBeInTheDocument()

      // Resolve the export
      resolveExport!()

      // After loading, should return to normal state
      await waitFor(() => {
        expect(screen.getByText('Export')).toBeInTheDocument()
      })
    })
  })

  describe('error handling', () => {
    it('should show error message when export fails', async () => {
      const user = userEvent.setup()
      vi.mocked(workflowsModule.workflowAPI.exportLogs).mockRejectedValueOnce(
        new Error('Export failed')
      )

      render(<LogExportButton {...defaultProps} />)

      await user.click(screen.getByRole('button', { name: /export logs/i }))
      await user.click(screen.getByText('Text (.txt)'))

      await waitFor(() => {
        expect(screen.getByRole('alert')).toBeInTheDocument()
        expect(screen.getByText('Export failed')).toBeInTheDocument()
      })
    })

    it('should show generic error for non-Error objects', async () => {
      const user = userEvent.setup()
      vi.mocked(workflowsModule.workflowAPI.exportLogs).mockRejectedValueOnce(
        'Unknown error'
      )

      render(<LogExportButton {...defaultProps} />)

      await user.click(screen.getByRole('button', { name: /export logs/i }))
      await user.click(screen.getByText('Text (.txt)'))

      await waitFor(() => {
        expect(screen.getByRole('alert')).toBeInTheDocument()
        expect(screen.getByText('Failed to export logs')).toBeInTheDocument()
      })
    })

// Note: Error auto-clear after 5s timeout tested manually - fake timers conflict with async userEvent
  })
})
