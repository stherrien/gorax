import React from 'react'
import type { AIModel } from '../../types/ai'

export interface TokenCounterProps {
  tokens: number
  outputTokens?: number
  model?: AIModel
  loading?: boolean
  label?: string
  showContextUsage?: boolean
  showBreakdown?: boolean
  size?: 'small' | 'medium' | 'large'
  className?: string
}

const formatTokens = (tokens: number, compact = false): string => {
  if (tokens >= 1000000) {
    return `${(tokens / 1000000).toFixed(1)}M${compact ? '' : ' tokens'}`
  }
  if (tokens >= 1000) {
    return `${(tokens / 1000).toFixed(1)}K${compact ? '' : ' tokens'}`
  }
  if (compact) {
    return tokens.toString()
  }
  return `${tokens} ${tokens === 1 ? 'token' : 'tokens'}`
}

const estimateCost = (
  model: AIModel,
  inputTokens: number,
  outputTokens: number = 0
): number => {
  // Costs are in cents per 1M tokens
  const inputCostCents = (inputTokens / 1000000) * model.inputCostPer1M
  const outputCostCents = (outputTokens / 1000000) * model.outputCostPer1M
  // Convert to dollars
  return (inputCostCents + outputCostCents) / 100
}

const formatCost = (dollars: number): string => {
  if (dollars < 0.01) {
    return '<$0.01'
  }
  return `$${dollars.toFixed(2)}`
}

export const TokenCounter: React.FC<TokenCounterProps> = ({
  tokens,
  outputTokens = 0,
  model,
  loading = false,
  label,
  showContextUsage = false,
  showBreakdown = false,
  size = 'medium',
  className = '',
}) => {
  const contextUsagePercent = model
    ? Math.round((tokens / model.contextWindow) * 100)
    : 0

  const getContextUsageClass = () => {
    if (contextUsagePercent > 100) return 'text-red-600'
    if (contextUsagePercent >= 90) return 'text-yellow-600'
    return 'text-gray-600'
  }

  const cost = model ? estimateCost(model, tokens, outputTokens) : 0

  const sizeClasses = {
    small: 'text-xs',
    medium: 'text-sm',
    large: 'text-base',
  }

  if (loading) {
    return (
      <div className={`${sizeClasses[size]} text-gray-500 ${className}`}>
        Counting...
      </div>
    )
  }

  return (
    <div className={`${sizeClasses[size]} ${className}`}>
      {label && <span className="text-gray-600">{label}: </span>}

      {showBreakdown ? (
        <div className="flex flex-col gap-1">
          <div className="flex items-center gap-2">
            <span className="text-gray-700">
              {formatTokens(tokens, size === 'small')} input
            </span>
            {outputTokens > 0 && (
              <span className="text-gray-700">
                {formatTokens(outputTokens, size === 'small')} output
              </span>
            )}
          </div>
          {model && (
            <span className="text-gray-500">
              Est. {formatCost(cost)}
            </span>
          )}
        </div>
      ) : (
        <div className="flex items-center gap-2">
          <span className="text-gray-700">
            {formatTokens(tokens, size === 'small')}
          </span>

          {showContextUsage && model && (
            <span className={getContextUsageClass()}>
              {contextUsagePercent}%
            </span>
          )}

          {model && (
            <span className="text-gray-500">
              {formatCost(cost)}
            </span>
          )}
        </div>
      )}
    </div>
  )
}
