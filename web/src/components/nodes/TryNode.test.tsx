import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { ReactFlowProvider } from '@xyflow/react'
import TryNode from './TryNode'
import type { TryNodeData } from './TryNode'

// Wrapper component for ReactFlow context
const Wrapper = ({ children }: { children: React.ReactNode }) => (
  <ReactFlowProvider>{children}</ReactFlowProvider>
)

describe('TryNode', () => {
  const defaultData: TryNodeData = {
    label: 'Error Handler',
    config: {
      tryNodes: ['node1', 'node2'],
      catchNodes: ['catch1'],
      finallyNodes: ['finally1'],
      errorBinding: 'error',
    },
  }

  it('renders with label and node type', () => {
    render(<TryNode id="try-1" data={defaultData} />, { wrapper: Wrapper })

    expect(screen.getByText('Error Handler')).toBeInTheDocument()
    expect(screen.getByText('Try/Catch/Finally')).toBeInTheDocument()
  })

  it('displays emoji icon', () => {
    render(<TryNode id="try-1" data={defaultData} />, { wrapper: Wrapper })

    expect(screen.getByText('🛡️')).toBeInTheDocument()
  })

  it('expands and collapses on button click', () => {
    render(<TryNode id="try-1" data={defaultData} />, { wrapper: Wrapper })

    // Initially collapsed
    expect(screen.queryByText(/Try Nodes:/)).not.toBeInTheDocument()

    // Expand
    const expandButton = screen.getByText('▶')
    fireEvent.click(expandButton)

    // Should show config
    expect(screen.getByText(/Try Nodes:/)).toBeInTheDocument()
    expect(screen.getByText(/Catch Nodes:/)).toBeInTheDocument()
    expect(screen.getByText(/Finally Nodes:/)).toBeInTheDocument()

    // Collapse
    const collapseButton = screen.getByText('▼')
    fireEvent.click(collapseButton)

    // Should hide config
    expect(screen.queryByText(/Try Nodes:/)).not.toBeInTheDocument()
  })

  it('displays try nodes count when expanded', () => {
    render(<TryNode id="try-1" data={defaultData} />, { wrapper: Wrapper })

    fireEvent.click(screen.getByText('▶'))

    expect(screen.getByText(/Try Nodes: 2/)).toBeInTheDocument()
  })

  it('displays catch nodes count when expanded', () => {
    render(<TryNode id="try-1" data={defaultData} />, { wrapper: Wrapper })

    fireEvent.click(screen.getByText('▶'))

    expect(screen.getByText(/Catch Nodes: 1/)).toBeInTheDocument()
  })

  it('displays finally nodes count when expanded', () => {
    render(<TryNode id="try-1" data={defaultData} />, { wrapper: Wrapper })

    fireEvent.click(screen.getByText('▶'))

    expect(screen.getByText(/Finally Nodes: 1/)).toBeInTheDocument()
  })

  it('displays error binding when configured', () => {
    render(<TryNode id="try-1" data={defaultData} />, { wrapper: Wrapper })

    fireEvent.click(screen.getByText('▶'))

    expect(screen.getByText(/Error Var: error/)).toBeInTheDocument()
  })

  it('displays retry config when configured', () => {
    const dataWithRetry: TryNodeData = {
      ...defaultData,
      config: {
        ...defaultData.config,
        retryConfig: {
          strategy: 'exponential',
          maxAttempts: 3,
          initialDelayMs: 1000,
        },
      },
    }

    render(<TryNode id="try-1" data={dataWithRetry} />, { wrapper: Wrapper })

    fireEvent.click(screen.getByText('▶'))

    expect(screen.getByText(/Retry: exponential \(max: 3\)/)).toBeInTheDocument()
  })

  it('does not show catch section when no catch nodes', () => {
    const dataWithoutCatch: TryNodeData = {
      label: 'Try Only',
      config: {
        tryNodes: ['node1'],
      },
    }

    render(<TryNode id="try-1" data={dataWithoutCatch} />, { wrapper: Wrapper })

    fireEvent.click(screen.getByText('▶'))

    expect(screen.queryByText(/Catch Nodes:/)).not.toBeInTheDocument()
  })

  it('does not show finally section when no finally nodes', () => {
    const dataWithoutFinally: TryNodeData = {
      label: 'Try/Catch',
      config: {
        tryNodes: ['node1'],
        catchNodes: ['catch1'],
      },
    }

    render(<TryNode id="try-1" data={dataWithoutFinally} />, { wrapper: Wrapper })

    fireEvent.click(screen.getByText('▶'))

    expect(screen.queryByText(/Finally Nodes:/)).not.toBeInTheDocument()
  })

  it('applies selected styles when selected', () => {
    const { container } = render(
      <TryNode id="try-1" data={defaultData} selected={true} />,
      { wrapper: Wrapper }
    )

    const nodeElement = container.querySelector('.ring-2')
    expect(nodeElement).toBeInTheDocument()
  })

  it('does not apply selected styles when not selected', () => {
    const { container } = render(
      <TryNode id="try-1" data={defaultData} selected={false} />,
      { wrapper: Wrapper }
    )

    const nodeElement = container.querySelector('.ring-2')
    expect(nodeElement).not.toBeInTheDocument()
  })

  it('shows all branch indicators', () => {
    render(<TryNode id="try-1" data={defaultData} />, { wrapper: Wrapper })

    expect(screen.getByText('Try')).toBeInTheDocument()
    expect(screen.getByText('Catch')).toBeInTheDocument()
    expect(screen.getByText('Finally')).toBeInTheDocument()
  })

  it('handles minimal configuration', () => {
    const minimalData: TryNodeData = {
      label: 'Simple Try',
      config: {
        tryNodes: ['node1'],
      },
    }

    render(<TryNode id="try-1" data={minimalData} />, { wrapper: Wrapper })

    expect(screen.getByText('Simple Try')).toBeInTheDocument()
    expect(screen.getByText('Try/Catch/Finally')).toBeInTheDocument()
  })
})
