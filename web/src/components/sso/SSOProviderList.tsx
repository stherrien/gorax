/**
 * SSO Provider List Component
 * Displays all configured SSO providers
 */

import React from 'react';
import { SSOProviderCard } from './SSOProviderCard';
import type { SSOProvider } from '../../types/sso';

interface SSOProviderListProps {
  providers: SSOProvider[];
  loading?: boolean;
  onEdit: (id: string) => void;
  onDelete: (id: string) => void;
  onTest: (id: string) => void;
  onToggle: (id: string, enabled: boolean) => void;
  onAdd: () => void;
}

export const SSOProviderList: React.FC<SSOProviderListProps> = ({
  providers,
  loading,
  onEdit,
  onDelete,
  onTest,
  onToggle,
  onAdd,
}) => {
  if (loading) {
    return (
      <div className="text-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
        <p className="mt-4 text-gray-500 dark:text-gray-400">
          Loading SSO providers...
        </p>
      </div>
    );
  }

  if (providers.length === 0) {
    return (
      <div className="text-center py-12 bg-gray-50 dark:bg-gray-800 rounded-lg border-2 border-dashed border-gray-300 dark:border-gray-600">
        <svg
          className="mx-auto h-12 w-12 text-gray-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
          />
        </svg>
        <h3 className="mt-4 text-lg font-medium text-gray-900 dark:text-white">
          No SSO Providers
        </h3>
        <p className="mt-2 text-sm text-gray-500 dark:text-gray-400 max-w-sm mx-auto">
          Get started by configuring your first SSO provider. Enable enterprise
          authentication with SAML or OIDC.
        </p>
        <button
          onClick={onAdd}
          className="mt-6 px-6 py-3 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md transition-colors"
        >
          Add SSO Provider
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
            SSO Providers
          </h2>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {providers.length} provider{providers.length !== 1 ? 's' : ''} configured
          </p>
        </div>
        <button
          onClick={onAdd}
          className="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md transition-colors"
        >
          Add Provider
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
          <p className="text-sm text-gray-500 dark:text-gray-400">Total Providers</p>
          <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">
            {providers.length}
          </p>
        </div>
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
          <p className="text-sm text-gray-500 dark:text-gray-400">Enabled</p>
          <p className="text-2xl font-bold text-green-600 dark:text-green-400 mt-1">
            {providers.filter((p) => p.enabled).length}
          </p>
        </div>
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
          <p className="text-sm text-gray-500 dark:text-gray-400">Total Domains</p>
          <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">
            {providers.reduce((sum, p) => sum + p.domains.length, 0)}
          </p>
        </div>
      </div>

      {/* Provider Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {providers.map((provider) => (
          <SSOProviderCard
            key={provider.id}
            provider={provider}
            onEdit={onEdit}
            onDelete={onDelete}
            onTest={onTest}
            onToggle={onToggle}
          />
        ))}
      </div>
    </div>
  );
};
