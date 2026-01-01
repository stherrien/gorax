import React, { useState, useEffect } from 'react'
import { useCredentialStore } from '../stores/credentialStore'
import { CredentialList } from '../components/credentials/CredentialList'
import { CredentialForm } from '../components/credentials/CredentialForm'
import { useThemeContext } from '../contexts/ThemeContext'
import type { Credential, CredentialType, CredentialCreateInput, CredentialUpdateInput } from '../api/credentials'

type ViewMode = 'list' | 'create' | 'edit'

export const CredentialManager: React.FC = () => {
  const { isDark } = useThemeContext()
  const {
    credentials,
    loading,
    error,
    fetchCredentials,
    createCredential,
    updateCredential,
    deleteCredential,
    testCredential,
    clearError,
  } = useCredentialStore()

  const [viewMode, setViewMode] = useState<ViewMode>('list')
  const [editingCredential, setEditingCredential] = useState<Credential | null>(null)
  const [deletingCredential, setDeletingCredential] = useState<Credential | null>(null)
  const [testResult, setTestResult] = useState<{ id: string; success: boolean; message: string } | null>(null)

  const [searchTerm, setSearchTerm] = useState('')
  const [filterType, setFilterType] = useState<CredentialType | ''>('')
  const [sortBy, setSortBy] = useState<'name' | 'created' | 'type'>('created')

  useEffect(() => {
    fetchCredentials()
  }, [fetchCredentials])

  const handleCreate = async (data: CredentialCreateInput | CredentialUpdateInput) => {
    // When creating, the form always provides all required fields for CredentialCreateInput
    await createCredential(data as CredentialCreateInput)
    if (!error) {
      setViewMode('list')
    }
  }

  const handleUpdate = async (data: CredentialCreateInput | CredentialUpdateInput) => {
    if (editingCredential) {
      await updateCredential(editingCredential.id, data as CredentialUpdateInput)
      if (!error) {
        setViewMode('list')
        setEditingCredential(null)
      }
    }
  }

  const handleDelete = async () => {
    if (deletingCredential) {
      await deleteCredential(deletingCredential.id)
      setDeletingCredential(null)
    }
  }

  const handleTest = async (id: string) => {
    try {
      const result = await testCredential(id)
      setTestResult({ id, success: result.success, message: result.message })
      // Auto-dismiss after 5 seconds
      setTimeout(() => setTestResult(null), 5000)
    } catch (err) {
      // Error handled by store
    }
  }

  const handleEdit = (id: string) => {
    const credential = (credentials || []).find((c) => c.id === id)
    if (credential) {
      setEditingCredential(credential)
      setViewMode('edit')
    }
  }

  const handleCancelForm = () => {
    setViewMode('list')
    setEditingCredential(null)
  }

  const handleRefresh = () => {
    fetchCredentials()
  }

  return (
    <div className="container mx-auto px-4 py-8 max-w-7xl">
      <div className="mb-8">
        <div className="flex items-center justify-between mb-4">
          <h1 className={`text-3xl font-bold ${isDark ? 'text-white' : 'text-gray-900'}`}>Credentials</h1>
          {viewMode === 'list' && (
            <div className="flex gap-2">
              <button
                onClick={handleRefresh}
                disabled={loading}
                className={`px-4 py-2 text-sm font-medium rounded-md disabled:opacity-50 disabled:cursor-not-allowed ${isDark ? 'text-gray-300 bg-gray-800 border border-gray-700 hover:bg-gray-700' : 'text-gray-700 bg-white border border-gray-300 hover:bg-gray-50'}`}
              >
                Refresh
              </button>
              <button
                onClick={() => setViewMode('create')}
                className="px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded-md hover:bg-primary-700"
              >
                Create Credential
              </button>
            </div>
          )}
        </div>

        {error && (
          <div className={`rounded-md p-4 flex items-start justify-between ${isDark ? 'bg-red-500/10 border border-red-500/30' : 'bg-red-50 border border-red-200'}`}>
            <p className={`text-sm ${isDark ? 'text-red-400' : 'text-red-800'}`}>{error}</p>
            <button
              onClick={clearError}
              className={isDark ? 'text-red-400 hover:text-red-300' : 'text-red-600 hover:text-red-800'}
              aria-label="Dismiss error"
            >
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fillRule="evenodd"
                  d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
                  clipRule="evenodd"
                />
              </svg>
            </button>
          </div>
        )}

        {testResult && (
          <div
            className={`border rounded-md p-4 ${
              testResult.success
                ? isDark ? 'bg-green-500/10 border-green-500/30' : 'bg-green-50 border-green-200'
                : isDark ? 'bg-red-500/10 border-red-500/30' : 'bg-red-50 border-red-200'
            }`}
          >
            <p
              className={`text-sm ${
                testResult.success
                  ? isDark ? 'text-green-400' : 'text-green-800'
                  : isDark ? 'text-red-400' : 'text-red-800'
              }`}
            >
              {testResult.message}
            </p>
          </div>
        )}
      </div>

      {viewMode === 'list' && (
        <div>
          {/* Filters and Search */}
          <div className={`rounded-lg p-4 mb-6 ${isDark ? 'bg-gray-800 border border-gray-700' : 'bg-white border border-gray-200'}`}>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label htmlFor="search" className={`block text-sm font-medium mb-1 ${isDark ? 'text-gray-300' : 'text-gray-700'}`}>
                  Search
                </label>
                <input
                  id="search"
                  type="text"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  placeholder="Search credentials..."
                  className={`w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500 ${isDark ? 'bg-gray-900 text-white border border-gray-700 placeholder-gray-500' : 'bg-white text-gray-900 border border-gray-300 placeholder-gray-400'}`}
                />
              </div>

              <div>
                <label htmlFor="filterType" className={`block text-sm font-medium mb-1 ${isDark ? 'text-gray-300' : 'text-gray-700'}`}>
                  Filter by type
                </label>
                <select
                  id="filterType"
                  value={filterType}
                  onChange={(e) => setFilterType(e.target.value as CredentialType | '')}
                  className={`w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500 ${isDark ? 'bg-gray-900 text-white border border-gray-700' : 'bg-white text-gray-900 border border-gray-300'}`}
                >
                  <option value="">All types</option>
                  <option value="api_key">API Key</option>
                  <option value="oauth2">OAuth2</option>
                  <option value="basic_auth">Basic Auth</option>
                  <option value="bearer_token">Bearer Token</option>
                </select>
              </div>

              <div>
                <label htmlFor="sortBy" className={`block text-sm font-medium mb-1 ${isDark ? 'text-gray-300' : 'text-gray-700'}`}>
                  Sort by
                </label>
                <select
                  id="sortBy"
                  value={sortBy}
                  onChange={(e) => setSortBy(e.target.value as 'name' | 'created' | 'type')}
                  className={`w-full px-3 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500 ${isDark ? 'bg-gray-900 text-white border border-gray-700' : 'bg-white text-gray-900 border border-gray-300'}`}
                >
                  <option value="created">Created date</option>
                  <option value="name">Name</option>
                  <option value="type">Type</option>
                </select>
              </div>
            </div>
          </div>

          {/* Credential List */}
          <CredentialList
            credentials={credentials || []}
            loading={loading}
            searchTerm={searchTerm}
            filterType={filterType || undefined}
            sortBy={sortBy}
            onSelect={() => {}}
            onEdit={handleEdit}
            onDelete={(id) => {
              const credential = (credentials || []).find((c) => c.id === id)
              if (credential) {
                setDeletingCredential(credential)
              }
            }}
            onTest={handleTest}
          />
        </div>
      )}

      {viewMode === 'create' && (
        <div className={`rounded-lg p-6 ${isDark ? 'bg-gray-800 border border-gray-700' : 'bg-white border border-gray-200'}`}>
          <CredentialForm
            loading={loading}
            error={error}
            onSubmit={handleCreate}
            onCancel={handleCancelForm}
          />
        </div>
      )}

      {viewMode === 'edit' && editingCredential && (
        <div className={`rounded-lg p-6 ${isDark ? 'bg-gray-800 border border-gray-700' : 'bg-white border border-gray-200'}`}>
          <CredentialForm
            credential={editingCredential}
            loading={loading}
            error={error}
            onSubmit={handleUpdate}
            onCancel={handleCancelForm}
          />
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {deletingCredential && (
        <div className="fixed inset-0 bg-black bg-opacity-70 flex items-center justify-center z-50">
          <div className={`rounded-lg p-6 max-w-md w-full mx-4 ${isDark ? 'bg-gray-800 border border-gray-700' : 'bg-white border border-gray-200'}`}>
            <h3 className={`text-lg font-semibold mb-4 ${isDark ? 'text-white' : 'text-gray-900'}`}>Delete Credential</h3>
            <p className={`text-sm mb-6 ${isDark ? 'text-gray-400' : 'text-gray-600'}`}>
              Are you sure you want to delete the credential "{deletingCredential.name}"? This action
              cannot be undone.
            </p>
            <div className="flex items-center justify-end gap-3">
              <button
                onClick={() => setDeletingCredential(null)}
                disabled={loading}
                className={`px-4 py-2 text-sm font-medium rounded-md disabled:opacity-50 disabled:cursor-not-allowed ${isDark ? 'text-gray-300 bg-gray-700 border border-gray-600 hover:bg-gray-600' : 'text-gray-700 bg-white border border-gray-300 hover:bg-gray-50'}`}
              >
                Cancel
              </button>
              <button
                onClick={handleDelete}
                disabled={loading}
                className="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-md hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loading ? 'Deleting...' : 'Confirm'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
