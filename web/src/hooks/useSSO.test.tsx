/**
 * Tests for SSO custom hooks
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ssoApi } from '../api/sso';
import {
  useSSOProviders,
  useSSOProvider,
  useCreateSSOProvider,
  useUpdateSSOProvider,
  useDeleteSSOProvider,
  useSSOMetadata,
  useTestSSOProvider,
  useDiscoverSSO,
} from './useSSO';
import type { SSOProvider, CreateProviderRequest, ProviderType } from '../types/sso';
import React from 'react';

// Mock the API
vi.mock('../api/sso');

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  const Wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
  return Wrapper;
};

const mockProvider: SSOProvider = {
  id: 'provider-123',
  tenant_id: 'tenant-123',
  name: 'Okta SSO',
  provider_type: 'saml' as ProviderType,
  enabled: true,
  enforce_sso: false,
  config: {
    entity_id: 'https://app.gorax.com',
    acs_url: 'https://app.gorax.com/api/v1/sso/acs',
    idp_entity_id: 'https://okta.com/entity',
    idp_sso_url: 'https://okta.com/sso',
    attribute_mapping: {
      email: 'NameID',
      first_name: 'firstName',
      last_name: 'lastName',
    },
    sign_authn_requests: false,
  },
  domains: ['example.com'],
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

describe('useSSOProviders', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch SSO providers', async () => {
    const mockProviders = [mockProvider];
    vi.mocked(ssoApi.listProviders).mockResolvedValue(mockProviders);

    const { result } = renderHook(() => useSSOProviders(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual(mockProviders);
    expect(ssoApi.listProviders).toHaveBeenCalledTimes(1);
  });

  it('should handle errors', async () => {
    const error = new Error('Failed to fetch providers');
    vi.mocked(ssoApi.listProviders).mockRejectedValue(error);

    const { result } = renderHook(() => useSSOProviders(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(result.current.error).toBeTruthy();
  });
});

describe('useSSOProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch a single provider', async () => {
    vi.mocked(ssoApi.getProvider).mockResolvedValue(mockProvider);

    const { result } = renderHook(() => useSSOProvider('provider-123'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual(mockProvider);
    expect(ssoApi.getProvider).toHaveBeenCalledWith('provider-123');
  });

  it('should not fetch when providerId is undefined', () => {
    const { result } = renderHook(() => useSSOProvider(undefined), {
      wrapper: createWrapper(),
    });

    expect(result.current.data).toBeUndefined();
    expect(ssoApi.getProvider).not.toHaveBeenCalled();
  });
});

describe('useCreateSSOProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should create a provider', async () => {
    vi.mocked(ssoApi.createProvider).mockResolvedValue(mockProvider);

    const { result } = renderHook(() => useCreateSSOProvider(), {
      wrapper: createWrapper(),
    });

    const createRequest: CreateProviderRequest = {
      name: 'Okta SSO',
      provider_type: 'saml',
      enabled: true,
      enforce_sso: false,
      config: mockProvider.config,
      domains: ['example.com'],
    };

    result.current.mutate(createRequest);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(ssoApi.createProvider).toHaveBeenCalledWith(createRequest);
  });

  it('should handle errors', async () => {
    const error = new Error('Failed to create provider');
    vi.mocked(ssoApi.createProvider).mockRejectedValue(error);

    const { result } = renderHook(() => useCreateSSOProvider(), {
      wrapper: createWrapper(),
    });

    const createRequest: CreateProviderRequest = {
      name: 'Okta SSO',
      provider_type: 'saml',
      enabled: true,
      enforce_sso: false,
      config: mockProvider.config,
      domains: ['example.com'],
    };

    result.current.mutate(createRequest);

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(result.current.error).toBeTruthy();
  });
});

describe('useUpdateSSOProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should update a provider', async () => {
    const updatedProvider = { ...mockProvider, name: 'Updated Okta' };
    vi.mocked(ssoApi.updateProvider).mockResolvedValue(updatedProvider);

    const { result } = renderHook(() => useUpdateSSOProvider(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      providerId: 'provider-123',
      updates: { name: 'Updated Okta' },
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(ssoApi.updateProvider).toHaveBeenCalledWith('provider-123', {
      name: 'Updated Okta',
    });
  });
});

describe('useDeleteSSOProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should delete a provider', async () => {
    vi.mocked(ssoApi.deleteProvider).mockResolvedValue();

    const { result } = renderHook(() => useDeleteSSOProvider(), {
      wrapper: createWrapper(),
    });

    result.current.mutate('provider-123');

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(ssoApi.deleteProvider).toHaveBeenCalledWith('provider-123');
  });
});

describe('useSSOMetadata', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch SAML metadata', async () => {
    const metadata = '<xml>metadata</xml>';
    vi.mocked(ssoApi.getMetadata).mockResolvedValue(metadata);

    const { result } = renderHook(() => useSSOMetadata('provider-123'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toBe(metadata);
    expect(ssoApi.getMetadata).toHaveBeenCalledWith('provider-123');
  });
});

describe('useTestSSOProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should test a provider', async () => {
    // Mock initiateLogin to avoid actual redirect
    vi.mocked(ssoApi.initiateLogin).mockImplementation(() => {});

    const { result } = renderHook(() => useTestSSOProvider(), {
      wrapper: createWrapper(),
    });

    result.current.mutate('provider-123');

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(ssoApi.initiateLogin).toHaveBeenCalledWith('provider-123', '/sso-test');
  });
});

describe('useDiscoverSSO', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should discover SSO provider', async () => {
    const discovery = {
      sso_available: true,
      provider_id: 'provider-123',
      provider_name: 'Okta SSO',
      provider_type: 'saml' as ProviderType,
      enforce_sso: false,
    };
    vi.mocked(ssoApi.discoverProvider).mockResolvedValue(discovery);

    const { result } = renderHook(() => useDiscoverSSO('user@example.com'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual(discovery);
    expect(ssoApi.discoverProvider).toHaveBeenCalledWith('user@example.com');
  });

  it('should not discover when email is empty', () => {
    const { result } = renderHook(() => useDiscoverSSO(''), {
      wrapper: createWrapper(),
    });

    expect(result.current.data).toBeUndefined();
    expect(ssoApi.discoverProvider).not.toHaveBeenCalled();
  });
});
