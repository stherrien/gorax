import React, { useMemo } from 'react'
import type { Credential, CredentialType } from '../../api/credentials'
import { useThemeContext } from '../../contexts/ThemeContext'

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
  const { isDark } = useThemeContext()
  const filteredAndSortedCredentials = useMemo(() => {
    let filtered = credentials || []

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

  if (loading && (!credentials || credentials.length === 0)) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className={isDark ? 'text-gray-400' : 'text-gray-500'}>Loading credentials...</div>
      </div>
    )
  }

  if (filteredAndSortedCredentials.length === 0) {
    return (
      <div className={`flex flex-col items-center justify-center h-64 ${isDark ? 'text-gray-400' : 'text-gray-500'}`}>
        <div className="text-lg font-medium">No credentials found</div>
        <div className="text-sm mt-2">Create your first credential to get started</div>
      </div>
    )
  }

  const count = filteredAndSortedCredentials.length
  const countText = count === 1 ? '1 credential' : `${count} credentials`

  return (
    <div className="space-y-4">
      <div className={`text-sm px-4 ${isDark ? 'text-gray-400' : 'text-gray-500'}`}>{countText}</div>

      <div className="space-y-2">
        {filteredAndSortedCredentials.map((credential) => {
          const expired = isExpired(credential.expiresAt)
          const isSelected = selectedId === credential.id

          return (
            <div
              key={credential.id}
              className={`border rounded-lg p-4 transition-colors ${
                isSelected
                  ? 'bg-primary-600/20 border-primary-500'
                  : isDark
                    ? 'bg-gray-800 border-gray-700 hover:border-gray-600'
                    : 'bg-white border-gray-200 hover:border-gray-300'
              } ${onSelect ? 'cursor-pointer' : ''}`}
              onClick={() => onSelect?.(credential.id)}
            >
              <div className="flex items-start justify-between">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h3
                      className={`text-base font-medium truncate ${isDark ? 'text-white' : 'text-gray-900'}`}
                      data-testid="credential-name"
                    >
                      {credential.name}
                    </h3>
                    <span
                      className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${isDark ? 'bg-gray-700 text-gray-300' : 'bg-gray-100 text-gray-600'}`}
                      data-testid="credential-type"
                    >
                      {formatCredentialType(credential.type)}
                    </span>
                    {expired && (
                      <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${isDark ? 'bg-red-500/20 text-red-400' : 'bg-red-100 text-red-700'}`}>
                        Expired
                      </span>
                    )}
                  </div>

                  {credential.description && (
                    <p className={`text-sm mt-1 ${isDark ? 'text-gray-400' : 'text-gray-500'}`}>{credential.description}</p>
                  )}

                  <div className={`flex items-center gap-4 mt-2 text-xs ${isDark ? 'text-gray-500' : 'text-gray-400'}`}>
                    <span>Created: {formatDate(credential.createdAt)}</span>
                    {credential.expiresAt && !expired && (
                      <span className={isDark ? 'text-orange-400' : 'text-orange-600'}>
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
                    className={`px-3 py-1 text-sm font-medium rounded disabled:opacity-50 disabled:cursor-not-allowed ${isDark ? 'text-primary-400 hover:text-primary-300 hover:bg-primary-500/10' : 'text-primary-600 hover:text-primary-700 hover:bg-primary-50'}`}
                  >
                    Test
                  </button>
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      onEdit(credential.id)
                    }}
                    className={`px-3 py-1 text-sm font-medium rounded ${isDark ? 'text-gray-400 hover:text-white hover:bg-gray-700' : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'}`}
                  >
                    Edit
                  </button>
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      onDelete(credential.id)
                    }}
                    className={`px-3 py-1 text-sm font-medium rounded ${isDark ? 'text-red-400 hover:text-red-300 hover:bg-red-500/10' : 'text-red-600 hover:text-red-700 hover:bg-red-50'}`}
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
