import React, { useMemo } from 'react'
import type { Credential, CredentialType } from '../../api/credentials'

export interface CredentialListProps {
  credentials: Credential[]
  loading?: boolean
  searchTerm?: string
  filterType?: CredentialType
  sortBy?: 'name' | 'created' | 'type'
  selectedId?: string
  onSelect?: (id: string) => void
  onEdit: (id: string) => void
  onDelete: (id: string) => void
  onTest: (id: string) => void
}

const formatCredentialType = (type: CredentialType): string => {
  const typeMap: Record<CredentialType, string> = {
    api_key: 'API Key',
    oauth2: 'OAuth2',
    basic_auth: 'Basic Auth',
    bearer_token: 'Bearer Token',
  }
  return typeMap[type] || type
}

const isExpired = (expiresAt?: string): boolean => {
  if (!expiresAt) return false
  return new Date(expiresAt) < new Date()
}

const formatDate = (dateStr: string): string => {
  const date = new Date(dateStr)
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

export const CredentialList: React.FC<CredentialListProps> = ({
  credentials,
  loading = false,
  searchTerm = '',
  filterType,
  sortBy = 'created',
  selectedId,
  onSelect,
  onEdit,
  onDelete,
  onTest,
}) => {
  const filteredAndSortedCredentials = useMemo(() => {
    let filtered = credentials

    // Apply search filter
    if (searchTerm) {
      const search = searchTerm.toLowerCase()
      filtered = filtered.filter(
        (cred) =>
          cred.name.toLowerCase().includes(search) ||
          cred.description?.toLowerCase().includes(search)
      )
    }

    // Apply type filter
    if (filterType) {
      filtered = filtered.filter((cred) => cred.type === filterType)
    }

    // Sort
    const sorted = [...filtered]
    switch (sortBy) {
      case 'name':
        sorted.sort((a, b) => a.name.localeCompare(b.name))
        break
      case 'type':
        sorted.sort((a, b) => a.type.localeCompare(b.type))
        break
      case 'created':
      default:
        sorted.sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime())
        break
    }

    return sorted
  }, [credentials, searchTerm, filterType, sortBy])

  if (loading && credentials.length === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-500">Loading credentials...</div>
      </div>
    )
  }

  if (filteredAndSortedCredentials.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-gray-500">
        <div className="text-lg font-medium">No credentials found</div>
        <div className="text-sm mt-2">Create your first credential to get started</div>
      </div>
    )
  }

  const count = filteredAndSortedCredentials.length
  const countText = count === 1 ? '1 credential' : `${count} credentials`

  return (
    <div className="space-y-4">
      <div className="text-sm text-gray-600 px-4">{countText}</div>

      <div className="space-y-2">
        {filteredAndSortedCredentials.map((credential) => {
          const expired = isExpired(credential.expiresAt)
          const isSelected = selectedId === credential.id

          return (
            <div
              key={credential.id}
              className={`border rounded-lg p-4 transition-colors ${
                isSelected
                  ? 'bg-blue-50 border-blue-300'
                  : 'bg-white border-gray-200 hover:border-gray-300'
              } ${onSelect ? 'cursor-pointer' : ''}`}
              onClick={() => onSelect?.(credential.id)}
            >
              <div className="flex items-start justify-between">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h3
                      className="text-base font-medium text-gray-900 truncate"
                      data-testid="credential-name"
                    >
                      {credential.name}
                    </h3>
                    <span
                      className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-700"
                      data-testid="credential-type"
                    >
                      {formatCredentialType(credential.type)}
                    </span>
                    {expired && (
                      <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-red-100 text-red-700">
                        Expired
                      </span>
                    )}
                  </div>

                  {credential.description && (
                    <p className="text-sm text-gray-600 mt-1">{credential.description}</p>
                  )}

                  <div className="flex items-center gap-4 mt-2 text-xs text-gray-500">
                    <span>Created: {formatDate(credential.createdAt)}</span>
                    {credential.expiresAt && !expired && (
                      <span className="text-orange-600">
                        Expires: {formatDate(credential.expiresAt)}
                      </span>
                    )}
                  </div>
                </div>

                <div className="flex items-center gap-2 ml-4">
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      onTest(credential.id)
                    }}
                    disabled={loading}
                    className="px-3 py-1 text-sm font-medium text-blue-600 hover:text-blue-700 hover:bg-blue-50 rounded disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Test
                  </button>
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      onEdit(credential.id)
                    }}
                    className="px-3 py-1 text-sm font-medium text-gray-600 hover:text-gray-700 hover:bg-gray-50 rounded"
                  >
                    Edit
                  </button>
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      onDelete(credential.id)
                    }}
                    className="px-3 py-1 text-sm font-medium text-red-600 hover:text-red-700 hover:bg-red-50 rounded"
                  >
                    Delete
                  </button>
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
