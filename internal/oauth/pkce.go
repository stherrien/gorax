package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const (
	// PKCEVerifierLength is the length of the PKCE code verifier (43-128 chars)
	PKCEVerifierLength = 64
)

// GenerateState generates a cryptographically secure random state string
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GeneratePKCEVerifier generates a PKCE code verifier
// Returns a random base64-url encoded string of 43-128 characters
func GeneratePKCEVerifier() (string, error) {
	b := make([]byte, PKCEVerifierLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate PKCE verifier: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GeneratePKCEChallenge generates a PKCE code challenge from a verifier
// Uses SHA256 hash and base64-url encoding
func GeneratePKCEChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
