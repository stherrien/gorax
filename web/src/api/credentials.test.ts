import { describe, it, expect, beforeEach, vi } from 'vitest'
import { credentialAPI } from './credentials'
import type { Credential, CredentialCreateInput, CredentialUpdateInput } from './credentials'

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

describe('Credential API', () => {
  const mockCredential: Credential = {
    id: 'cred-123',
    tenantId: 'tenant-1',
    name: 'My API Key',
    type: 'api_key',
    description: 'Production API key',
    expiresAt: '2025-12-31T23:59:59Z',
    createdAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-01-15T10:00:00Z',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('should fetch list of credentials', async () => {
      const mockCredentials = [mockCredential]
      ;(apiClient.get as any).mockResolvedValueOnce({ credentials: mockCredentials, total: 1 })

      const result = await credentialAPI.list()

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/credentials', undefined)
      expect(result).toEqual({ credentials: mockCredentials, total: 1 })
    })

    it('should handle empty list', async () => {
      (apiClient.get as any).mockResolvedValueOnce({ credentials: [], total: 0 })

      const result = await credentialAPI.list()

      expect(result.credentials).toEqual([])
      expect(result.total).toBe(0)
    })

    it('should support pagination parameters', async () => {
      (apiClient.get as any).mockResolvedValueOnce({ credentials: [], total: 0 })

      await credentialAPI.list({ page: 2, limit: 20 })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/credentials', {
        params: { page: 2, limit: 20 },
      })
    })

    it('should support type filter', async () => {
      (apiClient.get as any).mockResolvedValueOnce({ credentials: [], total: 0 })

      await credentialAPI.list({ type: 'oauth2' })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/credentials', {
        params: { type: 'oauth2' },
      })
    })

    it('should support search parameter', async () => {
      (apiClient.get as any).mockResolvedValueOnce({ credentials: [], total: 0 })

      await credentialAPI.list({ search: 'production' })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/credentials', {
        params: { search: 'production' },
      })
    })
  })

  describe('get', () => {
    it('should fetch single credential by ID', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockCredential)

      const result = await credentialAPI.get('cred-123')

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/credentials/cred-123')
      expect(result).toEqual(mockCredential)
    })

    it('should throw NotFoundError for invalid ID', async () => {
      const error = new Error('Not found')
      error.name = 'NotFoundError'
      ;(apiClient.get as any).mockRejectedValueOnce(error)

      await expect(credentialAPI.get('invalid-id')).rejects.toThrow('Not found')
    })
  })

  describe('create', () => {
    it('should create new API key credential', async () => {
      const createInput: CredentialCreateInput = {
        name: 'New API Key',
        type: 'api_key',
        description: 'Test API key',
        value: {
          apiKey: 'sk-test-123456',
        },
      }

      const createdCredential = { ...mockCredential, ...createInput }
      ;(apiClient.post as any).mockResolvedValueOnce(createdCredential)

      const result = await credentialAPI.create(createInput)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/credentials', createInput)
      expect(result).toEqual(createdCredential)
      expect(result.id).toBeDefined()
    })

    it('should create OAuth2 credential', async () => {
      const createInput: CredentialCreateInput = {
        name: 'OAuth App',
        type: 'oauth2',
        value: {
          clientId: 'client-123',
          clientSecret: 'secret-456',
          authUrl: 'https://auth.example.com',
          tokenUrl: 'https://token.example.com',
        },
      }

      ;(apiClient.post as any).mockResolvedValueOnce({ ...mockCredential, ...createInput })

      const result = await credentialAPI.create(createInput)

      expect(result.type).toBe('oauth2')
    })

    it('should create basic auth credential', async () => {
      const createInput: CredentialCreateInput = {
        name: 'Basic Auth',
        type: 'basic_auth',
        value: {
          username: 'user123',
          password: 'pass456',
        },
      }

      ;(apiClient.post as any).mockResolvedValueOnce({ ...mockCredential, ...createInput })

      const result = await credentialAPI.create(createInput)

      expect(result.type).toBe('basic_auth')
    })

    it('should validate required fields', async () => {
      const error = new Error('Name is required')
      error.name = 'ValidationError'
      ;(apiClient.post as any).mockRejectedValueOnce(error)

      const invalidInput = {
        name: '',
        type: 'api_key',
        value: {},
      } as CredentialCreateInput

      await expect(credentialAPI.create(invalidInput)).rejects.toThrow('Name is required')
    })

    it('should support optional expiration date', async () => {
      const createInput: CredentialCreateInput = {
        name: 'Expiring Key',
        type: 'api_key',
        value: { apiKey: 'key-123' },
        expiresAt: '2025-12-31T23:59:59Z',
      }

      ;(apiClient.post as any).mockResolvedValueOnce({ ...mockCredential, ...createInput })

      const result = await credentialAPI.create(createInput)

      expect(result.expiresAt).toBe('2025-12-31T23:59:59Z')
    })
  })

  describe('update', () => {
    it('should update credential metadata', async () => {
      const updates: CredentialUpdateInput = {
        name: 'Updated Name',
        description: 'Updated description',
      }

      const updatedCredential = { ...mockCredential, ...updates }
      ;(apiClient.put as any).mockResolvedValueOnce(updatedCredential)

      const result = await credentialAPI.update('cred-123', updates)

      expect(apiClient.put).toHaveBeenCalledWith('/api/v1/credentials/cred-123', updates)
      expect(result.name).toBe('Updated Name')
      expect(result.description).toBe('Updated description')
    })

    it('should update expiration date', async () => {
      const updates: CredentialUpdateInput = {
        expiresAt: '2026-12-31T23:59:59Z',
      }

      const updatedCredential = { ...mockCredential, ...updates }
      ;(apiClient.put as any).mockResolvedValueOnce(updatedCredential)

      const result = await credentialAPI.update('cred-123', updates)

      expect(result.expiresAt).toBe('2026-12-31T23:59:59Z')
    })

    it('should not allow updating credential value', async () => {
      // Value updates should go through rotate endpoint
      const updates: CredentialUpdateInput = {
        name: 'Test',
      }

      ;(apiClient.put as any).mockResolvedValueOnce({ ...mockCredential, ...updates })

      const result = await credentialAPI.update('cred-123', updates)

      expect(result).toBeDefined()
      // Value field should not be in update input type
    })

    it('should throw NotFoundError for non-existent credential', async () => {
      const error = new Error('Credential not found')
      error.name = 'NotFoundError'
      ;(apiClient.put as any).mockRejectedValueOnce(error)

      await expect(credentialAPI.update('invalid-id', { name: 'Test' })).rejects.toThrow(
        'Credential not found'
      )
    })
  })

  describe('delete', () => {
    it('should delete credential by ID', async () => {
      (apiClient.delete as any).mockResolvedValueOnce({})

      await credentialAPI.delete('cred-123')

      expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/credentials/cred-123')
    })

    it('should throw NotFoundError for non-existent credential', async () => {
      const error = new Error('Credential not found')
      error.name = 'NotFoundError'
      ;(apiClient.delete as any).mockRejectedValueOnce(error)

      await expect(credentialAPI.delete('invalid-id')).rejects.toThrow('Credential not found')
    })
  })

  describe('rotate', () => {
    it('should rotate credential value', async () => {
      const newValue = {
        apiKey: 'sk-new-789012',
      }

      const rotatedCredential = {
        ...mockCredential,
        updatedAt: '2024-01-16T10:00:00Z',
      }
      ;(apiClient.post as any).mockResolvedValueOnce(rotatedCredential)

      const result = await credentialAPI.rotate('cred-123', newValue)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/credentials/cred-123/rotate', {
        value: newValue,
      })
      expect(result.updatedAt).toBe('2024-01-16T10:00:00Z')
    })

    it('should handle OAuth2 rotation', async () => {
      const newValue = {
        clientId: 'new-client',
        clientSecret: 'new-secret',
        authUrl: 'https://auth.example.com',
        tokenUrl: 'https://token.example.com',
      }

      ;(apiClient.post as any).mockResolvedValueOnce({ ...mockCredential, type: 'oauth2' })

      const result = await credentialAPI.rotate('cred-123', newValue)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/credentials/cred-123/rotate', {
        value: newValue,
      })
    })
  })

  describe('test', () => {
    it('should test credential connectivity', async () => {
      const testResult = {
        success: true,
        message: 'Connection successful',
        testedAt: '2024-01-15T10:00:00Z',
      }

      ;(apiClient.post as any).mockResolvedValueOnce(testResult)

      const result = await credentialAPI.test('cred-123')

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/credentials/cred-123/test', {})
      expect(result.success).toBe(true)
      expect(result.message).toBe('Connection successful')
    })

    it('should return failure result for invalid credential', async () => {
      const testResult = {
        success: false,
        message: 'Authentication failed',
        testedAt: '2024-01-15T10:00:00Z',
      }

      ;(apiClient.post as any).mockResolvedValueOnce(testResult)

      const result = await credentialAPI.test('cred-123')

      expect(result.success).toBe(false)
      expect(result.message).toBe('Authentication failed')
    })
  })
})
