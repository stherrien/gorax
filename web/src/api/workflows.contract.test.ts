/**
 * Contract Tests for Workflow API
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
import { workflowAPI } from './workflows'

// Reset handlers after each test
beforeEach(() => {
  server.resetHandlers()
})

describe('Workflow API Contract Tests', () => {
  describe('list endpoint', () => {
    it('handles paginated response with data wrapper', async () => {
      // Backend response format from List handler
      const backendResponse = {
        data: [
          {
            id: 'wf-1',
            tenant_id: 'tenant-1',
            name: 'Test Workflow',
            description: 'A test workflow',
            status: 'active',
            definition: { nodes: [], edges: [] },
            version: 1,
            created_at: '2025-01-01T00:00:00Z',
            updated_at: '2025-01-01T00:00:00Z',
          },
        ],
        limit: 20,
        offset: 0,
      }

      server.use(
        http.get('*/api/v1/workflows', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await workflowAPI.list()

      // Frontend transforms backend {data, limit, offset} to {workflows, total}
      expect(result.workflows).toHaveLength(1)
      expect(result.workflows[0].id).toBe('wf-1')
      expect(result.workflows[0].name).toBe('Test Workflow')
      expect(result.total).toBe(1)
    })

    it('handles empty list response', async () => {
      server.use(
        http.get('*/api/v1/workflows', () => {
          return HttpResponse.json({
            data: [],
            limit: 20,
            offset: 0,
          })
        })
      )

      const result = await workflowAPI.list()

      expect(result.workflows).toHaveLength(0)
    })
  })

  describe('get endpoint', () => {
    it('handles single item response with data wrapper', async () => {
      // Backend response format from Get handler
      const backendResponse = {
        data: {
          id: 'wf-1',
          tenant_id: 'tenant-1',
          name: 'Test Workflow',
          description: 'A test workflow',
          status: 'active',
          definition: {
            nodes: [
              { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: {} },
            ],
            edges: [],
          },
          version: 1,
          created_at: '2025-01-01T00:00:00Z',
          updated_at: '2025-01-01T00:00:00Z',
        },
      }

      server.use(
        http.get('*/api/v1/workflows/wf-1', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await workflowAPI.get('wf-1')

      expect(result.id).toBe('wf-1')
      expect(result.name).toBe('Test Workflow')
      expect(result.definition.nodes).toHaveLength(1)
    })

    it('handles 404 not found error', async () => {
      server.use(
        http.get('*/api/v1/workflows/not-found', () => {
          return HttpResponse.json(
            {
              error: 'workflow not found',
              code: 'not_found',
            },
            { status: 404 }
          )
        })
      )

      await expect(workflowAPI.get('not-found')).rejects.toThrow()
    })
  })

  describe('create endpoint', () => {
    it('handles created response with data wrapper', async () => {
      const backendResponse = {
        data: {
          id: 'wf-new',
          tenant_id: 'tenant-1',
          name: 'New Workflow',
          description: 'A new workflow',
          status: 'draft',
          definition: { nodes: [], edges: [] },
          version: 1,
          created_at: '2025-01-20T00:00:00Z',
          updated_at: '2025-01-20T00:00:00Z',
        },
      }

      server.use(
        http.post('*/api/v1/workflows', () => {
          return HttpResponse.json(backendResponse, { status: 201 })
        })
      )

      const result = await workflowAPI.create({
        name: 'New Workflow',
        description: 'A new workflow',
        definition: { nodes: [], edges: [] },
      })

      expect(result.id).toBe('wf-new')
      expect(result.name).toBe('New Workflow')
      expect(result.status).toBe('draft')
    })

    it('handles validation error', async () => {
      server.use(
        http.post('*/api/v1/workflows', () => {
          return HttpResponse.json(
            {
              error: 'name is required',
              code: 'bad_request',
            },
            { status: 400 }
          )
        })
      )

      await expect(
        workflowAPI.create({
          name: '',
          definition: { nodes: [], edges: [] },
        })
      ).rejects.toThrow()
    })
  })

  describe('update endpoint', () => {
    it('handles updated response with data wrapper', async () => {
      const backendResponse = {
        data: {
          id: 'wf-1',
          tenant_id: 'tenant-1',
          name: 'Updated Workflow',
          description: 'Updated description',
          status: 'active',
          definition: { nodes: [], edges: [] },
          version: 2,
          created_at: '2025-01-01T00:00:00Z',
          updated_at: '2025-01-20T00:00:00Z',
        },
      }

      server.use(
        http.put('*/api/v1/workflows/wf-1', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await workflowAPI.update('wf-1', {
        name: 'Updated Workflow',
        description: 'Updated description',
      })

      expect(result.name).toBe('Updated Workflow')
      expect(result.version).toBe(2)
    })
  })

  describe('delete endpoint', () => {
    it('handles 204 no content response', async () => {
      server.use(
        http.delete('*/api/v1/workflows/wf-1', () => {
          return new HttpResponse(null, { status: 204 })
        })
      )

      // Should not throw
      await expect(workflowAPI.delete('wf-1')).resolves.not.toThrow()
    })
  })

  describe('execute endpoint', () => {
    it('handles execution response', async () => {
      const backendResponse = {
        execution_id: 'exec-123',
        workflow_id: 'wf-1',
        status: 'queued',
        queued_at: '2025-01-20T00:00:00Z',
      }

      server.use(
        http.post('*/api/v1/workflows/wf-1/execute', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await workflowAPI.execute('wf-1', { input: 'test' })

      expect(result.executionId || result.execution_id).toBeDefined()
      expect(result.status).toBe('queued')
    })
  })

  describe('versions endpoint', () => {
    it('handles list versions response', async () => {
      const backendResponse = {
        data: [
          {
            id: 'ver-1',
            workflow_id: 'wf-1',
            version: 1,
            definition: { nodes: [], edges: [] },
            created_by: 'user-1',
            created_at: '2025-01-01T00:00:00Z',
          },
          {
            id: 'ver-2',
            workflow_id: 'wf-1',
            version: 2,
            definition: { nodes: [], edges: [] },
            created_by: 'user-1',
            created_at: '2025-01-10T00:00:00Z',
          },
        ],
      }

      server.use(
        http.get('*/api/v1/workflows/wf-1/versions', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await workflowAPI.listVersions('wf-1')

      expect(result).toHaveLength(2)
      expect(result[0].version).toBe(1)
      expect(result[1].version).toBe(2)
    })

    it('handles get specific version', async () => {
      const backendResponse = {
        data: {
          id: 'ver-1',
          workflow_id: 'wf-1',
          version: 1,
          definition: { nodes: [], edges: [] },
          created_by: 'user-1',
          created_at: '2025-01-01T00:00:00Z',
        },
      }

      server.use(
        http.get('*/api/v1/workflows/wf-1/versions/1', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await workflowAPI.getVersion('wf-1', 1)

      expect(result.version).toBe(1)
      expect(result.workflowId || result.workflow_id).toBeDefined()
    })

    it('handles restore version', async () => {
      const backendResponse = {
        data: {
          id: 'wf-1',
          tenant_id: 'tenant-1',
          name: 'Restored Workflow',
          status: 'active',
          definition: { nodes: [], edges: [] },
          version: 3,
          created_at: '2025-01-01T00:00:00Z',
          updated_at: '2025-01-20T00:00:00Z',
        },
      }

      server.use(
        http.post('*/api/v1/workflows/wf-1/versions/1/restore', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await workflowAPI.restoreVersion('wf-1', 1)

      expect(result.version).toBe(3)
    })
  })

  describe('dry-run endpoint', () => {
    it('handles valid dry-run response', async () => {
      const backendResponse = {
        data: {
          valid: true,
          execution_order: ['node-1', 'node-2'],
          variable_mapping: { input: 'output' },
          warnings: [],
          errors: [],
        },
      }

      server.use(
        http.post('*/api/v1/workflows/wf-1/dry-run', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await workflowAPI.dryRun('wf-1', { testInput: 'value' })

      expect(result.valid).toBe(true)
      expect(result.executionOrder || result.execution_order).toHaveLength(2)
    })

    it('handles dry-run with errors', async () => {
      const backendResponse = {
        data: {
          valid: false,
          execution_order: [],
          variable_mapping: {},
          warnings: [{ nodeId: 'node-1', message: 'Deprecated action' }],
          errors: [{ nodeId: 'node-2', field: 'url', message: 'URL is required' }],
        },
      }

      server.use(
        http.post('*/api/v1/workflows/wf-1/dry-run', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await workflowAPI.dryRun('wf-1')

      expect(result.valid).toBe(false)
      expect(result.errors).toHaveLength(1)
      expect(result.warnings).toHaveLength(1)
    })
  })

  describe('bulk operations', () => {
    it('handles bulk delete response', async () => {
      const backendResponse = {
        success_count: 2,
        failures: [],
      }

      server.use(
        http.post('*/api/v1/workflows/bulk/delete', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await workflowAPI.bulkDelete(['wf-1', 'wf-2'])

      expect(result.success_count).toBe(2)
      expect(result.failures).toHaveLength(0)
    })

    it('handles bulk operation with partial failures', async () => {
      const backendResponse = {
        success_count: 1,
        failures: [
          { workflow_id: 'wf-2', error: 'workflow not found' },
        ],
      }

      server.use(
        http.post('*/api/v1/workflows/bulk/delete', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await workflowAPI.bulkDelete(['wf-1', 'wf-2'])

      expect(result.success_count).toBe(1)
      expect(result.failures).toHaveLength(1)
      expect(result.failures[0].workflow_id).toBe('wf-2')
    })

    it('handles bulk enable response', async () => {
      const backendResponse = {
        success_count: 3,
        failures: [],
      }

      server.use(
        http.post('*/api/v1/workflows/bulk/enable', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await workflowAPI.bulkEnable(['wf-1', 'wf-2', 'wf-3'])

      expect(result.success_count).toBe(3)
    })

    it('handles bulk disable response', async () => {
      const backendResponse = {
        success_count: 2,
        failures: [],
      }

      server.use(
        http.post('*/api/v1/workflows/bulk/disable', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await workflowAPI.bulkDisable(['wf-1', 'wf-2'])

      expect(result.success_count).toBe(2)
    })
  })

  describe('error response format', () => {
    it('handles standardized error response', async () => {
      server.use(
        http.get('*/api/v1/workflows/wf-1', () => {
          return HttpResponse.json(
            {
              error: 'workflow not found',
              code: 'not_found',
            },
            { status: 404 }
          )
        })
      )

      try {
        await workflowAPI.get('wf-1')
        expect.fail('Should have thrown')
      } catch (error: unknown) {
        expect((error as Error).message).toContain('workflow not found')
      }
    })

    it('handles internal server error', async () => {
      server.use(
        http.get('*/api/v1/workflows', () => {
          return HttpResponse.json(
            {
              error: 'database connection failed',
              code: 'internal_error',
            },
            { status: 500 }
          )
        })
      )

      await expect(workflowAPI.list()).rejects.toThrow()
    })
  })
})
