import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { useWebhooks, useWebhook, useWebhookEvents, useWebhookMutations } from './useWebhooks'
import type { Webhook, WebhookEvent } from '../api/webhooks'

// Mock the webhook API
import { createQueryWrapper } from "../test/test-utils"
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

// Helper to create a wrapper with QueryClient
function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
      mutations: {
        retry: false,
      },
    },
  })
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('useWebhooks', () => {
  // Valid RFC 4122 UUIDs (version 4, variant 1)
  const webhookId = '11111111-1111-4111-8111-111111111111'
  const webhookId2 = '22222222-2222-4222-8222-222222222222'
  const tenantId = '33333333-3333-4333-8333-333333333333'
  const workflowId = '44444444-4444-4444-8444-444444444444'
  const eventId = '55555555-5555-4555-8555-555555555555'
  const executionId = '66666666-6666-4666-8666-666666666666'

  const mockWebhook: Webhook = {
    id: webhookId,
    tenantId: tenantId,
    workflowId: workflowId,
    name: 'Test Webhook',
    path: `/webhooks/${workflowId}/${webhookId}`,
    authType: 'signature',
    enabled: true,
    priority: 1,
    triggerCount: 0,
    createdAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-01-15T10:00:00Z',
    url: `http://localhost:8080/webhooks/${workflowId}/${webhookId}`,
  }

  const mockWebhookEvent: WebhookEvent = {
    id: eventId,
    webhookId: webhookId,
    executionId: executionId,
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

      const { result } = renderHook(() => useWebhooks(), { wrapper: createWrapper() })

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

      const { result } = renderHook(() => useWebhooks(), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.webhooks).toEqual([])
      expect(result.current.total).toBe(0)
    })

    it('should handle errors', async () => {
      const error = new Error('Failed to load webhooks')
      ;(webhookAPI.list as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhooks(), { wrapper: createWrapper() })

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

      const { result } = renderHook(() => useWebhooks(), { wrapper: createWrapper() })

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

      const { result } = renderHook(() => useWebhooks(), { wrapper: createWrapper() })

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

      const { result } = renderHook(() => useWebhooks(), { wrapper: createWrapper() })

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

      renderHook(() => useWebhooks({ workflowId: 'wf-123', page: 2 }), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(webhookAPI.list).toHaveBeenCalledWith({ workflowId: 'wf-123', page: 2 })
      })
    })

    it('should filter by enabled status', async () => {
      (webhookAPI.list as any).mockResolvedValueOnce({
        webhooks: [],
        total: 0,
      })

      renderHook(() => useWebhooks({ enabled: true }), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(webhookAPI.list).toHaveBeenCalledWith({ enabled: true })
      })
    })

    it('should handle multiple webhooks', async () => {
      const secondWebhook: Webhook = {
        ...mockWebhook,
        id: webhookId2,
        name: 'Second Webhook',
      }

      ;(webhookAPI.list as any).mockResolvedValueOnce({
        webhooks: [mockWebhook, secondWebhook],
        total: 2,
      })

      const { result } = renderHook(() => useWebhooks(), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.webhooks).toHaveLength(2)
      expect(result.current.total).toBe(2)
      expect(result.current.webhooks[0].id).toBe(webhookId)
      expect(result.current.webhooks[1].id).toBe(webhookId2)
    })

    it('should pass pagination params', async () => {
      (webhookAPI.list as any).mockResolvedValueOnce({
        webhooks: [],
        total: 0,
      })

      renderHook(() => useWebhooks({ page: 3, limit: 25 }), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(webhookAPI.list).toHaveBeenCalledWith({ page: 3, limit: 25 })
      })
    })
  })

  describe('useWebhook - single webhook hook', () => {
    it('should load webhook by ID on mount', async () => {
      (webhookAPI.get as any).mockResolvedValueOnce(mockWebhook)

      const { result } = renderHook(() => useWebhook(webhookId), { wrapper: createWrapper() })

      expect(result.current.loading).toBe(true)

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.webhook).toEqual(mockWebhook)
      expect(result.current.error).toBeNull()
      expect(webhookAPI.get).toHaveBeenCalledWith(webhookId)
    })

    it('should not load if ID is null', () => {
      const { result } = renderHook(() => useWebhook(null), { wrapper: createWrapper() })

      expect(result.current.loading).toBe(false)
      expect(result.current.webhook).toBeNull()
      expect(result.current.error).toBeNull()
      expect(webhookAPI.get).not.toHaveBeenCalled()
    })

    it('should handle not found error', async () => {
      const error = new Error('Webhook not found')
      error.name = 'NotFoundError'
      ;(webhookAPI.get as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhook('99999999-9999-4999-8999-999999999999'), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe(error)
      expect(result.current.webhook).toBeNull()
    })

    it('should handle generic error', async () => {
      const error = new Error('Network timeout')
      ;(webhookAPI.get as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhook(webhookId), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe(error)
      expect(result.current.webhook).toBeNull()
    })

    it('should clear error on successful refetch', async () => {
      const error = new Error('Initial error')
      ;(webhookAPI.get as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhook(webhookId), { wrapper: createWrapper() })

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

      const { result } = renderHook(() => useWebhook(webhookId), { wrapper: createWrapper() })

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

      const { result } = renderHook(() => useWebhook(webhookId), { wrapper: createWrapper() })

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
        id: webhookId2,
        name: 'Second Webhook',
      }

      ;(webhookAPI.get as any).mockResolvedValueOnce(mockWebhook)

      const { result, rerender } = renderHook(
        ({ id }) => useWebhook(id),
        { initialProps: { id: webhookId }, wrapper: createWrapper() }
      )

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.webhook?.id).toBe(webhookId)

      ;(webhookAPI.get as any).mockResolvedValueOnce(secondWebhook)
      rerender({ id: webhookId2 })

      await waitFor(() => {
        expect(result.current.webhook?.id).toBe(webhookId2)
      })

      expect(webhookAPI.get).toHaveBeenCalledTimes(2)
      expect(webhookAPI.get).toHaveBeenCalledWith(webhookId)
      expect(webhookAPI.get).toHaveBeenCalledWith(webhookId2)
    })

    it('should handle transition from null to valid ID', async () => {
      (webhookAPI.get as any).mockResolvedValueOnce(mockWebhook)

      const { result, rerender } = renderHook(
        ({ id }) => useWebhook(id),
        { initialProps: { id: null }, wrapper: createWrapper() }
      )

      expect(result.current.loading).toBe(false)
      expect(webhookAPI.get).not.toHaveBeenCalled()

      rerender({ id: webhookId })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.webhook).toEqual(mockWebhook)
      expect(webhookAPI.get).toHaveBeenCalledWith(webhookId)
    })
  })

  describe('useWebhookEvents - webhook event history', () => {
    it('should load webhook events on mount', async () => {
      (webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: [mockWebhookEvent],
        total: 1,
      })

      const { result } = renderHook(() => useWebhookEvents(webhookId), { wrapper: createWrapper() })

      expect(result.current.loading).toBe(true)
      expect(result.current.events).toEqual([])

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.events).toEqual([mockWebhookEvent])
      expect(result.current.total).toBe(1)
      expect(result.current.error).toBeNull()
      expect(webhookAPI.getEvents).toHaveBeenCalledWith(webhookId, undefined)
    })

    it('should not load if webhook ID is null', () => {
      const { result } = renderHook(() => useWebhookEvents(null), { wrapper: createWrapper() })

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

      const { result } = renderHook(() => useWebhookEvents(webhookId), { wrapper: createWrapper() })

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

      renderHook(() => useWebhookEvents(webhookId, { page: 2, limit: 50 }), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(webhookAPI.getEvents).toHaveBeenCalledWith(webhookId, { page: 2, limit: 50 })
      })
    })

    it('should handle errors', async () => {
      const error = new Error('Failed to load events')
      ;(webhookAPI.getEvents as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhookEvents(webhookId), { wrapper: createWrapper() })

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

      const { result } = renderHook(() => useWebhookEvents(webhookId), { wrapper: createWrapper() })

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

      const { result } = renderHook(() => useWebhookEvents(webhookId), { wrapper: createWrapper() })

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

      const { result } = renderHook(() => useWebhookEvents(webhookId), { wrapper: createWrapper() })

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

      const { result } = renderHook(() => useWebhookEvents(webhookId), { wrapper: createWrapper() })

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
        webhookId: webhookId2,
      }

      ;(webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: [mockWebhookEvent],
        total: 1,
      })

      const { result, rerender } = renderHook(
        ({ webhookId }) => useWebhookEvents(webhookId),
        { initialProps: { webhookId: webhookId }, wrapper: createWrapper() }
      )

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.events[0].webhookId).toBe(webhookId)

      ;(webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: [webhook2Event],
        total: 1,
      })

      rerender({ webhookId: webhookId2 })

      await waitFor(() => {
        expect(result.current.events[0]?.webhookId).toBe(webhookId2)
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
        { initialProps: { webhookId: null }, wrapper: createWrapper() }
      )

      expect(result.current.loading).toBe(false)
      expect(webhookAPI.getEvents).not.toHaveBeenCalled()

      rerender({ webhookId: webhookId })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.events).toEqual([mockWebhookEvent])
      expect(webhookAPI.getEvents).toHaveBeenCalledWith(webhookId, undefined)
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

      const { result } = renderHook(() => useWebhookMutations(), { wrapper: createWrapper() })

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

      const { result } = renderHook(() => useWebhookMutations(), { wrapper: createWrapper() })

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

      const { result } = renderHook(() => useWebhookMutations(), { wrapper: createWrapper() })

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

      const { result } = renderHook(() => useWebhookMutations(), { wrapper: createWrapper() })

      expect(result.current.updating).toBe(false)

      let updated: Webhook
      await act(async () => {
        updated = await result.current.updateWebhook(webhookId, updates)
      })

      expect(result.current.updating).toBe(false)
      expect(updated!.enabled).toBe(false)
      expect(webhookAPI.update).toHaveBeenCalledWith(webhookId, updates)
    })

    it('should handle update error', async () => {
      const error = new Error('Update failed')
      ;(webhookAPI.update as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhookMutations(), { wrapper: createWrapper() })

      await expect(
        result.current.updateWebhook(webhookId, { name: 'Updated' })
      ).rejects.toThrow('Update failed')

      expect(result.current.updating).toBe(false)
    })

    it('should delete webhook', async () => {
      (webhookAPI.delete as any).mockResolvedValueOnce(undefined)

      const { result } = renderHook(() => useWebhookMutations(), { wrapper: createWrapper() })

      expect(result.current.deleting).toBe(false)

      await act(async () => {
        await result.current.deleteWebhook(webhookId)
      })

      expect(result.current.deleting).toBe(false)
      expect(webhookAPI.delete).toHaveBeenCalledWith(webhookId)
    })

    it('should handle delete error', async () => {
      const error = new Error('Delete failed')
      ;(webhookAPI.delete as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhookMutations(), { wrapper: createWrapper() })

      await expect(result.current.deleteWebhook(webhookId)).rejects.toThrow('Delete failed')

      expect(result.current.deleting).toBe(false)
    })

    it('should regenerate webhook secret', async () => {
      const secretResponse = { secret: 'new-secret-key' }
      ;(webhookAPI.regenerateSecret as any).mockResolvedValueOnce(secretResponse)

      const { result } = renderHook(() => useWebhookMutations(), { wrapper: createWrapper() })

      expect(result.current.regenerating).toBe(false)

      let response: { secret: string }
      await act(async () => {
        response = await result.current.regenerateSecret(webhookId)
      })

      expect(result.current.regenerating).toBe(false)
      expect(response!.secret).toBe('new-secret-key')
      expect(webhookAPI.regenerateSecret).toHaveBeenCalledWith(webhookId)
    })

    it('should handle regenerate error', async () => {
      const error = new Error('Regenerate failed')
      ;(webhookAPI.regenerateSecret as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhookMutations(), { wrapper: createWrapper() })

      await expect(result.current.regenerateSecret(webhookId)).rejects.toThrow(
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

      const { result } = renderHook(() => useWebhookMutations(), { wrapper: createWrapper() })

      expect(result.current.testing).toBe(false)

      const testData = { method: 'POST', body: { test: 'data' } }
      let response: any
      await act(async () => {
        response = await result.current.testWebhook(webhookId, testData)
      })

      expect(result.current.testing).toBe(false)
      expect(response.success).toBe(true)
      expect(response.executionId).toBe('exec-123')
      expect(webhookAPI.test).toHaveBeenCalledWith(webhookId, testData)
    })

    it('should test webhook without payload', async () => {
      const testResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 100,
        executionId: 'exec-456',
      }
      ;(webhookAPI.test as any).mockResolvedValueOnce(testResponse)

      const { result } = renderHook(() => useWebhookMutations(), { wrapper: createWrapper() })

      await result.current.testWebhook(webhookId)

      expect(webhookAPI.test).toHaveBeenCalledWith(webhookId, {})
    })

    it('should handle test error', async () => {
      const error = new Error('Test failed')
      ;(webhookAPI.test as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWebhookMutations(), { wrapper: createWrapper() })

      await expect(result.current.testWebhook(webhookId)).rejects.toThrow('Test failed')

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

      const { result } = renderHook(() => useWebhookMutations(), { wrapper: createWrapper() })

      const response = await result.current.testWebhook(webhookId)

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

      const { result } = renderHook(() => useWebhookMutations(), { wrapper: createWrapper() })

      await result.current.testWebhook(webhookId, testData)

      expect(webhookAPI.test).toHaveBeenCalledWith(webhookId, testData)
    })

    it('should maintain independent loading states for all mutations', async () => {
      (webhookAPI.create as any).mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve(mockWebhook), 50))
      )
      ;(webhookAPI.update as any).mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve(mockWebhook), 50))
      )

      const { result } = renderHook(() => useWebhookMutations(), { wrapper: createWrapper() })

      await act(async () => {
        const createPromise = result.current.createWebhook({
          name: 'Test',
          workflowId: 'wf-1',
          path: '/test',
          authType: 'signature',
        })

        const updatePromise = result.current.updateWebhook(webhookId, { name: 'Updated' })

        await Promise.all([createPromise, updatePromise])
      })

      expect(result.current.creating).toBe(false)
      expect(result.current.updating).toBe(false)
    })
  })
})
