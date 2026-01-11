# Enterprise SSO Setup Guide

## Overview

Gorax supports Enterprise Single Sign-On (SSO) integration with both SAML 2.0 and OpenID Connect (OIDC) protocols. This allows organizations to integrate with identity providers like Okta, Azure AD, Google Workspace, OneLogin, and Auth0.

## Features

- **SAML 2.0 Support**: Full SAML 2.0 implementation with signature verification
- **OIDC Support**: OpenID Connect with automatic discovery
- **Multi-Tenancy**: Different SSO providers per tenant
- **JIT Provisioning**: Automatic user creation on first SSO login
- **Attribute Mapping**: Flexible mapping of IdP attributes to user fields
- **Domain-Based Detection**: Automatic SSO provider detection by email domain
- **SSO Enforcement**: Optional requirement for SSO login per tenant
- **Audit Logging**: Complete audit trail of SSO login attempts

## Architecture

### Components

1. **SSO Providers**: SAML and OIDC provider implementations
2. **Repository**: Database operations for providers, connections, and events
3. **Service**: Business logic with JIT provisioning
4. **API Handlers**: HTTP endpoints for SSO management and authentication
5. **Frontend**: Admin UI for SSO configuration

### Authentication Flow

#### SAML 2.0 Flow
```
User → Gorax SP → IdP (Login) → IdP (Authenticate) → Gorax ACS → Session Created
```

#### OIDC Flow
```
User → Gorax RP → IdP (Authorize) → IdP (Authenticate) → Gorax Callback → Token Exchange → Session Created
```

## Configuration

### Environment Variables

```bash
# SSO Configuration
SSO_ENABLED=true
SSO_CALLBACK_BASE_URL=https://app.gorax.io
SSO_SESSION_DURATION=24h
```

### Database Schema

The SSO system uses three main tables:

- **sso_providers**: SSO provider configurations
- **sso_connections**: User-to-SSO mappings
- **sso_login_events**: Audit log of login attempts

## API Endpoints

### Provider Management (Authenticated)

```
POST   /api/v1/sso/providers          Create SSO provider
GET    /api/v1/sso/providers          List providers
GET    /api/v1/sso/providers/:id      Get provider
PUT    /api/v1/sso/providers/:id      Update provider
DELETE /api/v1/sso/providers/:id      Delete provider
```

### SSO Authentication (Public)

```
GET    /api/v1/sso/login/:id          Initiate SSO login
GET    /api/v1/sso/callback/:id       OIDC callback
POST   /api/v1/sso/callback/:id       SAML callback
POST   /api/v1/sso/acs                SAML assertion consumer
GET    /api/v1/sso/metadata/:id       Get SAML SP metadata
GET    /api/v1/sso/discover           Discover provider by email
```

## Provider Setup Guides

### [Okta SAML Setup](./SSO_OKTA_SAML.md)
### [Azure AD SAML Setup](./SSO_AZURE_SAML.md)
### [Google Workspace SAML Setup](./SSO_GOOGLE_SAML.md)
### [Okta OIDC Setup](./SSO_OKTA_OIDC.md)
### [Azure AD OIDC Setup](./SSO_AZURE_OIDC.md)

## Quick Start

### 1. Create SAML Provider

```bash
curl -X POST https://app.gorax.io/api/v1/sso/providers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Okta SAML",
    "provider_type": "saml",
    "enabled": true,
    "enforce_sso": false,
    "domains": ["company.com"],
    "config": {
      "entity_id": "https://app.gorax.io",
      "acs_url": "https://app.gorax.io/api/v1/sso/acs",
      "idp_metadata_url": "https://company.okta.com/app/metadata",
      "idp_entity_id": "http://www.okta.com/...",
      "idp_sso_url": "https://company.okta.com/app/...",
      "attribute_mapping": {
        "email": "NameID",
        "first_name": "firstName",
        "last_name": "lastName",
        "groups": "groups"
      }
    }
  }'
```

### 2. Create OIDC Provider

```bash
curl -X POST https://app.gorax.io/api/v1/sso/providers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Azure AD OIDC",
    "provider_type": "oidc",
    "enabled": true,
    "enforce_sso": false,
    "domains": ["company.com"],
    "config": {
      "client_id": "your-client-id",
      "client_secret": "your-client-secret",
      "discovery_url": "https://login.microsoftonline.com/{tenant}/v2.0/.well-known/openid-configuration",
      "redirect_url": "https://app.gorax.io/api/v1/sso/callback/{provider-id}",
      "scopes": ["openid", "profile", "email"],
      "attribute_mapping": {
        "email": "email",
        "first_name": "given_name",
        "last_name": "family_name",
        "groups": "groups"
      }
    }
  }'
```

### 3. Test SSO Login

1. Navigate to: `https://app.gorax.io/login`
2. Enter email: `user@company.com`
3. System detects SSO provider
4. Redirects to IdP for authentication
5. After authentication, redirects back to Gorax
6. User is automatically provisioned (JIT)
7. Session created and user logged in

## Attribute Mapping

Map IdP attributes to Gorax user fields:

| Gorax Field | Purpose | SAML Attribute | OIDC Claim |
|-------------|---------|----------------|------------|
| `email` | User email (required) | `NameID` or custom | `email` |
| `first_name` | First name | `firstName` | `given_name` |
| `last_name` | Last name | `lastName` | `family_name` |
| `groups` | User groups/roles | `groups` | `groups` |

### Example SAML Attribute Mapping

```json
{
  "attribute_mapping": {
    "email": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
    "first_name": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
    "last_name": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
    "groups": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/groups"
  }
}
```

### Example OIDC Attribute Mapping

```json
{
  "attribute_mapping": {
    "email": "email",
    "first_name": "given_name",
    "last_name": "family_name",
    "groups": "groups"
  }
}
```

## JIT User Provisioning

When a user logs in via SSO for the first time:

1. **User Lookup**: Check if user exists by email
2. **User Creation**: If not exists, create new user with:
   - Email from SSO attributes
   - Default role: `member`
   - Status: `active`
   - Kratos identity: `sso-{uuid}`
3. **Connection Creation**: Link user to SSO provider
4. **Attribute Storage**: Store SSO attributes for future use
5. **Session Creation**: Create authenticated session

## Security Considerations

### SAML Security

- ✅ Signature verification on assertions
- ✅ Audience validation
- ✅ Time-based validation (NotBefore/NotOnOrAfter)
- ✅ Replay attack prevention (assertion IDs tracked)
- ✅ HTTPS enforcement for endpoints

### OIDC Security

- ✅ State parameter for CSRF protection
- ✅ JWT signature verification
- ✅ Issuer validation
- ✅ Audience validation
- ✅ Nonce validation
- ✅ HTTPS enforcement for callbacks

### Best Practices

1. **Certificate Management**: Rotate SAML certificates regularly
2. **Secret Protection**: Store client secrets encrypted
3. **Session Duration**: Set appropriate session timeouts
4. **Audit Logging**: Monitor SSO login events
5. **Domain Validation**: Restrict SSO to verified domains
6. **MFA Enforcement**: Require MFA at IdP level

## Troubleshooting

### Common Issues

#### SAML: "Invalid signature"
- Verify certificate in IdP metadata matches
- Check clock synchronization between SP and IdP
- Ensure using correct signing algorithm

#### SAML: "Assertion expired"
- Check server time synchronization
- Adjust NotOnOrAfter duration at IdP

#### OIDC: "Invalid state"
- State token may have expired
- Check Redis/session storage is working
- Try initiating new login flow

#### OIDC: "Token verification failed"
- Verify client ID and secret
- Check discovery URL is accessible
- Ensure JWKS endpoint is reachable

#### JIT Provisioning: "User creation failed"
- Check email attribute is being sent
- Verify attribute mapping configuration
- Review database constraints (unique email)

### Debug Mode

Enable detailed logging:

```bash
SSO_DEBUG=true
LOG_LEVEL=debug
```

### Audit Logs

Query SSO login events:

```sql
SELECT * FROM sso_login_events
WHERE sso_provider_id = 'provider-uuid'
ORDER BY created_at DESC
LIMIT 100;
```

## Advanced Configuration

### Multiple SSO Providers

Support multiple IdPs per tenant:

```json
{
  "domains": ["company.com"],
  "enforce_sso": false
}
```

### SSO Enforcement

Require SSO for all logins:

```json
{
  "enforce_sso": true
}
```

### Role Mapping from Groups

Map IdP groups to Gorax roles:

```go
// In service.go, add group-to-role mapping
func mapGroupsToRole(groups []string) string {
    if contains(groups, "admins") {
        return "admin"
    }
    return "member"
}
```

## Testing

### Unit Tests

```bash
go test ./internal/sso/...
```

### Integration Tests

```bash
go test ./internal/sso/... -tags=integration
```

### Test with Mock IdP

Use SAMLtest.id or oidcdebugger.com for testing.

## Support

For issues or questions:
- GitHub Issues: https://github.com/gorax/gorax/issues
- Documentation: https://docs.gorax.io
- Email: support@gorax.io
