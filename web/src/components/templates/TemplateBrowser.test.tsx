import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { TemplateBrowser } from './TemplateBrowser'
import * as useTemplatesModule from '../../hooks/useTemplates'
import type { Template } from '../../api/templates'

// Mock the useTemplates hook
vi.mock('../../hooks/useTemplates', () => ({
  useTemplates: vi.fn(),
  useTemplateMutations: vi.fn(),
}))

describe('TemplateBrowser', () => {
  const mockTemplates: Template[] = [
    {
      id: 'template-1',
      name: 'Security Scan',
      description: 'Run security scans on your infrastructure',
      category: 'security',
      tags: ['security', 'scan'],
      isPublic: true,
      createdBy: 'user-1',
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
      definition: {
        nodes: [
          { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: { name: 'Start', config: {} } },
          { id: 'node-2', type: 'action', position: { x: 100, y: 100 }, data: { name: 'Scan', config: {} } },
        ],
        edges: [{ id: 'edge-1', source: 'node-1', target: 'node-2' }],
      },
    },
    {
      id: 'template-2',
      name: 'Data Pipeline',
      description: 'ETL workflow for data processing',
      category: 'dataops',
      tags: ['data', 'etl'],
      isPublic: false,
      createdBy: 'user-2',
      createdAt: '2025-01-02T00:00:00Z',
      updatedAt: '2025-01-02T00:00:00Z',
      definition: {
        nodes: [
          { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: { name: 'Trigger', config: {} } },
        ],
        edges: [],
      },
    },
    {
      id: 'template-3',
      name: 'Monitoring Setup',
      description: 'Set up monitoring for services',
      category: 'monitoring',
      tags: ['monitoring', 'alerts'],
      isPublic: true,
      createdBy: 'user-1',
      createdAt: '2025-01-03T00:00:00Z',
      updatedAt: '2025-01-03T00:00:00Z',
      definition: {
        nodes: [
          { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: { name: 'Start', config: {} } },
          { id: 'node-2', type: 'action', position: { x: 100, y: 100 }, data: { name: 'Setup', config: {} } },
          { id: 'node-3', type: 'action', position: { x: 200, y: 200 }, data: { name: 'Alert', config: {} } },
        ],
        edges: [
          { id: 'edge-1', source: 'node-1', target: 'node-2' },
          { id: 'edge-2', source: 'node-2', target: 'node-3' },
        ],
      },
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()

    vi.mocked(useTemplatesModule.useTemplates).mockReturnValue({
      templates: mockTemplates,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    vi.mocked(useTemplatesModule.useTemplateMutations).mockReturnValue({
      instantiating: false,
      creating: false,
      updating: false,
      deleting: false,
      createTemplate: vi.fn(),
      updateTemplate: vi.fn(),
      deleteTemplate: vi.fn(),
      createFromWorkflow: vi.fn(),
      instantiateTemplate: vi.fn(),
    })
  })

  describe('rendering', () => {
    it('should render with title', () => {
      render(<TemplateBrowser />)

      expect(screen.getByText('Template Library')).toBeInTheDocument()
    })

    it('should render search input', () => {
      render(<TemplateBrowser />)

      expect(screen.getByPlaceholderText(/search templates/i)).toBeInTheDocument()
    })

    it('should render all category filter buttons', () => {
      render(<TemplateBrowser />)

      expect(screen.getByRole('button', { name: /all/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /security/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /monitoring/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /integration/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /data ops/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /dev ops/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /other/i })).toBeInTheDocument()
    })

    it('should render template cards', () => {
      render(<TemplateBrowser />)

      expect(screen.getByText('Security Scan')).toBeInTheDocument()
      expect(screen.getByText('Data Pipeline')).toBeInTheDocument()
      expect(screen.getByText('Monitoring Setup')).toBeInTheDocument()
    })

    it('should show close button when onClose is provided', () => {
      const onClose = vi.fn()
      render(<TemplateBrowser onClose={onClose} />)

      const closeButton = screen.getAllByRole('button').find(
        btn => btn.textContent === '✕'
      )
      expect(closeButton).toBeDefined()
    })
  })

  describe('loading state', () => {
    it('should show loading message when loading', () => {
      vi.mocked(useTemplatesModule.useTemplates).mockReturnValue({
        templates: [],
        loading: true,
        error: null,
        refetch: vi.fn(),
      })

      render(<TemplateBrowser />)

      expect(screen.getByText(/loading templates/i)).toBeInTheDocument()
    })
  })

  describe('error state', () => {
    it('should show error message when there is an error', () => {
      vi.mocked(useTemplatesModule.useTemplates).mockReturnValue({
        templates: [],
        loading: false,
        error: new Error('Failed to fetch'),
        refetch: vi.fn(),
      })

      render(<TemplateBrowser />)

      expect(screen.getByText(/failed to load templates/i)).toBeInTheDocument()
      // The error message includes the error text
      const errorElement = document.querySelector('.error-message')
      expect(errorElement?.textContent).toContain('Failed to fetch')
    })
  })

  describe('empty state', () => {
    it('should show empty state when no templates found', () => {
      vi.mocked(useTemplatesModule.useTemplates).mockReturnValue({
        templates: [],
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(<TemplateBrowser />)

      expect(screen.getByText('No templates found')).toBeInTheDocument()
      expect(screen.getByText(/try adjusting your filters/i)).toBeInTheDocument()
    })
  })

  describe('template cards', () => {
    it('should display template name and description', () => {
      render(<TemplateBrowser />)

      expect(screen.getByText('Security Scan')).toBeInTheDocument()
      expect(screen.getByText('Run security scans on your infrastructure')).toBeInTheDocument()
    })

    it('should display template category in card', () => {
      render(<TemplateBrowser />)

      // Check category labels within template cards (in .template-category spans)
      const categorySpans = document.querySelectorAll('.template-category')
      expect(categorySpans.length).toBe(3)

      const categoryTexts = Array.from(categorySpans).map(span => span.textContent)
      expect(categoryTexts).toContain('security')
      expect(categoryTexts).toContain('dataops')
      expect(categoryTexts).toContain('monitoring')
    })

    it('should display node and connection counts', () => {
      render(<TemplateBrowser />)

      // Security Scan: 2 nodes, 1 connection
      expect(screen.getByText('2 nodes')).toBeInTheDocument()
      expect(screen.getByText('1 connections')).toBeInTheDocument()

      // Monitoring Setup: 3 nodes, 2 connections
      expect(screen.getByText('3 nodes')).toBeInTheDocument()
      expect(screen.getByText('2 connections')).toBeInTheDocument()
    })

    it('should display tags', () => {
      render(<TemplateBrowser />)

      // Check tags within .template-tags div (to distinguish from category filter buttons)
      const tagElements = document.querySelectorAll('.template-tags .tag')
      const tagTexts = Array.from(tagElements).map(el => el.textContent)

      expect(tagTexts).toContain('security')
      expect(tagTexts).toContain('scan')
      expect(tagTexts).toContain('data')
      expect(tagTexts).toContain('etl')
    })

    it('should show public badge for public templates', () => {
      render(<TemplateBrowser />)

      const publicBadges = screen.getAllByText('Public Template')
      // Security Scan and Monitoring Setup are public
      expect(publicBadges.length).toBe(2)
    })
  })

  describe('filtering', () => {
    it('should filter by category when category button is clicked', async () => {
      const user = userEvent.setup()
      render(<TemplateBrowser />)

      await user.click(screen.getByRole('button', { name: /security/i }))

      // useTemplates should be called with the selected category
      expect(useTemplatesModule.useTemplates).toHaveBeenCalledWith({
        category: 'security',
        search: undefined,
      })
    })

    it('should clear category filter when All is clicked', async () => {
      const user = userEvent.setup()
      render(<TemplateBrowser />)

      // First select a category
      await user.click(screen.getByRole('button', { name: /security/i }))

      // Then click All
      await user.click(screen.getByRole('button', { name: /all/i }))

      // useTemplates should be called with empty category
      expect(useTemplatesModule.useTemplates).toHaveBeenLastCalledWith({
        category: undefined,
        search: undefined,
      })
    })

    it('should filter by search query', async () => {
      const user = userEvent.setup()
      render(<TemplateBrowser />)

      const searchInput = screen.getByPlaceholderText(/search templates/i)
      await user.type(searchInput, 'monitoring')

      expect(useTemplatesModule.useTemplates).toHaveBeenCalledWith({
        category: undefined,
        search: 'monitoring',
      })
    })
  })

  describe('template selection', () => {
    it('should call onSelectTemplate when Use Template is clicked', async () => {
      const user = userEvent.setup()
      const onSelectTemplate = vi.fn()
      render(<TemplateBrowser onSelectTemplate={onSelectTemplate} />)

      const useButtons = screen.getAllByRole('button', { name: /use template/i })
      await user.click(useButtons[0])

      expect(onSelectTemplate).toHaveBeenCalledWith(mockTemplates[0])
    })

    it('should show using state when instantiating', () => {
      vi.mocked(useTemplatesModule.useTemplateMutations).mockReturnValue({
        instantiating: true,
        creating: false,
        updating: false,
        deleting: false,
        createTemplate: vi.fn(),
        updateTemplate: vi.fn(),
        deleteTemplate: vi.fn(),
        createFromWorkflow: vi.fn(),
        instantiateTemplate: vi.fn(),
      })

      render(<TemplateBrowser />)

      const usingButtons = screen.getAllByRole('button', { name: /using/i })
      expect(usingButtons.length).toBeGreaterThan(0)
      usingButtons.forEach(btn => expect(btn).toBeDisabled())
    })
  })

  describe('template details', () => {
    it('should toggle details when View Details is clicked', async () => {
      const user = userEvent.setup()
      render(<TemplateBrowser />)

      const viewDetailsButtons = screen.getAllByRole('button', { name: /view details/i })
      await user.click(viewDetailsButtons[0])

      // Should show template structure
      expect(screen.getByText('Template Structure')).toBeInTheDocument()

      // Should show node types
      expect(screen.getByText('trigger')).toBeInTheDocument()
      expect(screen.getByText('action')).toBeInTheDocument()
    })

    it('should hide details when Hide Details is clicked', async () => {
      const user = userEvent.setup()
      render(<TemplateBrowser />)

      const viewDetailsButtons = screen.getAllByRole('button', { name: /view details/i })
      await user.click(viewDetailsButtons[0])

      expect(screen.getByText('Template Structure')).toBeInTheDocument()

      const hideDetailsButton = screen.getByRole('button', { name: /hide details/i })
      await user.click(hideDetailsButton)

      expect(screen.queryByText('Template Structure')).not.toBeInTheDocument()
    })
  })

  describe('close behavior', () => {
    it('should call onClose when close button is clicked', async () => {
      const user = userEvent.setup()
      const onClose = vi.fn()
      render(<TemplateBrowser onClose={onClose} />)

      const closeButton = screen.getAllByRole('button').find(
        btn => btn.textContent === '✕'
      )
      if (closeButton) {
        await user.click(closeButton)
        expect(onClose).toHaveBeenCalledTimes(1)
      }
    })
  })
})
