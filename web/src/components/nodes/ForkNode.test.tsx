import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import ForkNode from './ForkNode'

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

interface ForkNodeData {
  label: string
  branchCount?: number
}

describe('ForkNode', () => {
  describe('Basic rendering', () => {
    it('should render fork node with label', () => {
      const data: ForkNodeData = {
        label: 'Split into branches',
      }

      render(<ForkNode data={data} selected={false} />)

      expect(screen.getByText('Split into branches')).toBeInTheDocument()
      expect(screen.getByText('Fork')).toBeInTheDocument()
    })

    it('should render with fork icon', () => {
      const data: ForkNodeData = {
        label: 'Fork Test',
      }

      render(<ForkNode data={data} selected={false} />)

      expect(screen.getByText('🔱')).toBeInTheDocument()
    })

    it('should have correct styling', () => {
      const data: ForkNodeData = {
        label: 'Styled Fork',
      }

      const { container } = render(<ForkNode data={data} selected={false} />)
      const node = container.firstChild as HTMLElement

      expect(node).toHaveClass('px-4')
      expect(node).toHaveClass('py-3')
      expect(node).toHaveClass('rounded-lg')
      expect(node).toHaveClass('shadow-lg')
    })
  })

  describe('Configuration display', () => {
    it('should display branch count when provided', () => {
      const data: ForkNodeData = {
        label: 'Fork Test',
        branchCount: 3,
      }

      render(<ForkNode data={data} selected={false} />)

      expect(screen.getByText(/3 branches/i)).toBeInTheDocument()
    })

    it('should display different branch counts', () => {
      const data: ForkNodeData = {
        label: 'Fork Test',
        branchCount: 5,
      }

      render(<ForkNode data={data} selected={false} />)

      expect(screen.getByText(/5 branches/i)).toBeInTheDocument()
    })

    it('should not display branch count when not provided', () => {
      const data: ForkNodeData = {
        label: 'Fork Test',
      }

      render(<ForkNode data={data} selected={false} />)

      expect(screen.queryByText(/branches/i)).not.toBeInTheDocument()
    })
  })

  describe('Selection state', () => {
    it('should apply selection styling when selected', () => {
      const data: ForkNodeData = {
        label: 'Selected Fork',
      }

      const { container } = render(<ForkNode data={data} selected={true} />)
      const node = container.firstChild as HTMLElement

      expect(node).toHaveClass('ring-2')
      expect(node).toHaveClass('ring-white')
    })

    it('should not apply selection styling when not selected', () => {
      const data: ForkNodeData = {
        label: 'Unselected Fork',
      }

      const { container } = render(<ForkNode data={data} selected={false} />)
      const node = container.firstChild as HTMLElement

      expect(node).not.toHaveClass('ring-2')
    })
  })

  describe('Handles', () => {
    it('should have input handle at the top', () => {
      const data: ForkNodeData = {
        label: 'Fork Test',
      }

      render(<ForkNode data={data} selected={false} />)

      expect(screen.getByText('Fork Test')).toBeInTheDocument()
    })

    it('should have output handle at the bottom', () => {
      const data: ForkNodeData = {
        label: 'Fork Test',
      }

      render(<ForkNode data={data} selected={false} />)

      expect(screen.getByText('Fork Test')).toBeInTheDocument()
    })
  })

  describe('Color scheme', () => {
    it('should use green gradient color scheme', () => {
      const data: ForkNodeData = {
        label: 'Fork Test',
      }

      const { container } = render(<ForkNode data={data} selected={false} />)
      const node = container.firstChild as HTMLElement

      expect(node.className).toMatch(/green|emerald/)
    })
  })

  describe('Minimal configuration', () => {
    it('should render with only label', () => {
      const data: ForkNodeData = {
        label: 'Minimal Fork',
      }

      render(<ForkNode data={data} selected={false} />)

      expect(screen.getByText('Minimal Fork')).toBeInTheDocument()
      expect(screen.getByText('Fork')).toBeInTheDocument()
    })
  })
})
