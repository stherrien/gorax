/**
 * SSO Provider Card Component
 * Displays a single SSO provider with actions
 */

import React from 'react';
import type { SSOProvider } from '../../types/sso';

interface SSOProviderCardProps {
  provider: SSOProvider;
  onEdit: (id: string) => void;
  onDelete: (id: string) => void;
  onTest: (id: string) => void;
  onToggle: (id: string, enabled: boolean) => void;
}

export const SSOProviderCard: React.FC<SSOProviderCardProps> = ({
  provider,
  onEdit,
  onDelete,
  onTest,
  onToggle,
}) => {
  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 shadow-sm hover:shadow-md transition-shadow">
      {/* Header */}
      <div className="flex items-start justify-between mb-4">
        <div className="flex-1">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
            {provider.name}
          </h3>
          <div className="flex flex-wrap gap-2">
            {/* Provider Type Badge */}
            <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
              {provider.provider_type.toUpperCase()}
            </span>

            {/* Status Badge */}
            <span
              className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                provider.enabled
                  ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                  : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
              }`}
            >
              {provider.enabled ? 'Enabled' : 'Disabled'}
            </span>

            {/* Enforcement Badge */}
            {provider.enforce_sso && (
              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200">
                Required
              </span>
            )}
          </div>
        </div>

        {/* Toggle Switch */}
        <label className="relative inline-flex items-center cursor-pointer">
          <input
            type="checkbox"
            role="switch"
            className="sr-only peer"
            checked={provider.enabled}
            onChange={(e) => onToggle(provider.id, e.target.checked)}
          />
          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
        </label>
      </div>

      {/* Domains */}
      <div className="mb-4">
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-1">
          Email Domains:
        </p>
        <p className="text-sm text-gray-900 dark:text-white font-medium">
          {provider.domains.join(', ')}
        </p>
      </div>

      {/* Timestamps */}
      <div className="mb-4 text-xs text-gray-500 dark:text-gray-500">
        <p>
          Created: {new Date(provider.created_at).toLocaleDateString()}
        </p>
        <p>
          Updated: {new Date(provider.updated_at).toLocaleDateString()}
        </p>
      </div>

      {/* Actions */}
      <div className="flex gap-2 pt-4 border-t border-gray-200 dark:border-gray-700">
        <button
          onClick={() => onTest(provider.id)}
          className="flex-1 px-3 py-2 text-sm font-medium text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20 hover:bg-blue-100 dark:hover:bg-blue-900/40 rounded-md transition-colors"
          aria-label="Test connection"
        >
          Test
        </button>
        <button
          onClick={() => onEdit(provider.id)}
          className="flex-1 px-3 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-md transition-colors"
          aria-label="Edit provider"
        >
          Edit
        </button>
        <button
          onClick={() => onDelete(provider.id)}
          className="flex-1 px-3 py-2 text-sm font-medium text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 hover:bg-red-100 dark:hover:bg-red-900/40 rounded-md transition-colors"
          aria-label="Delete provider"
        >
          Delete
        </button>
      </div>
    </div>
  );
};
