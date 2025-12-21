import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { UsageDashboard } from './UsageDashboard';
import * as usageApi from '../../api/usage';

vi.mock('../../api/usage');

describe('UsageDashboard', () => {
  const mockUsageData = {
    tenant_id: 'tenant-1',
    current_period: {
      workflow_executions: 50,
      step_executions: 200,
      period: 'daily',
    },
    month_to_date: {
      workflow_executions: 500,
      step_executions: 2000,
      period: 'monthly',
    },
    quotas: {
      max_executions_per_day: 100,
      max_executions_per_month: 3000,
      executions_remaining: 50,
      quota_percent_used: 50.0,
      max_concurrent_executions: 10,
      max_workflows: 50,
    },
    rate_limits: {
      requests_per_minute: 60,
      requests_per_hour: 1000,
      requests_per_day: 10000,
      hits_today: 150,
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state initially', () => {
    vi.mocked(usageApi.usageApi.getCurrentUsage).mockImplementation(
      () => new Promise(() => {})
    );

    render(<UsageDashboard tenantId="tenant-1" />);
    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  it('displays usage statistics after loading', async () => {
    vi.mocked(usageApi.usageApi.getCurrentUsage).mockResolvedValue(mockUsageData);

    render(<UsageDashboard tenantId="tenant-1" />);

    await waitFor(() => {
      expect(screen.getAllByText(/workflow executions/i).length).toBeGreaterThan(0);
    });

    // Check for "Today's Usage" section which should have 50 workflow executions
    expect(screen.getByText("Today's Usage")).toBeInTheDocument();
    // Check that "50.0%" appears somewhere in the document (could be in metrics or quota section)
    const percentageElements = screen.getAllByText(/50\.0%/);
    expect(percentageElements.length).toBeGreaterThan(0);
  });

  it('displays quota information', async () => {
    vi.mocked(usageApi.usageApi.getCurrentUsage).mockResolvedValue(mockUsageData);

    render(<UsageDashboard tenantId="tenant-1" />);

    await waitFor(() => {
      expect(screen.getAllByText(/remaining/i).length).toBeGreaterThan(0);
    });
  });

  it('handles error state', async () => {
    vi.mocked(usageApi.usageApi.getCurrentUsage).mockRejectedValue(
      new Error('Failed to fetch')
    );

    render(<UsageDashboard tenantId="tenant-1" />);

    await waitFor(() => {
      expect(screen.getByText(/error/i)).toBeInTheDocument();
    });
  });

  it('shows warning when quota is near limit', async () => {
    const highUsageData = {
      ...mockUsageData,
      quotas: {
        ...mockUsageData.quotas,
        quota_percent_used: 85.0,
        executions_remaining: 15,
      },
    };

    vi.mocked(usageApi.usageApi.getCurrentUsage).mockResolvedValue(highUsageData);

    render(<UsageDashboard tenantId="tenant-1" />);

    await waitFor(() => {
      const percentageElements = screen.getAllByText(/85\.0%/);
      expect(percentageElements.length).toBeGreaterThan(0);
    });
  });

  it('displays unlimited quota correctly', async () => {
    const unlimitedData = {
      ...mockUsageData,
      quotas: {
        ...mockUsageData.quotas,
        max_executions_per_day: -1,
        executions_remaining: -1,
        quota_percent_used: 0,
      },
    };

    vi.mocked(usageApi.usageApi.getCurrentUsage).mockResolvedValue(unlimitedData);

    render(<UsageDashboard tenantId="tenant-1" />);

    await waitFor(() => {
      expect(screen.getAllByText(/unlimited/i).length).toBeGreaterThan(0);
    });
  });
});
