import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import {
  APIClient,
  AuthError,
  NotFoundError,
  ValidationError,
  ServerError,
  NetworkError,
} from './client'

// Mock fetch for these unit tests (bypassing MSW)
const mockFetch = vi.fn()

describe('APIClient', () => {
  let client: APIClient
  let originalFetch: typeof fetch

  beforeEach(() => {
    // Store original fetch and replace with mock
    originalFetch = global.fetch
    global.fetch = mockFetch as typeof fetch
    mockFetch.mockClear()

    client = new APIClient('http://localhost:8080')
    vi.clearAllMocks()
  })

  afterEach(() => {
    // Restore original fetch
    global.fetch = originalFetch
  })

  describe('GET requests', () => {
    it('should make successful GET request', async () => {
      const mockData = { id: '123', name: 'Test' }
      ;mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => mockData,
      })

      const result = await client.get('/test')

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/test',
        expect.objectContaining({
          method: 'GET',
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
          }),
        })
      )
      expect(result).toEqual(mockData)
    })

    it('should include auth token in header when available', async () => {
      (localStorage.getItem as any).mockReturnValueOnce('test-token')
      ;mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({}),
      })

      await client.get('/test')

      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          headers: expect.objectContaining({
            Authorization: 'Bearer test-token',
          }),
        })
      )
    })

    it('should handle query parameters', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({}),
      })

      await client.get('/test', { params: { page: 1, limit: 10 } })

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/test?page=1&limit=10',
        expect.any(Object)
      )
    })
  })

  describe('POST requests', () => {
    it('should make successful POST request with body', async () => {
      const requestBody = { name: 'New Item' }
      const responseData = { id: '123', ...requestBody }
      ;mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 201,
        json: async () => responseData,
      })

      const result = await client.post('/items', requestBody)

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/items',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(requestBody),
        })
      )
      expect(result).toEqual(responseData)
    })
  })

  describe('PUT requests', () => {
    it('should make successful PUT request', async () => {
      const updates = { name: 'Updated' }
      ;mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => updates,
      })

      const result = await client.put('/items/123', updates)

      expect(result).toEqual(updates)
    })
  })

  describe('DELETE requests', () => {
    it('should make successful DELETE request', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 204,
        json: async () => ({}),
      })

      await client.delete('/items/123')

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/items/123',
        expect.objectContaining({ method: 'DELETE' })
      )
    })
  })

  describe('Error handling', () => {
    it('should throw AuthError on 401 response', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 401,
        text: async () => JSON.stringify({ error: 'Unauthorized' }),
      })

      await expect(client.get('/test')).rejects.toThrow(AuthError)
      await expect(client.get('/test')).rejects.toThrow('Unauthorized')
    })

    it('should throw AuthError on 403 response', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 403,
        text: async () => JSON.stringify({ error: 'Forbidden' }),
      })

      await expect(client.get('/test')).rejects.toThrow(AuthError)
    })

    it('should throw NotFoundError on 404 response', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 404,
        text: async () => JSON.stringify({ error: 'Not found' }),
      })

      await expect(client.get('/test')).rejects.toThrow(NotFoundError)
      await expect(client.get('/test')).rejects.toThrow('Not found')
    })

    it('should throw ValidationError on 400 response', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        text: async () => JSON.stringify({ error: 'Invalid input' }),
      })

      await expect(client.post('/test', {})).rejects.toThrow(ValidationError)
    })

    it('should throw ValidationError on 422 response', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 422,
        text: async () => JSON.stringify({ error: 'Validation failed' }),
      })

      await expect(client.post('/test', {})).rejects.toThrow(ValidationError)
    })

    it('should throw ServerError on 500 response', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        text: async () => JSON.stringify({ error: 'Internal server error' }),
      })

      await expect(client.get('/test')).rejects.toThrow(ServerError)
    })

    it('should throw ServerError on 503 response', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 503,
        text: async () => JSON.stringify({ error: 'Service unavailable' }),
      })

      await expect(client.get('/test')).rejects.toThrow(ServerError)
    })

    it('should throw NetworkError when fetch fails', async () => {
      mockFetch.mockRejectedValue(new Error('Network failure'))

      await expect(client.get('/test')).rejects.toThrow(NetworkError)
      await expect(client.get('/test')).rejects.toThrow('Network failure')
    })

    it('should handle non-JSON error responses', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: async () => {
          throw new Error('Invalid JSON')
        },
        text: async () => 'Internal Server Error',
      })

      await expect(client.get('/test')).rejects.toThrow(ServerError)
    })
  })

  describe('Retry logic', () => {
    it('should retry on 503 errors', async () => {
      mockFetch
        .mockResolvedValueOnce({
          ok: false,
          status: 503,
          text: async () => JSON.stringify({ error: 'Service unavailable' }),
        })
        .mockResolvedValueOnce({
          ok: true,
          status: 200,
          json: async () => ({ success: true }),
        })

      const result = await client.get('/test', { retries: 1 })

      expect(mockFetch).toHaveBeenCalledTimes(2)
      expect(result).toEqual({ success: true })
    })

    it('should not retry on 4xx errors', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        text: async () => JSON.stringify({ error: 'Not found' }),
      })

      await expect(client.get('/test', { retries: 3 })).rejects.toThrow(NotFoundError)
      expect(mockFetch).toHaveBeenCalledTimes(1)
    })

    it('should exhaust retry attempts and throw error', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 503,
        text: async () => JSON.stringify({ error: 'Service unavailable' }),
      })

      await expect(client.get('/test', { retries: 2 })).rejects.toThrow(ServerError)
      expect(mockFetch).toHaveBeenCalledTimes(3) // initial + 2 retries
    })
  })

  describe('Timeout handling', () => {
    // Skip: happy-dom doesn't fully support AbortController
    it.skip('should timeout long requests', async () => {
      mockFetch.mockImplementationOnce(
        () =>
          new Promise((resolve) =>
            setTimeout(() => resolve({ ok: true, json: async () => ({}) }), 5000)
          )
      )

      await expect(client.get('/test', { timeout: 50 })).rejects.toThrow('Request timeout')
    })
  })

  describe('Response unwrapping', () => {
    describe('getData', () => {
      it('should unwrap wrapped response', async () => {
        const mockData = { id: '123', name: 'Test' }
        ;mockFetch.mockResolvedValueOnce({
          ok: true,
          status: 200,
          json: async () => ({ data: mockData }),
        })

        const result = await client.getData<{ id: string; name: string }>('/test')

        expect(result).toEqual(mockData)
      })

      it('should return direct response if not wrapped', async () => {
        const mockData = { id: '123', name: 'Test' }
        ;mockFetch.mockResolvedValueOnce({
          ok: true,
          status: 200,
          json: async () => mockData,
        })

        const result = await client.getData<{ id: string; name: string }>('/test')

        expect(result).toEqual(mockData)
      })
    })

    describe('getPaginated', () => {
      it('should return paginated response as-is', async () => {
        const paginatedResponse = {
          data: [{ id: '1' }, { id: '2' }],
          limit: 10,
          offset: 0,
          total: 2,
        }
        ;mockFetch.mockResolvedValueOnce({
          ok: true,
          status: 200,
          json: async () => paginatedResponse,
        })

        const result = await client.getPaginated<{ id: string }>('/test')

        expect(result).toEqual(paginatedResponse)
      })

      it('should wrap plain array in paginated format', async () => {
        const arrayData = [{ id: '1' }, { id: '2' }]
        ;mockFetch.mockResolvedValueOnce({
          ok: true,
          status: 200,
          json: async () => arrayData,
        })

        const result = await client.getPaginated<{ id: string }>('/test')

        expect(result.data).toEqual(arrayData)
        expect(result.limit).toBe(2)
        expect(result.offset).toBe(0)
        expect(result.total).toBe(2)
      })

      it('should handle wrapped array without pagination metadata', async () => {
        const wrappedArray = { data: [{ id: '1' }, { id: '2' }] }
        ;mockFetch.mockResolvedValueOnce({
          ok: true,
          status: 200,
          json: async () => wrappedArray,
        })

        const result = await client.getPaginated<{ id: string }>('/test')

        expect(result.data).toEqual([{ id: '1' }, { id: '2' }])
        expect(result.limit).toBe(2)
        expect(result.offset).toBe(0)
      })

      it('should throw error for unexpected response format', async () => {
        mockFetch.mockResolvedValueOnce({
          ok: true,
          status: 200,
          json: async () => ({ unexpected: 'format' }),
        })

        await expect(client.getPaginated('/test')).rejects.toThrow(
          'Unexpected response format for paginated request'
        )
      })
    })

    describe('postData', () => {
      it('should unwrap wrapped POST response', async () => {
        const mockData = { id: '123', name: 'Created' }
        ;mockFetch.mockResolvedValueOnce({
          ok: true,
          status: 201,
          json: async () => ({ data: mockData }),
        })

        const result = await client.postData<{ id: string; name: string }>('/test', {
          name: 'Created',
        })

        expect(result).toEqual(mockData)
      })
    })

    describe('putData', () => {
      it('should unwrap wrapped PUT response', async () => {
        const mockData = { id: '123', name: 'Updated' }
        ;mockFetch.mockResolvedValueOnce({
          ok: true,
          status: 200,
          json: async () => ({ data: mockData }),
        })

        const result = await client.putData<{ id: string; name: string }>('/test/123', {
          name: 'Updated',
        })

        expect(result).toEqual(mockData)
      })
    })

    describe('patchData', () => {
      it('should unwrap wrapped PATCH response', async () => {
        const mockData = { id: '123', name: 'Patched' }
        ;mockFetch.mockResolvedValueOnce({
          ok: true,
          status: 200,
          json: async () => ({ data: mockData }),
        })

        const result = await client.patchData<{ id: string; name: string }>('/test/123', {
          name: 'Patched',
        })

        expect(result).toEqual(mockData)
      })
    })
  })
})
