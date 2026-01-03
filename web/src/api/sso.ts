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
    return await apiClient.post(
      `${SSO_BASE_PATH}/providers`,
      request
    );
  },

  /**
   * List all SSO providers for the current tenant
   */
  listProviders: async (): Promise<SSOProvider[]> => {
    return await apiClient.get(
      `${SSO_BASE_PATH}/providers`
    );
  },

  /**
   * Get a specific SSO provider
   */
  getProvider: async (providerId: string): Promise<SSOProvider> => {
    return await apiClient.get(
      `${SSO_BASE_PATH}/providers/${providerId}`
    );
  },

  /**
   * Update an SSO provider
   */
  updateProvider: async (
    providerId: string,
    request: UpdateProviderRequest
  ): Promise<SSOProvider> => {
    return await apiClient.put(
      `${SSO_BASE_PATH}/providers/${providerId}`,
      request
    );
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
    return await apiClient.get(
      `${SSO_BASE_PATH}/discover`,
      { params: { email } }
    );
  },

  /**
   * Get SAML metadata for a provider
   */
  getMetadata: async (providerId: string): Promise<string> => {
    return await apiClient.get(
      `${SSO_BASE_PATH}/metadata/${providerId}`
    );
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
    return await apiClient.get(
      `${SSO_BASE_PATH}/providers/${providerId}/events`,
      { params: { limit } }
    );
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
