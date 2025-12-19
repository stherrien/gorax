import { useEffect, useState } from 'react';
import { usageApi, UsageResponse } from '../../api/usage';
import { UsageMetrics } from './UsageMetrics';

interface UsageDashboardProps {
  tenantId: string;
}

export const UsageDashboard: React.FC<UsageDashboardProps> = ({ tenantId }) => {
  const [usage, setUsage] = useState<UsageResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchUsage = async () => {
      try {
        setLoading(true);
        setError(null);
        const data = await usageApi.getCurrentUsage(tenantId);
        setUsage(data);
      } catch (err) {
        setError(err as Error);
      } finally {
        setLoading(false);
      }
    };

    fetchUsage();
    // Refresh every 30 seconds
    const interval = setInterval(fetchUsage, 30000);

    return () => clearInterval(interval);
  }, [tenantId]);

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-gray-600">Loading usage data...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-red-600">Error loading usage: {error.message}</div>
      </div>
    );
  }

  if (!usage) {
    return null;
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-gray-900">Usage & Quotas</h2>
        <button
          onClick={() => window.location.reload()}
          className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Refresh
        </button>
      </div>

      <UsageMetrics usage={usage} />

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Current Period */}
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4">Today's Usage</h3>
          <div className="space-y-3">
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Workflow Executions</span>
              <span className="font-semibold text-lg">
                {usage.current_period.workflow_executions}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Step Executions</span>
              <span className="font-semibold text-lg">
                {usage.current_period.step_executions}
              </span>
            </div>
          </div>
        </div>

        {/* Month to Date */}
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4">This Month</h3>
          <div className="space-y-3">
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Workflow Executions</span>
              <span className="font-semibold text-lg">
                {usage.month_to_date.workflow_executions}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Step Executions</span>
              <span className="font-semibold text-lg">
                {usage.month_to_date.step_executions}
              </span>
            </div>
          </div>
        </div>

        {/* Quota Information */}
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4">Daily Quota</h3>
          <div className="space-y-3">
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Limit</span>
              <span className="font-semibold">
                {usage.quotas.max_executions_per_day === -1
                  ? 'Unlimited'
                  : usage.quotas.max_executions_per_day}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Remaining</span>
              <span className="font-semibold">
                {usage.quotas.executions_remaining === -1
                  ? 'Unlimited'
                  : usage.quotas.executions_remaining}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Used</span>
              <span
                className={`font-semibold ${
                  usage.quotas.quota_percent_used >= 90
                    ? 'text-red-600'
                    : usage.quotas.quota_percent_used >= 80
                    ? 'text-yellow-600'
                    : 'text-green-600'
                }`}
              >
                {usage.quotas.quota_percent_used.toFixed(1)}%
              </span>
            </div>
          </div>
        </div>

        {/* Rate Limits */}
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4">Rate Limits</h3>
          <div className="space-y-3">
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Per Minute</span>
              <span className="font-semibold">
                {usage.rate_limits.requests_per_minute === -1
                  ? 'Unlimited'
                  : usage.rate_limits.requests_per_minute}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Hits Today</span>
              <span className="font-semibold">{usage.rate_limits.hits_today}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
