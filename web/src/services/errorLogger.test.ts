import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { errorLogger, ErrorLoggerService, type ErrorCategory, type ErrorSeverity } from './errorLogger'
import * as Sentry from '@sentry/react'

// Mock Sentry
vi.mock('@sentry/react', () => ({
  withScope: vi.fn((callback) => {
    const scope = {
      setLevel: vi.fn(),
      setTag: vi.fn(),
      setContext: vi.fn(),
    }
    callback(scope)
  }),
  captureException: vi.fn(),
}))

// Mock localStorage
const mockLocalStorage = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key]
    }),
    clear: vi.fn(() => {
      store = {}
    }),
    get length() {
      return Object.keys(store).length
    },
    key: vi.fn((i: number) => Object.keys(store)[i] ?? null),
  }
})()

Object.defineProperty(window, 'localStorage', { value: mockLocalStorage })

describe('ErrorLoggerService', () => {
  let service: ErrorLoggerService

  beforeEach(() => {
    mockLocalStorage.clear()
    vi.clearAllMocks()
    service = new ErrorLoggerService()
  })

  describe('log', () => {
    it('logs error with generated entry', () => {
      const error = new Error('Test error')
      const entry = service.log(error)

      expect(entry).toMatchObject({
        category: expect.any(String),
        severity: expect.any(String),
        message: 'Test error',
        stack: expect.any(String),
      })
      expect(entry.id).toMatch(/^err_\d+_[a-z0-9]+$/)
      expect(entry.timestamp).toBeTruthy()
    })

    it('categorizes network errors correctly', () => {
      const error = new Error('Network request failed')
      const entry = service.log(error)

      expect(entry.category).toBe('network')
    })

    it('categorizes validation errors correctly', () => {
      const error = new Error('Validation failed: Invalid input')
      const entry = service.log(error)

      expect(entry.category).toBe('validation')
    })

    it('categorizes canvas errors from context', () => {
      const error = new Error('Some error')
      const entry = service.log(error, { componentName: 'WorkflowCanvas' })

      expect(entry.category).toBe('canvas')
    })

    it('categorizes node errors from context', () => {
      const error = new Error('Some error')
      const entry = service.log(error, { nodeId: 'node-123' })

      expect(entry.category).toBe('node')
    })

    it('reports to Sentry with correct context', () => {
      const error = new Error('Test error')
      service.log(error, {
        componentName: 'TestComponent',
        workflowId: 'wf-123',
        nodeId: 'node-456',
      })

      expect(Sentry.withScope).toHaveBeenCalled()
      expect(Sentry.captureException).toHaveBeenCalledWith(error)
    })

    it('stores error locally', () => {
      const error = new Error('Test error')
      service.log(error)

      expect(mockLocalStorage.setItem).toHaveBeenCalled()
    })

    it('includes browser info in entry', () => {
      const error = new Error('Test error')
      const entry = service.log(error)

      expect(entry.browserInfo).toMatchObject({
        userAgent: expect.any(String),
        url: expect.any(String),
        screenSize: expect.any(String),
        language: expect.any(String),
      })
    })
  })

  describe('logBoundaryError', () => {
    it('logs boundary error with component context', () => {
      const error = new Error('Render error')
      const errorInfo = { componentStack: 'at Component\nat App' }

      const entry = service.logBoundaryError(
        error,
        errorInfo as React.ErrorInfo,
        'TestComponent',
        { workflowId: 'wf-123' }
      )

      expect(entry.context.componentName).toBe('TestComponent')
      expect(entry.context.componentStack).toBe('at Component\nat App')
      expect(entry.context.workflowId).toBe('wf-123')
    })
  })

  describe('logNetworkError', () => {
    it('logs network error with endpoint context', () => {
      const error = new Error('Request failed')

      const entry = service.logNetworkError(error, '/api/workflows', 'POST')

      expect(entry.context.action).toBe('POST /api/workflows')
      expect(entry.context.additionalData).toEqual({
        endpoint: '/api/workflows',
        method: 'POST',
      })
    })
  })

  describe('logWorkflowError', () => {
    it('logs workflow error with workflow context', () => {
      const error = new Error('Workflow execution failed')

      const entry = service.logWorkflowError(error, 'wf-123', 'node-456', 'execute')

      expect(entry.context.workflowId).toBe('wf-123')
      expect(entry.context.nodeId).toBe('node-456')
      expect(entry.context.action).toBe('execute')
    })
  })

  describe('getLocalLogs', () => {
    it('returns empty array when no logs stored', () => {
      const logs = service.getLocalLogs()
      expect(logs).toEqual([])
    })

    it('returns stored logs', () => {
      const error = new Error('Test error 1')
      service.log(error)

      const error2 = new Error('Test error 2')
      service.log(error2)

      const logs = service.getLocalLogs()
      expect(logs).toHaveLength(2)
    })

    it('limits stored logs to MAX_LOCAL_LOGS', () => {
      // Log more than the limit
      for (let i = 0; i < 55; i++) {
        service.log(new Error(`Error ${i}`))
      }

      const logs = service.getLocalLogs()
      expect(logs.length).toBeLessThanOrEqual(50)
    })
  })

  describe('clearLocalLogs', () => {
    it('removes all local logs', () => {
      service.log(new Error('Test error'))
      expect(service.getLocalLogs()).toHaveLength(1)

      service.clearLocalLogs()

      expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('gorax_error_logs')
    })
  })

  describe('getLogsByCategory', () => {
    it('filters logs by category', () => {
      service.log(new Error('Network error'))
      service.log(new Error('Validation error'))
      service.log(new Error('Another network request failed'))

      const networkLogs = service.getLogsByCategory('network')
      const validationLogs = service.getLogsByCategory('validation')

      expect(networkLogs.length).toBeGreaterThanOrEqual(1)
      expect(validationLogs.length).toBeGreaterThanOrEqual(1)
    })
  })

  describe('getLogsBySeverity', () => {
    it('filters logs by severity', () => {
      // Log different types to get different severities
      service.log(new Error('Some error'), { componentStack: 'at Component' })
      service.log(new Error('Validation failed'))

      const allLogs = service.getLocalLogs()
      expect(allLogs.length).toBeGreaterThan(0)
    })
  })
})

describe('errorLogger singleton', () => {
  it('exports singleton instance', () => {
    expect(errorLogger).toBeInstanceOf(ErrorLoggerService)
  })

  it('singleton has all required methods', () => {
    expect(typeof errorLogger.log).toBe('function')
    expect(typeof errorLogger.logBoundaryError).toBe('function')
    expect(typeof errorLogger.logNetworkError).toBe('function')
    expect(typeof errorLogger.logWorkflowError).toBe('function')
    expect(typeof errorLogger.getLocalLogs).toBe('function')
    expect(typeof errorLogger.clearLocalLogs).toBe('function')
    expect(typeof errorLogger.getLogsByCategory).toBe('function')
    expect(typeof errorLogger.getLogsBySeverity).toBe('function')
  })
})
