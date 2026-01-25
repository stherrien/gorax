/**
 * SSO (Single Sign-On) Type Definitions
 */

export type ProviderType = 'saml' | 'oidc';

export type LoginStatus = 'success' | 'failure' | 'error';

/**
 * SAML Provider Configuration
 */
export interface SAMLConfig {
  entity_id: string;
  acs_url: string;
  idp_metadata_url?: string;
  idp_metadata_xml?: string;
  idp_entity_id: string;
  idp_sso_url: string;
  certificate?: string;
  private_key?: string;
  attribute_mapping: Record<string, string>;
  sign_authn_requests: boolean;
}

/**
 * OIDC Provider Configuration
 */
export interface OIDCConfig {
  client_id: string;
  client_secret: string;
  discovery_url: string;
  authorization_url?: string;
  token_url?: string;
  userinfo_url?: string;
  jwks_url?: string;
  redirect_url: string;
  scopes: string[];
  attribute_mapping: Record<string, string>;
}

/**
 * SSO Provider
 */
export interface SSOProvider {
  id: string;
  tenant_id: string;
  name: string;
  provider_type: ProviderType;
  enabled: boolean;
  enforce_sso: boolean;
  config: SAMLConfig | OIDCConfig;
  domains: string[];
  created_at: string;
  updated_at: string;
  created_by?: string;
  updated_by?: string;
}

/**
 * SSO Connection (User to Provider mapping)
 */
export interface SSOConnection {
  id: string;
  user_id: string;
  sso_provider_id: string;
  external_id: string;
  attributes?: Record<string, unknown>;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
}

/**
 * SSO Login Event
 */
export interface SSOLoginEvent {
  id: string;
  sso_provider_id: string;
  user_id?: string;
  external_id: string;
  status: LoginStatus;
  error_message?: string;
  ip_address?: string;
  user_agent?: string;
  created_at: string;
}

/**
 * Create Provider Request
 */
export interface CreateProviderRequest {
  name: string;
  provider_type: ProviderType;
  enabled: boolean;
  enforce_sso: boolean;
  config: SAMLConfig | OIDCConfig;
  domains: string[];
}

/**
 * Update Provider Request
 */
export interface UpdateProviderRequest {
  name?: string;
  enabled?: boolean;
  enforce_sso?: boolean;
  config?: SAMLConfig | OIDCConfig;
  domains?: string[];
}

/**
 * Provider Discovery Response
 */
export interface ProviderDiscoveryResponse {
  sso_available: boolean;
  provider_id?: string;
  provider_name?: string;
  provider_type?: ProviderType;
  enforce_sso?: boolean;
}

/**
 * Authentication Response
 */
export interface AuthenticationResponse {
  user_attributes: {
    external_id: string;
    email: string;
    first_name?: string;
    last_name?: string;
    groups?: string[];
    attributes?: Record<string, string>;
  };
  session_token: string;
  expires_at: string;
}
