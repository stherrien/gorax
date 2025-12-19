import React, { useState } from 'react';
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';
import { TriggerTypeBreakdown } from '../../api/metrics';

interface TriggerBreakdownProps {
  breakdown: TriggerTypeBreakdown[];
  loading?: boolean;
  error?: string;
  onFilterClick?: (triggerType: string) => void;
}

const COLORS = [
  '#6366f1', // Indigo
  '#10b981', // Green
  '#f59e0b', // Amber
  '#ef4444', // Red
  '#8b5cf6', // Purple
  '#ec4899', // Pink
  '#14b8a6', // Teal
  '#f97316', // Orange
];

const TriggerBreakdown: React.FC<TriggerBreakdownProps> = ({
  breakdown,
  loading = false,
  error,
  onFilterClick,
}) => {
  const [activeIndex, setActiveIndex] = useState<number | null>(null);

  const formatTriggerType = (type: string): string => {
    return type
      .split('_')
      .map(word => word.charAt(0).toUpperCase() + word.slice(1))
      .join(' ');
  };

  const getTriggerIcon = (type: string): string => {
    switch (type.toLowerCase()) {
      case 'webhook':
        return '🔗';
      case 'schedule':
        return '⏰';
      case 'manual':
        return '👆';
      case 'api':
        return '🔌';
      default:
        return '📋';
    }
  };

  const data = breakdown.map(item => ({
    name: formatTriggerType(item.triggerType),
    value: item.count,
    percentage: item.percentage,
    triggerType: item.triggerType,
  }));

  const renderCustomLabel = ({
    cx,
    cy,
    midAngle,
    innerRadius,
    outerRadius,
    percent,
  }: any) => {
    const RADIAN = Math.PI / 180;
    const radius = innerRadius + (outerRadius - innerRadius) * 0.5;
    const x = cx + radius * Math.cos(-midAngle * RADIAN);
    const y = cy + radius * Math.sin(-midAngle * RADIAN);

    if (percent < 0.05) return null; // Don't show label for small slices

    return (
      <text
        x={x}
        y={y}
        fill="white"
        textAnchor={x > cx ? 'start' : 'end'}
        dominantBaseline="central"
        className="font-semibold"
      >
        {`${(percent * 100).toFixed(0)}%`}
      </text>
    );
  };

  const onPieEnter = (_: any, index: number) => {
    setActiveIndex(index);
  };

  const onPieLeave = () => {
    setActiveIndex(null);
  };

  if (error) {
    return (
      <div className="bg-white p-6 rounded-lg shadow">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Trigger Types</h3>
        <div className="flex items-center justify-center h-64 text-red-600">
          {error}
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white p-6 rounded-lg shadow">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-gray-900">Execution by Trigger Type</h3>
        <span className="text-sm text-gray-500">
          Distribution of execution triggers
        </span>
      </div>

      {loading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      ) : data.length === 0 ? (
        <div className="flex items-center justify-center h-64 text-gray-500">
          No trigger data available
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Pie Chart */}
          <div>
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={data}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={renderCustomLabel}
                  outerRadius={100}
                  innerRadius={50}
                  fill="#8884d8"
                  dataKey="value"
                  onMouseEnter={onPieEnter}
                  onMouseLeave={onPieLeave}
                >
                  {data.map((entry, index) => (
                    <Cell
                      key={`cell-${index}`}
                      fill={COLORS[index % COLORS.length]}
                      opacity={activeIndex === null || activeIndex === index ? 1 : 0.6}
                      style={{ cursor: onFilterClick ? 'pointer' : 'default' }}
                      onClick={() => onFilterClick?.(entry.triggerType)}
                    />
                  ))}
                </Pie>
                <Tooltip
                  content={({ active, payload }) => {
                    if (active && payload && payload.length) {
                      const data = payload[0].payload;
                      return (
                        <div className="bg-white p-3 border border-gray-200 rounded-lg shadow-lg">
                          <p className="font-semibold text-gray-900">
                            {getTriggerIcon(data.triggerType)} {data.name}
                          </p>
                          <p className="text-sm text-gray-600 mt-1">
                            Count: <span className="font-medium">{data.value}</span>
                          </p>
                          <p className="text-sm text-gray-600">
                            Percentage: <span className="font-medium">{data.percentage.toFixed(2)}%</span>
                          </p>
                        </div>
                      );
                    }
                    return null;
                  }}
                />
              </PieChart>
            </ResponsiveContainer>
          </div>

          {/* Legend with counts */}
          <div className="flex flex-col justify-center">
            <div className="space-y-3">
              {data.map((entry, index) => (
                <div
                  key={entry.triggerType}
                  className={`flex items-center justify-between p-3 rounded-lg transition-all ${
                    activeIndex === index ? 'bg-gray-100 scale-105' : 'bg-gray-50'
                  } ${onFilterClick ? 'cursor-pointer hover:bg-gray-100' : ''}`}
                  onMouseEnter={() => setActiveIndex(index)}
                  onMouseLeave={() => setActiveIndex(null)}
                  onClick={() => onFilterClick?.(entry.triggerType)}
                >
                  <div className="flex items-center gap-3">
                    <div
                      className="w-4 h-4 rounded-full"
                      style={{ backgroundColor: COLORS[index % COLORS.length] }}
                    />
                    <div>
                      <div className="font-medium text-gray-900 flex items-center gap-2">
                        <span>{getTriggerIcon(entry.triggerType)}</span>
                        <span>{entry.name}</span>
                      </div>
                      <div className="text-xs text-gray-500">
                        {entry.percentage.toFixed(1)}% of total
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-lg font-bold text-gray-900">
                      {entry.value.toLocaleString()}
                    </div>
                    <div className="text-xs text-gray-500">executions</div>
                  </div>
                </div>
              ))}
            </div>

            {onFilterClick && (
              <div className="mt-4 p-3 bg-blue-50 rounded-lg text-sm text-blue-800">
                <p className="flex items-center gap-2">
                  <svg
                    className="h-4 w-4"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                  Click on a segment to filter executions by trigger type
                </p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default TriggerBreakdown;
