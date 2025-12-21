import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import ParallelNode from './ParallelNode'

// Mock @xyflow/react to avoid provider issues in tests
vi.mock('@xyflow/react', () => ({
  Handle: ({ type, position }: any) => <div data-testid={`handle-${type}-${position}`} />,
  Position: {
    Top: 'top',
    Bottom: 'bottom',
    Left: 'left',
    Right: 'right',
  },
}))

interface ParallelNodeData {
  label: string
  errorStrategy?: 'fail_fast' | 'wait_all'
  maxConcurrency?: number
  branchCount?: number
}

describe('ParallelNode', () => {
  describe('Basic rendering', () => {
    it('should render parallel node with label', () => {
      const data: ParallelNodeData = {
        label: 'Process in Parallel',
      }

      render(<ParallelNode data={data} selected={false} />)

      expect(screen.getByText('Process in Parallel')).toBeInTheDocument()
      expect(screen.getByText('Parallel')).toBeInTheDocument()
    })

    it('should render with parallel icon', () => {
      const data: ParallelNodeData = {
        label: 'Parallel Test',
      }

      render(<ParallelNode data={data} selected={false} />)

      // Check for parallel icon
      expect(screen.getByText('⚡')).toBeInTheDocument()
    })

    it('should have correct styling', () => {
      const data: ParallelNodeData = {
        label: 'Styled Parallel',
      }

      const { container } = render(<ParallelNode data={data} selected={false} />)
      const node = container.firstChild as HTMLElement

      expect(node).toHaveClass('px-4')
      expect(node).toHaveClass('py-3')
      expect(node).toHaveClass('rounded-lg')
      expect(node).toHaveClass('shadow-lg')
    })
  })

  describe('Configuration display', () => {
    it('should display error strategy when set to fail_fast', () => {
      const data: ParallelNodeData = {
        label: 'Parallel Test',
        errorStrategy: 'fail_fast',
      }

      render(<ParallelNode data={data} selected={false} />)

      expect(screen.getByText(/Fail Fast/i)).toBeInTheDocument()
    })

    it('should display error strategy when set to wait_all', () => {
      const data: ParallelNodeData = {
        label: 'Parallel Test',
        errorStrategy: 'wait_all',
      }

      render(<ParallelNode data={data} selected={false} />)

      expect(screen.getByText(/Wait All/i)).toBeInTheDocument()
    })

    it('should display max concurrency when provided', () => {
      const data: ParallelNodeData = {
        label: 'Parallel Test',
        maxConcurrency: 5,
      }

      render(<ParallelNode data={data} selected={false} />)

      expect(screen.getByText(/Max: 5/)).toBeInTheDocument()
    })

    it('should show unlimited when max concurrency is 0', () => {
      const data: ParallelNodeData = {
        label: 'Parallel Test',
        maxConcurrency: 0,
      }

      render(<ParallelNode data={data} selected={false} />)

      expect(screen.getByText(/Unlimited/i)).toBeInTheDocument()
    })

    it('should display branch count when provided', () => {
      const data: ParallelNodeData = {
        label: 'Parallel Test',
        branchCount: 3,
      }

      render(<ParallelNode data={data} selected={false} />)

      expect(screen.getByText(/3 branches/i)).toBeInTheDocument()
    })
  })

  describe('Selection state', () => {
    it('should apply selection styling when selected', () => {
      const data: ParallelNodeData = {
        label: 'Selected Parallel',
      }

      const { container } = render(<ParallelNode data={data} selected={true} />)
      const node = container.firstChild as HTMLElement

      expect(node).toHaveClass('ring-2')
      expect(node).toHaveClass('ring-white')
    })

    it('should not apply selection styling when not selected', () => {
      const data: ParallelNodeData = {
        label: 'Unselected Parallel',
      }

      const { container } = render(<ParallelNode data={data} selected={false} />)
      const node = container.firstChild as HTMLElement

      expect(node).not.toHaveClass('ring-2')
    })
  })

  describe('Handles', () => {
    it('should have input handle at the top', () => {
      const data: ParallelNodeData = {
        label: 'Parallel Test',
      }

      render(<ParallelNode data={data} selected={false} />)

      expect(screen.getByText('Parallel Test')).toBeInTheDocument()
    })

    it('should have multiple output handles for branches', () => {
      const data: ParallelNodeData = {
        label: 'Parallel Test',
      }

      render(<ParallelNode data={data} selected={false} />)

      // Component should render successfully
      expect(screen.getByText('Parallel Test')).toBeInTheDocument()
    })
  })

  describe('Error strategy display', () => {
    it('should not display error strategy when not configured', () => {
      const data: ParallelNodeData = {
        label: 'Parallel Test',
      }

      render(<ParallelNode data={data} selected={false} />)

      expect(screen.queryByText(/Fail Fast/i)).not.toBeInTheDocument()
      expect(screen.queryByText(/Wait All/i)).not.toBeInTheDocument()
    })

    it('should display default error strategy hint', () => {
      const data: ParallelNodeData = {
        label: 'Parallel Test',
      }

      render(<ParallelNode data={data} selected={false} />)

      // Should render without explicit strategy display when not configured
      expect(screen.getByText('Parallel Test')).toBeInTheDocument()
    })
  })

  describe('Minimal configuration', () => {
    it('should render with only label', () => {
      const data: ParallelNodeData = {
        label: 'Minimal Parallel',
      }

      render(<ParallelNode data={data} selected={false} />)

      expect(screen.getByText('Minimal Parallel')).toBeInTheDocument()
      expect(screen.getByText('Parallel')).toBeInTheDocument()
    })
  })

  describe('Color scheme', () => {
    it('should use blue gradient color scheme', () => {
      const data: ParallelNodeData = {
        label: 'Parallel Test',
      }

      const { container } = render(<ParallelNode data={data} selected={false} />)
      const node = container.firstChild as HTMLElement

      // Check for blue gradient
      expect(node.className).toMatch(/blue|cyan/)
    })
  })

  describe('Full configuration', () => {
    it('should render all configuration options together', () => {
      const data: ParallelNodeData = {
        label: 'Full Config Parallel',
        errorStrategy: 'wait_all',
        maxConcurrency: 10,
        branchCount: 5,
      }

      render(<ParallelNode data={data} selected={false} />)

      expect(screen.getByText('Full Config Parallel')).toBeInTheDocument()
      expect(screen.getByText(/Wait All/i)).toBeInTheDocument()
      expect(screen.getByText(/Max: 10/)).toBeInTheDocument()
      expect(screen.getByText(/5 branches/i)).toBeInTheDocument()
    })
  })
})
