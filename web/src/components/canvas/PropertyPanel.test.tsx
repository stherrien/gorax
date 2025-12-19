import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import PropertyPanel from './PropertyPanel'
import type { Node } from '@xyflow/react'

describe('PropertyPanel', () => {
  const mockOnUpdate = vi.fn()
  const mockOnClose = vi.fn()

  const defaultProps = {
    node: null as Node | null,
    onUpdate: mockOnUpdate,
    onClose: mockOnClose,
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Empty state', () => {
    it('should show empty state when no node selected', () => {
      render(<PropertyPanel {...defaultProps} />)

      expect(screen.getByText(/no node selected/i)).toBeInTheDocument()
    })

    it('should not show close button in empty state', () => {
      render(<PropertyPanel {...defaultProps} />)

      expect(screen.queryByRole('button', { name: /close/i })).not.toBeInTheDocument()
    })
  })

  describe('Node selection', () => {
    it('should display node properties when node selected', () => {
      const node: Node = {
        id: 'node-1',
        type: 'trigger',
        position: { x: 0, y: 0 },
        data: { nodeType: 'webhook', label: 'Webhook Trigger' },
      }

      render(<PropertyPanel {...defaultProps} node={node} />)

      expect(screen.getByText(/webhook trigger/i)).toBeInTheDocument()
    })

    it('should have close button when node selected', () => {
      const node: Node = {
        id: 'node-1',
        type: 'trigger',
        position: { x: 0, y: 0 },
        data: { nodeType: 'webhook', label: 'Webhook Trigger' },
      }

      render(<PropertyPanel {...defaultProps} node={node} />)

      expect(screen.getByRole('button', { name: /close/i })).toBeInTheDocument()
    })

    it('should call onClose when close button clicked', async () => {
      const user = userEvent.setup()
      const onClose = vi.fn()
      const node: Node = {
        id: 'node-1',
        type: 'trigger',
        position: { x: 0, y: 0 },
        data: { nodeType: 'webhook', label: 'Webhook Trigger' },
      }

      render(<PropertyPanel {...defaultProps} node={node} onClose={onClose} />)

      const closeButton = screen.getByRole('button', { name: /close/i })
      await user.click(closeButton)

      expect(onClose).toHaveBeenCalled()
    })
  })

  describe('Webhook trigger properties', () => {
    it('should show webhook trigger fields', () => {
      const node: Node = {
        id: 'node-1',
        type: 'trigger',
        position: { x: 0, y: 0 },
        data: { nodeType: 'webhook', label: 'Webhook Trigger' },
      }

      render(<PropertyPanel {...defaultProps} node={node} />)

      expect(screen.getByLabelText(/name/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/path/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/method/i)).toBeInTheDocument()
    })

    it('should have method dropdown with options', () => {
      const node: Node = {
        id: 'node-1',
        type: 'trigger',
        position: { x: 0, y: 0 },
        data: { nodeType: 'webhook', label: 'Webhook Trigger' },
      }

      render(<PropertyPanel {...defaultProps} node={node} />)

      const methodSelect = screen.getByLabelText(/method/i)
      expect(methodSelect).toBeInTheDocument()
      expect(methodSelect.tagName).toBe('SELECT')
    })
  })

  describe('HTTP action properties', () => {
    it('should show HTTP action fields', () => {
      const node: Node = {
        id: 'node-1',
        type: 'action',
        position: { x: 0, y: 0 },
        data: { nodeType: 'http', label: 'HTTP Request' },
      }

      render(<PropertyPanel {...defaultProps} node={node} />)

      expect(screen.getByLabelText(/name/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/url/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/method/i)).toBeInTheDocument()
    })

    it('should have headers field', () => {
      const node: Node = {
        id: 'node-1',
        type: 'action',
        position: { x: 0, y: 0 },
        data: { nodeType: 'http', label: 'HTTP Request' },
      }

      render(<PropertyPanel {...defaultProps} node={node} />)

      expect(screen.getByLabelText(/headers/i)).toBeInTheDocument()
    })

    it('should have body field', () => {
      const node: Node = {
        id: 'node-1',
        type: 'action',
        position: { x: 0, y: 0 },
        data: { nodeType: 'http', label: 'HTTP Request' },
      }

      render(<PropertyPanel {...defaultProps} node={node} />)

      expect(screen.getByLabelText(/body/i)).toBeInTheDocument()
    })
  })

  describe('Transform action properties', () => {
    it('should show transform action fields', () => {
      const node: Node = {
        id: 'node-1',
        type: 'action',
        position: { x: 0, y: 0 },
        data: { nodeType: 'transform', label: 'Transform Data' },
      }

      render(<PropertyPanel {...defaultProps} node={node} />)

      expect(screen.getByLabelText(/name/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/mapping/i)).toBeInTheDocument()
    })
  })

  describe('Conditional control properties', () => {
    it('should show conditional control fields', () => {
      const node: Node = {
        id: 'node-1',
        type: 'control',
        position: { x: 0, y: 0 },
        data: { nodeType: 'conditional', label: 'Conditional' },
      }

      render(<PropertyPanel {...defaultProps} node={node} />)

      expect(screen.getByLabelText(/name/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/condition/i)).toBeInTheDocument()
    })
  })

  describe('Save functionality', () => {
    it('should have save button', () => {
      const node: Node = {
        id: 'node-1',
        type: 'trigger',
        position: { x: 0, y: 0 },
        data: { nodeType: 'webhook', label: 'Webhook Trigger' },
      }

      render(<PropertyPanel {...defaultProps} node={node} />)

      expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument()
    })

    it('should call onUpdate with updated data when save clicked', async () => {
      const user = userEvent.setup()
      const onUpdate = vi.fn()
      const node: Node = {
        id: 'node-1',
        type: 'trigger',
        position: { x: 0, y: 0 },
        data: { nodeType: 'webhook', label: 'Webhook Trigger', path: '/webhook' },
      }

      render(<PropertyPanel {...defaultProps} node={node} onUpdate={onUpdate} />)

      const nameInput = screen.getByLabelText(/name/i)
      await user.clear(nameInput)
      await user.type(nameInput, 'Updated Webhook')

      const saveButton = screen.getByRole('button', { name: /save/i })
      await user.click(saveButton)

      await waitFor(() => {
        expect(onUpdate).toHaveBeenCalledWith('node-1', {
          label: 'Updated Webhook',
          nodeType: 'webhook',
          path: '/webhook',
        })
      })
    })

    it('should show success message after save', async () => {
      const user = userEvent.setup()
      const node: Node = {
        id: 'node-1',
        type: 'trigger',
        position: { x: 0, y: 0 },
        data: { nodeType: 'webhook', label: 'Webhook Trigger' },
      }

      render(<PropertyPanel {...defaultProps} node={node} />)

      const saveButton = screen.getByRole('button', { name: /save/i })
      await user.click(saveButton)

      await waitFor(() => {
        expect(screen.getByText(/saved successfully/i)).toBeInTheDocument()
      })
    })
  })

  describe('Validation', () => {
    it('should validate required fields', async () => {
      const user = userEvent.setup()
      const onUpdate = vi.fn()
      const node: Node = {
        id: 'node-1',
        type: 'trigger',
        position: { x: 0, y: 0 },
        data: { nodeType: 'webhook', label: 'Webhook Trigger', path: '/webhook' },
      }

      render(<PropertyPanel {...defaultProps} node={node} onUpdate={onUpdate} />)

      const nameInput = screen.getByLabelText(/name/i)
      await user.clear(nameInput)

      const saveButton = screen.getByRole('button', { name: /save/i })
      await user.click(saveButton)

      await waitFor(() => {
        expect(screen.getByText(/name is required/i)).toBeInTheDocument()
      })

      expect(onUpdate).not.toHaveBeenCalled()
    })

    it('should validate URL format for HTTP action', async () => {
      const user = userEvent.setup()
      const onUpdate = vi.fn()
      const node: Node = {
        id: 'node-1',
        type: 'action',
        position: { x: 0, y: 0 },
        data: { nodeType: 'http', label: 'HTTP Request' },
      }

      render(<PropertyPanel {...defaultProps} node={node} onUpdate={onUpdate} />)

      const urlInput = screen.getByLabelText(/url/i)
      await user.type(urlInput, 'not-a-valid-url')

      const saveButton = screen.getByRole('button', { name: /save/i })
      await user.click(saveButton)

      await waitFor(() => {
        expect(screen.getByText(/invalid url/i)).toBeInTheDocument()
      })

      expect(onUpdate).not.toHaveBeenCalled()
    })
  })

  describe('Field updates', () => {
    it('should update field values as user types', async () => {
      const user = userEvent.setup()
      const node: Node = {
        id: 'node-1',
        type: 'trigger',
        position: { x: 0, y: 0 },
        data: { nodeType: 'webhook', label: 'Webhook Trigger', path: '/webhook' },
      }

      render(<PropertyPanel {...defaultProps} node={node} />)

      const nameInput = screen.getByLabelText(/name/i) as HTMLInputElement
      await user.clear(nameInput)
      await user.type(nameInput, 'New Name')

      expect(nameInput.value).toBe('New Name')
    })

    it('should pre-populate fields with existing data', () => {
      const node: Node = {
        id: 'node-1',
        type: 'action',
        position: { x: 0, y: 0 },
        data: {
          nodeType: 'http',
          label: 'My HTTP Request',
          url: 'https://api.example.com',
          method: 'POST',
        },
      }

      render(<PropertyPanel {...defaultProps} node={node} />)

      const nameInput = screen.getByLabelText(/name/i) as HTMLInputElement
      const urlInput = screen.getByLabelText(/url/i) as HTMLInputElement
      const methodSelect = screen.getByLabelText(/method/i) as HTMLSelectElement

      expect(nameInput.value).toBe('My HTTP Request')
      expect(urlInput.value).toBe('https://api.example.com')
      expect(methodSelect.value).toBe('POST')
    })
  })

  describe('Reset functionality', () => {
    it('should have reset button', () => {
      const node: Node = {
        id: 'node-1',
        type: 'trigger',
        position: { x: 0, y: 0 },
        data: { nodeType: 'webhook', label: 'Webhook Trigger' },
      }

      render(<PropertyPanel {...defaultProps} node={node} />)

      expect(screen.getByRole('button', { name: /reset/i })).toBeInTheDocument()
    })

    it('should reset fields to original values when reset clicked', async () => {
      const user = userEvent.setup()
      const node: Node = {
        id: 'node-1',
        type: 'trigger',
        position: { x: 0, y: 0 },
        data: { nodeType: 'webhook', label: 'Webhook Trigger', path: '/webhook' },
      }

      render(<PropertyPanel {...defaultProps} node={node} />)

      const nameInput = screen.getByLabelText(/name/i) as HTMLInputElement
      await user.clear(nameInput)
      await user.type(nameInput, 'Changed Name')

      expect(nameInput.value).toBe('Changed Name')

      const resetButton = screen.getByRole('button', { name: /reset/i })
      await user.click(resetButton)

      await waitFor(() => {
        expect(nameInput.value).toBe('Webhook Trigger')
      })
    })
  })
})
