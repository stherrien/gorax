import { useState, useEffect, useCallback } from 'react'
import { webhookAPI } from '../api/webhooks'
import type {
  Webhook,
  WebhookEvent,
  WebhookListParams,
  WebhookCreateInput,
  WebhookUpdateInput,
  TestWebhookInput,
  TestWebhookResponse,
} from '../api/webhooks'

/**
 * Hook to fetch and manage list of webhooks
 */
export function useWebhooks(params?: WebhookListParams) {
  const [webhooks, setWebhooks] = useState<Webhook[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchWebhooks = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await webhookAPI.list(params)
      setWebhooks(response.webhooks)
      setTotal(response.total)
    } catch (err) {
      setError(err as Error)
      setWebhooks([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [params])

  useEffect(() => {
    fetchWebhooks()
  }, [fetchWebhooks])

  return {
    webhooks,
    total,
    loading,
    error,
    refetch: fetchWebhooks,
  }
}

/**
 * Hook to fetch a single webhook by ID
 */
export function useWebhook(id: string | null) {
  const [webhook, setWebhook] = useState<Webhook | null>(null)
  const [loading, setLoading] = useState(!!id)
  const [error, setError] = useState<Error | null>(null)

  const fetchWebhook = useCallback(async () => {
    if (!id) {
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const data = await webhookAPI.get(id)
      setWebhook(data)
    } catch (err) {
      setError(err as Error)
      setWebhook(null)
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    fetchWebhook()
  }, [fetchWebhook])

  return {
    webhook,
    loading,
    error,
    refetch: fetchWebhook,
  }
}

/**
 * Hook for webhook event history
 */
export function useWebhookEvents(
  webhookId: string | null,
  params?: { page?: number; limit?: number }
) {
  const [events, setEvents] = useState<WebhookEvent[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(!!webhookId)
  const [error, setError] = useState<Error | null>(null)

  const fetchEvents = useCallback(async () => {
    if (!webhookId) {
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const response = await webhookAPI.getEvents(webhookId, params)
      setEvents(response.events)
      setTotal(response.total)
    } catch (err) {
      setError(err as Error)
      setEvents([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [webhookId, params])

  useEffect(() => {
    fetchEvents()
  }, [fetchEvents])

  return {
    events,
    total,
    loading,
    error,
    refetch: fetchEvents,
  }
}

/**
 * Hook for webhook CRUD mutations
 */
export function useWebhookMutations() {
  const [creating, setCreating] = useState(false)
  const [updating, setUpdating] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [regenerating, setRegenerating] = useState(false)
  const [testing, setTesting] = useState(false)

  const createWebhook = async (input: WebhookCreateInput): Promise<Webhook> => {
    try {
      setCreating(true)
      const webhook = await webhookAPI.create(input)
      return webhook
    } finally {
      setCreating(false)
    }
  }

  const updateWebhook = async (
    id: string,
    updates: WebhookUpdateInput
  ): Promise<Webhook> => {
    try {
      setUpdating(true)
      const webhook = await webhookAPI.update(id, updates)
      return webhook
    } finally {
      setUpdating(false)
    }
  }

  const deleteWebhook = async (id: string): Promise<void> => {
    try {
      setDeleting(true)
      await webhookAPI.delete(id)
    } finally {
      setDeleting(false)
    }
  }

  const regenerateSecret = async (id: string): Promise<{ secret: string }> => {
    try {
      setRegenerating(true)
      const response = await webhookAPI.regenerateSecret(id)
      return response
    } finally {
      setRegenerating(false)
    }
  }

  const testWebhook = async (
    id: string,
    input?: TestWebhookInput
  ): Promise<TestWebhookResponse> => {
    try {
      setTesting(true)
      const response = await webhookAPI.test(id, input || {})
      return response
    } finally {
      setTesting(false)
    }
  }

  return {
    createWebhook,
    updateWebhook,
    deleteWebhook,
    regenerateSecret,
    testWebhook,
    creating,
    updating,
    deleting,
    regenerating,
    testing,
  }
}
