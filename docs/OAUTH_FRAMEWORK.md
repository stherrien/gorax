# OAuth 2.0 Integration Framework

## Overview

Gorax now includes a complete OAuth 2.0 framework for building pre-built integrations with third-party services. This framework supports the OAuth 2.0 authorization code flow with PKCE for enhanced security.

## Supported Providers

- **GitHub**: Repository management, issue tracking, pull requests
- **Google**: Google Workspace APIs, Google Sheets, Gmail
- **Slack**: Messaging, channel management, workspace integration
- **Microsoft**: Microsoft 365, Azure, Microsoft Graph API

## Architecture

### Core Components

1. **OAuth Domain Models** (`internal/oauth/domain.go`)
   - `OAuthProvider`: Provider configuration
   - `OAuthConnection`: User OAuth connections with encrypted tokens
   - `OAuthState`: CSRF protection state
   - `OAuthConnectionLog`: Audit logging

2. **OAuth Service** (`internal/oauth/service.go`)
   - Authorization flow orchestration
   - Token exchange and refresh
   - Connection management
   - Automatic token refresh

3. **OAuth Providers** (`internal/oauth/providers/`)
   - Provider-specific implementations
   - Standard OAuth 2.0 interface
   - Support for PKCE where available

4. **OAuth Repository** (`internal/oauth/repository.go`)
   - PostgreSQL persistence
   - Encrypted token storage using envelope encryption
   - State management for CSRF protection

5. **OAuth HTTP Handlers** (`internal/api/handlers/oauth_handler.go`)
   - RESTful API endpoints
   - Authorization flow endpoints
   - Connection management

## Database Schema

### oauth_providers
Stores OAuth provider configurations:
- Provider metadata (name, URLs, scopes)
- Encrypted client secrets
- Provider-specific configuration

### oauth_connections
Stores user OAuth connections:
- Encrypted access and refresh tokens (envelope encryption)
- Token expiry tracking
- Provider user information
- Connection status (active, revoked, expired)
- Last used and last refresh timestamps

### oauth_states
Temporary storage for OAuth state (CSRF protection):
- PKCE code verifier
- Requested scopes
- Expiry (10 minutes)
- One-time use flag

### oauth_connection_logs
Audit log for OAuth operations:
- Authorization attempts
- Token refreshes
- Connection tests
- Revocations

## Security Features

### 1. CSRF Protection
- Cryptographically secure random state generation
- State validation on callback
- One-time use enforcement
- 10-minute expiry

### 2. PKCE Support
- SHA256 code challenge
- Code verifier validation
- Support for Google, Slack, and Microsoft
- GitHub doesn't support PKCE but interface is consistent

### 3. Token Encryption
- Envelope encryption using existing credential encryption service
- AES-256-GCM encryption
- Separate encryption for access and refresh tokens
- KMS support for production

### 4. Connection Security
- Per-tenant isolation
- User ownership validation
- Revocation support
- Connection status tracking

## API Endpoints

### List Available Providers
```
GET /api/v1/oauth/providers
```

Returns list of configured OAuth providers with metadata (client secrets excluded).

### Start OAuth Authorization
```
GET /api/v1/oauth/authorize/:provider?scopes=scope1,scope2&redirect_uri=...
```

Initiates OAuth flow. Returns authorization URL or redirects to provider.

**Parameters:**
- `provider`: Provider key (github, google, slack, microsoft)
- `scopes`: Optional comma-separated list of scopes
- `redirect_uri`: Optional custom redirect URI

### OAuth Callback
```
GET /api/v1/oauth/callback/:provider?code=...&state=...
```

Handles OAuth callback, exchanges code for tokens, and stores connection.

### List User Connections
```
GET /api/v1/oauth/connections
```

Returns all OAuth connections for the authenticated user.

### Get Connection Details
```
GET /api/v1/oauth/connections/:id
```

Returns details for a specific connection (tokens excluded).

### Revoke Connection
```
DELETE /api/v1/oauth/connections/:id
```

Revokes an OAuth connection and marks it as revoked.

### Test Connection
```
POST /api/v1/oauth/connections/:id/test
```

Tests an OAuth connection by attempting to fetch user info.

## Configuration

### Environment Variables

```bash
# Base URL for OAuth callbacks
OAUTH_BASE_URL=http://localhost:8080

# GitHub OAuth Application
OAUTH_GITHUB_CLIENT_ID=your_github_client_id
OAUTH_GITHUB_CLIENT_SECRET=your_github_client_secret

# Google OAuth Application
OAUTH_GOOGLE_CLIENT_ID=your_google_client_id
OAUTH_GOOGLE_CLIENT_SECRET=your_google_client_secret

# Slack OAuth Application
OAUTH_SLACK_CLIENT_ID=your_slack_client_id
OAUTH_SLACK_CLIENT_SECRET=your_slack_client_secret

# Microsoft OAuth Application
OAUTH_MICROSOFT_CLIENT_ID=your_microsoft_client_id
OAUTH_MICROSOFT_CLIENT_SECRET=your_microsoft_client_secret
```

### Provider Registration

#### GitHub
1. Go to https://github.com/settings/developers
2. Click "New OAuth App"
3. Set callback URL: `${OAUTH_BASE_URL}/api/v1/oauth/callback/github`
4. Copy Client ID and Client Secret to environment variables

#### Google
1. Go to https://console.cloud.google.com/apis/credentials
2. Create OAuth 2.0 Client ID
3. Set redirect URI: `${OAUTH_BASE_URL}/api/v1/oauth/callback/google`
4. Copy Client ID and Client Secret to environment variables

#### Slack
1. Go to https://api.slack.com/apps
2. Create New App
3. Enable OAuth & Permissions
4. Set redirect URL: `${OAUTH_BASE_URL}/api/v1/oauth/callback/slack`
5. Copy Client ID and Client Secret to environment variables

#### Microsoft
1. Go to https://portal.azure.com/#blade/Microsoft_AAD_RegisteredApps/ApplicationsListBlade
2. Register new application
3. Add redirect URI: `${OAUTH_BASE_URL}/api/v1/oauth/callback/microsoft`
4. Create client secret
5. Copy Application (client) ID and Client Secret to environment variables

## Usage in Workflows

### Example: GitHub Action
```go
// Workflow action using OAuth connection
type GitHubCreateIssueAction struct {
    ConnectionID string
    Repository   string
    Title        string
    Body         string
}

func (a *GitHubCreateIssueAction) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Get OAuth service
    oauthSvc := // ... inject from context

    // Get access token (auto-refreshes if needed)
    accessToken, err := oauthSvc.GetAccessToken(ctx, a.ConnectionID)
    if err != nil {
        return nil, err
    }

    // Use token to call GitHub API
    // ...
}
```

### Token Refresh

The OAuth service automatically refreshes tokens when:
- Token expires in < 5 minutes
- Token is expired
- Next API call would use expired token

Refresh is transparent to the caller.

## Integration Process

1. **User initiates OAuth flow**
   - Clicks "Connect to GitHub" in UI
   - Frontend calls `/api/v1/oauth/authorize/github`

2. **Authorization**
   - Backend generates state and PKCE challenge
   - Stores state in database
   - Redirects user to provider

3. **User authorizes**
   - Approves scopes on provider's page
   - Provider redirects to callback URL

4. **Token exchange**
   - Backend validates state
   - Exchanges authorization code for tokens
   - Encrypts and stores tokens
   - Fetches and stores user info

5. **Use in workflows**
   - Workflow actions reference connection ID
   - Service auto-refreshes tokens as needed
   - Calls provider APIs with access token

## Monitoring and Audit

All OAuth operations are logged to `oauth_connection_logs`:
- Authorization attempts
- Token refresh operations
- API calls via connections
- Connection revocations

Query audit logs:
```sql
SELECT * FROM oauth_connection_logs
WHERE tenant_id = '...'
ORDER BY created_at DESC;
```

## Error Handling

### Common Errors

- `ErrInvalidProvider`: Provider not found or inactive
- `ErrInvalidState`: State expired, used, or invalid
- `ErrConnectionNotFound`: Connection doesn't exist
- `ErrConnectionRevoked`: Connection has been revoked
- `ErrTokenExpired`: Token expired and no refresh token
- `ErrTokenRefreshFailed`: Token refresh failed
- `ErrMissingRefreshToken`: No refresh token available

## Testing

Run OAuth tests:
```bash
go test ./internal/oauth/... -v
```

Test coverage includes:
- PKCE generation and validation
- State management
- Token encryption/decryption
- Connection lifecycle
- Auto-refresh logic

## Adding New Providers

To add a new OAuth provider:

1. **Create provider implementation**
   ```go
   // internal/oauth/providers/newprovider.go
   type NewProvider struct {
       httpClient *http.Client
   }

   func NewNewProvider() *NewProvider {
       return &NewProvider{
           httpClient: &http.Client{Timeout: 30 * time.Second},
       }
   }

   // Implement oauth.Provider interface
   func (p *NewProvider) Key() string { return "newprovider" }
   func (p *NewProvider) Name() string { return "New Provider" }
   // ... implement other methods
   ```

2. **Add to database migration**
   ```sql
   INSERT INTO oauth_providers (provider_key, name, auth_url, token_url, ...)
   VALUES ('newprovider', 'New Provider', '...', '...', ...);
   ```

3. **Register in service initialization**
   ```go
   providers := map[string]oauth.Provider{
       "github":      providers.NewGitHubProvider(),
       "google":      providers.NewGoogleProvider(),
       "slack":       providers.NewSlackProvider(),
       "microsoft":   providers.NewMicrosoftProvider(),
       "newprovider": providers.NewNewProvider(), // Add here
   }
   ```

4. **Add configuration**
   ```go
   // internal/config/config.go
   type OAuthConfig struct {
       // ... existing fields
       NewProviderClientID     string
       NewProviderClientSecret string
   }
   ```

5. **Update documentation**
   - Add provider to README
   - Document registration process
   - Document required scopes

## Best Practices

### 1. Scope Management
- Request minimum required scopes
- Document why each scope is needed
- Allow users to see granted scopes

### 2. Token Refresh
- Always use `GetAccessToken()` to get tokens
- Never cache tokens yourself
- Service handles refresh automatically

### 3. Error Handling
- Handle `ErrConnectionRevoked` gracefully
- Prompt user to reconnect if needed
- Log authorization failures for audit

### 4. Security
- Never expose access tokens in API responses
- Always validate user ownership of connections
- Use HTTPS in production for callbacks
- Rotate client secrets periodically

### 5. User Experience
- Show connection status in UI
- Allow users to manage connections
- Test connections before use
- Handle provider-specific errors

## Troubleshooting

### Issue: "Invalid state" error
**Cause**: State expired (> 10 min), already used, or CSRF attack
**Solution**: Restart OAuth flow

### Issue: Token refresh fails
**Cause**: Refresh token expired or revoked by user
**Solution**: User must re-authorize

### Issue: Provider returns error
**Cause**: Invalid client credentials, wrong scopes, or user denied
**Solution**: Check provider configuration and credentials

### Issue: Connection not found
**Cause**: Connection was deleted or user doesn't have access
**Solution**: User must authorize again

## Future Enhancements

Potential improvements:
- [ ] OAuth 2.0 Device Flow for CLI tools
- [ ] Automatic scope expansion requests
- [ ] Connection health monitoring
- [ ] Provider-specific action templates
- [ ] OAuth connection sharing within tenants
- [ ] Webhook support for token revocation events
- [ ] Rate limiting per connection
- [ ] Provider-specific error handling
