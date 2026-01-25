import { describe, it, expect } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { ReactFlowProvider } from '@xyflow/react'
import RetryNode from './RetryNode'
import type { RetryNodeData } from './RetryNode'

// Wrapper component for ReactFlow context
const Wrapper = ({ children }: { children: React.ReactNode }) => (
  <ReactFlowProvider>{children}</ReactFlowProvider>
)

describe('RetryNode', () => {
  const defaultData: RetryNodeData = {
    label: 'Retry HTTP',
    config: {
      strategy: 'exponential',
      maxAttempts: 3,
      initialDelayMs: 1000,
      maxDelayMs: 30000,
      multiplier: 2.0,
    },
  }

  it('renders with label and strategy', () => {
    render(<RetryNode id="retry-1" data={defaultData} />, { wrapper: Wrapper })

    expect(screen.getByText('Retry HTTP')).toBeInTheDocument()
    expect(screen.getByText('Retry: Exponential')).toBeInTheDocument()
  })

  it('displays correct icon for fixed strategy', () => {
    const data: RetryNodeData = {
      ...defaultData,
      config: { ...defaultData.config, strategy: 'fixed' },
    }

    render(<RetryNode id="retry-1" data={data} />, { wrapper: Wrapper })

    expect(screen.getByText('⏱️')).toBeInTheDocument()
    expect(screen.getByText('Retry: Fixed Delay')).toBeInTheDocument()
  })

  it('displays correct icon for exponential strategy', () => {
    render(<RetryNode id="retry-1" data={defaultData} />, { wrapper: Wrapper })

    expect(screen.getByText('📈')).toBeInTheDocument()
  })

  it('displays correct icon for exponential_jitter strategy', () => {
    const data: RetryNodeData = {
      ...defaultData,
      config: { ...defaultData.config, strategy: 'exponential_jitter' },
    }

    render(<RetryNode id="retry-1" data={data} />, { wrapper: Wrapper })

    expect(screen.getByText('🎲')).toBeInTheDocument()
    expect(screen.getByText('Retry: Exponential + Jitter')).toBeInTheDocument()
  })

  it('displays max attempts and initial delay', () => {
    render(<RetryNode id="retry-1" data={defaultData} />, { wrapper: Wrapper })

    expect(screen.getByText('Max Attempts: 3')).toBeInTheDocument()
    expect(screen.getByText('Delay: 1.0s')).toBeInTheDocument()
  })

  it('formats delay in milliseconds', () => {
    const data: RetryNodeData = {
      ...defaultData,
      config: { ...defaultData.config, initialDelayMs: 500 },
    }

    render(<RetryNode id="retry-1" data={data} />, { wrapper: Wrapper })

    expect(screen.getByText('Delay: 500ms')).toBeInTheDocument()
  })

  it('formats delay in seconds', () => {
    const data: RetryNodeData = {
      ...defaultData,
      config: { ...defaultData.config, initialDelayMs: 5000 },
    }

    render(<RetryNode id="retry-1" data={data} />, { wrapper: Wrapper })

    expect(screen.getByText('Delay: 5.0s')).toBeInTheDocument()
  })

  it('formats delay in minutes', () => {
    const data: RetryNodeData = {
      ...defaultData,
      config: { ...defaultData.config, initialDelayMs: 120000 },
    }

    render(<RetryNode id="retry-1" data={data} />, { wrapper: Wrapper })

    expect(screen.getByText('Delay: 2.0m')).toBeInTheDocument()
  })

  it('expands and collapses on button click', () => {
    render(<RetryNode id="retry-1" data={defaultData} />, { wrapper: Wrapper })

    // Initially collapsed
    expect(screen.queryByText(/Max Delay:/)).not.toBeInTheDocument()

    // Expand
    const expandButton = screen.getByText('▶')
    fireEvent.click(expandButton)

    // Should show config
    expect(screen.getByText(/Max Delay:/)).toBeInTheDocument()

    // Collapse
    const collapseButton = screen.getByText('▼')
    fireEvent.click(collapseButton)

    // Should hide config
    expect(screen.queryByText(/Max Delay:/)).not.toBeInTheDocument()
  })

  it('displays max delay when configured', () => {
    render(<RetryNode id="retry-1" data={defaultData} />, { wrapper: Wrapper })

    fireEvent.click(screen.getByText('▶'))

    // Text is split across elements: <strong>Max Delay:</strong> 30.0s
    const maxDelayLabel = screen.getByText(/Max Delay:/)
    expect(maxDelayLabel.closest('div')).toHaveTextContent('30.0s')
  })

  it('displays multiplier when configured', () => {
    render(<RetryNode id="retry-1" data={defaultData} />, { wrapper: Wrapper })

    fireEvent.click(screen.getByText('▶'))

    // Text is split across elements: <strong>Multiplier:</strong> 2x
    const multiplierLabel = screen.getByText(/Multiplier:/)
    expect(multiplierLabel.closest('div')).toHaveTextContent('2x')
  })

  it('displays jitter when enabled', () => {
    const data: RetryNodeData = {
      ...defaultData,
      config: { ...defaultData.config, jitter: true },
    }

    render(<RetryNode id="retry-1" data={data} />, { wrapper: Wrapper })

    fireEvent.click(screen.getByText('▶'))

    // Text is split across elements: <strong>Jitter:</strong> Enabled
    const jitterLabel = screen.getByText(/Jitter:/)
    expect(jitterLabel.closest('div')).toHaveTextContent('Enabled')
  })

  it('displays retryable errors when configured', () => {
    const data: RetryNodeData = {
      ...defaultData,
      config: {
        ...defaultData.config,
        retryableErrors: ['timeout', 'connection'],
      },
    }

    render(<RetryNode id="retry-1" data={data} />, { wrapper: Wrapper })

    fireEvent.click(screen.getByText('▶'))

    // Text is split across elements: <strong>Retry On:</strong> timeout, connection
    const retryOnLabel = screen.getByText(/Retry On:/)
    expect(retryOnLabel.closest('div')).toHaveTextContent('timeout, connection')
  })

  it('displays non-retryable errors when configured', () => {
    const data: RetryNodeData = {
      ...defaultData,
      config: {
        ...defaultData.config,
        nonRetryableErrors: ['invalid', 'forbidden'],
      },
    }

    render(<RetryNode id="retry-1" data={data} />, { wrapper: Wrapper })

    fireEvent.click(screen.getByText('▶'))

    // Text is split across elements: <strong>Don't Retry:</strong> invalid, forbidden
    const dontRetryLabel = screen.getByText(/Don't Retry:/)
    expect(dontRetryLabel.closest('div')).toHaveTextContent('invalid, forbidden')
  })

  it('displays retryable status codes when configured', () => {
    const data: RetryNodeData = {
      ...defaultData,
      config: {
        ...defaultData.config,
        retryableStatusCodes: [408, 429, 500, 502, 503, 504],
      },
    }

    render(<RetryNode id="retry-1" data={data} />, { wrapper: Wrapper })

    fireEvent.click(screen.getByText('▶'))

    // Text is split across elements: <strong>Status Codes:</strong> 408, 429, 500, 502, 503, 504
    const statusCodesLabel = screen.getByText(/Status Codes:/)
    expect(statusCodesLabel.closest('div')).toHaveTextContent('408, 429, 500, 502, 503, 504')
  })

  it('shows success and failed branch indicators', () => {
    render(<RetryNode id="retry-1" data={defaultData} />, { wrapper: Wrapper })

    expect(screen.getByText('Success')).toBeInTheDocument()
    expect(screen.getByText('Failed')).toBeInTheDocument()
  })

  it('applies selected styles when selected', () => {
    const { container } = render(
      <RetryNode id="retry-1" data={defaultData} selected={true} />,
      { wrapper: Wrapper }
    )

    const nodeElement = container.querySelector('.ring-2')
    expect(nodeElement).toBeInTheDocument()
  })

  it('does not apply selected styles when not selected', () => {
    const { container } = render(
      <RetryNode id="retry-1" data={defaultData} selected={false} />,
      { wrapper: Wrapper }
    )

    const nodeElement = container.querySelector('.ring-2')
    expect(nodeElement).not.toBeInTheDocument()
  })

  it('handles minimal configuration', () => {
    const minimalData: RetryNodeData = {
      label: 'Simple Retry',
      config: {
        strategy: 'fixed',
        maxAttempts: 1,
        initialDelayMs: 100,
      },
    }

    render(<RetryNode id="retry-1" data={minimalData} />, { wrapper: Wrapper })

    expect(screen.getByText('Simple Retry')).toBeInTheDocument()
    expect(screen.getByText('Retry: Fixed Delay')).toBeInTheDocument()
    expect(screen.getByText('Max Attempts: 1')).toBeInTheDocument()
    expect(screen.getByText('Delay: 100ms')).toBeInTheDocument()
  })

  it('handles all strategy types', () => {
    const strategies: Array<{ strategy: RetryNodeData['config']['strategy']; label: string }> = [
      { strategy: 'fixed', label: 'Fixed Delay' },
      { strategy: 'exponential', label: 'Exponential' },
      { strategy: 'exponential_jitter', label: 'Exponential + Jitter' },
    ]

    strategies.forEach(({ strategy, label }) => {
      const data: RetryNodeData = {
        label: `Test ${strategy}`,
        config: {
          strategy,
          maxAttempts: 3,
          initialDelayMs: 1000,
        },
      }

      const { unmount } = render(<RetryNode id="retry-1" data={data} />, { wrapper: Wrapper })

      expect(screen.getByText(`Retry: ${label}`)).toBeInTheDocument()

      unmount()
    })
  })
})
