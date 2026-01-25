/**
 * VersionSelector - Component for selecting workflow versions to compare
 * Provides dropdowns for base and compare version selection
 */

import { useState, useMemo } from 'react'
import { formatDistanceToNow, format } from 'date-fns'
import type { WorkflowVersion } from '../../../api/workflows'

interface VersionSelectorProps {
  versions: WorkflowVersion[]
  currentVersion: number
  baseVersionId: string | null
  compareVersionId: string | null
  onBaseVersionChange: (versionId: string | null) => void
  onCompareVersionChange: (versionId: string | null) => void
  loading?: boolean
}

export default function VersionSelector({
  versions,
  currentVersion,
  baseVersionId,
  compareVersionId,
  onBaseVersionChange,
  onCompareVersionChange,
  loading = false,
}: VersionSelectorProps) {
  const [baseOpen, setBaseOpen] = useState(false)
  const [compareOpen, setCompareOpen] = useState(false)

  const sortedVersions = useMemo(() => {
    return [...versions].sort((a, b) => b.version - a.version)
  }, [versions])

  const selectedBase = useMemo(() => {
    return versions.find((v) => v.id === baseVersionId)
  }, [versions, baseVersionId])

  const selectedCompare = useMemo(() => {
    return versions.find((v) => v.id === compareVersionId)
  }, [versions, compareVersionId])

  const handleBaseSelect = (version: WorkflowVersion) => {
    onBaseVersionChange(version.id)
    setBaseOpen(false)
  }

  const handleCompareSelect = (version: WorkflowVersion) => {
    onCompareVersionChange(version.id)
    setCompareOpen(false)
  }

  const handleSwapVersions = () => {
    const tempBase = baseVersionId
    onBaseVersionChange(compareVersionId)
    onCompareVersionChange(tempBase)
  }

  return (
    <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4 p-4 bg-gray-800 rounded-lg">
      {/* Base Version Selector */}
      <div className="flex-1 w-full sm:w-auto">
        <label className="block text-xs font-medium text-gray-400 mb-1">
          Base Version (older)
        </label>
        <VersionDropdown
          versions={sortedVersions}
          selectedVersion={selectedBase}
          currentVersion={currentVersion}
          isOpen={baseOpen}
          onToggle={() => setBaseOpen(!baseOpen)}
          onSelect={handleBaseSelect}
          onClose={() => setBaseOpen(false)}
          disabled={loading}
          excludeVersionId={compareVersionId}
          placeholder="Select base version"
        />
      </div>

      {/* Swap Button */}
      <div className="flex items-end pb-0.5">
        <button
          onClick={handleSwapVersions}
          disabled={loading || !baseVersionId || !compareVersionId}
          className="p-2 text-gray-400 hover:text-white hover:bg-gray-700 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          title="Swap versions"
          aria-label="Swap base and compare versions"
        >
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
          </svg>
        </button>
      </div>

      {/* Compare Version Selector */}
      <div className="flex-1 w-full sm:w-auto">
        <label className="block text-xs font-medium text-gray-400 mb-1">
          Compare Version (newer)
        </label>
        <VersionDropdown
          versions={sortedVersions}
          selectedVersion={selectedCompare}
          currentVersion={currentVersion}
          isOpen={compareOpen}
          onToggle={() => setCompareOpen(!compareOpen)}
          onSelect={handleCompareSelect}
          onClose={() => setCompareOpen(false)}
          disabled={loading}
          excludeVersionId={baseVersionId}
          placeholder="Select compare version"
        />
      </div>
    </div>
  )
}

// ============================================================================
// Version Dropdown Component
// ============================================================================

interface VersionDropdownProps {
  versions: WorkflowVersion[]
  selectedVersion?: WorkflowVersion
  currentVersion: number
  isOpen: boolean
  onToggle: () => void
  onSelect: (version: WorkflowVersion) => void
  onClose: () => void
  disabled?: boolean
  excludeVersionId?: string | null
  placeholder: string
}

function VersionDropdown({
  versions,
  selectedVersion,
  currentVersion,
  isOpen,
  onToggle,
  onSelect,
  onClose,
  disabled = false,
  excludeVersionId,
  placeholder,
}: VersionDropdownProps) {
  const filteredVersions = versions.filter((v) => v.id !== excludeVersionId)

  return (
    <div className="relative">
      {/* Selected Version Button */}
      <button
        type="button"
        onClick={onToggle}
        disabled={disabled}
        className="w-full flex items-center justify-between px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {selectedVersion ? (
          <div className="flex items-center gap-2 truncate">
            <span className="font-medium">v{selectedVersion.version}</span>
            {selectedVersion.version === currentVersion && (
              <span className="px-1.5 py-0.5 text-xs bg-primary-600 text-white rounded">
                Current
              </span>
            )}
            <span className="text-gray-400 text-xs truncate">
              {formatDistanceToNow(new Date(selectedVersion.createdAt), { addSuffix: true })}
            </span>
          </div>
        ) : (
          <span className="text-gray-400">{placeholder}</span>
        )}
        <svg
          className={`w-4 h-4 ml-2 transition-transform ${isOpen ? 'rotate-180' : ''}`}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {/* Dropdown Menu */}
      {isOpen && (
        <>
          {/* Backdrop */}
          <div
            className="fixed inset-0 z-10"
            onClick={onClose}
            aria-hidden="true"
          />

          {/* Menu */}
          <div className="absolute z-20 mt-1 w-full max-h-60 overflow-auto bg-gray-700 rounded-lg shadow-lg border border-gray-600">
            {filteredVersions.length === 0 ? (
              <div className="px-3 py-2 text-sm text-gray-400">
                No other versions available
              </div>
            ) : (
              filteredVersions.map((version) => (
                <button
                  key={version.id}
                  type="button"
                  onClick={() => onSelect(version)}
                  className={`w-full px-3 py-2 text-left hover:bg-gray-600 transition-colors ${
                    selectedVersion?.id === version.id ? 'bg-gray-600' : ''
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-white">v{version.version}</span>
                      {version.version === currentVersion && (
                        <span className="px-1.5 py-0.5 text-xs bg-primary-600 text-white rounded">
                          Current
                        </span>
                      )}
                    </div>
                  </div>
                  <div className="text-xs text-gray-400 mt-0.5">
                    {format(new Date(version.createdAt), 'MMM d, yyyy h:mm a')}
                  </div>
                  <div className="text-xs text-gray-500 mt-0.5">
                    {version.definition.nodes?.length || 0} nodes, {version.definition.edges?.length || 0} connections
                  </div>
                </button>
              ))
            )}
          </div>
        </>
      )}
    </div>
  )
}

// ============================================================================
// Quick Selection Helpers
// ============================================================================

interface QuickSelectButtonsProps {
  versions: WorkflowVersion[]
  currentVersion: number
  onSelectPair: (baseId: string, compareId: string) => void
  loading?: boolean
}

export function QuickSelectButtons({
  versions,
  currentVersion,
  onSelectPair,
  loading = false,
}: QuickSelectButtonsProps) {
  const sortedVersions = useMemo(() => {
    return [...versions].sort((a, b) => b.version - a.version)
  }, [versions])

  const canCompareWithPrevious = sortedVersions.length >= 2

  const handleCompareWithPrevious = () => {
    const currentIdx = sortedVersions.findIndex((v) => v.version === currentVersion)
    if (currentIdx >= 0 && currentIdx < sortedVersions.length - 1) {
      onSelectPair(sortedVersions[currentIdx + 1].id, sortedVersions[currentIdx].id)
    }
  }

  const handleCompareFirstWithLast = () => {
    if (sortedVersions.length >= 2) {
      onSelectPair(
        sortedVersions[sortedVersions.length - 1].id,
        sortedVersions[0].id
      )
    }
  }

  return (
    <div className="flex flex-wrap gap-2">
      <button
        onClick={handleCompareWithPrevious}
        disabled={loading || !canCompareWithPrevious}
        className="px-3 py-1.5 text-xs bg-gray-700 text-gray-300 hover:bg-gray-600 hover:text-white rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        Compare current with previous
      </button>
      {sortedVersions.length > 2 && (
        <button
          onClick={handleCompareFirstWithLast}
          disabled={loading}
          className="px-3 py-1.5 text-xs bg-gray-700 text-gray-300 hover:bg-gray-600 hover:text-white rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Compare first with latest
        </button>
      )}
    </div>
  )
}
