package google

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// extractString extracts a string value from a nested map using dot notation
func extractString(data map[string]interface{}, path string) (string, error) {
	keys := parsePath(path)
	current := data

	for i, key := range keys {
		if i == len(keys)-1 {
			if val, ok := current[key]; ok {
				if str, ok := val.(string); ok {
					return str, nil
				}
				return "", fmt.Errorf("value at '%s' is not a string", path)
			}
			return "", fmt.Errorf("key '%s' not found in context", path)
		}

		if val, ok := current[key]; ok {
			if m, ok := val.(map[string]interface{}); ok {
				current = m
				continue
			}
			return "", fmt.Errorf("value at '%s' is not a map", key)
		}
		return "", fmt.Errorf("key '%s' not found in context", key)
	}

	return "", fmt.Errorf("path '%s' is empty", path)
}

// parsePath splits a dot-notation path into keys
func parsePath(path string) []string {
	return strings.Split(path, ".")
}

// createOAuth2Token creates an OAuth2 token from credential data
func createOAuth2Token(credentialValue map[string]interface{}) (*oauth2.Token, error) {
	accessToken, ok := credentialValue["access_token"].(string)
	if !ok || accessToken == "" {
		return nil, fmt.Errorf("access_token not found in credential")
	}

	token := &oauth2.Token{
		AccessToken: accessToken,
	}

	// Add refresh token if present
	if refreshToken, ok := credentialValue["refresh_token"].(string); ok && refreshToken != "" {
		token.RefreshToken = refreshToken
	}

	// Add token type if present
	if tokenType, ok := credentialValue["token_type"].(string); ok && tokenType != "" {
		token.TokenType = tokenType
	} else {
		token.TokenType = "Bearer"
	}

	return token, nil
}

// createServiceAccountClient creates a Google API client using service account credentials
func createServiceAccountClient(ctx context.Context, credentialValue map[string]interface{}, scopes []string) (option.ClientOption, error) {
	// Try to get service account JSON
	var serviceAccountJSON []byte

	// Check if credential_json is a string (JSON string)
	if credJSON, ok := credentialValue["credential_json"].(string); ok && credJSON != "" {
		serviceAccountJSON = []byte(credJSON)
	} else if credJSON, ok := credentialValue["service_account_json"].(string); ok && credJSON != "" {
		serviceAccountJSON = []byte(credJSON)
	} else {
		// Try to marshal the entire credential value if it looks like a service account
		if _, hasType := credentialValue["type"]; hasType {
			var err error
			serviceAccountJSON, err = json.Marshal(credentialValue)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal service account credentials: %w", err)
			}
		} else {
			return nil, fmt.Errorf("service account credentials not found in credential")
		}
	}

	// Create credentials from JSON
	creds, err := google.CredentialsFromJSON(ctx, serviceAccountJSON, scopes...)
	if err != nil {
		return nil, fmt.Errorf("failed to create credentials from service account: %w", err)
	}

	return option.WithCredentials(creds), nil
}

// createOAuth2Client creates a Google API client using OAuth2 credentials
func createOAuth2Client(ctx context.Context, token *oauth2.Token) option.ClientOption {
	tokenSource := oauth2.StaticTokenSource(token)
	return option.WithTokenSource(tokenSource)
}
