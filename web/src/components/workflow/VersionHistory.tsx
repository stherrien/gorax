import { useState, useEffect, useCallback } from 'react'
import { workflowAPI, type WorkflowVersion } from '../../api/workflows'
import { formatDistanceToNow } from 'date-fns'

interface VersionHistoryProps {
  workflowId: string
  currentVersion: number
  onRestore?: (version: number) => void
  onClose?: () => void
}

export default function VersionHistory({ workflowId, currentVersion, onRestore, onClose }: VersionHistoryProps) {
  const [versions, setVersions] = useState<WorkflowVersion[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [restoring, setRestoring] = useState<number | null>(null)
  const [selectedVersion, setSelectedVersion] = useState<WorkflowVersion | null>(null)

  const loadVersions = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await workflowAPI.listVersions(workflowId)
      setVersions(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load versions')
    } finally {
      setLoading(false)
    }
  }, [workflowId])

  // Load versions on mount
  useEffect(() => {
    loadVersions()
  }, [loadVersions])

  async function handleRestore(version: number) {
    if (!confirm(`Are you sure you want to restore to version ${version}? This will create a new version with the old definition.`)) {
      return
    }

    try {
      setRestoring(version)
      await workflowAPI.restoreVersion(workflowId, version)
      if (onRestore) {
        onRestore(version)
      }
      // Reload versions to show the new one
      await loadVersions()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to restore version')
    } finally {
      setRestoring(null)
    }
  }

  async function handleViewVersion(version: WorkflowVersion) {
    setSelectedVersion(version)
  }

  function handleClosePreview() {
    setSelectedVersion(null)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-gray-400">Loading version history...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-4">
        <div className="text-red-400 mb-4">{error}</div>
        <button
          onClick={loadVersions}
          className="px-4 py-2 bg-primary-600 text-white rounded hover:bg-primary-700"
        >
          Retry
        </button>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-gray-700">
        <h2 className="text-lg font-semibold text-white">Version History</h2>
        {onClose && (
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-white transition-colors"
            aria-label="Close"
          >
            ✕
          </button>
        )}
      </div>

      {/* Version List */}
      <div className="flex-1 overflow-y-auto">
        {versions.length === 0 ? (
          <div className="p-8 text-center text-gray-400">
            No version history available
          </div>
        ) : (
          <div className="divide-y divide-gray-700">
            {versions.map((version) => (
              <div
                key={version.id}
                className={`p-4 hover:bg-gray-800 transition-colors ${
                  version.version === currentVersion ? 'bg-gray-800/50' : ''
                }`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-white font-medium">
                        Version {version.version}
                      </span>
                      {version.version === currentVersion && (
                        <span className="px-2 py-0.5 text-xs bg-primary-600 text-white rounded">
                          Current
                        </span>
                      )}
                    </div>
                    <div className="text-sm text-gray-400">
                      {formatDistanceToNow(new Date(version.createdAt), { addSuffix: true })}
                    </div>
                    <div className="text-xs text-gray-500 mt-1">
                      {version.definition.nodes?.length || 0} nodes, {version.definition.edges?.length || 0} edges
                    </div>
                  </div>

                  <div className="flex gap-2">
                    <button
                      onClick={() => handleViewVersion(version)}
                      className="px-3 py-1 text-sm text-primary-400 hover:text-primary-300 transition-colors"
                    >
                      Preview
                    </button>
                    {version.version !== currentVersion && (
                      <button
                        onClick={() => handleRestore(version.version)}
                        disabled={restoring === version.version}
                        className="px-3 py-1 text-sm bg-primary-600 text-white rounded hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                      >
                        {restoring === version.version ? 'Restoring...' : 'Restore'}
                      </button>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Preview Modal */}
      {selectedVersion && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={handleClosePreview}>
          <div className="bg-gray-800 rounded-lg max-w-2xl w-full max-h-[80vh] overflow-hidden" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between p-4 border-b border-gray-700">
              <h3 className="text-lg font-semibold text-white">
                Version {selectedVersion.version} Preview
              </h3>
              <button
                onClick={handleClosePreview}
                className="text-gray-400 hover:text-white transition-colors"
              >
                ✕
              </button>
            </div>
            <div className="p-4 overflow-y-auto max-h-[60vh]">
              <div className="text-sm text-gray-400 mb-4">
                Created {formatDistanceToNow(new Date(selectedVersion.createdAt), { addSuffix: true })}
              </div>
              <pre className="bg-gray-900 p-4 rounded text-sm text-gray-300 overflow-x-auto">
                {JSON.stringify(selectedVersion.definition, null, 2)}
              </pre>
            </div>
            <div className="flex justify-end gap-2 p-4 border-t border-gray-700">
              <button
                onClick={handleClosePreview}
                className="px-4 py-2 text-gray-400 hover:text-white transition-colors"
              >
                Close
              </button>
              {selectedVersion.version !== currentVersion && (
                <button
                  onClick={() => {
                    handleClosePreview()
                    handleRestore(selectedVersion.version)
                  }}
                  disabled={restoring === selectedVersion.version}
                  className="px-4 py-2 bg-primary-600 text-white rounded hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {restoring === selectedVersion.version ? 'Restoring...' : 'Restore This Version'}
                </button>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
