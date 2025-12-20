import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import JoinNode from './JoinNode'

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

interface JoinNodeData {
  label: string
  joinStrategy?: 'wait_all' | 'wait_n'
  requiredCount?: number
  timeoutMs?: number
  onTimeout?: 'fail' | 'continue'
}

describe('JoinNode', () => {
  describe('Basic rendering', () => {
    it('should render join node with label', () => {
      const data: JoinNodeData = {
        label: 'Merge branches',
      }

      render(<JoinNode data={data} selected={false} />)

      expect(screen.getByText('Merge branches')).toBeInTheDocument()
      expect(screen.getByText('Join')).toBeInTheDocument()
    })

    it('should render with join icon', () => {
      const data: JoinNodeData = {
        label: 'Join Test',
      }

      render(<JoinNode data={data} selected={false} />)

      expect(screen.getByText('⚡')).toBeInTheDocument()
    })

    it('should have correct styling', () => {
      const data: JoinNodeData = {
        label: 'Styled Join',
      }

      const { container } = render(<JoinNode data={data} selected={false} />)
      const node = container.firstChild as HTMLElement

      expect(node).toHaveClass('px-4')
      expect(node).toHaveClass('py-3')
      expect(node).toHaveClass('rounded-lg')
      expect(node).toHaveClass('shadow-lg')
    })
  })

  describe('Configuration display', () => {
    it('should display wait_all strategy', () => {
      const data: JoinNodeData = {
        label: 'Join Test',
        joinStrategy: 'wait_all',
      }

      render(<JoinNode data={data} selected={false} />)

      expect(screen.getByText(/Wait All/i)).toBeInTheDocument()
    })

    it('should display wait_n strategy with count', () => {
      const data: JoinNodeData = {
        label: 'Join Test',
        joinStrategy: 'wait_n',
        requiredCount: 2,
      }

      render(<JoinNode data={data} selected={false} />)

      expect(screen.getByText(/Wait 2/i)).toBeInTheDocument()
    })

    it('should display timeout when provided', () => {
      const data: JoinNodeData = {
        label: 'Join Test',
        timeoutMs: 5000,
      }

      render(<JoinNode data={data} selected={false} />)

      expect(screen.getByText(/Timeout: 5000ms/i)).toBeInTheDocument()
    })

    it('should display on timeout behavior', () => {
      const data: JoinNodeData = {
        label: 'Join Test',
        timeoutMs: 1000,
        onTimeout: 'continue',
      }

      render(<JoinNode data={data} selected={false} />)

      expect(screen.getByText(/Continue on timeout/i)).toBeInTheDocument()
    })

    it('should not display timeout when not provided', () => {
      const data: JoinNodeData = {
        label: 'Join Test',
      }

      render(<JoinNode data={data} selected={false} />)

      expect(screen.queryByText(/Timeout/i)).not.toBeInTheDocument()
    })
  })

  describe('Selection state', () => {
    it('should apply selection styling when selected', () => {
      const data: JoinNodeData = {
        label: 'Selected Join',
      }

      const { container } = render(<JoinNode data={data} selected={true} />)
      const node = container.firstChild as HTMLElement

      expect(node).toHaveClass('ring-2')
      expect(node).toHaveClass('ring-white')
    })

    it('should not apply selection styling when not selected', () => {
      const data: JoinNodeData = {
        label: 'Unselected Join',
      }

      const { container } = render(<JoinNode data={data} selected={false} />)
      const node = container.firstChild as HTMLElement

      expect(node).not.toHaveClass('ring-2')
    })
  })

  describe('Color scheme', () => {
    it('should use orange gradient color scheme', () => {
      const data: JoinNodeData = {
        label: 'Join Test',
      }

      const { container } = render(<JoinNode data={data} selected={false} />)
      const node = container.firstChild as HTMLElement

      expect(node.className).toMatch(/orange|amber/)
    })
  })

  describe('Minimal configuration', () => {
    it('should render with only label', () => {
      const data: JoinNodeData = {
        label: 'Minimal Join',
      }

      render(<JoinNode data={data} selected={false} />)

      expect(screen.getByText('Minimal Join')).toBeInTheDocument()
      expect(screen.getByText('Join')).toBeInTheDocument()
    })
  })

  describe('Full configuration', () => {
    it('should render all configuration options together', () => {
      const data: JoinNodeData = {
        label: 'Full Config Join',
        joinStrategy: 'wait_n',
        requiredCount: 3,
        timeoutMs: 10000,
        onTimeout: 'continue',
      }

      render(<JoinNode data={data} selected={false} />)

      expect(screen.getByText('Full Config Join')).toBeInTheDocument()
      expect(screen.getByText(/Wait 3/i)).toBeInTheDocument()
      expect(screen.getByText(/Timeout: 10000ms/i)).toBeInTheDocument()
      expect(screen.getByText(/Continue on timeout/i)).toBeInTheDocument()
    })
  })
})
