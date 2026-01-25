/**
 * ConnectionStatus - Displays WebSocket connection status with visual indicators
 *
 * A reusable component that shows the current connection state with:
 * - Color-coded status indicators (green/yellow/red)
 * - Animated pulse for active states
 * - Reconnection attempt counter
 * - Manual reconnect button for failed connections
 *
 * @example
 * ```tsx
 * <ConnectionStatus
 *   connected={connected}
 *   reconnecting={reconnecting}
 *   reconnectAttempt={reconnectAttempt}
 *   onReconnect={handleReconnect}
 * />
 * ```
 */

import { useCallback, useState, useEffect } from 'react'
import type { ConnectionState } from '../../types/websocket'

export interface ConnectionStatusProps {
  /** Whether the WebSocket is connected */
  connected: boolean
  /** Whether the WebSocket is attempting to reconnect */
  reconnecting?: boolean
  /** Current reconnection attempt number */
  reconnectAttempt?: number
  /** Maximum reconnection attempts */
  maxReconnectAttempts?: number
  /** Callback to manually trigger reconnection */
  onReconnect?: () => void
  /** Size variant */
  size?: 'sm' | 'md' | 'lg'
  /** Whether to show the text label */
  showLabel?: boolean
  /** Whether to show reconnect button when disconnected */
  showReconnectButton?: boolean
  /** Custom class name */
  className?: string
}

/**
 * Derive connection state from props
 */
function deriveConnectionState(
  connected: boolean,
  reconnecting: boolean
): ConnectionState {
  if (connected) return 'connected'
  if (reconnecting) return 'reconnecting'
  return 'disconnected'
}

/**
 * Get status indicator styles based on connection state
 */
function getStatusStyles(state: ConnectionState): {
  dotColor: string
  textColor: string
  animate: boolean
} {
  switch (state) {
    case 'connected':
      return {
        dotColor: 'bg-green-400',
        textColor: 'text-green-400',
        animate: true,
      }
    case 'connecting':
    case 'reconnecting':
      return {
        dotColor: 'bg-yellow-400',
        textColor: 'text-yellow-400',
        animate: true,
      }
    case 'disconnected':
    case 'error':
      return {
        dotColor: 'bg-red-400',
        textColor: 'text-red-400',
        animate: false,
      }
    default:
      return {
        dotColor: 'bg-gray-400',
        textColor: 'text-gray-400',
        animate: false,
      }
  }
}

/**
 * Get label text for connection state
 */
function getStatusLabel(
  state: ConnectionState,
  reconnectAttempt?: number
): string {
  switch (state) {
    case 'connected':
      return 'Connected'
    case 'connecting':
      return 'Connecting...'
    case 'reconnecting':
      return reconnectAttempt
        ? `Reconnecting... (Attempt ${reconnectAttempt})`
        : 'Reconnecting...'
    case 'disconnected':
      return 'Disconnected'
    case 'error':
      return 'Connection Error'
    default:
      return 'Unknown'
  }
}

/**
 * Get dot size based on size prop
 */
function getDotSize(size: ConnectionStatusProps['size']): string {
  switch (size) {
    case 'sm':
      return 'w-1.5 h-1.5'
    case 'lg':
      return 'w-3 h-3'
    case 'md':
    default:
      return 'w-2 h-2'
  }
}

/**
 * Get text size based on size prop
 */
function getTextSize(size: ConnectionStatusProps['size']): string {
  switch (size) {
    case 'sm':
      return 'text-xs'
    case 'lg':
      return 'text-base'
    case 'md':
    default:
      return 'text-sm'
  }
}

export function ConnectionStatus({
  connected,
  reconnecting = false,
  reconnectAttempt = 0,
  maxReconnectAttempts = 10,
  onReconnect,
  size = 'md',
  showLabel = true,
  showReconnectButton = true,
  className = '',
}: ConnectionStatusProps) {
  const [isReconnectDisabled, setIsReconnectDisabled] = useState(false)

  const state = deriveConnectionState(connected, reconnecting)
  const styles = getStatusStyles(state)
  const label = getStatusLabel(state, reconnectAttempt)
  const dotSize = getDotSize(size)
  const textSize = getTextSize(size)

  // Check if max reconnect attempts exceeded
  const maxAttemptsExceeded =
    !connected && !reconnecting && reconnectAttempt >= maxReconnectAttempts

  // Handle manual reconnection
  const handleReconnect = useCallback(() => {
    if (onReconnect && !isReconnectDisabled) {
      setIsReconnectDisabled(true)
      onReconnect()

      // Re-enable button after a short delay
      setTimeout(() => {
        setIsReconnectDisabled(false)
      }, 2000)
    }
  }, [onReconnect, isReconnectDisabled])

  // Reset disabled state when connection changes
  useEffect(() => {
    if (connected) {
      setIsReconnectDisabled(false)
    }
  }, [connected])

  const showButton =
    showReconnectButton &&
    onReconnect &&
    state === 'disconnected' &&
    !reconnecting

  return (
    <div
      className={`flex items-center gap-2 ${className}`}
      data-testid="connection-status"
      role="status"
      aria-live="polite"
      aria-label={`Connection status: ${label}`}
    >
      {/* Status Indicator Dot */}
      <div
        className={`${dotSize} rounded-full ${styles.dotColor} ${
          styles.animate ? 'animate-pulse' : ''
        }`}
        data-testid="connection-status-dot"
        aria-hidden="true"
      />

      {/* Status Label */}
      {showLabel && (
        <span
          className={`${textSize} ${styles.textColor}`}
          data-testid="connection-status-label"
        >
          {label}
        </span>
      )}

      {/* Max Attempts Warning */}
      {maxAttemptsExceeded && showLabel && (
        <span className="text-xs text-red-300" data-testid="max-attempts-warning">
          (Max attempts reached)
        </span>
      )}

      {/* Reconnect Button */}
      {showButton && (
        <button
          type="button"
          onClick={handleReconnect}
          disabled={isReconnectDisabled}
          className={`
            ${textSize} px-2 py-0.5 rounded
            bg-gray-700 hover:bg-gray-600
            text-gray-200 hover:text-white
            transition-colors duration-150
            disabled:opacity-50 disabled:cursor-not-allowed
            focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1 focus:ring-offset-gray-800
          `}
          data-testid="reconnect-button"
          aria-label="Reconnect to server"
        >
          Reconnect
        </button>
      )}
    </div>
  )
}

/**
 * Compact version of ConnectionStatus - just the dot
 */
export function ConnectionStatusDot({
  connected,
  reconnecting = false,
  size = 'md',
  className = '',
}: Pick<ConnectionStatusProps, 'connected' | 'reconnecting' | 'size' | 'className'>) {
  const state = deriveConnectionState(connected, reconnecting)
  const styles = getStatusStyles(state)
  const dotSize = getDotSize(size)
  const label = getStatusLabel(state)

  return (
    <div
      className={`${dotSize} rounded-full ${styles.dotColor} ${
        styles.animate ? 'animate-pulse' : ''
      } ${className}`}
      data-testid="connection-status-dot"
      role="status"
      aria-label={`Connection status: ${label}`}
      title={label}
    />
  )
}
