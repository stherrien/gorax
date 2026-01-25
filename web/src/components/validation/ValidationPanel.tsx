/**
 * ValidationPanel - Displays workflow validation results with actionable suggestions
 */

import { useMemo, useCallback } from 'react'
import { AlertCircle, AlertTriangle, Info, CheckCircle, ChevronDown, ChevronUp, Sparkles } from 'lucide-react'
import type { ValidationResult, ValidationIssue, ValidationSeverity } from '../../types/workflow'
import { getValidationSummary, filterIssuesBySeverity } from '../../utils/workflowValidator'

interface ValidationPanelProps {
  result: ValidationResult | null
  onNavigateToNode?: (nodeId: string) => void
  onAutoFix?: (issue: ValidationIssue) => void
  collapsed?: boolean
  onToggleCollapse?: () => void
}

export default function ValidationPanel({
  result,
  onNavigateToNode,
  onAutoFix,
  collapsed = false,
  onToggleCollapse,
}: ValidationPanelProps) {
  const errors = useMemo(
    () => (result ? filterIssuesBySeverity(result.issues, 'error') : []),
    [result]
  )

  const warnings = useMemo(
    () => (result ? filterIssuesBySeverity(result.issues, 'warning') : []),
    [result]
  )

  const infos = useMemo(
    () => (result ? filterIssuesBySeverity(result.issues, 'info') : []),
    [result]
  )

  const summary = useMemo(
    () => (result ? getValidationSummary(result) : 'No validation results'),
    [result]
  )

  const handleNodeClick = useCallback(
    (nodeId: string) => {
      onNavigateToNode?.(nodeId)
    },
    [onNavigateToNode]
  )

  if (!result) {
    return (
      <div className="bg-gray-800 border-t border-gray-700 p-4">
        <div className="text-gray-400 text-sm">Run validation to check your workflow</div>
      </div>
    )
  }

  return (
    <div className="bg-gray-800 border-t border-gray-700">
      {/* Header */}
      <button
        onClick={onToggleCollapse}
        className="w-full px-4 py-3 flex items-center justify-between hover:bg-gray-700/50 transition-colors"
      >
        <div className="flex items-center space-x-3">
          <StatusIcon valid={result.valid} />
          <span className={`font-medium text-sm ${result.valid ? 'text-green-400' : 'text-red-400'}`}>
            {summary}
          </span>
        </div>
        <div className="flex items-center space-x-2">
          {errors.length > 0 && (
            <span className="px-2 py-0.5 bg-red-900/30 text-red-400 text-xs rounded-full">
              {errors.length}
            </span>
          )}
          {warnings.length > 0 && (
            <span className="px-2 py-0.5 bg-yellow-900/30 text-yellow-400 text-xs rounded-full">
              {warnings.length}
            </span>
          )}
          {collapsed ? (
            <ChevronUp className="w-4 h-4 text-gray-400" />
          ) : (
            <ChevronDown className="w-4 h-4 text-gray-400" />
          )}
        </div>
      </button>

      {/* Issues List */}
      {!collapsed && result.issues.length > 0 && (
        <div className="border-t border-gray-700 max-h-64 overflow-y-auto">
          {/* Errors */}
          {errors.length > 0 && (
            <IssueSection
              title="Errors"
              issues={errors}
              severity="error"
              onNodeClick={handleNodeClick}
              onAutoFix={onAutoFix}
            />
          )}

          {/* Warnings */}
          {warnings.length > 0 && (
            <IssueSection
              title="Warnings"
              issues={warnings}
              severity="warning"
              onNodeClick={handleNodeClick}
              onAutoFix={onAutoFix}
            />
          )}

          {/* Info */}
          {infos.length > 0 && (
            <IssueSection
              title="Info"
              issues={infos}
              severity="info"
              onNodeClick={handleNodeClick}
              onAutoFix={onAutoFix}
            />
          )}
        </div>
      )}

      {/* Valid state */}
      {!collapsed && result.valid && result.issues.length === 0 && (
        <div className="p-4 border-t border-gray-700">
          <div className="flex items-center space-x-2 text-green-400">
            <CheckCircle className="w-5 h-5" />
            <span className="text-sm">Workflow is valid and ready to run</span>
          </div>
          {result.executionOrder && result.executionOrder.length > 0 && (
            <div className="mt-3">
              <div className="text-xs text-gray-500 uppercase tracking-wider mb-2">
                Execution Order
              </div>
              <div className="flex flex-wrap gap-1">
                {result.executionOrder.map((nodeId, index) => (
                  <span
                    key={nodeId}
                    className="inline-flex items-center px-2 py-1 bg-gray-700 rounded text-xs text-gray-300"
                  >
                    <span className="text-gray-500 mr-1">{index + 1}.</span>
                    {nodeId}
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// ============================================================================
// Sub-components
// ============================================================================

function StatusIcon({ valid }: { valid: boolean }) {
  if (valid) {
    return <CheckCircle className="w-5 h-5 text-green-400" />
  }
  return <AlertCircle className="w-5 h-5 text-red-400" />
}

interface IssueSectionProps {
  title: string
  issues: ValidationIssue[]
  severity: ValidationSeverity
  onNodeClick?: (nodeId: string) => void
  onAutoFix?: (issue: ValidationIssue) => void
}

function IssueSection({ title, issues, severity, onNodeClick, onAutoFix }: IssueSectionProps) {
  const bgColor = {
    error: 'bg-red-900/10',
    warning: 'bg-yellow-900/10',
    info: 'bg-blue-900/10',
  }[severity]

  return (
    <div className={bgColor}>
      <div className="px-4 py-2 border-b border-gray-700">
        <span className="text-xs text-gray-500 uppercase tracking-wider">{title}</span>
      </div>
      <div className="divide-y divide-gray-700/50">
        {issues.map((issue) => (
          <IssueItem
            key={issue.id}
            issue={issue}
            onNodeClick={onNodeClick}
            onAutoFix={onAutoFix}
          />
        ))}
      </div>
    </div>
  )
}

interface IssueItemProps {
  issue: ValidationIssue
  onNodeClick?: (nodeId: string) => void
  onAutoFix?: (issue: ValidationIssue) => void
}

function IssueItem({ issue, onNodeClick, onAutoFix }: IssueItemProps) {
  const Icon = {
    error: AlertCircle,
    warning: AlertTriangle,
    info: Info,
  }[issue.severity]

  const iconColor = {
    error: 'text-red-400',
    warning: 'text-yellow-400',
    info: 'text-blue-400',
  }[issue.severity]

  const handleNodeClick = () => {
    if (issue.nodeId && onNodeClick) {
      onNodeClick(issue.nodeId)
    }
  }

  const handleAutoFix = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (onAutoFix) {
      onAutoFix(issue)
    }
  }

  return (
    <div
      className={`px-4 py-3 hover:bg-gray-700/30 transition-colors ${
        issue.nodeId ? 'cursor-pointer' : ''
      }`}
      onClick={handleNodeClick}
    >
      <div className="flex items-start space-x-3">
        <Icon className={`w-4 h-4 mt-0.5 flex-shrink-0 ${iconColor}`} />
        <div className="flex-1 min-w-0">
          <div className="flex items-center space-x-2">
            <span className="text-sm text-white">{issue.message}</span>
            {issue.nodeId && (
              <span className="px-1.5 py-0.5 bg-gray-700 rounded text-xs text-gray-400">
                {issue.nodeId}
              </span>
            )}
            {issue.field && (
              <span className="px-1.5 py-0.5 bg-gray-700 rounded text-xs text-gray-500">
                {issue.field}
              </span>
            )}
          </div>
          {issue.suggestion && (
            <div className="mt-1 text-xs text-gray-400">{issue.suggestion}</div>
          )}
        </div>
        {issue.autoFixable && onAutoFix && (
          <button
            onClick={handleAutoFix}
            className="flex items-center space-x-1 px-2 py-1 bg-primary-600/20 text-primary-400 rounded text-xs hover:bg-primary-600/30 transition-colors"
            title="Auto-fix this issue"
          >
            <Sparkles className="w-3 h-3" />
            <span>Fix</span>
          </button>
        )}
      </div>
    </div>
  )
}

// ============================================================================
// Compact Validation Badge
// ============================================================================

interface ValidationBadgeProps {
  result: ValidationResult | null
  onClick?: () => void
}

export function ValidationBadge({ result, onClick }: ValidationBadgeProps) {
  if (!result) {
    return null
  }

  const errorCount = result.issues.filter((i) => i.severity === 'error').length
  const warningCount = result.issues.filter((i) => i.severity === 'warning').length

  if (result.valid && errorCount === 0 && warningCount === 0) {
    return (
      <button
        onClick={onClick}
        className="flex items-center space-x-1 px-2 py-1 bg-green-900/30 text-green-400 rounded text-sm hover:bg-green-900/50 transition-colors"
      >
        <CheckCircle className="w-4 h-4" />
        <span>Valid</span>
      </button>
    )
  }

  return (
    <button
      onClick={onClick}
      className="flex items-center space-x-2 px-2 py-1 bg-red-900/30 text-red-400 rounded text-sm hover:bg-red-900/50 transition-colors"
    >
      <AlertCircle className="w-4 h-4" />
      <span className="flex items-center space-x-1">
        {errorCount > 0 && <span>{errorCount} error{errorCount > 1 ? 's' : ''}</span>}
        {errorCount > 0 && warningCount > 0 && <span>,</span>}
        {warningCount > 0 && <span>{warningCount} warning{warningCount > 1 ? 's' : ''}</span>}
      </span>
    </button>
  )
}
