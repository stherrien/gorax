import { useState } from 'react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from 'recharts';
import { DurationStats } from '../../api/metrics';

interface DurationChartProps {
  stats: DurationStats[];
  loading?: boolean;
  error?: string;
}

type MetricType = 'avg' | 'p50' | 'p90' | 'p99';
type SortBy = 'duration' | 'runs';

const DurationChart: React.FC<DurationChartProps> = ({
  stats,
  loading = false,
  error,
}) => {
  const [metricType, setMetricType] = useState<MetricType>('avg');
  const [sortBy, setSortBy] = useState<SortBy>('duration');

  const formatDuration = (ms: number): string => {
    if (ms < 1000) return `${ms.toFixed(0)}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(2)}s`;
    return `${(ms / 60000).toFixed(2)}m`;
  };

  const getMetricValue = (stat: DurationStats): number => {
    switch (metricType) {
      case 'p50':
        return stat.p50Duration;
      case 'p90':
        return stat.p90Duration;
      case 'p99':
        return stat.p99Duration;
      default:
        return stat.avgDuration;
    }
  };

  const getMetricLabel = (): string => {
    switch (metricType) {
      case 'p50':
        return 'Median (P50)';
      case 'p90':
        return 'P90';
      case 'p99':
        return 'P99';
      default:
        return 'Average';
    }
  };

  const sortedStats = [...stats].sort((a, b) => {
    if (sortBy === 'runs') {
      return b.totalRuns - a.totalRuns;
    }
    return getMetricValue(b) - getMetricValue(a);
  });

  // Highlight outliers (top 20% in duration)
  const maxDuration = Math.max(...sortedStats.map(getMetricValue));
  const outlierThreshold = maxDuration * 0.8;

  const getBarColor = (value: number): string => {
    if (value >= outlierThreshold) return '#ef4444'; // Red for outliers
    if (value >= maxDuration * 0.5) return '#f59e0b'; // Orange for medium
    return '#10b981'; // Green for fast
  };

  const data = sortedStats.slice(0, 10).map(stat => ({
    name: stat.workflowName.length > 20
      ? stat.workflowName.substring(0, 20) + '...'
      : stat.workflowName,
    fullName: stat.workflowName,
    value: getMetricValue(stat),
    runs: stat.totalRuns,
    avg: stat.avgDuration,
    p50: stat.p50Duration,
    p90: stat.p90Duration,
    p99: stat.p99Duration,
  }));

  if (error) {
    return (
      <div className="bg-white p-6 rounded-lg shadow">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Execution Duration</h3>
        <div className="flex items-center justify-center h-64 text-red-600">
          {error}
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white p-6 rounded-lg shadow">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-gray-900">Execution Duration</h3>
        <div className="flex gap-2">
          {/* Metric Type Selector */}
          <select
            value={metricType}
            onChange={(e) => setMetricType(e.target.value as MetricType)}
            className="px-3 py-1 border border-gray-300 rounded-md text-sm"
          >
            <option value="avg">Average</option>
            <option value="p50">Median (P50)</option>
            <option value="p90">P90</option>
            <option value="p99">P99</option>
          </select>

          {/* Sort By Selector */}
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as SortBy)}
            className="px-3 py-1 border border-gray-300 rounded-md text-sm"
          >
            <option value="duration">Sort by Duration</option>
            <option value="runs">Sort by Run Count</option>
          </select>
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      ) : data.length === 0 ? (
        <div className="flex items-center justify-center h-64 text-gray-500">
          No duration data available
        </div>
      ) : (
        <>
          <ResponsiveContainer width="100%" height={350}>
            <BarChart
              data={data}
              layout="vertical"
              margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis
                type="number"
                stroke="#6b7280"
                tickFormatter={formatDuration}
              />
              <YAxis
                dataKey="name"
                type="category"
                stroke="#6b7280"
                width={150}
              />
              <Tooltip
                content={({ active, payload }) => {
                  if (active && payload && payload.length) {
                    const data = payload[0].payload;
                    return (
                      <div className="bg-white p-3 border border-gray-200 rounded-lg shadow-lg">
                        <p className="font-semibold text-gray-900 mb-2">
                          {data.fullName}
                        </p>
                        <div className="space-y-1 text-sm">
                          <p className="text-gray-600">
                            Avg: <span className="font-medium">{formatDuration(data.avg)}</span>
                          </p>
                          <p className="text-gray-600">
                            P50: <span className="font-medium">{formatDuration(data.p50)}</span>
                          </p>
                          <p className="text-gray-600">
                            P90: <span className="font-medium">{formatDuration(data.p90)}</span>
                          </p>
                          <p className="text-gray-600">
                            P99: <span className="font-medium">{formatDuration(data.p99)}</span>
                          </p>
                          <p className="text-gray-600 pt-1 border-t border-gray-200">
                            Total Runs: <span className="font-medium">{data.runs}</span>
                          </p>
                        </div>
                      </div>
                    );
                  }
                  return null;
                }}
              />
              <Bar dataKey="value" name={getMetricLabel()} radius={[0, 4, 4, 0]}>
                {data.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={getBarColor(entry.value)} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>

          {/* Legend */}
          <div className="mt-4 flex justify-center gap-6 text-sm">
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-green-500 rounded"></div>
              <span className="text-gray-600">Fast</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-orange-500 rounded"></div>
              <span className="text-gray-600">Medium</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-red-500 rounded"></div>
              <span className="text-gray-600">Slow (Outlier)</span>
            </div>
          </div>

          {/* Info note */}
          <div className="mt-4 p-3 bg-blue-50 rounded-lg text-sm text-blue-800">
            <p className="font-medium">About Percentiles:</p>
            <ul className="mt-1 ml-4 list-disc space-y-1">
              <li>P50 (Median): 50% of executions complete faster</li>
              <li>P90: 90% of executions complete faster</li>
              <li>P99: 99% of executions complete faster (captures worst-case scenarios)</li>
            </ul>
          </div>
        </>
      )}
    </div>
  );
};

export default DurationChart;
