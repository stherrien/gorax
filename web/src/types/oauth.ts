/**
 * OAuth provider configuration
 */
export interface OAuthProvider {
  id: string
  provider_key: string
  name: string
  description: string
  auth_url: string
  token_url: string
  user_info_url: string
  default_scopes: string[]
  client_id?: string
  status: ProviderStatus
  config?: Record<string, any>
  created_at: string
  updated_at: string
}

/**
 * OAuth provider status
 */
export type ProviderStatus = 'active' | 'inactive'

/**
 * OAuth connection status
 */
export type ConnectionStatus = 'active' | 'revoked' | 'expired'

/**
 * OAuth connection
 */
export interface OAuthConnection {
  id: string
  user_id: string
  tenant_id: string
  provider_key: string
  provider_user_id?: string
  provider_username?: string
  provider_email?: string
  token_expiry?: string
  scopes: string[]
  status: ConnectionStatus
  created_at: string
  updated_at: string
  last_used_at?: string
  last_refresh_at?: string
  metadata?: Record<string, any>
}

/**
 * OAuth authorization input
 */
export interface AuthorizeInput {
  provider_key: string
  scopes?: string[]
  redirect_uri?: string
}

/**
 * OAuth authorization response
 */
export interface AuthorizeResponse {
  authorization_url: string
}

/**
 * OAuth callback parameters
 */
export interface CallbackParams {
  code?: string
  state?: string
  error?: string
  error_description?: string
}

/**
 * OAuth callback response
 */
export interface CallbackResponse {
  success: boolean
  provider: string
  connection: OAuthConnection
}

/**
 * OAuth test connection response
 */
export interface TestConnectionResponse {
  success: boolean
  message?: string
  error?: string
}

/**
 * Provider branding configuration
 */
export interface ProviderBranding {
  name: string
  color: string
  icon: string
  iconBg: string
}

/**
 * Provider branding map
 */
export const PROVIDER_BRANDING: Record<string, ProviderBranding> = {
  github: {
    name: 'GitHub',
    color: '#333',
    icon: '🐙',
    iconBg: '#f6f8fa',
  },
  google: {
    name: 'Google',
    color: '#4285F4',
    icon: '🔍',
    iconBg: '#e8f0fe',
  },
  slack: {
    name: 'Slack',
    color: '#4A154B',
    icon: '💬',
    iconBg: '#f4ede4',
  },
  microsoft: {
    name: 'Microsoft',
    color: '#00A4EF',
    icon: '🏢',
    iconBg: '#e6f4ff',
  },
}
