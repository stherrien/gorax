import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import VersionHistory from './VersionHistory'
import * as workflowsModule from '../../api/workflows'

// Mock the workflows API
vi.mock('../../api/workflows', () => ({
  workflowAPI: {
    listVersions: vi.fn(),
    restoreVersion: vi.fn(),
  },
}))

// Mock window.confirm
const mockConfirm = vi.fn()
Object.defineProperty(window, 'confirm', {
  value: mockConfirm,
  writable: true,
})

describe('VersionHistory', () => {
  const mockVersions = [
    {
      id: 'ver-3',
      version: 3,
      workflowId: 'wf-123',
      createdAt: new Date().toISOString(),
      definition: {
        nodes: [{ id: 'n1' }, { id: 'n2' }, { id: 'n3' }],
        edges: [{ id: 'e1' }, { id: 'e2' }],
      },
    },
    {
      id: 'ver-2',
      version: 2,
      workflowId: 'wf-123',
      createdAt: new Date(Date.now() - 86400000).toISOString(), // 1 day ago
      definition: {
        nodes: [{ id: 'n1' }, { id: 'n2' }],
        edges: [{ id: 'e1' }],
      },
    },
    {
      id: 'ver-1',
      version: 1,
      workflowId: 'wf-123',
      createdAt: new Date(Date.now() - 172800000).toISOString(), // 2 days ago
      definition: {
        nodes: [{ id: 'n1' }],
        edges: [],
      },
    },
  ]

  const defaultProps = {
    workflowId: 'wf-123',
    currentVersion: 3,
  }

  beforeEach(() => {
    vi.clearAllMocks()
    mockConfirm.mockReturnValue(true)
  })

  describe('loading state', () => {
    it('should show loading message while fetching versions', () => {
      vi.mocked(workflowsModule.workflowAPI.listVersions).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      )

      render(<VersionHistory {...defaultProps} />)

      expect(screen.getByText('Loading version history...')).toBeInTheDocument()
    })
  })

  describe('error state', () => {
    it('should show error message when API fails', async () => {
      vi.mocked(workflowsModule.workflowAPI.listVersions).mockRejectedValueOnce(
        new Error('Failed to fetch')
      )

      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByText('Failed to fetch')).toBeInTheDocument()
      })
    })

    it('should show generic error for non-Error objects', async () => {
      vi.mocked(workflowsModule.workflowAPI.listVersions).mockRejectedValueOnce(
        'Unknown error'
      )

      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByText('Failed to load versions')).toBeInTheDocument()
      })
    })

    it('should show retry button on error', async () => {
      vi.mocked(workflowsModule.workflowAPI.listVersions).mockRejectedValueOnce(
        new Error('Network error')
      )

      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument()
      })
    })

    it('should retry loading when retry button is clicked', async () => {
      const user = userEvent.setup()
      vi.mocked(workflowsModule.workflowAPI.listVersions)
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce(mockVersions)

      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /retry/i }))

      await waitFor(() => {
        expect(screen.getByText('Version History')).toBeInTheDocument()
        expect(screen.getByText('Version 3')).toBeInTheDocument()
      })
    })
  })

  describe('empty state', () => {
    it('should show empty message when no versions', async () => {
      vi.mocked(workflowsModule.workflowAPI.listVersions).mockResolvedValueOnce([])

      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByText('No version history available')).toBeInTheDocument()
      })
    })
  })

  describe('version list', () => {
    beforeEach(() => {
      vi.mocked(workflowsModule.workflowAPI.listVersions).mockResolvedValue(mockVersions)
    })

    it('should render version history header', async () => {
      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByText('Version History')).toBeInTheDocument()
      })
    })

    it('should render all versions', async () => {
      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByText('Version 3')).toBeInTheDocument()
        expect(screen.getByText('Version 2')).toBeInTheDocument()
        expect(screen.getByText('Version 1')).toBeInTheDocument()
      })
    })

    it('should show Current badge for current version', async () => {
      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByText('Current')).toBeInTheDocument()
      })
    })

    it('should show node and edge counts', async () => {
      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByText('3 nodes, 2 edges')).toBeInTheDocument()
        expect(screen.getByText('2 nodes, 1 edges')).toBeInTheDocument()
        expect(screen.getByText('1 nodes, 0 edges')).toBeInTheDocument()
      })
    })

    it('should show Preview button for all versions', async () => {
      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        const previewButtons = screen.getAllByText('Preview')
        expect(previewButtons).toHaveLength(3)
      })
    })

    it('should show Restore button only for non-current versions', async () => {
      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        const restoreButtons = screen.getAllByText('Restore')
        expect(restoreButtons).toHaveLength(2) // Version 1 and 2, not 3
      })
    })

    it('should show close button when onClose is provided', async () => {
      const onClose = vi.fn()
      render(<VersionHistory {...defaultProps} onClose={onClose} />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /close/i })).toBeInTheDocument()
      })
    })

    it('should call onClose when close button is clicked', async () => {
      const user = userEvent.setup()
      const onClose = vi.fn()
      render(<VersionHistory {...defaultProps} onClose={onClose} />)

      await waitFor(() => {
        expect(screen.getByText('Version History')).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /close/i }))
      expect(onClose).toHaveBeenCalledTimes(1)
    })
  })

  describe('restore functionality', () => {
    beforeEach(() => {
      vi.mocked(workflowsModule.workflowAPI.listVersions).mockResolvedValue(mockVersions)
    })

    it('should show confirmation dialog before restore', async () => {
      const user = userEvent.setup()
      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByText('Version 2')).toBeInTheDocument()
      })

      const restoreButtons = screen.getAllByText('Restore')
      await user.click(restoreButtons[0]) // Restore version 2

      expect(mockConfirm).toHaveBeenCalled()
    })

    it('should not restore when confirmation is cancelled', async () => {
      const user = userEvent.setup()
      mockConfirm.mockReturnValue(false)
      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByText('Version 2')).toBeInTheDocument()
      })

      const restoreButtons = screen.getAllByText('Restore')
      await user.click(restoreButtons[0])

      expect(workflowsModule.workflowAPI.restoreVersion).not.toHaveBeenCalled()
    })

    it('should call restoreVersion API when confirmed', async () => {
      const user = userEvent.setup()
      vi.mocked(workflowsModule.workflowAPI.restoreVersion).mockResolvedValueOnce(undefined)

      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByText('Version 2')).toBeInTheDocument()
      })

      const restoreButtons = screen.getAllByText('Restore')
      await user.click(restoreButtons[0])

      await waitFor(() => {
        expect(workflowsModule.workflowAPI.restoreVersion).toHaveBeenCalledWith(
          'wf-123',
          2
        )
      })
    })

    it('should call onRestore callback after successful restore', async () => {
      const user = userEvent.setup()
      const onRestore = vi.fn()
      vi.mocked(workflowsModule.workflowAPI.restoreVersion).mockResolvedValueOnce(undefined)

      render(<VersionHistory {...defaultProps} onRestore={onRestore} />)

      await waitFor(() => {
        expect(screen.getByText('Version 2')).toBeInTheDocument()
      })

      const restoreButtons = screen.getAllByText('Restore')
      await user.click(restoreButtons[0])

      await waitFor(() => {
        expect(onRestore).toHaveBeenCalledWith(2)
      })
    })

    it('should show error when restore fails', async () => {
      const user = userEvent.setup()
      vi.mocked(workflowsModule.workflowAPI.restoreVersion).mockRejectedValueOnce(
        new Error('Restore failed')
      )

      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByText('Version 2')).toBeInTheDocument()
      })

      const restoreButtons = screen.getAllByText('Restore')
      await user.click(restoreButtons[0])

      await waitFor(() => {
        expect(screen.getByText('Restore failed')).toBeInTheDocument()
      })
    })

    it('should show Restoring... text during restore', async () => {
      const user = userEvent.setup()
      let resolveRestore: () => void
      vi.mocked(workflowsModule.workflowAPI.restoreVersion).mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveRestore = resolve
          })
      )

      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByText('Version 2')).toBeInTheDocument()
      })

      const restoreButtons = screen.getAllByText('Restore')
      await user.click(restoreButtons[0])

      await waitFor(() => {
        expect(screen.getByText('Restoring...')).toBeInTheDocument()
      })

      resolveRestore!()
    })
  })

  describe('preview modal', () => {
    beforeEach(() => {
      vi.mocked(workflowsModule.workflowAPI.listVersions).mockResolvedValue(mockVersions)
    })

    it('should open preview modal when Preview is clicked', async () => {
      const user = userEvent.setup()
      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByText('Version 2')).toBeInTheDocument()
      })

      const previewButtons = screen.getAllByText('Preview')
      await user.click(previewButtons[1]) // Preview version 2

      await waitFor(() => {
        expect(screen.getByText('Version 2 Preview')).toBeInTheDocument()
      })
    })

    it('should show version definition in preview', async () => {
      const user = userEvent.setup()
      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByText('Version 2')).toBeInTheDocument()
      })

      const previewButtons = screen.getAllByText('Preview')
      await user.click(previewButtons[1])

      await waitFor(() => {
        // Should show the JSON definition
        expect(screen.getByText(/"nodes"/)).toBeInTheDocument()
      })
    })

    it('should close preview modal when Close button is clicked', async () => {
      const user = userEvent.setup()
      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByText('Version 2')).toBeInTheDocument()
      })

      const previewButtons = screen.getAllByText('Preview')
      await user.click(previewButtons[1])

      await waitFor(() => {
        expect(screen.getByText('Version 2 Preview')).toBeInTheDocument()
      })

      // Find the Close button in the modal footer
      const closeButtons = screen.getAllByText('Close')
      await user.click(closeButtons[closeButtons.length - 1])

      await waitFor(() => {
        expect(screen.queryByText('Version 2 Preview')).not.toBeInTheDocument()
      })
    })

    it('should close preview modal when clicking backdrop', async () => {
      const user = userEvent.setup()
      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByText('Version 2')).toBeInTheDocument()
      })

      const previewButtons = screen.getAllByText('Preview')
      await user.click(previewButtons[1])

      await waitFor(() => {
        expect(screen.getByText('Version 2 Preview')).toBeInTheDocument()
      })

      // Click the backdrop (the outer fixed div)
      const backdrop = document.querySelector('.fixed.inset-0')
      if (backdrop) {
        await user.click(backdrop)
      }

      await waitFor(() => {
        expect(screen.queryByText('Version 2 Preview')).not.toBeInTheDocument()
      })
    })

    it('should show Restore button in preview for non-current versions', async () => {
      const user = userEvent.setup()
      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByText('Version 2')).toBeInTheDocument()
      })

      const previewButtons = screen.getAllByText('Preview')
      await user.click(previewButtons[1]) // Preview version 2

      await waitFor(() => {
        expect(screen.getByText('Restore This Version')).toBeInTheDocument()
      })
    })

    it('should not show Restore button in preview for current version', async () => {
      const user = userEvent.setup()
      render(<VersionHistory {...defaultProps} />)

      await waitFor(() => {
        expect(screen.getByText('Version 3')).toBeInTheDocument()
      })

      const previewButtons = screen.getAllByText('Preview')
      await user.click(previewButtons[0]) // Preview version 3 (current)

      await waitFor(() => {
        expect(screen.getByText('Version 3 Preview')).toBeInTheDocument()
        expect(screen.queryByText('Restore This Version')).not.toBeInTheDocument()
      })
    })
  })
})
