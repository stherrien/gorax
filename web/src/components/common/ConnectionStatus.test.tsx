/**
 * ConnectionStatus Tests
 * TDD: Tests written FIRST to define expected behavior
 */

import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ConnectionStatus, ConnectionStatusDot } from './ConnectionStatus'

describe('ConnectionStatus', () => {
  describe('Connected State', () => {
    it('renders connected status with green indicator', () => {
      render(<ConnectionStatus connected={true} />)

      const dot = screen.getByTestId('connection-status-dot')
      expect(dot).toHaveClass('bg-green-400')
      expect(dot).toHaveClass('animate-pulse')
    })

    it('displays "Connected" label when connected', () => {
      render(<ConnectionStatus connected={true} />)

      expect(screen.getByText('Connected')).toBeInTheDocument()
    })

    it('does not show reconnect button when connected', () => {
      render(<ConnectionStatus connected={true} onReconnect={vi.fn()} />)

      expect(screen.queryByTestId('reconnect-button')).not.toBeInTheDocument()
    })
  })

  describe('Disconnected State', () => {
    it('renders disconnected status with red indicator', () => {
      render(<ConnectionStatus connected={false} />)

      const dot = screen.getByTestId('connection-status-dot')
      expect(dot).toHaveClass('bg-red-400')
      expect(dot).not.toHaveClass('animate-pulse')
    })

    it('displays "Disconnected" label when disconnected', () => {
      render(<ConnectionStatus connected={false} />)

      expect(screen.getByText('Disconnected')).toBeInTheDocument()
    })

    it('shows reconnect button when disconnected and onReconnect provided', () => {
      render(<ConnectionStatus connected={false} onReconnect={vi.fn()} />)

      expect(screen.getByTestId('reconnect-button')).toBeInTheDocument()
    })

    it('does not show reconnect button when showReconnectButton is false', () => {
      render(
        <ConnectionStatus
          connected={false}
          onReconnect={vi.fn()}
          showReconnectButton={false}
        />
      )

      expect(screen.queryByTestId('reconnect-button')).not.toBeInTheDocument()
    })
  })

  describe('Reconnecting State', () => {
    it('renders reconnecting status with yellow indicator', () => {
      render(<ConnectionStatus connected={false} reconnecting={true} />)

      const dot = screen.getByTestId('connection-status-dot')
      expect(dot).toHaveClass('bg-yellow-400')
      expect(dot).toHaveClass('animate-pulse')
    })

    it('displays reconnecting label with attempt number', () => {
      render(
        <ConnectionStatus
          connected={false}
          reconnecting={true}
          reconnectAttempt={3}
        />
      )

      expect(screen.getByText('Reconnecting... (Attempt 3)')).toBeInTheDocument()
    })

    it('does not show reconnect button while reconnecting', () => {
      render(
        <ConnectionStatus
          connected={false}
          reconnecting={true}
          onReconnect={vi.fn()}
        />
      )

      expect(screen.queryByTestId('reconnect-button')).not.toBeInTheDocument()
    })
  })

  describe('Reconnect Button', () => {
    it('calls onReconnect when reconnect button is clicked', async () => {
      const onReconnect = vi.fn()
      const user = userEvent.setup()

      render(<ConnectionStatus connected={false} onReconnect={onReconnect} />)

      await user.click(screen.getByTestId('reconnect-button'))

      expect(onReconnect).toHaveBeenCalledTimes(1)
    })

    it('disables button temporarily after click to prevent spam', async () => {
      const onReconnect = vi.fn()
      const user = userEvent.setup()

      render(<ConnectionStatus connected={false} onReconnect={onReconnect} />)

      const button = screen.getByTestId('reconnect-button')
      await user.click(button)

      expect(button).toBeDisabled()
    })

    it('button is initially not disabled', () => {
      const onReconnect = vi.fn()

      render(<ConnectionStatus connected={false} onReconnect={onReconnect} />)

      const button = screen.getByTestId('reconnect-button')
      expect(button).not.toBeDisabled()
    })
  })

  describe('Max Reconnect Attempts', () => {
    it('shows warning when max attempts exceeded', () => {
      render(
        <ConnectionStatus
          connected={false}
          reconnectAttempt={10}
          maxReconnectAttempts={10}
        />
      )

      expect(screen.getByTestId('max-attempts-warning')).toBeInTheDocument()
      expect(screen.getByText('(Max attempts reached)')).toBeInTheDocument()
    })

    it('does not show warning when still under max attempts', () => {
      render(
        <ConnectionStatus
          connected={false}
          reconnectAttempt={5}
          maxReconnectAttempts={10}
        />
      )

      expect(screen.queryByTestId('max-attempts-warning')).not.toBeInTheDocument()
    })
  })

  describe('Size Variants', () => {
    it('renders small size variant', () => {
      render(<ConnectionStatus connected={true} size="sm" />)

      const dot = screen.getByTestId('connection-status-dot')
      expect(dot).toHaveClass('w-1.5', 'h-1.5')

      const label = screen.getByTestId('connection-status-label')
      expect(label).toHaveClass('text-xs')
    })

    it('renders medium size variant (default)', () => {
      render(<ConnectionStatus connected={true} />)

      const dot = screen.getByTestId('connection-status-dot')
      expect(dot).toHaveClass('w-2', 'h-2')

      const label = screen.getByTestId('connection-status-label')
      expect(label).toHaveClass('text-sm')
    })

    it('renders large size variant', () => {
      render(<ConnectionStatus connected={true} size="lg" />)

      const dot = screen.getByTestId('connection-status-dot')
      expect(dot).toHaveClass('w-3', 'h-3')

      const label = screen.getByTestId('connection-status-label')
      expect(label).toHaveClass('text-base')
    })
  })

  describe('Show Label Option', () => {
    it('hides label when showLabel is false', () => {
      render(<ConnectionStatus connected={true} showLabel={false} />)

      expect(screen.queryByTestId('connection-status-label')).not.toBeInTheDocument()
    })

    it('shows label by default', () => {
      render(<ConnectionStatus connected={true} />)

      expect(screen.getByTestId('connection-status-label')).toBeInTheDocument()
    })
  })

  describe('Accessibility', () => {
    it('has proper role attribute', () => {
      render(<ConnectionStatus connected={true} />)

      const status = screen.getByTestId('connection-status')
      expect(status).toHaveAttribute('role', 'status')
    })

    it('has aria-live for announcements', () => {
      render(<ConnectionStatus connected={true} />)

      const status = screen.getByTestId('connection-status')
      expect(status).toHaveAttribute('aria-live', 'polite')
    })

    it('has aria-label describing current state', () => {
      render(<ConnectionStatus connected={true} />)

      const status = screen.getByTestId('connection-status')
      expect(status).toHaveAttribute('aria-label', 'Connection status: Connected')
    })

    it('reconnect button has accessible label', () => {
      render(<ConnectionStatus connected={false} onReconnect={vi.fn()} />)

      const button = screen.getByTestId('reconnect-button')
      expect(button).toHaveAttribute('aria-label', 'Reconnect to server')
    })
  })

  describe('Custom className', () => {
    it('applies custom className', () => {
      render(<ConnectionStatus connected={true} className="custom-class" />)

      const status = screen.getByTestId('connection-status')
      expect(status).toHaveClass('custom-class')
    })
  })
})

describe('ConnectionStatusDot', () => {
  it('renders just the dot without label', () => {
    render(<ConnectionStatusDot connected={true} />)

    expect(screen.getByTestId('connection-status-dot')).toBeInTheDocument()
    expect(screen.queryByText('Connected')).not.toBeInTheDocument()
  })

  it('shows correct color for connected state', () => {
    render(<ConnectionStatusDot connected={true} />)

    const dot = screen.getByTestId('connection-status-dot')
    expect(dot).toHaveClass('bg-green-400')
  })

  it('shows correct color for disconnected state', () => {
    render(<ConnectionStatusDot connected={false} />)

    const dot = screen.getByTestId('connection-status-dot')
    expect(dot).toHaveClass('bg-red-400')
  })

  it('shows correct color for reconnecting state', () => {
    render(<ConnectionStatusDot connected={false} reconnecting={true} />)

    const dot = screen.getByTestId('connection-status-dot')
    expect(dot).toHaveClass('bg-yellow-400')
  })

  it('has tooltip with status', () => {
    render(<ConnectionStatusDot connected={true} />)

    const dot = screen.getByTestId('connection-status-dot')
    expect(dot).toHaveAttribute('title', 'Connected')
  })

  it('has proper accessibility attributes', () => {
    render(<ConnectionStatusDot connected={true} />)

    const dot = screen.getByTestId('connection-status-dot')
    expect(dot).toHaveAttribute('role', 'status')
    expect(dot).toHaveAttribute('aria-label', 'Connection status: Connected')
  })

  it('applies custom className', () => {
    render(<ConnectionStatusDot connected={true} className="custom-dot" />)

    const dot = screen.getByTestId('connection-status-dot')
    expect(dot).toHaveClass('custom-dot')
  })
})
