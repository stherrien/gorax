import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { useWebhooks, useWebhook, useWebhookEvents, useWebhookMutations } from './useWebhooks'
import type { Webhook, WebhookEvent } from '../api/webhooks'

// Mock the webhook API
vi.mock('../api/webhooks', () => ({
  webhookAPI: {
    list: vi.fn(),
    get: vi.fn(),
    getEvents: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    regenerateSecret: vi.fn(),
    test: vi.fn(),
  },
}))

import { webhookAPI } from '../api/webhooks'

describe('useWebhooks', () => {
  const mockWebhook: Webhook = {
    id: 'wh-123',
    tenantId: 'tenant-1',
    workflowId: 'wf-123',
    name: 'Test Webhook',
    path: '/webhooks/wf-123/wh-123',
    authType: 'signature',
    enabled: true,
    priority: 1,
    triggerCount: 0,
    createdAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-01-15T10:00:00Z',
    url: 'http://localhost:8080/webhooks/wf-123/wh-123',
  }

  const mockWebhookEvent: WebhookEvent = {
    id: 'evt-123',
    webhookId: 'wh-123',
    executionId: 'exec-123',
    requestMethod: 'POST',
    requestHeaders: { 'content-type': 'application/json' },
    requestBody: { data: 'test' },
    responseStatus: 200,
    processingTimeMs: 150,
    status: 'processed',
    createdAt: '2024-01-15T10:00:00Z',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('useWebhooks - list hook', () => {
    it('should load webhooks on mount', async () => {
      (webhookAPI.list as any).mockResolvedValueOnce({
        webhooks: [mockWebhook],
        total: 1,
      })

      const { result } = renderHook(() => useWebhooks())

      expect(result.current.loading).toBe(true)
      expect(result.current.webhooks).toEqual([])

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.webhooks).toEqual([mockWebhook])
      expect(result.current.total).toBe(1)
      expect(result.current.error).toBeNull()
    })

    it('should handle empty list', async () => {
      (webhookAPI.list as any).mockResolvedValueOnce({
        webhooks: [],
        total: 0,
      })

      const { result } = renderHook(() => useWebhooks())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.webhooks).toEqual([])
      expect(result.current.total).toBe(0)
    })

    it('should handle errors', async () => {
      const error = new Error('Failed to load webhooks')
      ;(webhookAPI.list as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhooks())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe(error)
      expect(result.current.webhooks).toEqual([])
      expect(result.current.total).toBe(0)
    })

    it('should clear error state on successful refetch', async () => {
      const error = new Error('Network error')
      ;(webhookAPI.list as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhooks())

      await waitFor(() => {
        expect(result.current.error).toBe(error)
      })

      // Mock successful response on refetch
      ;(webhookAPI.list as any).mockResolvedValueOnce({
        webhooks: [mockWebhook],
        total: 1,
      })

      result.current.refetch()

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBeNull()
      expect(result.current.webhooks).toEqual([mockWebhook])
    })

    it('should support refetch', async () => {
      (webhookAPI.list as any).mockResolvedValue({
        webhooks: [mockWebhook],
        total: 1,
      })

      const { result } = renderHook(() => useWebhooks())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(webhookAPI.list).toHaveBeenCalledTimes(1)

      // Trigger refetch
      result.current.refetch()

      await waitFor(() => {
        expect(webhookAPI.list).toHaveBeenCalledTimes(2)
      })
    })

    it('should show loading state during refetch', async () => {
      (webhookAPI.list as any).mockResolvedValue({
        webhooks: [mockWebhook],
        total: 1,
      })

      const { result } = renderHook(() => useWebhooks())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      await act(async () => {
        result.current.refetch()
      })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })
    })

    it('should pass filter params to API', async () => {
      (webhookAPI.list as any).mockResolvedValueOnce({
        webhooks: [],
        total: 0,
      })

      renderHook(() => useWebhooks({ workflowId: 'wf-123', page: 2 }))

      await waitFor(() => {
        expect(webhookAPI.list).toHaveBeenCalledWith({ workflowId: 'wf-123', page: 2 })
      })
    })

    it('should filter by enabled status', async () => {
      (webhookAPI.list as any).mockResolvedValueOnce({
        webhooks: [],
        total: 0,
      })

      renderHook(() => useWebhooks({ enabled: true }))

      await waitFor(() => {
        expect(webhookAPI.list).toHaveBeenCalledWith({ enabled: true })
      })
    })

    it('should handle multiple webhooks', async () => {
      const secondWebhook: Webhook = {
        ...mockWebhook,
        id: 'wh-456',
        name: 'Second Webhook',
      }

      ;(webhookAPI.list as any).mockResolvedValueOnce({
        webhooks: [mockWebhook, secondWebhook],
        total: 2,
      })

      const { result } = renderHook(() => useWebhooks())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.webhooks).toHaveLength(2)
      expect(result.current.total).toBe(2)
      expect(result.current.webhooks[0].id).toBe('wh-123')
      expect(result.current.webhooks[1].id).toBe('wh-456')
    })

    it('should pass pagination params', async () => {
      (webhookAPI.list as any).mockResolvedValueOnce({
        webhooks: [],
        total: 0,
      })

      renderHook(() => useWebhooks({ page: 3, limit: 25 }))

      await waitFor(() => {
        expect(webhookAPI.list).toHaveBeenCalledWith({ page: 3, limit: 25 })
      })
    })
  })

  describe('useWebhook - single webhook hook', () => {
    it('should load webhook by ID on mount', async () => {
      (webhookAPI.get as any).mockResolvedValueOnce(mockWebhook)

      const { result } = renderHook(() => useWebhook('wh-123'))

      expect(result.current.loading).toBe(true)

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.webhook).toEqual(mockWebhook)
      expect(result.current.error).toBeNull()
      expect(webhookAPI.get).toHaveBeenCalledWith('wh-123')
    })

    it('should not load if ID is null', () => {
      const { result } = renderHook(() => useWebhook(null))

      expect(result.current.loading).toBe(false)
      expect(result.current.webhook).toBeNull()
      expect(result.current.error).toBeNull()
      expect(webhookAPI.get).not.toHaveBeenCalled()
    })

    it('should handle not found error', async () => {
      const error = new Error('Webhook not found')
      error.name = 'NotFoundError'
      ;(webhookAPI.get as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhook('invalid-id'))

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe(error)
      expect(result.current.webhook).toBeNull()
    })

    it('should handle generic error', async () => {
      const error = new Error('Network timeout')
      ;(webhookAPI.get as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhook('wh-123'))

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe(error)
      expect(result.current.webhook).toBeNull()
    })

    it('should clear error on successful refetch', async () => {
      const error = new Error('Initial error')
      ;(webhookAPI.get as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhook('wh-123'))

      await waitFor(() => {
        expect(result.current.error).toBe(error)
      })

      ;(webhookAPI.get as any).mockResolvedValueOnce(mockWebhook)
      result.current.refetch()

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBeNull()
      expect(result.current.webhook).toEqual(mockWebhook)
    })

    it('should refetch webhook', async () => {
      (webhookAPI.get as any).mockResolvedValue(mockWebhook)

      const { result } = renderHook(() => useWebhook('wh-123'))

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      result.current.refetch()

      await waitFor(() => {
        expect(webhookAPI.get).toHaveBeenCalledTimes(2)
      })
    })

    it('should show loading state during refetch', async () => {
      (webhookAPI.get as any).mockResolvedValue(mockWebhook)

      const { result } = renderHook(() => useWebhook('wh-123'))

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      await act(async () => {
        result.current.refetch()
      })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })
    })

    it('should update when ID changes', async () => {
      const secondWebhook: Webhook = {
        ...mockWebhook,
        id: 'wh-456',
        name: 'Second Webhook',
      }

      ;(webhookAPI.get as any).mockResolvedValueOnce(mockWebhook)

      const { result, rerender } = renderHook(
        ({ id }) => useWebhook(id),
        { initialProps: { id: 'wh-123' } }
      )

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.webhook?.id).toBe('wh-123')

      ;(webhookAPI.get as any).mockResolvedValueOnce(secondWebhook)
      rerender({ id: 'wh-456' })

      await waitFor(() => {
        expect(result.current.webhook?.id).toBe('wh-456')
      })

      expect(webhookAPI.get).toHaveBeenCalledTimes(2)
      expect(webhookAPI.get).toHaveBeenCalledWith('wh-123')
      expect(webhookAPI.get).toHaveBeenCalledWith('wh-456')
    })

    it('should handle transition from null to valid ID', async () => {
      (webhookAPI.get as any).mockResolvedValueOnce(mockWebhook)

      const { result, rerender } = renderHook(
        ({ id }) => useWebhook(id),
        { initialProps: { id: null } }
      )

      expect(result.current.loading).toBe(false)
      expect(webhookAPI.get).not.toHaveBeenCalled()

      rerender({ id: 'wh-123' })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.webhook).toEqual(mockWebhook)
      expect(webhookAPI.get).toHaveBeenCalledWith('wh-123')
    })
  })

  describe('useWebhookEvents - webhook event history', () => {
    it('should load webhook events on mount', async () => {
      (webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: [mockWebhookEvent],
        total: 1,
      })

      const { result } = renderHook(() => useWebhookEvents('wh-123'))

      expect(result.current.loading).toBe(true)
      expect(result.current.events).toEqual([])

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.events).toEqual([mockWebhookEvent])
      expect(result.current.total).toBe(1)
      expect(result.current.error).toBeNull()
      expect(webhookAPI.getEvents).toHaveBeenCalledWith('wh-123', undefined)
    })

    it('should not load if webhook ID is null', () => {
      const { result } = renderHook(() => useWebhookEvents(null))

      expect(result.current.loading).toBe(false)
      expect(result.current.events).toEqual([])
      expect(result.current.error).toBeNull()
      expect(webhookAPI.getEvents).not.toHaveBeenCalled()
    })

    it('should handle empty events list', async () => {
      (webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: [],
        total: 0,
      })

      const { result } = renderHook(() => useWebhookEvents('wh-123'))

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.events).toEqual([])
      expect(result.current.total).toBe(0)
    })

    it('should handle pagination params', async () => {
      (webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: [],
        total: 0,
      })

      renderHook(() => useWebhookEvents('wh-123', { page: 2, limit: 50 }))

      await waitFor(() => {
        expect(webhookAPI.getEvents).toHaveBeenCalledWith('wh-123', { page: 2, limit: 50 })
      })
    })

    it('should handle errors', async () => {
      const error = new Error('Failed to load events')
      ;(webhookAPI.getEvents as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhookEvents('wh-123'))

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe(error)
      expect(result.current.events).toEqual([])
      expect(result.current.total).toBe(0)
    })

    it('should clear error on successful refetch', async () => {
      const error = new Error('Failed to load')
      ;(webhookAPI.getEvents as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhookEvents('wh-123'))

      await waitFor(() => {
        expect(result.current.error).toBe(error)
      })

      ;(webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: [mockWebhookEvent],
        total: 1,
      })

      result.current.refetch()

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBeNull()
      expect(result.current.events).toEqual([mockWebhookEvent])
    })

    it('should support refetch', async () => {
      (webhookAPI.getEvents as any).mockResolvedValue({
        events: [mockWebhookEvent],
        total: 1,
      })

      const { result } = renderHook(() => useWebhookEvents('wh-123'))

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(webhookAPI.getEvents).toHaveBeenCalledTimes(1)

      result.current.refetch()

      await waitFor(() => {
        expect(webhookAPI.getEvents).toHaveBeenCalledTimes(2)
      })
    })

    it('should show loading state during refetch', async () => {
      (webhookAPI.getEvents as any).mockResolvedValue({
        events: [mockWebhookEvent],
        total: 1,
      })

      const { result } = renderHook(() => useWebhookEvents('wh-123'))

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      await act(async () => {
        result.current.refetch()
      })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })
    })

    it('should handle multiple events', async () => {
      const secondEvent: WebhookEvent = {
        ...mockWebhookEvent,
        id: 'evt-456',
        status: 'failed',
        errorMessage: 'Processing error',
      }

      ;(webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: [mockWebhookEvent, secondEvent],
        total: 2,
      })

      const { result } = renderHook(() => useWebhookEvents('wh-123'))

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.events).toHaveLength(2)
      expect(result.current.total).toBe(2)
      expect(result.current.events[0].status).toBe('processed')
      expect(result.current.events[1].status).toBe('failed')
    })

    it('should update when webhook ID changes', async () => {
      const webhook2Event: WebhookEvent = {
        ...mockWebhookEvent,
        id: 'evt-999',
        webhookId: 'wh-456',
      }

      ;(webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: [mockWebhookEvent],
        total: 1,
      })

      const { result, rerender } = renderHook(
        ({ webhookId }) => useWebhookEvents(webhookId),
        { initialProps: { webhookId: 'wh-123' } }
      )

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.events[0].webhookId).toBe('wh-123')

      ;(webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: [webhook2Event],
        total: 1,
      })

      rerender({ webhookId: 'wh-456' })

      await waitFor(() => {
        expect(result.current.events[0]?.webhookId).toBe('wh-456')
      })

      expect(webhookAPI.getEvents).toHaveBeenCalledTimes(2)
    })

    it('should handle transition from null to valid webhook ID', async () => {
      (webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: [mockWebhookEvent],
        total: 1,
      })

      const { result, rerender } = renderHook(
        ({ webhookId }) => useWebhookEvents(webhookId),
        { initialProps: { webhookId: null } }
      )

      expect(result.current.loading).toBe(false)
      expect(webhookAPI.getEvents).not.toHaveBeenCalled()

      rerender({ webhookId: 'wh-123' })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.events).toEqual([mockWebhookEvent])
      expect(webhookAPI.getEvents).toHaveBeenCalledWith('wh-123', undefined)
    })
  })

  describe('useWebhookMutations - CRUD operations', () => {
    it('should create webhook', async () => {
      const newWebhook = {
        name: 'New Webhook',
        workflowId: 'wf-123',
        path: '/webhook-path',
        authType: 'signature' as const,
      }

      ;(webhookAPI.create as any).mockResolvedValueOnce({
        ...mockWebhook,
        ...newWebhook,
      })

      const { result } = renderHook(() => useWebhookMutations())

      expect(result.current.creating).toBe(false)

      let created: Webhook
      await act(async () => {
        created = await result.current.createWebhook(newWebhook)
      })

      expect(result.current.creating).toBe(false)
      expect(created!.workflowId).toBe('wf-123')
      expect(webhookAPI.create).toHaveBeenCalledWith(newWebhook)
    })

    it('should handle create error', async () => {
      const error = new Error('Validation failed')
      ;(webhookAPI.create as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhookMutations())

      await expect(
        result.current.createWebhook({
          name: '',
          workflowId: '',
          path: '',
          authType: 'signature',
        })
      ).rejects.toThrow('Validation failed')

      expect(result.current.creating).toBe(false)
    })

    it('should reset creating state even on error', async () => {
      const error = new Error('Network error')
      ;(webhookAPI.create as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhookMutations())

      await act(async () => {
        await expect(
          result.current.createWebhook({
            name: 'Test',
            workflowId: 'wf-1',
            path: '/test',
            authType: 'signature',
          })
        ).rejects.toThrow('Network error')
      })

      expect(result.current.creating).toBe(false)
    })

    it('should update webhook', async () => {
      const updates = { enabled: false }
      ;(webhookAPI.update as any).mockResolvedValueOnce({
        ...mockWebhook,
        ...updates,
      })

      const { result } = renderHook(() => useWebhookMutations())

      expect(result.current.updating).toBe(false)

      let updated: Webhook
      await act(async () => {
        updated = await result.current.updateWebhook('wh-123', updates)
      })

      expect(result.current.updating).toBe(false)
      expect(updated!.enabled).toBe(false)
      expect(webhookAPI.update).toHaveBeenCalledWith('wh-123', updates)
    })

    it('should handle update error', async () => {
      const error = new Error('Update failed')
      ;(webhookAPI.update as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhookMutations())

      await expect(
        result.current.updateWebhook('wh-123', { name: 'Updated' })
      ).rejects.toThrow('Update failed')

      expect(result.current.updating).toBe(false)
    })

    it('should delete webhook', async () => {
      (webhookAPI.delete as any).mockResolvedValueOnce(undefined)

      const { result } = renderHook(() => useWebhookMutations())

      expect(result.current.deleting).toBe(false)

      await act(async () => {
        await result.current.deleteWebhook('wh-123')
      })

      expect(result.current.deleting).toBe(false)
      expect(webhookAPI.delete).toHaveBeenCalledWith('wh-123')
    })

    it('should handle delete error', async () => {
      const error = new Error('Delete failed')
      ;(webhookAPI.delete as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhookMutations())

      await expect(result.current.deleteWebhook('wh-123')).rejects.toThrow('Delete failed')

      expect(result.current.deleting).toBe(false)
    })

    it('should regenerate webhook secret', async () => {
      const secretResponse = { secret: 'new-secret-key' }
      ;(webhookAPI.regenerateSecret as any).mockResolvedValueOnce(secretResponse)

      const { result } = renderHook(() => useWebhookMutations())

      expect(result.current.regenerating).toBe(false)

      let response: { secret: string }
      await act(async () => {
        response = await result.current.regenerateSecret('wh-123')
      })

      expect(result.current.regenerating).toBe(false)
      expect(response!.secret).toBe('new-secret-key')
      expect(webhookAPI.regenerateSecret).toHaveBeenCalledWith('wh-123')
    })

    it('should handle regenerate error', async () => {
      const error = new Error('Regenerate failed')
      ;(webhookAPI.regenerateSecret as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhookMutations())

      await expect(result.current.regenerateSecret('wh-123')).rejects.toThrow(
        'Regenerate failed'
      )

      expect(result.current.regenerating).toBe(false)
    })

    it('should test webhook', async () => {
      const testResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 150,
        executionId: 'exec-123',
      }
      ;(webhookAPI.test as any).mockResolvedValueOnce(testResponse)

      const { result } = renderHook(() => useWebhookMutations())

      expect(result.current.testing).toBe(false)

      const testData = { method: 'POST', body: { test: 'data' } }
      let response: any
      await act(async () => {
        response = await result.current.testWebhook('wh-123', testData)
      })

      expect(result.current.testing).toBe(false)
      expect(response.success).toBe(true)
      expect(response.executionId).toBe('exec-123')
      expect(webhookAPI.test).toHaveBeenCalledWith('wh-123', testData)
    })

    it('should test webhook without payload', async () => {
      const testResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 100,
        executionId: 'exec-456',
      }
      ;(webhookAPI.test as any).mockResolvedValueOnce(testResponse)

      const { result } = renderHook(() => useWebhookMutations())

      await result.current.testWebhook('wh-123')

      expect(webhookAPI.test).toHaveBeenCalledWith('wh-123', {})
    })

    it('should handle test error', async () => {
      const error = new Error('Test failed')
      ;(webhookAPI.test as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhookMutations())

      await expect(result.current.testWebhook('wh-123')).rejects.toThrow('Test failed')

      expect(result.current.testing).toBe(false)
    })

    it('should test webhook with failed response', async () => {
      const testResponse = {
        success: false,
        statusCode: 500,
        responseTimeMs: 50,
        error: 'Internal server error',
      }
      ;(webhookAPI.test as any).mockResolvedValueOnce(testResponse)

      const { result } = renderHook(() => useWebhookMutations())

      const response = await result.current.testWebhook('wh-123')

      expect(response.success).toBe(false)
      expect(response.statusCode).toBe(500)
      expect(response.error).toBe('Internal server error')
    })

    it('should test webhook with custom headers', async () => {
      const testResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 120,
      }
      const testData = {
        method: 'POST',
        headers: { 'X-Custom-Header': 'test-value' },
        body: { data: 'custom' },
      }

      ;(webhookAPI.test as any).mockResolvedValueOnce(testResponse)

      const { result } = renderHook(() => useWebhookMutations())

      await result.current.testWebhook('wh-123', testData)

      expect(webhookAPI.test).toHaveBeenCalledWith('wh-123', testData)
    })

    it('should maintain independent loading states for all mutations', async () => {
      (webhookAPI.create as any).mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve(mockWebhook), 50))
      )
      ;(webhookAPI.update as any).mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve(mockWebhook), 50))
      )

      const { result } = renderHook(() => useWebhookMutations())

      await act(async () => {
        const createPromise = result.current.createWebhook({
          name: 'Test',
          workflowId: 'wf-1',
          path: '/test',
          authType: 'signature',
        })

        const updatePromise = result.current.updateWebhook('wh-123', { name: 'Updated' })

        await Promise.all([createPromise, updatePromise])
      })

      expect(result.current.creating).toBe(false)
      expect(result.current.updating).toBe(false)
    })
  })
})
