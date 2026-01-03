package sso

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorax/gorax/internal/user"
)

// Service handles SSO operations and orchestration
type Service interface {
	// Provider management
	CreateProvider(ctx context.Context, tenantID uuid.UUID, req *CreateProviderRequest, createdBy uuid.UUID) (*Provider, error)
	GetProvider(ctx context.Context, tenantID, providerID uuid.UUID) (*Provider, error)
	ListProviders(ctx context.Context, tenantID uuid.UUID) ([]*Provider, error)
	UpdateProvider(ctx context.Context, tenantID, providerID uuid.UUID, req *UpdateProviderRequest, updatedBy uuid.UUID) (*Provider, error)
	DeleteProvider(ctx context.Context, tenantID, providerID uuid.UUID) error
	GetProviderByDomain(ctx context.Context, emailDomain string) (*Provider, error)

	// SSO authentication flow
	InitiateLogin(ctx context.Context, providerID uuid.UUID, relayState string) (string, error)
	HandleCallback(ctx context.Context, providerID uuid.UUID, r *http.Request) (*AuthenticationResponse, error)
	GetMetadata(ctx context.Context, providerID uuid.UUID) (string, error)

	// Testing and validation
	ValidateProvider(ctx context.Context, provider *Provider) error
}

// ServiceImpl implements the SSO service
type ServiceImpl struct {
	repo            Repository
	userRepo        user.Repository
	providerFactory ProviderFactory
}

// NewService creates a new SSO service
func NewService(repo Repository, userRepo user.Repository, providerFactory ProviderFactory) Service {
	return &ServiceImpl{
		repo:            repo,
		userRepo:        userRepo,
		providerFactory: providerFactory,
	}
}

// CreateProvider creates a new SSO provider
func (s *ServiceImpl) CreateProvider(ctx context.Context, tenantID uuid.UUID, req *CreateProviderRequest, createdBy uuid.UUID) (*Provider, error) {
	// Validate request
	if err := validateCreateProviderRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Create provider model
	provider := &Provider{
		TenantID:   tenantID,
		Name:       req.Name,
		Type:       req.Type,
		Enabled:    req.Enabled,
		EnforceSSO: req.EnforceSSO,
		Config:     req.Config,
		Domains:    req.Domains,
		CreatedBy:  &createdBy,
		UpdatedBy:  &createdBy,
	}

	// Validate provider configuration
	if err := s.ValidateProvider(ctx, provider); err != nil {
		return nil, fmt.Errorf("provider validation failed: %w", err)
	}

	// Create in database
	if err := s.repo.CreateProvider(ctx, provider); err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return provider, nil
}

// GetProvider retrieves an SSO provider
func (s *ServiceImpl) GetProvider(ctx context.Context, tenantID, providerID uuid.UUID) (*Provider, error) {
	provider, err := s.repo.GetProviderByTenant(ctx, tenantID, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Mask sensitive configuration
	maskedConfig, err := MaskSensitiveConfig(provider.Type, provider.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to mask config: %w", err)
	}
	provider.Config = maskedConfig

	return provider, nil
}

// ListProviders lists all SSO providers for a tenant
func (s *ServiceImpl) ListProviders(ctx context.Context, tenantID uuid.UUID) ([]*Provider, error) {
	providers, err := s.repo.ListProviders(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}

	// Mask sensitive configuration for all providers
	for _, provider := range providers {
		maskedConfig, err := MaskSensitiveConfig(provider.Type, provider.Config)
		if err != nil {
			continue // Skip masking if error, but don't fail
		}
		provider.Config = maskedConfig
	}

	return providers, nil
}

// UpdateProvider updates an SSO provider
func (s *ServiceImpl) UpdateProvider(ctx context.Context, tenantID, providerID uuid.UUID, req *UpdateProviderRequest, updatedBy uuid.UUID) (*Provider, error) {
	// Get existing provider
	provider, err := s.repo.GetProviderByTenant(ctx, tenantID, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Apply updates
	if req.Name != nil {
		provider.Name = *req.Name
	}
	if req.Enabled != nil {
		provider.Enabled = *req.Enabled
	}
	if req.EnforceSSO != nil {
		provider.EnforceSSO = *req.EnforceSSO
	}
	if req.Config != nil {
		provider.Config = req.Config
	}
	if req.Domains != nil {
		provider.Domains = req.Domains
	}
	provider.UpdatedBy = &updatedBy

	// Validate updated provider
	if err := s.ValidateProvider(ctx, provider); err != nil {
		return nil, fmt.Errorf("provider validation failed: %w", err)
	}

	// Update in database
	if err := s.repo.UpdateProvider(ctx, provider); err != nil {
		return nil, fmt.Errorf("failed to update provider: %w", err)
	}

	// Mask sensitive configuration before returning
	maskedConfig, err := MaskSensitiveConfig(provider.Type, provider.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to mask config: %w", err)
	}
	provider.Config = maskedConfig

	return provider, nil
}

// DeleteProvider deletes an SSO provider
func (s *ServiceImpl) DeleteProvider(ctx context.Context, tenantID, providerID uuid.UUID) error {
	// Verify provider belongs to tenant
	_, err := s.repo.GetProviderByTenant(ctx, tenantID, providerID)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Delete provider
	if err := s.repo.DeleteProvider(ctx, providerID); err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}

	return nil
}

// GetProviderByDomain retrieves an SSO provider by email domain
func (s *ServiceImpl) GetProviderByDomain(ctx context.Context, emailDomain string) (*Provider, error) {
	provider, err := s.repo.GetProviderByDomain(ctx, emailDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider by domain: %w", err)
	}

	return provider, nil
}

// InitiateLogin initiates the SSO login flow
func (s *ServiceImpl) InitiateLogin(ctx context.Context, providerID uuid.UUID, relayState string) (string, error) {
	// Get provider
	provider, err := s.repo.GetProvider(ctx, providerID)
	if err != nil {
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	if !provider.Enabled {
		return "", fmt.Errorf("SSO provider is disabled")
	}

	// Create SSO provider instance
	ssoProvider, err := s.providerFactory.CreateProvider(ctx, provider)
	if err != nil {
		return "", fmt.Errorf("failed to create SSO provider: %w", err)
	}

	// Initiate login
	redirectURL, err := ssoProvider.InitiateLogin(ctx, relayState)
	if err != nil {
		return "", fmt.Errorf("failed to initiate login: %w", err)
	}

	return redirectURL, nil
}

// HandleCallback handles the SSO callback and performs JIT provisioning
func (s *ServiceImpl) HandleCallback(ctx context.Context, providerID uuid.UUID, r *http.Request) (*AuthenticationResponse, error) {
	// Get provider
	provider, err := s.repo.GetProvider(ctx, providerID)
	if err != nil {
		return nil, s.logLoginEvent(ctx, providerID, "", LoginStatusError, fmt.Sprintf("provider not found: %v", err), r)
	}

	if !provider.Enabled {
		return nil, s.logLoginEvent(ctx, providerID, "", LoginStatusError, "provider disabled", r)
	}

	// Create SSO provider instance
	ssoProvider, err := s.providerFactory.CreateProvider(ctx, provider)
	if err != nil {
		return nil, s.logLoginEvent(ctx, providerID, "", LoginStatusError, fmt.Sprintf("failed to create provider: %v", err), r)
	}

	// Handle callback and extract user attributes
	userAttrs, err := ssoProvider.HandleCallback(ctx, r)
	if err != nil {
		return nil, s.logLoginEvent(ctx, providerID, "", LoginStatusFailure, fmt.Sprintf("callback failed: %v", err), r)
	}

	// JIT user provisioning
	usr, conn, err := s.provisionUser(ctx, provider, userAttrs)
	if err != nil {
		return nil, s.logLoginEvent(ctx, providerID, userAttrs.ExternalID, LoginStatusError, fmt.Sprintf("provisioning failed: %v", err), r)
	}

	// Update connection last login
	now := time.Now()
	conn.LastLoginAt = &now
	if err := s.repo.UpdateConnection(ctx, conn); err != nil {
		// Log but don't fail
		fmt.Printf("failed to update connection last login: %v\n", err)
	}

	// Log successful login
	userID, _ := uuid.Parse(usr.ID)
	if err := s.logLoginEvent(ctx, providerID, userAttrs.ExternalID, LoginStatusSuccess, "", r); err != nil {
		// Log but don't fail
		fmt.Printf("failed to log login event: %v\n", err)
	}

	// Create session token (in production, integrate with Kratos or JWT)
	sessionToken := fmt.Sprintf("sso-session-%s", uuid.New().String())
	expiresAt := time.Now().Add(24 * time.Hour)

	response := &AuthenticationResponse{
		UserAttributes: *userAttrs,
		SessionToken:   sessionToken,
		ExpiresAt:      expiresAt,
	}

	// Store user ID in context for downstream processing
	_ = userID

	return response, nil
}

// provisionUser performs JIT user provisioning
func (s *ServiceImpl) provisionUser(ctx context.Context, provider *Provider, attrs *UserAttributes) (*user.User, *Connection, error) {
	// Check if connection already exists
	existingConn, err := s.repo.GetConnectionByExternalID(ctx, provider.ID, attrs.ExternalID)
	if err != nil && err.Error() != "SSO connection not found" {
		return nil, nil, fmt.Errorf("failed to check existing connection: %w", err)
	}

	// If connection exists, return existing user
	if existingConn != nil {
		usr, err := s.userRepo.GetByID(ctx, existingConn.UserID.String())
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get existing user: %w", err)
		}
		return usr, existingConn, nil
	}

	// Check if user exists by email
	existingUser, err := s.userRepo.GetByEmail(ctx, attrs.Email)
	if err != nil && err != user.ErrUserNotFound {
		return nil, nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	var usr *user.User

	// Create new user if doesn't exist
	if existingUser == nil {
		usr = &user.User{
			ID:               uuid.New().String(),
			TenantID:         provider.TenantID.String(),
			KratosIdentityID: fmt.Sprintf("sso-%s", uuid.New().String()), // SSO users get special identity ID
			Email:            attrs.Email,
			Role:             "member", // Default role
			Status:           "active",
		}

		if err := s.userRepo.Create(ctx, usr); err != nil {
			return nil, nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		usr = existingUser
	}

	// Create SSO connection
	attrsJSON, err := json.Marshal(attrs.Attributes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal attributes: %w", err)
	}

	userID, err := uuid.Parse(usr.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid user ID: %w", err)
	}

	conn := &Connection{
		UserID:     userID,
		ProviderID: provider.ID,
		ExternalID: attrs.ExternalID,
		Attributes: attrsJSON,
	}

	if err := s.repo.CreateConnection(ctx, conn); err != nil {
		return nil, nil, fmt.Errorf("failed to create connection: %w", err)
	}

	return usr, conn, nil
}

// logLoginEvent logs an SSO login event and returns the error
func (s *ServiceImpl) logLoginEvent(ctx context.Context, providerID uuid.UUID, externalID string, status LoginStatus, errorMsg string, r *http.Request) error {
	var errMsgPtr *string
	if errorMsg != "" {
		errMsgPtr = &errorMsg
	}

	ipAddr := r.RemoteAddr
	userAgent := r.UserAgent()

	event := &LoginEvent{
		ProviderID:   providerID,
		ExternalID:   externalID,
		Status:       status,
		ErrorMessage: errMsgPtr,
		IPAddress:    &ipAddr,
		UserAgent:    &userAgent,
	}

	if err := s.repo.CreateLoginEvent(ctx, event); err != nil {
		fmt.Printf("failed to create login event: %v\n", err)
	}

	if errorMsg != "" {
		return fmt.Errorf("%s", errorMsg)
	}
	return nil
}

// GetMetadata returns the SSO provider metadata (for SAML)
func (s *ServiceImpl) GetMetadata(ctx context.Context, providerID uuid.UUID) (string, error) {
	// Get provider
	provider, err := s.repo.GetProvider(ctx, providerID)
	if err != nil {
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	// Create SSO provider instance
	ssoProvider, err := s.providerFactory.CreateProvider(ctx, provider)
	if err != nil {
		return "", fmt.Errorf("failed to create SSO provider: %w", err)
	}

	// Get metadata
	metadata, err := ssoProvider.GetMetadata(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get metadata: %w", err)
	}

	return metadata, nil
}

// ValidateProvider validates an SSO provider configuration
func (s *ServiceImpl) ValidateProvider(ctx context.Context, provider *Provider) error {
	// Create provider instance
	ssoProvider, err := s.providerFactory.CreateProvider(ctx, provider)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Validate
	if err := ssoProvider.Validate(ctx); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// validateCreateProviderRequest validates the create provider request
func validateCreateProviderRequest(req *CreateProviderRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	if req.Type != ProviderTypeSAML && req.Type != ProviderTypeOIDC {
		return fmt.Errorf("invalid provider type: %s", req.Type)
	}

	if len(req.Config) == 0 {
		return fmt.Errorf("config is required")
	}

	if len(req.Domains) == 0 {
		return fmt.Errorf("at least one domain is required")
	}

	// Validate domains format
	for _, domain := range req.Domains {
		if !isValidDomain(domain) {
			return fmt.Errorf("invalid domain format: %s", domain)
		}
	}

	return nil
}

// isValidDomain validates domain format
func isValidDomain(domain string) bool {
	if domain == "" {
		return false
	}

	// Basic validation - should not contain @ or spaces
	if strings.Contains(domain, "@") || strings.Contains(domain, " ") {
		return false
	}

	// Should contain at least one dot
	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}
