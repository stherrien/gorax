import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import userEvent from '@testing-library/user-event'
import WorkflowList from './WorkflowList'
import type { Workflow } from '../api/workflows'

// Mock the hooks
vi.mock('../hooks/useWorkflows', () => ({
  useWorkflows: vi.fn(),
  useWorkflowMutations: vi.fn(),
}))

import { useWorkflows, useWorkflowMutations } from '../hooks/useWorkflows'

describe('WorkflowList Integration', () => {
  const mockWorkflows: Workflow[] = [
    {
      id: 'wf-1',
      tenantId: 'tenant-1',
      name: 'Workflow 1',
      description: 'First workflow',
      status: 'active',
      definition: { nodes: [], edges: [] },
      version: 1,
      createdAt: '2024-01-15T10:00:00Z',
      updatedAt: '2024-01-15T10:00:00Z',
    },
    {
      id: 'wf-2',
      tenantId: 'tenant-1',
      name: 'Workflow 2',
      description: 'Second workflow',
      status: 'draft',
      definition: { nodes: [], edges: [] },
      version: 1,
      createdAt: '2024-01-14T10:00:00Z',
      updatedAt: '2024-01-14T10:00:00Z',
    },
  ]

  const mockMutations = {
    createWorkflow: vi.fn(),
    updateWorkflow: vi.fn(),
    deleteWorkflow: vi.fn(),
    executeWorkflow: vi.fn(),
    creating: false,
    updating: false,
    deleting: false,
    executing: false,
  }

  beforeEach(() => {
    vi.clearAllMocks()
    ;(useWorkflowMutations as any).mockReturnValue(mockMutations)
  })

  describe('Load workflows', () => {
    it('should display list of workflows from API', async () => {
      (useWorkflows as any).mockReturnValue({
        workflows: mockWorkflows,
        total: 2,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <WorkflowList />
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('Workflow 1')).toBeInTheDocument()
        expect(screen.getByText('Workflow 2')).toBeInTheDocument()
      })
    })

    it('should show loading state while fetching', () => {
      (useWorkflows as any).mockReturnValue({
        workflows: [],
        total: 0,
        loading: true,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <WorkflowList />
        </MemoryRouter>
      )

      expect(screen.getByText(/loading/i)).toBeInTheDocument()
    })

    it('should show error message if fetch fails', () => {
      const error = new Error('Failed to fetch workflows')
      ;(useWorkflows as any).mockReturnValue({
        workflows: [],
        total: 0,
        loading: false,
        error,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <WorkflowList />
        </MemoryRouter>
      )

      const errorTexts = screen.getAllByText(/failed to fetch/i)
      expect(errorTexts.length).toBeGreaterThan(0)
    })

    it('should show empty state when no workflows exist', () => {
      (useWorkflows as any).mockReturnValue({
        workflows: [],
        total: 0,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <WorkflowList />
        </MemoryRouter>
      )

      expect(screen.getByText(/no workflows/i)).toBeInTheDocument()
    })
  })

  describe('Workflow status badges', () => {
    it('should display correct status badge for active workflow', async () => {
      (useWorkflows as any).mockReturnValue({
        workflows: [mockWorkflows[0]], // active workflow
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <WorkflowList />
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('active')).toBeInTheDocument()
      })
    })

    it('should display correct status badge for draft workflow', async () => {
      (useWorkflows as any).mockReturnValue({
        workflows: [mockWorkflows[1]], // draft workflow
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <WorkflowList />
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('draft')).toBeInTheDocument()
      })
    })
  })

  describe('Delete workflow', () => {
    it('should delete workflow when delete button clicked', async () => {
      const user = userEvent.setup()
      const refetch = vi.fn()

      ;(useWorkflows as any).mockReturnValue({
        workflows: mockWorkflows,
        total: 2,
        loading: false,
        error: null,
        refetch,
      })

      mockMutations.deleteWorkflow.mockResolvedValueOnce(undefined)

      render(
        <MemoryRouter>
          <WorkflowList />
        </MemoryRouter>
      )

      // Find delete button for first workflow
      const deleteButtons = screen.getAllByRole('button', { name: /delete/i })
      await user.click(deleteButtons[0])

      // Should show confirmation dialog
      await waitFor(() => {
        expect(screen.getByText(/are you sure/i)).toBeInTheDocument()
      })

      // Confirm deletion
      const confirmButton = screen.getByRole('button', { name: /confirm/i })
      await user.click(confirmButton)

      await waitFor(() => {
        expect(mockMutations.deleteWorkflow).toHaveBeenCalledWith('wf-1')
      })

      // Should refetch workflows after deletion
      await waitFor(() => {
        expect(refetch).toHaveBeenCalled()
      })
    })

    it('should show error if delete fails', async () => {
      const user = userEvent.setup()

      ;(useWorkflows as any).mockReturnValue({
        workflows: mockWorkflows,
        total: 2,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      mockMutations.deleteWorkflow.mockRejectedValueOnce(new Error('Delete failed'))

      render(
        <MemoryRouter>
          <WorkflowList />
        </MemoryRouter>
      )

      const deleteButtons = screen.getAllByRole('button', { name: /delete/i })
      await user.click(deleteButtons[0])

      const confirmButton = screen.getByRole('button', { name: /confirm/i })
      await user.click(confirmButton)

      await waitFor(() => {
        expect(screen.getByText(/delete failed/i)).toBeInTheDocument()
      })
    })
  })

  describe('Navigation', () => {
    it('should have link to create new workflow', () => {
      (useWorkflows as any).mockReturnValue({
        workflows: mockWorkflows,
        total: 2,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <WorkflowList />
        </MemoryRouter>
      )

      const newButton = screen.getByRole('link', { name: /new workflow/i })
      expect(newButton).toHaveAttribute('href', '/workflows/new')
    })

    it('should have links to edit each workflow', async () => {
      (useWorkflows as any).mockReturnValue({
        workflows: mockWorkflows,
        total: 2,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <WorkflowList />
        </MemoryRouter>
      )

      await waitFor(() => {
        const editLinks = screen.getAllByRole('link', { name: /edit/i })
        expect(editLinks[0]).toHaveAttribute('href', '/workflows/wf-1')
        expect(editLinks[1]).toHaveAttribute('href', '/workflows/wf-2')
      })
    })
  })

  describe('Execute workflow', () => {
    it('should execute workflow when run button clicked', async () => {
      const user = userEvent.setup()

      ;(useWorkflows as any).mockReturnValue({
        workflows: [mockWorkflows[0]], // active workflow
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      mockMutations.executeWorkflow.mockResolvedValueOnce({
        executionId: 'exec-123',
        workflowId: 'wf-1',
        status: 'queued',
        queuedAt: '2024-01-15T10:00:00Z',
      })

      render(
        <MemoryRouter>
          <WorkflowList />
        </MemoryRouter>
      )

      const runButton = screen.getByRole('button', { name: /run/i })
      await user.click(runButton)

      await waitFor(() => {
        expect(mockMutations.executeWorkflow).toHaveBeenCalledWith('wf-1')
      })

      // Should show success message
      await waitFor(() => {
        expect(screen.getByText(/execution started/i)).toBeInTheDocument()
      })
    })

    it('should only show run button for active workflows', async () => {
      (useWorkflows as any).mockReturnValue({
        workflows: mockWorkflows, // one active, one draft
        total: 2,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <WorkflowList />
        </MemoryRouter>
      )

      await waitFor(() => {
        const runButtons = screen.queryAllByRole('button', { name: /run/i })
        // Should only have 1 run button (for active workflow)
        expect(runButtons.length).toBe(1)
      })
    })
  })
})
