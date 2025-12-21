import React from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts';
import { ExecutionTrend } from '../../api/metrics';

interface SuccessRateChartProps {
  trends: ExecutionTrend[];
  loading?: boolean;
  error?: string;
  targetRate?: number; // Target success rate (default 95%)
}

interface SuccessRateDataPoint {
  date: string;
  successRate: number;
  count: number;
}

const SuccessRateChart: React.FC<SuccessRateChartProps> = ({
  trends,
  loading = false,
  error,
  targetRate = 95,
}) => {
  const calculateSuccessRate = (trend: ExecutionTrend): number => {
    if (trend.count === 0) return 0;
    return (trend.success / trend.count) * 100;
  };

  const data: SuccessRateDataPoint[] = trends.map(trend => ({
    date: trend.date,
    successRate: calculateSuccessRate(trend),
    count: trend.count,
  }));

  const formatXAxis = (value: string) => {
    const date = new Date(value);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  };

  const formatTooltip = (value: number | undefined, _name: string | undefined) => {
    if (value === undefined) return '';
    return `${value.toFixed(2)}%`;
  };

  // Calculate overall success rate
  const totalSuccess = trends.reduce((sum, t) => sum + t.success, 0);
  const totalCount = trends.reduce((sum, t) => sum + t.count, 0);
  const overallRate = totalCount > 0 ? (totalSuccess / totalCount) * 100 : 0;

  // Determine trend
  const getTrend = () => {
    if (data.length < 2) return 'stable';
    const recentRates = data.slice(-7).map(d => d.successRate);
    const firstHalf = recentRates.slice(0, Math.floor(recentRates.length / 2));
    const secondHalf = recentRates.slice(Math.floor(recentRates.length / 2));
    const firstAvg = firstHalf.reduce((sum, r) => sum + r, 0) / firstHalf.length;
    const secondAvg = secondHalf.reduce((sum, r) => sum + r, 0) / secondHalf.length;

    if (secondAvg > firstAvg + 2) return 'improving';
    if (secondAvg < firstAvg - 2) return 'declining';
    return 'stable';
  };

  const trend = getTrend();
  const getStatusColor = (rate: number) => {
    if (rate >= 90) return 'text-green-600';
    if (rate >= 80) return 'text-yellow-600';
    return 'text-red-600';
  };

  const getTrendIcon = () => {
    if (trend === 'improving') return '↗';
    if (trend === 'declining') return '↘';
    return '→';
  };

  if (error) {
    return (
      <div className="bg-white p-6 rounded-lg shadow">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Success Rate</h3>
        <div className="flex items-center justify-center h-64 text-red-600">
          {error}
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white p-6 rounded-lg shadow">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-gray-900">Success Rate</h3>
        <div className="flex items-center gap-4">
          <div className="text-right">
            <div className={`text-2xl font-bold ${getStatusColor(overallRate)}`}>
              {overallRate.toFixed(1)}%
            </div>
            <div className="text-xs text-gray-500 flex items-center gap-1">
              <span>{getTrendIcon()}</span>
              <span className="capitalize">{trend}</span>
            </div>
          </div>
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      ) : data.length === 0 ? (
        <div className="flex items-center justify-center h-64 text-gray-500">
          No data available
        </div>
      ) : (
        <>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={data}>
              <defs>
                <linearGradient id="successRateGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#10b981" stopOpacity={0.3} />
                  <stop offset="100%" stopColor="#10b981" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis
                dataKey="date"
                tickFormatter={formatXAxis}
                stroke="#6b7280"
              />
              <YAxis
                domain={[0, 100]}
                stroke="#6b7280"
                tickFormatter={(value) => `${value}%`}
              />
              <Tooltip
                formatter={formatTooltip}
                contentStyle={{
                  backgroundColor: '#fff',
                  border: '1px solid #e5e7eb',
                  borderRadius: '0.5rem'
                }}
              />
              <Legend />
              {/* Target line */}
              <ReferenceLine
                y={targetRate}
                stroke="#f59e0b"
                strokeDasharray="3 3"
                label={{
                  value: `Target ${targetRate}%`,
                  position: 'right',
                  fill: '#f59e0b',
                  fontSize: 12,
                }}
              />
              {/* Color zones */}
              <ReferenceLine y={90} stroke="#10b981" strokeOpacity={0.2} />
              <ReferenceLine y={80} stroke="#f59e0b" strokeOpacity={0.2} />
              <Line
                type="monotone"
                dataKey="successRate"
                name="Success Rate"
                stroke="#10b981"
                strokeWidth={3}
                dot={{ fill: '#10b981', r: 5 }}
                activeDot={{ r: 7 }}
                fill="url(#successRateGradient)"
              />
            </LineChart>
          </ResponsiveContainer>
          <div className="mt-4 grid grid-cols-3 gap-4 text-center text-sm">
            <div className="p-3 bg-green-50 rounded-lg">
              <div className="text-green-700 font-semibold">Excellent</div>
              <div className="text-gray-600">&gt; 90%</div>
            </div>
            <div className="p-3 bg-yellow-50 rounded-lg">
              <div className="text-yellow-700 font-semibold">Good</div>
              <div className="text-gray-600">80-90%</div>
            </div>
            <div className="p-3 bg-red-50 rounded-lg">
              <div className="text-red-700 font-semibold">Needs Attention</div>
              <div className="text-gray-600">&lt; 80%</div>
            </div>
          </div>
        </>
      )}
    </div>
  );
};

export default SuccessRateChart;
