/**
 * Custom hooks for SSO operations
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { ssoApi } from '../api/sso';
import type {
  CreateProviderRequest,
  UpdateProviderRequest,
} from '../types/sso';

// Query keys
export const ssoKeys = {
  all: ['sso'] as const,
  providers: () => [...ssoKeys.all, 'providers'] as const,
  provider: (id: string) => [...ssoKeys.all, 'provider', id] as const,
  metadata: (id: string) => [...ssoKeys.all, 'metadata', id] as const,
  discover: (email: string) => [...ssoKeys.all, 'discover', email] as const,
};

/**
 * Hook to fetch all SSO providers
 */
export const useSSOProviders = () => {
  return useQuery({
    queryKey: ssoKeys.providers(),
    queryFn: () => ssoApi.listProviders(),
  });
};

/**
 * Hook to fetch a single SSO provider
 */
export const useSSOProvider = (providerId: string | undefined) => {
  return useQuery({
    queryKey: ssoKeys.provider(providerId || ''),
    queryFn: () => ssoApi.getProvider(providerId!),
    enabled: !!providerId,
  });
};

/**
 * Hook to create an SSO provider
 */
export const useCreateSSOProvider = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateProviderRequest) =>
      ssoApi.createProvider(request),
    onSuccess: () => {
      // Invalidate providers list
      queryClient.invalidateQueries({ queryKey: ssoKeys.providers() });
    },
  });
};

/**
 * Hook to update an SSO provider
 */
export const useUpdateSSOProvider = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      providerId,
      updates,
    }: {
      providerId: string;
      updates: UpdateProviderRequest;
    }) => ssoApi.updateProvider(providerId, updates),
    onSuccess: (data) => {
      // Invalidate providers list and specific provider
      queryClient.invalidateQueries({ queryKey: ssoKeys.providers() });
      queryClient.invalidateQueries({ queryKey: ssoKeys.provider(data.id) });
    },
  });
};

/**
 * Hook to delete an SSO provider
 */
export const useDeleteSSOProvider = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (providerId: string) => ssoApi.deleteProvider(providerId),
    onSuccess: (_, providerId) => {
      // Invalidate providers list and remove specific provider from cache
      queryClient.invalidateQueries({ queryKey: ssoKeys.providers() });
      queryClient.removeQueries({ queryKey: ssoKeys.provider(providerId) });
    },
  });
};

/**
 * Hook to fetch SAML metadata for a provider
 */
export const useSSOMetadata = (providerId: string | undefined) => {
  return useQuery({
    queryKey: ssoKeys.metadata(providerId || ''),
    queryFn: () => ssoApi.getMetadata(providerId!),
    enabled: !!providerId,
  });
};

/**
 * Hook to test an SSO provider connection
 * This initiates a test login flow
 */
export const useTestSSOProvider = () => {
  return useMutation({
    mutationFn: (providerId: string) => {
      // Initiate login with a test relay state
      ssoApi.initiateLogin(providerId, '/sso-test');
      return Promise.resolve();
    },
  });
};

/**
 * Hook to discover SSO provider by email
 */
export const useDiscoverSSO = (email: string) => {
  return useQuery({
    queryKey: ssoKeys.discover(email),
    queryFn: () => ssoApi.discoverProvider(email),
    enabled: !!email && email.includes('@'),
    retry: false, // Don't retry discovery failures
  });
};

/**
 * Hook to toggle provider enabled status
 */
export const useToggleProviderStatus = () => {
  const updateMutation = useUpdateSSOProvider();

  return useMutation({
    mutationFn: ({ providerId, enabled }: { providerId: string; enabled: boolean }) =>
      updateMutation.mutateAsync({ providerId, updates: { enabled } }),
  });
};

/**
 * Hook to update provider domains
 */
export const useUpdateProviderDomains = () => {
  const updateMutation = useUpdateSSOProvider();

  return useMutation({
    mutationFn: ({ providerId, domains }: { providerId: string; domains: string[] }) =>
      updateMutation.mutateAsync({ providerId, updates: { domains } }),
  });
};
