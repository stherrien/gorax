import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import TopFailuresTable from './TopFailuresTable'
import { TopFailure } from '../../api/metrics'

// Mock useNavigate
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

describe('TopFailuresTable', () => {
  const mockFailures: TopFailure[] = [
    {
      workflowId: 'wf-12345678-abcd',
      workflowName: 'Data Processing Workflow',
      failureCount: 25,
      lastFailedAt: '2025-01-15T10:30:00Z',
      errorPreview: 'Connection timeout: Unable to reach database server',
    },
    {
      workflowId: 'wf-87654321-efgh',
      workflowName: 'Email Notification Workflow',
      failureCount: 10,
      lastFailedAt: '2025-01-14T15:45:00Z',
      errorPreview: 'SMTP authentication failed',
    },
  ]

  const renderWithRouter = (ui: React.ReactNode) => {
    return render(<MemoryRouter>{ui}</MemoryRouter>)
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Loading state', () => {
    it('should show loading spinner when loading', () => {
      renderWithRouter(<TopFailuresTable failures={[]} loading={true} />)

      expect(screen.getByText('Top Failures')).toBeInTheDocument()
      const spinner = document.querySelector('.animate-spin')
      expect(spinner).toBeInTheDocument()
    })

    it('should not show table when loading', () => {
      renderWithRouter(<TopFailuresTable failures={mockFailures} loading={true} />)

      expect(screen.queryByRole('table')).not.toBeInTheDocument()
    })
  })

  describe('Error state', () => {
    it('should display error message', () => {
      renderWithRouter(<TopFailuresTable failures={[]} error="Failed to load failures" />)

      expect(screen.getByText('Failed to load failures')).toBeInTheDocument()
    })

    it('should show title with error', () => {
      renderWithRouter(<TopFailuresTable failures={[]} error="Error" />)

      expect(screen.getByText('Top Failures')).toBeInTheDocument()
    })
  })

  describe('Empty state', () => {
    it('should display success message when no failures', () => {
      renderWithRouter(<TopFailuresTable failures={[]} />)

      expect(screen.getByText('No failures found')).toBeInTheDocument()
      expect(screen.getByText('All workflows are running smoothly')).toBeInTheDocument()
    })

    it('should show success icon in empty state', () => {
      renderWithRouter(<TopFailuresTable failures={[]} />)

      // SVG checkmark icon should be present
      const svg = document.querySelector('svg')
      expect(svg).toBeInTheDocument()
    })
  })

  describe('Table rendering', () => {
    it('should render table with data', () => {
      renderWithRouter(<TopFailuresTable failures={mockFailures} />)

      expect(screen.getByRole('table')).toBeInTheDocument()
    })

    it('should display workflow names', () => {
      renderWithRouter(<TopFailuresTable failures={mockFailures} />)

      expect(screen.getByText('Data Processing Workflow')).toBeInTheDocument()
      expect(screen.getByText('Email Notification Workflow')).toBeInTheDocument()
    })

    it('should display failure counts', () => {
      renderWithRouter(<TopFailuresTable failures={mockFailures} />)

      expect(screen.getByText('25')).toBeInTheDocument()
      expect(screen.getByText('10')).toBeInTheDocument()
    })

    it('should display truncated workflow IDs', () => {
      renderWithRouter(<TopFailuresTable failures={mockFailures} />)

      // substring(0, 8) + '...' = 'wf-12345' + '...'
      expect(screen.getByText('wf-12345...')).toBeInTheDocument()
      expect(screen.getByText('wf-87654...')).toBeInTheDocument()
    })

    it('should display table headers', () => {
      renderWithRouter(<TopFailuresTable failures={mockFailures} />)

      expect(screen.getByText('Workflow')).toBeInTheDocument()
      expect(screen.getByText('Failures')).toBeInTheDocument()
      expect(screen.getByText('Last Failed')).toBeInTheDocument()
      expect(screen.getByText('Error Preview')).toBeInTheDocument()
      expect(screen.getByText('Actions')).toBeInTheDocument()
    })

    it('should display error previews', () => {
      renderWithRouter(<TopFailuresTable failures={mockFailures} />)

      expect(screen.getByText(/Connection timeout/)).toBeInTheDocument()
      expect(screen.getByText(/SMTP authentication failed/)).toBeInTheDocument()
    })
  })

  describe('Date formatting', () => {
    it('should format date correctly', () => {
      renderWithRouter(<TopFailuresTable failures={mockFailures} />)

      // Should show formatted date like "Jan 15, 10:30 AM"
      const formattedDates = screen.getAllByText(/Jan \d+/)
      expect(formattedDates.length).toBeGreaterThan(0)
    })

    it('should show N/A for missing date', () => {
      const failuresWithoutDate: TopFailure[] = [
        {
          workflowId: 'wf-123',
          workflowName: 'Test Workflow',
          failureCount: 5,
          errorPreview: 'Error',
        },
      ]

      renderWithRouter(<TopFailuresTable failures={failuresWithoutDate} />)

      expect(screen.getByText('N/A')).toBeInTheDocument()
    })
  })

  describe('Error truncation', () => {
    it('should truncate long error messages', () => {
      const longError = 'A'.repeat(150)
      const failuresWithLongError: TopFailure[] = [
        {
          workflowId: 'wf-123',
          workflowName: 'Test Workflow',
          failureCount: 5,
          errorPreview: longError,
        },
      ]

      renderWithRouter(<TopFailuresTable failures={failuresWithLongError} />)

      // Should show first 100 characters + "..."
      const truncatedError = 'A'.repeat(100) + '...'
      expect(screen.getByText(truncatedError)).toBeInTheDocument()
    })

    it('should show No error message for missing error', () => {
      const failuresWithoutError: TopFailure[] = [
        {
          workflowId: 'wf-123',
          workflowName: 'Test Workflow',
          failureCount: 5,
        },
      ]

      renderWithRouter(<TopFailuresTable failures={failuresWithoutError} />)

      expect(screen.getByText('No error message')).toBeInTheDocument()
    })
  })

  describe('Navigation', () => {
    it('should navigate to workflow detail on row click', () => {
      renderWithRouter(<TopFailuresTable failures={mockFailures} />)

      const row = screen.getByText('Data Processing Workflow').closest('tr')
      expect(row).toBeInTheDocument()

      fireEvent.click(row!)
      expect(mockNavigate).toHaveBeenCalledWith('/workflows/wf-12345678-abcd')
    })

    it('should navigate to failed executions on View Executions click', () => {
      renderWithRouter(<TopFailuresTable failures={mockFailures} />)

      const viewButtons = screen.getAllByText('View Executions')
      fireEvent.click(viewButtons[0])

      expect(mockNavigate).toHaveBeenCalledWith(
        '/executions?workflowId=wf-12345678-abcd&status=failed'
      )
    })

    it('should stop propagation when View Executions clicked', () => {
      renderWithRouter(<TopFailuresTable failures={mockFailures} />)

      const viewButtons = screen.getAllByText('View Executions')
      fireEvent.click(viewButtons[0])

      // Navigate should only be called once (not for row click as well)
      expect(mockNavigate).toHaveBeenCalledTimes(1)
    })
  })

  describe('Warning message', () => {
    it('should show warning message when failures exist', () => {
      renderWithRouter(<TopFailuresTable failures={mockFailures} />)

      expect(screen.getByText(/These workflows require attention/)).toBeInTheDocument()
    })

    it('should not show warning message when no failures', () => {
      renderWithRouter(<TopFailuresTable failures={[]} />)

      expect(screen.queryByText(/These workflows require attention/)).not.toBeInTheDocument()
    })
  })

  describe('Subtitle', () => {
    it('should display subtitle', () => {
      renderWithRouter(<TopFailuresTable failures={mockFailures} />)

      expect(screen.getByText('Workflows with most failures')).toBeInTheDocument()
    })
  })
})
