/**
 * SSO Settings Page
 * Main admin page for managing SSO providers
 */

import React, { useState } from 'react';
import {
  useSSOProviders,
  useDeleteSSOProvider,
  useToggleProviderStatus,
  useTestSSOProvider,
} from '../../hooks/useSSO';
import { SSOProviderList } from '../../components/sso/SSOProviderList';
import { SSOProviderWizard } from '../../components/sso/SSOProviderWizard';
import type { SSOProvider } from '../../types/sso';

type ViewMode = 'list' | 'add' | 'edit';

export const SSOSettings: React.FC = () => {
  const [viewMode, setViewMode] = useState<ViewMode>('list');
  const [editingProvider, setEditingProvider] = useState<SSOProvider | null>(null);
  const [deletingProvider, setDeletingProvider] = useState<SSOProvider | null>(null);

  const { data: providers, isLoading } = useSSOProviders();
  const deleteMutation = useDeleteSSOProvider();
  const toggleMutation = useToggleProviderStatus();
  const testMutation = useTestSSOProvider();

  const handleAdd = () => {
    setEditingProvider(null);
    setViewMode('add');
  };

  const handleEdit = (id: string) => {
    const provider = providers?.find((p) => p.id === id);
    if (provider) {
      setEditingProvider(provider);
      setViewMode('edit');
    }
  };

  const handleDelete = (id: string) => {
    const provider = providers?.find((p) => p.id === id);
    if (provider) {
      setDeletingProvider(provider);
    }
  };

  const confirmDelete = async () => {
    if (deletingProvider) {
      await deleteMutation.mutateAsync(deletingProvider.id);
      setDeletingProvider(null);
    }
  };

  const handleTest = (id: string) => {
    testMutation.mutate(id);
  };

  const handleToggle = async (id: string, enabled: boolean) => {
    await toggleMutation.mutateAsync({ providerId: id, enabled });
  };

  const handleWizardComplete = () => {
    setViewMode('list');
    setEditingProvider(null);
  };

  const handleWizardCancel = () => {
    setViewMode('list');
    setEditingProvider(null);
  };

  return (
    <div className="container mx-auto px-4 py-8 max-w-7xl">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          SSO Configuration
        </h1>
        <p className="mt-2 text-gray-600 dark:text-gray-400">
          Configure Single Sign-On providers for enterprise authentication
        </p>
      </div>

      {/* Security Notice */}
      <div className="mb-6 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
        <div className="flex items-start">
          <svg
            className="h-5 w-5 text-blue-600 dark:text-blue-400 mt-0.5 mr-3"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <div>
            <h3 className="text-sm font-medium text-blue-900 dark:text-blue-300">
              Security Best Practices
            </h3>
            <ul className="mt-2 text-sm text-blue-800 dark:text-blue-400 list-disc list-inside space-y-1">
              <li>Always use HTTPS for SSO endpoints</li>
              <li>Verify IdP certificates and metadata</li>
              <li>Test thoroughly before enabling for production</li>
              <li>Keep client secrets secure and rotate regularly</li>
              <li>Monitor SSO login events for suspicious activity</li>
            </ul>
          </div>
        </div>
      </div>

      {/* Main Content */}
      {viewMode === 'list' ? (
        <SSOProviderList
          providers={providers || []}
          loading={isLoading}
          onAdd={handleAdd}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onTest={handleTest}
          onToggle={handleToggle}
        />
      ) : (
        <SSOProviderWizard
          provider={editingProvider}
          onComplete={handleWizardComplete}
          onCancel={handleWizardCancel}
        />
      )}

      {/* Delete Confirmation Modal */}
      {deletingProvider && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              Delete SSO Provider
            </h3>
            <p className="text-gray-600 dark:text-gray-400 mb-6">
              Are you sure you want to delete "{deletingProvider.name}"? This
              action cannot be undone. Users from domains mapped to this provider
              will no longer be able to sign in via SSO.
            </p>
            <div className="flex gap-3 justify-end">
              <button
                onClick={() => setDeletingProvider(null)}
                className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-md transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={confirmDelete}
                disabled={deleteMutation.isPending}
                className="px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 disabled:bg-red-400 disabled:cursor-not-allowed rounded-md transition-colors"
              >
                {deleteMutation.isPending ? 'Deleting...' : 'Delete Provider'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
