import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ReactFlowProvider } from '@xyflow/react'
import TriggerNode from './TriggerNode'
import type { TriggerNodeData } from '../../stores/workflowStore'

// Wrapper component to provide ReactFlow context
function TestWrapper({ children }: { children: React.ReactNode }) {
  return <ReactFlowProvider>{children}</ReactFlowProvider>
}

describe('TriggerNode', () => {
  const defaultProps = {
    id: 'trigger-1',
    data: {
      label: 'My Trigger',
      triggerType: 'webhook',
    } as TriggerNodeData,
  }

  describe('rendering', () => {
    it('should render with webhook trigger type', () => {
      render(
        <TestWrapper>
          <TriggerNode {...defaultProps} />
        </TestWrapper>
      )

      expect(screen.getByText('My Trigger')).toBeInTheDocument()
      expect(screen.getByText('webhook')).toBeInTheDocument()
      expect(screen.getByText('🔗')).toBeInTheDocument()
    })

    it('should render with schedule trigger type', () => {
      const props = {
        ...defaultProps,
        data: {
          label: 'Daily Schedule',
          triggerType: 'schedule',
        } as TriggerNodeData,
      }

      render(
        <TestWrapper>
          <TriggerNode {...props} />
        </TestWrapper>
      )

      expect(screen.getByText('Daily Schedule')).toBeInTheDocument()
      expect(screen.getByText('schedule')).toBeInTheDocument()
      expect(screen.getByText('⏰')).toBeInTheDocument()
    })

    it('should render with default icon for unknown trigger type', () => {
      const props = {
        ...defaultProps,
        data: {
          label: 'Unknown Trigger',
          triggerType: 'custom',
        } as TriggerNodeData,
      }

      render(
        <TestWrapper>
          <TriggerNode {...props} />
        </TestWrapper>
      )

      expect(screen.getByText('Unknown Trigger')).toBeInTheDocument()
      expect(screen.getByText('custom')).toBeInTheDocument()
      expect(screen.getByText('📥')).toBeInTheDocument()
    })
  })

  describe('selection state', () => {
    it('should not have ring classes when not selected', () => {
      const { container } = render(
        <TestWrapper>
          <TriggerNode {...defaultProps} selected={false} />
        </TestWrapper>
      )

      const nodeDiv = container.querySelector('.rounded-lg')
      expect(nodeDiv?.className).not.toContain('ring-2')
    })

    it('should have ring classes when selected', () => {
      const { container } = render(
        <TestWrapper>
          <TriggerNode {...defaultProps} selected={true} />
        </TestWrapper>
      )

      const nodeDiv = container.querySelector('.rounded-lg')
      expect(nodeDiv?.className).toContain('ring-2')
      expect(nodeDiv?.className).toContain('ring-white')
    })
  })
})
