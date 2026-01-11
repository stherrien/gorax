/**
 * Tests for SSOProviderCard component
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SSOProviderCard } from './SSOProviderCard';
import type { SSOProvider } from '../../types/sso';

const mockProvider: SSOProvider = {
  id: 'provider-123',
  tenant_id: 'tenant-123',
  name: 'Okta SSO',
  provider_type: 'saml',
  enabled: true,
  enforce_sso: false,
  config: {
    entity_id: 'https://app.gorax.com',
    acs_url: 'https://app.gorax.com/api/v1/sso/acs',
    idp_entity_id: 'https://okta.com/entity',
    idp_sso_url: 'https://okta.com/sso',
    attribute_mapping: {},
    sign_authn_requests: false,
  },
  domains: ['example.com', 'test.com'],
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

describe('SSOProviderCard', () => {
  it('should render provider information', () => {
    render(
      <SSOProviderCard
        provider={mockProvider}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onTest={vi.fn()}
        onToggle={vi.fn()}
      />
    );

    expect(screen.getByText('Okta SSO')).toBeInTheDocument();
    expect(screen.getByText('SAML')).toBeInTheDocument();
    expect(screen.getByText('example.com, test.com')).toBeInTheDocument();
  });

  it('should show enabled status badge', () => {
    render(
      <SSOProviderCard
        provider={mockProvider}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onTest={vi.fn()}
        onToggle={vi.fn()}
      />
    );

    expect(screen.getByText('Enabled')).toBeInTheDocument();
  });

  it('should show disabled status badge', () => {
    const disabledProvider = { ...mockProvider, enabled: false };
    render(
      <SSOProviderCard
        provider={disabledProvider}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onTest={vi.fn()}
        onToggle={vi.fn()}
      />
    );

    expect(screen.getByText('Disabled')).toBeInTheDocument();
  });

  it('should call onEdit when edit button is clicked', async () => {
    const onEdit = vi.fn();
    render(
      <SSOProviderCard
        provider={mockProvider}
        onEdit={onEdit}
        onDelete={vi.fn()}
        onTest={vi.fn()}
        onToggle={vi.fn()}
      />
    );

    const editButton = screen.getByRole('button', { name: /edit/i });
    await userEvent.click(editButton);

    expect(onEdit).toHaveBeenCalledWith(mockProvider.id);
  });

  it('should call onDelete when delete button is clicked', async () => {
    const onDelete = vi.fn();
    render(
      <SSOProviderCard
        provider={mockProvider}
        onEdit={vi.fn()}
        onDelete={onDelete}
        onTest={vi.fn()}
        onToggle={vi.fn()}
      />
    );

    const deleteButton = screen.getByRole('button', { name: /delete/i });
    await userEvent.click(deleteButton);

    expect(onDelete).toHaveBeenCalledWith(mockProvider.id);
  });

  it('should call onTest when test button is clicked', async () => {
    const onTest = vi.fn();
    render(
      <SSOProviderCard
        provider={mockProvider}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onTest={onTest}
        onToggle={vi.fn()}
      />
    );

    const testButton = screen.getByRole('button', { name: /test/i });
    await userEvent.click(testButton);

    expect(onTest).toHaveBeenCalledWith(mockProvider.id);
  });

  it('should call onToggle when toggle switch is clicked', async () => {
    const onToggle = vi.fn();
    render(
      <SSOProviderCard
        provider={mockProvider}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onTest={vi.fn()}
        onToggle={onToggle}
      />
    );

    const toggleSwitch = screen.getByRole('switch');
    await userEvent.click(toggleSwitch);

    expect(onToggle).toHaveBeenCalledWith(mockProvider.id, false);
  });

  it('should display OIDC badge for OIDC provider', () => {
    const oidcProvider = { ...mockProvider, provider_type: 'oidc' as const };
    render(
      <SSOProviderCard
        provider={oidcProvider}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onTest={vi.fn()}
        onToggle={vi.fn()}
      />
    );

    expect(screen.getByText('OIDC')).toBeInTheDocument();
  });

  it('should show enforcement badge when SSO is enforced', () => {
    const enforcedProvider = { ...mockProvider, enforce_sso: true };
    render(
      <SSOProviderCard
        provider={enforcedProvider}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onTest={vi.fn()}
        onToggle={vi.fn()}
      />
    );

    expect(screen.getByText('Required')).toBeInTheDocument();
  });
});
