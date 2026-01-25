import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import {
  WorkflowErrorBoundary,
  CanvasErrorBoundary,
  PanelErrorBoundary,
  NodeErrorBoundary,
} from './WorkflowErrorBoundary'

// Mock the error logger
vi.mock('../../services/errorLogger', () => ({
  errorLogger: {
    logBoundaryError: vi.fn(),
  },
}))

// Component that throws an error
function ThrowingComponent({ shouldThrow = true }: { shouldThrow?: boolean }) {
  if (shouldThrow) {
    throw new Error('Test error')
  }
  return <div>Healthy component</div>
}

// Suppress console.error for cleaner test output
const originalError = console.error
beforeEach(() => {
  console.error = vi.fn()
})
afterEach(() => {
  console.error = originalError
})

describe('WorkflowErrorBoundary', () => {
  it('renders children when there is no error', () => {
    render(
      <WorkflowErrorBoundary>
        <div>Test content</div>
      </WorkflowErrorBoundary>
    )

    expect(screen.getByText('Test content')).toBeInTheDocument()
  })

  it('renders inline fallback by default when error occurs', () => {
    render(
      <WorkflowErrorBoundary>
        <ThrowingComponent />
      </WorkflowErrorBoundary>
    )

    expect(screen.getByText('Component Error')).toBeInTheDocument()
    expect(screen.getByText('This component encountered an error.')).toBeInTheDocument()
  })

  it('calls onError callback when error occurs', () => {
    const onError = vi.fn()

    render(
      <WorkflowErrorBoundary onError={onError}>
        <ThrowingComponent />
      </WorkflowErrorBoundary>
    )

    expect(onError).toHaveBeenCalledTimes(1)
    expect(onError).toHaveBeenCalledWith(
      expect.any(Error),
      expect.objectContaining({ componentStack: expect.any(String) })
    )
  })

  it('renders custom fallback when provided', () => {
    render(
      <WorkflowErrorBoundary fallback={<div>Custom fallback</div>}>
        <ThrowingComponent />
      </WorkflowErrorBoundary>
    )

    expect(screen.getByText('Custom fallback')).toBeInTheDocument()
  })

  it('allows retry after error', () => {
    render(
      <WorkflowErrorBoundary>
        <ThrowingComponent shouldThrow={true} />
      </WorkflowErrorBoundary>
    )

    expect(screen.getByText('Component Error')).toBeInTheDocument()

    // Click retry
    fireEvent.click(screen.getByText('Retry'))

    // The boundary should reset but component still throws
    // In real usage, the component would be fixed before retry
    expect(screen.getByText('Component Error')).toBeInTheDocument()
  })

  it('renders canvas fallback when fallbackType is canvas', () => {
    render(
      <WorkflowErrorBoundary fallbackType="canvas">
        <ThrowingComponent />
      </WorkflowErrorBoundary>
    )

    expect(screen.getByText('Canvas Error')).toBeInTheDocument()
    expect(screen.getByText(/The workflow canvas encountered an error/)).toBeInTheDocument()
  })

  it('renders panel fallback when fallbackType is panel', () => {
    render(
      <WorkflowErrorBoundary fallbackType="panel">
        <ThrowingComponent />
      </WorkflowErrorBoundary>
    )

    expect(screen.getByText('Panel Error')).toBeInTheDocument()
    expect(screen.getByText('This panel encountered an error.')).toBeInTheDocument()
  })
})

describe('CanvasErrorBoundary', () => {
  it('renders children when there is no error', () => {
    render(
      <CanvasErrorBoundary>
        <div>Canvas content</div>
      </CanvasErrorBoundary>
    )

    expect(screen.getByText('Canvas content')).toBeInTheDocument()
  })

  it('renders canvas-specific fallback when error occurs', () => {
    render(
      <CanvasErrorBoundary workflowId="test-workflow-123">
        <ThrowingComponent />
      </CanvasErrorBoundary>
    )

    expect(screen.getByText('Canvas Error')).toBeInTheDocument()
    expect(screen.getByText(/The workflow canvas encountered an error/)).toBeInTheDocument()
    expect(screen.getByText('Retry')).toBeInTheDocument()
    expect(screen.getByText('Reset Canvas')).toBeInTheDocument()
    expect(screen.getByText('Refresh Page')).toBeInTheDocument()
  })

  it('calls onError with workflowId context', () => {
    const onError = vi.fn()

    render(
      <CanvasErrorBoundary workflowId="test-workflow-123" onError={onError}>
        <ThrowingComponent />
      </CanvasErrorBoundary>
    )

    expect(onError).toHaveBeenCalledTimes(1)
  })
})

describe('PanelErrorBoundary', () => {
  it('renders children when there is no error', () => {
    render(
      <PanelErrorBoundary>
        <div>Panel content</div>
      </PanelErrorBoundary>
    )

    expect(screen.getByText('Panel content')).toBeInTheDocument()
  })

  it('renders panel-specific fallback when error occurs', () => {
    render(
      <PanelErrorBoundary title="Custom Panel Title">
        <ThrowingComponent />
      </PanelErrorBoundary>
    )

    expect(screen.getByText('Custom Panel Title')).toBeInTheDocument()
    expect(screen.getByText('This panel encountered an error.')).toBeInTheDocument()
  })

  it('uses default title when not provided', () => {
    render(
      <PanelErrorBoundary>
        <ThrowingComponent />
      </PanelErrorBoundary>
    )

    expect(screen.getByText('Panel Error')).toBeInTheDocument()
  })
})

describe('NodeErrorBoundary', () => {
  it('renders children when there is no error', () => {
    render(
      <NodeErrorBoundary>
        <div>Node content</div>
      </NodeErrorBoundary>
    )

    expect(screen.getByText('Node content')).toBeInTheDocument()
  })

  it('renders minimal node fallback when error occurs', () => {
    render(
      <NodeErrorBoundary nodeId="node-123">
        <ThrowingComponent />
      </NodeErrorBoundary>
    )

    expect(screen.getByText('Node Error')).toBeInTheDocument()
    expect(screen.getByText('Retry')).toBeInTheDocument()
  })

  it('allows retry for node errors', () => {
    render(
      <NodeErrorBoundary nodeId="node-123">
        <ThrowingComponent />
      </NodeErrorBoundary>
    )

    expect(screen.getByText('Node Error')).toBeInTheDocument()

    fireEvent.click(screen.getByText('Retry'))

    // Component still throws, so error still shown
    expect(screen.getByText('Node Error')).toBeInTheDocument()
  })
})

describe('Error boundary reset functionality', () => {
  it('WorkflowErrorBoundary resets on button click', () => {
    const onReset = vi.fn()

    render(
      <WorkflowErrorBoundary onReset={onReset}>
        <ThrowingComponent />
      </WorkflowErrorBoundary>
    )

    expect(screen.getByText('Component Error')).toBeInTheDocument()

    fireEvent.click(screen.getByText('Reset'))

    expect(onReset).toHaveBeenCalledTimes(1)
  })

  it('CanvasErrorBoundary resets on button click', () => {
    const onReset = vi.fn()

    render(
      <CanvasErrorBoundary onReset={onReset}>
        <ThrowingComponent />
      </CanvasErrorBoundary>
    )

    expect(screen.getByText('Canvas Error')).toBeInTheDocument()

    fireEvent.click(screen.getByText('Reset Canvas'))

    expect(onReset).toHaveBeenCalledTimes(1)
  })

  it('PanelErrorBoundary resets on button click', () => {
    const onReset = vi.fn()

    render(
      <PanelErrorBoundary onReset={onReset}>
        <ThrowingComponent />
      </PanelErrorBoundary>
    )

    expect(screen.getByText('Panel Error')).toBeInTheDocument()

    fireEvent.click(screen.getByText('Reset'))

    expect(onReset).toHaveBeenCalledTimes(1)
  })
})
