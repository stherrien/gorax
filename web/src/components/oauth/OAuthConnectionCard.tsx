import { useState } from 'react'
import type { OAuthConnection } from '../../types/oauth'
import { PROVIDER_BRANDING } from '../../types/oauth'
import { useRevokeConnection, useTestConnection } from '../../hooks/useOAuth'

interface OAuthConnectionCardProps {
  connection: OAuthConnection
}

export function OAuthConnectionCard({ connection }: OAuthConnectionCardProps) {
  const [showRevokeConfirm, setShowRevokeConfirm] = useState(false)
  const revokeConnection = useRevokeConnection()
  const testConnection = useTestConnection()

  const branding = PROVIDER_BRANDING[connection.provider_key] || {
    name: connection.provider_key,
    color: '#666',
    icon: '🔗',
    iconBg: '#f0f0f0',
  }

  const handleRevoke = async () => {
    try {
      await revokeConnection.mutateAsync(connection.id)
      setShowRevokeConfirm(false)
    } catch (error) {
      console.error('Failed to revoke connection:', error)
    }
  }

  const handleTest = async () => {
    try {
      await testConnection.mutateAsync(connection.id)
    } catch (error) {
      console.error('Failed to test connection:', error)
    }
  }

  const getStatusColor = () => {
    switch (connection.status) {
      case 'active':
        return 'text-green-600 bg-green-50'
      case 'expired':
        return 'text-yellow-600 bg-yellow-50'
      case 'revoked':
        return 'text-red-600 bg-red-50'
      default:
        return 'text-gray-600 bg-gray-50'
    }
  }

  const getStatusIcon = () => {
    switch (connection.status) {
      case 'active':
        return (
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M5 13l4 4L19 7"
            />
          </svg>
        )
      case 'expired':
        return (
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        )
      case 'revoked':
        return (
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        )
      default:
        return null
    }
  }

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return 'Never'
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  }

  const isExpiringSoon = () => {
    if (!connection.token_expiry) return false
    const expiry = new Date(connection.token_expiry)
    const fiveMinutes = 5 * 60 * 1000
    return expiry.getTime() - Date.now() < fiveMinutes
  }

  return (
    <div className="border border-gray-200 rounded-lg p-4">
      <div className="flex items-start justify-between">
        <div className="flex items-start gap-3 flex-1">
          <div
            className="w-12 h-12 rounded-lg flex items-center justify-center text-2xl flex-shrink-0"
            style={{ backgroundColor: branding.iconBg }}
          >
            {branding.icon}
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <h3 className="font-semibold text-gray-900">{branding.name}</h3>
              <span
                className={`inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full ${getStatusColor()}`}
              >
                {getStatusIcon()}
                {connection.status}
              </span>
            </div>
            {connection.provider_username && (
              <p className="text-sm text-gray-600 mt-1">@{connection.provider_username}</p>
            )}
            {connection.provider_email && (
              <p className="text-sm text-gray-600">{connection.provider_email}</p>
            )}
            <div className="mt-2 space-y-1">
              <div className="flex items-center gap-4 text-xs text-gray-500">
                <span>Connected: {formatDate(connection.created_at)}</span>
                {connection.last_used_at && (
                  <span>Last used: {formatDate(connection.last_used_at)}</span>
                )}
              </div>
              {connection.token_expiry && (
                <div className="text-xs">
                  <span
                    className={
                      isExpiringSoon() ? 'text-yellow-600 font-medium' : 'text-gray-500'
                    }
                  >
                    Expires: {formatDate(connection.token_expiry)}
                  </span>
                </div>
              )}
            </div>
            {connection.scopes && connection.scopes.length > 0 && (
              <div className="mt-2">
                <p className="text-xs text-gray-500 mb-1">Scopes:</p>
                <div className="flex flex-wrap gap-1">
                  {connection.scopes.map((scope) => (
                    <span
                      key={scope}
                      className="text-xs bg-gray-100 text-gray-700 px-2 py-0.5 rounded"
                    >
                      {scope}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
        <div className="flex flex-col gap-2 ml-4">
          {connection.status === 'active' && (
            <>
              <button
                onClick={handleTest}
                disabled={testConnection.isPending}
                className="px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-50"
              >
                {testConnection.isPending ? 'Testing...' : 'Test'}
              </button>
              {testConnection.isSuccess && (
                <div className="text-xs">
                  {testConnection.data?.success ? (
                    <span className="text-green-600">✓ Connection works</span>
                  ) : (
                    <span className="text-red-600">✗ {testConnection.data?.error}</span>
                  )}
                </div>
              )}
            </>
          )}
          {!showRevokeConfirm ? (
            <button
              onClick={() => setShowRevokeConfirm(true)}
              className="px-3 py-1.5 text-sm font-medium text-red-700 bg-white border border-red-300 rounded hover:bg-red-50"
            >
              Disconnect
            </button>
          ) : (
            <div className="flex flex-col gap-1">
              <p className="text-xs text-gray-600">Are you sure?</p>
              <div className="flex gap-1">
                <button
                  onClick={handleRevoke}
                  disabled={revokeConnection.isPending}
                  className="px-2 py-1 text-xs font-medium text-white bg-red-600 rounded hover:bg-red-700 disabled:opacity-50"
                >
                  {revokeConnection.isPending ? '...' : 'Yes'}
                </button>
                <button
                  onClick={() => setShowRevokeConfirm(false)}
                  className="px-2 py-1 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded hover:bg-gray-50"
                >
                  No
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
