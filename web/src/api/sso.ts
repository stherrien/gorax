/**
 * SSO API Client
 */

import { apiClient } from './client';
import type {
  SSOProvider,
  CreateProviderRequest,
  UpdateProviderRequest,
  ProviderDiscoveryResponse,
  SSOLoginEvent,
} from '../types/sso';

const SSO_BASE_PATH = '/api/v1/sso';

/**
 * SSO Provider Management
 */
export const ssoApi = {
  /**
   * Create a new SSO provider
   */
  createProvider: async (request: CreateProviderRequest): Promise<SSOProvider> => {
    const response = await apiClient.post<SSOProvider>(
      `${SSO_BASE_PATH}/providers`,
      request
    );
    return response.data;
  },

  /**
   * List all SSO providers for the current tenant
   */
  listProviders: async (): Promise<SSOProvider[]> => {
    const response = await apiClient.get<SSOProvider[]>(
      `${SSO_BASE_PATH}/providers`
    );
    return response.data;
  },

  /**
   * Get a specific SSO provider
   */
  getProvider: async (providerId: string): Promise<SSOProvider> => {
    const response = await apiClient.get<SSOProvider>(
      `${SSO_BASE_PATH}/providers/${providerId}`
    );
    return response.data;
  },

  /**
   * Update an SSO provider
   */
  updateProvider: async (
    providerId: string,
    request: UpdateProviderRequest
  ): Promise<SSOProvider> => {
    const response = await apiClient.put<SSOProvider>(
      `${SSO_BASE_PATH}/providers/${providerId}`,
      request
    );
    return response.data;
  },

  /**
   * Delete an SSO provider
   */
  deleteProvider: async (providerId: string): Promise<void> => {
    await apiClient.delete(`${SSO_BASE_PATH}/providers/${providerId}`);
  },

  /**
   * Discover SSO provider by email domain
   */
  discoverProvider: async (email: string): Promise<ProviderDiscoveryResponse> => {
    const response = await apiClient.get<ProviderDiscoveryResponse>(
      `${SSO_BASE_PATH}/discover`,
      { params: { email } }
    );
    return response.data;
  },

  /**
   * Get SAML metadata for a provider
   */
  getMetadata: async (providerId: string): Promise<string> => {
    const response = await apiClient.get<string>(
      `${SSO_BASE_PATH}/metadata/${providerId}`,
      { responseType: 'text' }
    );
    return response.data;
  },

  /**
   * Initiate SSO login
   * This will redirect the browser to the IdP
   */
  initiateLogin: (providerId: string, relayState?: string): void => {
    const url = new URL(
      `${SSO_BASE_PATH}/login/${providerId}`,
      window.location.origin
    );
    if (relayState) {
      url.searchParams.set('relay_state', relayState);
    }
    window.location.href = url.toString();
  },

  /**
   * Get login events for a provider (admin only)
   */
  getLoginEvents: async (
    providerId: string,
    limit = 100
  ): Promise<SSOLoginEvent[]> => {
    const response = await apiClient.get<SSOLoginEvent[]>(
      `${SSO_BASE_PATH}/providers/${providerId}/events`,
      { params: { limit } }
    );
    return response.data;
  },
};

/**
 * Helper function to build SSO login URL
 */
export const buildSSOLoginUrl = (
  providerId: string,
  relayState?: string
): string => {
  const url = new URL(
    `${SSO_BASE_PATH}/login/${providerId}`,
    window.location.origin
  );
  if (relayState) {
    url.searchParams.set('relay_state', relayState);
  }
  return url.toString();
};

/**
 * Helper function to extract domain from email
 */
export const extractEmailDomain = (email: string): string | null => {
  const match = email.match(/@(.+)$/);
  return match ? match[1] : null;
};

/**
 * Validate provider configuration
 */
export const validateProviderConfig = (
  providerType: 'saml' | 'oidc',
  config: unknown
): string[] => {
  const errors: string[] = [];

  if (providerType === 'saml') {
    const samlConfig = config as Record<string, unknown>;
    if (!samlConfig.entity_id) {
      errors.push('Entity ID is required');
    }
    if (!samlConfig.acs_url) {
      errors.push('ACS URL is required');
    }
    if (!samlConfig.idp_metadata_url && !samlConfig.idp_metadata_xml) {
      errors.push('IdP metadata URL or XML is required');
    }
  } else if (providerType === 'oidc') {
    const oidcConfig = config as Record<string, unknown>;
    if (!oidcConfig.client_id) {
      errors.push('Client ID is required');
    }
    if (!oidcConfig.client_secret) {
      errors.push('Client Secret is required');
    }
    if (!oidcConfig.discovery_url) {
      errors.push('Discovery URL is required');
    }
    if (!oidcConfig.redirect_url) {
      errors.push('Redirect URL is required');
    }
  }

  return errors;
};

/**
 * Generate default SAML config
 */
export const getDefaultSAMLConfig = (baseUrl: string) => ({
  entity_id: baseUrl,
  acs_url: `${baseUrl}/api/v1/sso/acs`,
  idp_metadata_url: '',
  idp_entity_id: '',
  idp_sso_url: '',
  attribute_mapping: {
    email: 'NameID',
    first_name: 'firstName',
    last_name: 'lastName',
    groups: 'groups',
  },
  sign_authn_requests: false,
});

/**
 * Generate default OIDC config
 */
export const getDefaultOIDCConfig = (baseUrl: string, providerId: string) => ({
  client_id: '',
  client_secret: '',
  discovery_url: '',
  redirect_url: `${baseUrl}/api/v1/sso/callback/${providerId}`,
  scopes: ['openid', 'profile', 'email'],
  attribute_mapping: {
    email: 'email',
    first_name: 'given_name',
    last_name: 'family_name',
    groups: 'groups',
  },
});
