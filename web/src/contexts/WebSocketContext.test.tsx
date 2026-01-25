/**
 * WebSocketContext Tests
 * TDD: Tests written FIRST to define expected behavior
 */

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook } from '@testing-library/react'
import type { ReactNode } from 'react'
import {
  WebSocketProvider,
  useWebSocketContext,
  useOptionalWebSocketContext,
  useWebSocketStatus,
} from './WebSocketContext'

// Mock WebSocket client
vi.mock('../lib/websocket', () => ({
  WebSocketClient: vi.fn().mockImplementation(() => ({
    connect: vi.fn(),
    disconnect: vi.fn(),
    on: vi.fn().mockReturnValue(() => {}),
    isConnected: vi.fn().mockReturnValue(false),
  })),
}))

// Mock config
vi.mock('../config/websocket', () => ({
  createTenantWebSocketUrl: vi.fn().mockReturnValue('ws://localhost:8080/api/v1/ws'),
  getWebSocketConfig: vi.fn().mockReturnValue({
    reconnectDelay: 3000,
    maxReconnectAttempts: 10,
    connectionTimeout: 10000,
    heartbeatInterval: 30000,
  }),
}))

describe('useWebSocketContext', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('throws error when used outside provider', () => {
    // Suppress console.error for this test
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    expect(() => {
      renderHook(() => useWebSocketContext())
    }).toThrow('useWebSocketContext must be used within a WebSocketProvider')

    consoleSpy.mockRestore()
  })

  it('does not throw when used inside provider', () => {
    const wrapper = ({ children }: { children: ReactNode }) => (
      <WebSocketProvider autoConnect={false}>{children}</WebSocketProvider>
    )

    expect(() => {
      renderHook(() => useWebSocketContext(), { wrapper })
    }).not.toThrow()
  })
})

describe('useOptionalWebSocketContext', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns null when used outside provider', () => {
    const { result } = renderHook(() => useOptionalWebSocketContext())

    expect(result.current).toBeNull()
  })

  it('returns context when used inside provider', () => {
    const wrapper = ({ children }: { children: ReactNode }) => (
      <WebSocketProvider autoConnect={false}>{children}</WebSocketProvider>
    )

    const { result } = renderHook(() => useOptionalWebSocketContext(), { wrapper })

    expect(result.current).not.toBeNull()
    expect(result.current).toHaveProperty('connected')
  })
})

describe('useWebSocketStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns connection status fields', () => {
    const wrapper = ({ children }: { children: ReactNode }) => (
      <WebSocketProvider autoConnect={false}>{children}</WebSocketProvider>
    )

    const { result } = renderHook(() => useWebSocketStatus(), { wrapper })

    expect(result.current).toHaveProperty('connected')
    expect(result.current).toHaveProperty('reconnecting')
    expect(result.current).toHaveProperty('reconnectAttempt')
    expect(result.current).toHaveProperty('lastError')
  })

  it('does not include subscribe or reconnect functions', () => {
    const wrapper = ({ children }: { children: ReactNode }) => (
      <WebSocketProvider autoConnect={false}>{children}</WebSocketProvider>
    )

    const { result } = renderHook(() => useWebSocketStatus(), { wrapper })

    expect(result.current).not.toHaveProperty('subscribe')
    expect(result.current).not.toHaveProperty('reconnect')
  })
})

describe('WebSocketProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('provides initial state values', () => {
    const wrapper = ({ children }: { children: ReactNode }) => (
      <WebSocketProvider autoConnect={false}>{children}</WebSocketProvider>
    )

    const { result } = renderHook(() => useWebSocketContext(), { wrapper })

    expect(result.current.connected).toBe(false)
    expect(result.current.reconnecting).toBe(false)
    expect(result.current.reconnectAttempt).toBe(0)
    expect(result.current.lastError).toBeNull()
    expect(result.current.maxReconnectAttempts).toBe(10)
    expect(typeof result.current.subscribe).toBe('function')
    expect(typeof result.current.reconnect).toBe('function')
    expect(typeof result.current.disconnect).toBe('function')
  })

  it('subscribe returns a function', () => {
    const wrapper = ({ children }: { children: ReactNode }) => (
      <WebSocketProvider autoConnect={false}>{children}</WebSocketProvider>
    )

    const { result } = renderHook(() => useWebSocketContext(), { wrapper })

    const handler = vi.fn()
    const unsubscribe = result.current.subscribe(handler)

    expect(typeof unsubscribe).toBe('function')
  })
})
