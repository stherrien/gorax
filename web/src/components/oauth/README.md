# OAuth Components

Complete UI components for managing OAuth connections in Gorax.

## Overview

This directory contains all frontend components for OAuth 2.0 connection management, including provider cards, connection management, and OAuth callback handling.

## Components

### OAuthProviderCard

Displays available OAuth provider with connect button.

**Props:**
- `provider: OAuthProvider` - Provider configuration
- `isConnected: boolean` - Whether user has active connection
- `onConnect: () => void` - Callback when connect button clicked

**Features:**
- Provider branding (icon, colors)
- Default scopes display
- Connect/Connected status
- Responsive design

### OAuthConnectionCard

Displays user's OAuth connection with management actions.

**Props:**
- `connection: OAuthConnection` - User's OAuth connection

**Features:**
- Connection status (active/expired/revoked)
- Provider user info (username, email)
- Scopes display
- Token expiry warning
- Test connection button
- Disconnect with confirmation
- Last used date

### OAuthConnectionList

Lists all user's OAuth connections grouped by provider.

**Props:**
- `connections: OAuthConnection[]` - List of connections
- `isLoading?: boolean` - Loading state

**Features:**
- Groups by provider
- Empty state
- Loading skeleton
- Counts per provider

### OAuthConnectButton

Reusable button to initiate OAuth flow.

**Props:**
- `providerKey: string` - Provider key (github, google, etc.)
- `scopes?: string[]` - Optional custom scopes
- `onSuccess?: () => void` - Success callback
- `onError?: (error: Error) => void` - Error callback
- `children?: React.ReactNode` - Custom button text
- `className?: string` - Custom CSS classes

**Features:**
- Opens OAuth popup
- Handles authorization flow
- Refreshes connections after success
- Loading state

**Usage:**
```tsx
<OAuthConnectButton
  providerKey="github"
  scopes={['read:user', 'repo']}
  onSuccess={() => console.log('Connected!')}
/>
```

### OAuthCallback

Handles OAuth callback from provider.

**Features:**
- Parses callback parameters
- Handles errors
- Shows success/error states
- Closes popup or redirects
- Communicates with parent window

**Route:** `/oauth/callback/:provider`

## Pages

### OAuthConnections

Main page for OAuth connection management.

**Location:** `/oauth/connections`

**Features:**
- List all available providers
- Show connected providers
- Connect new providers
- Manage existing connections
- HTTPS warning
- Help documentation

## API Client

### Methods

- `listProviders()` - Get available providers
- `authorize(input)` - Get authorization URL
- `listConnections()` - Get user connections
- `getConnection(id)` - Get specific connection
- `revokeConnection(id)` - Revoke connection
- `testConnection(id)` - Test if connection works
- `handleCallback(provider, code, state)` - Handle OAuth callback

## Custom Hooks

### useOAuthProviders()

Fetches available OAuth providers.

**Returns:** `{ data, isLoading, error }`

### useOAuthConnections()

Fetches user's OAuth connections.

**Returns:** `{ data, isLoading, error }`

### useOAuthConnection(id)

Fetches specific connection.

**Returns:** `{ data, isLoading, error }`

### useAuthorize()

Starts OAuth authorization flow.

**Returns:** `{ mutateAsync, isPending, error }`

### useRevokeConnection()

Revokes an OAuth connection.

**Returns:** `{ mutateAsync, isPending, error }`

### useTestConnection()

Tests an OAuth connection.

**Returns:** `{ mutateAsync, isPending, error, data }`

### useHasConnection(providerKey)

Check if user has active connection for provider.

**Returns:** `{ hasConnection, connection }`

### openOAuthPopup(authUrl, provider)

Opens OAuth flow in popup window.

**Returns:** `Promise<void>`

## Types

All TypeScript types are defined in `/web/src/types/oauth.ts`:

- `OAuthProvider` - Provider configuration
- `OAuthConnection` - User connection
- `ProviderStatus` - active | inactive
- `ConnectionStatus` - active | revoked | expired
- `AuthorizeInput` - Authorization parameters
- `AuthorizeResponse` - Authorization URL
- `CallbackParams` - OAuth callback parameters
- `CallbackResponse` - Callback result
- `TestConnectionResponse` - Test result
- `ProviderBranding` - UI branding config
- `PROVIDER_BRANDING` - Branding map

## Security

- All tokens encrypted with AES-256-GCM
- Tokens never exposed in frontend
- CSRF protection via state parameter
- PKCE support for public clients
- HTTPS enforcement warnings
- Secure popup handling
- Confirmation for revocation

## Testing

All components have comprehensive tests:

- `oauth.test.ts` - API client tests
- `OAuthProviderCard.test.tsx`
- `OAuthConnectionCard.test.tsx`
- `OAuthConnectionList.test.tsx`
- `OAuthConnectButton.test.tsx`
- `OAuthCallback.test.tsx`

**Run tests:**
```bash
npm test src/api/oauth.test.ts src/components/oauth
```

## Usage Examples

### Basic Provider Card
```tsx
import { OAuthProviderCard } from './components/oauth'

<OAuthProviderCard
  provider={provider}
  isConnected={false}
  onConnect={() => console.log('Connect clicked')}
/>
```

### Connection Management
```tsx
import { OAuthConnectionList } from './components/oauth'
import { useOAuthConnections } from './hooks/useOAuth'

function MyComponent() {
  const { data: connections, isLoading } = useOAuthConnections()

  return (
    <OAuthConnectionList
      connections={connections || []}
      isLoading={isLoading}
    />
  )
}
```

### Custom Connect Button
```tsx
import { OAuthConnectButton } from './components/oauth'

<OAuthConnectButton
  providerKey="github"
  scopes={['read:user', 'repo']}
  className="custom-button-class"
  onSuccess={() => {
    toast.success('Connected to GitHub!')
  }}
  onError={(error) => {
    toast.error(`Failed: ${error.message}`)
  }}
>
  Connect Your GitHub
</OAuthConnectButton>
```

### Check Connection Status
```tsx
import { useHasConnection } from './hooks/useOAuth'

function WorkflowEditor() {
  const { hasConnection, connection } = useHasConnection('github')

  if (!hasConnection) {
    return <div>Please connect GitHub first</div>
  }

  return <div>Connected as {connection.provider_username}</div>
}
```

## Integration with Workflow Editor

When adding OAuth-based actions (GitHub, Slack, etc.):

1. Check if user has connection: `useHasConnection(providerKey)`
2. Show "Connect Provider" button if not connected
3. Show connection selector if multiple connections
4. Display scopes required vs granted
5. Warn if token is expiring soon

## Architecture

```
types/oauth.ts              # TypeScript definitions
api/oauth.ts               # API client
hooks/useOAuth.ts          # React Query hooks
components/oauth/
  ├── OAuthProviderCard.tsx
  ├── OAuthConnectionCard.tsx
  ├── OAuthConnectionList.tsx
  ├── OAuthConnectButton.tsx
  ├── OAuthCallback.tsx
  └── index.ts
pages/
  └── OAuthConnections.tsx  # Main management page
```

## Backend Integration

This frontend integrates with:

- `internal/oauth/` - OAuth domain logic
- `internal/api/handlers/oauth_handler.go` - HTTP handlers
- `internal/oauth/service.go` - OAuth service
- `internal/oauth/repository.go` - Data access

See backend documentation for provider setup and configuration.
