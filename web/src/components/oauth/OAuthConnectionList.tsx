import type { OAuthConnection } from '../../types/oauth'
import { OAuthConnectionCard } from './OAuthConnectionCard'

interface OAuthConnectionListProps {
  connections: OAuthConnection[]
  isLoading?: boolean
}

export function OAuthConnectionList({ connections, isLoading }: OAuthConnectionListProps) {
  if (isLoading) {
    return (
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="border border-gray-200 rounded-lg p-4 animate-pulse">
            <div className="flex items-start gap-3">
              <div className="w-12 h-12 bg-gray-200 rounded-lg"></div>
              <div className="flex-1">
                <div className="h-5 bg-gray-200 rounded w-32 mb-2"></div>
                <div className="h-4 bg-gray-200 rounded w-48"></div>
              </div>
            </div>
          </div>
        ))}
      </div>
    )
  }

  if (!connections || connections.length === 0) {
    return (
      <div className="text-center py-12 px-4 border-2 border-dashed border-gray-300 rounded-lg">
        <div className="text-4xl mb-3">🔗</div>
        <h3 className="text-lg font-medium text-gray-900 mb-2">No OAuth Connections</h3>
        <p className="text-gray-600 max-w-md mx-auto">
          You haven't connected any OAuth providers yet. Connect a provider to use it in your
          workflows.
        </p>
      </div>
    )
  }

  // Group connections by provider
  const groupedConnections = connections.reduce(
    (acc, connection) => {
      const key = connection.provider_key
      if (!acc[key]) {
        acc[key] = []
      }
      acc[key].push(connection)
      return acc
    },
    {} as Record<string, OAuthConnection[]>
  )

  return (
    <div className="space-y-6">
      {Object.entries(groupedConnections).map(([providerKey, providerConnections]) => (
        <div key={providerKey}>
          <h3 className="text-sm font-medium text-gray-700 mb-3 uppercase">
            {providerKey} ({providerConnections.length})
          </h3>
          <div className="space-y-3">
            {providerConnections.map((connection) => (
              <OAuthConnectionCard key={connection.id} connection={connection} />
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}
