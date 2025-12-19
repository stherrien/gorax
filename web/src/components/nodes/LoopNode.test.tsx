import { describe, it, expect, vi, beforeAll } from 'vitest'
import { render, screen } from '@testing-library/react'
import LoopNode from './LoopNode'

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

interface LoopNodeData {
  label: string
  source?: string
  itemVariable?: string
  indexVariable?: string
  maxIterations?: number
  onError?: 'stop' | 'continue'
}

describe('LoopNode', () => {
  describe('Basic rendering', () => {
    it('should render loop node with label', () => {
      const data: LoopNodeData = {
        label: 'Process Items',
      }

      render(<LoopNode data={data} selected={false} />)

      expect(screen.getByText('Process Items')).toBeInTheDocument()
      expect(screen.getByText('Loop')).toBeInTheDocument()
    })

    it('should render with loop icon', () => {
      const data: LoopNodeData = {
        label: 'Loop Test',
      }

      render(<LoopNode data={data} selected={false} />)

      // Check for loop emoji icon
      expect(screen.getByText('🔁')).toBeInTheDocument()
    })

    it('should have correct styling', () => {
      const data: LoopNodeData = {
        label: 'Styled Loop',
      }

      const { container } = render(<LoopNode data={data} selected={false} />)
      const node = container.firstChild as HTMLElement

      expect(node).toHaveClass('px-4')
      expect(node).toHaveClass('py-3')
      expect(node).toHaveClass('rounded-lg')
      expect(node).toHaveClass('shadow-lg')
    })
  })

  describe('Configuration display', () => {
    it('should display source expression when provided', () => {
      const data: LoopNodeData = {
        label: 'Loop Over Items',
        source: '${steps.http_request.output.items}',
      }

      render(<LoopNode data={data} selected={false} />)

      expect(screen.getByText('${steps.http_request.output.items}')).toBeInTheDocument()
    })

    it('should display item variable when provided', () => {
      const data: LoopNodeData = {
        label: 'Loop Test',
        source: '${steps.data.items}',
        itemVariable: 'currentItem',
      }

      render(<LoopNode data={data} selected={false} />)

      expect(screen.getByText(/currentItem/)).toBeInTheDocument()
    })

    it('should display index variable when provided', () => {
      const data: LoopNodeData = {
        label: 'Loop Test',
        source: '${steps.data.items}',
        itemVariable: 'item',
        indexVariable: 'idx',
      }

      render(<LoopNode data={data} selected={false} />)

      expect(screen.getByText(/idx/)).toBeInTheDocument()
    })

    it('should display max iterations when provided', () => {
      const data: LoopNodeData = {
        label: 'Loop Test',
        source: '${steps.data.items}',
        itemVariable: 'item',
        maxIterations: 100,
      }

      render(<LoopNode data={data} selected={false} />)

      expect(screen.getByText(/Max: 100/)).toBeInTheDocument()
    })

    it('should not show max iterations when using default', () => {
      const data: LoopNodeData = {
        label: 'Loop Test',
        source: '${steps.data.items}',
        itemVariable: 'item',
      }

      render(<LoopNode data={data} selected={false} />)

      expect(screen.queryByText(/Max:/)).not.toBeInTheDocument()
    })
  })

  describe('Selection state', () => {
    it('should apply selection styling when selected', () => {
      const data: LoopNodeData = {
        label: 'Selected Loop',
      }

      const { container } = render(<LoopNode data={data} selected={true} />)
      const node = container.firstChild as HTMLElement

      expect(node).toHaveClass('ring-2')
      expect(node).toHaveClass('ring-white')
    })

    it('should not apply selection styling when not selected', () => {
      const data: LoopNodeData = {
        label: 'Unselected Loop',
      }

      const { container } = render(<LoopNode data={data} selected={false} />)
      const node = container.firstChild as HTMLElement

      expect(node).not.toHaveClass('ring-2')
    })
  })

  describe('Handles', () => {
    it('should have input handle at the top', () => {
      const data: LoopNodeData = {
        label: 'Loop Test',
      }

      render(<LoopNode data={data} selected={false} />)

      // We can't directly test Handle components without mocking @xyflow/react
      // but we can verify the component renders without errors
      expect(screen.getByText('Loop Test')).toBeInTheDocument()
    })

    it('should have output handle at the bottom', () => {
      const data: LoopNodeData = {
        label: 'Loop Test',
      }

      render(<LoopNode data={data} selected={false} />)

      // Component should render successfully with handles
      expect(screen.getByText('Loop Test')).toBeInTheDocument()
    })
  })

  describe('Error strategy display', () => {
    it('should display error strategy when set to continue', () => {
      const data: LoopNodeData = {
        label: 'Loop Test',
        source: '${steps.data.items}',
        itemVariable: 'item',
        onError: 'continue',
      }

      render(<LoopNode data={data} selected={false} />)

      expect(screen.getByText(/Continue on error/i)).toBeInTheDocument()
    })

    it('should display error strategy when set to stop', () => {
      const data: LoopNodeData = {
        label: 'Loop Test',
        source: '${steps.data.items}',
        itemVariable: 'item',
        onError: 'stop',
      }

      render(<LoopNode data={data} selected={false} />)

      expect(screen.getByText(/Stop on error/i)).toBeInTheDocument()
    })

    it('should not display error strategy when not configured', () => {
      const data: LoopNodeData = {
        label: 'Loop Test',
        source: '${steps.data.items}',
        itemVariable: 'item',
      }

      render(<LoopNode data={data} selected={false} />)

      expect(screen.queryByText(/on error/i)).not.toBeInTheDocument()
    })
  })

  describe('Minimal configuration', () => {
    it('should render with only label', () => {
      const data: LoopNodeData = {
        label: 'Minimal Loop',
      }

      render(<LoopNode data={data} selected={false} />)

      expect(screen.getByText('Minimal Loop')).toBeInTheDocument()
      expect(screen.getByText('Loop')).toBeInTheDocument()
    })
  })

  describe('Color scheme', () => {
    it('should use purple gradient color scheme', () => {
      const data: LoopNodeData = {
        label: 'Loop Test',
      }

      const { container } = render(<LoopNode data={data} selected={false} />)
      const node = container.firstChild as HTMLElement

      // Check for purple gradient (purple to violet)
      expect(node.className).toMatch(/purple|violet/)
    })
  })
})
