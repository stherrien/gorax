/**
 * SAML Configuration Form Component
 */

import React, { useState } from 'react';
import type { SAMLConfig } from '../../types/sso';
import { AttributeMappingBuilder } from './AttributeMappingBuilder';

interface SAMLConfigFormProps {
  config: Partial<SAMLConfig>;
  onChange: (config: Partial<SAMLConfig>) => void;
  baseUrl: string;
}

export const SAMLConfigForm: React.FC<SAMLConfigFormProps> = ({
  config,
  onChange,
  baseUrl,
}) => {
  const [metadataSource, setMetadataSource] = useState<'url' | 'xml'>(
    config.idp_metadata_url ? 'url' : 'xml'
  );

  const handleChange = (field: keyof SAMLConfig, value: unknown) => {
    onChange({
      ...config,
      [field]: value,
    });
  };

  return (
    <div className="space-y-6">
      {/* SP Information (Read-only) */}
      <div className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
        <h3 className="text-sm font-medium text-gray-900 dark:text-white mb-3">
          Service Provider (SP) Information
        </h3>
        <div className="space-y-3">
          <div>
            <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
              Entity ID
            </label>
            <input
              type="text"
              value={config.entity_id || baseUrl}
              readOnly
              className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-900 border border-gray-300 dark:border-gray-600 rounded-md text-sm text-gray-700 dark:text-gray-300"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
              ACS URL (Assertion Consumer Service)
            </label>
            <input
              type="text"
              value={config.acs_url || `${baseUrl}/api/v1/sso/acs`}
              readOnly
              className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-900 border border-gray-300 dark:border-gray-600 rounded-md text-sm text-gray-700 dark:text-gray-300"
            />
          </div>
        </div>
      </div>

      {/* IdP Metadata Source */}
      <div>
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
          Identity Provider (IdP) Metadata
        </label>
        <div className="flex gap-4 mb-4">
          <label className="flex items-center">
            <input
              type="radio"
              name="metadata-source"
              checked={metadataSource === 'url'}
              onChange={() => setMetadataSource('url')}
              className="mr-2"
            />
            <span className="text-sm text-gray-700 dark:text-gray-300">
              Metadata URL
            </span>
          </label>
          <label className="flex items-center">
            <input
              type="radio"
              name="metadata-source"
              checked={metadataSource === 'xml'}
              onChange={() => setMetadataSource('xml')}
              className="mr-2"
            />
            <span className="text-sm text-gray-700 dark:text-gray-300">
              Upload XML
            </span>
          </label>
        </div>

        {metadataSource === 'url' ? (
          <input
            type="url"
            value={config.idp_metadata_url || ''}
            onChange={(e) => handleChange('idp_metadata_url', e.target.value)}
            placeholder="https://idp.example.com/metadata"
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
          />
        ) : (
          <textarea
            value={config.idp_metadata_xml || ''}
            onChange={(e) => handleChange('idp_metadata_xml', e.target.value)}
            placeholder="Paste IdP metadata XML here..."
            rows={6}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white font-mono text-xs"
          />
        )}
      </div>

      {/* Manual IdP Configuration */}
      <div className="space-y-4">
        <h3 className="text-sm font-medium text-gray-900 dark:text-white">
          IdP Configuration
        </h3>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            IdP Entity ID
          </label>
          <input
            type="text"
            value={config.idp_entity_id || ''}
            onChange={(e) => handleChange('idp_entity_id', e.target.value)}
            placeholder="https://idp.example.com/entity"
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            IdP SSO URL
          </label>
          <input
            type="url"
            value={config.idp_sso_url || ''}
            onChange={(e) => handleChange('idp_sso_url', e.target.value)}
            placeholder="https://idp.example.com/sso"
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
          />
        </div>
      </div>

      {/* Advanced Options */}
      <div className="space-y-4">
        <h3 className="text-sm font-medium text-gray-900 dark:text-white">
          Advanced Options
        </h3>
        <label className="flex items-center">
          <input
            type="checkbox"
            checked={config.sign_authn_requests || false}
            onChange={(e) =>
              handleChange('sign_authn_requests', e.target.checked)
            }
            className="mr-2"
          />
          <span className="text-sm text-gray-700 dark:text-gray-300">
            Sign authentication requests
          </span>
        </label>
      </div>

      {/* Attribute Mapping */}
      <AttributeMappingBuilder
        mapping={config.attribute_mapping || {}}
        onChange={(mapping) => handleChange('attribute_mapping', mapping)}
        providerType="saml"
      />
    </div>
  );
};
