/**
 * SSO Test Panel Component
 * Tests SSO connection and displays results
 */

import React, { useState } from 'react';
import { useTestSSOProvider } from '../../hooks/useSSO';

interface SSOTestPanelProps {
  providerId: string;
  providerName: string;
  enabled: boolean;
}

export const SSOTestPanel: React.FC<SSOTestPanelProps> = ({
  providerId,
  providerName,
  enabled,
}) => {
  const testMutation = useTestSSOProvider();
  const [showWarning, setShowWarning] = useState(false);

  const handleTest = () => {
    if (!enabled) {
      setShowWarning(true);
      return;
    }
    testMutation.mutate(providerId);
  };

  return (
    <div className="space-y-4">
      {/* Header */}
      <div>
        <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
          Test SSO Connection
        </h3>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          Test the SSO flow to ensure everything is configured correctly
        </p>
      </div>

      {/* Warning for disabled provider */}
      {!enabled && showWarning && (
        <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-md p-4">
          <p className="text-sm text-amber-800 dark:text-amber-400">
            This provider is currently disabled. Enable it before testing.
          </p>
        </div>
      )}

      {/* Test Button */}
      <div className="flex items-center gap-4">
        <button
          onClick={handleTest}
          disabled={!enabled}
          className="px-6 py-3 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed rounded-md transition-colors"
        >
          Test SSO Login
        </button>
        {!enabled && (
          <span className="text-sm text-gray-500 dark:text-gray-400">
            Provider must be enabled to test
          </span>
        )}
      </div>

      {/* Instructions */}
      <div className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-md p-4">
        <h4 className="text-sm font-medium text-gray-900 dark:text-white mb-3">
          Testing Process
        </h4>
        <ol className="text-sm text-gray-700 dark:text-gray-300 space-y-2 list-decimal list-inside">
          <li>
            Click "Test SSO Login" to open the SSO login flow in a new tab
          </li>
          <li>
            You'll be redirected to {providerName} to authenticate
          </li>
          <li>
            After successful authentication, you'll be redirected back to Gorax
          </li>
          <li>
            If the test succeeds, you'll see a success message
          </li>
          <li>
            If it fails, check the error message and verify your configuration
          </li>
        </ol>
      </div>

      {/* Troubleshooting */}
      <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-md p-4">
        <h4 className="text-sm font-medium text-blue-900 dark:text-blue-300 mb-2">
          Common Issues
        </h4>
        <ul className="text-sm text-blue-800 dark:text-blue-400 space-y-1 list-disc list-inside">
          <li>Verify your IdP metadata is correct</li>
          <li>Check that attribute mappings match your IdP configuration</li>
          <li>Ensure the redirect URL is configured in your IdP</li>
          <li>Verify your domain mappings are correct</li>
          <li>Check that your IdP is accessible from this server</li>
        </ul>
      </div>

      {/* Security Note */}
      <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-md p-4">
        <h4 className="text-sm font-medium text-amber-900 dark:text-amber-300 mb-2">
          Security Note
        </h4>
        <p className="text-sm text-amber-800 dark:text-amber-400">
          Always test SSO configuration thoroughly before enabling for production use.
          Ensure all certificate validations and security checks are in place.
        </p>
      </div>
    </div>
  );
};
