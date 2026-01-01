import { useState, useMemo } from 'react'
import {
  useTenantOverview,
  useExecutionTrends,
  useTopWorkflows,
  useErrorBreakdown,
} from '../hooks/useAnalytics'
import type { Granularity } from '../types/analytics'

export default function Analytics() {
  const [dateRange, setDateRange] = useState({
    startDate: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
    endDate: new Date().toISOString(),
  })
  const [granularity, setGranularity] = useState<Granularity>('day')

  const { data: overview, loading: overviewLoading } = useTenantOverview(dateRange)
  const { data: trends, loading: trendsLoading } = useExecutionTrends({
    ...dateRange,
    granularity,
  })
  const { data: topWorkflows, loading: topWorkflowsLoading } = useTopWorkflows({
    ...dateRange,
    limit: 10,
  })
  const { data: errorBreakdown, loading: errorBreakdownLoading } =
    useErrorBreakdown(dateRange)

  const successRatePercentage = useMemo(() => {
    if (!overview) return 0
    return Math.round(overview.successRate * 100)
  }, [overview])

  const avgDurationSeconds = useMemo(() => {
    if (!overview) return 0
    return (overview.avgDurationMs / 1000).toFixed(2)
  }, [overview])

  const handleDateRangeChange = (range: 'day' | 'week' | 'month' | 'year') => {
    const now = new Date()
    const end = now.toISOString()
    let start: Date

    switch (range) {
      case 'day':
        start = new Date(now.getTime() - 24 * 60 * 60 * 1000)
        break
      case 'week':
        start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
        break
      case 'month':
        start = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
        break
      case 'year':
        start = new Date(now.getTime() - 365 * 24 * 60 * 60 * 1000)
        break
    }

    setDateRange({
      startDate: start.toISOString(),
      endDate: end,
    })
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Analytics Dashboard</h1>
        <p className="mt-2 text-gray-600">
          Monitor workflow performance and execution trends
        </p>
      </div>

      <div className="mb-6 flex items-center justify-between">
        <div className="flex gap-2">
          <button
            onClick={() => handleDateRangeChange('day')}
            className="px-4 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
          >
            Last 24h
          </button>
          <button
            onClick={() => handleDateRangeChange('week')}
            className="px-4 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
          >
            Last Week
          </button>
          <button
            onClick={() => handleDateRangeChange('month')}
            className="px-4 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
          >
            Last Month
          </button>
          <button
            onClick={() => handleDateRangeChange('year')}
            className="px-4 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
          >
            Last Year
          </button>
        </div>

        <select
          value={granularity}
          onChange={(e) => setGranularity(e.target.value as Granularity)}
          className="px-4 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
        >
          <option value="hour">Hourly</option>
          <option value="day">Daily</option>
          <option value="week">Weekly</option>
          <option value="month">Monthly</option>
        </select>
      </div>

      {overviewLoading ? (
        <div className="text-center py-12">Loading overview...</div>
      ) : overview ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 uppercase">
              Total Executions
            </h3>
            <p className="mt-2 text-3xl font-bold text-gray-900">
              {overview.totalExecutions.toLocaleString()}
            </p>
            <p className="mt-1 text-sm text-green-600">
              {overview.successfulExecutions} successful
            </p>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 uppercase">
              Success Rate
            </h3>
            <p className="mt-2 text-3xl font-bold text-gray-900">
              {successRatePercentage}%
            </p>
            <p className="mt-1 text-sm text-red-600">
              {overview.failedExecutions} failed
            </p>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 uppercase">
              Avg Duration
            </h3>
            <p className="mt-2 text-3xl font-bold text-gray-900">
              {avgDurationSeconds}s
            </p>
            <p className="mt-1 text-sm text-gray-600">Per execution</p>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-sm font-medium text-gray-500 uppercase">
              Active Workflows
            </h3>
            <p className="mt-2 text-3xl font-bold text-gray-900">
              {overview.activeWorkflows}
            </p>
            <p className="mt-1 text-sm text-gray-600">
              of {overview.totalWorkflows} total
            </p>
          </div>
        </div>
      ) : null}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-bold text-gray-900 mb-4">
            Execution Trends
          </h2>
          {trendsLoading ? (
            <div className="text-center py-8">Loading trends...</div>
          ) : trends && trends.dataPoints.length > 0 ? (
            <div className="space-y-2">
              {trends.dataPoints.map((point, index) => (
                <div
                  key={index}
                  className="flex items-center justify-between p-2 hover:bg-gray-50 rounded"
                >
                  <span className="text-sm text-gray-600">
                    {new Date(point.timestamp).toLocaleDateString()}
                  </span>
                  <div className="flex items-center gap-4">
                    <span className="text-sm font-medium text-gray-900">
                      {point.executionCount} executions
                    </span>
                    <span
                      className={`text-sm ${
                        point.successRate >= 0.9
                          ? 'text-green-600'
                          : point.successRate >= 0.7
                          ? 'text-yellow-600'
                          : 'text-red-600'
                      }`}
                    >
                      {Math.round(point.successRate * 100)}% success
                    </span>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">No data available</div>
          )}
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-bold text-gray-900 mb-4">Top Workflows</h2>
          {topWorkflowsLoading ? (
            <div className="text-center py-8">Loading workflows...</div>
          ) : topWorkflows && topWorkflows.workflows.length > 0 ? (
            <div className="space-y-3">
              {topWorkflows.workflows.map((workflow) => (
                <div
                  key={workflow.workflowId}
                  className="flex items-center justify-between p-3 border border-gray-200 rounded-lg hover:border-indigo-300 transition-colors"
                >
                  <div className="flex-1">
                    <h3 className="font-medium text-gray-900">
                      {workflow.workflowName}
                    </h3>
                    <p className="text-sm text-gray-500">
                      {workflow.executionCount} executions •{' '}
                      {Math.round(workflow.successRate * 100)}% success rate
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-medium text-gray-900">
                      {(workflow.avgDurationMs / 1000).toFixed(2)}s
                    </p>
                    <p className="text-xs text-gray-500">avg duration</p>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">No workflows found</div>
          )}
        </div>
      </div>

      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-xl font-bold text-gray-900 mb-4">Error Breakdown</h2>
        {errorBreakdownLoading ? (
          <div className="text-center py-8">Loading errors...</div>
        ) : errorBreakdown && errorBreakdown.errorsByType.length > 0 ? (
          <div className="space-y-3">
            {errorBreakdown.errorsByType.map((error, index) => (
              <div
                key={index}
                className="p-4 border border-red-200 rounded-lg bg-red-50"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="font-medium text-gray-900">
                      {error.errorMessage}
                    </h3>
                    <p className="text-sm text-gray-600 mt-1">
                      {error.workflowName}
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="text-lg font-bold text-red-600">
                      {error.errorCount}
                    </p>
                    <p className="text-xs text-gray-500">
                      {error.percentage.toFixed(1)}%
                    </p>
                  </div>
                </div>
                <p className="text-xs text-gray-500 mt-2">
                  Last occurred:{' '}
                  {new Date(error.lastOccurrence).toLocaleString()}
                </p>
              </div>
            ))}
          </div>
        ) : (
          <div className="text-center py-8 text-gray-500">
            No errors found - great job!
          </div>
        )}
      </div>
    </div>
  )
}
