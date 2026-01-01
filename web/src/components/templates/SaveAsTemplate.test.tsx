import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SaveAsTemplate } from './SaveAsTemplate'
import * as useTemplatesModule from '../../hooks/useTemplates'
import type { WorkflowDefinition } from '../../api/templates'

// Mock the useTemplates hook
vi.mock('../../hooks/useTemplates', () => ({
  useTemplateMutations: vi.fn(),
}))

describe('SaveAsTemplate', () => {
  const mockWorkflowDefinition: WorkflowDefinition = {
    nodes: [
      { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: { name: 'Trigger', config: {} } },
      { id: 'node-2', type: 'action', position: { x: 100, y: 100 }, data: { name: 'Action', config: {} } },
    ],
    edges: [
      { id: 'edge-1', source: 'node-1', target: 'node-2' },
    ],
  }

  const defaultProps = {
    workflowId: 'wf-123',
    workflowName: 'Test Workflow',
    definition: mockWorkflowDefinition,
  }

  let mockCreateFromWorkflow: ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.clearAllMocks()
    mockCreateFromWorkflow = vi.fn().mockResolvedValue({
      id: 'template-1',
      name: 'Test Template',
    })

    vi.mocked(useTemplatesModule.useTemplateMutations).mockReturnValue({
      createFromWorkflow: mockCreateFromWorkflow,
      creating: false,
      createTemplate: vi.fn(),
      updateTemplate: vi.fn(),
      deleteTemplate: vi.fn(),
      instantiateTemplate: vi.fn(),
      updating: false,
      deleting: false,
      instantiating: false,
    })
  })

  describe('rendering', () => {
    it('should render dialog with title', () => {
      render(<SaveAsTemplate {...defaultProps} />)

      expect(screen.getByText('Save as Template')).toBeInTheDocument()
    })

    it('should pre-fill name with workflow name', () => {
      render(<SaveAsTemplate {...defaultProps} />)

      const nameInput = screen.getByLabelText(/template name/i)
      expect(nameInput).toHaveValue('Test Workflow')
    })

    it('should display all category options', () => {
      render(<SaveAsTemplate {...defaultProps} />)

      expect(screen.getByText('Security')).toBeInTheDocument()
      expect(screen.getByText('Monitoring')).toBeInTheDocument()
      expect(screen.getByText('Integration')).toBeInTheDocument()
      expect(screen.getByText('Data Ops')).toBeInTheDocument()
      expect(screen.getByText('Dev Ops')).toBeInTheDocument()
      expect(screen.getByText('Other')).toBeInTheDocument()
    })

    it('should display template preview with node and edge counts', () => {
      render(<SaveAsTemplate {...defaultProps} />)

      expect(screen.getByText('Template Preview')).toBeInTheDocument()
      expect(screen.getByText('2')).toBeInTheDocument() // 2 nodes
      expect(screen.getByText('1')).toBeInTheDocument() // 1 edge
    })

    it('should show close button when onCancel is provided', () => {
      const onCancel = vi.fn()
      render(<SaveAsTemplate {...defaultProps} onCancel={onCancel} />)

      // Should have two close buttons - one in header and one as Cancel button
      const buttons = screen.getAllByRole('button')
      expect(buttons.length).toBeGreaterThan(1)
    })
  })

  describe('form validation', () => {
    it('should show error when name is empty after whitespace only', async () => {
      const user = userEvent.setup()
      render(<SaveAsTemplate {...defaultProps} />)

      // Clear the name field and add just whitespace
      const nameInput = screen.getByLabelText(/template name/i)
      await user.clear(nameInput)
      await user.type(nameInput, '   ') // Whitespace only - passes HTML required but fails our validation

      // Submit form
      await user.click(screen.getByRole('button', { name: /save template/i }))

      expect(await screen.findByText('Template name is required')).toBeInTheDocument()
      expect(mockCreateFromWorkflow).not.toHaveBeenCalled()
    })
  })

  describe('tag management', () => {
    it('should add tags when clicking add button', async () => {
      const user = userEvent.setup()
      render(<SaveAsTemplate {...defaultProps} />)

      const tagInput = screen.getByPlaceholderText(/add a tag/i)
      await user.type(tagInput, 'automation')
      await user.click(screen.getByRole('button', { name: /add/i }))

      expect(screen.getByText('automation')).toBeInTheDocument()
    })

    it('should add tags when pressing Enter', async () => {
      const user = userEvent.setup()
      render(<SaveAsTemplate {...defaultProps} />)

      const tagInput = screen.getByPlaceholderText(/add a tag/i)
      await user.type(tagInput, 'workflow{enter}')

      expect(screen.getByText('workflow')).toBeInTheDocument()
    })

    it('should remove tags when clicking remove button', async () => {
      const user = userEvent.setup()
      render(<SaveAsTemplate {...defaultProps} />)

      // Add a tag
      const tagInput = screen.getByPlaceholderText(/add a tag/i)
      await user.type(tagInput, 'test-tag')
      await user.click(screen.getByRole('button', { name: /add/i }))

      // Verify tag is displayed
      expect(screen.getByText('test-tag')).toBeInTheDocument()

      // Remove the tag
      await user.click(screen.getByRole('button', { name: /remove test-tag tag/i }))

      // Verify tag is removed
      expect(screen.queryByText('test-tag')).not.toBeInTheDocument()
    })

    it('should not add duplicate tags', async () => {
      const user = userEvent.setup()
      render(<SaveAsTemplate {...defaultProps} />)

      const tagInput = screen.getByPlaceholderText(/add a tag/i)

      // Add same tag twice
      await user.type(tagInput, 'duplicate')
      await user.click(screen.getByRole('button', { name: /add/i }))
      await user.type(tagInput, 'duplicate')
      await user.click(screen.getByRole('button', { name: /add/i }))

      // Should only appear once
      const tags = screen.getAllByText('duplicate')
      expect(tags.length).toBe(1)
    })

    it('should not add empty tags', async () => {
      const user = userEvent.setup()
      render(<SaveAsTemplate {...defaultProps} />)

      // Click add without typing anything
      const addButton = screen.getByRole('button', { name: /add/i })
      expect(addButton).toBeDisabled()
    })
  })

  describe('form submission', () => {
    it('should submit form with correct data', async () => {
      const user = userEvent.setup()
      const onSuccess = vi.fn()
      render(<SaveAsTemplate {...defaultProps} onSuccess={onSuccess} />)

      // Fill out form
      const nameInput = screen.getByLabelText(/template name/i)
      await user.clear(nameInput)
      await user.type(nameInput, 'My Template')

      const descInput = screen.getByLabelText(/description/i)
      await user.type(descInput, 'Template description')

      const categorySelect = screen.getByLabelText(/category/i)
      await user.selectOptions(categorySelect, 'security')

      // Add a tag
      const tagInput = screen.getByPlaceholderText(/add a tag/i)
      await user.type(tagInput, 'test')
      await user.click(screen.getByRole('button', { name: /add/i }))

      // Check public checkbox
      const publicCheckbox = screen.getByRole('checkbox')
      await user.click(publicCheckbox)

      // Submit
      await user.click(screen.getByRole('button', { name: /save template/i }))

      await waitFor(() => {
        expect(mockCreateFromWorkflow).toHaveBeenCalledWith('wf-123', {
          name: 'My Template',
          description: 'Template description',
          category: 'security',
          definition: mockWorkflowDefinition,
          tags: ['test'],
          isPublic: true,
        })
      })

      await waitFor(() => {
        expect(onSuccess).toHaveBeenCalled()
      })
    })

    it('should show saving state while creating', () => {
      vi.mocked(useTemplatesModule.useTemplateMutations).mockReturnValue({
        createFromWorkflow: mockCreateFromWorkflow,
        creating: true,
        createTemplate: vi.fn(),
        updateTemplate: vi.fn(),
        deleteTemplate: vi.fn(),
        instantiateTemplate: vi.fn(),
        updating: false,
        deleting: false,
        instantiating: false,
      })

      render(<SaveAsTemplate {...defaultProps} />)

      expect(screen.getByRole('button', { name: /saving/i })).toBeDisabled()
    })

    it('should display error when submission fails', async () => {
      const user = userEvent.setup()
      mockCreateFromWorkflow.mockRejectedValueOnce(new Error('API Error'))

      render(<SaveAsTemplate {...defaultProps} />)

      await user.click(screen.getByRole('button', { name: /save template/i }))

      await waitFor(() => {
        expect(screen.getByText('API Error')).toBeInTheDocument()
      })
    })

    it('should display generic error for non-Error objects', async () => {
      const user = userEvent.setup()
      mockCreateFromWorkflow.mockRejectedValueOnce('Unknown error')

      render(<SaveAsTemplate {...defaultProps} />)

      await user.click(screen.getByRole('button', { name: /save template/i }))

      await waitFor(() => {
        expect(screen.getByText('Failed to save template')).toBeInTheDocument()
      })
    })
  })

  describe('cancel behavior', () => {
    it('should call onCancel when cancel button is clicked', async () => {
      const user = userEvent.setup()
      const onCancel = vi.fn()
      render(<SaveAsTemplate {...defaultProps} onCancel={onCancel} />)

      await user.click(screen.getByRole('button', { name: /cancel/i }))

      expect(onCancel).toHaveBeenCalledTimes(1)
    })

    it('should disable cancel button while creating', () => {
      vi.mocked(useTemplatesModule.useTemplateMutations).mockReturnValue({
        createFromWorkflow: mockCreateFromWorkflow,
        creating: true,
        createTemplate: vi.fn(),
        updateTemplate: vi.fn(),
        deleteTemplate: vi.fn(),
        instantiateTemplate: vi.fn(),
        updating: false,
        deleting: false,
        instantiating: false,
      })

      const onCancel = vi.fn()
      render(<SaveAsTemplate {...defaultProps} onCancel={onCancel} />)

      expect(screen.getByRole('button', { name: /cancel/i })).toBeDisabled()
    })
  })
})
