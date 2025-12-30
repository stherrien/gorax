import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import TriggerBreakdown from './TriggerBreakdown'
import { TriggerTypeBreakdown } from '../../api/metrics'

// Mock recharts
vi.mock('recharts', () => ({
  PieChart: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="pie-chart">{children}</div>
  ),
  Pie: ({
    children,
    onMouseEnter,
    onMouseLeave,
  }: {
    children: React.ReactNode
    onMouseEnter?: (data: any, index: number) => void
    onMouseLeave?: () => void
  }) => (
    <div
      data-testid="pie"
      onMouseEnter={() => onMouseEnter?.({}, 0)}
      onMouseLeave={onMouseLeave}
    >
      {children}
    </div>
  ),
  Cell: ({
    fill,
    onClick,
  }: {
    fill: string
    onClick?: () => void
  }) => (
    <div
      data-testid="cell"
      data-fill={fill}
      onClick={onClick}
    />
  ),
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="responsive-container">{children}</div>
  ),
  Tooltip: () => <div data-testid="tooltip" />,
}))

describe('TriggerBreakdown', () => {
  const mockBreakdown: TriggerTypeBreakdown[] = [
    { triggerType: 'webhook', count: 500, percentage: 50 },
    { triggerType: 'schedule', count: 300, percentage: 30 },
    { triggerType: 'manual', count: 150, percentage: 15 },
    { triggerType: 'api', count: 50, percentage: 5 },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Loading state', () => {
    it('should show loading spinner when loading', () => {
      render(<TriggerBreakdown breakdown={[]} loading={true} />)

      expect(screen.getByText('Execution by Trigger Type')).toBeInTheDocument()
      const spinner = document.querySelector('.animate-spin')
      expect(spinner).toBeInTheDocument()
    })

    it('should not show chart when loading', () => {
      render(<TriggerBreakdown breakdown={mockBreakdown} loading={true} />)

      expect(screen.queryByTestId('responsive-container')).not.toBeInTheDocument()
    })
  })

  describe('Error state', () => {
    it('should display error message', () => {
      render(<TriggerBreakdown breakdown={[]} error="Failed to load trigger breakdown" />)

      expect(screen.getByText('Failed to load trigger breakdown')).toBeInTheDocument()
    })

    it('should show title with error', () => {
      render(<TriggerBreakdown breakdown={[]} error="Error" />)

      expect(screen.getByText('Trigger Types')).toBeInTheDocument()
    })
  })

  describe('Empty state', () => {
    it('should display empty message when no data', () => {
      render(<TriggerBreakdown breakdown={[]} />)

      expect(screen.getByText('No trigger data available')).toBeInTheDocument()
    })
  })

  describe('Chart rendering', () => {
    it('should render pie chart with data', () => {
      render(<TriggerBreakdown breakdown={mockBreakdown} />)

      expect(screen.getByTestId('responsive-container')).toBeInTheDocument()
      expect(screen.getByTestId('pie-chart')).toBeInTheDocument()
    })
  })

  describe('Legend', () => {
    it('should display formatted trigger types', () => {
      render(<TriggerBreakdown breakdown={mockBreakdown} />)

      expect(screen.getByText('Webhook')).toBeInTheDocument()
      expect(screen.getByText('Schedule')).toBeInTheDocument()
      expect(screen.getByText('Manual')).toBeInTheDocument()
      expect(screen.getByText('Api')).toBeInTheDocument()
    })

    it('should display execution counts', () => {
      render(<TriggerBreakdown breakdown={mockBreakdown} />)

      expect(screen.getByText('500')).toBeInTheDocument()
      expect(screen.getByText('300')).toBeInTheDocument()
      expect(screen.getByText('150')).toBeInTheDocument()
      expect(screen.getByText('50')).toBeInTheDocument()
    })

    it('should display percentages', () => {
      render(<TriggerBreakdown breakdown={mockBreakdown} />)

      expect(screen.getByText('50.0% of total')).toBeInTheDocument()
      expect(screen.getByText('30.0% of total')).toBeInTheDocument()
      expect(screen.getByText('15.0% of total')).toBeInTheDocument()
      expect(screen.getByText('5.0% of total')).toBeInTheDocument()
    })

    it('should display executions label', () => {
      render(<TriggerBreakdown breakdown={mockBreakdown} />)

      const executionsLabels = screen.getAllByText('executions')
      expect(executionsLabels.length).toBe(4)
    })
  })

  describe('Trigger icons', () => {
    it('should display webhook icon', () => {
      render(<TriggerBreakdown breakdown={[{ triggerType: 'webhook', count: 100, percentage: 100 }]} />)

      expect(screen.getByText('🔗')).toBeInTheDocument()
    })

    it('should display schedule icon', () => {
      render(<TriggerBreakdown breakdown={[{ triggerType: 'schedule', count: 100, percentage: 100 }]} />)

      expect(screen.getByText('⏰')).toBeInTheDocument()
    })

    it('should display manual icon', () => {
      render(<TriggerBreakdown breakdown={[{ triggerType: 'manual', count: 100, percentage: 100 }]} />)

      expect(screen.getByText('👆')).toBeInTheDocument()
    })

    it('should display api icon', () => {
      render(<TriggerBreakdown breakdown={[{ triggerType: 'api', count: 100, percentage: 100 }]} />)

      expect(screen.getByText('🔌')).toBeInTheDocument()
    })

    it('should display default icon for unknown type', () => {
      render(<TriggerBreakdown breakdown={[{ triggerType: 'unknown', count: 100, percentage: 100 }]} />)

      expect(screen.getByText('📋')).toBeInTheDocument()
    })
  })

  describe('Filter click callback', () => {
    it('should call onFilterClick when legend item clicked', () => {
      const onFilterClick = vi.fn()
      render(<TriggerBreakdown breakdown={mockBreakdown} onFilterClick={onFilterClick} />)

      // Click on a legend item
      const webhookItem = screen.getByText('Webhook').closest('div[class*="flex items-center justify-between"]')
      fireEvent.click(webhookItem!)

      expect(onFilterClick).toHaveBeenCalledWith('webhook')
    })

    it('should show info message when onFilterClick provided', () => {
      render(<TriggerBreakdown breakdown={mockBreakdown} onFilterClick={vi.fn()} />)

      expect(screen.getByText(/Click on a segment to filter/)).toBeInTheDocument()
    })

    it('should not show info message without onFilterClick', () => {
      render(<TriggerBreakdown breakdown={mockBreakdown} />)

      expect(screen.queryByText(/Click on a segment to filter/)).not.toBeInTheDocument()
    })

    it('should have cursor pointer when onFilterClick provided', () => {
      const onFilterClick = vi.fn()
      render(<TriggerBreakdown breakdown={mockBreakdown} onFilterClick={onFilterClick} />)

      const legendItem = screen.getByText('Webhook').closest('div[class*="flex items-center justify-between"]')
      expect(legendItem).toHaveClass('cursor-pointer')
    })
  })

  describe('Hover interactions', () => {
    it('should handle mouse enter on legend item', () => {
      render(<TriggerBreakdown breakdown={mockBreakdown} />)

      const legendItem = screen.getByText('Webhook').closest('div[class*="flex items-center justify-between"]')
      fireEvent.mouseEnter(legendItem!)

      // Should have bg-gray-100 and scale-105 class on active
      expect(legendItem).toHaveClass('bg-gray-100', 'scale-105')
    })

    it('should handle mouse leave on legend item', () => {
      render(<TriggerBreakdown breakdown={mockBreakdown} />)

      const legendItem = screen.getByText('Webhook').closest('div[class*="flex items-center justify-between"]')
      fireEvent.mouseEnter(legendItem!)
      fireEvent.mouseLeave(legendItem!)

      // Should not have active class after mouse leave
      expect(legendItem).not.toHaveClass('scale-105')
    })
  })

  describe('Header', () => {
    it('should display title', () => {
      render(<TriggerBreakdown breakdown={mockBreakdown} />)

      expect(screen.getByText('Execution by Trigger Type')).toBeInTheDocument()
    })

    it('should display subtitle', () => {
      render(<TriggerBreakdown breakdown={mockBreakdown} />)

      expect(screen.getByText('Distribution of execution triggers')).toBeInTheDocument()
    })
  })

  describe('Trigger type formatting', () => {
    it('should format underscore-separated types', () => {
      render(<TriggerBreakdown breakdown={[{ triggerType: 'webhook_external', count: 100, percentage: 100 }]} />)

      expect(screen.getByText('Webhook External')).toBeInTheDocument()
    })

    it('should capitalize each word', () => {
      render(<TriggerBreakdown breakdown={[{ triggerType: 'manual_internal_test', count: 100, percentage: 100 }]} />)

      expect(screen.getByText('Manual Internal Test')).toBeInTheDocument()
    })
  })
})
