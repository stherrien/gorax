import React, { useState, useEffect } from 'react'
import type { Credential, CredentialType, CredentialCreateInput, CredentialUpdateInput } from '../../api/credentials'

export interface CredentialFormProps {
  credential?: Credential
  loading?: boolean
  error?: string | null
  onSubmit: (data: CredentialCreateInput | CredentialUpdateInput) => void
  onCancel: () => void
}

const CREDENTIAL_TYPE_OPTIONS: { value: CredentialType; label: string }[] = [
  { value: 'api_key', label: 'API Key' },
  { value: 'oauth2', label: 'OAuth2' },
  { value: 'basic_auth', label: 'Basic Auth' },
  { value: 'bearer_token', label: 'Bearer Token' },
]

const calculatePasswordStrength = (password: string): 'Weak' | 'Medium' | 'Strong' => {
  if (password.length < 8) return 'Weak'

  const hasUpper = /[A-Z]/.test(password)
  const hasLower = /[a-z]/.test(password)
  const hasNumber = /[0-9]/.test(password)
  const hasSpecial = /[!@#$%^&*(),.?":{}|<>]/.test(password)

  const score = [hasUpper, hasLower, hasNumber, hasSpecial].filter(Boolean).length

  if (score >= 3 && password.length >= 12) return 'Strong'
  if (score >= 2) return 'Medium'
  return 'Weak'
}

export const CredentialForm: React.FC<CredentialFormProps> = ({
  credential,
  loading = false,
  error,
  onSubmit,
  onCancel,
}) => {
  const isEditMode = !!credential

  // Form state
  const [name, setName] = useState(credential?.name || '')
  const [description, setDescription] = useState(credential?.description || '')
  const [type, setType] = useState<CredentialType>(credential?.type || 'api_key')
  const [expiresAt, setExpiresAt] = useState('')

  // Credential value state
  const [apiKey, setApiKey] = useState('')
  const [clientId, setClientId] = useState('')
  const [clientSecret, setClientSecret] = useState('')
  const [authUrl, setAuthUrl] = useState('')
  const [tokenUrl, setTokenUrl] = useState('')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [token, setToken] = useState('')

  // Validation errors
  const [errors, setErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    if (credential?.expiresAt) {
      // Format date for input field (YYYY-MM-DD)
      const date = new Date(credential.expiresAt)
      setExpiresAt(date.toISOString().split('T')[0])
    }
  }, [credential])

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {}

    if (!name.trim()) {
      newErrors.name = 'Name is required'
    }

    if (!isEditMode) {
      // Validate credential values for create mode
      switch (type) {
        case 'api_key':
          if (!apiKey.trim()) {
            newErrors.apiKey = 'API Key is required'
          }
          break
        case 'oauth2':
          if (!clientId.trim()) {
            newErrors.clientId = 'Client ID is required'
          }
          if (!clientSecret.trim()) {
            newErrors.clientSecret = 'Client Secret is required'
          }
          if (!authUrl.trim()) {
            newErrors.authUrl = 'Auth URL is required'
          }
          if (!tokenUrl.trim()) {
            newErrors.tokenUrl = 'Token URL is required'
          }
          break
        case 'basic_auth':
          if (!username.trim()) {
            newErrors.username = 'Username is required'
          }
          if (!password.trim()) {
            newErrors.password = 'Password is required'
          }
          break
        case 'bearer_token':
          if (!token.trim()) {
            newErrors.token = 'Token is required'
          }
          break
      }
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) {
      return
    }

    if (isEditMode) {
      // Edit mode: only send metadata updates
      const updates: CredentialUpdateInput = {
        name,
        description: description || undefined,
        expiresAt: expiresAt ? new Date(expiresAt).toISOString() : undefined,
      }
      onSubmit(updates)
    } else {
      // Create mode: send full credential data
      const value: Record<string, any> = {}

      switch (type) {
        case 'api_key':
          value.apiKey = apiKey
          break
        case 'oauth2':
          value.clientId = clientId
          value.clientSecret = clientSecret
          value.authUrl = authUrl
          value.tokenUrl = tokenUrl
          break
        case 'basic_auth':
          value.username = username
          value.password = password
          break
        case 'bearer_token':
          value.token = token
          break
      }

      const createData: CredentialCreateInput = {
        name,
        description: description || undefined,
        type,
        value,
        expiresAt: expiresAt ? new Date(expiresAt).toISOString() : undefined,
      }
      onSubmit(createData)
    }
  }

  const passwordStrength = password ? calculatePasswordStrength(password) : null

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-900 mb-6">
          {isEditMode ? 'Edit Credential' : 'Create Credential'}
        </h2>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      <div className="space-y-4">
        {/* Name */}
        <div>
          <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
            Name *
          </label>
          <input
            id="name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            disabled={loading}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
            placeholder="My API Credential"
          />
          {errors.name && <p className="mt-1 text-sm text-red-600">{errors.name}</p>}
        </div>

        {/* Description */}
        <div>
          <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-1">
            Description
          </label>
          <textarea
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            disabled={loading}
            rows={3}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
            placeholder="Optional description"
          />
        </div>

        {/* Type */}
        <div>
          <label htmlFor="type" className="block text-sm font-medium text-gray-700 mb-1">
            Type *
          </label>
          <select
            id="type"
            value={type}
            onChange={(e) => setType(e.target.value as CredentialType)}
            disabled={loading || isEditMode}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
          >
            {CREDENTIAL_TYPE_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>

        {/* Credential Value Fields (Create mode only) */}
        {!isEditMode && (
          <div className="border-t border-gray-200 pt-4 mt-4">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Credential Details</h3>

            {type === 'api_key' && (
              <div>
                <label htmlFor="apiKey" className="block text-sm font-medium text-gray-700 mb-1">
                  API Key *
                </label>
                <input
                  id="apiKey"
                  type="password"
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  disabled={loading}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
                  placeholder="sk-..."
                />
                {errors.apiKey && <p className="mt-1 text-sm text-red-600">{errors.apiKey}</p>}
              </div>
            )}

            {type === 'oauth2' && (
              <div className="space-y-4">
                <div>
                  <label htmlFor="clientId" className="block text-sm font-medium text-gray-700 mb-1">
                    Client ID *
                  </label>
                  <input
                    id="clientId"
                    type="text"
                    value={clientId}
                    onChange={(e) => setClientId(e.target.value)}
                    disabled={loading}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
                  />
                  {errors.clientId && <p className="mt-1 text-sm text-red-600">{errors.clientId}</p>}
                </div>

                <div>
                  <label htmlFor="clientSecret" className="block text-sm font-medium text-gray-700 mb-1">
                    Client Secret *
                  </label>
                  <input
                    id="clientSecret"
                    type="password"
                    value={clientSecret}
                    onChange={(e) => setClientSecret(e.target.value)}
                    disabled={loading}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
                  />
                  {errors.clientSecret && <p className="mt-1 text-sm text-red-600">{errors.clientSecret}</p>}
                </div>

                <div>
                  <label htmlFor="authUrl" className="block text-sm font-medium text-gray-700 mb-1">
                    Auth URL *
                  </label>
                  <input
                    id="authUrl"
                    type="url"
                    value={authUrl}
                    onChange={(e) => setAuthUrl(e.target.value)}
                    disabled={loading}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
                    placeholder="https://auth.example.com"
                  />
                  {errors.authUrl && <p className="mt-1 text-sm text-red-600">{errors.authUrl}</p>}
                </div>

                <div>
                  <label htmlFor="tokenUrl" className="block text-sm font-medium text-gray-700 mb-1">
                    Token URL *
                  </label>
                  <input
                    id="tokenUrl"
                    type="url"
                    value={tokenUrl}
                    onChange={(e) => setTokenUrl(e.target.value)}
                    disabled={loading}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
                    placeholder="https://token.example.com"
                  />
                  {errors.tokenUrl && <p className="mt-1 text-sm text-red-600">{errors.tokenUrl}</p>}
                </div>
              </div>
            )}

            {type === 'basic_auth' && (
              <div className="space-y-4">
                <div>
                  <label htmlFor="username" className="block text-sm font-medium text-gray-700 mb-1">
                    Username *
                  </label>
                  <input
                    id="username"
                    type="text"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    disabled={loading}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
                  />
                  {errors.username && <p className="mt-1 text-sm text-red-600">{errors.username}</p>}
                </div>

                <div>
                  <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
                    Password *
                  </label>
                  <input
                    id="password"
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    disabled={loading}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
                  />
                  {errors.password && <p className="mt-1 text-sm text-red-600">{errors.password}</p>}

                  {passwordStrength && (
                    <div className="mt-2">
                      <div className="flex items-center gap-2">
                        <span className="text-xs text-gray-600">Strength:</span>
                        <span
                          className={`text-xs font-medium ${
                            passwordStrength === 'Strong'
                              ? 'text-green-600'
                              : passwordStrength === 'Medium'
                              ? 'text-yellow-600'
                              : 'text-red-600'
                          }`}
                        >
                          {passwordStrength}
                        </span>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}

            {type === 'bearer_token' && (
              <div>
                <label htmlFor="token" className="block text-sm font-medium text-gray-700 mb-1">
                  Token *
                </label>
                <input
                  id="token"
                  type="password"
                  value={token}
                  onChange={(e) => setToken(e.target.value)}
                  disabled={loading}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
                  placeholder="Bearer token"
                />
                {errors.token && <p className="mt-1 text-sm text-red-600">{errors.token}</p>}
              </div>
            )}
          </div>
        )}

        {/* Rotate credential link (Edit mode only) */}
        {isEditMode && (
          <div className="border-t border-gray-200 pt-4 mt-4">
            <p className="text-sm text-gray-600">
              To update the credential value, use the{' '}
              <button type="button" className="text-blue-600 hover:text-blue-700 font-medium">
                Rotate credential
              </button>{' '}
              feature.
            </p>
          </div>
        )}

        {/* Expiration Date */}
        <div>
          <label htmlFor="expiresAt" className="block text-sm font-medium text-gray-700 mb-1">
            Expiration Date (Optional)
          </label>
          <input
            id="expiresAt"
            type="date"
            value={expiresAt}
            onChange={(e) => setExpiresAt(e.target.value)}
            disabled={loading}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed"
          />
        </div>
      </div>

      {/* Form Actions */}
      <div className="flex items-center justify-end gap-3 pt-4 border-t border-gray-200">
        <button
          type="button"
          onClick={onCancel}
          disabled={loading}
          className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={loading}
          className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? (isEditMode ? 'Saving...' : 'Creating...') : isEditMode ? 'Save' : 'Create'}
        </button>
      </div>
    </form>
  )
}
