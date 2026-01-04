import type { OAuthProvider } from '../../types/oauth'
import { PROVIDER_BRANDING } from '../../types/oauth'

interface OAuthProviderCardProps {
  provider: OAuthProvider
  isConnected: boolean
  onConnect: () => void
}

export function OAuthProviderCard({
  provider,
  isConnected,
  onConnect,
}: OAuthProviderCardProps) {
  const branding = PROVIDER_BRANDING[provider.provider_key] || {
    name: provider.name,
    color: '#666',
    icon: '🔗',
    iconBg: '#f0f0f0',
  }

  return (
    <div className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between">
        <div className="flex items-start gap-3">
          <div
            className="w-12 h-12 rounded-lg flex items-center justify-center text-2xl"
            style={{ backgroundColor: branding.iconBg }}
          >
            {branding.icon}
          </div>
          <div>
            <h3 className="font-semibold text-gray-900">{branding.name}</h3>
            <p className="text-sm text-gray-600 mt-1">{provider.description}</p>
            {provider.default_scopes && provider.default_scopes.length > 0 && (
              <div className="mt-2">
                <p className="text-xs text-gray-500 mb-1">Default scopes:</p>
                <div className="flex flex-wrap gap-1">
                  {provider.default_scopes.map((scope) => (
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
        <div>
          {isConnected ? (
            <span className="inline-flex items-center gap-1 text-sm text-green-600 bg-green-50 px-3 py-1 rounded-full">
              <svg
                className="w-4 h-4"
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
              Connected
            </span>
          ) : (
            <button
              onClick={onConnect}
              className="px-4 py-2 text-sm font-medium text-white rounded-lg transition-colors"
              style={{ backgroundColor: branding.color }}
              onMouseEnter={(e) => {
                const target = e.currentTarget
                target.style.opacity = '0.9'
              }}
              onMouseLeave={(e) => {
                const target = e.currentTarget
                target.style.opacity = '1'
              }}
            >
              Connect
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
