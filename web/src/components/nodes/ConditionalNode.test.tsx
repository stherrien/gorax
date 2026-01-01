import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ReactFlowProvider } from '@xyflow/react'
import ConditionalNode from './ConditionalNode'
import type { ConditionalNodeData } from './ConditionalNode'

// Wrapper component to provide ReactFlow context
function TestWrapper({ children }: { children: React.ReactNode }) {
  return <ReactFlowProvider>{children}</ReactFlowProvider>
}

describe('ConditionalNode', () => {
  const defaultProps = {
    id: 'condition-1',
    data: {
      label: 'My Condition',
    } as ConditionalNodeData,
  }

  describe('rendering', () => {
    it('should render with label', () => {
      render(
        <TestWrapper>
          <ConditionalNode {...defaultProps} />
        </TestWrapper>
      )

      expect(screen.getByText('My Condition')).toBeInTheDocument()
      expect(screen.getByText('Conditional')).toBeInTheDocument()
      expect(screen.getByText('🔀')).toBeInTheDocument()
    })

    it('should render condition when provided', () => {
      const props = {
        ...defaultProps,
        data: {
          label: 'Check Status',
          condition: 'status === "success"',
        } as ConditionalNodeData,
      }

      render(
        <TestWrapper>
          <ConditionalNode {...props} />
        </TestWrapper>
      )

      expect(screen.getByText('Check Status')).toBeInTheDocument()
      expect(screen.getByText('status === "success"')).toBeInTheDocument()
    })

    it('should render description when provided', () => {
      const props = {
        ...defaultProps,
        data: {
          label: 'Validate Input',
          description: 'Checks if input is valid',
        } as ConditionalNodeData,
      }

      render(
        <TestWrapper>
          <ConditionalNode {...props} />
        </TestWrapper>
      )

      expect(screen.getByText('Validate Input')).toBeInTheDocument()
      expect(screen.getByText('Checks if input is valid')).toBeInTheDocument()
    })

    it('should render both condition and description when provided', () => {
      const props = {
        ...defaultProps,
        data: {
          label: 'Complex Check',
          condition: 'value > 100 && isActive',
          description: 'Checks value threshold and status',
        } as ConditionalNodeData,
      }

      render(
        <TestWrapper>
          <ConditionalNode {...props} />
        </TestWrapper>
      )

      expect(screen.getByText('Complex Check')).toBeInTheDocument()
      expect(screen.getByText('value > 100 && isActive')).toBeInTheDocument()
      expect(screen.getByText('Checks value threshold and status')).toBeInTheDocument()
    })

    it('should not render condition block when not provided', () => {
      render(
        <TestWrapper>
          <ConditionalNode {...defaultProps} />
        </TestWrapper>
      )

      // The condition block has font-mono class
      const conditionBlock = document.querySelector('.font-mono')
      expect(conditionBlock).toBeNull()
    })

    it('should render True and False output labels', () => {
      render(
        <TestWrapper>
          <ConditionalNode {...defaultProps} />
        </TestWrapper>
      )

      expect(screen.getByText('True')).toBeInTheDocument()
      expect(screen.getByText('False')).toBeInTheDocument()
    })
  })

  describe('selection state', () => {
    it('should not have ring classes when not selected', () => {
      const { container } = render(
        <TestWrapper>
          <ConditionalNode {...defaultProps} selected={false} />
        </TestWrapper>
      )

      const nodeDiv = container.querySelector('.rounded-lg')
      expect(nodeDiv?.className).not.toContain('ring-2')
    })

    it('should have ring classes when selected', () => {
      const { container } = render(
        <TestWrapper>
          <ConditionalNode {...defaultProps} selected={true} />
        </TestWrapper>
      )

      const nodeDiv = container.querySelector('.rounded-lg')
      expect(nodeDiv?.className).toContain('ring-2')
      expect(nodeDiv?.className).toContain('ring-white')
    })
  })
})
