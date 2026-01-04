package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorax/gorax/internal/oauth"
)

// OAuthHandler handles OAuth-related HTTP requests
type OAuthHandler struct {
	service oauth.OAuthService
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(service oauth.OAuthService) *OAuthHandler {
	return &OAuthHandler{
		service: service,
	}
}

// ListProviders returns available OAuth providers
// GET /api/v1/oauth/providers
func (h *OAuthHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	providers, err := h.service.ListProviders(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list providers: %v", err), http.StatusInternalServerError)
		return
	}

	// Remove sensitive data before returning
	for _, provider := range providers {
		provider.ClientSecretEncrypted = nil
		provider.ClientSecretNonce = nil
		provider.ClientSecretAuthTag = nil
		provider.ClientSecretEncDEK = nil
		provider.ClientSecretKMSKeyID = ""
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(providers); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Authorize starts the OAuth authorization flow
// GET /api/v1/oauth/authorize/:provider
func (h *OAuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	providerKey := chi.URLParam(r, "provider")

	// Get user and tenant from context
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok {
		http.Error(w, "Missing tenant context", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	scopes := r.URL.Query()["scopes"]
	redirectURI := r.URL.Query().Get("redirect_uri")

	input := &oauth.AuthorizeInput{
		ProviderKey: providerKey,
		Scopes:      scopes,
		RedirectURI: redirectURI,
	}

	authURL, err := h.service.Authorize(ctx, userID, tenantID, input)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start authorization: %v", err), http.StatusInternalServerError)
		return
	}

	// Return authorization URL or redirect
	accept := r.Header.Get("Accept")
	if accept == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"authorization_url": authURL}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	} else {
		// Redirect to authorization URL
		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

// Callback handles OAuth callback
// GET /api/v1/oauth/callback/:provider
func (h *OAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	providerKey := chi.URLParam(r, "provider")

	// Get user and tenant from context (if available)
	// Note: User might not be authenticated yet in callback
	userID, _ := ctx.Value("user_id").(string)
	tenantID, _ := ctx.Value("tenant_id").(string)

	// Parse callback parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	input := &oauth.CallbackInput{
		Code:  code,
		State: state,
		Error: errorParam,
	}

	// Handle callback
	conn, err := h.service.HandleCallback(ctx, userID, tenantID, input)
	if err != nil {
		http.Error(w, fmt.Sprintf("OAuth callback failed: %v", err), http.StatusBadRequest)
		return
	}

	// Remove sensitive data
	conn.AccessTokenEncrypted = nil
	conn.AccessTokenNonce = nil
	conn.AccessTokenAuthTag = nil
	conn.AccessTokenEncDEK = nil
	conn.RefreshTokenEncrypted = nil
	conn.RefreshTokenNonce = nil
	conn.RefreshTokenAuthTag = nil
	conn.RefreshTokenEncDEK = nil

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success":    true,
		"provider":   providerKey,
		"connection": conn,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ListConnections lists user's OAuth connections
// GET /api/v1/oauth/connections
func (h *OAuthHandler) ListConnections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user and tenant from context
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok {
		http.Error(w, "Missing tenant context", http.StatusBadRequest)
		return
	}

	connections, err := h.service.ListConnections(ctx, userID, tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list connections: %v", err), http.StatusInternalServerError)
		return
	}

	// Remove sensitive data
	for _, conn := range connections {
		conn.AccessTokenEncrypted = nil
		conn.AccessTokenNonce = nil
		conn.AccessTokenAuthTag = nil
		conn.AccessTokenEncDEK = nil
		conn.RefreshTokenEncrypted = nil
		conn.RefreshTokenNonce = nil
		conn.RefreshTokenAuthTag = nil
		conn.RefreshTokenEncDEK = nil
		conn.RawTokenResponse = nil
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(connections); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetConnection retrieves a specific OAuth connection
// GET /api/v1/oauth/connections/:id
func (h *OAuthHandler) GetConnection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	connectionID := chi.URLParam(r, "id")

	// Get user and tenant from context
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok {
		http.Error(w, "Missing tenant context", http.StatusBadRequest)
		return
	}

	// Get connection from service
	// Note: We need to implement a GetConnectionByID method that validates ownership
	conn, err := h.service.GetConnection(ctx, userID, tenantID, connectionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Connection not found: %v", err), http.StatusNotFound)
		return
	}

	// Remove sensitive data
	conn.AccessTokenEncrypted = nil
	conn.AccessTokenNonce = nil
	conn.AccessTokenAuthTag = nil
	conn.AccessTokenEncDEK = nil
	conn.RefreshTokenEncrypted = nil
	conn.RefreshTokenNonce = nil
	conn.RefreshTokenAuthTag = nil
	conn.RefreshTokenEncDEK = nil
	conn.RawTokenResponse = nil

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(conn); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// RevokeConnection revokes an OAuth connection
// DELETE /api/v1/oauth/connections/:id
func (h *OAuthHandler) RevokeConnection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	connectionID := chi.URLParam(r, "id")

	// Get user and tenant from context
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok {
		http.Error(w, "Missing tenant context", http.StatusBadRequest)
		return
	}

	if err := h.service.RevokeConnection(ctx, userID, tenantID, connectionID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to revoke connection: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TestConnection tests an OAuth connection
// POST /api/v1/oauth/connections/:id/test
func (h *OAuthHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	connectionID := chi.URLParam(r, "id")

	// Get user and tenant from context
	_, ok := ctx.Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, ok = ctx.Value("tenant_id").(string)
	if !ok {
		http.Error(w, "Missing tenant context", http.StatusBadRequest)
		return
	}

	if err := h.service.TestConnection(ctx, connectionID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success": true,
		"message": "Connection test successful",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
