/**
 * WebSocket configuration
 *
 * Centralized configuration for WebSocket connections including
 * reconnection parameters, timeouts, and environment-specific settings.
 */

import type { WebSocketConfig } from '../types/websocket'

// ============================================================================
// Environment Configuration
// ============================================================================

/**
 * Get the API base URL from environment or use default
 */
export function getApiBaseUrl(): string {
  return import.meta.env.VITE_API_URL || 'http://localhost:8080'
}

/**
 * Get the WebSocket base URL derived from API URL
 */
export function getWebSocketBaseUrl(): string {
  const apiUrl = getApiBaseUrl()
  const wsProtocol = apiUrl.startsWith('https') ? 'wss' : 'ws'
  const host = apiUrl.replace(/^https?:\/\//, '')
  return `${wsProtocol}://${host}`
}

// ============================================================================
// Default Configuration
// ============================================================================

/**
 * Default WebSocket configuration values
 */
export const DEFAULT_WEBSOCKET_CONFIG: Required<Omit<WebSocketConfig, 'url'>> = {
  /** Initial delay before reconnection (3 seconds) */
  reconnectDelay: 3000,

  /** Maximum reconnection attempts before giving up */
  maxReconnectAttempts: 10,

  /** Connection timeout in milliseconds (10 seconds) */
  connectionTimeout: 10000,

  /** Heartbeat/ping interval in milliseconds (30 seconds) */
  heartbeatInterval: 30000,
}

/**
 * Development-specific configuration overrides
 */
export const DEV_WEBSOCKET_CONFIG: Partial<typeof DEFAULT_WEBSOCKET_CONFIG> = {
  /** Faster reconnection in development */
  reconnectDelay: 1000,

  /** More attempts in development */
  maxReconnectAttempts: 20,
}

/**
 * Production-specific configuration overrides
 */
export const PROD_WEBSOCKET_CONFIG: Partial<typeof DEFAULT_WEBSOCKET_CONFIG> = {
  /** Standard reconnection delay */
  reconnectDelay: 3000,

  /** Standard attempts */
  maxReconnectAttempts: 10,
}

/**
 * Get environment-specific WebSocket configuration
 */
export function getWebSocketConfig(): Required<Omit<WebSocketConfig, 'url'>> {
  const isDev = import.meta.env.DEV
  const envConfig = isDev ? DEV_WEBSOCKET_CONFIG : PROD_WEBSOCKET_CONFIG

  return {
    ...DEFAULT_WEBSOCKET_CONFIG,
    ...envConfig,
  }
}

// ============================================================================
// URL Builders
// ============================================================================

/**
 * Create WebSocket URL for execution updates
 * @param executionId - The execution ID to subscribe to
 * @param baseURL - Optional base URL override
 */
export function createExecutionWebSocketUrl(
  executionId: string,
  baseURL?: string
): string {
  const wsBase = baseURL
    ? baseURL.replace(/^https?:\/\//, (m) => (m === 'https://' ? 'wss://' : 'ws://'))
    : getWebSocketBaseUrl()
  return `${wsBase}/api/v1/ws/executions/${executionId}`
}

/**
 * Create WebSocket URL for workflow collaboration
 * @param workflowId - The workflow ID to subscribe to
 * @param baseURL - Optional base URL override
 */
export function createWorkflowWebSocketUrl(
  workflowId: string,
  baseURL?: string
): string {
  const wsBase = baseURL
    ? baseURL.replace(/^https?:\/\//, (m) => (m === 'https://' ? 'wss://' : 'ws://'))
    : getWebSocketBaseUrl()
  return `${wsBase}/api/v1/ws/workflows/${workflowId}`
}

/**
 * Create WebSocket URL for tenant-wide updates
 * @param baseURL - Optional base URL override
 */
export function createTenantWebSocketUrl(baseURL?: string): string {
  const wsBase = baseURL
    ? baseURL.replace(/^https?:\/\//, (m) => (m === 'https://' ? 'wss://' : 'ws://'))
    : getWebSocketBaseUrl()
  return `${wsBase}/api/v1/ws?subscribe_tenant=true`
}

// ============================================================================
// Reconnection Strategy
// ============================================================================

/**
 * Calculate exponential backoff delay for reconnection
 * @param attempt - Current attempt number (1-based)
 * @param baseDelay - Base delay in milliseconds
 * @param maxDelay - Maximum delay cap in milliseconds
 */
export function calculateReconnectDelay(
  attempt: number,
  baseDelay: number = DEFAULT_WEBSOCKET_CONFIG.reconnectDelay,
  maxDelay: number = 30000
): number {
  // Exponential backoff: baseDelay * 2^(attempt-1)
  // Capped at maxDelay
  const exponentialDelay = baseDelay * Math.pow(2, attempt - 1)
  return Math.min(exponentialDelay, maxDelay)
}

/**
 * Calculate delay with jitter to prevent thundering herd
 * @param delay - Base delay in milliseconds
 * @param jitterFactor - Jitter factor (0-1), default 0.2 (20%)
 */
export function addJitter(delay: number, jitterFactor: number = 0.2): number {
  const jitter = delay * jitterFactor * (Math.random() * 2 - 1)
  return Math.max(0, delay + jitter)
}

// ============================================================================
// Message Buffer Configuration
// ============================================================================

/**
 * Configuration for message buffering during disconnection
 */
export const MESSAGE_BUFFER_CONFIG = {
  /** Maximum number of messages to buffer */
  maxBufferSize: 100,

  /** Maximum age of buffered messages in milliseconds (5 minutes) */
  maxMessageAge: 5 * 60 * 1000,

  /** Whether to flush buffer on reconnection */
  flushOnReconnect: true,
}

// ============================================================================
// Throttling Configuration
// ============================================================================

/**
 * Configuration for update throttling to prevent UI flooding
 */
export const THROTTLE_CONFIG = {
  /** Minimum interval between UI updates in milliseconds */
  minUpdateInterval: 100,

  /** Maximum updates per second */
  maxUpdatesPerSecond: 10,

  /** Whether to batch rapid updates */
  batchUpdates: true,
}
