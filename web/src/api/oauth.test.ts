import { describe, it, expect, vi, beforeEach } from 'vitest'
import * as oauthAPI from './oauth'
import { apiClient } from './client'
import type { OAuthProvider, OAuthConnection } from '../types/oauth'

vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('OAuth API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('listProviders', () => {
    it('should fetch OAuth providers', async () => {
      const mockProviders: OAuthProvider[] = [
        {
          id: '1',
          provider_key: 'github',
          name: 'GitHub',
          description: 'GitHub OAuth',
          auth_url: 'https://github.com/login/oauth/authorize',
          token_url: 'https://github.com/login/oauth/access_token',
          user_info_url: 'https://api.github.com/user',
          default_scopes: ['read:user'],
          status: 'active',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ]

      vi.mocked(apiClient.get).mockResolvedValue(mockProviders)

      const result = await oauthAPI.listProviders()

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/oauth/providers')
      expect(result).toEqual(mockProviders)
    })
  })

  describe('authorize', () => {
    it('should get authorization URL with scopes', async () => {
      const mockResponse = {
        authorization_url: 'https://github.com/login/oauth/authorize?client_id=123',
      }

      vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

      const result = await oauthAPI.authorize({
        provider_key: 'github',
        scopes: ['read:user', 'repo'],
      })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/oauth/authorize/github', {
        params: {
          scopes: ['read:user', 'repo'],
        },
        headers: { Accept: 'application/json' },
      })
      expect(result).toEqual(mockResponse)
    })

    it('should get authorization URL without scopes', async () => {
      const mockResponse = {
        authorization_url: 'https://github.com/login/oauth/authorize?client_id=123',
      }

      vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

      const result = await oauthAPI.authorize({
        provider_key: 'github',
      })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/oauth/authorize/github', {
        params: {},
        headers: { Accept: 'application/json' },
      })
      expect(result).toEqual(mockResponse)
    })

    it('should include redirect_uri if provided', async () => {
      const mockResponse = {
        authorization_url: 'https://github.com/login/oauth/authorize?client_id=123',
      }

      vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

      await oauthAPI.authorize({
        provider_key: 'github',
        redirect_uri: 'https://example.com/callback',
      })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/oauth/authorize/github', {
        params: {
          redirect_uri: 'https://example.com/callback',
        },
        headers: { Accept: 'application/json' },
      })
    })
  })

  describe('listConnections', () => {
    it('should fetch user OAuth connections', async () => {
      const mockConnections: OAuthConnection[] = [
        {
          id: 'conn-1',
          user_id: 'user-1',
          tenant_id: 'tenant-1',
          provider_key: 'github',
          provider_user_id: 'gh-123',
          provider_username: 'testuser',
          scopes: ['read:user'],
          status: 'active',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ]

      vi.mocked(apiClient.get).mockResolvedValue(mockConnections)

      const result = await oauthAPI.listConnections()

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/oauth/connections')
      expect(result).toEqual(mockConnections)
    })
  })

  describe('getConnection', () => {
    it('should fetch a specific connection', async () => {
      const mockConnection: OAuthConnection = {
        id: 'conn-1',
        user_id: 'user-1',
        tenant_id: 'tenant-1',
        provider_key: 'github',
        scopes: ['read:user'],
        status: 'active',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }

      vi.mocked(apiClient.get).mockResolvedValue(mockConnection)

      const result = await oauthAPI.getConnection('conn-1')

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/oauth/connections/conn-1')
      expect(result).toEqual(mockConnection)
    })
  })

  describe('revokeConnection', () => {
    it('should revoke a connection', async () => {
      vi.mocked(apiClient.delete).mockResolvedValue({})

      await oauthAPI.revokeConnection('conn-1')

      expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/oauth/connections/conn-1')
    })
  })

  describe('testConnection', () => {
    it('should test a connection successfully', async () => {
      const mockResponse = {
        success: true,
        message: 'Connection test successful',
      }

      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const result = await oauthAPI.testConnection('conn-1')

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/oauth/connections/conn-1/test', {})
      expect(result).toEqual(mockResponse)
    })

    it('should handle test failure', async () => {
      const mockResponse = {
        success: false,
        error: 'Token expired',
      }

      vi.mocked(apiClient.post).mockResolvedValue(mockResponse)

      const result = await oauthAPI.testConnection('conn-1')

      expect(result.success).toBe(false)
      expect(result.error).toBe('Token expired')
    })
  })

  describe('handleCallback', () => {
    it('should handle OAuth callback', async () => {
      const mockResponse = {
        success: true,
        provider: 'github',
        connection: {
          id: 'conn-1',
          user_id: 'user-1',
          tenant_id: 'tenant-1',
          provider_key: 'github',
          scopes: ['read:user'],
          status: 'active' as const,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      }

      vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

      const result = await oauthAPI.handleCallback('github', 'code123', 'state456')

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/oauth/callback/github', {
        params: { code: 'code123', state: 'state456' },
      })
      expect(result).toEqual(mockResponse)
    })
  })
})
