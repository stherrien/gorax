package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/config"
	"github.com/gorax/gorax/internal/user"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userService  *user.Service
	kratosConfig config.KratosConfig
	logger       *slog.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService *user.Service, kratosConfig config.KratosConfig, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		userService:  userService,
		kratosConfig: kratosConfig,
		logger:       logger,
	}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TenantID string `json:"tenant_id"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// PasswordResetRequest represents a password reset request
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// PasswordResetConfirm represents a password reset confirmation
type PasswordResetConfirm struct {
	Code     string `json:"code"`
	Password string `json:"password"`
}

// VerificationRequest represents an email verification request
type VerificationRequest struct {
	Email string `json:"email"`
}

// VerificationConfirm represents an email verification confirmation
type VerificationConfirm struct {
	Code string `json:"code"`
}

// InitiateRegistration initiates the registration flow with Kratos
func (h *AuthHandler) InitiateRegistration(w http.ResponseWriter, r *http.Request) {
	// Create a new registration flow with Kratos
	flowURL := fmt.Sprintf("%s/self-service/registration/api", h.kratosConfig.PublicURL)

	req, err := http.NewRequest("GET", flowURL, nil)
	if err != nil {
		h.logger.Error("failed to create registration flow request", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("failed to initiate registration flow", "error", err)
		http.Error(w, "failed to initiate registration", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Forward the response from Kratos
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"flow_id": resp.Header.Get("X-Flow-Id"),
		"ui":      resp.Header.Get("Location"),
	})
}

// Register handles user registration via Kratos
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Submit registration to Kratos
	flowID := r.URL.Query().Get("flow")
	if flowID == "" {
		http.Error(w, "missing flow parameter", http.StatusBadRequest)
		return
	}

	registrationURL := fmt.Sprintf("%s/self-service/registration?flow=%s", h.kratosConfig.PublicURL, flowID)

	payload := map[string]interface{}{
		"method": "password",
		"password": req.Password,
		"traits": map[string]interface{}{
			"email":     req.Email,
			"tenant_id": req.TenantID,
		},
	}

	payloadBytes, _ := json.Marshal(payload)
	kratosReq, err := http.NewRequest("POST", registrationURL, nil)
	if err != nil {
		h.logger.Error("failed to create kratos request", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	kratosReq.Header.Set("Content-Type", "application/json")
	kratosReq.Body = http.NoBody

	client := &http.Client{}
	resp, err := client.Do(kratosReq)
	if err != nil {
		h.logger.Error("failed to register with kratos", "error", err)
		http.Error(w, "registration failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Parse response from Kratos
	var kratosResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&kratosResp); err != nil {
		h.logger.Error("failed to decode kratos response", "error", err)
		http.Error(w, "registration failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	json.NewEncoder(w).Encode(kratosResp)

	_ = payloadBytes // Use the variable to avoid unused warning
}

// InitiateLogin initiates the login flow with Kratos
func (h *AuthHandler) InitiateLogin(w http.ResponseWriter, r *http.Request) {
	flowURL := fmt.Sprintf("%s/self-service/login/api", h.kratosConfig.PublicURL)

	req, err := http.NewRequest("GET", flowURL, nil)
	if err != nil {
		h.logger.Error("failed to create login flow request", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("failed to initiate login flow", "error", err)
		http.Error(w, "failed to initiate login", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var flowResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&flowResp); err != nil {
		h.logger.Error("failed to decode flow response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flowResp)
}

// Login handles user login via Kratos
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	flowID := r.URL.Query().Get("flow")
	if flowID == "" {
		http.Error(w, "missing flow parameter", http.StatusBadRequest)
		return
	}

	loginURL := fmt.Sprintf("%s/self-service/login?flow=%s", h.kratosConfig.PublicURL, flowID)

	payload := map[string]interface{}{
		"method":     "password",
		"password":   req.Password,
		"identifier": req.Email,
	}

	payloadBytes, _ := json.Marshal(payload)
	kratosReq, err := http.NewRequest("POST", loginURL, nil)
	if err != nil {
		h.logger.Error("failed to create kratos request", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	kratosReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(kratosReq)
	if err != nil {
		h.logger.Error("failed to login with kratos", "error", err)
		http.Error(w, "login failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var kratosResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&kratosResp); err != nil {
		h.logger.Error("failed to decode kratos response", "error", err)
		http.Error(w, "login failed", http.StatusInternalServerError)
		return
	}

	// Copy session cookie from Kratos response with security attributes
	if cookies := resp.Cookies(); len(cookies) > 0 {
		for _, cookie := range cookies {
			// Ensure security attributes are set
			cookie.Secure = true
			cookie.HttpOnly = true
			cookie.SameSite = http.SameSiteStrictMode
			http.SetCookie(w, cookie)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	json.NewEncoder(w).Encode(kratosResp)

	_ = payloadBytes
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get session token
	sessionToken := extractSessionToken(r)
	if sessionToken == "" {
		http.Error(w, "no active session", http.StatusUnauthorized)
		return
	}

	// Create logout flow
	logoutURL := fmt.Sprintf("%s/self-service/logout/api", h.kratosConfig.PublicURL)
	req, err := http.NewRequest("GET", logoutURL, nil)
	if err != nil {
		h.logger.Error("failed to create logout request", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	req.Header.Set("X-Session-Token", sessionToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("failed to logout with kratos", "error", err)
		http.Error(w, "logout failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Clear session cookie with security attributes
	http.SetCookie(w, &http.Cookie{
		Name:     "ory_kratos_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	w.WriteHeader(http.StatusNoContent)
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	authUser := middleware.GetUser(r)
	if authUser == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch full user details from database
	user, err := h.userService.GetUserByKratosIdentityID(r.Context(), authUser.ID)
	if err != nil {
		// User might not be synced yet, return basic info from Kratos
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(authUser)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// RequestPasswordReset initiates password reset flow
func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Create recovery flow
	recoveryURL := fmt.Sprintf("%s/self-service/recovery/api", h.kratosConfig.PublicURL)
	kratosReq, err := http.NewRequest("GET", recoveryURL, nil)
	if err != nil {
		h.logger.Error("failed to create recovery request", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(kratosReq)
	if err != nil {
		h.logger.Error("failed to initiate recovery", "error", err)
		http.Error(w, "failed to initiate password reset", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var flowResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&flowResp); err != nil {
		h.logger.Error("failed to decode flow response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flowResp)
}

// ConfirmPasswordReset confirms password reset
func (h *AuthHandler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetConfirm
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	flowID := r.URL.Query().Get("flow")
	if flowID == "" {
		http.Error(w, "missing flow parameter", http.StatusBadRequest)
		return
	}

	// Submit recovery code to Kratos
	recoveryURL := fmt.Sprintf("%s/self-service/recovery?flow=%s", h.kratosConfig.PublicURL, flowID)

	payload := map[string]interface{}{
		"method": "code",
		"code":   req.Code,
	}

	payloadBytes, _ := json.Marshal(payload)
	kratosReq, err := http.NewRequest("POST", recoveryURL, nil)
	if err != nil {
		h.logger.Error("failed to create kratos request", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	kratosReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(kratosReq)
	if err != nil {
		h.logger.Error("failed to confirm password reset", "error", err)
		http.Error(w, "password reset failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var kratosResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&kratosResp); err != nil {
		h.logger.Error("failed to decode kratos response", "error", err)
		http.Error(w, "password reset failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	json.NewEncoder(w).Encode(kratosResp)

	_ = payloadBytes
}

// RequestEmailVerification initiates email verification flow
func (h *AuthHandler) RequestEmailVerification(w http.ResponseWriter, r *http.Request) {
	// Create verification flow
	verificationURL := fmt.Sprintf("%s/self-service/verification/api", h.kratosConfig.PublicURL)
	kratosReq, err := http.NewRequest("GET", verificationURL, nil)
	if err != nil {
		h.logger.Error("failed to create verification request", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(kratosReq)
	if err != nil {
		h.logger.Error("failed to initiate verification", "error", err)
		http.Error(w, "failed to initiate email verification", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var flowResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&flowResp); err != nil {
		h.logger.Error("failed to decode flow response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flowResp)
}

// ConfirmEmailVerification confirms email verification
func (h *AuthHandler) ConfirmEmailVerification(w http.ResponseWriter, r *http.Request) {
	var req VerificationConfirm
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	flowID := r.URL.Query().Get("flow")
	if flowID == "" {
		http.Error(w, "missing flow parameter", http.StatusBadRequest)
		return
	}

	// Submit verification code to Kratos
	verificationURL := fmt.Sprintf("%s/self-service/verification?flow=%s", h.kratosConfig.PublicURL, flowID)

	payload := map[string]interface{}{
		"method": "code",
		"code":   req.Code,
	}

	payloadBytes, _ := json.Marshal(payload)
	kratosReq, err := http.NewRequest("POST", verificationURL, nil)
	if err != nil {
		h.logger.Error("failed to create kratos request", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	kratosReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(kratosReq)
	if err != nil {
		h.logger.Error("failed to confirm email verification", "error", err)
		http.Error(w, "email verification failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var kratosResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&kratosResp); err != nil {
		h.logger.Error("failed to decode kratos response", "error", err)
		http.Error(w, "email verification failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	json.NewEncoder(w).Encode(kratosResp)

	_ = payloadBytes
}

// KratosWebhook handles webhooks from Kratos for identity lifecycle events
func (h *AuthHandler) KratosWebhook(w http.ResponseWriter, r *http.Request) {
	// Verify webhook secret
	secret := r.Header.Get("X-Webhook-Secret")
	expectedSecret := os.Getenv("KRATOS_WEBHOOK_SECRET")
	if expectedSecret == "" {
		expectedSecret = "YOUR_WEBHOOK_SECRET" // Default for development
	}

	if secret != expectedSecret {
		h.logger.Warn("invalid webhook secret")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var webhook user.KratosIdentityWebhook
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		h.logger.Error("failed to decode webhook payload", "error", err)
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	h.logger.Info("received kratos webhook",
		"identity_id", webhook.IdentityID,
		"email", webhook.Email,
	)

	// Sync user to database
	_, err := h.userService.SyncFromKratosWebhook(r.Context(), webhook)
	if err != nil {
		h.logger.Error("failed to sync user from webhook", "error", err)
		http.Error(w, "failed to sync user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

func extractSessionToken(r *http.Request) string {
	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}

	// Try session cookie
	cookie, err := r.Cookie("ory_kratos_session")
	if err == nil {
		return cookie.Value
	}

	return ""
}
