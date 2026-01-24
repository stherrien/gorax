/**
 * Contract Tests for Schedule API
 *
 * These tests verify that the frontend API client correctly handles
 * the actual response format from the backend API.
 *
 * Uses MSW to simulate real backend responses, ensuring type safety
 * and response handling are correct.
 */

import { describe, it, expect, beforeEach } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '../test/mocks/server'
import { scheduleAPI } from './schedules'

// Reset handlers after each test
beforeEach(() => {
  server.resetHandlers()
})

describe('Schedule API Contract Tests', () => {
  describe('list endpoint', () => {
    it('handles paginated response with data wrapper', async () => {
      // Backend response format from response.Paginated()
      const backendResponse = {
        data: [
          {
            id: 'sched-1',
            tenant_id: 'tenant-1',
            workflow_id: 'wf-1',
            name: 'Daily Backup',
            cron_expression: '0 0 * * *',
            timezone: 'UTC',
            enabled: true,
            next_run_at: '2025-01-20T00:00:00Z',
            last_run_at: '2025-01-19T00:00:00Z',
            created_at: '2025-01-01T00:00:00Z',
            updated_at: '2025-01-01T00:00:00Z',
          },
        ],
        limit: 20,
        offset: 0,
        total: 1,
      }

      server.use(
        http.get('*/api/v1/schedules', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await scheduleAPI.list()

      // Frontend transforms backend {data, limit, offset, total} to {schedules, total}
      expect(result.schedules).toHaveLength(1)
      expect(result.schedules[0].id).toBe('sched-1')
      expect(result.schedules[0].name).toBe('Daily Backup')
      expect(result.total).toBe(1)
    })

    it('handles empty list response', async () => {
      server.use(
        http.get('*/api/v1/schedules', () => {
          return HttpResponse.json({
            data: [],
            limit: 20,
            offset: 0,
            total: 0,
          })
        })
      )

      const result = await scheduleAPI.list()

      expect(result.schedules).toHaveLength(0)
    })
  })

  describe('get endpoint', () => {
    it('handles single item response with data wrapper', async () => {
      // Backend response format from response.OK()
      const backendResponse = {
        data: {
          id: 'sched-1',
          tenant_id: 'tenant-1',
          workflow_id: 'wf-1',
          name: 'Daily Backup',
          cron_expression: '0 0 * * *',
          timezone: 'UTC',
          enabled: true,
          next_run_at: '2025-01-20T00:00:00Z',
          last_run_at: '2025-01-19T00:00:00Z',
          created_at: '2025-01-01T00:00:00Z',
          updated_at: '2025-01-01T00:00:00Z',
        },
      }

      server.use(
        http.get('*/api/v1/schedules/sched-1', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await scheduleAPI.get('sched-1')

      expect(result.id).toBe('sched-1')
      expect(result.name).toBe('Daily Backup')
      expect(result.cron_expression).toBe('0 0 * * *')
    })

    it('handles 404 not found error', async () => {
      // Backend response format from response.NotFound()
      server.use(
        http.get('*/api/v1/schedules/not-found', () => {
          return HttpResponse.json(
            {
              error: 'schedule not found',
              code: 'not_found',
            },
            { status: 404 }
          )
        })
      )

      await expect(scheduleAPI.get('not-found')).rejects.toThrow()
    })
  })

  describe('create endpoint', () => {
    it('handles created response with data wrapper', async () => {
      // Backend response format from response.Created()
      const backendResponse = {
        data: {
          id: 'sched-new',
          tenant_id: 'tenant-1',
          workflow_id: 'wf-1',
          name: 'New Schedule',
          cron_expression: '0 9 * * *',
          timezone: 'UTC',
          enabled: true,
          next_run_at: '2025-01-20T09:00:00Z',
          last_run_at: null,
          created_at: '2025-01-20T00:00:00Z',
          updated_at: '2025-01-20T00:00:00Z',
        },
      }

      server.use(
        http.post('*/api/v1/workflows/wf-1/schedules', () => {
          return HttpResponse.json(backendResponse, { status: 201 })
        })
      )

      const result = await scheduleAPI.create('wf-1', {
        name: 'New Schedule',
        cronExpression: '0 9 * * *',
        timezone: 'UTC',
        enabled: true,
      })

      expect(result.id).toBe('sched-new')
      expect(result.name).toBe('New Schedule')
    })

    it('handles validation error', async () => {
      // Backend response format from response.BadRequest()
      server.use(
        http.post('*/api/v1/workflows/wf-1/schedules', () => {
          return HttpResponse.json(
            {
              error: 'invalid cron expression',
              code: 'bad_request',
            },
            { status: 400 }
          )
        })
      )

      await expect(
        scheduleAPI.create('wf-1', {
          name: 'Bad Schedule',
          cronExpression: 'invalid',
          timezone: 'UTC',
          enabled: true,
        })
      ).rejects.toThrow()
    })
  })

  describe('update endpoint', () => {
    it('handles updated response with data wrapper', async () => {
      const backendResponse = {
        data: {
          id: 'sched-1',
          tenant_id: 'tenant-1',
          workflow_id: 'wf-1',
          name: 'Updated Schedule',
          cron_expression: '0 9 * * 1-5',
          timezone: 'America/New_York',
          enabled: false,
          next_run_at: null,
          last_run_at: '2025-01-19T00:00:00Z',
          created_at: '2025-01-01T00:00:00Z',
          updated_at: '2025-01-20T00:00:00Z',
        },
      }

      server.use(
        http.put('*/api/v1/schedules/sched-1', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await scheduleAPI.update('sched-1', {
        name: 'Updated Schedule',
        enabled: false,
      })

      expect(result.name).toBe('Updated Schedule')
      expect(result.enabled).toBe(false)
    })
  })

  describe('delete endpoint', () => {
    it('handles 204 no content response', async () => {
      server.use(
        http.delete('*/api/v1/schedules/sched-1', () => {
          return new HttpResponse(null, { status: 204 })
        })
      )

      // Should not throw
      await expect(scheduleAPI.delete('sched-1')).resolves.not.toThrow()
    })
  })

  describe('parseCron endpoint', () => {
    it('handles valid cron response', async () => {
      // Note: parseCron uses response.JSON() directly, not response.OK()
      const backendResponse = {
        valid: true,
        next_run: '2025-01-20T00:00:00Z',
      }

      server.use(
        http.post('*/api/v1/schedules/parse-cron', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await scheduleAPI.parseCron('0 0 * * *', 'UTC')

      expect(result.valid).toBe(true)
      expect(result.next_run).toBe('2025-01-20T00:00:00Z')
    })

    it('handles invalid cron error', async () => {
      server.use(
        http.post('*/api/v1/schedules/parse-cron', () => {
          return HttpResponse.json(
            {
              error: 'invalid cron expression: invalid value',
              code: 'bad_request',
            },
            { status: 400 }
          )
        })
      )

      await expect(scheduleAPI.parseCron('invalid', 'UTC')).rejects.toThrow()
    })
  })

  describe('preview endpoint', () => {
    it('handles preview response with next runs', async () => {
      // Note: preview uses response.JSON() directly
      const backendResponse = {
        valid: true,
        next_runs: [
          '2025-01-20T09:00:00Z',
          '2025-01-21T09:00:00Z',
          '2025-01-22T09:00:00Z',
        ],
        count: 3,
        timezone: 'UTC',
      }

      server.use(
        http.post('*/api/v1/schedules/preview', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await scheduleAPI.preview('0 9 * * *', 'UTC', 3)

      expect(result.valid).toBe(true)
      expect(result.next_runs).toHaveLength(3)
      expect(result.count).toBe(3)
      expect(result.timezone).toBe('UTC')
    })
  })

  describe('error response format', () => {
    it('handles standardized error response with code', async () => {
      server.use(
        http.get('*/api/v1/schedules/sched-1', () => {
          return HttpResponse.json(
            {
              error: 'schedule not found',
              code: 'not_found',
            },
            { status: 404 }
          )
        })
      )

      try {
        await scheduleAPI.get('sched-1')
        expect.fail('Should have thrown')
      } catch (error: unknown) {
        // The error should contain the backend error message
        expect((error as Error).message).toContain('schedule not found')
      }
    })

    it('handles validation error with details', async () => {
      server.use(
        http.post('*/api/v1/workflows/wf-1/schedules', () => {
          return HttpResponse.json(
            {
              error: 'cron_expression is required',
              code: 'validation_error',
              details: {
                field: 'cron_expression',
              },
            },
            { status: 400 }
          )
        })
      )

      try {
        await scheduleAPI.create('wf-1', {
          name: 'Test',
          cronExpression: '',
          timezone: 'UTC',
          enabled: true,
        })
        expect.fail('Should have thrown')
      } catch (error: unknown) {
        expect((error as Error).message).toContain('cron_expression')
      }
    })

    it('handles internal server error', async () => {
      server.use(
        http.get('*/api/v1/schedules', () => {
          return HttpResponse.json(
            {
              error: 'database connection failed',
              code: 'internal_error',
            },
            { status: 500 }
          )
        })
      )

      await expect(scheduleAPI.list()).rejects.toThrow()
    })
  })
})
