/**
 * Contract Tests for Credentials API
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
import { credentialAPI } from './credentials'

// Reset handlers after each test
beforeEach(() => {
  server.resetHandlers()
})

describe('Credential API Contract Tests', () => {
  describe('list endpoint', () => {
    it('handles paginated response', async () => {
      const backendResponse = {
        data: [
          {
            id: 'cred-1',
            tenant_id: 'tenant-1',
            name: 'API Key Credential',
            type: 'api_key',
            description: 'Test API key',
            created_at: '2025-01-01T00:00:00Z',
            updated_at: '2025-01-01T00:00:00Z',
          },
          {
            id: 'cred-2',
            tenant_id: 'tenant-1',
            name: 'OAuth Credential',
            type: 'oauth2',
            created_at: '2025-01-02T00:00:00Z',
            updated_at: '2025-01-02T00:00:00Z',
          },
        ],
        limit: 20,
        offset: 0,
      }

      server.use(
        http.get('*/api/v1/credentials', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await credentialAPI.list()

      expect(result.data).toHaveLength(2)
      expect(result.data[0].id).toBe('cred-1')
      expect(result.data[0].name).toBe('API Key Credential')
      expect(result.limit).toBe(20)
      expect(result.offset).toBe(0)
    })

    it('handles empty list response', async () => {
      server.use(
        http.get('*/api/v1/credentials', () => {
          return HttpResponse.json({
            data: [],
            limit: 20,
            offset: 0,
          })
        })
      )

      const result = await credentialAPI.list()

      expect(result.data).toHaveLength(0)
    })

    it('handles filtered list with type parameter', async () => {
      const backendResponse = {
        data: [
          {
            id: 'cred-1',
            tenant_id: 'tenant-1',
            name: 'API Key',
            type: 'api_key',
            created_at: '2025-01-01T00:00:00Z',
            updated_at: '2025-01-01T00:00:00Z',
          },
        ],
        limit: 20,
        offset: 0,
      }

      server.use(
        http.get('*/api/v1/credentials', ({ request }) => {
          const url = new URL(request.url)
          const type = url.searchParams.get('type')
          if (type === 'api_key') {
            return HttpResponse.json(backendResponse)
          }
          return HttpResponse.json({ data: [], limit: 20, offset: 0 })
        })
      )

      const result = await credentialAPI.list({ type: 'api_key' })

      expect(result.data).toHaveLength(1)
      expect(result.data[0].type).toBe('api_key')
    })
  })

  describe('get endpoint', () => {
    it('handles single item response with data wrapper', async () => {
      const backendResponse = {
        data: {
          id: 'cred-1',
          tenant_id: 'tenant-1',
          name: 'API Key Credential',
          type: 'api_key',
          description: 'Test API key',
          expires_at: '2026-01-01T00:00:00Z',
          created_at: '2025-01-01T00:00:00Z',
          updated_at: '2025-01-01T00:00:00Z',
        },
      }

      server.use(
        http.get('*/api/v1/credentials/cred-1', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await credentialAPI.get('cred-1')

      expect(result.id).toBe('cred-1')
      expect(result.name).toBe('API Key Credential')
      expect(result.type).toBe('api_key')
    })

    it('handles 404 not found error', async () => {
      server.use(
        http.get('*/api/v1/credentials/not-found', () => {
          return HttpResponse.json(
            {
              error: 'credential not found',
              code: 'not_found',
            },
            { status: 404 }
          )
        })
      )

      await expect(credentialAPI.get('not-found')).rejects.toThrow()
    })
  })

  describe('create endpoint', () => {
    it('handles created response with data wrapper', async () => {
      const backendResponse = {
        data: {
          id: 'cred-new',
          tenant_id: 'tenant-1',
          name: 'New API Key',
          type: 'api_key',
          description: 'A new API key',
          created_at: '2025-01-20T00:00:00Z',
          updated_at: '2025-01-20T00:00:00Z',
        },
      }

      server.use(
        http.post('*/api/v1/credentials', () => {
          return HttpResponse.json(backendResponse, { status: 201 })
        })
      )

      const result = await credentialAPI.create({
        name: 'New API Key',
        type: 'api_key',
        description: 'A new API key',
        value: { apiKey: 'secret-key' },
      })

      expect(result.id).toBe('cred-new')
      expect(result.name).toBe('New API Key')
    })

    it('handles validation error for missing required fields', async () => {
      server.use(
        http.post('*/api/v1/credentials', () => {
          return HttpResponse.json(
            {
              error: 'name is required',
              code: 'validation_error',
            },
            { status: 400 }
          )
        })
      )

      await expect(
        credentialAPI.create({
          name: '',
          type: 'api_key',
          value: { apiKey: 'test' },
        })
      ).rejects.toThrow()
    })
  })

  describe('update endpoint', () => {
    it('handles updated response with data wrapper', async () => {
      const backendResponse = {
        data: {
          id: 'cred-1',
          tenant_id: 'tenant-1',
          name: 'Updated Credential',
          type: 'api_key',
          description: 'Updated description',
          created_at: '2025-01-01T00:00:00Z',
          updated_at: '2025-01-20T00:00:00Z',
        },
      }

      server.use(
        http.put('*/api/v1/credentials/cred-1', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await credentialAPI.update('cred-1', {
        name: 'Updated Credential',
        description: 'Updated description',
      })

      expect(result.name).toBe('Updated Credential')
      expect(result.description).toBe('Updated description')
    })
  })

  describe('delete endpoint', () => {
    it('handles 204 no content response', async () => {
      server.use(
        http.delete('*/api/v1/credentials/cred-1', () => {
          return new HttpResponse(null, { status: 204 })
        })
      )

      // Should not throw
      await expect(credentialAPI.delete('cred-1')).resolves.not.toThrow()
    })
  })

  describe('rotate endpoint', () => {
    it('handles rotated credential response', async () => {
      const backendResponse = {
        data: {
          id: 'cred-1',
          tenant_id: 'tenant-1',
          name: 'API Key',
          type: 'api_key',
          created_at: '2025-01-01T00:00:00Z',
          updated_at: '2025-01-20T00:00:00Z',
        },
      }

      server.use(
        http.post('*/api/v1/credentials/cred-1/rotate', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await credentialAPI.rotate('cred-1', {
        apiKey: 'new-secret-key',
      })

      expect(result.id).toBe('cred-1')
    })
  })

  describe('test endpoint', () => {
    it('handles successful test response', async () => {
      // Backend wraps response in data
      const backendResponse = {
        data: {
          success: true,
          message: 'Credential is valid',
          tested_at: '2025-01-20T00:00:00Z',
        },
      }

      server.use(
        http.post('*/api/v1/credentials/cred-1/test', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await credentialAPI.test('cred-1')

      expect(result.success).toBe(true)
      expect(result.message).toBe('Credential is valid')
    })

    it('handles failed test response', async () => {
      const backendResponse = {
        data: {
          success: false,
          message: 'Authentication failed: invalid credentials',
          tested_at: '2025-01-20T00:00:00Z',
        },
      }

      server.use(
        http.post('*/api/v1/credentials/cred-1/test', () => {
          return HttpResponse.json(backendResponse)
        })
      )

      const result = await credentialAPI.test('cred-1')

      expect(result.success).toBe(false)
      expect(result.message).toContain('Authentication failed')
    })
  })

  describe('error response format', () => {
    it('handles standardized error response', async () => {
      server.use(
        http.get('*/api/v1/credentials/cred-1', () => {
          return HttpResponse.json(
            {
              error: 'credential not found',
              code: 'not_found',
            },
            { status: 404 }
          )
        })
      )

      try {
        await credentialAPI.get('cred-1')
        expect.fail('Should have thrown')
      } catch (error: unknown) {
        expect((error as Error).message).toContain('credential not found')
      }
    })

    it('handles internal server error', async () => {
      server.use(
        http.get('*/api/v1/credentials', () => {
          return HttpResponse.json(
            {
              error: 'database connection failed',
              code: 'internal_error',
            },
            { status: 500 }
          )
        })
      )

      await expect(credentialAPI.list()).rejects.toThrow()
    })
  })
})
