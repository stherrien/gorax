package oauth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/gorax/gorax/internal/credential"
)

// Service implements OAuthService
type Service struct {
	repo          OAuthRepository
	encryptionSvc EncryptionService
	providers     map[string]Provider
	baseURL       string
}

// EncryptionService defines the encryption interface for OAuth tokens
type EncryptionService interface {
	Encrypt(ctx context.Context, tenantID string, data *credential.CredentialData) (*credential.EncryptedSecret, error)
	Decrypt(ctx context.Context, encrypted *credential.EncryptedSecret) (*credential.CredentialData, error)
}

// NewService creates a new OAuth service
func NewService(repo OAuthRepository, encryptionSvc EncryptionService, providers map[string]Provider, baseURL string) *Service {
	return &Service{
		repo:          repo,
		encryptionSvc: encryptionSvc,
		baseURL:       strings.TrimSuffix(baseURL, "/"),
		providers:     providers,
	}
}

// GetProvider retrieves an OAuth provider by key
func (s *Service) GetProvider(ctx context.Context, providerKey string) (*OAuthProvider, error) {
	return s.repo.GetProviderByKey(ctx, providerKey)
}

// ListProviders lists all available OAuth providers
func (s *Service) ListProviders(ctx context.Context) ([]*OAuthProvider, error) {
	return s.repo.ListProviders(ctx)
}

// Authorize starts the OAuth authorization flow
func (s *Service) Authorize(ctx context.Context, userID, tenantID string, input *AuthorizeInput) (string, error) {
	// Get provider config
	providerConfig, err := s.repo.GetProviderByKey(ctx, input.ProviderKey)
	if err != nil {
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	// Get provider implementation
	provider, ok := s.providers[input.ProviderKey]
	if !ok {
		return "", ErrInvalidProvider
	}

	// Generate state for CSRF protection
	state, err := GenerateState()
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Generate PKCE verifier and challenge
	codeVerifier, err := GeneratePKCEVerifier()
	if err != nil {
		return "", fmt.Errorf("failed to generate PKCE verifier: %w", err)
	}
	codeChallenge := GeneratePKCEChallenge(codeVerifier)

	// Determine scopes
	scopes := input.Scopes
	if len(scopes) == 0 {
		scopes = providerConfig.DefaultScopes
	}

	// Store state in database
	oauthState := &OAuthState{
		State:        state,
		UserID:       userID,
		TenantID:     tenantID,
		ProviderKey:  input.ProviderKey,
		RedirectURI:  input.RedirectURI,
		CodeVerifier: codeVerifier,
		Scopes:       scopes,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	}

	if err := s.repo.CreateState(ctx, oauthState); err != nil {
		return "", fmt.Errorf("failed to store state: %w", err)
	}

	// Determine redirect URI
	redirectURI := input.RedirectURI
	if redirectURI == "" {
		redirectURI = fmt.Sprintf("%s/api/v1/oauth/callback/%s", s.baseURL, input.ProviderKey)
	}

	// Generate authorization URL
	authURL := provider.GetAuthURL(providerConfig.ClientID, redirectURI, state, scopes, codeChallenge)

	return authURL, nil
}

// HandleCallback handles the OAuth callback and exchanges code for tokens
func (s *Service) HandleCallback(ctx context.Context, userID, tenantID string, input *CallbackInput) (*OAuthConnection, error) {
	// Handle error from OAuth provider
	if input.Error != "" {
		return nil, fmt.Errorf("OAuth provider error: %s", input.Error)
	}

	// Validate state
	oauthState, err := s.repo.GetState(ctx, input.State)
	if err != nil {
		return nil, fmt.Errorf("invalid state: %w", err)
	}

	if oauthState.Used {
		return nil, fmt.Errorf("state already used")
	}

	if oauthState.IsExpired() {
		return nil, ErrInvalidState
	}

	// Verify state matches user and tenant
	if oauthState.UserID != userID || oauthState.TenantID != tenantID {
		return nil, fmt.Errorf("state mismatch")
	}

	// Mark state as used
	if err := s.repo.MarkStateUsed(ctx, input.State); err != nil {
		return nil, fmt.Errorf("failed to mark state as used: %w", err)
	}

	// Get provider config
	providerConfig, err := s.repo.GetProviderByKey(ctx, oauthState.ProviderKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Get provider implementation
	provider, ok := s.providers[oauthState.ProviderKey]
	if !ok {
		return nil, ErrInvalidProvider
	}

	// Decrypt client secret if available
	var clientSecret string
	if len(providerConfig.ClientSecretEncrypted) > 0 {
		encrypted := &credential.EncryptedSecret{
			Ciphertext:   providerConfig.ClientSecretEncrypted,
			Nonce:        providerConfig.ClientSecretNonce,
			AuthTag:      providerConfig.ClientSecretAuthTag,
			EncryptedDEK: providerConfig.ClientSecretEncDEK,
			KMSKeyID:     providerConfig.ClientSecretKMSKeyID,
		}
		decrypted, err := s.encryptionSvc.Decrypt(ctx, encrypted)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt client secret: %w", err)
		}
		if secret, ok := decrypted.Value["secret"].(string); ok {
			clientSecret = secret
		}
	}

	// Determine redirect URI
	redirectURI := oauthState.RedirectURI
	if redirectURI == "" {
		redirectURI = fmt.Sprintf("%s/api/v1/oauth/callback/%s", s.baseURL, oauthState.ProviderKey)
	}

	// Exchange code for tokens
	tokenResp, err := provider.ExchangeCode(ctx, providerConfig.ClientID, clientSecret, input.Code, redirectURI, oauthState.CodeVerifier)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info
	userInfo, err := provider.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Encrypt access token
	accessTokenData := &credential.CredentialData{
		Value: map[string]interface{}{
			"token": tokenResp.AccessToken,
		},
	}
	encryptedAccessToken, err := s.encryptionSvc.Encrypt(ctx, tenantID, accessTokenData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt access token: %w", err)
	}

	// Encrypt refresh token if present
	var encryptedRefreshToken *credential.EncryptedSecret
	if tokenResp.RefreshToken != "" {
		refreshTokenData := &credential.CredentialData{
			Value: map[string]interface{}{
				"token": tokenResp.RefreshToken,
			},
		}
		encryptedRefreshToken, err = s.encryptionSvc.Encrypt(ctx, tenantID, refreshTokenData)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
	}

	// Calculate token expiry
	var tokenExpiry *time.Time
	if tokenResp.ExpiresIn > 0 {
		expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		tokenExpiry = &expiry
	}

	// Parse scopes
	scopes := oauthState.Scopes
	if tokenResp.Scope != "" {
		scopes = strings.Split(tokenResp.Scope, " ")
	}

	// Create connection
	conn := &OAuthConnection{
		ID:                   uuid.New().String(),
		UserID:               userID,
		TenantID:             tenantID,
		ProviderKey:          oauthState.ProviderKey,
		ProviderUserID:       userInfo.ID,
		ProviderUsername:     userInfo.Username,
		ProviderEmail:        userInfo.Email,
		AccessTokenEncrypted: encryptedAccessToken.Ciphertext,
		AccessTokenNonce:     encryptedAccessToken.Nonce,
		AccessTokenAuthTag:   encryptedAccessToken.AuthTag,
		AccessTokenEncDEK:    encryptedAccessToken.EncryptedDEK,
		AccessTokenKMSKeyID:  encryptedAccessToken.KMSKeyID,
		TokenExpiry:          tokenExpiry,
		Scopes:               scopes,
		Status:               ConnectionStatusActive,
		RawTokenResponse:     map[string]interface{}{},
	}

	if encryptedRefreshToken != nil {
		conn.RefreshTokenEncrypted = encryptedRefreshToken.Ciphertext
		conn.RefreshTokenNonce = encryptedRefreshToken.Nonce
		conn.RefreshTokenAuthTag = encryptedRefreshToken.AuthTag
		conn.RefreshTokenEncDEK = encryptedRefreshToken.EncryptedDEK
		conn.RefreshTokenKMSKeyID = encryptedRefreshToken.KMSKeyID
	}

	// Save connection
	if err := s.repo.CreateConnection(ctx, conn); err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	// Log successful authorization
	_ = s.logConnectionAction(ctx, conn.ID, userID, tenantID, "authorize", true, "")

	return conn, nil
}

// GetConnection retrieves a user's OAuth connection
func (s *Service) GetConnection(ctx context.Context, userID, tenantID, providerKey string) (*OAuthConnection, error) {
	conn, err := s.repo.GetConnectionByUserProvider(ctx, userID, tenantID, providerKey)
	if err != nil {
		return nil, err
	}

	// Check if connection is revoked
	if conn.Status == ConnectionStatusRevoked {
		return nil, ErrConnectionRevoked
	}

	return conn, nil
}

// ListConnections lists all OAuth connections for a user
func (s *Service) ListConnections(ctx context.Context, userID, tenantID string) ([]*OAuthConnection, error) {
	return s.repo.ListConnectionsByUser(ctx, userID, tenantID)
}

// RevokeConnection revokes an OAuth connection
func (s *Service) RevokeConnection(ctx context.Context, userID, tenantID, connectionID string) error {
	conn, err := s.repo.GetConnection(ctx, connectionID)
	if err != nil {
		return err
	}

	// Verify ownership
	if conn.UserID != userID || conn.TenantID != tenantID {
		return fmt.Errorf("unauthorized")
	}

	// Mark as revoked
	conn.Status = ConnectionStatusRevoked
	if err := s.repo.UpdateConnection(ctx, conn); err != nil {
		return fmt.Errorf("failed to revoke connection: %w", err)
	}

	// Log revocation
	_ = s.logConnectionAction(ctx, conn.ID, userID, tenantID, "revoke", true, "")

	return nil
}

// RefreshToken refreshes an expired OAuth token
func (s *Service) RefreshToken(ctx context.Context, connectionID string) error {
	conn, err := s.repo.GetConnection(ctx, connectionID)
	if err != nil {
		return err
	}

	// Check if refresh token exists
	if len(conn.RefreshTokenEncrypted) == 0 {
		return ErrMissingRefreshToken
	}

	// Get provider
	provider, ok := s.providers[conn.ProviderKey]
	if !ok {
		return ErrInvalidProvider
	}

	// Get provider config
	providerConfig, err := s.repo.GetProviderByKey(ctx, conn.ProviderKey)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Decrypt client secret
	var clientSecret string
	if len(providerConfig.ClientSecretEncrypted) > 0 {
		encrypted := &credential.EncryptedSecret{
			Ciphertext:   providerConfig.ClientSecretEncrypted,
			Nonce:        providerConfig.ClientSecretNonce,
			AuthTag:      providerConfig.ClientSecretAuthTag,
			EncryptedDEK: providerConfig.ClientSecretEncDEK,
			KMSKeyID:     providerConfig.ClientSecretKMSKeyID,
		}
		decrypted, err := s.encryptionSvc.Decrypt(ctx, encrypted)
		if err != nil {
			return fmt.Errorf("failed to decrypt client secret: %w", err)
		}
		if secret, ok := decrypted.Value["secret"].(string); ok {
			clientSecret = secret
		}
	}

	// Decrypt refresh token
	encryptedRefreshToken := &credential.EncryptedSecret{
		Ciphertext:   conn.RefreshTokenEncrypted,
		Nonce:        conn.RefreshTokenNonce,
		AuthTag:      conn.RefreshTokenAuthTag,
		EncryptedDEK: conn.RefreshTokenEncDEK,
		KMSKeyID:     conn.RefreshTokenKMSKeyID,
	}
	decryptedRefreshToken, err := s.encryptionSvc.Decrypt(ctx, encryptedRefreshToken)
	if err != nil {
		return fmt.Errorf("failed to decrypt refresh token: %w", err)
	}

	refreshToken, ok := decryptedRefreshToken.Value["token"].(string)
	if !ok {
		return fmt.Errorf("invalid refresh token format")
	}

	// Refresh the token
	tokenResp, err := provider.RefreshToken(ctx, providerConfig.ClientID, clientSecret, refreshToken)
	if err != nil {
		_ = s.logConnectionAction(ctx, conn.ID, conn.UserID, conn.TenantID, "token_refresh", false, err.Error())
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Encrypt new access token
	accessTokenData := &credential.CredentialData{
		Value: map[string]interface{}{
			"token": tokenResp.AccessToken,
		},
	}
	encryptedAccessToken, err := s.encryptionSvc.Encrypt(ctx, conn.TenantID, accessTokenData)
	if err != nil {
		return fmt.Errorf("failed to encrypt access token: %w", err)
	}

	// Update connection
	conn.AccessTokenEncrypted = encryptedAccessToken.Ciphertext
	conn.AccessTokenNonce = encryptedAccessToken.Nonce
	conn.AccessTokenAuthTag = encryptedAccessToken.AuthTag
	conn.AccessTokenEncDEK = encryptedAccessToken.EncryptedDEK
	conn.AccessTokenKMSKeyID = encryptedAccessToken.KMSKeyID

	// Update expiry
	if tokenResp.ExpiresIn > 0 {
		expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		conn.TokenExpiry = &expiry
	}

	// Update refresh token if provided
	if tokenResp.RefreshToken != "" {
		refreshTokenData := &credential.CredentialData{
			Value: map[string]interface{}{
				"token": tokenResp.RefreshToken,
			},
		}
		encryptedRefreshToken, err := s.encryptionSvc.Encrypt(ctx, conn.TenantID, refreshTokenData)
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
		conn.RefreshTokenEncrypted = encryptedRefreshToken.Ciphertext
		conn.RefreshTokenNonce = encryptedRefreshToken.Nonce
		conn.RefreshTokenAuthTag = encryptedRefreshToken.AuthTag
		conn.RefreshTokenEncDEK = encryptedRefreshToken.EncryptedDEK
		conn.RefreshTokenKMSKeyID = encryptedRefreshToken.KMSKeyID
	}

	now := time.Now()
	conn.LastRefreshAt = &now

	if err := s.repo.UpdateConnection(ctx, conn); err != nil {
		return fmt.Errorf("failed to update connection: %w", err)
	}

	// Log successful refresh
	_ = s.logConnectionAction(ctx, conn.ID, conn.UserID, conn.TenantID, "token_refresh", true, "")

	return nil
}

// TestConnection tests an OAuth connection
func (s *Service) TestConnection(ctx context.Context, connectionID string) error {
	conn, err := s.repo.GetConnection(ctx, connectionID)
	if err != nil {
		return err
	}

	// Get provider
	provider, ok := s.providers[conn.ProviderKey]
	if !ok {
		return ErrInvalidProvider
	}

	// Get access token
	accessToken, err := s.GetAccessToken(ctx, connectionID)
	if err != nil {
		return err
	}

	// Try to get user info
	_, err = provider.GetUserInfo(ctx, accessToken)
	if err != nil {
		_ = s.logConnectionAction(ctx, conn.ID, conn.UserID, conn.TenantID, "test_connection", false, err.Error())
		return fmt.Errorf("connection test failed: %w", err)
	}

	_ = s.logConnectionAction(ctx, conn.ID, conn.UserID, conn.TenantID, "test_connection", true, "")
	return nil
}

// GetAccessToken retrieves and refreshes if needed the access token
func (s *Service) GetAccessToken(ctx context.Context, connectionID string) (string, error) {
	conn, err := s.repo.GetConnection(ctx, connectionID)
	if err != nil {
		return "", err
	}

	// Check if token needs refresh
	if conn.NeedsRefresh() {
		if err := s.RefreshToken(ctx, connectionID); err != nil {
			return "", fmt.Errorf("failed to refresh token: %w", err)
		}
		// Reload connection after refresh
		conn, err = s.repo.GetConnection(ctx, connectionID)
		if err != nil {
			return "", err
		}
	}

	// Decrypt access token
	encryptedAccessToken := &credential.EncryptedSecret{
		Ciphertext:   conn.AccessTokenEncrypted,
		Nonce:        conn.AccessTokenNonce,
		AuthTag:      conn.AccessTokenAuthTag,
		EncryptedDEK: conn.AccessTokenEncDEK,
		KMSKeyID:     conn.AccessTokenKMSKeyID,
	}
	decryptedAccessToken, err := s.encryptionSvc.Decrypt(ctx, encryptedAccessToken)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt access token: %w", err)
	}

	accessToken, ok := decryptedAccessToken.Value["token"].(string)
	if !ok {
		return "", fmt.Errorf("invalid access token format")
	}

	// Update last used time
	now := time.Now()
	conn.LastUsedAt = &now
	_ = s.repo.UpdateConnection(ctx, conn)

	return accessToken, nil
}

// logConnectionAction logs an OAuth connection action
func (s *Service) logConnectionAction(ctx context.Context, connectionID, userID, tenantID, action string, success bool, errorMsg string) error {
	log := &OAuthConnectionLog{
		ID:           uuid.New().String(),
		ConnectionID: connectionID,
		UserID:       userID,
		TenantID:     tenantID,
		Action:       action,
		Success:      success,
		ErrorMessage: errorMsg,
	}
	return s.repo.CreateLog(ctx, log)
}
