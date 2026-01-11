package oauth

import (
	"context"
	"errors"
	"time"
)

// Common errors
var (
	ErrInvalidProvider     = errors.New("invalid OAuth provider")
	ErrInvalidState        = errors.New("invalid or expired OAuth state")
	ErrConnectionNotFound  = errors.New("OAuth connection not found")
	ErrConnectionRevoked   = errors.New("OAuth connection has been revoked")
	ErrTokenExpired        = errors.New("OAuth token has expired")
	ErrInvalidCode         = errors.New("invalid authorization code")
	ErrTokenRefreshFailed  = errors.New("failed to refresh OAuth token")
	ErrMissingRefreshToken = errors.New("refresh token not available")
)

// ProviderStatus represents the status of an OAuth provider
type ProviderStatus string

const (
	ProviderStatusActive   ProviderStatus = "active"
	ProviderStatusInactive ProviderStatus = "inactive"
)

// ConnectionStatus represents the status of an OAuth connection
type ConnectionStatus string

const (
	ConnectionStatusActive  ConnectionStatus = "active"
	ConnectionStatusRevoked ConnectionStatus = "revoked"
	ConnectionStatusExpired ConnectionStatus = "expired"
)

// OAuthProvider represents an OAuth 2.0 provider configuration
type OAuthProvider struct {
	ID                    string                 `json:"id" db:"id"`
	ProviderKey           string                 `json:"provider_key" db:"provider_key"`
	Name                  string                 `json:"name" db:"name"`
	Description           string                 `json:"description" db:"description"`
	AuthURL               string                 `json:"auth_url" db:"auth_url"`
	TokenURL              string                 `json:"token_url" db:"token_url"`
	UserInfoURL           string                 `json:"user_info_url" db:"user_info_url"`
	DefaultScopes         []string               `json:"default_scopes" db:"default_scopes"`
	ClientID              string                 `json:"client_id,omitempty" db:"client_id"`
	ClientSecretEncrypted []byte                 `json:"-" db:"client_secret_encrypted"`
	ClientSecretNonce     []byte                 `json:"-" db:"client_secret_nonce"`
	ClientSecretAuthTag   []byte                 `json:"-" db:"client_secret_auth_tag"`
	ClientSecretEncDEK    []byte                 `json:"-" db:"client_secret_encrypted_dek"`
	ClientSecretKMSKeyID  string                 `json:"-" db:"client_secret_kms_key_id"`
	Status                ProviderStatus         `json:"status" db:"status"`
	Config                map[string]interface{} `json:"config,omitempty" db:"config"`
	CreatedAt             time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at" db:"updated_at"`
}

// OAuthConnection represents a user's OAuth connection to a provider
type OAuthConnection struct {
	ID               string `json:"id" db:"id"`
	UserID           string `json:"user_id" db:"user_id"`
	TenantID         string `json:"tenant_id" db:"tenant_id"`
	ProviderKey      string `json:"provider_key" db:"provider_key"`
	ProviderUserID   string `json:"provider_user_id,omitempty" db:"provider_user_id"`
	ProviderUsername string `json:"provider_username,omitempty" db:"provider_username"`
	ProviderEmail    string `json:"provider_email,omitempty" db:"provider_email"`

	// Encrypted tokens
	AccessTokenEncrypted []byte `json:"-" db:"access_token_encrypted"`
	AccessTokenNonce     []byte `json:"-" db:"access_token_nonce"`
	AccessTokenAuthTag   []byte `json:"-" db:"access_token_auth_tag"`
	AccessTokenEncDEK    []byte `json:"-" db:"access_token_encrypted_dek"`
	AccessTokenKMSKeyID  string `json:"-" db:"access_token_kms_key_id"`

	RefreshTokenEncrypted []byte `json:"-" db:"refresh_token_encrypted"`
	RefreshTokenNonce     []byte `json:"-" db:"refresh_token_nonce"`
	RefreshTokenAuthTag   []byte `json:"-" db:"refresh_token_auth_tag"`
	RefreshTokenEncDEK    []byte `json:"-" db:"refresh_token_encrypted_dek"`
	RefreshTokenKMSKeyID  string `json:"-" db:"refresh_token_kms_key_id"`

	TokenExpiry *time.Time       `json:"token_expiry,omitempty" db:"token_expiry"`
	Scopes      []string         `json:"scopes" db:"scopes"`
	Status      ConnectionStatus `json:"status" db:"status"`

	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	LastRefreshAt *time.Time `json:"last_refresh_at,omitempty" db:"last_refresh_at"`

	RawTokenResponse map[string]interface{} `json:"-" db:"raw_token_response"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// IsExpired checks if the OAuth token has expired
func (c *OAuthConnection) IsExpired() bool {
	if c.TokenExpiry == nil {
		return false
	}
	return time.Now().After(*c.TokenExpiry)
}

// NeedsRefresh checks if token should be refreshed (expires in < 5 minutes)
func (c *OAuthConnection) NeedsRefresh() bool {
	if c.TokenExpiry == nil {
		return false
	}
	return time.Now().Add(5 * time.Minute).After(*c.TokenExpiry)
}

// OAuthState represents temporary OAuth state for CSRF protection
type OAuthState struct {
	State        string                 `json:"state" db:"state"`
	UserID       string                 `json:"user_id" db:"user_id"`
	TenantID     string                 `json:"tenant_id" db:"tenant_id"`
	ProviderKey  string                 `json:"provider_key" db:"provider_key"`
	RedirectURI  string                 `json:"redirect_uri,omitempty" db:"redirect_uri"`
	CodeVerifier string                 `json:"-" db:"code_verifier"` // For PKCE
	Scopes       []string               `json:"scopes,omitempty" db:"scopes"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	ExpiresAt    time.Time              `json:"expires_at" db:"expires_at"`
	Used         bool                   `json:"used" db:"used"`
}

// IsExpired checks if the OAuth state has expired
func (s *OAuthState) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// OAuthConnectionLog represents an audit log entry
type OAuthConnectionLog struct {
	ID           string                 `json:"id" db:"id"`
	ConnectionID string                 `json:"connection_id" db:"connection_id"`
	UserID       string                 `json:"user_id" db:"user_id"`
	TenantID     string                 `json:"tenant_id" db:"tenant_id"`
	Action       string                 `json:"action" db:"action"`
	Success      bool                   `json:"success" db:"success"`
	ErrorMessage string                 `json:"error_message,omitempty" db:"error_message"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// AuthorizeInput represents input for starting OAuth authorization
type AuthorizeInput struct {
	ProviderKey string   `json:"provider_key"`
	Scopes      []string `json:"scopes,omitempty"`
	RedirectURI string   `json:"redirect_uri,omitempty"`
}

// CallbackInput represents OAuth callback parameters
type CallbackInput struct {
	Code  string `json:"code"`
	State string `json:"state"`
	Error string `json:"error,omitempty"`
}

// OAuthService defines the OAuth service interface
type OAuthService interface {
	// GetProvider retrieves an OAuth provider by key
	GetProvider(ctx context.Context, providerKey string) (*OAuthProvider, error)

	// ListProviders lists all available OAuth providers
	ListProviders(ctx context.Context) ([]*OAuthProvider, error)

	// Authorize starts the OAuth authorization flow
	Authorize(ctx context.Context, userID, tenantID string, input *AuthorizeInput) (string, error)

	// HandleCallback handles the OAuth callback and exchanges code for tokens
	HandleCallback(ctx context.Context, userID, tenantID string, input *CallbackInput) (*OAuthConnection, error)

	// GetConnection retrieves a user's OAuth connection
	GetConnection(ctx context.Context, userID, tenantID, providerKey string) (*OAuthConnection, error)

	// ListConnections lists all OAuth connections for a user
	ListConnections(ctx context.Context, userID, tenantID string) ([]*OAuthConnection, error)

	// RevokeConnection revokes an OAuth connection
	RevokeConnection(ctx context.Context, userID, tenantID, connectionID string) error

	// RefreshToken refreshes an expired OAuth token
	RefreshToken(ctx context.Context, connectionID string) error

	// TestConnection tests an OAuth connection
	TestConnection(ctx context.Context, connectionID string) error

	// GetAccessToken retrieves and refreshes if needed the access token
	GetAccessToken(ctx context.Context, connectionID string) (string, error)
}

// OAuthRepository defines the OAuth repository interface
type OAuthRepository interface {
	// Provider operations
	GetProviderByKey(ctx context.Context, providerKey string) (*OAuthProvider, error)
	ListProviders(ctx context.Context) ([]*OAuthProvider, error)

	// Connection operations
	CreateConnection(ctx context.Context, conn *OAuthConnection) error
	GetConnection(ctx context.Context, id string) (*OAuthConnection, error)
	GetConnectionByUserProvider(ctx context.Context, userID, tenantID, providerKey string) (*OAuthConnection, error)
	ListConnectionsByUser(ctx context.Context, userID, tenantID string) ([]*OAuthConnection, error)
	UpdateConnection(ctx context.Context, conn *OAuthConnection) error
	DeleteConnection(ctx context.Context, id string) error

	// State operations
	CreateState(ctx context.Context, state *OAuthState) error
	GetState(ctx context.Context, stateStr string) (*OAuthState, error)
	MarkStateUsed(ctx context.Context, stateStr string) error
	DeleteExpiredStates(ctx context.Context) (int, error)

	// Log operations
	CreateLog(ctx context.Context, log *OAuthConnectionLog) error
}
