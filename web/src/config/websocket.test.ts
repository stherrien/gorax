/**
 * WebSocket Configuration Tests
 * TDD: Tests written FIRST to define expected behavior
 */

import { describe, it, expect } from 'vitest'
import {
  getApiBaseUrl,
  getWebSocketBaseUrl,
  getWebSocketConfig,
  createExecutionWebSocketUrl,
  createWorkflowWebSocketUrl,
  createTenantWebSocketUrl,
  calculateReconnectDelay,
  addJitter,
  DEFAULT_WEBSOCKET_CONFIG,
} from './websocket'

describe('WebSocket Configuration', () => {
  describe('getApiBaseUrl', () => {
    it('returns a URL string', () => {
      const result = getApiBaseUrl()
      expect(typeof result).toBe('string')
      expect(result.startsWith('http')).toBe(true)
    })
  })

  describe('getWebSocketBaseUrl', () => {
    it('returns a WebSocket URL string', () => {
      const result = getWebSocketBaseUrl()
      expect(typeof result).toBe('string')
      expect(result.startsWith('ws')).toBe(true)
    })
  })

  describe('getWebSocketConfig', () => {
    it('returns merged config with defaults', () => {
      const config = getWebSocketConfig()

      expect(config.reconnectDelay).toBeDefined()
      expect(config.maxReconnectAttempts).toBeDefined()
      expect(config.connectionTimeout).toBeDefined()
      expect(config.heartbeatInterval).toBeDefined()
    })

    it('returns numeric values for all config properties', () => {
      const config = getWebSocketConfig()

      expect(typeof config.reconnectDelay).toBe('number')
      expect(typeof config.maxReconnectAttempts).toBe('number')
      expect(typeof config.connectionTimeout).toBe('number')
      expect(typeof config.heartbeatInterval).toBe('number')
    })
  })

  describe('URL Builders', () => {
    describe('createExecutionWebSocketUrl', () => {
      it('uses provided baseURL when given with http', () => {
        const url = createExecutionWebSocketUrl('exec-123', 'http://api.example.com')

        expect(url).toBe('ws://api.example.com/api/v1/ws/executions/exec-123')
      })

      it('uses provided baseURL when given with https', () => {
        const url = createExecutionWebSocketUrl('exec-123', 'https://api.example.com')

        expect(url).toBe('wss://api.example.com/api/v1/ws/executions/exec-123')
      })

      it('includes execution ID in URL path', () => {
        const url = createExecutionWebSocketUrl('my-execution-id', 'http://localhost:8080')

        expect(url).toContain('my-execution-id')
        expect(url).toContain('/executions/')
      })
    })

    describe('createWorkflowWebSocketUrl', () => {
      it('uses provided baseURL when given with http', () => {
        const url = createWorkflowWebSocketUrl('workflow-456', 'http://api.example.com')

        expect(url).toBe('ws://api.example.com/api/v1/ws/workflows/workflow-456')
      })

      it('uses provided baseURL when given with https', () => {
        const url = createWorkflowWebSocketUrl('workflow-456', 'https://api.example.com')

        expect(url).toBe('wss://api.example.com/api/v1/ws/workflows/workflow-456')
      })

      it('includes workflow ID in URL path', () => {
        const url = createWorkflowWebSocketUrl('my-workflow-id', 'http://localhost:8080')

        expect(url).toContain('my-workflow-id')
        expect(url).toContain('/workflows/')
      })
    })

    describe('createTenantWebSocketUrl', () => {
      it('uses provided baseURL when given with http', () => {
        const url = createTenantWebSocketUrl('http://api.example.com')

        expect(url).toBe('ws://api.example.com/api/v1/ws?subscribe_tenant=true')
      })

      it('uses provided baseURL when given with https', () => {
        const url = createTenantWebSocketUrl('https://api.example.com')

        expect(url).toBe('wss://api.example.com/api/v1/ws?subscribe_tenant=true')
      })

      it('includes subscribe_tenant query parameter', () => {
        const url = createTenantWebSocketUrl('http://localhost:8080')

        expect(url).toContain('subscribe_tenant=true')
      })
    })
  })

  describe('Reconnection Strategy', () => {
    describe('calculateReconnectDelay', () => {
      it('calculates exponential backoff correctly', () => {
        const baseDelay = 1000

        expect(calculateReconnectDelay(1, baseDelay)).toBe(1000) // 1000 * 2^0
        expect(calculateReconnectDelay(2, baseDelay)).toBe(2000) // 1000 * 2^1
        expect(calculateReconnectDelay(3, baseDelay)).toBe(4000) // 1000 * 2^2
        expect(calculateReconnectDelay(4, baseDelay)).toBe(8000) // 1000 * 2^3
      })

      it('caps delay at maxDelay', () => {
        const baseDelay = 1000
        const maxDelay = 5000

        // 1000 * 2^4 = 16000, but capped at 5000
        expect(calculateReconnectDelay(5, baseDelay, maxDelay)).toBe(5000)
      })

      it('uses default values when not provided', () => {
        const delay = calculateReconnectDelay(1)

        expect(delay).toBe(DEFAULT_WEBSOCKET_CONFIG.reconnectDelay)
      })

      it('uses default maxDelay of 30 seconds', () => {
        // Large attempt number to hit the cap
        const delay = calculateReconnectDelay(10)

        expect(delay).toBeLessThanOrEqual(30000)
      })
    })

    describe('addJitter', () => {
      it('returns a value close to the input', () => {
        const delay = 1000
        const result = addJitter(delay)

        // With 20% jitter, result should be between 800 and 1200
        expect(result).toBeGreaterThanOrEqual(800)
        expect(result).toBeLessThanOrEqual(1200)
      })

      it('respects custom jitter factor', () => {
        const delay = 1000
        const jitterFactor = 0.5

        const result = addJitter(delay, jitterFactor)

        // With 50% jitter, result should be between 500 and 1500
        expect(result).toBeGreaterThanOrEqual(500)
        expect(result).toBeLessThanOrEqual(1500)
      })

      it('never returns negative values', () => {
        const delay = 100
        const jitterFactor = 0.2

        // Run multiple times to check for negative values
        for (let i = 0; i < 100; i++) {
          expect(addJitter(delay, jitterFactor)).toBeGreaterThanOrEqual(0)
        }
      })

      it('returns exactly the delay when jitter factor is 0', () => {
        const delay = 1000
        const result = addJitter(delay, 0)

        expect(result).toBe(1000)
      })
    })
  })

  describe('Default Configuration', () => {
    it('has reasonable default values', () => {
      expect(DEFAULT_WEBSOCKET_CONFIG.reconnectDelay).toBe(3000)
      expect(DEFAULT_WEBSOCKET_CONFIG.maxReconnectAttempts).toBe(10)
      expect(DEFAULT_WEBSOCKET_CONFIG.connectionTimeout).toBe(10000)
      expect(DEFAULT_WEBSOCKET_CONFIG.heartbeatInterval).toBe(30000)
    })
  })
})
