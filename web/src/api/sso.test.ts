import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import {
  ssoApi,
  buildSSOLoginUrl,
  extractEmailDomain,
  validateProviderConfig,
  getDefaultSAMLConfig,
  getDefaultOIDCConfig,
} from './sso'
import type {
  SSOProvider,
  CreateProviderRequest,
  UpdateProviderRequest,
  ProviderDiscoveryResponse,
  SSOLoginEvent,
  SAMLConfig,
  OIDCConfig,
} from '../types/sso'

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

describe('SSO API', () => {
  const mockSAMLConfig: SAMLConfig = {
    entity_id: 'https://app.example.com',
    acs_url: 'https://app.example.com/api/v1/sso/acs',
    idp_metadata_url: 'https://idp.example.com/metadata',
    idp_entity_id: 'https://idp.example.com',
    idp_sso_url: 'https://idp.example.com/sso',
    attribute_mapping: {
      email: 'NameID',
      first_name: 'firstName',
      last_name: 'lastName',
    },
    sign_authn_requests: true,
  }

  const mockOIDCConfig: OIDCConfig = {
    client_id: 'client-123',
    client_secret: 'secret-456',
    discovery_url: 'https://idp.example.com/.well-known/openid-configuration',
    redirect_url: 'https://app.example.com/api/v1/sso/callback/provider-123',
    scopes: ['openid', 'profile', 'email'],
    attribute_mapping: {
      email: 'email',
      first_name: 'given_name',
      last_name: 'family_name',
    },
  }

  const mockSAMLProvider: SSOProvider = {
    id: 'provider-123',
    tenant_id: 'tenant-1',
    name: 'Corporate SAML',
    provider_type: 'saml',
    enabled: true,
    enforce_sso: false,
    config: mockSAMLConfig,
    domains: ['example.com', 'corp.example.com'],
    created_at: '2024-01-15T10:00:00Z',
    updated_at: '2024-01-15T10:00:00Z',
  }

  const mockOIDCProvider: SSOProvider = {
    ...mockSAMLProvider,
    id: 'provider-456',
    name: 'Corporate OIDC',
    provider_type: 'oidc',
    config: mockOIDCConfig,
  }

  const mockLoginEvent: SSOLoginEvent = {
    id: 'event-123',
    sso_provider_id: 'provider-123',
    user_id: 'user-123',
    external_id: 'external-user-456',
    status: 'success',
    ip_address: '192.168.1.1',
    user_agent: 'Mozilla/5.0',
    created_at: '2024-01-15T10:00:00Z',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Provider Management', () => {
    describe('createProvider', () => {
      it('should create SAML provider', async () => {
        const request: CreateProviderRequest = {
          name: 'New SAML Provider',
          provider_type: 'saml',
          enabled: true,
          enforce_sso: false,
          config: mockSAMLConfig,
          domains: ['example.com'],
        }
        ;(apiClient.post as any).mockResolvedValueOnce(mockSAMLProvider)

        const result = await ssoApi.createProvider(request)

        expect(apiClient.post).toHaveBeenCalledWith('/api/v1/sso/providers', request)
        expect(result).toEqual(mockSAMLProvider)
      })

      it('should create OIDC provider', async () => {
        const request: CreateProviderRequest = {
          name: 'New OIDC Provider',
          provider_type: 'oidc',
          enabled: true,
          enforce_sso: true,
          config: mockOIDCConfig,
          domains: ['example.com'],
        }
        ;(apiClient.post as any).mockResolvedValueOnce(mockOIDCProvider)

        const result = await ssoApi.createProvider(request)

        expect(apiClient.post).toHaveBeenCalledWith('/api/v1/sso/providers', request)
        expect(result).toEqual(mockOIDCProvider)
      })

      it('should handle validation error', async () => {
        const error = new Error('Name is required')
        ;(apiClient.post as any).mockRejectedValueOnce(error)

        await expect(
          ssoApi.createProvider({
            name: '',
            provider_type: 'saml',
            enabled: true,
            enforce_sso: false,
            config: mockSAMLConfig,
            domains: [],
          })
        ).rejects.toThrow('Name is required')
      })
    })

    describe('listProviders', () => {
      it('should fetch list of providers', async () => {
        const providers = [mockSAMLProvider, mockOIDCProvider]
        ;(apiClient.get as any).mockResolvedValueOnce(providers)

        const result = await ssoApi.listProviders()

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/sso/providers')
        expect(result).toEqual(providers)
        expect(result).toHaveLength(2)
      })

      it('should handle empty list', async () => {
        (apiClient.get as any).mockResolvedValueOnce([])

        const result = await ssoApi.listProviders()

        expect(result).toEqual([])
      })
    })

    describe('getProvider', () => {
      it('should fetch single provider by ID', async () => {
        (apiClient.get as any).mockResolvedValueOnce(mockSAMLProvider)

        const result = await ssoApi.getProvider('provider-123')

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/sso/providers/provider-123')
        expect(result).toEqual(mockSAMLProvider)
      })

      it('should throw NotFoundError for invalid ID', async () => {
        const error = new Error('Provider not found')
        ;(apiClient.get as any).mockRejectedValueOnce(error)

        await expect(ssoApi.getProvider('invalid-id')).rejects.toThrow('Provider not found')
      })
    })

    describe('updateProvider', () => {
      it('should update provider name', async () => {
        const updates: UpdateProviderRequest = {
          name: 'Updated Provider Name',
        }
        const updatedProvider = { ...mockSAMLProvider, name: 'Updated Provider Name' }
        ;(apiClient.put as any).mockResolvedValueOnce(updatedProvider)

        const result = await ssoApi.updateProvider('provider-123', updates)

        expect(apiClient.put).toHaveBeenCalledWith(
          '/api/v1/sso/providers/provider-123',
          updates
        )
        expect(result.name).toBe('Updated Provider Name')
      })

      it('should update provider enabled status', async () => {
        const updates: UpdateProviderRequest = {
          enabled: false,
        }
        const updatedProvider = { ...mockSAMLProvider, enabled: false }
        ;(apiClient.put as any).mockResolvedValueOnce(updatedProvider)

        const result = await ssoApi.updateProvider('provider-123', updates)

        expect(result.enabled).toBe(false)
      })

      it('should update provider domains', async () => {
        const updates: UpdateProviderRequest = {
          domains: ['newdomain.com', 'anotherdomain.com'],
        }
        const updatedProvider = { ...mockSAMLProvider, domains: updates.domains! }
        ;(apiClient.put as any).mockResolvedValueOnce(updatedProvider)

        const result = await ssoApi.updateProvider('provider-123', updates)

        expect(result.domains).toEqual(['newdomain.com', 'anotherdomain.com'])
      })

      it('should update provider config', async () => {
        const newConfig: SAMLConfig = {
          ...mockSAMLConfig,
          sign_authn_requests: false,
        }
        const updates: UpdateProviderRequest = {
          config: newConfig,
        }
        const updatedProvider = { ...mockSAMLProvider, config: newConfig }
        ;(apiClient.put as any).mockResolvedValueOnce(updatedProvider)

        const result = await ssoApi.updateProvider('provider-123', updates)

        expect((result.config as SAMLConfig).sign_authn_requests).toBe(false)
      })
    })

    describe('deleteProvider', () => {
      it('should delete provider by ID', async () => {
        (apiClient.delete as any).mockResolvedValueOnce(undefined)

        await ssoApi.deleteProvider('provider-123')

        expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/sso/providers/provider-123')
      })

      it('should throw error for non-existent provider', async () => {
        const error = new Error('Provider not found')
        ;(apiClient.delete as any).mockRejectedValueOnce(error)

        await expect(ssoApi.deleteProvider('invalid-id')).rejects.toThrow('Provider not found')
      })
    })
  })

  describe('Provider Discovery', () => {
    describe('discoverProvider', () => {
      it('should discover provider by email domain', async () => {
        const discoveryResponse: ProviderDiscoveryResponse = {
          sso_available: true,
          provider_id: 'provider-123',
          provider_name: 'Corporate SAML',
          provider_type: 'saml',
          enforce_sso: false,
        }
        ;(apiClient.get as any).mockResolvedValueOnce(discoveryResponse)

        const result = await ssoApi.discoverProvider('user@example.com')

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/sso/discover', {
          params: { email: 'user@example.com' },
        })
        expect(result).toEqual(discoveryResponse)
      })

      it('should return sso_available false for unknown domain', async () => {
        const discoveryResponse: ProviderDiscoveryResponse = {
          sso_available: false,
        }
        ;(apiClient.get as any).mockResolvedValueOnce(discoveryResponse)

        const result = await ssoApi.discoverProvider('user@unknown.com')

        expect(result.sso_available).toBe(false)
        expect(result.provider_id).toBeUndefined()
      })

      it('should discover enforced SSO provider', async () => {
        const discoveryResponse: ProviderDiscoveryResponse = {
          sso_available: true,
          provider_id: 'provider-456',
          provider_name: 'Corporate OIDC',
          provider_type: 'oidc',
          enforce_sso: true,
        }
        ;(apiClient.get as any).mockResolvedValueOnce(discoveryResponse)

        const result = await ssoApi.discoverProvider('user@corp.example.com')

        expect(result.enforce_sso).toBe(true)
      })
    })
  })

  describe('SAML Metadata', () => {
    describe('getMetadata', () => {
      it('should fetch SAML metadata for provider', async () => {
        const metadata = '<EntityDescriptor>...</EntityDescriptor>'
        ;(apiClient.get as any).mockResolvedValueOnce(metadata)

        const result = await ssoApi.getMetadata('provider-123')

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/sso/metadata/provider-123')
        expect(result).toBe(metadata)
      })

      it('should handle error for OIDC provider', async () => {
        const error = new Error('Metadata not available for OIDC providers')
        ;(apiClient.get as any).mockRejectedValueOnce(error)

        await expect(ssoApi.getMetadata('oidc-provider')).rejects.toThrow(
          'Metadata not available for OIDC providers'
        )
      })
    })
  })

  describe('Login Events', () => {
    describe('getLoginEvents', () => {
      it('should fetch login events with default limit', async () => {
        const events = [mockLoginEvent]
        ;(apiClient.get as any).mockResolvedValueOnce(events)

        const result = await ssoApi.getLoginEvents('provider-123')

        expect(apiClient.get).toHaveBeenCalledWith(
          '/api/v1/sso/providers/provider-123/events',
          { params: { limit: 100 } }
        )
        expect(result).toEqual(events)
      })

      it('should fetch login events with custom limit', async () => {
        const events = [mockLoginEvent]
        ;(apiClient.get as any).mockResolvedValueOnce(events)

        const result = await ssoApi.getLoginEvents('provider-123', 50)

        expect(apiClient.get).toHaveBeenCalledWith(
          '/api/v1/sso/providers/provider-123/events',
          { params: { limit: 50 } }
        )
        expect(result).toEqual(events)
      })

      it('should return events with different statuses', async () => {
        const events: SSOLoginEvent[] = [
          mockLoginEvent,
          { ...mockLoginEvent, id: 'event-456', status: 'failure', error_message: 'Invalid assertion' },
          { ...mockLoginEvent, id: 'event-789', status: 'error', error_message: 'IdP unreachable' },
        ]
        ;(apiClient.get as any).mockResolvedValueOnce(events)

        const result = await ssoApi.getLoginEvents('provider-123')

        expect(result).toHaveLength(3)
        expect(result[0].status).toBe('success')
        expect(result[1].status).toBe('failure')
        expect(result[2].status).toBe('error')
      })

      it('should handle empty events list', async () => {
        (apiClient.get as any).mockResolvedValueOnce([])

        const result = await ssoApi.getLoginEvents('provider-123')

        expect(result).toEqual([])
      })
    })
  })

  describe('Helper Functions', () => {
    describe('buildSSOLoginUrl', () => {
      const originalLocation = window.location

      beforeEach(() => {
        // Mock window.location
        delete (window as any).location
        window.location = {
          ...originalLocation,
          origin: 'https://app.example.com',
        } as Location
      })

      afterEach(() => {
        window.location = originalLocation
      })

      it('should build login URL without relay state', () => {
        const url = buildSSOLoginUrl('provider-123')

        expect(url).toBe('https://app.example.com/api/v1/sso/login/provider-123')
      })

      it('should build login URL with relay state', () => {
        const url = buildSSOLoginUrl('provider-123', '/dashboard')

        expect(url).toBe(
          'https://app.example.com/api/v1/sso/login/provider-123?relay_state=%2Fdashboard'
        )
      })

      it('should handle relay state with special characters', () => {
        const url = buildSSOLoginUrl('provider-123', '/path?param=value&other=123')

        expect(url).toContain('relay_state=')
        expect(url).toContain('%2Fpath%3Fparam%3Dvalue%26other%3D123')
      })
    })

    describe('extractEmailDomain', () => {
      it('should extract domain from valid email', () => {
        expect(extractEmailDomain('user@example.com')).toBe('example.com')
      })

      it('should extract subdomain', () => {
        expect(extractEmailDomain('user@corp.example.com')).toBe('corp.example.com')
      })

      it('should return null for invalid email', () => {
        expect(extractEmailDomain('invalid-email')).toBeNull()
      })

      it('should return null for empty string', () => {
        expect(extractEmailDomain('')).toBeNull()
      })

      it('should handle email with plus sign', () => {
        expect(extractEmailDomain('user+test@example.com')).toBe('example.com')
      })
    })

    describe('validateProviderConfig', () => {
      describe('SAML validation', () => {
        it('should pass for valid SAML config', () => {
          const config = {
            entity_id: 'https://app.example.com',
            acs_url: 'https://app.example.com/acs',
            idp_metadata_url: 'https://idp.example.com/metadata',
          }

          const errors = validateProviderConfig('saml', config)

          expect(errors).toEqual([])
        })

        it('should fail without entity_id', () => {
          const config = {
            acs_url: 'https://app.example.com/acs',
            idp_metadata_url: 'https://idp.example.com/metadata',
          }

          const errors = validateProviderConfig('saml', config)

          expect(errors).toContain('Entity ID is required')
        })

        it('should fail without acs_url', () => {
          const config = {
            entity_id: 'https://app.example.com',
            idp_metadata_url: 'https://idp.example.com/metadata',
          }

          const errors = validateProviderConfig('saml', config)

          expect(errors).toContain('ACS URL is required')
        })

        it('should fail without IdP metadata', () => {
          const config = {
            entity_id: 'https://app.example.com',
            acs_url: 'https://app.example.com/acs',
          }

          const errors = validateProviderConfig('saml', config)

          expect(errors).toContain('IdP metadata URL or XML is required')
        })

        it('should pass with idp_metadata_xml instead of URL', () => {
          const config = {
            entity_id: 'https://app.example.com',
            acs_url: 'https://app.example.com/acs',
            idp_metadata_xml: '<EntityDescriptor>...</EntityDescriptor>',
          }

          const errors = validateProviderConfig('saml', config)

          expect(errors).toEqual([])
        })

        it('should return multiple errors', () => {
          const errors = validateProviderConfig('saml', {})

          expect(errors).toContain('Entity ID is required')
          expect(errors).toContain('ACS URL is required')
          expect(errors).toContain('IdP metadata URL or XML is required')
          expect(errors).toHaveLength(3)
        })
      })

      describe('OIDC validation', () => {
        it('should pass for valid OIDC config', () => {
          const config = {
            client_id: 'client-123',
            client_secret: 'secret-456',
            discovery_url: 'https://idp.example.com/.well-known/openid-configuration',
            redirect_url: 'https://app.example.com/callback',
          }

          const errors = validateProviderConfig('oidc', config)

          expect(errors).toEqual([])
        })

        it('should fail without client_id', () => {
          const config = {
            client_secret: 'secret-456',
            discovery_url: 'https://idp.example.com/.well-known/openid-configuration',
            redirect_url: 'https://app.example.com/callback',
          }

          const errors = validateProviderConfig('oidc', config)

          expect(errors).toContain('Client ID is required')
        })

        it('should fail without client_secret', () => {
          const config = {
            client_id: 'client-123',
            discovery_url: 'https://idp.example.com/.well-known/openid-configuration',
            redirect_url: 'https://app.example.com/callback',
          }

          const errors = validateProviderConfig('oidc', config)

          expect(errors).toContain('Client Secret is required')
        })

        it('should fail without discovery_url', () => {
          const config = {
            client_id: 'client-123',
            client_secret: 'secret-456',
            redirect_url: 'https://app.example.com/callback',
          }

          const errors = validateProviderConfig('oidc', config)

          expect(errors).toContain('Discovery URL is required')
        })

        it('should fail without redirect_url', () => {
          const config = {
            client_id: 'client-123',
            client_secret: 'secret-456',
            discovery_url: 'https://idp.example.com/.well-known/openid-configuration',
          }

          const errors = validateProviderConfig('oidc', config)

          expect(errors).toContain('Redirect URL is required')
        })

        it('should return multiple errors', () => {
          const errors = validateProviderConfig('oidc', {})

          expect(errors).toContain('Client ID is required')
          expect(errors).toContain('Client Secret is required')
          expect(errors).toContain('Discovery URL is required')
          expect(errors).toContain('Redirect URL is required')
          expect(errors).toHaveLength(4)
        })
      })
    })

    describe('getDefaultSAMLConfig', () => {
      it('should generate default SAML config', () => {
        const baseUrl = 'https://app.example.com'
        const config = getDefaultSAMLConfig(baseUrl)

        expect(config.entity_id).toBe(baseUrl)
        expect(config.acs_url).toBe(`${baseUrl}/api/v1/sso/acs`)
        expect(config.idp_metadata_url).toBe('')
        expect(config.sign_authn_requests).toBe(false)
        expect(config.attribute_mapping).toEqual({
          email: 'NameID',
          first_name: 'firstName',
          last_name: 'lastName',
          groups: 'groups',
        })
      })
    })

    describe('getDefaultOIDCConfig', () => {
      it('should generate default OIDC config', () => {
        const baseUrl = 'https://app.example.com'
        const providerId = 'provider-123'
        const config = getDefaultOIDCConfig(baseUrl, providerId)

        expect(config.client_id).toBe('')
        expect(config.client_secret).toBe('')
        expect(config.discovery_url).toBe('')
        expect(config.redirect_url).toBe(`${baseUrl}/api/v1/sso/callback/${providerId}`)
        expect(config.scopes).toEqual(['openid', 'profile', 'email'])
        expect(config.attribute_mapping).toEqual({
          email: 'email',
          first_name: 'given_name',
          last_name: 'family_name',
          groups: 'groups',
        })
      })
    })
  })

  describe('initiateLogin', () => {
    const originalLocation = window.location

    beforeEach(() => {
      // Mock window.location
      delete (window as any).location
      window.location = {
        ...originalLocation,
        origin: 'https://app.example.com',
        href: '',
      } as Location
    })

    afterEach(() => {
      window.location = originalLocation
    })

    it('should redirect to login URL without relay state', () => {
      ssoApi.initiateLogin('provider-123')

      expect(window.location.href).toBe('https://app.example.com/api/v1/sso/login/provider-123')
    })

    it('should redirect to login URL with relay state', () => {
      ssoApi.initiateLogin('provider-123', '/dashboard')

      expect(window.location.href).toBe(
        'https://app.example.com/api/v1/sso/login/provider-123?relay_state=%2Fdashboard'
      )
    })
  })
})
