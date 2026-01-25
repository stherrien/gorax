/**
 * SSO Metadata Display Component
 * Displays SAML SP metadata for copying to IdP
 */

import React, { useState } from 'react';
import { useSSOMetadata } from '../../hooks/useSSO';

interface SSOMetadataDisplayProps {
  providerId: string;
}

export const SSOMetadataDisplay: React.FC<SSOMetadataDisplayProps> = ({
  providerId,
}) => {
  const { data: metadata, isLoading, error } = useSSOMetadata(providerId);
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    if (metadata) {
      await navigator.clipboard.writeText(metadata);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleDownload = () => {
    if (metadata) {
      const blob = new Blob([metadata], { type: 'application/xml' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'metadata.xml';
      a.click();
      URL.revokeObjectURL(url);
    }
  };

  if (isLoading) {
    return (
      <div className="text-center py-4">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
        <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
          Loading metadata...
        </p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
        <p className="text-sm text-red-800 dark:text-red-400">
          Failed to load metadata
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <div>
        <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
          Service Provider Metadata
        </h3>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          Copy this metadata or download the XML file to configure your IdP
        </p>
      </div>

      {/* Actions */}
      <div className="flex gap-2">
        <button
          onClick={handleCopy}
          className="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md transition-colors"
        >
          {copied ? 'Copied!' : 'Copy to Clipboard'}
        </button>
        <button
          onClick={handleDownload}
          className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-md transition-colors"
        >
          Download metadata.xml
        </button>
      </div>

      {/* Metadata Display */}
      <div className="relative">
        <pre className="bg-gray-900 text-gray-100 p-4 rounded-md overflow-x-auto text-xs font-mono max-h-96">
          <code>{metadata}</code>
        </pre>
      </div>

      {/* Instructions */}
      <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-md p-4">
        <h4 className="text-sm font-medium text-blue-900 dark:text-blue-300 mb-2">
          How to use this metadata
        </h4>
        <ol className="text-sm text-blue-800 dark:text-blue-400 space-y-1 list-decimal list-inside">
          <li>Copy the metadata or download the XML file</li>
          <li>Go to your Identity Provider configuration</li>
          <li>Find the section for adding a new SAML application</li>
          <li>Paste the metadata or upload the XML file</li>
          <li>Complete the IdP configuration and save</li>
        </ol>
      </div>
    </div>
  );
};
