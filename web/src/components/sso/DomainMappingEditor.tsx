/**
 * Domain Mapping Editor Component
 * Allows adding/removing email domains for SSO provider
 */

import React, { useState } from 'react';

interface DomainMappingEditorProps {
  domains: string[];
  onChange: (domains: string[]) => void;
}

export const DomainMappingEditor: React.FC<DomainMappingEditorProps> = ({
  domains,
  onChange,
}) => {
  const [newDomain, setNewDomain] = useState('');
  const [error, setError] = useState('');

  const validateDomain = (domain: string): boolean => {
    // Basic domain validation
    // eslint-disable-next-line no-useless-escape
    const domainRegex = /^[a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,}$/i;
    return domainRegex.test(domain);
  };

  const handleAddDomain = () => {
    const normalized = newDomain.trim().toLowerCase();

    if (!normalized) {
      return;
    }

    // Validate format
    if (!validateDomain(normalized)) {
      setError('Invalid domain format. Example: example.com');
      return;
    }

    // Check for duplicates
    if (domains.includes(normalized)) {
      setError('Domain already exists');
      return;
    }

    // Add domain
    onChange([...domains, normalized]);
    setNewDomain('');
    setError('');
  };

  const handleRemoveDomain = (domainToRemove: string) => {
    onChange(domains.filter((d) => d !== domainToRemove));
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleAddDomain();
    }
  };

  return (
    <div className="space-y-4">
      {/* Header */}
      <div>
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Email Domains
        </label>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          Users with email addresses from these domains will be redirected to this SSO provider
        </p>
      </div>

      {/* Add Domain Input */}
      <div className="flex gap-2">
        <div className="flex-1">
          <input
            type="text"
            value={newDomain}
            onChange={(e) => {
              setNewDomain(e.target.value);
              setError('');
            }}
            onKeyDown={handleKeyDown}
            placeholder="Enter domain (e.g., example.com)"
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
          />
          {error && (
            <p className="mt-1 text-sm text-red-600 dark:text-red-400">{error}</p>
          )}
        </div>
        <button
          type="button"
          onClick={handleAddDomain}
          disabled={!newDomain.trim()}
          className="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed rounded-md transition-colors"
          aria-label="Add domain"
        >
          Add Domain
        </button>
      </div>

      {/* Domain List */}
      <div className="space-y-2">
        {domains.length === 0 ? (
          <div className="text-center py-8 bg-gray-50 dark:bg-gray-800 rounded-lg border-2 border-dashed border-gray-300 dark:border-gray-600">
            <p className="text-sm text-gray-500 dark:text-gray-400">
              No domains configured yet
            </p>
          </div>
        ) : (
          <div className="space-y-2">
            {domains.map((domain) => (
              <div
                key={domain}
                className="flex items-center justify-between px-4 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-md"
              >
                <span className="text-sm font-medium text-gray-900 dark:text-white">
                  {domain}
                </span>
                <button
                  type="button"
                  onClick={() => handleRemoveDomain(domain)}
                  className="text-sm text-red-600 dark:text-red-400 hover:text-red-800 dark:hover:text-red-300"
                  aria-label={`Remove ${domain}`}
                >
                  Remove
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Info Box */}
      <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-md p-4">
        <h4 className="text-sm font-medium text-blue-900 dark:text-blue-300 mb-2">
          Domain Mapping
        </h4>
        <ul className="text-sm text-blue-800 dark:text-blue-400 space-y-1 list-disc list-inside">
          <li>Enter only the domain part (e.g., example.com, not user@example.com)</li>
          <li>Users logging in with matching email domains will use this SSO provider</li>
          <li>Multiple domains can be mapped to a single provider</li>
        </ul>
      </div>
    </div>
  );
};
