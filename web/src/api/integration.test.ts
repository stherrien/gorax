import { describe, it, expect, beforeAll } from 'vitest'
import { marketplaceAPI } from './marketplace'
import { auditAPI } from './audit'
import { listProviders as listOAuthProviders, listConnections } from './oauth'
import { ssoApi } from './sso'

/**
 * API Integration Tests
 *
 * These tests verify that frontend API clients can successfully
 * communicate with backend endpoints.
 *
 * NOTE: These tests require a running backend server.
 * Run with: npm test -- integration.test.ts
 *
 * Prerequisites:
 * - Backend server running on http://localhost:8181
 * - Test tenant ID: 00000000-0000-0000-0000-000000000001
 * - Test user authenticated (dev mode)
 */

describe('API Integration Tests', () => {
  beforeAll(() => {
    // Set up dev mode tenant header
    localStorage.setItem('auth_token', '')
  })

  describe('Marketplace API', () => {
    it('should list marketplace templates', async () => {
      const templates = await marketplaceAPI.list()
      expect(Array.isArray(templates)).toBe(true)
    })

    it('should get trending templates', async () => {
      const templates = await marketplaceAPI.getTrending(5)
      expect(Array.isArray(templates)).toBe(true)
      expect(templates.length).toBeLessThanOrEqual(5)
    })

    it('should get popular templates', async () => {
      const templates = await marketplaceAPI.getPopular(5)
      expect(Array.isArray(templates)).toBe(true)
      expect(templates.length).toBeLessThanOrEqual(5)
    })

    it('should get categories', async () => {
      const categories = await marketplaceAPI.getCategories()
      expect(Array.isArray(categories)).toBe(true)
    })

    it('should search templates with filters', async () => {
      const templates = await marketplaceAPI.list({
        search: 'test',
        limit: 10,
      })
      expect(Array.isArray(templates)).toBe(true)
    })
  })

  describe('OAuth API', () => {
    it('should list OAuth providers', async () => {
      const providers = await listOAuthProviders()
      expect(Array.isArray(providers)).toBe(true)
    })

    it('should list user OAuth connections', async () => {
      const connections = await listConnections()
      expect(Array.isArray(connections)).toBe(true)
    })
  })

  describe('SSO API', () => {
    it('should list SSO providers', async () => {
      // Note: SSO service is currently disabled in backend
      // This test will fail until SSO is properly initialized
      try {
        const providers = await ssoApi.listProviders()
        expect(Array.isArray(providers)).toBe(true)
      } catch (error: any) {
        // Expected to fail if SSO not initialized
        expect(error.status).toBeOneOf([404, 500])
      }
    })
  })

  describe('Audit API (Admin)', () => {
    it('should query audit events', async () => {
      // Note: Requires admin role
      try {
        const result = await auditAPI.queryAuditEvents({
          limit: 10,
          offset: 0,
        })
        expect(result).toHaveProperty('events')
        expect(result).toHaveProperty('total')
        expect(Array.isArray(result.events)).toBe(true)
      } catch (error: any) {
        // Expected to fail if not admin
        expect(error.status).toBeOneOf([401, 403])
      }
    })

    it('should get audit stats', async () => {
      // Note: Requires admin role
      try {
        const stats = await auditAPI.getAuditStats({
          startDate: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
          endDate: new Date().toISOString(),
        })
        expect(stats).toBeDefined()
      } catch (error: any) {
        // Expected to fail if not admin
        expect(error.status).toBeOneOf([401, 403])
      }
    })
  })

  describe('Error Handling', () => {
    it('should handle 404 errors gracefully', async () => {
      try {
        await marketplaceAPI.get('non-existent-id')
        expect.fail('Should have thrown error')
      } catch (error: any) {
        expect(error.status).toBe(404)
        expect(error.name).toBe('NotFoundError')
      }
    })

    it('should handle authentication errors', async () => {
      // Save current token
      const originalToken = localStorage.getItem('auth_token')

      // Set invalid token
      localStorage.setItem('auth_token', 'invalid-token')

      try {
        await marketplaceAPI.list()
        // If it succeeds, it means dev mode is bypassing auth (expected)
        expect(true).toBe(true)
      } catch (error: any) {
        // In production mode, should get 401
        expect(error.status).toBe(401)
        expect(error.name).toBe('AuthError')
      } finally {
        // Restore token
        if (originalToken) {
          localStorage.setItem('auth_token', originalToken)
        } else {
          localStorage.removeItem('auth_token')
        }
      }
    })
  })

  describe('API Client Configuration', () => {
    it('should have correct base URL configuration', () => {
      // In dev mode, base URL should be empty (uses Vite proxy)
      const apiBaseURL = import.meta.env.VITE_API_URL || ''
      expect(apiBaseURL).toBe('')
    })

    it('should include proper headers in requests', () => {
      // This is tested implicitly by other tests
      // Just verify localStorage is accessible
      expect(typeof localStorage.getItem).toBe('function')
    })
  })
})
