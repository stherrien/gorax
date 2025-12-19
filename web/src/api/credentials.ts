import { apiClient } from './client'

// Credential types
export type CredentialType = 'api_key' | 'oauth2' | 'basic_auth' | 'bearer_token'

export interface Credential {
  id: string
  tenantId: string
  name: string
  type: CredentialType
  description?: string
  expiresAt?: string
  createdAt: string
  updatedAt: string
}

export interface CredentialListResponse {
  credentials: Credential[]
  total: number
}

export interface CredentialListParams {
  page?: number
  limit?: number
  type?: CredentialType
  search?: string
}

export interface CredentialValue {
  // API Key
  apiKey?: string

  // OAuth2
  clientId?: string
  clientSecret?: string
  authUrl?: string
  tokenUrl?: string
  scopes?: string[]

  // Basic Auth
  username?: string
  password?: string

  // Bearer Token
  token?: string

  // Custom fields
  [key: string]: any
}

export interface CredentialCreateInput {
  name: string
  type: CredentialType
  description?: string
  value: CredentialValue
  expiresAt?: string
}

export interface CredentialUpdateInput {
  name?: string
  description?: string
  expiresAt?: string
}

export interface CredentialTestResult {
  success: boolean
  message: string
  testedAt: string
}

class CredentialAPI {
  /**
   * List all credentials
   */
  async list(params?: CredentialListParams): Promise<CredentialListResponse> {
    const options = params ? { params } : undefined
    return await apiClient.get('/api/v1/credentials', options)
  }

  /**
   * Get a single credential by ID
   */
  async get(id: string): Promise<Credential> {
    return await apiClient.get(`/api/v1/credentials/${id}`)
  }

  /**
   * Create a new credential
   */
  async create(credential: CredentialCreateInput): Promise<Credential> {
    return await apiClient.post('/api/v1/credentials', credential)
  }

  /**
   * Update an existing credential (metadata only)
   */
  async update(id: string, updates: CredentialUpdateInput): Promise<Credential> {
    return await apiClient.put(`/api/v1/credentials/${id}`, updates)
  }

  /**
   * Delete a credential
   */
  async delete(id: string): Promise<void> {
    await apiClient.delete(`/api/v1/credentials/${id}`)
  }

  /**
   * Rotate credential value
   */
  async rotate(id: string, value: CredentialValue): Promise<Credential> {
    return await apiClient.post(`/api/v1/credentials/${id}/rotate`, { value })
  }

  /**
   * Test credential connectivity
   */
  async test(id: string): Promise<CredentialTestResult> {
    return await apiClient.post(`/api/v1/credentials/${id}/test`, {})
  }
}

export const credentialAPI = new CredentialAPI()
