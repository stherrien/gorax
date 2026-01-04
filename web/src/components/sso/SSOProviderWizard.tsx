/**
 * SSO Provider Wizard Component
 * Multi-step wizard for creating/editing SSO providers
 */

import React, { useState, useEffect } from 'react';
import { useCreateSSOProvider, useUpdateSSOProvider } from '../../hooks/useSSO';
import { SAMLConfigForm } from './SAMLConfigForm';
import { OIDCConfigForm } from './OIDCConfigForm';
import { DomainMappingEditor } from './DomainMappingEditor';
import { SSOTestPanel } from './SSOTestPanel';
import type { SSOProvider, ProviderType, SAMLConfig, OIDCConfig } from '../../types/sso';

interface SSOProviderWizardProps {
  provider?: SSOProvider | null;
  onComplete: () => void;
  onCancel: () => void;
}

type WizardStep = 'type' | 'config' | 'domains' | 'test' | 'review';

export const SSOProviderWizard: React.FC<SSOProviderWizardProps> = ({
  provider,
  onComplete,
  onCancel,
}) => {
  const [currentStep, setCurrentStep] = useState<WizardStep>('type');
  const [providerType, setProviderType] = useState<ProviderType>(
    provider?.provider_type || 'saml'
  );
  const [providerName, setProviderName] = useState(provider?.name || '');
  const [config, setConfig] = useState<Partial<SAMLConfig | OIDCConfig>>(
    provider?.config || {}
  );
  const [domains, setDomains] = useState<string[]>(provider?.domains || []);
  const [enabled, setEnabled] = useState(provider?.enabled ?? false);
  const [enforceSso, setEnforceSso] = useState(provider?.enforce_sso ?? false);
  const [error, setError] = useState('');

  const createMutation = useCreateSSOProvider();
  const updateMutation = useUpdateSSOProvider();

  const baseUrl = window.location.origin;
  const isEditMode = !!provider;

  // Skip type selection in edit mode
  useEffect(() => {
    if (isEditMode) {
      setCurrentStep('config');
    }
  }, [isEditMode]);

  const validateStep = (): boolean => {
    setError('');

    switch (currentStep) {
      case 'type':
        if (!providerName.trim()) {
          setError('Provider name is required');
          return false;
        }
        return true;

      case 'config':
        if (providerType === 'saml') {
          const samlConfig = config as Partial<SAMLConfig>;
          if (!samlConfig.idp_entity_id || !samlConfig.idp_sso_url) {
            setError('IdP Entity ID and SSO URL are required');
            return false;
          }
        } else {
          const oidcConfig = config as Partial<OIDCConfig>;
          if (!oidcConfig.client_id || !oidcConfig.client_secret) {
            setError('Client ID and Client Secret are required');
            return false;
          }
          if (!oidcConfig.discovery_url && !oidcConfig.authorization_url) {
            setError('Discovery URL or Authorization URL is required');
            return false;
          }
        }
        return true;

      case 'domains':
        if (domains.length === 0) {
          setError('At least one domain is required');
          return false;
        }
        return true;

      default:
        return true;
    }
  };

  const handleNext = () => {
    if (!validateStep()) {
      return;
    }

    const steps: WizardStep[] = ['type', 'config', 'domains', 'test', 'review'];
    const currentIndex = steps.indexOf(currentStep);
    if (currentIndex < steps.length - 1) {
      setCurrentStep(steps[currentIndex + 1]);
    }
  };

  const handleBack = () => {
    const steps: WizardStep[] = ['type', 'config', 'domains', 'test', 'review'];
    const currentIndex = steps.indexOf(currentStep);
    if (currentIndex > 0) {
      setCurrentStep(steps[currentIndex - 1]);
    }
  };

  const handleSave = async () => {
    if (!validateStep()) {
      return;
    }

    try {
      if (isEditMode && provider) {
        await updateMutation.mutateAsync({
          providerId: provider.id,
          updates: {
            name: providerName,
            config: config as SAMLConfig | OIDCConfig,
            domains,
            enabled,
            enforce_sso: enforceSso,
          },
        });
      } else {
        await createMutation.mutateAsync({
          name: providerName,
          provider_type: providerType,
          config: config as SAMLConfig | OIDCConfig,
          domains,
          enabled,
          enforce_sso: enforceSso,
        });
      }
      onComplete();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save provider');
    }
  };

  const renderStepContent = () => {
    switch (currentStep) {
      case 'type':
        return (
          <div className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Provider Name <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={providerName}
                onChange={(e) => setProviderName(e.target.value)}
                placeholder="e.g., Okta SSO, Azure AD"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
                Provider Type
              </label>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <button
                  type="button"
                  onClick={() => setProviderType('saml')}
                  className={`p-6 border-2 rounded-lg text-left transition-all ${
                    providerType === 'saml'
                      ? 'border-blue-600 bg-blue-50 dark:bg-blue-900/20'
                      : 'border-gray-300 dark:border-gray-600 hover:border-gray-400'
                  }`}
                >
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
                    SAML 2.0
                  </h3>
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    Standard enterprise SSO protocol. Compatible with Okta, Azure AD,
                    and most enterprise IdPs.
                  </p>
                </button>

                <button
                  type="button"
                  onClick={() => setProviderType('oidc')}
                  className={`p-6 border-2 rounded-lg text-left transition-all ${
                    providerType === 'oidc'
                      ? 'border-blue-600 bg-blue-50 dark:bg-blue-900/20'
                      : 'border-gray-300 dark:border-gray-600 hover:border-gray-400'
                  }`}
                >
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
                    OIDC (OpenID Connect)
                  </h3>
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    Modern OAuth 2.0-based protocol. Works with Google, Auth0, and
                    modern IdPs.
                  </p>
                </button>
              </div>
            </div>
          </div>
        );

      case 'config':
        return providerType === 'saml' ? (
          <SAMLConfigForm
            config={config as Partial<SAMLConfig>}
            onChange={setConfig}
            baseUrl={baseUrl}
          />
        ) : (
          <OIDCConfigForm
            config={config as Partial<OIDCConfig>}
            onChange={setConfig}
            providerId={provider?.id}
            baseUrl={baseUrl}
          />
        );

      case 'domains':
        return <DomainMappingEditor domains={domains} onChange={setDomains} />;

      case 'test':
        return provider ? (
          <SSOTestPanel
            providerId={provider.id}
            providerName={providerName}
            enabled={enabled}
          />
        ) : (
          <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg p-4">
            <p className="text-sm text-amber-800 dark:text-amber-400">
              Save the provider first to test the SSO connection
            </p>
          </div>
        );

      case 'review':
        return (
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
                Review Configuration
              </h3>
            </div>

            <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 space-y-4">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Provider Name</p>
                <p className="text-base font-medium text-gray-900 dark:text-white">
                  {providerName}
                </p>
              </div>
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Type</p>
                <p className="text-base font-medium text-gray-900 dark:text-white">
                  {providerType.toUpperCase()}
                </p>
              </div>
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Domains</p>
                <p className="text-base font-medium text-gray-900 dark:text-white">
                  {domains.join(', ')}
                </p>
              </div>
            </div>

            <div className="space-y-3">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={enabled}
                  onChange={(e) => setEnabled(e.target.checked)}
                  className="mr-2"
                />
                <span className="text-sm text-gray-700 dark:text-gray-300">
                  Enable this provider immediately
                </span>
              </label>
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={enforceSso}
                  onChange={(e) => setEnforceSso(e.target.checked)}
                  className="mr-2"
                />
                <span className="text-sm text-gray-700 dark:text-gray-300">
                  Require SSO (disable password login for these domains)
                </span>
              </label>
            </div>
          </div>
        );

      default:
        return null;
    }
  };

  const getStepTitle = () => {
    switch (currentStep) {
      case 'type':
        return 'Choose Provider Type';
      case 'config':
        return `Configure ${providerType.toUpperCase()}`;
      case 'domains':
        return 'Domain Mapping';
      case 'test':
        return 'Test Connection';
      case 'review':
        return 'Review & Confirm';
      default:
        return '';
    }
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 max-w-4xl mx-auto">
      {/* Progress Indicator */}
      <div className="mb-8">
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
            {isEditMode ? 'Edit' : 'Add'} SSO Provider
          </h2>
        </div>
        <h3 className="text-lg text-gray-700 dark:text-gray-300">{getStepTitle()}</h3>
      </div>

      {/* Error Message */}
      {error && (
        <div className="mb-6 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
          <p className="text-sm text-red-800 dark:text-red-400">{error}</p>
        </div>
      )}

      {/* Step Content */}
      <div className="mb-8">{renderStepContent()}</div>

      {/* Actions */}
      <div className="flex justify-between border-t border-gray-200 dark:border-gray-700 pt-6">
        <button
          onClick={onCancel}
          className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-md transition-colors"
        >
          Cancel
        </button>

        <div className="flex gap-2">
          {currentStep !== 'type' && (
            <button
              onClick={handleBack}
              className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-md transition-colors"
            >
              Back
            </button>
          )}

          {currentStep === 'review' ? (
            <button
              onClick={handleSave}
              disabled={createMutation.isPending || updateMutation.isPending}
              className="px-6 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 disabled:cursor-not-allowed rounded-md transition-colors"
            >
              {createMutation.isPending || updateMutation.isPending
                ? 'Saving...'
                : isEditMode
                ? 'Update Provider'
                : 'Create Provider'}
            </button>
          ) : (
            <button
              onClick={handleNext}
              className="px-6 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md transition-colors"
            >
              Next
            </button>
          )}
        </div>
      </div>
    </div>
  );
};
