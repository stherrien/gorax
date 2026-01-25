import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import WorkflowComparison from './WorkflowComparison'
import type { WorkflowVersion } from '../../../api/workflows'

describe('WorkflowComparison', () => {
  const mockVersions: WorkflowVersion[] = [
    {
      id: 'ver-3',
      workflowId: 'wf-123',
      version: 3,
      createdAt: new Date().toISOString(),
      createdBy: 'user1',
      definition: {
        nodes: [
          { id: 'n1', type: 'action', position: { x: 0, y: 0 }, data: { label: 'Node 1' } },
          { id: 'n2', type: 'action', position: { x: 100, y: 0 }, data: { label: 'Node 2' } },
          { id: 'n3', type: 'action', position: { x: 200, y: 0 }, data: { label: 'New Node' } },
        ],
        edges: [
          { id: 'e1', source: 'n1', target: 'n2' },
          { id: 'e2', source: 'n2', target: 'n3' },
        ],
      },
    },
    {
      id: 'ver-2',
      workflowId: 'wf-123',
      version: 2,
      createdAt: new Date(Date.now() - 86400000).toISOString(),
      createdBy: 'user1',
      definition: {
        nodes: [
          { id: 'n1', type: 'action', position: { x: 0, y: 0 }, data: { label: 'Node 1' } },
          { id: 'n2', type: 'action', position: { x: 100, y: 0 }, data: { label: 'Modified' } },
        ],
        edges: [
          { id: 'e1', source: 'n1', target: 'n2' },
        ],
      },
    },
    {
      id: 'ver-1',
      workflowId: 'wf-123',
      version: 1,
      createdAt: new Date(Date.now() - 172800000).toISOString(),
      createdBy: 'user1',
      definition: {
        nodes: [
          { id: 'n1', type: 'action', position: { x: 0, y: 0 }, data: { label: 'Node 1' } },
        ],
        edges: [],
      },
    },
  ]

  const defaultProps = {
    currentVersion: 3,
    versions: mockVersions,
  }

  describe('loading state', () => {
    it('should show loading message when loading', () => {
      render(<WorkflowComparison {...defaultProps} loading={true} />)

      expect(screen.getByText('Loading version history...')).toBeInTheDocument()
    })
  })

  describe('error state', () => {
    it('should show error message when error occurs', () => {
      render(<WorkflowComparison {...defaultProps} error="Failed to load versions" />)

      expect(screen.getByText('Failed to load versions')).toBeInTheDocument()
    })
  })

  describe('empty state', () => {
    it('should show message when less than 2 versions available', () => {
      render(
        <WorkflowComparison
          {...defaultProps}
          versions={[mockVersions[0]]}
        />
      )

      expect(screen.getByText('At least two versions are required to compare.')).toBeInTheDocument()
      expect(screen.getByText('Only one version exists.')).toBeInTheDocument()
    })

    it('should show message when no versions available', () => {
      render(<WorkflowComparison {...defaultProps} versions={[]} />)

      expect(screen.getByText('At least two versions are required to compare.')).toBeInTheDocument()
      expect(screen.getByText('No version history available.')).toBeInTheDocument()
    })
  })

  describe('version selection', () => {
    it('should show version selectors', () => {
      render(<WorkflowComparison {...defaultProps} />)

      expect(screen.getByText('Base Version (older)')).toBeInTheDocument()
      expect(screen.getByText('Compare Version (newer)')).toBeInTheDocument()
    })

    it('should show prompt to select versions initially', () => {
      render(<WorkflowComparison {...defaultProps} />)

      expect(screen.getByText('Select versions to compare')).toBeInTheDocument()
    })

    it('should show quick select buttons', () => {
      render(<WorkflowComparison {...defaultProps} />)

      expect(screen.getByText('Compare current with previous')).toBeInTheDocument()
    })

    it('should show compare first with latest button when more than 2 versions', () => {
      render(<WorkflowComparison {...defaultProps} />)

      expect(screen.getByText('Compare first with latest')).toBeInTheDocument()
    })
  })

  describe('view mode toggle', () => {
    it('should show Visual and JSON toggle buttons', () => {
      render(<WorkflowComparison {...defaultProps} />)

      expect(screen.getByRole('button', { name: 'Visual' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'JSON' })).toBeInTheDocument()
    })

    it('should default to Visual mode', () => {
      render(<WorkflowComparison {...defaultProps} />)

      const visualButton = screen.getByRole('button', { name: 'Visual' })
      expect(visualButton).toHaveClass('bg-primary-600')
    })

    it('should toggle to JSON mode when clicked', async () => {
      const user = userEvent.setup()
      render(<WorkflowComparison {...defaultProps} />)

      await user.click(screen.getByRole('button', { name: 'JSON' }))

      const jsonButton = screen.getByRole('button', { name: 'JSON' })
      expect(jsonButton).toHaveClass('bg-primary-600')
    })
  })

  describe('show unchanged toggle', () => {
    it('should show "Show unchanged" checkbox in visual mode', () => {
      render(<WorkflowComparison {...defaultProps} />)

      expect(screen.getByLabelText(/show unchanged/i)).toBeInTheDocument()
    })

    it('should be unchecked by default', () => {
      render(<WorkflowComparison {...defaultProps} />)

      const checkbox = screen.getByLabelText(/show unchanged/i)
      expect(checkbox).not.toBeChecked()
    })
  })

  describe('close button', () => {
    it('should show close button when onClose provided', () => {
      const onClose = vi.fn()
      render(<WorkflowComparison {...defaultProps} onClose={onClose} />)

      expect(screen.getByRole('button', { name: /close comparison/i })).toBeInTheDocument()
    })

    it('should not show close button when onClose not provided', () => {
      render(<WorkflowComparison {...defaultProps} />)

      expect(screen.queryByRole('button', { name: /close comparison/i })).not.toBeInTheDocument()
    })

    it('should call onClose when close button clicked', async () => {
      const user = userEvent.setup()
      const onClose = vi.fn()
      render(<WorkflowComparison {...defaultProps} onClose={onClose} />)

      await user.click(screen.getByRole('button', { name: /close comparison/i }))

      expect(onClose).toHaveBeenCalledTimes(1)
    })
  })

  describe('header', () => {
    it('should show Version Comparison title', () => {
      render(<WorkflowComparison {...defaultProps} />)

      expect(screen.getByText('Version Comparison')).toBeInTheDocument()
    })
  })
})
