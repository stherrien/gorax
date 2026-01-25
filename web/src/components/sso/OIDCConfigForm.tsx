/**
 * OIDC Configuration Form Component
 */

import React, { useState } from 'react';
import type { OIDCConfig } from '../../types/sso';
import { AttributeMappingBuilder } from './AttributeMappingBuilder';

interface OIDCConfigFormProps {
  config: Partial<OIDCConfig>;
  onChange: (config: Partial<OIDCConfig>) => void;
  providerId?: string;
  baseUrl: string;
}

const DEFAULT_SCOPES = ['openid', 'profile', 'email'];

export const OIDCConfigForm: React.FC<OIDCConfigFormProps> = ({
  config,
  onChange,
  providerId,
  baseUrl,
}) => {
  const [useDiscovery, setUseDiscovery] = useState(!!config.discovery_url);

  const handleChange = (field: keyof OIDCConfig, value: unknown) => {
    onChange({
      ...config,
      [field]: value,
    });
  };

  const handleScopesChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const scopes = e.target.value.split(',').map((s) => s.trim()).filter(Boolean);
    handleChange('scopes', scopes);
  };

  return (
    <div className="space-y-6">
      {/* Client Credentials */}
      <div className="space-y-4">
        <h3 className="text-sm font-medium text-gray-900 dark:text-white">
          Client Credentials
        </h3>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Client ID <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            value={config.client_id || ''}
            onChange={(e) => handleChange('client_id', e.target.value)}
            placeholder="Your OIDC client ID"
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Client Secret <span className="text-red-500">*</span>
          </label>
          <input
            type="password"
            value={config.client_secret || ''}
            onChange={(e) => handleChange('client_secret', e.target.value)}
            placeholder="Your OIDC client secret"
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
          />
        </div>
      </div>

      {/* Redirect URL (Read-only) */}
      <div>
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Redirect URL
        </label>
        <input
          type="text"
          value={
            config.redirect_url ||
            `${baseUrl}/api/v1/sso/callback/${providerId || 'new'}`
          }
          readOnly
          className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-900 border border-gray-300 dark:border-gray-600 rounded-md text-sm text-gray-700 dark:text-gray-300"
        />
        <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
          Use this URL when configuring your OIDC provider
        </p>
      </div>

      {/* Discovery or Manual Configuration */}
      <div>
        <label className="flex items-center mb-4">
          <input
            type="checkbox"
            checked={useDiscovery}
            onChange={(e) => setUseDiscovery(e.target.checked)}
            className="mr-2"
          />
          <span className="text-sm text-gray-700 dark:text-gray-300">
            Use discovery URL (recommended)
          </span>
        </label>

        {useDiscovery ? (
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Discovery URL <span className="text-red-500">*</span>
            </label>
            <input
              type="url"
              value={config.discovery_url || ''}
              onChange={(e) => handleChange('discovery_url', e.target.value)}
              placeholder="https://idp.example.com/.well-known/openid-configuration"
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
            />
            <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
              Most OIDC providers support automatic discovery
            </p>
          </div>
        ) : (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Authorization URL <span className="text-red-500">*</span>
              </label>
              <input
                type="url"
                value={config.authorization_url || ''}
                onChange={(e) =>
                  handleChange('authorization_url', e.target.value)
                }
                placeholder="https://idp.example.com/oauth/authorize"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Token URL <span className="text-red-500">*</span>
              </label>
              <input
                type="url"
                value={config.token_url || ''}
                onChange={(e) => handleChange('token_url', e.target.value)}
                placeholder="https://idp.example.com/oauth/token"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Userinfo URL
              </label>
              <input
                type="url"
                value={config.userinfo_url || ''}
                onChange={(e) => handleChange('userinfo_url', e.target.value)}
                placeholder="https://idp.example.com/oauth/userinfo"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                JWKS URL
              </label>
              <input
                type="url"
                value={config.jwks_url || ''}
                onChange={(e) => handleChange('jwks_url', e.target.value)}
                placeholder="https://idp.example.com/.well-known/jwks.json"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
              />
            </div>
          </div>
        )}
      </div>

      {/* Scopes */}
      <div>
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Scopes
        </label>
        <input
          type="text"
          value={(config.scopes || DEFAULT_SCOPES).join(', ')}
          onChange={handleScopesChange}
          placeholder="openid, profile, email"
          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
        />
        <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
          Comma-separated list of OAuth scopes
        </p>
      </div>

      {/* Attribute Mapping */}
      <AttributeMappingBuilder
        mapping={config.attribute_mapping || {}}
        onChange={(mapping) => handleChange('attribute_mapping', mapping)}
        providerType="oidc"
      />
    </div>
  );
};
