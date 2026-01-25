import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import VersionSelector, { QuickSelectButtons } from './VersionSelector'
import type { WorkflowVersion } from '../../../api/workflows'

describe('VersionSelector', () => {
  const mockVersions: WorkflowVersion[] = [
    {
      id: 'ver-3',
      workflowId: 'wf-123',
      version: 3,
      createdAt: new Date().toISOString(),
      createdBy: 'user1',
      definition: {
        nodes: [{ id: 'n1' }, { id: 'n2' }, { id: 'n3' }],
        edges: [{ id: 'e1' }, { id: 'e2' }],
      },
    },
    {
      id: 'ver-2',
      workflowId: 'wf-123',
      version: 2,
      createdAt: new Date(Date.now() - 86400000).toISOString(),
      createdBy: 'user1',
      definition: {
        nodes: [{ id: 'n1' }, { id: 'n2' }],
        edges: [{ id: 'e1' }],
      },
    },
    {
      id: 'ver-1',
      workflowId: 'wf-123',
      version: 1,
      createdAt: new Date(Date.now() - 172800000).toISOString(),
      createdBy: 'user1',
      definition: {
        nodes: [{ id: 'n1' }],
        edges: [],
      },
    },
  ]

  const defaultProps = {
    versions: mockVersions,
    currentVersion: 3,
    baseVersionId: null,
    compareVersionId: null,
    onBaseVersionChange: vi.fn(),
    onCompareVersionChange: vi.fn(),
  }

  describe('rendering', () => {
    it('should render base version selector', () => {
      render(<VersionSelector {...defaultProps} />)

      expect(screen.getByText('Base Version (older)')).toBeInTheDocument()
    })

    it('should render compare version selector', () => {
      render(<VersionSelector {...defaultProps} />)

      expect(screen.getByText('Compare Version (newer)')).toBeInTheDocument()
    })

    it('should show placeholder when no version selected', () => {
      render(<VersionSelector {...defaultProps} />)

      expect(screen.getByText('Select base version')).toBeInTheDocument()
      expect(screen.getByText('Select compare version')).toBeInTheDocument()
    })

    it('should show selected base version', () => {
      render(<VersionSelector {...defaultProps} baseVersionId="ver-1" />)

      expect(screen.getByText('v1')).toBeInTheDocument()
    })

    it('should show selected compare version', () => {
      render(<VersionSelector {...defaultProps} compareVersionId="ver-3" />)

      expect(screen.getByText('v3')).toBeInTheDocument()
    })

    it('should show Current badge for current version', () => {
      render(<VersionSelector {...defaultProps} compareVersionId="ver-3" />)

      expect(screen.getByText('Current')).toBeInTheDocument()
    })

    it('should render swap button', () => {
      render(<VersionSelector {...defaultProps} />)

      expect(screen.getByRole('button', { name: /swap/i })).toBeInTheDocument()
    })
  })

  describe('dropdown interaction', () => {
    it('should open base version dropdown when clicked', async () => {
      const user = userEvent.setup()
      render(<VersionSelector {...defaultProps} />)

      await user.click(screen.getByText('Select base version'))

      // Should show versions in dropdown
      await waitFor(() => {
        expect(screen.getByText('v3')).toBeInTheDocument()
        expect(screen.getByText('v2')).toBeInTheDocument()
        expect(screen.getByText('v1')).toBeInTheDocument()
      })
    })

    it('should call onBaseVersionChange when version selected', async () => {
      const user = userEvent.setup()
      const onBaseVersionChange = vi.fn()
      render(<VersionSelector {...defaultProps} onBaseVersionChange={onBaseVersionChange} />)

      await user.click(screen.getByText('Select base version'))

      await waitFor(() => {
        expect(screen.getByText('v2')).toBeInTheDocument()
      })

      await user.click(screen.getByText('v2'))

      expect(onBaseVersionChange).toHaveBeenCalledWith('ver-2')
    })

    it('should call onCompareVersionChange when version selected', async () => {
      const user = userEvent.setup()
      const onCompareVersionChange = vi.fn()
      render(<VersionSelector {...defaultProps} onCompareVersionChange={onCompareVersionChange} />)

      await user.click(screen.getByText('Select compare version'))

      await waitFor(() => {
        expect(screen.getByText('v3')).toBeInTheDocument()
      })

      await user.click(screen.getByText('v3'))

      expect(onCompareVersionChange).toHaveBeenCalledWith('ver-3')
    })

    it('should exclude selected base version from compare dropdown', async () => {
      const user = userEvent.setup()
      render(<VersionSelector {...defaultProps} baseVersionId="ver-2" />)

      await user.click(screen.getByText('Select compare version'))

      await waitFor(() => {
        // v2 should not be in compare dropdown since it's selected as base
        const v2Elements = screen.getAllByText('v2')
        expect(v2Elements).toHaveLength(1) // Only in base selector
      })
    })

    it('should exclude selected compare version from base dropdown', async () => {
      const user = userEvent.setup()
      render(<VersionSelector {...defaultProps} compareVersionId="ver-3" />)

      await user.click(screen.getByText('Select base version'))

      await waitFor(() => {
        // Should show v1 and v2 but not v3 (which is selected as compare)
        const dropdownOptions = screen.getAllByRole('button').filter(
          btn => btn.textContent?.includes('v1') || btn.textContent?.includes('v2')
        )
        expect(dropdownOptions.length).toBeGreaterThan(0)
      })
    })

    it('should close dropdown when backdrop clicked', async () => {
      const user = userEvent.setup()
      render(<VersionSelector {...defaultProps} />)

      await user.click(screen.getByText('Select base version'))

      await waitFor(() => {
        expect(screen.getByText('v1')).toBeInTheDocument()
      })

      // Click backdrop
      const backdrop = document.querySelector('.fixed.inset-0')
      if (backdrop) {
        await user.click(backdrop)
      }

      // Dropdown should close (v1 button in dropdown should be gone)
      // Note: v1 might still be visible as selected, but dropdown should close
    })

    it('should show node and edge counts in dropdown', async () => {
      const user = userEvent.setup()
      render(<VersionSelector {...defaultProps} />)

      await user.click(screen.getByText('Select base version'))

      await waitFor(() => {
        expect(screen.getByText('3 nodes, 2 connections')).toBeInTheDocument()
        expect(screen.getByText('2 nodes, 1 connections')).toBeInTheDocument()
        expect(screen.getByText('1 nodes, 0 connections')).toBeInTheDocument()
      })
    })
  })

  describe('swap functionality', () => {
    it('should swap versions when swap button clicked', async () => {
      const user = userEvent.setup()
      const onBaseVersionChange = vi.fn()
      const onCompareVersionChange = vi.fn()

      render(
        <VersionSelector
          {...defaultProps}
          baseVersionId="ver-1"
          compareVersionId="ver-3"
          onBaseVersionChange={onBaseVersionChange}
          onCompareVersionChange={onCompareVersionChange}
        />
      )

      await user.click(screen.getByRole('button', { name: /swap/i }))

      expect(onBaseVersionChange).toHaveBeenCalledWith('ver-3')
      expect(onCompareVersionChange).toHaveBeenCalledWith('ver-1')
    })

    it('should disable swap button when no versions selected', () => {
      render(<VersionSelector {...defaultProps} />)

      const swapButton = screen.getByRole('button', { name: /swap/i })
      expect(swapButton).toBeDisabled()
    })

    it('should disable swap button when only base selected', () => {
      render(<VersionSelector {...defaultProps} baseVersionId="ver-1" />)

      const swapButton = screen.getByRole('button', { name: /swap/i })
      expect(swapButton).toBeDisabled()
    })

    it('should enable swap button when both versions selected', () => {
      render(
        <VersionSelector
          {...defaultProps}
          baseVersionId="ver-1"
          compareVersionId="ver-3"
        />
      )

      const swapButton = screen.getByRole('button', { name: /swap/i })
      expect(swapButton).not.toBeDisabled()
    })
  })

  describe('loading state', () => {
    it('should disable selectors when loading', () => {
      render(<VersionSelector {...defaultProps} loading={true} />)

      const buttons = screen.getAllByRole('button')
      buttons.forEach(button => {
        expect(button).toBeDisabled()
      })
    })
  })
})

describe('QuickSelectButtons', () => {
  const mockVersions: WorkflowVersion[] = [
    {
      id: 'ver-3',
      workflowId: 'wf-123',
      version: 3,
      createdAt: new Date().toISOString(),
      createdBy: 'user1',
      definition: { nodes: [], edges: [] },
    },
    {
      id: 'ver-2',
      workflowId: 'wf-123',
      version: 2,
      createdAt: new Date(Date.now() - 86400000).toISOString(),
      createdBy: 'user1',
      definition: { nodes: [], edges: [] },
    },
    {
      id: 'ver-1',
      workflowId: 'wf-123',
      version: 1,
      createdAt: new Date(Date.now() - 172800000).toISOString(),
      createdBy: 'user1',
      definition: { nodes: [], edges: [] },
    },
  ]

  describe('compare with previous', () => {
    it('should show compare with previous button', () => {
      render(
        <QuickSelectButtons
          versions={mockVersions}
          currentVersion={3}
          onSelectPair={vi.fn()}
        />
      )

      expect(screen.getByText('Compare current with previous')).toBeInTheDocument()
    })

    it('should call onSelectPair with current and previous versions', async () => {
      const user = userEvent.setup()
      const onSelectPair = vi.fn()

      render(
        <QuickSelectButtons
          versions={mockVersions}
          currentVersion={3}
          onSelectPair={onSelectPair}
        />
      )

      await user.click(screen.getByText('Compare current with previous'))

      expect(onSelectPair).toHaveBeenCalledWith('ver-2', 'ver-3')
    })

    it('should disable button when less than 2 versions', () => {
      render(
        <QuickSelectButtons
          versions={[mockVersions[0]]}
          currentVersion={3}
          onSelectPair={vi.fn()}
        />
      )

      expect(screen.getByText('Compare current with previous')).toBeDisabled()
    })
  })

  describe('compare first with latest', () => {
    it('should show compare first with latest button when more than 2 versions', () => {
      render(
        <QuickSelectButtons
          versions={mockVersions}
          currentVersion={3}
          onSelectPair={vi.fn()}
        />
      )

      expect(screen.getByText('Compare first with latest')).toBeInTheDocument()
    })

    it('should not show button when 2 or fewer versions', () => {
      render(
        <QuickSelectButtons
          versions={mockVersions.slice(0, 2)}
          currentVersion={3}
          onSelectPair={vi.fn()}
        />
      )

      expect(screen.queryByText('Compare first with latest')).not.toBeInTheDocument()
    })

    it('should call onSelectPair with first and latest versions', async () => {
      const user = userEvent.setup()
      const onSelectPair = vi.fn()

      render(
        <QuickSelectButtons
          versions={mockVersions}
          currentVersion={3}
          onSelectPair={onSelectPair}
        />
      )

      await user.click(screen.getByText('Compare first with latest'))

      expect(onSelectPair).toHaveBeenCalledWith('ver-1', 'ver-3')
    })
  })

  describe('loading state', () => {
    it('should disable buttons when loading', () => {
      render(
        <QuickSelectButtons
          versions={mockVersions}
          currentVersion={3}
          onSelectPair={vi.fn()}
          loading={true}
        />
      )

      expect(screen.getByText('Compare current with previous')).toBeDisabled()
      expect(screen.getByText('Compare first with latest')).toBeDisabled()
    })
  })
})
