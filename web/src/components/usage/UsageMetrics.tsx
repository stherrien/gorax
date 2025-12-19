import { UsageResponse } from '../../api/usage';

interface UsageMetricsProps {
  usage: UsageResponse;
}

export const UsageMetrics: React.FC<UsageMetricsProps> = ({ usage }) => {
  const getQuotaColor = (percent: number): string => {
    if (percent >= 90) return 'bg-red-500';
    if (percent >= 80) return 'bg-yellow-500';
    return 'bg-green-500';
  };

  const formatNumber = (num: number): string => {
    if (num === -1) return 'Unlimited';
    return num.toLocaleString();
  };

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
      {/* Workflow Executions */}
      <div className="bg-white rounded-lg shadow p-6">
        <div className="flex items-center justify-between mb-2">
          <h3 className="text-sm font-medium text-gray-600">Workflow Executions</h3>
          <svg
            className="w-5 h-5 text-blue-500"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M13 10V3L4 14h7v7l9-11h-7z"
            />
          </svg>
        </div>
        <div className="text-3xl font-bold text-gray-900">
          {usage.current_period.workflow_executions}
        </div>
        <div className="text-sm text-gray-500 mt-1">Today</div>

        {/* Progress bar */}
        {usage.quotas.max_executions_per_day !== -1 && (
          <div className="mt-4">
            <div className="flex justify-between text-xs text-gray-600 mb-1">
              <span>
                {usage.current_period.workflow_executions} /{' '}
                {usage.quotas.max_executions_per_day}
              </span>
              <span>{usage.quotas.quota_percent_used.toFixed(1)}%</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div
                className={`h-2 rounded-full transition-all ${getQuotaColor(
                  usage.quotas.quota_percent_used
                )}`}
                style={{
                  width: `${Math.min(usage.quotas.quota_percent_used, 100)}%`,
                }}
              />
            </div>
          </div>
        )}
      </div>

      {/* Remaining Quota */}
      <div className="bg-white rounded-lg shadow p-6">
        <div className="flex items-center justify-between mb-2">
          <h3 className="text-sm font-medium text-gray-600">Remaining Quota</h3>
          <svg
            className="w-5 h-5 text-green-500"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        </div>
        <div className="text-3xl font-bold text-gray-900">
          {formatNumber(usage.quotas.executions_remaining)}
        </div>
        <div className="text-sm text-gray-500 mt-1">Executions left today</div>
      </div>

      {/* Concurrent Executions */}
      <div className="bg-white rounded-lg shadow p-6">
        <div className="flex items-center justify-between mb-2">
          <h3 className="text-sm font-medium text-gray-600">Concurrent Limit</h3>
          <svg
            className="w-5 h-5 text-purple-500"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
            />
          </svg>
        </div>
        <div className="text-3xl font-bold text-gray-900">
          {formatNumber(usage.quotas.max_concurrent_executions)}
        </div>
        <div className="text-sm text-gray-500 mt-1">Max parallel workflows</div>
      </div>
    </div>
  );
};
