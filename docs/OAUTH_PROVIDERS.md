# OAuth Provider Configuration Guide

This guide provides detailed information on configuring OAuth 2.0 integrations with various third-party services in gorax.

## Table of Contents

- [Overview](#overview)
- [General Configuration](#general-configuration)
- [Supported Providers](#supported-providers)
  - [GitHub](#github)
  - [Google](#google)
  - [Slack](#slack)
  - [Microsoft](#microsoft)
  - [Twitter/X](#twitterx)
  - [LinkedIn](#linkedin)
  - [Salesforce](#salesforce)
  - [Auth0](#auth0)
- [Security Considerations](#security-considerations)
- [Troubleshooting](#troubleshooting)

## Overview

gorax supports OAuth 2.0 authentication with multiple third-party providers. OAuth connections allow users to securely integrate external services with their workflows.

### Features

- **Secure Token Storage**: All access and refresh tokens are encrypted using envelope encryption
- **PKCE Support**: Providers that support PKCE (Proof Key for Code Exchange) use it for enhanced security
- **Automatic Token Refresh**: Expired tokens are automatically refreshed when available
- **Audit Logging**: All OAuth operations are logged for security auditing

## General Configuration

All OAuth providers require:

1. **Base URL**: Set in `.env` as `OAUTH_BASE_URL`
2. **Client ID**: Obtained from the provider's developer portal
3. **Client Secret**: Obtained from the provider's developer portal
4. **Redirect URI**: Must match the registered callback URL: `${OAUTH_BASE_URL}/api/v1/oauth/callback/{provider}`

## Supported Providers

### GitHub

**Use Case**: GitHub repository access, issue management, code management

#### Setup Instructions

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Click "New OAuth App"
3. Fill in application details:
   - **Application name**: gorax (or your custom name)
   - **Homepage URL**: Your gorax instance URL
   - **Authorization callback URL**: `${OAUTH_BASE_URL}/api/v1/oauth/callback/github`
4. Click "Register application"
5. Copy the **Client ID**
6. Generate and copy a new **Client Secret**

#### Configuration

```bash
OAUTH_GITHUB_CLIENT_ID=your_client_id_here
OAUTH_GITHUB_CLIENT_SECRET=your_client_secret_here
```

#### Default Scopes

- `user`: Access user profile information
- `repo`: Access repositories (read and write)

#### Notes

- GitHub OAuth Apps **do not support refresh tokens**
- Access tokens are long-lived
- Does **not** support PKCE

---

### Google

**Use Case**: Google Workspace integration, Gmail, Google Drive, Calendar

#### Setup Instructions

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Navigate to **APIs & Services > Credentials**
4. Click **Create Credentials > OAuth 2.0 Client ID**
5. Configure consent screen if prompted
6. Select **Web application** as application type
7. Add authorized redirect URI: `${OAUTH_BASE_URL}/api/v1/oauth/callback/google`
8. Copy **Client ID** and **Client Secret**

#### Configuration

```bash
OAUTH_GOOGLE_CLIENT_ID=your_client_id_here.apps.googleusercontent.com
OAUTH_GOOGLE_CLIENT_SECRET=your_client_secret_here
```

#### Default Scopes

- `https://www.googleapis.com/auth/userinfo.email`: Access email address
- `https://www.googleapis.com/auth/userinfo.profile`: Access profile information

#### Notes

- **Supports PKCE** for enhanced security
- **Supports refresh tokens** (configured to request offline access)
- Tokens expire after a period of inactivity

---

### Slack

**Use Case**: Slack workspace integration, messaging, notifications

#### Setup Instructions

1. Go to [Slack API Applications](https://api.slack.com/apps)
2. Click **Create New App > From scratch**
3. Enter **App Name** and select your **Slack Workspace**
4. Navigate to **OAuth & Permissions**
5. Add redirect URL: `${OAUTH_BASE_URL}/api/v1/oauth/callback/slack`
6. Under **Scopes**, add required OAuth scopes:
   - `chat:write`: Send messages
   - `channels:read`: View channels
7. Copy **Client ID** and **Client Secret**

#### Configuration

```bash
OAUTH_SLACK_CLIENT_ID=your_client_id_here.apps.sla ckapp.com
OAUTH_SLACK_CLIENT_SECRET=your_client_secret_here
```

#### Default Scopes

- `chat:write`: Send messages as the app
- `channels:read`: View basic channel information

#### Notes

- **Supports PKCE**
- **Supports refresh tokens** (configure with `offline_access` scope)
- Workspace-specific installation

---

### Microsoft

**Use Case**: Microsoft 365, Azure AD, Outlook, OneDrive, Teams

#### Setup Instructions

1. Go to [Azure Portal](https://portal.azure.com/)
2. Navigate to **Azure Active Directory > App registrations**
3. Click **New registration**
4. Enter application details:
   - **Name**: gorax
   - **Supported account types**: Choose appropriate option
   - **Redirect URI**: Web - `${OAUTH_BASE_URL}/api/v1/oauth/callback/microsoft`
5. After creation, go to **Certificates & secrets**
6. Create a **New client secret**
7. Copy **Application (client) ID** and the **client secret value**

#### Configuration

```bash
OAUTH_MICROSOFT_CLIENT_ID=your_application_client_id_here
OAUTH_MICROSOFT_CLIENT_SECRET=your_client_secret_here
```

#### Default Scopes

- `user.read`: Read user profile
- `mail.read`: Read user email

#### Notes

- **Supports PKCE**
- **Supports refresh tokens**
- Uses Microsoft Identity Platform (v2.0 endpoint)

---

### Twitter/X

**Use Case**: Twitter/X API integration, tweet posting, user data

#### Setup Instructions

1. Go to [Twitter Developer Portal](https://developer.twitter.com/en/portal/projects-and-apps)
2. Create a new **Project** and **App** (or use existing)
3. Navigate to your App settings
4. Under **User authentication settings**, click **Set up**
5. Select **OAuth 2.0** (not OAuth 1.0a)
6. Configure OAuth 2.0 settings:
   - **Type of App**: Web App
   - **Callback URI**: `${OAUTH_BASE_URL}/api/v1/oauth/callback/twitter`
7. Copy **Client ID** and **Client Secret**

#### Configuration

```bash
OAUTH_TWITTER_CLIENT_ID=your_client_id_here
OAUTH_TWITTER_CLIENT_SECRET=your_client_secret_here
```

#### Default Scopes

- `tweet.read`: Read tweets
- `users.read`: Read user information
- `offline.access`: Request refresh tokens

#### Notes

- **Supports PKCE** (required by Twitter)
- **Supports refresh tokens** (with `offline.access` scope)
- Uses Twitter API v2 OAuth 2.0 (not the legacy v1.0a)

---

### LinkedIn

**Use Case**: LinkedIn profile access, professional networking data

#### Setup Instructions

1. Go to [LinkedIn Developers](https://www.linkedin.com/developers/apps)
2. Click **Create app**
3. Fill in app details:
   - **App name**: gorax
   - **LinkedIn Page**: Select or create a company page
4. After creation, go to **Auth** tab
5. Add **Authorized redirect URL**: `${OAUTH_BASE_URL}/api/v1/oauth/callback/linkedin`
6. Under **Products**, request access to required products (e.g., "Sign In with LinkedIn using OpenID Connect")
7. Copy **Client ID** and **Client Secret**

#### Configuration

```bash
OAUTH_LINKEDIN_CLIENT_ID=your_client_id_here
OAUTH_LINKEDIN_CLIENT_SECRET=your_client_secret_here
```

#### Default Scopes

- `profile`: Access user profile
- `email`: Access email address
- `openid`: OpenID Connect authentication

#### Notes

- **Does NOT support PKCE**
- **Does NOT support refresh tokens** - access tokens are long-lived (60 days)
- Uses OpenID Connect for user authentication

---

### Salesforce

**Use Case**: Salesforce CRM integration, customer data, automation

#### Setup Instructions

1. Log in to Salesforce (production or sandbox)
2. Go to **Setup** (gear icon)
3. Navigate to **Platform Tools > Apps > App Manager**
4. Click **New Connected App**
5. Fill in basic information:
   - **Connected App Name**: gorax
   - **API Name**: gorax (auto-filled)
   - **Contact Email**: Your email
6. Under **API (Enable OAuth Settings)**:
   - Check **Enable OAuth Settings**
   - **Callback URL**: `${OAUTH_BASE_URL}/api/v1/oauth/callback/salesforce`
   - **Selected OAuth Scopes**: Add required scopes (api, refresh_token, openid, profile, email)
   - Check **Enable for Device Flow** (optional)
7. Save and wait for ~10 minutes for changes to propagate
8. Click **Manage Consumer Details** to view **Consumer Key** (Client ID) and **Consumer Secret**

#### Configuration

```bash
OAUTH_SALESFORCE_CLIENT_ID=your_consumer_key_here
OAUTH_SALESFORCE_CLIENT_SECRET=your_consumer_secret_here
OAUTH_SALESFORCE_ENVIRONMENT=production  # or sandbox
```

#### Default Scopes

- `api`: Access Salesforce APIs
- `refresh_token`: Request refresh tokens
- `openid`: OpenID Connect authentication
- `profile`: Access user profile
- `email`: Access email address

#### Environment Configuration

- **Production**: Uses `login.salesforce.com` (default)
- **Sandbox**: Uses `test.salesforce.com` (set `OAUTH_SALESFORCE_ENVIRONMENT=sandbox`)

#### Notes

- **Supports PKCE**
- **Supports refresh tokens**
- Requires approval process for production use
- Separate configuration required for sandbox vs production

---

### Auth0

**Use Case**: Universal authentication platform, identity management

#### Setup Instructions

1. Go to [Auth0 Dashboard](https://manage.auth0.com/)
2. Navigate to **Applications > Applications**
3. Click **Create Application**
4. Select **Regular Web Applications**
5. After creation, go to **Settings** tab
6. Configure:
   - **Allowed Callback URLs**: `${OAUTH_BASE_URL}/api/v1/oauth/callback/auth0`
   - **Allowed Web Origins**: Your gorax instance URL
7. Copy **Domain**, **Client ID**, and **Client Secret**

#### Configuration

```bash
OAUTH_AUTH0_DOMAIN=your-tenant.auth0.com  # or custom domain
OAUTH_AUTH0_CLIENT_ID=your_client_id_here
OAUTH_AUTH0_CLIENT_SECRET=your_client_secret_here
```

#### Domain Variations

Auth0 supports several domain configurations:
- **Standard tenant**: `your-tenant.auth0.com`
- **Regional tenant**: `your-tenant.us.auth0.com`, `your-tenant.eu.auth0.com`
- **Custom domain**: `login.yourcompany.com` (requires setup in Auth0)

#### Default Scopes

- `openid`: OpenID Connect authentication
- `profile`: Access user profile
- `email`: Access email address
- `offline_access`: Request refresh tokens

#### Notes

- **Supports PKCE**
- **Supports refresh tokens** (with `offline_access` scope)
- Highly configurable authentication flows
- Can be configured to use social providers (Google, Facebook, etc.)

---

## Security Considerations

### Token Storage

- **Encryption**: All access and refresh tokens are encrypted at rest using AES-256-GCM envelope encryption
- **Key Management**: Production deployments should use AWS KMS for key management
- **Database Security**: Tokens are never stored in plaintext

### PKCE (Proof Key for Code Exchange)

PKCE adds an additional security layer to prevent authorization code interception attacks. The following providers support PKCE:

- ✅ Google
- ✅ Slack
- ✅ Microsoft
- ✅ Twitter
- ❌ GitHub (not supported)
- ❌ LinkedIn (not supported)
- ✅ Salesforce
- ✅ Auth0

### Refresh Tokens

Refresh tokens allow long-term access without requiring users to re-authenticate. Support by provider:

- ❌ GitHub (not supported - long-lived access tokens)
- ✅ Google
- ✅ Slack
- ✅ Microsoft
- ✅ Twitter (with `offline.access` scope)
- ❌ LinkedIn (not supported - 60-day access tokens)
- ✅ Salesforce
- ✅ Auth0 (with `offline_access` scope)

### Best Practices

1. **Rotate Secrets**: Regularly rotate client secrets
2. **Scope Minimization**: Only request necessary scopes
3. **Audit Logs**: Review OAuth connection logs regularly
4. **Revocation**: Implement token revocation for compromised connections
5. **HTTPS Only**: Always use HTTPS in production
6. **Environment Separation**: Use separate OAuth apps for dev/staging/production

---

## Troubleshooting

### Common Issues

#### "Invalid redirect URI" Error

**Cause**: The callback URL doesn't match what's registered with the provider.

**Solution**:
1. Verify `OAUTH_BASE_URL` is correctly set
2. Ensure the redirect URI in provider settings matches: `${OAUTH_BASE_URL}/api/v1/oauth/callback/{provider}`
3. Check for trailing slashes (most providers are strict about exact matches)

#### "Invalid client" Error

**Cause**: Client ID or secret is incorrect.

**Solution**:
1. Verify credentials are correctly copied from provider
2. Check for extra spaces or newlines in `.env` file
3. Ensure the OAuth app is enabled in the provider's dashboard

#### Token Refresh Failures

**Cause**: Refresh token expired or revoked.

**Solution**:
1. Check if provider supports refresh tokens
2. Verify `offline_access` or equivalent scope is requested
3. Users may need to re-authenticate if refresh token is expired

#### Salesforce "Authentication failure" Error

**Cause**: Salesforce has a propagation delay after creating Connected Apps.

**Solution**:
1. Wait 10-15 minutes after creating the Connected App
2. Verify the environment setting (production vs sandbox)
3. Check IP restrictions in Connected App settings

#### Auth0 Domain Issues

**Cause**: Incorrect domain configuration.

**Solution**:
1. Ensure domain includes full hostname: `tenant.auth0.com`
2. For regional tenants, include region: `tenant.us.auth0.com`
3. For custom domains, use the custom domain URL

### Testing OAuth Connections

You can test OAuth connections using the gorax API:

```bash
# Start OAuth authorization flow
curl -X POST http://localhost:8080/api/v1/oauth/authorize \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"provider_key": "github"}'

# Test existing connection
curl -X POST http://localhost:8080/api/v1/oauth/connections/{connection_id}/test \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Debug Logging

Enable debug logging for OAuth operations by setting log level:

```bash
LOG_LEVEL=debug
```

This will log:
- Authorization URL generation
- Token exchange requests/responses
- Token refresh attempts
- User info API calls

### Getting Help

For additional assistance:
1. Check the [main documentation](../README.md)
2. Review audit logs in the database: `oauth_connection_logs` table
3. Enable debug logging and review application logs
4. Verify provider-specific documentation for any recent API changes

---

## Provider Comparison Matrix

| Provider   | PKCE | Refresh Tokens | Token Lifetime | User Info API | Notes |
|------------|------|----------------|----------------|---------------|-------|
| GitHub     | ❌   | ❌             | Long-lived     | ✅            | Simple, no refresh |
| Google     | ✅   | ✅             | 1 hour         | ✅            | Most flexible |
| Slack      | ✅   | ✅             | Varies         | ✅            | Workspace-specific |
| Microsoft  | ✅   | ✅             | 1 hour         | ✅            | Enterprise features |
| Twitter    | ✅   | ✅             | 2 hours        | ✅            | Requires PKCE |
| LinkedIn   | ❌   | ❌             | 60 days        | ✅            | Long-lived tokens |
| Salesforce | ✅   | ✅             | Configurable   | ✅            | Sandbox support |
| Auth0      | ✅   | ✅             | Configurable   | ✅            | Most configurable |

---

## Frequently Asked Questions

### Can I use multiple OAuth apps for the same provider?

Currently, gorax supports one OAuth app configuration per provider. For multi-tenant scenarios, configure OAuth apps at the organization/tenant level in the provider's settings.

### How are tokens encrypted?

gorax uses envelope encryption:
1. A unique Data Encryption Key (DEK) is generated for each token
2. The DEK encrypts the token using AES-256-GCM
3. The DEK itself is encrypted using a master key (or AWS KMS in production)
4. Only the encrypted token and encrypted DEK are stored

### What happens when a token expires?

If a refresh token is available, gorax automatically refreshes the access token before use. If no refresh token is available (GitHub, LinkedIn), users must re-authenticate when the token expires.

### Can users revoke OAuth connections?

Yes, users can revoke connections through the gorax UI or API. This marks the connection as revoked and attempts to revoke the token with the provider (if supported).

### How do I migrate from OAuth 1.0a to OAuth 2.0?

Twitter is the primary provider that moved from OAuth 1.0a to 2.0:
1. Create a new OAuth 2.0 app in Twitter Developer Portal
2. Update your gorax configuration with the new credentials
3. Users will need to re-authenticate with the new OAuth 2.0 flow

---

**Last Updated**: January 2026
**gorax Version**: 1.0.0
