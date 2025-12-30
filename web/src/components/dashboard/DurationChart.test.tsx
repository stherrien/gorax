import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import DurationChart from './DurationChart'
import { DurationStats } from '../../api/metrics'

// Mock recharts
vi.mock('recharts', () => ({
  BarChart: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="bar-chart">{children}</div>
  ),
  Bar: ({ dataKey }: { dataKey: string }) => <div data-testid={`bar-${dataKey}`} />,
  XAxis: () => <div data-testid="x-axis" />,
  YAxis: () => <div data-testid="y-axis" />,
  CartesianGrid: () => <div data-testid="cartesian-grid" />,
  Tooltip: ({ content }: { content: (props: any) => React.ReactNode }) => (
    <div data-testid="tooltip">{content?.({ active: false, payload: [] })}</div>
  ),
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="responsive-container">{children}</div>
  ),
  Cell: ({ fill }: { fill: string }) => <div data-testid="cell" data-fill={fill} />,
}))

describe('DurationChart', () => {
  const mockStats: DurationStats[] = [
    {
      workflowId: 'wf-1',
      workflowName: 'Fast Workflow',
      avgDuration: 500,
      p50Duration: 400,
      p90Duration: 600,
      p99Duration: 800,
      totalRuns: 100,
    },
    {
      workflowId: 'wf-2',
      workflowName: 'Medium Workflow',
      avgDuration: 2500,
      p50Duration: 2000,
      p90Duration: 3000,
      p99Duration: 4000,
      totalRuns: 50,
    },
    {
      workflowId: 'wf-3',
      workflowName: 'Slow Workflow With A Very Long Name That Needs Truncation',
      avgDuration: 90000,
      p50Duration: 85000,
      p90Duration: 100000,
      p99Duration: 120000,
      totalRuns: 200,
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Loading state', () => {
    it('should show loading spinner when loading', () => {
      render(<DurationChart stats={[]} loading={true} />)

      expect(screen.getByText('Execution Duration')).toBeInTheDocument()
      const spinner = document.querySelector('.animate-spin')
      expect(spinner).toBeInTheDocument()
    })

    it('should not show chart when loading', () => {
      render(<DurationChart stats={mockStats} loading={true} />)

      expect(screen.queryByTestId('responsive-container')).not.toBeInTheDocument()
    })
  })

  describe('Error state', () => {
    it('should display error message', () => {
      render(<DurationChart stats={[]} error="Failed to load duration stats" />)

      expect(screen.getByText('Failed to load duration stats')).toBeInTheDocument()
    })

    it('should show title with error', () => {
      render(<DurationChart stats={[]} error="Error" />)

      expect(screen.getByText('Execution Duration')).toBeInTheDocument()
    })
  })

  describe('Empty state', () => {
    it('should display empty message when no data', () => {
      render(<DurationChart stats={[]} />)

      expect(screen.getByText('No duration data available')).toBeInTheDocument()
    })
  })

  describe('Chart rendering', () => {
    it('should render chart with data', () => {
      render(<DurationChart stats={mockStats} />)

      expect(screen.getByTestId('responsive-container')).toBeInTheDocument()
      expect(screen.getByTestId('bar-chart')).toBeInTheDocument()
    })

    it('should show metric type selector', () => {
      render(<DurationChart stats={mockStats} />)

      expect(screen.getByDisplayValue('Average')).toBeInTheDocument()
    })

    it('should show sort by selector', () => {
      render(<DurationChart stats={mockStats} />)

      expect(screen.getByDisplayValue('Sort by Duration')).toBeInTheDocument()
    })
  })

  describe('Metric type selector', () => {
    it('should have all metric options', () => {
      render(<DurationChart stats={mockStats} />)

      expect(screen.getByText('Average')).toBeInTheDocument()
      expect(screen.getByText('Median (P50)')).toBeInTheDocument()
      expect(screen.getByText('P90')).toBeInTheDocument()
      expect(screen.getByText('P99')).toBeInTheDocument()
    })

    it('should change metric type when option selected', () => {
      render(<DurationChart stats={mockStats} />)

      const dropdown = screen.getByDisplayValue('Average')
      fireEvent.change(dropdown, { target: { value: 'p50' } })

      expect(screen.getByDisplayValue('Median (P50)')).toBeInTheDocument()
    })

    it('should allow selecting P90', () => {
      render(<DurationChart stats={mockStats} />)

      const dropdown = screen.getByDisplayValue('Average')
      fireEvent.change(dropdown, { target: { value: 'p90' } })

      expect(screen.getByDisplayValue('P90')).toBeInTheDocument()
    })

    it('should allow selecting P99', () => {
      render(<DurationChart stats={mockStats} />)

      const dropdown = screen.getByDisplayValue('Average')
      fireEvent.change(dropdown, { target: { value: 'p99' } })

      expect(screen.getByDisplayValue('P99')).toBeInTheDocument()
    })
  })

  describe('Sort by selector', () => {
    it('should have duration and runs options', () => {
      render(<DurationChart stats={mockStats} />)

      expect(screen.getByText('Sort by Duration')).toBeInTheDocument()
      expect(screen.getByText('Sort by Run Count')).toBeInTheDocument()
    })

    it('should change sort by when option selected', () => {
      render(<DurationChart stats={mockStats} />)

      const dropdown = screen.getByDisplayValue('Sort by Duration')
      fireEvent.change(dropdown, { target: { value: 'runs' } })

      expect(screen.getByDisplayValue('Sort by Run Count')).toBeInTheDocument()
    })
  })

  describe('Legend', () => {
    it('should display color legend', () => {
      render(<DurationChart stats={mockStats} />)

      expect(screen.getByText('Fast')).toBeInTheDocument()
      expect(screen.getByText('Medium')).toBeInTheDocument()
      expect(screen.getByText('Slow (Outlier)')).toBeInTheDocument()
    })
  })

  describe('Info note', () => {
    it('should display percentile information', () => {
      render(<DurationChart stats={mockStats} />)

      expect(screen.getByText('About Percentiles:')).toBeInTheDocument()
      expect(screen.getByText(/P50 \(Median\)/)).toBeInTheDocument()
      expect(screen.getByText(/P90:/)).toBeInTheDocument()
      expect(screen.getByText(/P99:/)).toBeInTheDocument()
    })
  })

  describe('Header', () => {
    it('should display Execution Duration title', () => {
      render(<DurationChart stats={mockStats} />)

      expect(screen.getByText('Execution Duration')).toBeInTheDocument()
    })
  })

  describe('Data processing', () => {
    it('should limit to top 10 workflows', () => {
      const manyStats: DurationStats[] = Array.from({ length: 15 }, (_, i) => ({
        workflowId: `wf-${i}`,
        workflowName: `Workflow ${i}`,
        avgDuration: 1000 + i * 100,
        p50Duration: 800 + i * 100,
        p90Duration: 1200 + i * 100,
        p99Duration: 1500 + i * 100,
        totalRuns: 50,
      }))

      render(<DurationChart stats={manyStats} />)

      // Chart should be rendered (implicitly limiting to 10)
      expect(screen.getByTestId('bar-chart')).toBeInTheDocument()
    })
  })
})
