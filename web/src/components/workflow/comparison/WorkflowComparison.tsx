/**
 * WorkflowComparison - Main container for workflow version comparison
 * Orchestrates version selection, diff computation, and view modes
 */

import { useState, useMemo, useCallback } from 'react'
import type { WorkflowVersion } from '../../../api/workflows'
import type { WorkflowDiff } from '../../../types/diff'
import VersionSelector, { QuickSelectButtons } from './VersionSelector'
import SideBySideView from './SideBySideView'
import JsonDiffView from './JsonDiffView'
import { computeWorkflowDiff } from './DiffHighlight'

type ViewMode = 'visual' | 'json'

interface WorkflowComparisonProps {
  currentVersion: number
  versions: WorkflowVersion[]
  loading?: boolean
  error?: string | null
  onClose?: () => void
}

export default function WorkflowComparison({
  currentVersion,
  versions,
  loading = false,
  error = null,
  onClose,
}: WorkflowComparisonProps) {
  const [baseVersionId, setBaseVersionId] = useState<string | null>(null)
  const [compareVersionId, setCompareVersionId] = useState<string | null>(null)
  const [viewMode, setViewMode] = useState<ViewMode>('visual')
  const [showUnchanged, setShowUnchanged] = useState(false)

  // Get selected versions
  const baseVersion = useMemo(() => {
    return versions.find((v) => v.id === baseVersionId)
  }, [versions, baseVersionId])

  const compareVersion = useMemo(() => {
    return versions.find((v) => v.id === compareVersionId)
  }, [versions, compareVersionId])

  // Compute diff when both versions are selected
  const diff = useMemo<WorkflowDiff | null>(() => {
    if (!baseVersion || !compareVersion) return null

    return computeWorkflowDiff(
      baseVersion.definition,
      compareVersion.definition,
      baseVersion.version,
      compareVersion.version
    )
  }, [baseVersion, compareVersion])

  const handleSelectPair = useCallback((baseId: string, compareId: string) => {
    setBaseVersionId(baseId)
    setCompareVersionId(compareId)
  }, [])

  // Loading state
  if (loading) {
    return (
      <div className="flex flex-col h-full bg-gray-900 rounded-lg">
        <ComparisonHeader
          viewMode={viewMode}
          onViewModeChange={setViewMode}
          showUnchanged={showUnchanged}
          onShowUnchangedChange={setShowUnchanged}
          onClose={onClose}
        />
        <div className="flex-1 flex items-center justify-center">
          <div className="text-gray-400">Loading version history...</div>
        </div>
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className="flex flex-col h-full bg-gray-900 rounded-lg">
        <ComparisonHeader
          viewMode={viewMode}
          onViewModeChange={setViewMode}
          showUnchanged={showUnchanged}
          onShowUnchangedChange={setShowUnchanged}
          onClose={onClose}
        />
        <div className="flex-1 flex items-center justify-center">
          <div className="text-center">
            <div className="text-red-400 mb-4">{error}</div>
          </div>
        </div>
      </div>
    )
  }

  // No versions available
  if (versions.length < 2) {
    return (
      <div className="flex flex-col h-full bg-gray-900 rounded-lg">
        <ComparisonHeader
          viewMode={viewMode}
          onViewModeChange={setViewMode}
          showUnchanged={showUnchanged}
          onShowUnchangedChange={setShowUnchanged}
          onClose={onClose}
        />
        <div className="flex-1 flex items-center justify-center">
          <div className="text-center text-gray-400">
            <p className="mb-2">At least two versions are required to compare.</p>
            <p className="text-sm">
              {versions.length === 0
                ? 'No version history available.'
                : 'Only one version exists.'}
            </p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full bg-gray-900 rounded-lg overflow-hidden">
      {/* Header */}
      <ComparisonHeader
        viewMode={viewMode}
        onViewModeChange={setViewMode}
        showUnchanged={showUnchanged}
        onShowUnchangedChange={setShowUnchanged}
        onClose={onClose}
      />

      {/* Version Selection */}
      <div className="p-4 border-b border-gray-700">
        <VersionSelector
          versions={versions}
          currentVersion={currentVersion}
          baseVersionId={baseVersionId}
          compareVersionId={compareVersionId}
          onBaseVersionChange={setBaseVersionId}
          onCompareVersionChange={setCompareVersionId}
          loading={loading}
        />

        {/* Quick Selection */}
        <div className="mt-3">
          <QuickSelectButtons
            versions={versions}
            currentVersion={currentVersion}
            onSelectPair={handleSelectPair}
            loading={loading}
          />
        </div>
      </div>

      {/* Diff Content */}
      <div className="flex-1 min-h-0 overflow-hidden">
        {!baseVersion || !compareVersion ? (
          <div className="h-full flex items-center justify-center">
            <div className="text-center text-gray-400">
              <svg
                className="w-16 h-16 mx-auto mb-4 text-gray-600"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={1}
                  d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                />
              </svg>
              <p className="text-lg font-medium mb-1">Select versions to compare</p>
              <p className="text-sm text-gray-500">
                Choose a base version and a compare version to see the differences
              </p>
            </div>
          </div>
        ) : diff && viewMode === 'visual' ? (
          <SideBySideView
            baseDefinition={baseVersion.definition}
            compareDefinition={compareVersion.definition}
            baseVersion={baseVersion.version}
            compareVersion={compareVersion.version}
            diff={diff}
            showUnchanged={showUnchanged}
          />
        ) : diff && viewMode === 'json' ? (
          <JsonDiffView
            baseDefinition={baseVersion.definition}
            compareDefinition={compareVersion.definition}
            baseVersion={baseVersion.version}
            compareVersion={compareVersion.version}
          />
        ) : null}
      </div>
    </div>
  )
}

// ============================================================================
// Header Component
// ============================================================================

interface ComparisonHeaderProps {
  viewMode: ViewMode
  onViewModeChange: (mode: ViewMode) => void
  showUnchanged: boolean
  onShowUnchangedChange: (show: boolean) => void
  onClose?: () => void
}

function ComparisonHeader({
  viewMode,
  onViewModeChange,
  showUnchanged,
  onShowUnchangedChange,
  onClose,
}: ComparisonHeaderProps) {
  return (
    <div className="flex items-center justify-between p-4 bg-gray-800 border-b border-gray-700">
      <h2 className="text-lg font-semibold text-white">Version Comparison</h2>

      <div className="flex items-center gap-4">
        {/* View Mode Toggle */}
        <div className="flex items-center bg-gray-700 rounded-lg p-0.5">
          <button
            onClick={() => onViewModeChange('visual')}
            className={`px-3 py-1.5 text-sm rounded transition-colors ${
              viewMode === 'visual'
                ? 'bg-primary-600 text-white'
                : 'text-gray-400 hover:text-white'
            }`}
          >
            Visual
          </button>
          <button
            onClick={() => onViewModeChange('json')}
            className={`px-3 py-1.5 text-sm rounded transition-colors ${
              viewMode === 'json'
                ? 'bg-primary-600 text-white'
                : 'text-gray-400 hover:text-white'
            }`}
          >
            JSON
          </button>
        </div>

        {/* Show Unchanged Toggle (only for visual mode) */}
        {viewMode === 'visual' && (
          <label className="flex items-center gap-2 text-sm text-gray-400">
            <input
              type="checkbox"
              checked={showUnchanged}
              onChange={(e) => onShowUnchangedChange(e.target.checked)}
              className="rounded bg-gray-700 border-gray-600 text-primary-600 focus:ring-primary-500"
            />
            Show unchanged
          </label>
        )}

        {/* Close Button */}
        {onClose && (
          <button
            onClick={onClose}
            className="p-2 text-gray-400 hover:text-white hover:bg-gray-700 rounded transition-colors"
            aria-label="Close comparison"
          >
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        )}
      </div>
    </div>
  )
}

// Re-exports are handled via index.ts
