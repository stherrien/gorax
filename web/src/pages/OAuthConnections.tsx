import { useState } from 'react'
import { useOAuthProviders, useOAuthConnections } from '../hooks/useOAuth'
import { OAuthProviderCard } from '../components/oauth/OAuthProviderCard'
import { OAuthConnectionList } from '../components/oauth/OAuthConnectionList'

export function OAuthConnections() {
  const { data: providers, isLoading: providersLoading } = useOAuthProviders()
  const { data: connections, isLoading: connectionsLoading } = useOAuthConnections()
  const [connectionSuccess, setConnectionSuccess] = useState<string | null>(null)

  // Filter active providers
  const activeProviders = providers?.filter((p) => p.status === 'active') || []

  // Get connection status for all providers
  const connectedProviderKeys = new Set(
    connections?.filter((c) => c.status === 'active').map((c) => c.provider_key) || []
  )

  // Separate connected and available providers
  const availableProviders = activeProviders.filter(
    (provider) => !connectedProviderKeys.has(provider.provider_key)
  )

  const handleConnectionSuccess = (providerKey: string) => {
    setConnectionSuccess(providerKey)
    setTimeout(() => setConnectionSuccess(null), 5000)
  }

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-2">OAuth Connections</h1>
        <p className="text-gray-600">
          Connect your OAuth providers to use them in workflows. All tokens are securely encrypted
          and never exposed.
        </p>
      </div>

      {/* Success message */}
      {connectionSuccess && (
        <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
          <div className="flex items-center gap-2">
            <svg
              className="w-5 h-5 text-green-600"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M5 13l4 4L19 7"
              />
            </svg>
            <span className="text-green-800 font-medium">
              Successfully connected {connectionSuccess}!
            </span>
          </div>
        </div>
      )}

      {/* HTTPS Warning */}
      {window.location.protocol !== 'https:' && window.location.hostname !== 'localhost' && (
        <div className="mb-6 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
          <div className="flex items-start gap-2">
            <svg
              className="w-5 h-5 text-yellow-600 mt-0.5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
            <div>
              <h3 className="text-yellow-900 font-medium">HTTPS Required</h3>
              <p className="text-yellow-800 text-sm mt-1">
                OAuth connections require HTTPS for security. Please use HTTPS in production.
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Connected Providers Section */}
      {connections && connections.length > 0 && (
        <div className="mb-12">
          <div className="mb-4">
            <h2 className="text-xl font-semibold text-gray-900">Your Connections</h2>
            <p className="text-sm text-gray-600 mt-1">
              Manage your active OAuth connections. You can test or disconnect them at any time.
            </p>
          </div>
          <OAuthConnectionList connections={connections} isLoading={connectionsLoading} />
        </div>
      )}

      {/* Available Providers Section */}
      <div>
        <div className="mb-4">
          <h2 className="text-xl font-semibold text-gray-900">Available Providers</h2>
          <p className="text-sm text-gray-600 mt-1">
            Connect to OAuth providers to use them in your workflows.
          </p>
        </div>

        {providersLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="border border-gray-200 rounded-lg p-4 animate-pulse">
                <div className="flex items-start gap-3">
                  <div className="w-12 h-12 bg-gray-200 rounded-lg"></div>
                  <div className="flex-1">
                    <div className="h-5 bg-gray-200 rounded w-32 mb-2"></div>
                    <div className="h-4 bg-gray-200 rounded w-full"></div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : availableProviders.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {availableProviders.map((provider) => (
              <OAuthProviderCard
                key={provider.id}
                provider={provider}
                isConnected={false}
                onConnect={() => {
                  // Connection handled by OAuthConnectButton in card
                  // Just show success message after popup closes
                  handleConnectionSuccess(provider.name)
                }}
              />
            ))}
          </div>
        ) : connections && connections.length > 0 ? (
          <div className="text-center py-12 px-4 border-2 border-dashed border-gray-300 rounded-lg">
            <div className="text-4xl mb-3">🎉</div>
            <h3 className="text-lg font-medium text-gray-900 mb-2">All Connected!</h3>
            <p className="text-gray-600">You've connected all available OAuth providers.</p>
          </div>
        ) : (
          <div className="text-center py-12 px-4 border-2 border-dashed border-gray-300 rounded-lg">
            <div className="text-4xl mb-3">🔌</div>
            <h3 className="text-lg font-medium text-gray-900 mb-2">No Providers Available</h3>
            <p className="text-gray-600 max-w-md mx-auto">
              No OAuth providers are currently configured. Contact your administrator to set up
              OAuth providers.
            </p>
          </div>
        )}
      </div>

      {/* Help Section */}
      <div className="mt-12 p-6 bg-blue-50 border border-blue-200 rounded-lg">
        <h3 className="text-lg font-semibold text-blue-900 mb-2">How OAuth Connections Work</h3>
        <ul className="space-y-2 text-sm text-blue-800">
          <li className="flex items-start gap-2">
            <svg
              className="w-5 h-5 text-blue-600 mt-0.5 flex-shrink-0"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span>
              <strong>Secure:</strong> All tokens are encrypted using AES-256-GCM encryption.
            </span>
          </li>
          <li className="flex items-start gap-2">
            <svg
              className="w-5 h-5 text-blue-600 mt-0.5 flex-shrink-0"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span>
              <strong>Automatic Refresh:</strong> Tokens are automatically refreshed when they
              expire.
            </span>
          </li>
          <li className="flex items-start gap-2">
            <svg
              className="w-5 h-5 text-blue-600 mt-0.5 flex-shrink-0"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span>
              <strong>Revocable:</strong> You can disconnect at any time to revoke access.
            </span>
          </li>
          <li className="flex items-start gap-2">
            <svg
              className="w-5 h-5 text-blue-600 mt-0.5 flex-shrink-0"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span>
              <strong>Workflow Integration:</strong> Use connected providers in workflow actions.
            </span>
          </li>
        </ul>
      </div>
    </div>
  )
}
