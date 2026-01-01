import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import ExecutionTrendChart from './ExecutionTrendChart'
import { ExecutionTrend } from '../../api/metrics'

// Mock recharts to avoid rendering issues in tests
vi.mock('recharts', () => ({
  LineChart: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="line-chart">{children}</div>
  ),
  BarChart: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="bar-chart">{children}</div>
  ),
  Line: () => <div data-testid="line" />,
  Bar: ({ dataKey, stackId }: { dataKey: string; stackId?: string }) => (
    <div data-testid={`bar-${dataKey}`} data-stack-id={stackId} />
  ),
  XAxis: () => <div data-testid="x-axis" />,
  YAxis: () => <div data-testid="y-axis" />,
  CartesianGrid: () => <div data-testid="cartesian-grid" />,
  Tooltip: () => <div data-testid="tooltip" />,
  Legend: () => <div data-testid="legend" />,
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="responsive-container">{children}</div>
  ),
}))

describe('ExecutionTrendChart', () => {
  const mockTrends: ExecutionTrend[] = [
    { date: '2025-01-15', count: 100, success: 90, failed: 10 },
    { date: '2025-01-16', count: 150, success: 140, failed: 10 },
    { date: '2025-01-17', count: 120, success: 100, failed: 20 },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Loading state', () => {
    it('should show loading spinner when loading', () => {
      render(<ExecutionTrendChart trends={[]} loading={true} />)

      expect(screen.getByText('Execution Trends')).toBeInTheDocument()
      const spinner = document.querySelector('.animate-spin')
      expect(spinner).toBeInTheDocument()
    })

    it('should not show chart when loading', () => {
      render(<ExecutionTrendChart trends={mockTrends} loading={true} />)

      expect(screen.queryByTestId('responsive-container')).not.toBeInTheDocument()
    })
  })

  describe('Error state', () => {
    it('should display error message', () => {
      render(<ExecutionTrendChart trends={[]} error="Failed to load trends" />)

      expect(screen.getByText('Failed to load trends')).toBeInTheDocument()
    })

    it('should show Try Again button when onRefresh provided', () => {
      const onRefresh = vi.fn()
      render(
        <ExecutionTrendChart
          trends={[]}
          error="Failed to load"
          onRefresh={onRefresh}
        />
      )

      const tryAgainButton = screen.getByText('Try Again')
      expect(tryAgainButton).toBeInTheDocument()

      fireEvent.click(tryAgainButton)
      expect(onRefresh).toHaveBeenCalledTimes(1)
    })

    it('should not show Try Again button without onRefresh', () => {
      render(<ExecutionTrendChart trends={[]} error="Failed to load" />)

      expect(screen.queryByText('Try Again')).not.toBeInTheDocument()
    })
  })

  describe('Empty state', () => {
    it('should display empty message when no data', () => {
      render(<ExecutionTrendChart trends={[]} />)

      expect(screen.getByText('No execution data available')).toBeInTheDocument()
    })
  })

  describe('Chart rendering', () => {
    it('should render line chart by default', () => {
      render(<ExecutionTrendChart trends={mockTrends} />)

      expect(screen.getByTestId('responsive-container')).toBeInTheDocument()
      expect(screen.getByTestId('line-chart')).toBeInTheDocument()
    })

    it('should show chart type toggle buttons', () => {
      render(<ExecutionTrendChart trends={mockTrends} />)

      expect(screen.getByText('Line')).toBeInTheDocument()
      expect(screen.getByText('Bar')).toBeInTheDocument()
      expect(screen.getByText('Stacked')).toBeInTheDocument()
    })
  })

  describe('Chart type toggle', () => {
    it('should switch to bar chart when Bar button clicked', () => {
      render(<ExecutionTrendChart trends={mockTrends} />)

      fireEvent.click(screen.getByText('Bar'))

      expect(screen.getByTestId('bar-chart')).toBeInTheDocument()
      expect(screen.queryByTestId('line-chart')).not.toBeInTheDocument()
    })

    it('should switch to stacked chart when Stacked button clicked', () => {
      render(<ExecutionTrendChart trends={mockTrends} />)

      fireEvent.click(screen.getByText('Stacked'))

      expect(screen.getByTestId('bar-chart')).toBeInTheDocument()
      expect(screen.getByTestId('bar-success')).toHaveAttribute('data-stack-id', 'a')
      expect(screen.getByTestId('bar-failed')).toHaveAttribute('data-stack-id', 'a')
    })

    it('should switch back to line chart', () => {
      render(<ExecutionTrendChart trends={mockTrends} />)

      fireEvent.click(screen.getByText('Bar'))
      expect(screen.getByTestId('bar-chart')).toBeInTheDocument()

      fireEvent.click(screen.getByText('Line'))
      expect(screen.getByTestId('line-chart')).toBeInTheDocument()
    })

    it('should highlight active chart type button', () => {
      render(<ExecutionTrendChart trends={mockTrends} />)

      const lineButton = screen.getByText('Line')
      expect(lineButton).toHaveClass('bg-white')

      fireEvent.click(screen.getByText('Bar'))
      expect(screen.getByText('Bar')).toHaveClass('bg-white')
      expect(lineButton).not.toHaveClass('bg-white')
    })
  })

  describe('Time range selector', () => {
    it('should render time range dropdown', () => {
      render(<ExecutionTrendChart trends={mockTrends} />)

      const dropdown = screen.getByDisplayValue('Last 7 days')
      expect(dropdown).toBeInTheDocument()
    })

    it('should have all time range options', () => {
      render(<ExecutionTrendChart trends={mockTrends} />)

      expect(screen.getByText('Last 7 days')).toBeInTheDocument()
      expect(screen.getByText('Last 30 days')).toBeInTheDocument()
      expect(screen.getByText('Last 90 days')).toBeInTheDocument()
    })

    it('should change time range when option selected', () => {
      render(<ExecutionTrendChart trends={mockTrends} />)

      const dropdown = screen.getByDisplayValue('Last 7 days')
      fireEvent.change(dropdown, { target: { value: '30d' } })

      expect(screen.getByDisplayValue('Last 30 days')).toBeInTheDocument()
    })
  })

  describe('Group by selector', () => {
    it('should render group by dropdown', () => {
      render(<ExecutionTrendChart trends={mockTrends} />)

      const dropdown = screen.getByDisplayValue('Daily')
      expect(dropdown).toBeInTheDocument()
    })

    it('should have hourly and daily options', () => {
      render(<ExecutionTrendChart trends={mockTrends} />)

      expect(screen.getByText('Hourly')).toBeInTheDocument()
      expect(screen.getByText('Daily')).toBeInTheDocument()
    })

    it('should change group by when option selected', () => {
      render(<ExecutionTrendChart trends={mockTrends} />)

      const dropdown = screen.getByDisplayValue('Daily')
      fireEvent.change(dropdown, { target: { value: 'hour' } })

      expect(screen.getByDisplayValue('Hourly')).toBeInTheDocument()
    })
  })

  describe('Grouped bar chart', () => {
    it('should render grouped bars without stack ID', () => {
      render(<ExecutionTrendChart trends={mockTrends} />)

      fireEvent.click(screen.getByText('Bar'))

      expect(screen.getByTestId('bar-success')).not.toHaveAttribute('data-stack-id', 'a')
      expect(screen.getByTestId('bar-failed')).not.toHaveAttribute('data-stack-id', 'a')
    })
  })

  describe('Header', () => {
    it('should display Execution Trends title', () => {
      render(<ExecutionTrendChart trends={mockTrends} />)

      expect(screen.getByText('Execution Trends')).toBeInTheDocument()
    })
  })
})
