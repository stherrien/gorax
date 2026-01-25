/**
 * MSW request handlers for API mocking.
 * These handlers intercept API requests and return mock responses.
 */

import { http, HttpResponse, delay } from 'msw'
import {
  mockSchedules,
  mockWorkflows,
  createMockSchedule,
  wrapData,
  wrapPaginated,
  wrapError,
} from './data'

const API_BASE = '/api/v1'

// Schedule handlers
export const scheduleHandlers = [
  // List all schedules
  http.get(`${API_BASE}/schedules`, async ({ request }) => {
    const url = new URL(request.url)
    const limit = parseInt(url.searchParams.get('limit') || '10', 10)
    const offset = parseInt(url.searchParams.get('offset') || '0', 10)

    await delay(50) // Simulate network delay

    return HttpResponse.json(
      wrapPaginated(mockSchedules.list, limit, offset)
    )
  }),

  // List schedules for a workflow
  http.get(`${API_BASE}/workflows/:workflowId/schedules`, async ({ params }) => {
    const { workflowId } = params

    await delay(50)

    const schedules = mockSchedules.list.filter(
      (s) => s.workflow_id === workflowId
    )

    return HttpResponse.json(wrapPaginated(schedules))
  }),

  // Get single schedule
  http.get(`${API_BASE}/schedules/:id`, async ({ params }) => {
    const { id } = params

    await delay(50)

    const schedule = mockSchedules.list.find((s) => s.id === id)
    if (!schedule) {
      return HttpResponse.json(
        wrapError('schedule not found', 'not_found'),
        { status: 404 }
      )
    }

    return HttpResponse.json(wrapData(schedule))
  }),

  // Create schedule
  http.post(`${API_BASE}/workflows/:workflowId/schedules`, async ({ request, params }) => {
    const body = await request.json() as Record<string, unknown>
    const { workflowId } = params

    await delay(100)

    // Validation
    if (!body.name || typeof body.name !== 'string') {
      return HttpResponse.json(
        wrapError('name is required', 'validation_error'),
        { status: 400 }
      )
    }

    if (!body.cron_expression || typeof body.cron_expression !== 'string') {
      return HttpResponse.json(
        wrapError('cron_expression is required', 'validation_error'),
        { status: 400 }
      )
    }

    const newSchedule = createMockSchedule({
      workflow_id: workflowId as string,
      name: body.name as string,
      cron_expression: body.cron_expression as string,
      timezone: (body.timezone as string) || 'UTC',
      enabled: body.enabled !== false,
    })

    return HttpResponse.json(wrapData(newSchedule), { status: 201 })
  }),

  // Update schedule
  http.put(`${API_BASE}/schedules/:id`, async ({ request, params }) => {
    const body = await request.json() as Record<string, unknown>
    const { id } = params

    await delay(100)

    const schedule = mockSchedules.list.find((s) => s.id === id)
    if (!schedule) {
      return HttpResponse.json(
        wrapError('schedule not found', 'not_found'),
        { status: 404 }
      )
    }

    const updated = {
      ...schedule,
      ...body,
      updated_at: new Date().toISOString(),
    }

    return HttpResponse.json(wrapData(updated))
  }),

  // Delete schedule
  http.delete(`${API_BASE}/schedules/:id`, async ({ params }) => {
    const { id } = params

    await delay(100)

    const schedule = mockSchedules.list.find((s) => s.id === id)
    if (!schedule) {
      return HttpResponse.json(
        wrapError('schedule not found', 'not_found'),
        { status: 404 }
      )
    }

    return new HttpResponse(null, { status: 204 })
  }),

  // Parse cron expression
  http.post(`${API_BASE}/schedules/parse-cron`, async ({ request }) => {
    const body = await request.json() as Record<string, unknown>

    await delay(50)

    const cron = body.cron_expression as string
    if (!cron) {
      return HttpResponse.json(
        wrapError('cron_expression is required', 'validation_error'),
        { status: 400 }
      )
    }

    // Basic cron validation (5 or 6 parts)
    const parts = cron.trim().split(/\s+/)
    if (parts.length < 5 || parts.length > 6) {
      return HttpResponse.json(
        wrapError('invalid cron expression: must have 5 or 6 parts', 'validation_error'),
        { status: 400 }
      )
    }

    const nextRun = new Date(Date.now() + 3600000).toISOString()

    return HttpResponse.json({
      valid: true,
      next_run: nextRun,
    })
  }),

  // Preview schedule
  http.post(`${API_BASE}/schedules/preview`, async ({ request }) => {
    const body = await request.json() as Record<string, unknown>

    await delay(50)

    const count = Math.min((body.count as number) || 10, 50)
    const nextRuns: string[] = []

    for (let i = 0; i < count; i++) {
      nextRuns.push(new Date(Date.now() + (i + 1) * 3600000).toISOString())
    }

    return HttpResponse.json({
      valid: true,
      next_runs: nextRuns,
      count: nextRuns.length,
      timezone: (body.timezone as string) || 'UTC',
    })
  }),
]

// Workflow handlers
export const workflowHandlers = [
  // List workflows
  http.get(`${API_BASE}/workflows`, async ({ request }) => {
    const url = new URL(request.url)
    const limit = parseInt(url.searchParams.get('limit') || '10', 10)
    const offset = parseInt(url.searchParams.get('offset') || '0', 10)

    await delay(50)

    return HttpResponse.json(wrapPaginated(mockWorkflows.list, limit, offset))
  }),

  // Get single workflow
  http.get(`${API_BASE}/workflows/:id`, async ({ params }) => {
    const { id } = params

    await delay(50)

    const workflow = mockWorkflows.list.find((w) => w.id === id)
    if (!workflow) {
      return HttpResponse.json(
        wrapError('workflow not found', 'not_found'),
        { status: 404 }
      )
    }

    return HttpResponse.json(wrapData(workflow))
  }),
]

// Auth handlers
export const authHandlers = [
  // CSRF token
  http.get(`${API_BASE}/auth/csrf`, async () => {
    await delay(20)
    return HttpResponse.json({ token: 'mock-csrf-token' })
  }),

  // Current user
  http.get(`${API_BASE}/auth/me`, async () => {
    await delay(50)
    return HttpResponse.json(
      wrapData({
        id: 'user-1',
        email: 'test@example.com',
        name: 'Test User',
        role: 'admin',
        tenant_id: 'tenant-1',
      })
    )
  }),
]

// Error simulation handlers (for testing error states)
export const errorHandlers = {
  // Returns 500 for any matching route
  serverError: (path: string) =>
    http.all(`${API_BASE}${path}`, async () => {
      await delay(50)
      return HttpResponse.json(
        wrapError('internal server error', 'internal_error'),
        { status: 500 }
      )
    }),

  // Returns 401 for any matching route
  unauthorized: (path: string) =>
    http.all(`${API_BASE}${path}`, async () => {
      await delay(50)
      return HttpResponse.json(
        wrapError('unauthorized', 'unauthorized'),
        { status: 401 }
      )
    }),

  // Returns 429 for any matching route
  rateLimited: (path: string) =>
    http.all(`${API_BASE}${path}`, async () => {
      await delay(50)
      return HttpResponse.json(
        wrapError('rate limit exceeded', 'rate_limit_exceeded'),
        { status: 429 }
      )
    }),

  // Network error (no response)
  networkError: (path: string) =>
    http.all(`${API_BASE}${path}`, async () => {
      throw new Error('Network error')
    }),
}

// All default handlers
export const handlers = [
  ...scheduleHandlers,
  ...workflowHandlers,
  ...authHandlers,
]
