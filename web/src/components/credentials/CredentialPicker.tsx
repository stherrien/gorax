import React, { useState, useRef, useEffect, useMemo } from 'react'
import type { Credential, CredentialType } from '../../api/credentials'

export interface CredentialPickerProps {
  credentials: Credential[]
  value?: string
  filterType?: CredentialType
  placeholder?: string
  loading?: boolean
  disabled?: boolean
  onSelect: (template: string) => void
  onCreate?: () => void
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

const extractCredentialName = (template: string): string | null => {
  const match = template.match(/\{\{credentials\.(.+?)\}\}/)
  return match ? match[1] : null
}

export const CredentialPicker: React.FC<CredentialPickerProps> = ({
  credentials,
  value,
  filterType,
  placeholder = 'Select Credential',
  loading = false,
  disabled = false,
  onSelect,
  onCreate,
}) => {
  const [isOpen, setIsOpen] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')
  const dropdownRef = useRef<HTMLDivElement>(null)

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside)
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [isOpen])

  // Filter and search credentials
  const filteredCredentials = useMemo(() => {
    let filtered = credentials

    // Apply type filter
    if (filterType) {
      filtered = filtered.filter((cred) => cred.type === filterType)
    }

    // Apply search
    if (searchTerm) {
      const search = searchTerm.toLowerCase()
      filtered = filtered.filter(
        (cred) =>
          cred.name.toLowerCase().includes(search) ||
          cred.description?.toLowerCase().includes(search)
      )
    }

    return filtered
  }, [credentials, filterType, searchTerm])

  const handleSelect = (credential: Credential) => {
    const template = `{{credentials.${credential.name}}}`
    onSelect(template)
    setIsOpen(false)
    setSearchTerm('')
  }

  const handleToggle = () => {
    if (!disabled && !loading) {
      setIsOpen(!isOpen)
    }
  }

  // Get display text
  const displayText = useMemo(() => {
    if (value) {
      const credName = extractCredentialName(value)
      if (credName) {
        const credential = credentials.find((c) => c.name === credName)
        if (credential) {
          return credential.name
        }
      }
    }
    return placeholder
  }, [value, credentials, placeholder])

  const showSearch = credentials.length > 5

  return (
    <div ref={dropdownRef} className="relative">
      <button
        type="button"
        onClick={handleToggle}
        disabled={disabled || loading}
        className="w-full px-3 py-2 text-left bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed flex items-center justify-between"
      >
        <span className={value ? 'text-gray-900' : 'text-gray-500'}>
          {loading ? 'Loading...' : displayText}
        </span>
        <svg
          className={`w-5 h-5 text-gray-400 transition-transform ${isOpen ? 'rotate-180' : ''}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {isOpen && (
        <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-80 overflow-auto">
          {showSearch && (
            <div className="sticky top-0 bg-white border-b border-gray-200 p-2">
              <input
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="Search credentials..."
                className="w-full px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                onClick={(e) => e.stopPropagation()}
              />
            </div>
          )}

          {filteredCredentials.length > 0 && (
            <div className="px-2 py-1 text-xs text-gray-500">
              {filteredCredentials.length === 1
                ? '1 credential'
                : `${filteredCredentials.length} credentials`}
            </div>
          )}

          {filteredCredentials.length > 0 ? (
            <ul className="py-1">
              {filteredCredentials.map((credential) => (
                <li key={credential.id}>
                  <button
                    type="button"
                    onClick={() => handleSelect(credential)}
                    className="w-full px-4 py-2 text-left hover:bg-gray-100 focus:bg-gray-100 focus:outline-none"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-medium text-gray-900 truncate">
                          {credential.name}
                        </div>
                        {credential.description && (
                          <div className="text-xs text-gray-500 truncate">
                            {credential.description}
                          </div>
                        )}
                      </div>
                      <span className="ml-2 px-2 py-1 text-xs font-medium bg-gray-100 text-gray-700 rounded">
                        {formatCredentialType(credential.type)}
                      </span>
                    </div>
                  </button>
                </li>
              ))}
            </ul>
          ) : (
            <div className="px-4 py-8 text-center text-gray-500">
              <div className="text-sm font-medium">No credentials available</div>
              {onCreate && (
                <button
                  type="button"
                  onClick={() => {
                    onCreate()
                    setIsOpen(false)
                  }}
                  className="mt-2 text-sm text-blue-600 hover:text-blue-700 font-medium"
                >
                  Create a credential
                </button>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
