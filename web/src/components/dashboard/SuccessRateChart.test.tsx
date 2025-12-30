import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import SuccessRateChart from './SuccessRateChart'
import { ExecutionTrend } from '../../api/metrics'

// Mock recharts
vi.mock('recharts', () => ({
  LineChart: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="line-chart">{children}</div>
  ),
  Line: () => <div data-testid="line" />,
  XAxis: () => <div data-testid="x-axis" />,
  YAxis: () => <div data-testid="y-axis" />,
  CartesianGrid: () => <div data-testid="cartesian-grid" />,
  Tooltip: () => <div data-testid="tooltip" />,
  Legend: () => <div data-testid="legend" />,
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="responsive-container">{children}</div>
  ),
  ReferenceLine: ({ y, label }: { y: number; label?: any }) => (
    <div data-testid="reference-line" data-y={y}>
      {label?.value}
    </div>
  ),
}))

describe('SuccessRateChart', () => {
  const mockTrends: ExecutionTrend[] = [
    { date: '2025-01-10', count: 100, success: 95, failed: 5 },
    { date: '2025-01-11', count: 120, success: 110, failed: 10 },
    { date: '2025-01-12', count: 80, success: 75, failed: 5 },
    { date: '2025-01-13', count: 100, success: 92, failed: 8 },
    { date: '2025-01-14', count: 110, success: 100, failed: 10 },
    { date: '2025-01-15', count: 90, success: 85, failed: 5 },
    { date: '2025-01-16', count: 100, success: 90, failed: 10 },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Loading state', () => {
    it('should show loading spinner when loading', () => {
      render(<SuccessRateChart trends={[]} loading={true} />)

      expect(screen.getByText('Success Rate')).toBeInTheDocument()
      const spinner = document.querySelector('.animate-spin')
      expect(spinner).toBeInTheDocument()
    })

    it('should not show chart when loading', () => {
      render(<SuccessRateChart trends={mockTrends} loading={true} />)

      expect(screen.queryByTestId('responsive-container')).not.toBeInTheDocument()
    })
  })

  describe('Error state', () => {
    it('should display error message', () => {
      render(<SuccessRateChart trends={[]} error="Failed to load success rate" />)

      expect(screen.getByText('Failed to load success rate')).toBeInTheDocument()
    })

    it('should show title with error', () => {
      render(<SuccessRateChart trends={[]} error="Error" />)

      expect(screen.getByText('Success Rate')).toBeInTheDocument()
    })
  })

  describe('Empty state', () => {
    it('should display empty message when no data', () => {
      render(<SuccessRateChart trends={[]} />)

      expect(screen.getByText('No data available')).toBeInTheDocument()
    })
  })

  describe('Chart rendering', () => {
    it('should render chart with data', () => {
      render(<SuccessRateChart trends={mockTrends} />)

      expect(screen.getByTestId('responsive-container')).toBeInTheDocument()
      expect(screen.getByTestId('line-chart')).toBeInTheDocument()
    })
  })

  describe('Overall rate calculation', () => {
    it('should calculate and display overall success rate', () => {
      render(<SuccessRateChart trends={mockTrends} />)

      // Total: 700 count, 647 success = 92.43%
      expect(screen.getByText('92.4%')).toBeInTheDocument()
    })

    it('should handle zero count gracefully', () => {
      const emptyTrends: ExecutionTrend[] = [
        { date: '2025-01-15', count: 0, success: 0, failed: 0 },
      ]

      render(<SuccessRateChart trends={emptyTrends} />)

      expect(screen.getByText('0.0%')).toBeInTheDocument()
    })
  })

  describe('Status color', () => {
    it('should show green for high success rate (>= 90%)', () => {
      render(<SuccessRateChart trends={mockTrends} />)

      const rateElement = screen.getByText('92.4%')
      expect(rateElement).toHaveClass('text-green-600')
    })

    it('should show yellow for medium success rate (80-90%)', () => {
      const mediumTrends: ExecutionTrend[] = [
        { date: '2025-01-15', count: 100, success: 85, failed: 15 },
      ]

      render(<SuccessRateChart trends={mediumTrends} />)

      const rateElement = screen.getByText('85.0%')
      expect(rateElement).toHaveClass('text-yellow-600')
    })

    it('should show red for low success rate (< 80%)', () => {
      const lowTrends: ExecutionTrend[] = [
        { date: '2025-01-15', count: 100, success: 70, failed: 30 },
      ]

      render(<SuccessRateChart trends={lowTrends} />)

      const rateElement = screen.getByText('70.0%')
      expect(rateElement).toHaveClass('text-red-600')
    })
  })

  describe('Trend determination', () => {
    it('should show stable trend indicator', () => {
      const stableTrends: ExecutionTrend[] = [
        { date: '2025-01-10', count: 100, success: 90, failed: 10 },
        { date: '2025-01-11', count: 100, success: 90, failed: 10 },
        { date: '2025-01-12', count: 100, success: 90, failed: 10 },
        { date: '2025-01-13', count: 100, success: 90, failed: 10 },
      ]

      render(<SuccessRateChart trends={stableTrends} />)

      expect(screen.getByText('→')).toBeInTheDocument()
      expect(screen.getByText('stable')).toBeInTheDocument()
    })

    it('should show improving trend indicator', () => {
      const improvingTrends: ExecutionTrend[] = [
        { date: '2025-01-10', count: 100, success: 70, failed: 30 },
        { date: '2025-01-11', count: 100, success: 75, failed: 25 },
        { date: '2025-01-12', count: 100, success: 80, failed: 20 },
        { date: '2025-01-13', count: 100, success: 85, failed: 15 },
        { date: '2025-01-14', count: 100, success: 90, failed: 10 },
        { date: '2025-01-15', count: 100, success: 95, failed: 5 },
        { date: '2025-01-16', count: 100, success: 98, failed: 2 },
      ]

      render(<SuccessRateChart trends={improvingTrends} />)

      expect(screen.getByText('↗')).toBeInTheDocument()
      expect(screen.getByText('improving')).toBeInTheDocument()
    })

    it('should show declining trend indicator', () => {
      const decliningTrends: ExecutionTrend[] = [
        { date: '2025-01-10', count: 100, success: 98, failed: 2 },
        { date: '2025-01-11', count: 100, success: 95, failed: 5 },
        { date: '2025-01-12', count: 100, success: 90, failed: 10 },
        { date: '2025-01-13', count: 100, success: 85, failed: 15 },
        { date: '2025-01-14', count: 100, success: 80, failed: 20 },
        { date: '2025-01-15', count: 100, success: 75, failed: 25 },
        { date: '2025-01-16', count: 100, success: 70, failed: 30 },
      ]

      render(<SuccessRateChart trends={decliningTrends} />)

      expect(screen.getByText('↘')).toBeInTheDocument()
      expect(screen.getByText('declining')).toBeInTheDocument()
    })

    it('should show stable for single data point', () => {
      const singleTrend: ExecutionTrend[] = [
        { date: '2025-01-15', count: 100, success: 90, failed: 10 },
      ]

      render(<SuccessRateChart trends={singleTrend} />)

      expect(screen.getByText('→')).toBeInTheDocument()
      expect(screen.getByText('stable')).toBeInTheDocument()
    })
  })

  describe('Target rate', () => {
    it('should show default target rate of 95%', () => {
      render(<SuccessRateChart trends={mockTrends} />)

      expect(screen.getByText('Target 95%')).toBeInTheDocument()
    })

    it('should show custom target rate', () => {
      render(<SuccessRateChart trends={mockTrends} targetRate={99} />)

      expect(screen.getByText('Target 99%')).toBeInTheDocument()
    })
  })

  describe('Legend categories', () => {
    it('should display rate categories', () => {
      render(<SuccessRateChart trends={mockTrends} />)

      expect(screen.getByText('Excellent')).toBeInTheDocument()
      expect(screen.getByText('> 90%')).toBeInTheDocument()
      expect(screen.getByText('Good')).toBeInTheDocument()
      expect(screen.getByText('80-90%')).toBeInTheDocument()
      expect(screen.getByText('Needs Attention')).toBeInTheDocument()
      expect(screen.getByText('< 80%')).toBeInTheDocument()
    })
  })

  describe('Header', () => {
    it('should display Success Rate title', () => {
      render(<SuccessRateChart trends={mockTrends} />)

      expect(screen.getByText('Success Rate')).toBeInTheDocument()
    })
  })
})
