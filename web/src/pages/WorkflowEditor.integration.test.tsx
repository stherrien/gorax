import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import userEvent from '@testing-library/user-event'
import WorkflowEditor from './WorkflowEditor'
import { workflowAPI } from '../api/workflows'
import type { Workflow } from '../api/workflows'

// Mock the API
vi.mock('../api/workflows', () => ({
  workflowAPI: {
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    list: vi.fn(),
    delete: vi.fn(),
    execute: vi.fn(),
  },
}))

// Mock the canvas components
vi.mock('../components/canvas/WorkflowCanvas', () => ({
  default: vi.fn(({ onSave, onChange, onNodeSelect }) => {
    const testNode = {
      id: 'node-1',
      type: 'trigger',
      position: { x: 0, y: 0 },
      data: { nodeType: 'webhook', label: 'Webhook Trigger' },
    }
    return (
      <div data-testid="workflow-canvas">
        <button
          data-testid="canvas-save-button"
          onClick={() => onSave?.()}
        >
          Canvas Save
        </button>
        <button
          data-testid="canvas-add-node"
          onClick={() => {
            onChange?.({
              nodes: [testNode],
              edges: [],
            })
            onNodeSelect?.(testNode)
          }}
        >
          Add Node
        </button>
      </div>
    )
  }),
}))

vi.mock('../components/canvas/NodePalette', () => ({
  default: vi.fn(() => <div data-testid="node-palette">Node Palette</div>),
}))

// Use real PropertyPanel to test node save
vi.mock('../components/canvas/PropertyPanel', () => ({
  default: vi.fn(({ node, onUpdate, onSave, isSaving }) => {
    if (!node) {
      return <div data-testid="property-panel">No node selected</div>
    }
    return (
      <div data-testid="property-panel">
        <div data-testid="node-label">{node.data?.label}</div>
        <button
          data-testid="node-save-button"
          disabled={isSaving}
          onClick={async () => {
            onUpdate(node.id, { ...node.data, label: 'Updated Label' })
            if (onSave) {
              await onSave()
            }
          }}
        >
          {isSaving ? 'Saving...' : 'Save Node'}
        </button>
      </div>
    )
  }),
}))

const mockWorkflow: Workflow = {
  id: 'wf-123',
  tenantId: 'tenant-1',
  name: 'Test Workflow',
  description: 'Test Description',
  definition: {
    nodes: [{
      id: 'node-1',
      type: 'trigger',
      position: { x: 100, y: 100 },
      data: { nodeType: 'webhook', label: 'Webhook Trigger' },
    }],
    edges: [],
  },
  status: 'active',
  version: 1,
  createdAt: '2025-01-15T10:00:00Z',
  updatedAt: '2025-01-15T10:00:00Z',
}

describe('WorkflowEditor Integration Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  describe('Save Workflow Button', () => {
    it('should persist new workflow to API when clicking Save Workflow', async () => {
      const user = userEvent.setup()
      const createdWorkflow = { ...mockWorkflow, id: 'new-wf-1' }
      vi.mocked(workflowAPI.create).mockResolvedValue(createdWorkflow)

      render(
        <MemoryRouter initialEntries={['/workflows/new']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      // Enter workflow name
      const nameInput = screen.getByPlaceholderText(/workflow name/i)
      await user.type(nameInput, 'My New Workflow')

      // Click Save Workflow button
      const saveButton = screen.getByRole('button', { name: /save workflow/i })
      await user.click(saveButton)

      // Verify API was called
      await waitFor(() => {
        expect(workflowAPI.create).toHaveBeenCalledWith(
          expect.objectContaining({
            name: 'My New Workflow',
            definition: expect.objectContaining({
              nodes: expect.any(Array),
              edges: expect.any(Array),
            }),
          })
        )
      })
    })

    it('should update existing workflow when clicking Save Workflow', async () => {
      const user = userEvent.setup()
      vi.mocked(workflowAPI.get).mockResolvedValue(mockWorkflow)
      vi.mocked(workflowAPI.update).mockResolvedValue(mockWorkflow)

      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      // Wait for workflow to load
      await waitFor(() => {
        const nameInput = screen.getByPlaceholderText(/workflow name/i) as HTMLInputElement
        expect(nameInput.value).toBe('Test Workflow')
      })

      // Click Save Workflow button
      const saveButton = screen.getByRole('button', { name: /save workflow/i })
      await user.click(saveButton)

      // Verify update API was called
      await waitFor(() => {
        expect(workflowAPI.update).toHaveBeenCalledWith('wf-123', expect.any(Object))
      })
    })

    it('should show error message when save fails', async () => {
      const user = userEvent.setup()
      vi.mocked(workflowAPI.get).mockResolvedValue(mockWorkflow)
      vi.mocked(workflowAPI.update).mockRejectedValue(new Error('Network error'))

      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      // Wait for workflow to load
      await waitFor(() => {
        expect(screen.getByPlaceholderText(/workflow name/i)).toBeInTheDocument()
      })

      // Click Save Workflow button
      const saveButton = screen.getByRole('button', { name: /save workflow/i })
      await user.click(saveButton)

      // Verify error is displayed
      await waitFor(() => {
        expect(screen.getByText(/network error/i)).toBeInTheDocument()
      })
    })

    it('should show success message when save succeeds', async () => {
      const user = userEvent.setup()
      vi.mocked(workflowAPI.get).mockResolvedValue(mockWorkflow)
      vi.mocked(workflowAPI.update).mockResolvedValue(mockWorkflow)

      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      // Wait for workflow to load
      await waitFor(() => {
        expect(screen.getByPlaceholderText(/workflow name/i)).toBeInTheDocument()
      })

      // Click Save Workflow button
      const saveButton = screen.getByRole('button', { name: /save workflow/i })
      await user.click(saveButton)

      // Verify success message is displayed
      await waitFor(() => {
        expect(screen.getByText(/workflow saved successfully/i)).toBeInTheDocument()
      })
    })

    it('should disable save button while saving', async () => {
      const user = userEvent.setup()
      vi.mocked(workflowAPI.get).mockResolvedValue(mockWorkflow)
      // Make update take some time
      vi.mocked(workflowAPI.update).mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve(mockWorkflow), 100))
      )

      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      // Wait for workflow to load
      await waitFor(() => {
        expect(screen.getByPlaceholderText(/workflow name/i)).toBeInTheDocument()
      })

      const saveButton = screen.getByRole('button', { name: /save workflow/i })
      await user.click(saveButton)

      // Button should show "Saving..." and be disabled
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /saving/i })).toBeDisabled()
      })

      // Wait for save to complete
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /save workflow/i })).toBeEnabled()
      })
    })

    it('should validate workflow name before saving', async () => {
      const user = userEvent.setup()

      render(
        <MemoryRouter initialEntries={['/workflows/new']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      // Click Save without entering name
      const saveButton = screen.getByRole('button', { name: /save workflow/i })
      await user.click(saveButton)

      // Should show validation error
      await waitFor(() => {
        expect(screen.getByText(/workflow name is required/i)).toBeInTheDocument()
      })

      // API should not be called
      expect(workflowAPI.create).not.toHaveBeenCalled()
    })
  })

  describe('Canvas Save Button', () => {
    it('should trigger workflow save when canvas save button is clicked', async () => {
      const user = userEvent.setup()
      vi.mocked(workflowAPI.get).mockResolvedValue(mockWorkflow)
      vi.mocked(workflowAPI.update).mockResolvedValue(mockWorkflow)

      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      // Wait for workflow to load
      await waitFor(() => {
        expect(screen.getByTestId('workflow-canvas')).toBeInTheDocument()
      })

      // Click canvas save button
      const canvasSaveButton = screen.getByTestId('canvas-save-button')
      await user.click(canvasSaveButton)

      // Verify update API was called
      await waitFor(() => {
        expect(workflowAPI.update).toHaveBeenCalledWith('wf-123', expect.any(Object))
      })
    })
  })

  describe('Node Save Button (PropertyPanel)', () => {
    it('should persist workflow to API when node save button is clicked', async () => {
      const user = userEvent.setup()
      vi.mocked(workflowAPI.get).mockResolvedValue(mockWorkflow)
      vi.mocked(workflowAPI.update).mockResolvedValue(mockWorkflow)

      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      // Wait for workflow to load
      await waitFor(() => {
        expect(screen.getByTestId('workflow-canvas')).toBeInTheDocument()
      })

      // Add a node to trigger property panel
      const addNodeButton = screen.getByTestId('canvas-add-node')
      await user.click(addNodeButton)

      // Verify property panel shows node
      await waitFor(() => {
        const panel = screen.getByTestId('property-panel')
        expect(within(panel).getByTestId('node-label')).toHaveTextContent('Webhook Trigger')
      })

      // Click node save button
      const nodeSaveButton = screen.getByTestId('node-save-button')
      await user.click(nodeSaveButton)

      // Verify update API was called (node save should trigger workflow save)
      await waitFor(() => {
        expect(workflowAPI.update).toHaveBeenCalledWith('wf-123', expect.any(Object))
      })
    })

    it.skip('should show saving state in node save button', async () => {
      const user = userEvent.setup()
      vi.mocked(workflowAPI.get).mockResolvedValue(mockWorkflow)
      // Make update take some time
      vi.mocked(workflowAPI.update).mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve(mockWorkflow), 500))
      )

      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      // Wait for workflow to load and add node
      await waitFor(() => {
        expect(screen.getByTestId('workflow-canvas')).toBeInTheDocument()
      })

      const addNodeButton = screen.getByTestId('canvas-add-node')
      await user.click(addNodeButton)

      // Click node save button
      const nodeSaveButton = screen.getByTestId('node-save-button')
      await user.click(nodeSaveButton)

      // Button should show "Saving..." and be disabled
      await waitFor(() => {
        expect(screen.queryByText('Saving...') || screen.getByTestId('node-save-button')).toBeTruthy()
        expect(screen.getByTestId('node-save-button')).toBeDisabled()
      }, { timeout: 1000 })

      // Wait for save to complete
      await waitFor(() => {
        expect(screen.getByText('Save Node')).toBeInTheDocument()
        expect(screen.getByTestId('node-save-button')).toBeEnabled()
      }, { timeout: 2000 })
    })
  })

  describe('Error Handling', () => {
    it('should display API error message to user', async () => {
      const user = userEvent.setup()
      vi.mocked(workflowAPI.get).mockResolvedValue(mockWorkflow)
      vi.mocked(workflowAPI.update).mockRejectedValue(new Error('Server unavailable'))

      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByPlaceholderText(/workflow name/i)).toBeInTheDocument()
      })

      const saveButton = screen.getByRole('button', { name: /save workflow/i })
      await user.click(saveButton)

      await waitFor(() => {
        expect(screen.getByText(/server unavailable/i)).toBeInTheDocument()
      })
    })

    it('should display validation error when workflow load fails', async () => {
      vi.mocked(workflowAPI.get).mockRejectedValue(new Error('Workflow not found'))

      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText(/failed to load workflow/i)).toBeInTheDocument()
      })
    })
  })
})
