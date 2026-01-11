/**
 * Tests for AttributeMappingBuilder component
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AttributeMappingBuilder } from './AttributeMappingBuilder';

const defaultMapping = {
  email: 'NameID',
  first_name: 'firstName',
  last_name: 'lastName',
  groups: 'groups',
};

describe('AttributeMappingBuilder', () => {
  it('should render all attribute fields', () => {
    render(
      <AttributeMappingBuilder
        mapping={defaultMapping}
        onChange={vi.fn()}
        providerType="saml"
      />
    );

    expect(screen.getByLabelText(/email attribute/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/first name attribute/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/last name attribute/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/groups attribute/i)).toBeInTheDocument();
  });

  it('should display current mappings', () => {
    render(
      <AttributeMappingBuilder
        mapping={defaultMapping}
        onChange={vi.fn()}
        providerType="saml"
      />
    );

    expect(screen.getByDisplayValue('NameID')).toBeInTheDocument();
    expect(screen.getByDisplayValue('firstName')).toBeInTheDocument();
  });

  it('should call onChange when mapping is updated', async () => {
    const onChange = vi.fn();
    render(
      <AttributeMappingBuilder
        mapping={{}}
        onChange={onChange}
        providerType="saml"
      />
    );

    const emailInput = screen.getByLabelText(/email attribute/i);
    await userEvent.type(emailInput, 'x');

    // Should be called when user types
    expect(onChange).toHaveBeenCalled();
    // Should pass email in the mapping
    expect(onChange.mock.calls[0][0]).toHaveProperty('email');
  });

  it('should show SAML examples', () => {
    render(
      <AttributeMappingBuilder
        mapping={defaultMapping}
        onChange={vi.fn()}
        providerType="saml"
      />
    );

    expect(screen.getByText(/Map SAML attributes/i)).toBeInTheDocument();
  });

  it('should show OIDC examples', () => {
    render(
      <AttributeMappingBuilder
        mapping={{}}
        onChange={vi.fn()}
        providerType="oidc"
      />
    );

    expect(screen.getByText(/Map OIDC attributes/i)).toBeInTheDocument();
  });

  it('should handle empty mapping', () => {
    render(
      <AttributeMappingBuilder
        mapping={{}}
        onChange={vi.fn()}
        providerType="saml"
      />
    );

    const emailInput = screen.getByLabelText(/email attribute/i) as HTMLInputElement;
    expect(emailInput.value).toBe('');
  });
});
