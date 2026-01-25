/**
 * JsonDiffView - Technical JSON diff viewer for workflow definitions
 * Shows line-by-line JSON comparison with syntax highlighting
 */

import { useMemo, useState, useCallback } from 'react'
import { diffLines, type Change } from 'diff'
import type { WorkflowDefinition } from '../../../api/workflows'

interface JsonDiffViewProps {
  baseDefinition: WorkflowDefinition
  compareDefinition: WorkflowDefinition
  baseVersion: number
  compareVersion: number
}

type DiffViewMode = 'unified' | 'split'

export default function JsonDiffView({
  baseDefinition,
  compareDefinition,
  baseVersion,
  compareVersion,
}: JsonDiffViewProps) {
  const [viewMode, setViewMode] = useState<DiffViewMode>('unified')
  const [showLineNumbers, setShowLineNumbers] = useState(true)

  const baseJson = useMemo(() => {
    return JSON.stringify(baseDefinition, null, 2)
  }, [baseDefinition])

  const compareJson = useMemo(() => {
    return JSON.stringify(compareDefinition, null, 2)
  }, [compareDefinition])

  const diffResult = useMemo(() => {
    return diffLines(baseJson, compareJson)
  }, [baseJson, compareJson])

  const stats = useMemo(() => {
    let additions = 0
    let deletions = 0

    for (const change of diffResult) {
      const lineCount = (change.value.match(/\n/g) || []).length
      if (change.added) additions += lineCount
      if (change.removed) deletions += lineCount
    }

    return { additions, deletions }
  }, [diffResult])

  const handleExport = useCallback(() => {
    const content = generatePatchContent(diffResult, baseVersion, compareVersion)
    downloadFile(content, `workflow-diff-v${baseVersion}-v${compareVersion}.patch`, 'text/plain')
  }, [diffResult, baseVersion, compareVersion])

  const handleCopyDiff = useCallback(() => {
    const content = generatePatchContent(diffResult, baseVersion, compareVersion)
    navigator.clipboard.writeText(content)
  }, [diffResult, baseVersion, compareVersion])

  return (
    <div className="flex flex-col h-full bg-gray-900 rounded-lg overflow-hidden">
      {/* Header */}
      <div className="p-4 bg-gray-800 border-b border-gray-700">
        <div className="flex items-center justify-between flex-wrap gap-3">
          <div className="flex items-center gap-4">
            <h3 className="text-lg font-semibold text-white">JSON Diff</h3>
            <div className="flex items-center gap-3 text-sm">
              <span className="flex items-center gap-1">
                <span className="w-3 h-3 rounded bg-green-500" />
                <span className="text-green-400">+{stats.additions}</span>
              </span>
              <span className="flex items-center gap-1">
                <span className="w-3 h-3 rounded bg-red-500" />
                <span className="text-red-400">-{stats.deletions}</span>
              </span>
            </div>
          </div>

          <div className="flex items-center gap-4">
            {/* View Mode Toggle */}
            <div className="flex items-center bg-gray-700 rounded-lg p-0.5">
              <button
                onClick={() => setViewMode('unified')}
                className={`px-3 py-1 text-xs rounded transition-colors ${
                  viewMode === 'unified'
                    ? 'bg-gray-600 text-white'
                    : 'text-gray-400 hover:text-white'
                }`}
              >
                Unified
              </button>
              <button
                onClick={() => setViewMode('split')}
                className={`px-3 py-1 text-xs rounded transition-colors ${
                  viewMode === 'split'
                    ? 'bg-gray-600 text-white'
                    : 'text-gray-400 hover:text-white'
                }`}
              >
                Split
              </button>
            </div>

            {/* Options */}
            <label className="flex items-center gap-2 text-sm text-gray-400">
              <input
                type="checkbox"
                checked={showLineNumbers}
                onChange={(e) => setShowLineNumbers(e.target.checked)}
                className="rounded bg-gray-700 border-gray-600 text-primary-600 focus:ring-primary-500"
              />
              Line numbers
            </label>

            {/* Actions */}
            <div className="flex gap-2">
              <button
                onClick={handleCopyDiff}
                className="px-3 py-1.5 text-xs bg-gray-700 text-gray-300 hover:bg-gray-600 hover:text-white rounded transition-colors"
                title="Copy diff to clipboard"
              >
                Copy
              </button>
              <button
                onClick={handleExport}
                className="px-3 py-1.5 text-xs bg-gray-700 text-gray-300 hover:bg-gray-600 hover:text-white rounded transition-colors"
                title="Export as patch file"
              >
                Export
              </button>
            </div>
          </div>
        </div>

        {/* Version Labels */}
        <div className="flex items-center gap-4 mt-3 text-sm">
          <div className="flex items-center gap-2">
            <span className="text-gray-400">Base:</span>
            <span className="px-2 py-0.5 bg-gray-700 text-white rounded">v{baseVersion}</span>
          </div>
          <span className="text-gray-500">→</span>
          <div className="flex items-center gap-2">
            <span className="text-gray-400">Compare:</span>
            <span className="px-2 py-0.5 bg-primary-600 text-white rounded">v{compareVersion}</span>
          </div>
        </div>
      </div>

      {/* Diff Content */}
      <div className="flex-1 overflow-auto">
        {viewMode === 'unified' ? (
          <UnifiedDiffView
            diffResult={diffResult}
            showLineNumbers={showLineNumbers}
          />
        ) : (
          <SplitDiffView
            baseJson={baseJson}
            compareJson={compareJson}
            diffResult={diffResult}
            showLineNumbers={showLineNumbers}
          />
        )}
      </div>
    </div>
  )
}

// ============================================================================
// Unified Diff View
// ============================================================================

interface UnifiedDiffViewProps {
  diffResult: Change[]
  showLineNumbers: boolean
}

function UnifiedDiffView({ diffResult, showLineNumbers }: UnifiedDiffViewProps) {
  let oldLineNum = 1
  let newLineNum = 1

  return (
    <div className="font-mono text-sm">
      {diffResult.map((change, idx) => {
        const lines = change.value.split('\n').filter((_, i, arr) => {
          // Don't show the last empty line from split
          return i < arr.length - 1 || arr[i] !== ''
        })

        return lines.map((line, lineIdx) => {
          let lineClass = 'bg-gray-900'
          let linePrefix = ' '
          let oldNum: number | null = null
          let newNum: number | null = null

          if (change.added) {
            lineClass = 'bg-green-900/30'
            linePrefix = '+'
            newNum = newLineNum++
          } else if (change.removed) {
            lineClass = 'bg-red-900/30'
            linePrefix = '-'
            oldNum = oldLineNum++
          } else {
            oldNum = oldLineNum++
            newNum = newLineNum++
          }

          return (
            <div
              key={`${idx}-${lineIdx}`}
              className={`flex hover:bg-gray-800/50 ${lineClass}`}
            >
              {showLineNumbers && (
                <div className="flex-shrink-0 w-20 flex text-gray-500 text-xs select-none border-r border-gray-700">
                  <span className="w-10 px-2 py-0.5 text-right">
                    {oldNum || ''}
                  </span>
                  <span className="w-10 px-2 py-0.5 text-right">
                    {newNum || ''}
                  </span>
                </div>
              )}
              <div className="flex-shrink-0 w-6 text-center py-0.5 select-none">
                <span className={change.added ? 'text-green-400' : change.removed ? 'text-red-400' : 'text-gray-500'}>
                  {linePrefix}
                </span>
              </div>
              <div className="flex-1 py-0.5 px-2 whitespace-pre overflow-x-auto">
                <SyntaxHighlightedLine line={line} />
              </div>
            </div>
          )
        })
      })}
    </div>
  )
}

// ============================================================================
// Split Diff View
// ============================================================================

interface SplitDiffViewProps {
  baseJson: string
  compareJson: string
  diffResult: Change[]
  showLineNumbers: boolean
}

function SplitDiffView({ diffResult, showLineNumbers }: SplitDiffViewProps) {
  // Build aligned lines for split view
  const alignedLines = useMemo(() => {
    const result: Array<{
      leftLine: string | null
      rightLine: string | null
      leftNum: number | null
      rightNum: number | null
      leftType: 'removed' | 'unchanged' | 'empty'
      rightType: 'added' | 'unchanged' | 'empty'
    }> = []

    let leftNum = 1
    let rightNum = 1

    for (const change of diffResult) {
      const lines = change.value.split('\n').filter((_, i, arr) => {
        return i < arr.length - 1 || arr[i] !== ''
      })

      if (change.added) {
        for (const line of lines) {
          result.push({
            leftLine: null,
            rightLine: line,
            leftNum: null,
            rightNum: rightNum++,
            leftType: 'empty',
            rightType: 'added',
          })
        }
      } else if (change.removed) {
        for (const line of lines) {
          result.push({
            leftLine: line,
            rightLine: null,
            leftNum: leftNum++,
            rightNum: null,
            leftType: 'removed',
            rightType: 'empty',
          })
        }
      } else {
        for (const line of lines) {
          result.push({
            leftLine: line,
            rightLine: line,
            leftNum: leftNum++,
            rightNum: rightNum++,
            leftType: 'unchanged',
            rightType: 'unchanged',
          })
        }
      }
    }

    return result
  }, [diffResult])

  return (
    <div className="flex font-mono text-sm">
      {/* Left Side */}
      <div className="flex-1 border-r border-gray-700">
        {alignedLines.map((line, idx) => (
          <div
            key={`left-${idx}`}
            className={`flex hover:bg-gray-800/50 ${
              line.leftType === 'removed'
                ? 'bg-red-900/30'
                : line.leftType === 'empty'
                ? 'bg-gray-800/30'
                : 'bg-gray-900'
            }`}
          >
            {showLineNumbers && (
              <div className="flex-shrink-0 w-10 px-2 py-0.5 text-right text-gray-500 text-xs select-none border-r border-gray-700">
                {line.leftNum || ''}
              </div>
            )}
            <div className="flex-1 py-0.5 px-2 whitespace-pre overflow-x-auto">
              {line.leftLine !== null ? (
                <SyntaxHighlightedLine line={line.leftLine} />
              ) : (
                <span className="text-gray-600 italic">-</span>
              )}
            </div>
          </div>
        ))}
      </div>

      {/* Right Side */}
      <div className="flex-1">
        {alignedLines.map((line, idx) => (
          <div
            key={`right-${idx}`}
            className={`flex hover:bg-gray-800/50 ${
              line.rightType === 'added'
                ? 'bg-green-900/30'
                : line.rightType === 'empty'
                ? 'bg-gray-800/30'
                : 'bg-gray-900'
            }`}
          >
            {showLineNumbers && (
              <div className="flex-shrink-0 w-10 px-2 py-0.5 text-right text-gray-500 text-xs select-none border-r border-gray-700">
                {line.rightNum || ''}
              </div>
            )}
            <div className="flex-1 py-0.5 px-2 whitespace-pre overflow-x-auto">
              {line.rightLine !== null ? (
                <SyntaxHighlightedLine line={line.rightLine} />
              ) : (
                <span className="text-gray-600 italic">-</span>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

// ============================================================================
// Syntax Highlighting Component
// ============================================================================

interface SyntaxHighlightedLineProps {
  line: string
}

function SyntaxHighlightedLine({ line }: SyntaxHighlightedLineProps) {
  // Simple JSON syntax highlighting
  const highlighted = useMemo(() => {
    const parts: Array<{ text: string; className: string }> = []
    let remaining = line

    // Match patterns in order
    const patterns = [
      { regex: /^(\s*)"([^"]+)"(?=\s*:)/, className: 'text-blue-400' }, // Keys
      { regex: /^(\s*)"([^"]*)"/, className: 'text-green-400' }, // String values
      { regex: /^(\s*)(true|false)/, className: 'text-yellow-400' }, // Booleans
      { regex: /^(\s*)(null)/, className: 'text-gray-500' }, // Null
      { regex: /^(\s*)(-?\d+\.?\d*)/, className: 'text-orange-400' }, // Numbers
      { regex: /^(\s*)([{}[\],:])/,className: 'text-gray-400' }, // Punctuation
    ]

    while (remaining.length > 0) {
      let matched = false

      for (const { regex, className } of patterns) {
        const match = remaining.match(regex)
        if (match) {
          if (match[1]) {
            parts.push({ text: match[1], className: '' }) // Leading whitespace
          }
          parts.push({ text: match[2] || match[0].slice(match[1]?.length || 0), className })
          remaining = remaining.slice(match[0].length)
          matched = true
          break
        }
      }

      if (!matched) {
        // Add single character and continue
        parts.push({ text: remaining[0], className: 'text-gray-300' })
        remaining = remaining.slice(1)
      }
    }

    return parts
  }, [line])

  return (
    <>
      {highlighted.map((part, idx) => (
        <span key={idx} className={part.className}>
          {part.text}
        </span>
      ))}
    </>
  )
}

// ============================================================================
// Utility Functions
// ============================================================================

function generatePatchContent(diffResult: Change[], baseVersion: number, compareVersion: number): string {
  const lines: string[] = [
    `--- workflow v${baseVersion}`,
    `+++ workflow v${compareVersion}`,
    '',
  ]

  for (const change of diffResult) {
    const changeLines = change.value.split('\n').filter((_, i, arr) => {
      return i < arr.length - 1 || arr[i] !== ''
    })

    for (const line of changeLines) {
      if (change.added) {
        lines.push(`+ ${line}`)
      } else if (change.removed) {
        lines.push(`- ${line}`)
      } else {
        lines.push(`  ${line}`)
      }
    }
  }

  return lines.join('\n')
}

function downloadFile(content: string, filename: string, mimeType: string) {
  const blob = new Blob([content], { type: mimeType })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}
