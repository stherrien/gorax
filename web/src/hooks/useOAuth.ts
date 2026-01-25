import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as oauthAPI from '../api/oauth'
import type {
  OAuthProvider,
  OAuthConnection,
  AuthorizeInput,
  TestConnectionResponse,
} from '../types/oauth'

/**
 * Query key factory for OAuth queries
 */
const oauthKeys = {
  all: ['oauth'] as const,
  providers: () => [...oauthKeys.all, 'providers'] as const,
  connections: () => [...oauthKeys.all, 'connections'] as const,
  connection: (id: string) => [...oauthKeys.connections(), id] as const,
}

/**
 * Fetch available OAuth providers
 */
export function useOAuthProviders() {
  return useQuery<OAuthProvider[], Error>({
    queryKey: oauthKeys.providers(),
    queryFn: oauthAPI.listProviders,
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}

/**
 * Fetch user's OAuth connections
 */
export function useOAuthConnections() {
  return useQuery<OAuthConnection[], Error>({
    queryKey: oauthKeys.connections(),
    queryFn: oauthAPI.listConnections,
    staleTime: 30 * 1000, // 30 seconds
  })
}

/**
 * Fetch a specific OAuth connection
 */
export function useOAuthConnection(id: string) {
  return useQuery<OAuthConnection, Error>({
    queryKey: oauthKeys.connection(id),
    queryFn: () => oauthAPI.getConnection(id),
    enabled: !!id,
  })
}

/**
 * Start OAuth authorization flow
 * Opens authorization URL in popup or redirects
 */
export function useAuthorize() {
  return useMutation<string, Error, AuthorizeInput>({
    mutationFn: async (input: AuthorizeInput) => {
      const response = await oauthAPI.authorize(input)
      return response.authorization_url
    },
  })
}

/**
 * Revoke an OAuth connection
 */
export function useRevokeConnection() {
  const queryClient = useQueryClient()

  return useMutation<void, Error, string>({
    mutationFn: oauthAPI.revokeConnection,
    onSuccess: () => {
      // Invalidate connections list after revoke
      queryClient.invalidateQueries({ queryKey: oauthKeys.connections() })
    },
  })
}

/**
 * Test an OAuth connection
 */
export function useTestConnection() {
  return useMutation<TestConnectionResponse, Error, string>({
    mutationFn: oauthAPI.testConnection,
  })
}

/**
 * Open OAuth authorization in popup window
 * Returns a promise that resolves when popup closes
 */
export function openOAuthPopup(authUrl: string, provider: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const width = 600
    const height = 700
    const left = window.screenX + (window.outerWidth - width) / 2
    const top = window.screenY + (window.outerHeight - height) / 2

    const popup = window.open(
      authUrl,
      `oauth_${provider}`,
      `width=${width},height=${height},left=${left},top=${top},toolbar=0,scrollbars=1,status=1,resizable=1`
    )

    if (!popup) {
      reject(new Error('Failed to open popup. Please allow popups for this site.'))
      return
    }

    // Poll to detect when popup closes
    const pollTimer = setInterval(() => {
      if (popup.closed) {
        clearInterval(pollTimer)
        resolve()
      }
    }, 500)

    // Cleanup after 5 minutes
    setTimeout(() => {
      clearInterval(pollTimer)
      if (!popup.closed) {
        popup.close()
      }
      reject(new Error('OAuth authorization timeout'))
    }, 5 * 60 * 1000)
  })
}

/**
 * Check if user has an active connection for a provider
 */
export function useHasConnection(providerKey: string) {
  const { data: connections } = useOAuthConnections()

  const connection = connections?.find(
    (conn) => conn.provider_key === providerKey && conn.status === 'active'
  )

  return {
    hasConnection: !!connection,
    connection,
  }
}
