import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ExecutionStatusNode } from './ExecutionStatusNode'
import { useExecutionTraceStore } from '../../stores/executionTraceStore'
import type { NodeProps } from '@xyflow/react'

// Mock the execution trace store
vi.mock('../../stores/executionTraceStore', () => ({
  useExecutionTraceStore: vi.fn(),
}))

// Sample base node component for testing
function TestNode({ data, selected }: { data: any; selected?: boolean }) {
  return (
    <div data-testid="base-node" data-selected={selected}>
      <div data-testid="node-label">{data.label}</div>
    </div>
  )
}

describe('ExecutionStatusNode', () => {
  beforeEach(() => {
    // Reset mock before each test
    vi.clearAllMocks()
  })

  describe('Node status rendering', () => {
    it('should render base node without status when no execution is active', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        nodeStatuses: {},
      } as any)

      const WrappedNode = ExecutionStatusNode(TestNode)
      const props: NodeProps = {
        id: 'node-1',
        data: { label: 'Test Node' },
        selected: false,
      } as any

      render(<WrappedNode {...props} />)

      expect(screen.getByTestId('base-node')).toBeInTheDocument()
      expect(screen.getByTestId('node-label')).toHaveTextContent('Test Node')
    })

    it('should render with pending status', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        nodeStatuses: { 'node-1': 'pending' },
      } as any)

      const WrappedNode = ExecutionStatusNode(TestNode)
      const props: NodeProps = {
        id: 'node-1',
        data: { label: 'Test Node' },
        selected: false,
      } as any

      render(<WrappedNode {...props} />)

      const wrapper = screen.getByTestId('execution-status-wrapper')
      expect(wrapper).toHaveClass('execution-status-pending')
    })

    it('should render with running status', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        nodeStatuses: { 'node-1': 'running' },
      } as any)

      const WrappedNode = ExecutionStatusNode(TestNode)
      const props: NodeProps = {
        id: 'node-1',
        data: { label: 'Test Node' },
        selected: false,
      } as any

      render(<WrappedNode {...props} />)

      const wrapper = screen.getByTestId('execution-status-wrapper')
      expect(wrapper).toHaveClass('execution-status-running')
    })

    it('should render with completed status', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        nodeStatuses: { 'node-1': 'completed' },
      } as any)

      const WrappedNode = ExecutionStatusNode(TestNode)
      const props: NodeProps = {
        id: 'node-1',
        data: { label: 'Test Node' },
        selected: false,
      } as any

      render(<WrappedNode {...props} />)

      const wrapper = screen.getByTestId('execution-status-wrapper')
      expect(wrapper).toHaveClass('execution-status-completed')
    })

    it('should render with failed status', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        nodeStatuses: { 'node-1': 'failed' },
      } as any)

      const WrappedNode = ExecutionStatusNode(TestNode)
      const props: NodeProps = {
        id: 'node-1',
        data: { label: 'Test Node' },
        selected: false,
      } as any

      render(<WrappedNode {...props} />)

      const wrapper = screen.getByTestId('execution-status-wrapper')
      expect(wrapper).toHaveClass('execution-status-failed')
    })
  })

  describe('Progress indicators', () => {
    it('should show spinner icon for running status', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        nodeStatuses: { 'node-1': 'running' },
      } as any)

      const WrappedNode = ExecutionStatusNode(TestNode)
      const props: NodeProps = {
        id: 'node-1',
        data: { label: 'Test Node' },
        selected: false,
      } as any

      render(<WrappedNode {...props} />)

      expect(screen.getByTestId('status-indicator-running')).toBeInTheDocument()
    })

    it('should show checkmark icon for completed status', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        nodeStatuses: { 'node-1': 'completed' },
      } as any)

      const WrappedNode = ExecutionStatusNode(TestNode)
      const props: NodeProps = {
        id: 'node-1',
        data: { label: 'Test Node' },
        selected: false,
      } as any

      render(<WrappedNode {...props} />)

      expect(screen.getByTestId('status-indicator-completed')).toBeInTheDocument()
    })

    it('should show error icon for failed status', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        nodeStatuses: { 'node-1': 'failed' },
      } as any)

      const WrappedNode = ExecutionStatusNode(TestNode)
      const props: NodeProps = {
        id: 'node-1',
        data: { label: 'Test Node' },
        selected: false,
      } as any

      render(<WrappedNode {...props} />)

      expect(screen.getByTestId('status-indicator-failed')).toBeInTheDocument()
    })

    it('should show clock icon for pending status', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        nodeStatuses: { 'node-1': 'pending' },
      } as any)

      const WrappedNode = ExecutionStatusNode(TestNode)
      const props: NodeProps = {
        id: 'node-1',
        data: { label: 'Test Node' },
        selected: false,
      } as any

      render(<WrappedNode {...props} />)

      expect(screen.getByTestId('status-indicator-pending')).toBeInTheDocument()
    })

    it('should not show indicator when no execution is active', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        nodeStatuses: {},
      } as any)

      const WrappedNode = ExecutionStatusNode(TestNode)
      const props: NodeProps = {
        id: 'node-1',
        data: { label: 'Test Node' },
        selected: false,
      } as any

      render(<WrappedNode {...props} />)

      expect(screen.queryByTestId(/status-indicator/)).not.toBeInTheDocument()
    })
  })

  describe('Node highlighting and animations', () => {
    it('should apply pulse animation class for running status', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        nodeStatuses: { 'node-1': 'running' },
      } as any)

      const WrappedNode = ExecutionStatusNode(TestNode)
      const props: NodeProps = {
        id: 'node-1',
        data: { label: 'Test Node' },
        selected: false,
      } as any

      render(<WrappedNode {...props} />)

      const wrapper = screen.getByTestId('execution-status-wrapper')
      expect(wrapper).toHaveClass('animate-pulse-glow')
    })

    it('should apply success animation class for completed status', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        nodeStatuses: { 'node-1': 'completed' },
      } as any)

      const WrappedNode = ExecutionStatusNode(TestNode)
      const props: NodeProps = {
        id: 'node-1',
        data: { label: 'Test Node' },
        selected: false,
      } as any

      render(<WrappedNode {...props} />)

      const wrapper = screen.getByTestId('execution-status-wrapper')
      expect(wrapper).toHaveClass('animate-checkmark')
    })

    it('should apply error animation class for failed status', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        nodeStatuses: { 'node-1': 'failed' },
      } as any)

      const WrappedNode = ExecutionStatusNode(TestNode)
      const props: NodeProps = {
        id: 'node-1',
        data: { label: 'Test Node' },
        selected: false,
      } as any

      render(<WrappedNode {...props} />)

      const wrapper = screen.getByTestId('execution-status-wrapper')
      expect(wrapper).toHaveClass('animate-shake')
    })
  })

  describe('Integration with base node', () => {
    it('should pass through all props to base node', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        nodeStatuses: {},
      } as any)

      const WrappedNode = ExecutionStatusNode(TestNode)
      const props: NodeProps = {
        id: 'node-1',
        data: { label: 'Test Node', custom: 'value' },
        selected: true,
      } as any

      render(<WrappedNode {...props} />)

      expect(screen.getByTestId('base-node')).toBeInTheDocument()
      expect(screen.getByTestId('node-label')).toHaveTextContent('Test Node')
      expect(screen.getByTestId('base-node')).toHaveAttribute('data-selected', 'true')
    })

    it('should work with different node types', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        nodeStatuses: { 'node-1': 'running' },
      } as any)

      function AnotherNode({ data }: { data: any }) {
        return <div data-testid="another-node">{data.title}</div>
      }

      const WrappedNode = ExecutionStatusNode(AnotherNode)
      const props: NodeProps = {
        id: 'node-1',
        data: { title: 'Another Node' },
        selected: false,
      } as any

      render(<WrappedNode {...props} />)

      expect(screen.getByTestId('another-node')).toHaveTextContent('Another Node')
      expect(screen.getByTestId('execution-status-wrapper')).toHaveClass('execution-status-running')
    })

    it('should preserve base node functionality', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        nodeStatuses: {},
      } as any)

      function InteractiveNode({ data, selected }: { data: any; selected?: boolean }) {
        return (
          <div data-testid="interactive-node">
            <button onClick={() => data.onClick?.()}>Click me</button>
            {selected && <span data-testid="selected-indicator">Selected</span>}
          </div>
        )
      }

      const WrappedNode = ExecutionStatusNode(InteractiveNode)
      const mockClick = vi.fn()
      const props: NodeProps = {
        id: 'node-1',
        data: { label: 'Interactive', onClick: mockClick },
        selected: true,
      } as any

      render(<WrappedNode {...props} />)

      expect(screen.getByTestId('selected-indicator')).toBeInTheDocument()
      screen.getByText('Click me').click()
      expect(mockClick).toHaveBeenCalled()
    })
  })

  describe('TypeScript generics', () => {
    it('should work with strongly typed node data', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        nodeStatuses: {},
      } as any)

      interface CustomNodeData {
        label: string
        count: number
        enabled: boolean
      }

      function TypedNode({ data }: { data: CustomNodeData }) {
        return (
          <div data-testid="typed-node">
            <span>{data.label}</span>
            <span>{data.count}</span>
            <span>{data.enabled ? 'enabled' : 'disabled'}</span>
          </div>
        )
      }

      const WrappedNode = ExecutionStatusNode(TypedNode)
      const props: NodeProps<CustomNodeData> = {
        id: 'node-1',
        data: { label: 'Typed', count: 42, enabled: true },
        selected: false,
      } as any

      render(<WrappedNode {...props} />)

      const node = screen.getByTestId('typed-node')
      expect(node).toHaveTextContent('Typed')
      expect(node).toHaveTextContent('42')
      expect(node).toHaveTextContent('enabled')
    })
  })

  describe('Multiple node instances', () => {
    it('should track different statuses for different nodes', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        nodeStatuses: {
          'node-1': 'completed',
          'node-2': 'running',
          'node-3': 'failed',
        },
      } as any)

      const WrappedNode = ExecutionStatusNode(TestNode)

      const { rerender } = render(
        <WrappedNode
          id="node-1"
          data={{ label: 'Node 1' }}
          selected={false}
          {...({} as any)}
        />
      )

      expect(screen.getByTestId('execution-status-wrapper')).toHaveClass('execution-status-completed')

      rerender(
        <WrappedNode
          id="node-2"
          data={{ label: 'Node 2' }}
          selected={false}
          {...({} as any)}
        />
      )

      expect(screen.getByTestId('execution-status-wrapper')).toHaveClass('execution-status-running')

      rerender(
        <WrappedNode
          id="node-3"
          data={{ label: 'Node 3' }}
          selected={false}
          {...({} as any)}
        />
      )

      expect(screen.getByTestId('execution-status-wrapper')).toHaveClass('execution-status-failed')
    })
  })
})
