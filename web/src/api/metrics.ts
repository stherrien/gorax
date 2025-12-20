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

    const response = await apiClient.get(`/metrics/trends?${queryParams.toString()}`);
    return response.data;
  },

  getDurationStats: async (params?: MetricsQueryParams): Promise<DurationStatsResponse> => {
    const queryParams = new URLSearchParams();
    if (params?.days) queryParams.append('days', params.days.toString());
    if (params?.startDate) queryParams.append('startDate', params.startDate);
    if (params?.endDate) queryParams.append('endDate', params.endDate);

    const response = await apiClient.get(`/metrics/duration?${queryParams.toString()}`);
    return response.data;
  },

  getTopFailures: async (params?: MetricsQueryParams): Promise<TopFailuresResponse> => {
    const queryParams = new URLSearchParams();
    if (params?.days) queryParams.append('days', params.days.toString());
    if (params?.startDate) queryParams.append('startDate', params.startDate);
    if (params?.endDate) queryParams.append('endDate', params.endDate);
    if (params?.limit) queryParams.append('limit', params.limit.toString());

    const response = await apiClient.get(`/metrics/failures?${queryParams.toString()}`);
    return response.data;
  },

  getTriggerBreakdown: async (params?: MetricsQueryParams): Promise<TriggerBreakdownResponse> => {
    const queryParams = new URLSearchParams();
    if (params?.days) queryParams.append('days', params.days.toString());
    if (params?.startDate) queryParams.append('startDate', params.startDate);
    if (params?.endDate) queryParams.append('endDate', params.endDate);

    const response = await apiClient.get(`/metrics/trigger-breakdown?${queryParams.toString()}`);
    return response.data;
  },
};

export default metricsApi;
