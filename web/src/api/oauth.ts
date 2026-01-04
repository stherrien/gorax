import { apiClient } from './client'
import type {
  OAuthProvider,
  OAuthConnection,
  AuthorizeInput,
  AuthorizeResponse,
  CallbackResponse,
  TestConnectionResponse,
} from '../types/oauth'

/**
 * List available OAuth providers
 */
export async function listProviders(): Promise<OAuthProvider[]> {
  return apiClient.get('/api/v1/oauth/providers')
}

/**
 * Start OAuth authorization flow
 * Returns authorization URL to redirect user to
 */
export async function authorize(input: AuthorizeInput): Promise<AuthorizeResponse> {
  const params: Record<string, any> = {}
  if (input.scopes && input.scopes.length > 0) {
    input.scopes.forEach((scope) => {
      if (!params.scopes) {
        params.scopes = []
      }
      params.scopes.push(scope)
    })
  }
  if (input.redirect_uri) {
    params.redirect_uri = input.redirect_uri
  }

  return apiClient.get(`/api/v1/oauth/authorize/${input.provider_key}`, {
    params,
    headers: { Accept: 'application/json' },
  })
}

/**
 * List user's OAuth connections
 */
export async function listConnections(): Promise<OAuthConnection[]> {
  return apiClient.get('/api/v1/oauth/connections')
}

/**
 * Get a specific OAuth connection
 */
export async function getConnection(id: string): Promise<OAuthConnection> {
  return apiClient.get(`/api/v1/oauth/connections/${id}`)
}

/**
 * Revoke an OAuth connection
 */
export async function revokeConnection(id: string): Promise<void> {
  return apiClient.delete(`/api/v1/oauth/connections/${id}`)
}

/**
 * Test an OAuth connection
 */
export async function testConnection(id: string): Promise<TestConnectionResponse> {
  return apiClient.post(`/api/v1/oauth/connections/${id}/test`, {})
}

/**
 * Handle OAuth callback
 * Note: This is typically called by the backend, not directly by frontend
 */
export async function handleCallback(
  provider: string,
  code: string,
  state: string
): Promise<CallbackResponse> {
  return apiClient.get(`/api/v1/oauth/callback/${provider}`, {
    params: { code, state },
  })
}
