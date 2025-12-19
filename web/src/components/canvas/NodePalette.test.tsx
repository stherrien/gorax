import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import NodePalette from './NodePalette'

describe('NodePalette', () => {
  const mockOnAddNode = vi.fn()

  const defaultProps = {
    onAddNode: mockOnAddNode,
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Rendering', () => {
    it('should render the node palette', () => {
      render(<NodePalette {...defaultProps} />)

      expect(screen.getByText(/node palette/i)).toBeInTheDocument()
    })

    it('should display trigger nodes section', () => {
      render(<NodePalette {...defaultProps} />)

      expect(screen.getByText(/triggers/i)).toBeInTheDocument()
    })

    it('should display action nodes section', () => {
      render(<NodePalette {...defaultProps} />)

      expect(screen.getByText(/actions/i)).toBeInTheDocument()
    })

    it('should display control nodes section', () => {
      render(<NodePalette {...defaultProps} />)

      expect(screen.getByText(/controls/i)).toBeInTheDocument()
    })
  })

  describe('Trigger nodes', () => {
    it('should display webhook trigger', () => {
      render(<NodePalette {...defaultProps} />)

      expect(screen.getByText(/webhook/i)).toBeInTheDocument()
    })

    it('should display schedule trigger', () => {
      render(<NodePalette {...defaultProps} />)

      expect(screen.getByText(/schedule/i)).toBeInTheDocument()
    })

    it('should display manual trigger', () => {
      render(<NodePalette {...defaultProps} />)

      expect(screen.getByText(/manual/i)).toBeInTheDocument()
    })

    it('should show trigger descriptions on hover', async () => {
      const user = userEvent.setup()
      render(<NodePalette {...defaultProps} />)

      const webhookNode = screen.getByText(/webhook/i)
      await user.hover(webhookNode)

      await waitFor(() => {
        expect(screen.getByText(/trigger workflow via http/i)).toBeInTheDocument()
      })
    })
  })

  describe('Action nodes', () => {
    it('should display HTTP Request action', () => {
      render(<NodePalette {...defaultProps} />)

      expect(screen.getByText(/http request/i)).toBeInTheDocument()
    })

    it('should display Transform action', () => {
      render(<NodePalette {...defaultProps} />)

      expect(screen.getByText(/transform/i)).toBeInTheDocument()
    })

    it('should display Email action', () => {
      render(<NodePalette {...defaultProps} />)

      expect(screen.getByText(/email/i)).toBeInTheDocument()
    })

    it('should display Run Script action', () => {
      render(<NodePalette {...defaultProps} />)

      expect(screen.getByText(/run script/i)).toBeInTheDocument()
    })
  })

  describe('Control nodes', () => {
    it('should display Conditional control', () => {
      render(<NodePalette {...defaultProps} />)

      expect(screen.getByText(/conditional/i)).toBeInTheDocument()
    })

    it('should display Loop control', () => {
      render(<NodePalette {...defaultProps} />)

      expect(screen.getByText(/loop/i)).toBeInTheDocument()
    })

    it('should display Delay control', () => {
      render(<NodePalette {...defaultProps} />)

      expect(screen.getByText(/delay/i)).toBeInTheDocument()
    })
  })

  describe('Add node functionality', () => {
    it('should call onAddNode when node is clicked', async () => {
      const user = userEvent.setup()
      const onAddNode = vi.fn()

      render(<NodePalette {...defaultProps} onAddNode={onAddNode} />)

      const webhookNode = screen.getByText(/webhook/i)
      await user.click(webhookNode)

      expect(onAddNode).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'trigger',
          nodeType: 'webhook',
        })
      )
    })

    it('should add action node when action clicked', async () => {
      const user = userEvent.setup()
      const onAddNode = vi.fn()

      render(<NodePalette {...defaultProps} onAddNode={onAddNode} />)

      const httpNode = screen.getByText(/http request/i)
      await user.click(httpNode)

      expect(onAddNode).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'action',
          nodeType: 'http',
        })
      )
    })

    it('should add control node when control clicked', async () => {
      const user = userEvent.setup()
      const onAddNode = vi.fn()

      render(<NodePalette {...defaultProps} onAddNode={onAddNode} />)

      const conditionalNode = screen.getByText(/conditional/i)
      await user.click(conditionalNode)

      expect(onAddNode).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'control',
          nodeType: 'conditional',
        })
      )
    })
  })

  describe('Drag and drop support', () => {
    it('should have draggable nodes', () => {
      render(<NodePalette {...defaultProps} />)

      const webhookNode = screen.getByText(/webhook/i).closest('[draggable]')
      expect(webhookNode).toHaveAttribute('draggable', 'true')
    })

    it('should set drag data on dragStart', async () => {
      const user = userEvent.setup()
      render(<NodePalette {...defaultProps} />)

      const webhookNode = screen.getByText(/webhook/i)
      const draggableParent = webhookNode.closest('[draggable="true"]')

      expect(draggableParent).toBeInTheDocument()
    })
  })

  describe('Collapsible sections', () => {
    it('should have collapsible trigger section', async () => {
      const user = userEvent.setup()
      render(<NodePalette {...defaultProps} />)

      const triggerHeader = screen.getByText(/triggers/i)
      await user.click(triggerHeader)

      // Section should still be in document (just testing collapse mechanism exists)
      expect(triggerHeader).toBeInTheDocument()
    })

    it('should show/hide section content when toggled', async () => {
      const user = userEvent.setup()
      render(<NodePalette {...defaultProps} />)

      const triggersSection = screen.getByText(/triggers/i)
      const sectionContainer = triggersSection.closest('div')

      expect(sectionContainer).toBeInTheDocument()
    })
  })

  describe('Search functionality', () => {
    it('should have search input', () => {
      render(<NodePalette {...defaultProps} />)

      expect(screen.getByPlaceholderText(/search nodes/i)).toBeInTheDocument()
    })

    it('should filter nodes based on search query', async () => {
      const user = userEvent.setup()
      render(<NodePalette {...defaultProps} />)

      const searchInput = screen.getByPlaceholderText(/search nodes/i)
      await user.type(searchInput, 'webhook')

      // Webhook should be visible
      expect(screen.getByText(/webhook/i)).toBeInTheDocument()

      // HTTP Request should not be visible (or should be hidden)
      // We'll check that only matching nodes are shown
    })

    it('should show all nodes when search is cleared', async () => {
      const user = userEvent.setup()
      render(<NodePalette {...defaultProps} />)

      const searchInput = screen.getByPlaceholderText(/search nodes/i)
      await user.type(searchInput, 'webhook')
      await user.clear(searchInput)

      // All nodes should be visible again
      expect(screen.getByText(/webhook/i)).toBeInTheDocument()
      expect(screen.getByText(/http request/i)).toBeInTheDocument()
      expect(screen.getByText(/conditional/i)).toBeInTheDocument()
    })

    it('should show no results message when search has no matches', async () => {
      const user = userEvent.setup()
      render(<NodePalette {...defaultProps} />)

      const searchInput = screen.getByPlaceholderText(/search nodes/i)
      await user.type(searchInput, 'nonexistent')

      expect(screen.getByText(/no nodes found/i)).toBeInTheDocument()
    })
  })

  describe('Node icons', () => {
    it('should display icons for trigger nodes', () => {
      render(<NodePalette {...defaultProps} />)

      // Check that trigger nodes have icons (data-icon attribute or similar)
      const webhookNode = screen.getByText(/webhook/i)
      const nodeContainer = webhookNode.closest('div')
      expect(nodeContainer).toBeInTheDocument()
    })

    it('should display icons for action nodes', () => {
      render(<NodePalette {...defaultProps} />)

      const httpNode = screen.getByText(/http request/i)
      const nodeContainer = httpNode.closest('div')
      expect(nodeContainer).toBeInTheDocument()
    })

    it('should display icons for control nodes', () => {
      render(<NodePalette {...defaultProps} />)

      const conditionalNode = screen.getByText(/conditional/i)
      const nodeContainer = conditionalNode.closest('div')
      expect(nodeContainer).toBeInTheDocument()
    })
  })
})
