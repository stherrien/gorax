import { describe, it, expect, beforeEach, vi } from 'vitest'
import { webhookAPI } from './webhooks'
import type {
  Webhook,
  WebhookCreateInput,
  WebhookUpdateInput,
  TestWebhookInput,
} from './webhooks'

// Mock the API client
vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

import { apiClient } from './client'

describe('Webhook API', () => {
  const mockWebhook: Webhook = {
    id: 'wh-123',
    tenantId: 'tenant-1',
    workflowId: 'wf-123',
    name: 'Test Webhook',
    path: '/webhook/test',
    authType: 'signature',
    enabled: true,
    priority: 0,
    triggerCount: 42,
    lastTriggeredAt: '2024-01-15T10:00:00Z',
    createdAt: '2024-01-15T09:00:00Z',
    updatedAt: '2024-01-15T10:00:00Z',
    url: 'https://api.example.com/webhook/test',
  }

  const mockWebhookEvent = {
    id: 'evt-123',
    webhookId: 'wh-123',
    executionId: 'exec-123',
    requestMethod: 'POST',
    requestHeaders: { 'content-type': 'application/json' },
    requestBody: { test: 'data' },
    responseStatus: 200,
    processingTimeMs: 150,
    status: 'processed' as const,
    createdAt: '2024-01-15T10:00:00Z',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('should fetch list of webhooks', async () => {
      const mockWebhooks = [mockWebhook]
      ;(apiClient.get as any).mockResolvedValueOnce({ webhooks: mockWebhooks, total: 1 })

      const result = await webhookAPI.list()

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/webhooks', undefined)
      expect(result).toEqual({ webhooks: mockWebhooks, total: 1 })
    })

    it('should handle empty list', async () => {
      (apiClient.get as any).mockResolvedValueOnce({ webhooks: [], total: 0 })

      const result = await webhookAPI.list()

      expect(result.webhooks).toEqual([])
      expect(result.total).toBe(0)
    })

    it('should support pagination parameters', async () => {
      (apiClient.get as any).mockResolvedValueOnce({ webhooks: [], total: 0 })

      await webhookAPI.list({ page: 2, limit: 20 })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/webhooks', {
        params: { page: 2, limit: 20 },
      })
    })

    it('should support workflowId filter', async () => {
      (apiClient.get as any).mockResolvedValueOnce({ webhooks: [], total: 0 })

      await webhookAPI.list({ workflowId: 'wf-123' })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/webhooks', {
        params: { workflowId: 'wf-123' },
      })
    })

    it('should support enabled filter', async () => {
      (apiClient.get as any).mockResolvedValueOnce({ webhooks: [], total: 0 })

      await webhookAPI.list({ enabled: true })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/webhooks', {
        params: { enabled: true },
      })
    })

    it('should support multiple filters', async () => {
      (apiClient.get as any).mockResolvedValueOnce({ webhooks: [], total: 0 })

      await webhookAPI.list({ workflowId: 'wf-123', enabled: true, page: 1, limit: 10 })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/webhooks', {
        params: { workflowId: 'wf-123', enabled: true, page: 1, limit: 10 },
      })
    })
  })

  describe('get', () => {
    it('should fetch single webhook by ID', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockWebhook)

      const result = await webhookAPI.get('wh-123')

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/webhooks/wh-123')
      expect(result).toEqual(mockWebhook)
    })

    it('should throw NotFoundError for invalid ID', async () => {
      const error = new Error('Webhook not found')
      error.name = 'NotFoundError'
      ;(apiClient.get as any).mockRejectedValueOnce(error)

      await expect(webhookAPI.get('invalid-id')).rejects.toThrow('Webhook not found')
    })
  })

  describe('create', () => {
    it('should create new webhook with required fields', async () => {
      const createInput: WebhookCreateInput = {
        name: 'New Webhook',
        workflowId: 'wf-123',
        path: '/webhook/new',
      }

      const createdWebhook = { ...mockWebhook, ...createInput }
      ;(apiClient.post as any).mockResolvedValueOnce(createdWebhook)

      const result = await webhookAPI.create(createInput)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/webhooks', createInput)
      expect(result).toEqual(createdWebhook)
      expect(result.id).toBeDefined()
      expect(result.url).toBeDefined()
    })

    it('should create webhook with optional fields', async () => {
      const createInput: WebhookCreateInput = {
        name: 'New Webhook',
        workflowId: 'wf-123',
        path: '/webhook/new',
        authType: 'api_key',
        description: 'Test description',
        priority: 10,
      }

      const createdWebhook = { ...mockWebhook, ...createInput }
      ;(apiClient.post as any).mockResolvedValueOnce(createdWebhook)

      const result = await webhookAPI.create(createInput)

      expect(result.authType).toBe('api_key')
      expect(result.priority).toBe(10)
    })

    it('should validate required fields', async () => {
      const error = new Error('Name is required')
      error.name = 'ValidationError'
      ;(apiClient.post as any).mockRejectedValueOnce(error)

      const invalidInput = {
        name: '',
        workflowId: 'wf-123',
        path: '/test',
      } as WebhookCreateInput

      await expect(webhookAPI.create(invalidInput)).rejects.toThrow('Name is required')
    })

    it('should validate path format', async () => {
      const error = new Error('Path must start with /')
      error.name = 'ValidationError'
      ;(apiClient.post as any).mockRejectedValueOnce(error)

      const invalidInput: WebhookCreateInput = {
        name: 'Test',
        workflowId: 'wf-123',
        path: 'invalid-path',
      }

      await expect(webhookAPI.create(invalidInput)).rejects.toThrow('Path must start with /')
    })
  })

  describe('update', () => {
    it('should update existing webhook', async () => {
      const updates: WebhookUpdateInput = {
        name: 'Updated Webhook',
        enabled: false,
      }

      const updatedWebhook = { ...mockWebhook, ...updates }
      ;(apiClient.put as any).mockResolvedValueOnce(updatedWebhook)

      const result = await webhookAPI.update('wh-123', updates)

      expect(apiClient.put).toHaveBeenCalledWith('/api/v1/webhooks/wh-123', updates)
      expect(result.name).toBe('Updated Webhook')
      expect(result.enabled).toBe(false)
    })

    it('should preserve unchanged fields', async () => {
      const updates: WebhookUpdateInput = {
        name: 'Updated Name Only',
      }

      const updatedWebhook = { ...mockWebhook, ...updates }
      ;(apiClient.put as any).mockResolvedValueOnce(updatedWebhook)

      const result = await webhookAPI.update('wh-123', updates)

      expect(result.name).toBe('Updated Name Only')
      expect(result.path).toBe(mockWebhook.path)
      expect(result.enabled).toBe(mockWebhook.enabled)
    })

    it('should update auth type', async () => {
      const updates: WebhookUpdateInput = {
        authType: 'basic',
      }

      const updatedWebhook = { ...mockWebhook, ...updates }
      ;(apiClient.put as any).mockResolvedValueOnce(updatedWebhook)

      const result = await webhookAPI.update('wh-123', updates)

      expect(result.authType).toBe('basic')
    })

    it('should update priority', async () => {
      const updates: WebhookUpdateInput = {
        priority: 100,
      }

      const updatedWebhook = { ...mockWebhook, priority: 100 }
      ;(apiClient.put as any).mockResolvedValueOnce(updatedWebhook)

      const result = await webhookAPI.update('wh-123', updates)

      expect(result.priority).toBe(100)
    })

    it('should throw NotFoundError for non-existent webhook', async () => {
      const error = new Error('Webhook not found')
      error.name = 'NotFoundError'
      ;(apiClient.put as any).mockRejectedValueOnce(error)

      await expect(webhookAPI.update('invalid-id', { name: 'Test' })).rejects.toThrow(
        'Webhook not found'
      )
    })
  })

  describe('delete', () => {
    it('should delete webhook by ID', async () => {
      (apiClient.delete as any).mockResolvedValueOnce({})

      await webhookAPI.delete('wh-123')

      expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/webhooks/wh-123')
    })

    it('should throw NotFoundError for non-existent webhook', async () => {
      const error = new Error('Webhook not found')
      error.name = 'NotFoundError'
      ;(apiClient.delete as any).mockRejectedValueOnce(error)

      await expect(webhookAPI.delete('invalid-id')).rejects.toThrow('Webhook not found')
    })
  })

  describe('regenerateSecret', () => {
    it('should regenerate webhook secret', async () => {
      const mockResponse = { secret: 'new-secret-key-12345' }
      ;(apiClient.post as any).mockResolvedValueOnce(mockResponse)

      const result = await webhookAPI.regenerateSecret('wh-123')

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/webhooks/wh-123/regenerate-secret', {})
      expect(result.secret).toBe('new-secret-key-12345')
    })

    it('should throw NotFoundError for non-existent webhook', async () => {
      const error = new Error('Webhook not found')
      error.name = 'NotFoundError'
      ;(apiClient.post as any).mockRejectedValueOnce(error)

      await expect(webhookAPI.regenerateSecret('invalid-id')).rejects.toThrow('Webhook not found')
    })
  })

  describe('test', () => {
    it('should test webhook with default parameters', async () => {
      const mockResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 150,
        executionId: 'exec-123',
      }
      ;(apiClient.post as any).mockResolvedValueOnce(mockResponse)

      const result = await webhookAPI.test('wh-123', {})

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/webhooks/wh-123/test', {})
      expect(result.success).toBe(true)
      expect(result.statusCode).toBe(200)
      expect(result.executionId).toBe('exec-123')
    })

    it('should test webhook with custom request data', async () => {
      const testInput: TestWebhookInput = {
        method: 'POST',
        headers: { 'x-custom': 'header' },
        body: { test: 'data' },
      }

      const mockResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 120,
      }
      ;(apiClient.post as any).mockResolvedValueOnce(mockResponse)

      const result = await webhookAPI.test('wh-123', testInput)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/webhooks/wh-123/test', testInput)
      expect(result.success).toBe(true)
    })

    it('should handle failed webhook test', async () => {
      const mockResponse = {
        success: false,
        statusCode: 500,
        responseTimeMs: 50,
        error: 'Workflow execution failed',
      }
      ;(apiClient.post as any).mockResolvedValueOnce(mockResponse)

      const result = await webhookAPI.test('wh-123', {})

      expect(result.success).toBe(false)
      expect(result.error).toBe('Workflow execution failed')
    })

    it('should throw NotFoundError for non-existent webhook', async () => {
      const error = new Error('Webhook not found')
      error.name = 'NotFoundError'
      ;(apiClient.post as any).mockRejectedValueOnce(error)

      await expect(webhookAPI.test('invalid-id', {})).rejects.toThrow('Webhook not found')
    })
  })

  describe('getEvents', () => {
    it('should fetch webhook events', async () => {
      const mockEvents = [mockWebhookEvent]
      ;(apiClient.get as any).mockResolvedValueOnce({ events: mockEvents, total: 1 })

      const result = await webhookAPI.getEvents('wh-123')

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/webhooks/wh-123/events', undefined)
      expect(result.events).toEqual(mockEvents)
      expect(result.total).toBe(1)
    })

    it('should handle empty events list', async () => {
      (apiClient.get as any).mockResolvedValueOnce({ events: [], total: 0 })

      const result = await webhookAPI.getEvents('wh-123')

      expect(result.events).toEqual([])
      expect(result.total).toBe(0)
    })

    it('should support pagination parameters', async () => {
      (apiClient.get as any).mockResolvedValueOnce({ events: [], total: 0 })

      await webhookAPI.getEvents('wh-123', { page: 2, limit: 50 })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/webhooks/wh-123/events', {
        params: { page: 2, limit: 50 },
      })
    })

    it('should include all event details', async () => {
      const mockEvents = [mockWebhookEvent]
      ;(apiClient.get as any).mockResolvedValueOnce({ events: mockEvents, total: 1 })

      const result = await webhookAPI.getEvents('wh-123')

      const event = result.events[0]
      expect(event.id).toBe('evt-123')
      expect(event.webhookId).toBe('wh-123')
      expect(event.executionId).toBe('exec-123')
      expect(event.requestMethod).toBe('POST')
      expect(event.requestHeaders).toEqual({ 'content-type': 'application/json' })
      expect(event.requestBody).toEqual({ test: 'data' })
      expect(event.responseStatus).toBe(200)
      expect(event.processingTimeMs).toBe(150)
      expect(event.status).toBe('processed')
    })

    it('should throw NotFoundError for non-existent webhook', async () => {
      const error = new Error('Webhook not found')
      error.name = 'NotFoundError'
      ;(apiClient.get as any).mockRejectedValueOnce(error)

      await expect(webhookAPI.getEvents('invalid-id')).rejects.toThrow('Webhook not found')
    })
  })

  describe('replayEvent', () => {
    it('should replay a single event without modified payload', async () => {
      const mockResult = {
        success: true,
        executionId: 'exec-123'
      }

      ;(apiClient.post as any).mockResolvedValueOnce(mockResult)

      const result = await webhookAPI.replayEvent('event-123')

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/events/event-123/replay', {})
      expect(result).toEqual(mockResult)
      expect(result.success).toBe(true)
      expect(result.executionId).toBe('exec-123')
    })

    it('should replay a single event with modified payload', async () => {
      const mockResult = {
        success: true,
        executionId: 'exec-456'
      }

      const modifiedPayload = { test: 'modified data' }

      ;(apiClient.post as any).mockResolvedValueOnce(mockResult)

      const result = await webhookAPI.replayEvent('event-123', modifiedPayload)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/events/event-123/replay', {
        modifiedPayload
      })
      expect(result.executionId).toBe('exec-456')
    })

    it('should handle replay failure - event not found', async () => {
      const mockResult = {
        success: false,
        error: 'event not found'
      }

      ;(apiClient.post as any).mockResolvedValueOnce(mockResult)

      const result = await webhookAPI.replayEvent('event-999')

      expect(result.success).toBe(false)
      expect(result.error).toBe('event not found')
      expect(result.executionId).toBeUndefined()
    })

    it('should handle replay failure - max replay count exceeded', async () => {
      const mockResult = {
        success: false,
        error: 'max replay count (5) exceeded'
      }

      ;(apiClient.post as any).mockResolvedValueOnce(mockResult)

      const result = await webhookAPI.replayEvent('event-123')

      expect(result.success).toBe(false)
      expect(result.error).toContain('max replay count')
    })

    it('should handle replay failure - webhook disabled', async () => {
      const mockResult = {
        success: false,
        error: 'webhook is disabled'
      }

      ;(apiClient.post as any).mockResolvedValueOnce(mockResult)

      const result = await webhookAPI.replayEvent('event-123')

      expect(result.success).toBe(false)
      expect(result.error).toBe('webhook is disabled')
    })
  })

  describe('batchReplayEvents', () => {
    it('should replay multiple events successfully', async () => {
      const mockResponse = {
        results: {
          'event-1': { success: true, executionId: 'exec-1' },
          'event-2': { success: true, executionId: 'exec-2' }
        }
      }

      ;(apiClient.post as any).mockResolvedValueOnce(mockResponse)

      const result = await webhookAPI.batchReplayEvents('webhook-123', ['event-1', 'event-2'])

      expect(apiClient.post).toHaveBeenCalledWith(
        '/api/v1/webhooks/webhook-123/events/replay',
        { eventIds: ['event-1', 'event-2'] }
      )
      expect(result.results['event-1'].success).toBe(true)
      expect(result.results['event-1'].executionId).toBe('exec-1')
      expect(result.results['event-2'].success).toBe(true)
      expect(result.results['event-2'].executionId).toBe('exec-2')
    })

    it('should handle partial success in batch replay', async () => {
      const mockResponse = {
        results: {
          'event-1': { success: true, executionId: 'exec-1' },
          'event-2': { success: false, error: 'event not found' },
          'event-3': { success: false, error: 'max replay count (5) exceeded' }
        }
      }

      ;(apiClient.post as any).mockResolvedValueOnce(mockResponse)

      const result = await webhookAPI.batchReplayEvents('webhook-123', ['event-1', 'event-2', 'event-3'])

      expect(result.results['event-1'].success).toBe(true)
      expect(result.results['event-2'].success).toBe(false)
      expect(result.results['event-2'].error).toBe('event not found')
      expect(result.results['event-3'].success).toBe(false)
      expect(result.results['event-3'].error).toContain('max replay count')
    })

    it('should handle batch size exceeded error', async () => {
      const eventIds = Array.from({ length: 11 }, (_, i) => `event-${i}`)
      const mockResponse = {
        results: Object.fromEntries(
          eventIds.map(id => [
            id,
            { success: false, error: 'batch size exceeds maximum of 10 events' }
          ])
        )
      }

      ;(apiClient.post as any).mockResolvedValueOnce(mockResponse)

      const result = await webhookAPI.batchReplayEvents('webhook-123', eventIds)

      expect(Object.keys(result.results)).toHaveLength(11)
      expect(result.results['event-0'].success).toBe(false)
      expect(result.results['event-0'].error).toContain('batch size exceeds maximum')
      expect(result.results['event-10'].success).toBe(false)
      expect(result.results['event-10'].error).toContain('batch size exceeds maximum')
    })

    it('should handle empty event IDs array', async () => {
      const mockResponse = {
        results: {}
      }

      ;(apiClient.post as any).mockResolvedValueOnce(mockResponse)

      const result = await webhookAPI.batchReplayEvents('webhook-123', [])

      expect(apiClient.post).toHaveBeenCalledWith(
        '/api/v1/webhooks/webhook-123/events/replay',
        { eventIds: [] }
      )
      expect(Object.keys(result.results)).toHaveLength(0)
    })

    it('should replay single event in batch', async () => {
      const mockResponse = {
        results: {
          'event-1': { success: true, executionId: 'exec-1' }
        }
      }

      ;(apiClient.post as any).mockResolvedValueOnce(mockResponse)

      const result = await webhookAPI.batchReplayEvents('webhook-123', ['event-1'])

      expect(result.results['event-1'].success).toBe(true)
      expect(Object.keys(result.results)).toHaveLength(1)
    })
  })
})
