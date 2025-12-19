import apiClient from './client';

export interface UsageResponse {
  tenant_id: string;
  current_period: PeriodUsage;
  month_to_date: PeriodUsage;
  quotas: QuotaInfo;
  rate_limits: RateLimitInfo;
}

export interface PeriodUsage {
  workflow_executions: number;
  step_executions: number;
  period: string;
}

export interface QuotaInfo {
  max_executions_per_day: number;
  max_executions_per_month: number;
  executions_remaining: number;
  quota_percent_used: number;
  max_concurrent_executions: number;
  max_workflows: number;
}

export interface RateLimitInfo {
  requests_per_minute: number;
  requests_per_hour: number;
  requests_per_day: number;
  hits_today: number;
}

export interface UsageByDate {
  date: string;
  workflow_executions: number;
  step_executions: number;
}

export interface UsageHistoryResponse {
  usage: UsageByDate[];
  total: number;
  page: number;
  limit: number;
  start_date: string;
  end_date: string;
}

export const usageApi = {
  getCurrentUsage: async (tenantId: string): Promise<UsageResponse> => {
    return apiClient.get(`/api/tenants/${tenantId}/usage`);
  },

  getUsageHistory: async (
    tenantId: string,
    startDate?: string,
    endDate?: string,
    page: number = 1,
    limit: number = 30
  ): Promise<UsageHistoryResponse> => {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
    });

    if (startDate) {
      params.append('start_date', startDate);
    }
    if (endDate) {
      params.append('end_date', endDate);
    }

    return apiClient.get(`/api/tenants/${tenantId}/usage/history?${params}`);
  },
};

export default usageApi;
