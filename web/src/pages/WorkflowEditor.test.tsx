import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
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

vi.mock('../components/workflow/VersionHistory', () => ({
  default: vi.fn(({ workflowId, currentVersion, onRestore, onClose }) => (
    <div data-testid="version-history">
      <button onClick={() => onRestore(1)}>Restore Version 1</button>
      <button onClick={onClose}>Close Version History</button>
    </div>
  )),
}))

vi.mock('../components/templates', () => ({
  TemplateBrowser: vi.fn(({ onSelectTemplate, onClose }) => (
    <div data-testid="template-browser">
      <button onClick={() => onSelectTemplate({ id: 'tpl-1', name: 'Test Template' })}>
        Select Template
      </button>
      <button onClick={onClose}>Close Template Browser</button>
    </div>
  )),
  SaveAsTemplate: vi.fn(({ onSuccess, onCancel }) => (
    <div data-testid="save-as-template">
      <button onClick={onSuccess}>Save Template</button>
      <button onClick={onCancel}>Cancel Save Template</button>
    </div>
  )),
}))

// Mock the hooks
vi.mock('../hooks/useWorkflows', () => ({
  useWorkflow: vi.fn(),
  useWorkflowMutations: vi.fn(),
}))

vi.mock('../hooks/useTemplates', () => ({
  useTemplateMutations: vi.fn(() => ({
    instantiateTemplate: vi.fn(),
  })),
}))

// Mock the workflow API
vi.mock('../api/workflows', () => ({
  workflowAPI: {
    dryRun: vi.fn(),
  },
}))

import { useWorkflow, useWorkflowMutations } from '../hooks/useWorkflows'
import { useTemplateMutations } from '../hooks/useTemplates'
import { workflowAPI } from '../api/workflows'

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

      (useWorkflowMutations as any).mockReturnValue({
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

      (useWorkflowMutations as any).mockReturnValue({
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

      (useWorkflowMutations as any).mockReturnValue({
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

      (useWorkflow as any).mockReturnValue({
        workflow: mockWorkflow,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      (useWorkflowMutations as any).mockReturnValue({
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

      (useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      (useWorkflowMutations as any).mockReturnValue({
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

      (useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      (useWorkflow as any).mockReturnValue({
        workflow: mockWorkflow,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      (useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      (useWorkflowMutations as any).mockReturnValue({
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

  describe('Dry Run / Test Workflow', () => {
    const mockWorkflow: Workflow = {
      id: 'wf-123',
      name: 'Test Workflow',
      description: 'Test Description',
      definition: { nodes: [], edges: [] },
      status: 'active',
      version: 1,
      createdAt: '2025-01-15T10:00:00Z',
      updatedAt: '2025-01-15T10:00:00Z',
    }

    beforeEach(() => {
      (useWorkflow as any).mockReturnValue({
        workflow: mockWorkflow,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
        createWorkflow: vi.fn(),
        updateWorkflow: vi.fn(),
        creating: false,
        updating: false,
      })
    })

    it('should show Test Workflow button for existing workflows', () => {
      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByRole('button', { name: /test workflow/i })).toBeInTheDocument()
    })

    it('should not show Test Workflow button for new workflows', () => {
      (useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error: null,
      })

      render(
        <MemoryRouter initialEntries={['/workflows/new']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      expect(screen.queryByRole('button', { name: /test workflow/i })).not.toBeInTheDocument()
    })

    it('should show dry run results on successful test', async () => {
      const user = userEvent.setup()
      const mockDryRunResult = {
        valid: true,
        executionOrder: ['node-1', 'node-2'],
        errors: [],
        warnings: [],
        variableMapping: { '{{trigger.data}}': 'trigger_node' },
      }

      vi.mocked(workflowAPI.dryRun).mockResolvedValue(mockDryRunResult)

      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      await user.click(screen.getByRole('button', { name: /test workflow/i }))

      await waitFor(() => {
        expect(screen.getByText(/workflow test results/i)).toBeInTheDocument()
      })

      expect(screen.getByText(/workflow is valid/i)).toBeInTheDocument()
    })

    it('should show errors in dry run results', async () => {
      const user = userEvent.setup()
      const mockDryRunResult = {
        valid: false,
        executionOrder: [],
        errors: [{ nodeId: 'node-1', field: 'url', message: 'URL is required' }],
        warnings: [],
        variableMapping: {},
      }

      vi.mocked(workflowAPI.dryRun).mockResolvedValue(mockDryRunResult)

      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      await user.click(screen.getByRole('button', { name: /test workflow/i }))

      await waitFor(() => {
        expect(screen.getByText(/workflow has errors/i)).toBeInTheDocument()
      })

      expect(screen.getByText(/url is required/i)).toBeInTheDocument()
    })

    it('should show warnings in dry run results', async () => {
      const user = userEvent.setup()
      const mockDryRunResult = {
        valid: true,
        executionOrder: ['node-1'],
        errors: [],
        warnings: [{ nodeId: 'node-1', message: 'Deprecated action type' }],
        variableMapping: {},
      }

      vi.mocked(workflowAPI.dryRun).mockResolvedValue(mockDryRunResult)

      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      await user.click(screen.getByRole('button', { name: /test workflow/i }))

      await waitFor(() => {
        expect(screen.getByText(/deprecated action type/i)).toBeInTheDocument()
      })
    })

    it('should show error message when dry run fails', async () => {
      const user = userEvent.setup()
      vi.mocked(workflowAPI.dryRun).mockRejectedValue(new Error('Server error'))

      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      await user.click(screen.getByRole('button', { name: /test workflow/i }))

      await waitFor(() => {
        expect(screen.getByText(/server error/i)).toBeInTheDocument()
      })
    })

    it('should close dry run results modal', async () => {
      const user = userEvent.setup()
      const mockDryRunResult = {
        valid: true,
        executionOrder: ['node-1'],
        errors: [],
        warnings: [],
        variableMapping: {},
      }

      vi.mocked(workflowAPI.dryRun).mockResolvedValue(mockDryRunResult)

      render(
        <MemoryRouter initialEntries={['/workflows/wf-123']}>
          <Routes>
            <Route path="/workflows/:id" element={<WorkflowEditor />} />
          </Routes>
        </MemoryRouter>
      )

      await user.click(screen.getByRole('button', { name: /test workflow/i }))

      await waitFor(() => {
        expect(screen.getByText(/workflow test results/i)).toBeInTheDocument()
      })

      // Click close button
      const closeButtons = screen.getAllByRole('button', { name: /close/i })
      await user.click(closeButtons[closeButtons.length - 1])

      await waitFor(() => {
        expect(screen.queryByText(/workflow test results/i)).not.toBeInTheDocument()
      })
    })
  })

  describe('Template Browser', () => {
    it('should show Browse Templates button', () => {
      (useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      expect(screen.getByRole('button', { name: /browse templates/i })).toBeInTheDocument()
    })

    it('should open template browser when button clicked', async () => {
      const user = userEvent.setup()

      (useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      await user.click(screen.getByRole('button', { name: /browse templates/i }))

      await waitFor(() => {
        expect(screen.getByTestId('template-browser')).toBeInTheDocument()
      })
    })

    it('should close template browser', async () => {
      const user = userEvent.setup()

      (useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      await user.click(screen.getByRole('button', { name: /browse templates/i }))

      await waitFor(() => {
        expect(screen.getByTestId('template-browser')).toBeInTheDocument()
      })

      await user.click(screen.getByText(/close template browser/i))

      await waitFor(() => {
        expect(screen.queryByTestId('template-browser')).not.toBeInTheDocument()
      })
    })

    it('should load template when selected', async () => {
      const user = userEvent.setup()
      const instantiateTemplate = vi.fn().mockResolvedValue({
        definition: { nodes: [{ id: 'n1' }], edges: [] },
        workflowName: 'Template Workflow',
      })

      (useTemplateMutations as any).mockReturnValue({ instantiateTemplate })

      (useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      await user.click(screen.getByRole('button', { name: /browse templates/i }))
      await user.click(screen.getByText(/select template/i))

      await waitFor(() => {
        expect(instantiateTemplate).toHaveBeenCalledWith('tpl-1', expect.any(Object))
      })
    })
  })

  describe('Save as Template', () => {
    const mockWorkflow: Workflow = {
      id: 'wf-123',
      name: 'Test Workflow',
      description: 'Test Description',
      definition: { nodes: [], edges: [] },
      status: 'active',
      version: 1,
      createdAt: '2025-01-15T10:00:00Z',
      updatedAt: '2025-01-15T10:00:00Z',
    }

    it('should show Save as Template button for existing workflows', () => {
      (useWorkflow as any).mockReturnValue({
        workflow: mockWorkflow,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      expect(screen.getByRole('button', { name: /save as template/i })).toBeInTheDocument()
    })

    it('should not show Save as Template button for new workflows', () => {
      (useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      expect(screen.queryByRole('button', { name: /save as template/i })).not.toBeInTheDocument()
    })

    it('should open Save as Template modal when clicked', async () => {
      const user = userEvent.setup()

      (useWorkflow as any).mockReturnValue({
        workflow: mockWorkflow,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      await user.click(screen.getByRole('button', { name: /save as template/i }))

      await waitFor(() => {
        expect(screen.getByTestId('save-as-template')).toBeInTheDocument()
      })
    })

    it('should close Save as Template modal on cancel', async () => {
      const user = userEvent.setup()

      (useWorkflow as any).mockReturnValue({
        workflow: mockWorkflow,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      await user.click(screen.getByRole('button', { name: /save as template/i }))
      await user.click(screen.getByText(/cancel save template/i))

      await waitFor(() => {
        expect(screen.queryByTestId('save-as-template')).not.toBeInTheDocument()
      })
    })

    it('should show success message when template saved', async () => {
      const user = userEvent.setup()

      (useWorkflow as any).mockReturnValue({
        workflow: mockWorkflow,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      await user.click(screen.getByRole('button', { name: /save as template/i }))

      const modal = screen.getByTestId('save-as-template')
      await user.click(within(modal).getByRole('button', { name: /^save template$/i }))

      await waitFor(() => {
        expect(screen.getByText(/template saved successfully/i)).toBeInTheDocument()
      })
    })
  })

  describe('Version History', () => {
    const mockWorkflow: Workflow = {
      id: 'wf-123',
      name: 'Test Workflow',
      description: 'Test Description',
      definition: { nodes: [], edges: [] },
      status: 'active',
      version: 3,
      createdAt: '2025-01-15T10:00:00Z',
      updatedAt: '2025-01-15T10:00:00Z',
    }

    it('should show Version History button for existing workflows', () => {
      (useWorkflow as any).mockReturnValue({
        workflow: mockWorkflow,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      expect(screen.getByRole('button', { name: /version history/i })).toBeInTheDocument()
    })

    it('should show version number for existing workflows', () => {
      (useWorkflow as any).mockReturnValue({
        workflow: mockWorkflow,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      expect(screen.getByText(/v3/i)).toBeInTheDocument()
    })

    it('should open Version History panel when clicked', async () => {
      const user = userEvent.setup()

      (useWorkflow as any).mockReturnValue({
        workflow: mockWorkflow,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      await user.click(screen.getByRole('button', { name: /version history/i }))

      await waitFor(() => {
        expect(screen.getByTestId('version-history')).toBeInTheDocument()
      })
    })

    it('should close Version History panel', async () => {
      const user = userEvent.setup()

      (useWorkflow as any).mockReturnValue({
        workflow: mockWorkflow,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      await user.click(screen.getByRole('button', { name: /version history/i }))
      await user.click(screen.getByText(/close version history/i))

      await waitFor(() => {
        expect(screen.queryByTestId('version-history')).not.toBeInTheDocument()
      })
    })
  })

  describe('Save error handling', () => {
    it('should show error message when save fails', async () => {
      const user = userEvent.setup()
      const createWorkflow = vi.fn().mockRejectedValue(new Error('Network error'))

      (useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      await user.type(screen.getByPlaceholderText(/workflow name/i), 'Test Workflow')
      await user.click(screen.getByRole('button', { name: /save workflow/i }))

      await waitFor(() => {
        expect(screen.getByText(/network error/i)).toBeInTheDocument()
      })
    })

    it('should show success message when save succeeds', async () => {
      const user = userEvent.setup()
      const updateWorkflow = vi.fn().mockResolvedValue({})

      const mockWorkflow: Workflow = {
        id: 'wf-123',
        name: 'Test Workflow',
        description: '',
        definition: { nodes: [], edges: [] },
        status: 'active',
        createdAt: '2025-01-15T10:00:00Z',
        updatedAt: '2025-01-15T10:00:00Z',
      }

      (useWorkflow as any).mockReturnValue({
        workflow: mockWorkflow,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      await user.click(screen.getByRole('button', { name: /save workflow/i }))

      await waitFor(() => {
        expect(screen.getByText(/workflow saved successfully/i)).toBeInTheDocument()
      })
    })

    it('should clear validation error when name is entered', async () => {
      const user = userEvent.setup()

      (useWorkflow as any).mockReturnValue({
        workflow: null,
        loading: false,
        error: null,
      })

      (useWorkflowMutations as any).mockReturnValue({
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

      // Trigger validation error
      await user.click(screen.getByRole('button', { name: /save workflow/i }))
      expect(screen.getByText(/workflow name is required/i)).toBeInTheDocument()

      // Enter name - should clear error
      await user.type(screen.getByPlaceholderText(/workflow name/i), 'Test')

      await waitFor(() => {
        expect(screen.queryByText(/workflow name is required/i)).not.toBeInTheDocument()
      })
    })
  })
})
