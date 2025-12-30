import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ReactFlowProvider } from '@xyflow/react'
import ActionNode from './ActionNode'
import type { ActionNodeData } from '../../stores/workflowStore'

// Wrapper component to provide ReactFlow context
function TestWrapper({ children }: { children: React.ReactNode }) {
  return <ReactFlowProvider>{children}</ReactFlowProvider>
}

describe('ActionNode', () => {
  const defaultProps = {
    id: 'action-1',
    data: {
      label: 'My Action',
      actionType: 'http',
    } as ActionNodeData,
  }

  describe('rendering', () => {
    it('should render with http action type', () => {
      render(
        <TestWrapper>
          <ActionNode {...defaultProps} />
        </TestWrapper>
      )

      expect(screen.getByText('My Action')).toBeInTheDocument()
      expect(screen.getByText('http')).toBeInTheDocument()
      expect(screen.getByText('🌐')).toBeInTheDocument()
    })

    it('should render with transform action type', () => {
      const props = {
        ...defaultProps,
        data: {
          label: 'Transform Data',
          actionType: 'transform',
        } as ActionNodeData,
      }

      render(
        <TestWrapper>
          <ActionNode {...props} />
        </TestWrapper>
      )

      expect(screen.getByText('Transform Data')).toBeInTheDocument()
      expect(screen.getByText('transform')).toBeInTheDocument()
      expect(screen.getByText('🔄')).toBeInTheDocument()
    })

    it('should render with formula action type', () => {
      const props = {
        ...defaultProps,
        data: {
          label: 'Calculate',
          actionType: 'formula',
        } as ActionNodeData,
      }

      render(
        <TestWrapper>
          <ActionNode {...props} />
        </TestWrapper>
      )

      expect(screen.getByText('Calculate')).toBeInTheDocument()
      expect(screen.getByText('🔢')).toBeInTheDocument()
    })

    it('should render with code action type', () => {
      const props = {
        ...defaultProps,
        data: {
          label: 'Run Code',
          actionType: 'code',
        } as ActionNodeData,
      }

      render(
        <TestWrapper>
          <ActionNode {...props} />
        </TestWrapper>
      )

      expect(screen.getByText('Run Code')).toBeInTheDocument()
      expect(screen.getByText('💻')).toBeInTheDocument()
    })

    it('should render with script action type', () => {
      const props = {
        ...defaultProps,
        data: {
          label: 'Execute Script',
          actionType: 'script',
        } as ActionNodeData,
      }

      render(
        <TestWrapper>
          <ActionNode {...props} />
        </TestWrapper>
      )

      expect(screen.getByText('Execute Script')).toBeInTheDocument()
      expect(screen.getByText('📜')).toBeInTheDocument()
    })

    it('should render with email action type', () => {
      const props = {
        ...defaultProps,
        data: {
          label: 'Send Email',
          actionType: 'email',
        } as ActionNodeData,
      }

      render(
        <TestWrapper>
          <ActionNode {...props} />
        </TestWrapper>
      )

      expect(screen.getByText('Send Email')).toBeInTheDocument()
      expect(screen.getByText('📧')).toBeInTheDocument()
    })

    it('should render with slack_send_message action type', () => {
      const props = {
        ...defaultProps,
        data: {
          label: 'Slack Message',
          actionType: 'slack_send_message',
        } as ActionNodeData,
      }

      render(
        <TestWrapper>
          <ActionNode {...props} />
        </TestWrapper>
      )

      expect(screen.getByText('Slack Message')).toBeInTheDocument()
      expect(screen.getByText('💬')).toBeInTheDocument()
    })

    it('should render with slack_send_dm action type', () => {
      const props = {
        ...defaultProps,
        data: {
          label: 'Direct Message',
          actionType: 'slack_send_dm',
        } as ActionNodeData,
      }

      render(
        <TestWrapper>
          <ActionNode {...props} />
        </TestWrapper>
      )

      expect(screen.getByText('Direct Message')).toBeInTheDocument()
      expect(screen.getByText('✉️')).toBeInTheDocument()
    })

    it('should render with slack_update_message action type', () => {
      const props = {
        ...defaultProps,
        data: {
          label: 'Update Message',
          actionType: 'slack_update_message',
        } as ActionNodeData,
      }

      render(
        <TestWrapper>
          <ActionNode {...props} />
        </TestWrapper>
      )

      expect(screen.getByText('Update Message')).toBeInTheDocument()
      expect(screen.getByText('✏️')).toBeInTheDocument()
    })

    it('should render with slack_add_reaction action type', () => {
      const props = {
        ...defaultProps,
        data: {
          label: 'Add Reaction',
          actionType: 'slack_add_reaction',
        } as ActionNodeData,
      }

      render(
        <TestWrapper>
          <ActionNode {...props} />
        </TestWrapper>
      )

      expect(screen.getByText('Add Reaction')).toBeInTheDocument()
      expect(screen.getByText('👍')).toBeInTheDocument()
    })

    it('should render with default icon for unknown action type', () => {
      const props = {
        ...defaultProps,
        data: {
          label: 'Custom Action',
          actionType: 'unknown_type',
        } as ActionNodeData,
      }

      render(
        <TestWrapper>
          <ActionNode {...props} />
        </TestWrapper>
      )

      expect(screen.getByText('Custom Action')).toBeInTheDocument()
      expect(screen.getByText('unknown_type')).toBeInTheDocument()
      expect(screen.getByText('⚙️')).toBeInTheDocument()
    })
  })

  describe('selection state', () => {
    it('should not have ring classes when not selected', () => {
      const { container } = render(
        <TestWrapper>
          <ActionNode {...defaultProps} selected={false} />
        </TestWrapper>
      )

      const nodeDiv = container.querySelector('.rounded-lg')
      expect(nodeDiv?.className).not.toContain('ring-2')
    })

    it('should have ring classes when selected', () => {
      const { container } = render(
        <TestWrapper>
          <ActionNode {...defaultProps} selected={true} />
        </TestWrapper>
      )

      const nodeDiv = container.querySelector('.rounded-lg')
      expect(nodeDiv?.className).toContain('ring-2')
      expect(nodeDiv?.className).toContain('ring-white')
    })
  })
})
