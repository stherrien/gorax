import React, { useState } from 'react';
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { ExecutionTrend } from '../../api/metrics';

interface ExecutionTrendChartProps {
  trends: ExecutionTrend[];
  loading?: boolean;
  error?: string;
  onRefresh?: () => void;
}

type ChartType = 'line' | 'bar' | 'stacked';
type TimeRange = '7d' | '30d' | '90d';
type GroupBy = 'hour' | 'day';

const ExecutionTrendChart: React.FC<ExecutionTrendChartProps> = ({
  trends,
  loading = false,
  error,
  onRefresh,
}) => {
  const [chartType, setChartType] = useState<ChartType>('line');
  const [timeRange, setTimeRange] = useState<TimeRange>('7d');
  const [groupBy, setGroupBy] = useState<GroupBy>('day');

  const formatXAxis = (value: string) => {
    if (groupBy === 'hour') {
      // Format: "2024-01-15 14:00" -> "14:00"
      const parts = value.split(' ');
      return parts[1] || value;
    }
    // Format: "2024-01-15" -> "Jan 15"
    const date = new Date(value);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  };

  const renderChart = () => {
    if (chartType === 'line') {
      return (
        <LineChart data={trends}>
          <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
          <XAxis
            dataKey="date"
            tickFormatter={formatXAxis}
            stroke="#6b7280"
          />
          <YAxis stroke="#6b7280" />
          <Tooltip
            contentStyle={{
              backgroundColor: '#fff',
              border: '1px solid #e5e7eb',
              borderRadius: '0.5rem'
            }}
          />
          <Legend />
          <Line
            type="monotone"
            dataKey="success"
            name="Success"
            stroke="#10b981"
            strokeWidth={2}
            dot={{ fill: '#10b981', r: 4 }}
          />
          <Line
            type="monotone"
            dataKey="failed"
            name="Failed"
            stroke="#ef4444"
            strokeWidth={2}
            dot={{ fill: '#ef4444', r: 4 }}
          />
          <Line
            type="monotone"
            dataKey="count"
            name="Total"
            stroke="#6366f1"
            strokeWidth={2}
            strokeDasharray="5 5"
            dot={{ fill: '#6366f1', r: 4 }}
          />
        </LineChart>
      );
    }

    if (chartType === 'stacked') {
      return (
        <BarChart data={trends}>
          <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
          <XAxis
            dataKey="date"
            tickFormatter={formatXAxis}
            stroke="#6b7280"
          />
          <YAxis stroke="#6b7280" />
          <Tooltip
            contentStyle={{
              backgroundColor: '#fff',
              border: '1px solid #e5e7eb',
              borderRadius: '0.5rem'
            }}
          />
          <Legend />
          <Bar dataKey="success" stackId="a" fill="#10b981" name="Success" />
          <Bar dataKey="failed" stackId="a" fill="#ef4444" name="Failed" />
        </BarChart>
      );
    }

    // Grouped bar chart
    return (
      <BarChart data={trends}>
        <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
        <XAxis
          dataKey="date"
          tickFormatter={formatXAxis}
          stroke="#6b7280"
        />
        <YAxis stroke="#6b7280" />
        <Tooltip
          contentStyle={{
            backgroundColor: '#fff',
            border: '1px solid #e5e7eb',
            borderRadius: '0.5rem'
          }}
        />
        <Legend />
        <Bar dataKey="success" fill="#10b981" name="Success" />
        <Bar dataKey="failed" fill="#ef4444" name="Failed" />
      </BarChart>
    );
  };

  if (error) {
    return (
      <div className="bg-white p-6 rounded-lg shadow">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900">Execution Trends</h3>
        </div>
        <div className="flex items-center justify-center h-64 text-red-600">
          <div className="text-center">
            <p className="mb-2">{error}</p>
            {onRefresh && (
              <button
                onClick={onRefresh}
                className="text-sm text-blue-600 hover:text-blue-800"
              >
                Try Again
              </button>
            )}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white p-6 rounded-lg shadow">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-gray-900">Execution Trends</h3>
        <div className="flex gap-2">
          {/* Chart Type Toggle */}
          <div className="flex bg-gray-100 rounded-md p-1">
            <button
              onClick={() => setChartType('line')}
              className={`px-3 py-1 rounded text-sm ${
                chartType === 'line'
                  ? 'bg-white shadow text-gray-900'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              Line
            </button>
            <button
              onClick={() => setChartType('bar')}
              className={`px-3 py-1 rounded text-sm ${
                chartType === 'bar'
                  ? 'bg-white shadow text-gray-900'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              Bar
            </button>
            <button
              onClick={() => setChartType('stacked')}
              className={`px-3 py-1 rounded text-sm ${
                chartType === 'stacked'
                  ? 'bg-white shadow text-gray-900'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              Stacked
            </button>
          </div>

          {/* Time Range Selector */}
          <select
            value={timeRange}
            onChange={(e) => setTimeRange(e.target.value as TimeRange)}
            className="px-3 py-1 border border-gray-300 rounded-md text-sm"
          >
            <option value="7d">Last 7 days</option>
            <option value="30d">Last 30 days</option>
            <option value="90d">Last 90 days</option>
          </select>

          {/* Group By Selector */}
          <select
            value={groupBy}
            onChange={(e) => setGroupBy(e.target.value as GroupBy)}
            className="px-3 py-1 border border-gray-300 rounded-md text-sm"
          >
            <option value="hour">Hourly</option>
            <option value="day">Daily</option>
          </select>
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      ) : trends.length === 0 ? (
        <div className="flex items-center justify-center h-64 text-gray-500">
          No execution data available
        </div>
      ) : (
        <ResponsiveContainer width="100%" height={300}>
          {renderChart()}
        </ResponsiveContainer>
      )}
    </div>
  );
};

export default ExecutionTrendChart;
