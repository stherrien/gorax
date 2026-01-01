import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import userEvent from '@testing-library/user-event'
import Analytics from './Analytics'
import * as analyticsHooks from '../hooks/useAnalytics'

// Mock the analytics hooks
vi.mock('../hooks/useAnalytics', () => ({
  useTenantOverview: vi.fn(),
  useExecutionTrends: vi.fn(),
  useTopWorkflows: vi.fn(),
  useErrorBreakdown: vi.fn(),
}))

const mockOverview = {
  totalExecutions: 1000,
  successfulExecutions: 850,
  failedExecutions: 100,
  cancelledExecutions: 30,
  pendingExecutions: 10,
  runningExecutions: 10,
  successRate: 0.85,
  avgDurationMs: 1500,
  activeWorkflows: 15,
  totalWorkflows: 20,
}

const mockWorkflowStats = [
  {
    workflowId: 'wf-1',
    workflowName: 'API Integration',
    executionCount: 500,
    successCount: 475,
    failureCount: 25,
    successRate: 0.95,
    avgDurationMs: 1200,
  },
  {
    workflowId: 'wf-2',
    workflowName: 'Data Pipeline',
    executionCount: 300,
    successCount: 270,
    failureCount: 30,
    successRate: 0.90,
    avgDurationMs: 2500,
  },
]

const mockTrends = {
  granularity: 'day',
  dataPoints: [
    {
      timestamp: '2025-01-20T00:00:00Z',
      executionCount: 100,
      successCount: 85,
      failureCount: 15,
      successRate: 0.85,
      avgDurationMs: 1400,
    },
    {
      timestamp: '2025-01-21T00:00:00Z',
      executionCount: 120,
      successCount: 108,
      failureCount: 12,
      successRate: 0.90,
      avgDurationMs: 1300,
    },
  ],
}

const mockErrors = {
  totalErrors: 100,
  errorsByType: [
    {
      errorMessage: 'Connection timeout',
      errorCount: 50,
      percentage: 50.0,
      workflowId: 'wf-1',
      workflowName: 'API Integration',
      lastOccurrence: '2025-01-21T12:00:00Z',
    },
    {
      errorMessage: 'Invalid response',
      errorCount: 30,
      percentage: 30.0,
      workflowId: 'wf-1',
      workflowName: 'API Integration',
      lastOccurrence: '2025-01-21T11:00:00Z',
    },
  ],
}

function renderAnalytics() {
  return render(
    <MemoryRouter>
      <Analytics />
    </MemoryRouter>
  )
}

function setupMocksWithData() {
  vi.mocked(analyticsHooks.useTenantOverview).mockReturnValue({
    data: mockOverview,
    loading: false,
    error: null,
    refetch: vi.fn(),
  })
  vi.mocked(analyticsHooks.useExecutionTrends).mockReturnValue({
    data: mockTrends,
    loading: false,
    error: null,
    refetch: vi.fn(),
  })
  vi.mocked(analyticsHooks.useTopWorkflows).mockReturnValue({
    data: { workflows: mockWorkflowStats },
    loading: false,
    error: null,
    refetch: vi.fn(),
  })
  vi.mocked(analyticsHooks.useErrorBreakdown).mockReturnValue({
    data: mockErrors,
    loading: false,
    error: null,
    refetch: vi.fn(),
  })
}

function setupMocksLoading() {
  vi.mocked(analyticsHooks.useTenantOverview).mockReturnValue({
    data: null,
    loading: true,
    error: null,
    refetch: vi.fn(),
  })
  vi.mocked(analyticsHooks.useExecutionTrends).mockReturnValue({
    data: null,
    loading: true,
    error: null,
    refetch: vi.fn(),
  })
  vi.mocked(analyticsHooks.useTopWorkflows).mockReturnValue({
    data: null,
    loading: true,
    error: null,
    refetch: vi.fn(),
  })
  vi.mocked(analyticsHooks.useErrorBreakdown).mockReturnValue({
    data: null,
    loading: true,
    error: null,
    refetch: vi.fn(),
  })
}

describe('Analytics Integration Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  describe('Overview Dashboard', () => {
    it('should load and display overview statistics', () => {
      setupMocksWithData()
      renderAnalytics()

      // Total executions
      expect(screen.getByText('1,000')).toBeInTheDocument()
      // Successful count
      expect(screen.getByText('850 successful')).toBeInTheDocument()
      // Success rate (0.85 * 100 = 85%)
      expect(screen.getByText('85%')).toBeInTheDocument()
      // Failed count
      expect(screen.getByText('100 failed')).toBeInTheDocument()
    })

    it('should display metrics cards', () => {
      setupMocksWithData()
      renderAnalytics()

      expect(screen.getByText('Total Executions')).toBeInTheDocument()
      expect(screen.getByText('Success Rate')).toBeInTheDocument()
      expect(screen.getByText('Avg Duration')).toBeInTheDocument()
      expect(screen.getByText('Active Workflows')).toBeInTheDocument()
    })

    it('should display active workflows count', () => {
      setupMocksWithData()
      renderAnalytics()

      // Active workflows
      expect(screen.getByText('15')).toBeInTheDocument()
      // Total workflows
      expect(screen.getByText('of 20 total')).toBeInTheDocument()
    })

    it('should display average duration in seconds', () => {
      setupMocksWithData()
      renderAnalytics()

      // 1500ms = 1.50s
      expect(screen.getByText('1.50s')).toBeInTheDocument()
      expect(screen.getByText('Per execution')).toBeInTheDocument()
    })
  })

  describe('Date Range Controls', () => {
    it('should render date range buttons', () => {
      setupMocksWithData()
      renderAnalytics()

      expect(screen.getByText('Last 24h')).toBeInTheDocument()
      expect(screen.getByText('Last Week')).toBeInTheDocument()
      expect(screen.getByText('Last Month')).toBeInTheDocument()
      expect(screen.getByText('Last Year')).toBeInTheDocument()
    })

    it('should call hooks when date range button is clicked', async () => {
      setupMocksWithData()
      const user = userEvent.setup()
      renderAnalytics()

      // Click Last Week button
      await user.click(screen.getByText('Last Week'))

      // Hooks should be called with updated params (component re-renders)
      expect(analyticsHooks.useTenantOverview).toHaveBeenCalled()
    })

    it('should render granularity selector', () => {
      setupMocksWithData()
      renderAnalytics()

      const select = screen.getByRole('combobox')
      expect(select).toBeInTheDocument()
      expect(screen.getByText('Hourly')).toBeInTheDocument()
      expect(screen.getByText('Daily')).toBeInTheDocument()
      expect(screen.getByText('Weekly')).toBeInTheDocument()
      expect(screen.getByText('Monthly')).toBeInTheDocument()
    })

    it('should update granularity when changed', async () => {
      setupMocksWithData()
      const user = userEvent.setup()
      renderAnalytics()

      const select = screen.getByRole('combobox')
      await user.selectOptions(select, 'hour')

      // Trends hook should be called with hourly granularity
      expect(analyticsHooks.useExecutionTrends).toHaveBeenCalled()
    })
  })

  describe('Top Workflows Section', () => {
    it('should display top workflows', () => {
      setupMocksWithData()
      renderAnalytics()

      // Top Workflows section header
      const topWorkflowsHeaders = screen.getAllByText('Top Workflows')
      expect(topWorkflowsHeaders.length).toBeGreaterThanOrEqual(1)
      // Workflow names appear in multiple sections (Top Workflows and Error Breakdown)
      const apiIntegrationElements = screen.getAllByText('API Integration')
      expect(apiIntegrationElements.length).toBeGreaterThanOrEqual(1)
      expect(screen.getByText('Data Pipeline')).toBeInTheDocument()
    })

    it('should show workflow execution counts and success rates', () => {
      setupMocksWithData()
      renderAnalytics()

      // API Integration: 500 executions, 95% success
      expect(screen.getByText(/500 executions/)).toBeInTheDocument()
      expect(screen.getByText(/95% success rate/)).toBeInTheDocument()

      // Data Pipeline: 300 executions, 90% success
      expect(screen.getByText(/300 executions/)).toBeInTheDocument()
      expect(screen.getByText(/90% success rate/)).toBeInTheDocument()
    })

    it('should display workflow average durations', () => {
      setupMocksWithData()
      renderAnalytics()

      // API Integration: 1200ms = 1.20s
      expect(screen.getByText('1.20s')).toBeInTheDocument()
      // Data Pipeline: 2500ms = 2.50s
      expect(screen.getByText('2.50s')).toBeInTheDocument()
    })

    it('should show empty state when no workflows', () => {
      vi.mocked(analyticsHooks.useTenantOverview).mockReturnValue({
        data: mockOverview,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useExecutionTrends).mockReturnValue({
        data: mockTrends,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useTopWorkflows).mockReturnValue({
        data: { workflows: [] },
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useErrorBreakdown).mockReturnValue({
        data: mockErrors,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderAnalytics()

      expect(screen.getByText('No workflows found')).toBeInTheDocument()
    })
  })

  describe('Execution Trends Section', () => {
    it('should display execution trends', () => {
      setupMocksWithData()
      renderAnalytics()

      expect(screen.getByText('Execution Trends')).toBeInTheDocument()
      expect(screen.getByText('100 executions')).toBeInTheDocument()
      expect(screen.getByText('120 executions')).toBeInTheDocument()
    })

    it('should display success rates for trends', () => {
      setupMocksWithData()
      renderAnalytics()

      // 85% and 90% success rates
      expect(screen.getByText('85% success')).toBeInTheDocument()
      expect(screen.getByText('90% success')).toBeInTheDocument()
    })

    it('should show empty state when no trend data', () => {
      vi.mocked(analyticsHooks.useTenantOverview).mockReturnValue({
        data: mockOverview,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useExecutionTrends).mockReturnValue({
        data: { granularity: 'day', dataPoints: [] },
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useTopWorkflows).mockReturnValue({
        data: { workflows: mockWorkflowStats },
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useErrorBreakdown).mockReturnValue({
        data: mockErrors,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderAnalytics()

      expect(screen.getByText('No data available')).toBeInTheDocument()
    })
  })

  describe('Error Breakdown Section', () => {
    it('should display error breakdown', () => {
      setupMocksWithData()
      renderAnalytics()

      expect(screen.getByText('Error Breakdown')).toBeInTheDocument()
      expect(screen.getByText('Connection timeout')).toBeInTheDocument()
      expect(screen.getByText('Invalid response')).toBeInTheDocument()
    })

    it('should show error counts', () => {
      setupMocksWithData()
      renderAnalytics()

      // Error counts
      expect(screen.getByText('50')).toBeInTheDocument()
      expect(screen.getByText('30')).toBeInTheDocument()
    })

    it('should show error percentages', () => {
      setupMocksWithData()
      renderAnalytics()

      expect(screen.getByText('50.0%')).toBeInTheDocument()
      expect(screen.getByText('30.0%')).toBeInTheDocument()
    })

    it('should display associated workflow names', () => {
      setupMocksWithData()
      renderAnalytics()

      // Both errors are from API Integration
      const workflowNames = screen.getAllByText('API Integration')
      expect(workflowNames.length).toBeGreaterThanOrEqual(2)
    })

    it('should show empty state when no errors', () => {
      vi.mocked(analyticsHooks.useTenantOverview).mockReturnValue({
        data: mockOverview,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useExecutionTrends).mockReturnValue({
        data: mockTrends,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useTopWorkflows).mockReturnValue({
        data: { workflows: mockWorkflowStats },
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useErrorBreakdown).mockReturnValue({
        data: { totalErrors: 0, errorsByType: [] },
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderAnalytics()

      expect(screen.getByText('No errors found - great job!')).toBeInTheDocument()
    })
  })

  describe('Loading States', () => {
    it('should show loading state for overview', () => {
      setupMocksLoading()
      renderAnalytics()

      expect(screen.getByText('Loading overview...')).toBeInTheDocument()
    })

    it('should show loading state for trends', () => {
      vi.mocked(analyticsHooks.useTenantOverview).mockReturnValue({
        data: mockOverview,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useExecutionTrends).mockReturnValue({
        data: null,
        loading: true,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useTopWorkflows).mockReturnValue({
        data: { workflows: mockWorkflowStats },
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useErrorBreakdown).mockReturnValue({
        data: mockErrors,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderAnalytics()

      expect(screen.getByText('Loading trends...')).toBeInTheDocument()
    })

    it('should show loading state for workflows', () => {
      vi.mocked(analyticsHooks.useTenantOverview).mockReturnValue({
        data: mockOverview,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useExecutionTrends).mockReturnValue({
        data: mockTrends,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useTopWorkflows).mockReturnValue({
        data: null,
        loading: true,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useErrorBreakdown).mockReturnValue({
        data: mockErrors,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderAnalytics()

      expect(screen.getByText('Loading workflows...')).toBeInTheDocument()
    })

    it('should show loading state for errors', () => {
      vi.mocked(analyticsHooks.useTenantOverview).mockReturnValue({
        data: mockOverview,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useExecutionTrends).mockReturnValue({
        data: mockTrends,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useTopWorkflows).mockReturnValue({
        data: { workflows: mockWorkflowStats },
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useErrorBreakdown).mockReturnValue({
        data: null,
        loading: true,
        error: null,
        refetch: vi.fn(),
      })

      renderAnalytics()

      expect(screen.getByText('Loading errors...')).toBeInTheDocument()
    })
  })

  describe('Empty Data States', () => {
    it('should not render overview cards when data is null', () => {
      vi.mocked(analyticsHooks.useTenantOverview).mockReturnValue({
        data: null,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useExecutionTrends).mockReturnValue({
        data: mockTrends,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useTopWorkflows).mockReturnValue({
        data: { workflows: mockWorkflowStats },
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useErrorBreakdown).mockReturnValue({
        data: mockErrors,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderAnalytics()

      // Overview specific text should not be present
      expect(screen.queryByText('1,000')).not.toBeInTheDocument()
    })

    it('should not render trends when data is null', () => {
      vi.mocked(analyticsHooks.useTenantOverview).mockReturnValue({
        data: mockOverview,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useExecutionTrends).mockReturnValue({
        data: null,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useTopWorkflows).mockReturnValue({
        data: { workflows: mockWorkflowStats },
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useErrorBreakdown).mockReturnValue({
        data: mockErrors,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderAnalytics()

      // Execution trends section header should still be there
      expect(screen.getByText('Execution Trends')).toBeInTheDocument()
      // But no data points
      expect(screen.queryByText('100 executions')).not.toBeInTheDocument()
    })

    it('should not render workflows when data is null', () => {
      vi.mocked(analyticsHooks.useTenantOverview).mockReturnValue({
        data: mockOverview,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useExecutionTrends).mockReturnValue({
        data: mockTrends,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useTopWorkflows).mockReturnValue({
        data: null,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useErrorBreakdown).mockReturnValue({
        data: mockErrors,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderAnalytics()

      // Top workflows header should still be there
      const topWorkflowsHeaders = screen.getAllByText('Top Workflows')
      expect(topWorkflowsHeaders.length).toBeGreaterThanOrEqual(1)
      // Data Pipeline only appears in Top Workflows, not in Error Breakdown
      expect(screen.queryByText('Data Pipeline')).not.toBeInTheDocument()
    })

    it('should not render error breakdown when data is null', () => {
      vi.mocked(analyticsHooks.useTenantOverview).mockReturnValue({
        data: mockOverview,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useExecutionTrends).mockReturnValue({
        data: mockTrends,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useTopWorkflows).mockReturnValue({
        data: { workflows: mockWorkflowStats },
        loading: false,
        error: null,
        refetch: vi.fn(),
      })
      vi.mocked(analyticsHooks.useErrorBreakdown).mockReturnValue({
        data: null,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderAnalytics()

      // Error breakdown header should still be there
      expect(screen.getByText('Error Breakdown')).toBeInTheDocument()
      // But no error data
      expect(screen.queryByText('Connection timeout')).not.toBeInTheDocument()
    })
  })

  describe('Page Structure', () => {
    it('should render the page title', () => {
      setupMocksWithData()
      renderAnalytics()

      expect(screen.getByText('Analytics Dashboard')).toBeInTheDocument()
    })

    it('should render the page description', () => {
      setupMocksWithData()
      renderAnalytics()

      expect(
        screen.getByText('Monitor workflow performance and execution trends')
      ).toBeInTheDocument()
    })
  })
})
