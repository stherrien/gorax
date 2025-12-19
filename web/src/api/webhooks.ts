import { apiClient } from './client'

// Webhook types
export type WebhookAuthType = 'none' | 'signature' | 'basic' | 'api_key'

export interface Webhook {
  id: string
  tenantId: string
  workflowId: string
  name: string
  path: string
  authType: WebhookAuthType
  enabled: boolean
  priority: number
  triggerCount: number
  lastTriggeredAt?: string
  createdAt: string
  updatedAt: string
  url: string
}

export interface WebhookListResponse {
  webhooks: Webhook[]
  total: number
}

export interface WebhookListParams {
  page?: number
  limit?: number
  workflowId?: string
  enabled?: boolean
}

export interface WebhookCreateInput {
  name: string
  workflowId: string
  path: string
  authType?: WebhookAuthType
  description?: string
  priority?: number
}

export interface WebhookUpdateInput {
  name?: string
  path?: string
  authType?: WebhookAuthType
  enabled?: boolean
  description?: string
  priority?: number
}

export interface WebhookEvent {
  id: string
  webhookId: string
  executionId?: string
  requestMethod: string
  requestHeaders: Record<string, string>
  requestBody: unknown
  responseStatus?: number
  processingTimeMs?: number
  status: 'received' | 'processed' | 'filtered' | 'failed'
  errorMessage?: string
  createdAt: string
}

export interface WebhookEventListResponse {
  events: WebhookEvent[]
  total: number
}

export interface TestWebhookInput {
  method?: string
  headers?: Record<string, string>
  body?: unknown
}

export interface TestWebhookResponse {
  success: boolean
  statusCode: number
  responseTimeMs: number
  executionId?: string
  error?: string
}

class WebhookAPI {
  /**
   * List all webhooks
   */
  async list(params?: WebhookListParams): Promise<WebhookListResponse> {
    const options = params ? { params } : undefined
    const response = await apiClient.get('/api/v1/webhooks', options)
    // Backend returns { data: [], limit, offset } or { webhooks: [], total }
    if (response.data && Array.isArray(response.data)) {
      return { webhooks: response.data, total: response.data.length }
    }
    return response
  }

  /**
   * Get a single webhook by ID
   */
  async get(id: string): Promise<Webhook> {
    const response = await apiClient.get(`/api/v1/webhooks/${id}`)
    return response.data || response
  }

  /**
   * Create a new webhook
   */
  async create(webhook: WebhookCreateInput): Promise<Webhook> {
    const response = await apiClient.post('/api/v1/webhooks', webhook)
    return response.data || response
  }

  /**
   * Update an existing webhook
   */
  async update(id: string, updates: WebhookUpdateInput): Promise<Webhook> {
    const response = await apiClient.put(`/api/v1/webhooks/${id}`, updates)
    return response.data || response
  }

  /**
   * Delete a webhook
   */
  async delete(id: string): Promise<void> {
    await apiClient.delete(`/api/v1/webhooks/${id}`)
  }

  /**
   * Regenerate webhook secret
   */
  async regenerateSecret(id: string): Promise<{ secret: string }> {
    return await apiClient.post(`/api/v1/webhooks/${id}/regenerate-secret`, {})
  }

  /**
   * Test webhook with sample data
   */
  async test(id: string, input: TestWebhookInput): Promise<TestWebhookResponse> {
    return await apiClient.post(`/api/v1/webhooks/${id}/test`, input)
  }

  /**
   * Get webhook events history
   */
  async getEvents(
    id: string,
    params?: { page?: number; limit?: number }
  ): Promise<WebhookEventListResponse> {
    const options = params ? { params } : undefined
    const response = await apiClient.get(`/api/v1/webhooks/${id}/events`, options)
    // Backend returns { data: [], limit, offset } or { events: [], total }
    if (response.data && Array.isArray(response.data)) {
      return { events: response.data, total: response.data.length }
    }
    return response
  }

  /**
   * Replay a single webhook event
   */
  async replayEvent(
    eventId: string,
    modifiedPayload?: unknown
  ): Promise<ReplayResult> {
    const body = modifiedPayload ? { modifiedPayload } : {}
    return await apiClient.post(`/api/v1/events/${eventId}/replay`, body)
  }

  /**
   * Batch replay multiple webhook events
   */
  async batchReplayEvents(
    webhookId: string,
    eventIds: string[]
  ): Promise<BatchReplayResponse> {
    return await apiClient.post(`/api/v1/webhooks/${webhookId}/events/replay`, {
      eventIds
    })
  }
}

export const webhookAPI = new WebhookAPI()

// Replay types
export interface ReplayResult {
  success: boolean
  executionId?: string
  error?: string
}

export interface BatchReplayResponse {
  results: Record<string, ReplayResult>
}
