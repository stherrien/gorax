import React, { useState, useEffect } from 'react'
import type { Credential, CredentialType, CredentialCreateInput, CredentialUpdateInput } from '../../api/credentials'
import { useThemeContext } from '../../contexts/ThemeContext'

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
  const { isDark } = useThemeContext()
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

  // Theme-aware class helpers
  const labelClass = `block text-sm font-medium mb-1 ${isDark ? 'text-gray-300' : 'text-gray-700'}`
  const inputClass = `w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent disabled:cursor-not-allowed ${
    isDark
      ? 'bg-gray-900 text-white border border-gray-700 disabled:bg-gray-800 disabled:text-gray-500'
      : 'bg-white text-gray-900 border border-gray-300 disabled:bg-gray-100 disabled:text-gray-400'
  }`
  const errorClass = `mt-1 text-sm ${isDark ? 'text-red-400' : 'text-red-600'}`

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div>
        <h2 className={`text-2xl font-bold mb-6 ${isDark ? 'text-white' : 'text-gray-900'}`}>
          {isEditMode ? 'Edit Credential' : 'Create Credential'}
        </h2>
      </div>

      {error && (
        <div className={`rounded-md p-4 ${isDark ? 'bg-red-500/10 border border-red-500/30' : 'bg-red-50 border border-red-200'}`}>
          <p className={`text-sm ${isDark ? 'text-red-400' : 'text-red-800'}`}>{error}</p>
        </div>
      )}

      <div className="space-y-4">
        {/* Name */}
        <div>
          <label htmlFor="name" className={labelClass}>
            Name *
          </label>
          <input
            id="name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            disabled={loading}
            className={inputClass}
            placeholder="My API Credential"
          />
          {errors.name && <p className={errorClass}>{errors.name}</p>}
        </div>

        {/* Description */}
        <div>
          <label htmlFor="description" className={labelClass}>
            Description
          </label>
          <textarea
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            disabled={loading}
            rows={3}
            className={inputClass}
            placeholder="Optional description"
          />
        </div>

        {/* Type */}
        <div>
          <label htmlFor="type" className={labelClass}>
            Type *
          </label>
          <select
            id="type"
            value={type}
            onChange={(e) => setType(e.target.value as CredentialType)}
            disabled={loading || isEditMode}
            className={inputClass}
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
          <div className={`border-t pt-4 mt-4 ${isDark ? 'border-gray-700' : 'border-gray-200'}`}>
            <h3 className={`text-lg font-medium mb-4 ${isDark ? 'text-white' : 'text-gray-900'}`}>Credential Details</h3>

            {type === 'api_key' && (
              <div>
                <label htmlFor="apiKey" className={labelClass}>
                  API Key *
                </label>
                <input
                  id="apiKey"
                  type="password"
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  disabled={loading}
                  className={inputClass}
                  placeholder="sk-..."
                />
                {errors.apiKey && <p className={errorClass}>{errors.apiKey}</p>}
              </div>
            )}

            {type === 'oauth2' && (
              <div className="space-y-4">
                <div>
                  <label htmlFor="clientId" className={labelClass}>
                    Client ID *
                  </label>
                  <input
                    id="clientId"
                    type="text"
                    value={clientId}
                    onChange={(e) => setClientId(e.target.value)}
                    disabled={loading}
                    className={inputClass}
                  />
                  {errors.clientId && <p className={errorClass}>{errors.clientId}</p>}
                </div>

                <div>
                  <label htmlFor="clientSecret" className={labelClass}>
                    Client Secret *
                  </label>
                  <input
                    id="clientSecret"
                    type="password"
                    value={clientSecret}
                    onChange={(e) => setClientSecret(e.target.value)}
                    disabled={loading}
                    className={inputClass}
                  />
                  {errors.clientSecret && <p className={errorClass}>{errors.clientSecret}</p>}
                </div>

                <div>
                  <label htmlFor="authUrl" className={labelClass}>
                    Auth URL *
                  </label>
                  <input
                    id="authUrl"
                    type="url"
                    value={authUrl}
                    onChange={(e) => setAuthUrl(e.target.value)}
                    disabled={loading}
                    className={inputClass}
                    placeholder="https://auth.example.com"
                  />
                  {errors.authUrl && <p className={errorClass}>{errors.authUrl}</p>}
                </div>

                <div>
                  <label htmlFor="tokenUrl" className={labelClass}>
                    Token URL *
                  </label>
                  <input
                    id="tokenUrl"
                    type="url"
                    value={tokenUrl}
                    onChange={(e) => setTokenUrl(e.target.value)}
                    disabled={loading}
                    className={inputClass}
                    placeholder="https://token.example.com"
                  />
                  {errors.tokenUrl && <p className={errorClass}>{errors.tokenUrl}</p>}
                </div>
              </div>
            )}

            {type === 'basic_auth' && (
              <div className="space-y-4">
                <div>
                  <label htmlFor="username" className={labelClass}>
                    Username *
                  </label>
                  <input
                    id="username"
                    type="text"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    disabled={loading}
                    className={inputClass}
                  />
                  {errors.username && <p className={errorClass}>{errors.username}</p>}
                </div>

                <div>
                  <label htmlFor="password" className={labelClass}>
                    Password *
                  </label>
                  <input
                    id="password"
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    disabled={loading}
                    className={inputClass}
                  />
                  {errors.password && <p className={errorClass}>{errors.password}</p>}

                  {passwordStrength && (
                    <div className="mt-2">
                      <div className="flex items-center gap-2">
                        <span className={`text-xs ${isDark ? 'text-gray-400' : 'text-gray-500'}`}>Strength:</span>
                        <span
                          className={`text-xs font-medium ${
                            passwordStrength === 'Strong'
                              ? isDark ? 'text-green-400' : 'text-green-600'
                              : passwordStrength === 'Medium'
                              ? isDark ? 'text-yellow-400' : 'text-yellow-600'
                              : isDark ? 'text-red-400' : 'text-red-600'
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
                <label htmlFor="token" className={labelClass}>
                  Token *
                </label>
                <input
                  id="token"
                  type="password"
                  value={token}
                  onChange={(e) => setToken(e.target.value)}
                  disabled={loading}
                  className={inputClass}
                  placeholder="Bearer token"
                />
                {errors.token && <p className={errorClass}>{errors.token}</p>}
              </div>
            )}
          </div>
        )}

        {/* Rotate credential link (Edit mode only) */}
        {isEditMode && (
          <div className={`border-t pt-4 mt-4 ${isDark ? 'border-gray-700' : 'border-gray-200'}`}>
            <p className={`text-sm ${isDark ? 'text-gray-400' : 'text-gray-500'}`}>
              To update the credential value, use the{' '}
              <button type="button" className={`font-medium ${isDark ? 'text-primary-400 hover:text-primary-300' : 'text-primary-600 hover:text-primary-700'}`}>
                Rotate credential
              </button>{' '}
              feature.
            </p>
          </div>
        )}

        {/* Expiration Date */}
        <div>
          <label htmlFor="expiresAt" className={labelClass}>
            Expiration Date (Optional)
          </label>
          <input
            id="expiresAt"
            type="date"
            value={expiresAt}
            onChange={(e) => setExpiresAt(e.target.value)}
            disabled={loading}
            className={`${inputClass} ${isDark ? '[color-scheme:dark]' : ''}`}
          />
        </div>
      </div>

      {/* Form Actions */}
      <div className={`flex items-center justify-end gap-3 pt-4 border-t ${isDark ? 'border-gray-700' : 'border-gray-200'}`}>
        <button
          type="button"
          onClick={onCancel}
          disabled={loading}
          className={`px-4 py-2 text-sm font-medium rounded-md disabled:opacity-50 disabled:cursor-not-allowed ${isDark ? 'text-gray-300 bg-gray-700 border border-gray-600 hover:bg-gray-600' : 'text-gray-700 bg-white border border-gray-300 hover:bg-gray-50'}`}
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={loading}
          className="px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded-md hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? (isEditMode ? 'Saving...' : 'Creating...') : isEditMode ? 'Save' : 'Create'}
        </button>
      </div>
    </form>
  )
}
