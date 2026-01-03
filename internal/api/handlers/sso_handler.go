package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorax/gorax/internal/sso"
)

// SSOHandler handles SSO-related HTTP requests
type SSOHandler struct {
	ssoService sso.Service
}

// NewSSOHandler creates a new SSO handler
func NewSSOHandler(ssoService sso.Service) *SSOHandler {
	return &SSOHandler{
		ssoService: ssoService,
	}
}

// RegisterRoutes registers SSO routes
func (h *SSOHandler) RegisterRoutes(r chi.Router) {
	r.Route("/sso", func(r chi.Router) {
		// Provider management (authenticated, tenant-scoped)
		r.Post("/providers", h.CreateProvider)
		r.Get("/providers", h.ListProviders)
		r.Get("/providers/{id}", h.GetProvider)
		r.Put("/providers/{id}", h.UpdateProvider)
		r.Delete("/providers/{id}", h.DeleteProvider)

		// SSO login flow (public)
		r.Get("/login/{id}", h.InitiateLogin)
		r.Post("/callback/{id}", h.HandleCallback)
		r.Get("/callback/{id}", h.HandleCallback) // For OIDC redirect
		r.Get("/metadata/{id}", h.GetMetadata)
		r.Post("/acs", h.HandleSAMLAssertion) // SAML Assertion Consumer Service

		// Helper endpoints
		r.Get("/discover", h.DiscoverProvider)
	})
}

// CreateProvider creates a new SSO provider
func (h *SSOHandler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant ID from context
	tenantID, err := getTenantIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user ID from context
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var req sso.CreateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Create provider
	provider, err := h.ssoService.CreateProvider(ctx, tenantID, &req, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create provider: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(provider); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ListProviders lists all SSO providers for a tenant
func (h *SSOHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant ID from context
	tenantID, err := getTenantIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// List providers
	providers, err := h.ssoService.ListProviders(ctx, tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list providers: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(providers); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetProvider retrieves an SSO provider
func (h *SSOHandler) GetProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant ID from context
	tenantID, err := getTenantIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get provider ID from URL
	providerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	// Get provider
	provider, err := h.ssoService.GetProvider(ctx, tenantID, providerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get provider: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(provider); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateProvider updates an SSO provider
func (h *SSOHandler) UpdateProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant ID from context
	tenantID, err := getTenantIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user ID from context
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get provider ID from URL
	providerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	// Parse request
	var req sso.UpdateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Update provider
	provider, err := h.ssoService.UpdateProvider(ctx, tenantID, providerID, &req, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update provider: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(provider); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DeleteProvider deletes an SSO provider
func (h *SSOHandler) DeleteProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant ID from context
	tenantID, err := getTenantIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get provider ID from URL
	providerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	// Delete provider
	if err := h.ssoService.DeleteProvider(ctx, tenantID, providerID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete provider: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// InitiateLogin initiates the SSO login flow
func (h *SSOHandler) InitiateLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get provider ID from URL
	providerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	// Get relay state from query params
	relayState := r.URL.Query().Get("relay_state")

	// Initiate login
	redirectURL, err := h.ssoService.InitiateLogin(ctx, providerID, relayState)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to initiate login: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to IdP
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// HandleCallback handles the SSO callback
func (h *SSOHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get provider ID from URL
	providerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	// Handle callback
	authResp, err := h.ssoService.HandleCallback(ctx, providerID, r)
	if err != nil {
		// Redirect to error page or return error
		http.Error(w, fmt.Sprintf("SSO authentication failed: %v", err), http.StatusUnauthorized)
		return
	}

	// In production, set session cookie and redirect to app
	// For now, return the authentication response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(authResp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleSAMLAssertion handles SAML assertion consumer service
func (h *SSOHandler) HandleSAMLAssertion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse form to get relay state
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get provider ID from relay state or form
	// In production, relay state should contain provider ID
	relayState := r.PostForm.Get("RelayState")

	// For simplicity, assume relay state contains provider ID
	// In production, use a more secure state management
	var providerID uuid.UUID
	if relayState != "" {
		var err error
		providerID, err = uuid.Parse(relayState)
		if err != nil {
			http.Error(w, "Invalid relay state", http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "Relay state required", http.StatusBadRequest)
		return
	}

	// Handle callback (same as HandleCallback)
	authResp, err := h.ssoService.HandleCallback(ctx, providerID, r)
	if err != nil {
		http.Error(w, fmt.Sprintf("SAML authentication failed: %v", err), http.StatusUnauthorized)
		return
	}

	// In production, set session cookie and redirect to app
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(authResp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetMetadata returns the SSO provider metadata (for SAML)
func (h *SSOHandler) GetMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get provider ID from URL
	providerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	// Get metadata
	metadata, err := h.ssoService.GetMetadata(ctx, providerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get metadata: %v", err), http.StatusInternalServerError)
		return
	}

	// Return XML metadata
	w.Header().Set("Content-Type", "application/xml")
	if _, err := w.Write([]byte(metadata)); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

// DiscoverProvider discovers SSO provider by email domain
func (h *SSOHandler) DiscoverProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get email from query params
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email parameter required", http.StatusBadRequest)
		return
	}

	// Extract domain from email
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}
	domain := parts[1]

	// Get provider by domain
	provider, err := h.ssoService.GetProviderByDomain(ctx, domain)
	if err != nil {
		// No SSO provider found - return 404
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"sso_available": false,
		}); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
		return
	}

	// Return provider info
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"sso_available": true,
		"provider_id":   provider.ID,
		"provider_name": provider.Name,
		"provider_type": provider.Type,
		"enforce_sso":   provider.EnforceSSO,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Helper functions to extract context values
func getTenantIDFromContext(ctx context.Context) (uuid.UUID, error) {
	// In production, extract from authenticated context
	// For now, return a dummy UUID
	tenantIDStr, ok := ctx.Value("tenant_id").(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("tenant ID not found in context")
	}
	return uuid.Parse(tenantIDStr)
}

func getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	// In production, extract from authenticated context
	userIDStr, ok := ctx.Value("user_id").(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}
	return uuid.Parse(userIDStr)
}
