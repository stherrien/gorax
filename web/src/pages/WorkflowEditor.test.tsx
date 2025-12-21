import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import userEvent from '@testing-library/user-event'
import WorkflowEditor from './WorkflowEditor'
import type { Workflow } from '../api/workflows'

// Mock the components
vi.mock('../components/canvas/WorkflowCanvas', () => ({
  default: vi.fn(({ initialNodes, initialEdges, onSave, onChange }) => (
    <div data-testid="workflow-canvas">
      <button onClick={() => onSave?.({ nodes: [], edges: [] })}>Save Canvas</button>
    </div>
  )),
}))

vi.mock('../components/canvas/NodePalette', () => ({
  default: vi.fn(({ onAddNode }) => (
    <div data-testid="node-palette">Node Palette</div>
  )),
}))

vi.mock('../components/canvas/PropertyPanel', () => ({
  default: vi.fn(({ node }) => (
    <div data-testid="property-panel">Property Panel</div>
  )),
}))

// Mock the hooks
vi.mock('../hooks/useWorkflows', () => ({
  useWorkflow: vi.fn(),
  useWorkflowMutations: vi.fn(),
}))

import { useWorkflow, useWorkflowMutations } from '../hooks/useWorkflows'

describe('WorkflowEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('New workflow', () => {
    it('should render editor for new workflow', () => {
      (useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error: null,
      })

      ;(useWorkflowMutations as any).mockReturnValue({
        createWorkflow: vi.fn(),
        updateWorkflow: vi.fn(),
        creating: false,
        updating: false,
      })

      render(
        <MemoryRouter initialEntries={['/workflows/new']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByTestId('workflow-canvas')).toBeInTheDocument()
      expect(screen.getByTestId('node-palette')).toBeInTheDocument()
      expect(screen.getByTestId('property-panel')).toBeInTheDocument()
    })

    it('should have workflow name input', () => {
      (useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error: null,
      })

      ;(useWorkflowMutations as any).mockReturnValue({
        createWorkflow: vi.fn(),
        updateWorkflow: vi.fn(),
        creating: false,
        updating: false,
      })

      render(
        <MemoryRouter initialEntries={['/workflows/new']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByPlaceholderText(/workflow name/i)).toBeInTheDocument()
    })

    it('should have description textarea', () => {
      (useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error: null,
      })

      ;(useWorkflowMutations as any).mockReturnValue({
        createWorkflow: vi.fn(),
        updateWorkflow: vi.fn(),
        creating: false,
        updating: false,
      })

      render(
        <MemoryRouter initialEntries={['/workflows/new']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByPlaceholderText(/description/i)).toBeInTheDocument()
    })
  })

  describe('Edit existing workflow', () => {
    it('should load and display existing workflow', async () => {
      const mockWorkflow: Workflow = {
        id: 'wf-123',
        name: 'Test Workflow',
        description: 'Test Description',
        definition: {
          nodes: [],
          edges: [],
        },
        status: 'active',
        createdAt: '2025-01-15T10:00:00Z',
        updatedAt: '2025-01-15T10:00:00Z',
      }

      ;(useWorkflow as any).mockReturnValue({
        workflow: mockWorkflow,
        loading: false,
        error: null,
      })

      ;(useWorkflowMutations as any).mockReturnValue({
        createWorkflow: vi.fn(),
        updateWorkflow: vi.fn(),
        creating: false,
        updating: false,
      })

      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        const nameInput = screen.getByPlaceholderText(/workflow name/i) as HTMLInputElement
        expect(nameInput.value).toBe('Test Workflow')
      })
    })

    it('should show loading state while fetching workflow', () => {
      (useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: true,
        error: null,
      })

      ;(useWorkflowMutations as any).mockReturnValue({
        createWorkflow: vi.fn(),
        updateWorkflow: vi.fn(),
        creating: false,
        updating: false,
      })

      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByText(/loading/i)).toBeInTheDocument()
    })

    it('should show error if workflow load fails', () => {
      const error = new Error('Failed to load workflow')

      ;(useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error,
      })

      ;(useWorkflowMutations as any).mockReturnValue({
        createWorkflow: vi.fn(),
        updateWorkflow: vi.fn(),
        creating: false,
        updating: false,
      })

      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      const errorMessages = screen.getAllByText(/failed to load/i)
      expect(errorMessages.length).toBeGreaterThan(0)
    })
  })

  describe('Save functionality', () => {
    it('should have save button', () => {
      (useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error: null,
      })

      ;(useWorkflowMutations as any).mockReturnValue({
        createWorkflow: vi.fn(),
        updateWorkflow: vi.fn(),
        creating: false,
        updating: false,
      })

      render(
        <MemoryRouter initialEntries={['/workflows/new']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByRole('button', { name: /save workflow/i })).toBeInTheDocument()
    })

    it('should create new workflow when saving new', async () => {
      const user = userEvent.setup()
      const createWorkflow = vi.fn().mockResolvedValue({ id: 'new-wf-123' })

      ;(useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error: null,
      })

      ;(useWorkflowMutations as any).mockReturnValue({
        createWorkflow,
        updateWorkflow: vi.fn(),
        creating: false,
        updating: false,
      })

      render(
        <MemoryRouter initialEntries={['/workflows/new']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      const nameInput = screen.getByPlaceholderText(/workflow name/i)
      await user.type(nameInput, 'My New Workflow')

      const saveButton = screen.getByRole('button', { name: /save workflow/i })
      await user.click(saveButton)

      await waitFor(() => {
        expect(createWorkflow).toHaveBeenCalledWith(
          expect.objectContaining({
            name: 'My New Workflow',
          })
        )
      })
    })

    it('should update existing workflow when saving', async () => {
      const user = userEvent.setup()
      const updateWorkflow = vi.fn()

      const mockWorkflow: Workflow = {
        id: 'wf-123',
        name: 'Test Workflow',
        description: 'Test Description',
        definition: {
          nodes: [],
          edges: [],
        },
        status: 'active',
        createdAt: '2025-01-15T10:00:00Z',
        updatedAt: '2025-01-15T10:00:00Z',
      }

      ;(useWorkflow as any).mockReturnValue({
        workflow: mockWorkflow,
        loading: false,
        error: null,
      })

      ;(useWorkflowMutations as any).mockReturnValue({
        createWorkflow: vi.fn(),
        updateWorkflow,
        creating: false,
        updating: false,
      })

      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      const saveButton = screen.getByRole('button', { name: /save workflow/i })
      await user.click(saveButton)

      await waitFor(() => {
        expect(updateWorkflow).toHaveBeenCalledWith('wf-123', expect.any(Object))
      })
    })

    it('should show validation error if name is empty', async () => {
      const user = userEvent.setup()
      const createWorkflow = vi.fn()

      ;(useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error: null,
      })

      ;(useWorkflowMutations as any).mockReturnValue({
        createWorkflow,
        updateWorkflow: vi.fn(),
        creating: false,
        updating: false,
      })

      render(
        <MemoryRouter initialEntries={['/workflows/new']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      const saveButton = screen.getByRole('button', { name: /save workflow/i })
      await user.click(saveButton)

      await waitFor(() => {
        expect(screen.getByText(/workflow name is required/i)).toBeInTheDocument()
      })

      expect(createWorkflow).not.toHaveBeenCalled()
    })
  })

  describe('Navigation', () => {
    it('should have back button', () => {
      (useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error: null,
      })

      ;(useWorkflowMutations as any).mockReturnValue({
        createWorkflow: vi.fn(),
        updateWorkflow: vi.fn(),
        creating: false,
        updating: false,
      })

      render(
        <MemoryRouter initialEntries={['/workflows/new']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      const backLink = screen.getByRole('link', { name: /back to workflows/i })
      expect(backLink).toHaveAttribute('href', '/workflows')
    })
  })
})
