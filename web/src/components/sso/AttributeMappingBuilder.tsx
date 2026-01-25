/**
 * Attribute Mapping Builder Component
 * Maps IdP attributes to Gorax user fields
 */

import React from 'react';
import type { ProviderType } from '../../types/sso';

interface AttributeMappingBuilderProps {
  mapping: Record<string, string>;
  onChange: (mapping: Record<string, string>) => void;
  providerType: ProviderType;
}

interface AttributeField {
  key: string;
  label: string;
  description: string;
  required: boolean;
  samlExample: string;
  oidcExample: string;
}

const attributeFields: AttributeField[] = [
  {
    key: 'email',
    label: 'Email Attribute',
    description: 'The attribute containing the user email address',
    required: true,
    samlExample: 'NameID or email',
    oidcExample: 'email',
  },
  {
    key: 'first_name',
    label: 'First Name Attribute',
    description: 'The attribute containing the user first name',
    required: false,
    samlExample: 'firstName or givenName',
    oidcExample: 'given_name',
  },
  {
    key: 'last_name',
    label: 'Last Name Attribute',
    description: 'The attribute containing the user last name',
    required: false,
    samlExample: 'lastName or surname',
    oidcExample: 'family_name',
  },
  {
    key: 'groups',
    label: 'Groups Attribute',
    description: 'The attribute containing user group memberships',
    required: false,
    samlExample: 'groups or memberOf',
    oidcExample: 'groups',
  },
];

export const AttributeMappingBuilder: React.FC<AttributeMappingBuilderProps> = ({
  mapping,
  onChange,
  providerType,
}) => {
  const handleChange = (key: string, value: string) => {
    onChange({
      ...mapping,
      [key]: value,
    });
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
          Attribute Mapping
        </h3>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          Map {providerType.toUpperCase()} attributes to Gorax user fields
        </p>
      </div>

      {/* Attribute Fields */}
      <div className="space-y-4">
        {attributeFields.map((field) => (
          <div key={field.key}>
            <label
              htmlFor={`attr-${field.key}`}
              className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
            >
              {field.label}
              {field.required && (
                <span className="text-red-500 ml-1">*</span>
              )}
            </label>
            <input
              type="text"
              id={`attr-${field.key}`}
              value={mapping[field.key] || ''}
              onChange={(e) => handleChange(field.key, e.target.value)}
              placeholder={
                providerType === 'saml'
                  ? field.samlExample
                  : field.oidcExample
              }
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
            />
            <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {field.description}
            </p>
            <p className="mt-1 text-xs text-gray-400 dark:text-gray-500">
              Example: {providerType === 'saml' ? field.samlExample : field.oidcExample}
            </p>
          </div>
        ))}
      </div>

      {/* Info Box */}
      <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-md p-4">
        <h4 className="text-sm font-medium text-amber-900 dark:text-amber-300 mb-2">
          Attribute Mapping Guide
        </h4>
        <ul className="text-sm text-amber-800 dark:text-amber-400 space-y-1 list-disc list-inside">
          {providerType === 'saml' ? (
            <>
              <li>SAML attributes come from the assertion sent by your IdP</li>
              <li>Check your IdP documentation for available attribute names</li>
              <li>Common SAML attributes: NameID, email, firstName, lastName, groups</li>
            </>
          ) : (
            <>
              <li>OIDC claims come from the ID token or userinfo endpoint</li>
              <li>Standard OIDC claims: sub, email, given_name, family_name</li>
              <li>Custom claims depend on your IdP configuration</li>
            </>
          )}
        </ul>
      </div>

      {/* Preview */}
      <div className="border border-gray-200 dark:border-gray-700 rounded-md p-4">
        <h4 className="text-sm font-medium text-gray-900 dark:text-white mb-3">
          Mapping Preview
        </h4>
        <div className="space-y-2">
          {Object.entries(mapping).filter(([, value]) => value).length === 0 ? (
            <p className="text-sm text-gray-500 dark:text-gray-400">
              No mappings configured
            </p>
          ) : (
            Object.entries(mapping)
              .filter(([, value]) => value)
              .map(([key, value]) => (
                <div
                  key={key}
                  className="flex items-center justify-between text-sm bg-gray-50 dark:bg-gray-800 px-3 py-2 rounded"
                >
                  <span className="font-medium text-gray-900 dark:text-white">
                    {key}
                  </span>
                  <span className="text-gray-600 dark:text-gray-400">→</span>
                  <span className="text-gray-700 dark:text-gray-300 font-mono">
                    {value}
                  </span>
                </div>
              ))
          )}
        </div>
      </div>
    </div>
  );
};
