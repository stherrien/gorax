import { useState, useEffect } from 'react';
import metricsApi, {
  ExecutionTrend,
  DurationStats,
  TopFailure,
  TriggerTypeBreakdown,
  MetricsQueryParams,
} from '../api/metrics';

export const useExecutionTrends = (params?: MetricsQueryParams) => {
  const [trends, setTrends] = useState<ExecutionTrend[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchTrends = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await metricsApi.getExecutionTrends(params);
      setTrends(response.trends);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch trends');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTrends();
  }, [JSON.stringify(params)]);

  return { trends, loading, error, refetch: fetchTrends };
};

export const useDurationStats = (params?: MetricsQueryParams) => {
  const [stats, setStats] = useState<DurationStats[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchStats = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await metricsApi.getDurationStats(params);
      setStats(response.stats);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch duration stats');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStats();
  }, [JSON.stringify(params)]);

  return { stats, loading, error, refetch: fetchStats };
};

export const useTopFailures = (params?: MetricsQueryParams) => {
  const [failures, setFailures] = useState<TopFailure[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchFailures = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await metricsApi.getTopFailures(params);
      setFailures(response.failures);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch failures');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFailures();
  }, [JSON.stringify(params)]);

  return { failures, loading, error, refetch: fetchFailures };
};

export const useTriggerBreakdown = (params?: MetricsQueryParams) => {
  const [breakdown, setBreakdown] = useState<TriggerTypeBreakdown[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchBreakdown = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await metricsApi.getTriggerBreakdown(params);
      setBreakdown(response.breakdown);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch trigger breakdown');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchBreakdown();
  }, [JSON.stringify(params)]);

  return { breakdown, loading, error, refetch: fetchBreakdown };
};

export const useAllMetrics = (params?: MetricsQueryParams) => {
  const trends = useExecutionTrends(params);
  const durationStats = useDurationStats(params);
  const failures = useTopFailures(params);
  const triggerBreakdown = useTriggerBreakdown(params);

  const loading = trends.loading || durationStats.loading || failures.loading || triggerBreakdown.loading;
  const error = trends.error || durationStats.error || failures.error || triggerBreakdown.error;

  const refetchAll = () => {
    trends.refetch();
    durationStats.refetch();
    failures.refetch();
    triggerBreakdown.refetch();
  };

  return {
    trends: trends.trends,
    durationStats: durationStats.stats,
    failures: failures.failures,
    triggerBreakdown: triggerBreakdown.breakdown,
    loading,
    error,
    refetch: refetchAll,
  };
};
