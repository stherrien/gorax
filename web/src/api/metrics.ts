import { apiClient } from './client';

export interface ExecutionTrend {
  date: string;
  count: number;
  success: number;
  failed: number;
}

export interface DurationStats {
  workflowId: string;
  workflowName: string;
  avgDuration: number;
  p50Duration: number;
  p90Duration: number;
  p99Duration: number;
  totalRuns: number;
}

export interface TopFailure {
  workflowId: string;
  workflowName: string;
  failureCount: number;
  lastFailedAt?: string;
  errorPreview?: string;
}

export interface TriggerTypeBreakdown {
  triggerType: string;
  count: number;
  percentage: number;
}

export interface ExecutionTrendsResponse {
  trends: ExecutionTrend[];
  startDate: string;
  endDate: string;
  groupBy: string;
}

export interface DurationStatsResponse {
  stats: DurationStats[];
  startDate: string;
  endDate: string;
}

export interface TopFailuresResponse {
  failures: TopFailure[];
  startDate: string;
  endDate: string;
  limit: number;
}

export interface TriggerBreakdownResponse {
  breakdown: TriggerTypeBreakdown[];
  startDate: string;
  endDate: string;
}

export interface MetricsQueryParams {
  days?: number;
  startDate?: string;
  endDate?: string;
  groupBy?: 'hour' | 'day';
  limit?: number;
}

const metricsApi = {
  getExecutionTrends: async (params?: MetricsQueryParams): Promise<ExecutionTrendsResponse> => {
    const queryParams = new URLSearchParams();
    if (params?.days) queryParams.append('days', params.days.toString());
    if (params?.startDate) queryParams.append('startDate', params.startDate);
    if (params?.endDate) queryParams.append('endDate', params.endDate);
    if (params?.groupBy) queryParams.append('groupBy', params.groupBy);

    const query = queryParams.toString();
    const response = await apiClient.get(`/api/v1/metrics/trends${query ? `?${query}` : ''}`);
    return response;
  },

  getDurationStats: async (params?: MetricsQueryParams): Promise<DurationStatsResponse> => {
    const queryParams = new URLSearchParams();
    if (params?.days) queryParams.append('days', params.days.toString());
    if (params?.startDate) queryParams.append('startDate', params.startDate);
    if (params?.endDate) queryParams.append('endDate', params.endDate);

    const query = queryParams.toString();
    const response = await apiClient.get(`/api/v1/metrics/duration${query ? `?${query}` : ''}`);
    return response;
  },

  getTopFailures: async (params?: MetricsQueryParams): Promise<TopFailuresResponse> => {
    const queryParams = new URLSearchParams();
    if (params?.days) queryParams.append('days', params.days.toString());
    if (params?.startDate) queryParams.append('startDate', params.startDate);
    if (params?.endDate) queryParams.append('endDate', params.endDate);
    if (params?.limit) queryParams.append('limit', params.limit.toString());

    const query = queryParams.toString();
    const response = await apiClient.get(`/api/v1/metrics/failures${query ? `?${query}` : ''}`);
    return response;
  },

  getTriggerBreakdown: async (params?: MetricsQueryParams): Promise<TriggerBreakdownResponse> => {
    const queryParams = new URLSearchParams();
    if (params?.days) queryParams.append('days', params.days.toString());
    if (params?.startDate) queryParams.append('startDate', params.startDate);
    if (params?.endDate) queryParams.append('endDate', params.endDate);

    const query = queryParams.toString();
    const response = await apiClient.get(`/api/v1/metrics/trigger-breakdown${query ? `?${query}` : ''}`);
    return response;
  },
};

export default metricsApi;
