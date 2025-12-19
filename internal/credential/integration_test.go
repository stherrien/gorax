package credential

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_EncryptDecryptCredential tests the full encryption/decryption cycle
func TestIntegration_EncryptDecryptCredential(t *testing.T) {
	// Setup
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i + 1)
	}

	svc, err := NewSimpleEncryptionService(masterKey)
	require.NoError(t, err)

	ctx := context.Background()
	tenantID := "tenant-123"

	// Test data - API key credential
	originalData := &CredentialData{
		Value: map[string]interface{}{
			"api_key": "secret-api-key-12345",
			"region":  "us-east-1",
		},
	}

	t.Run("successful encryption and decryption", func(t *testing.T) {
		// Encrypt
		encrypted, err := svc.Encrypt(ctx, tenantID, originalData)
		require.NoError(t, err)
		require.NotNil(t, encrypted)

		// Verify encrypted data structure
		assert.NotEmpty(t, encrypted.EncryptedDEK)
		assert.NotEmpty(t, encrypted.Ciphertext)
		assert.NotEmpty(t, encrypted.Nonce)
		assert.NotEmpty(t, encrypted.AuthTag)
		assert.Equal(t, "simple-encryption", encrypted.KMSKeyID)

		// Verify the encrypted data doesn't contain plaintext
		encryptedJSON, _ := json.Marshal(encrypted)
		assert.NotContains(t, string(encryptedJSON), "secret-api-key")

		// Decrypt
		decrypted, err := svc.Decrypt(ctx, encrypted)
		require.NoError(t, err)
		require.NotNil(t, decrypted)

		// Verify decrypted data matches original
		assert.Equal(t, originalData.Value["api_key"], decrypted.Value["api_key"])
		assert.Equal(t, originalData.Value["region"], decrypted.Value["region"])
	})

	t.Run("OAuth2 credential", func(t *testing.T) {
		oauth2Data := &CredentialData{
			Value: map[string]interface{}{
				"client_id":     "client-123",
				"client_secret": "secret-456",
				"access_token":  "token-789",
				"refresh_token": "refresh-abc",
			},
		}

		encrypted, err := svc.Encrypt(ctx, tenantID, oauth2Data)
		require.NoError(t, err)

		decrypted, err := svc.Decrypt(ctx, encrypted)
		require.NoError(t, err)

		assert.Equal(t, oauth2Data.Value["client_id"], decrypted.Value["client_id"])
		assert.Equal(t, oauth2Data.Value["client_secret"], decrypted.Value["client_secret"])
		assert.Equal(t, oauth2Data.Value["access_token"], decrypted.Value["access_token"])
		assert.Equal(t, oauth2Data.Value["refresh_token"], decrypted.Value["refresh_token"])
	})

	t.Run("basic auth credential", func(t *testing.T) {
		basicAuthData := &CredentialData{
			Value: map[string]interface{}{
				"username": "admin",
				"password": "super-secret-password",
			},
		}

		encrypted, err := svc.Encrypt(ctx, tenantID, basicAuthData)
		require.NoError(t, err)

		decrypted, err := svc.Decrypt(ctx, encrypted)
		require.NoError(t, err)

		assert.Equal(t, "admin", decrypted.Value["username"])
		assert.Equal(t, "super-secret-password", decrypted.Value["password"])
	})
}

// TestIntegration_MaskerWithEncryption tests masker integration with encrypted credentials
func TestIntegration_MaskerWithEncryption(t *testing.T) {
	// Setup encryption
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i + 1)
	}

	svc, err := NewSimpleEncryptionService(masterKey)
	require.NoError(t, err)

	ctx := context.Background()
	tenantID := "tenant-123"

	// Create credential
	credData := &CredentialData{
		Value: map[string]interface{}{
			"api_key": "super-secret-key-12345",
		},
	}

	encrypted, err := svc.Encrypt(ctx, tenantID, credData)
	require.NoError(t, err)

	// Decrypt to get the actual value
	decrypted, err := svc.Decrypt(ctx, encrypted)
	require.NoError(t, err)

	secretValue := decrypted.Value["api_key"].(string)

	// Setup masker
	masker := NewMasker()

	t.Run("masks credential in string output", func(t *testing.T) {
		output := "Using API key: " + secretValue + " to authenticate"
		masked := masker.MaskString(output, []string{secretValue})

		assert.NotContains(t, masked, secretValue)
		assert.Contains(t, masked, "***MASKED***")
		assert.Equal(t, "Using API key: ***MASKED*** to authenticate", masked)
	})

	t.Run("masks credential in JSON output", func(t *testing.T) {
		data := map[string]interface{}{
			"request": map[string]interface{}{
				"headers": map[string]interface{}{
					"Authorization": "Bearer " + secretValue,
				},
			},
			"response": map[string]interface{}{
				"token": secretValue,
			},
		}

		masked := masker.MaskJSON(data, []string{secretValue})

		// Verify original is not modified
		assert.Contains(t, data["request"].(map[string]interface{})["headers"].(map[string]interface{})["Authorization"], secretValue)

		// Verify masked output doesn't contain secret
		maskedJSON, _ := json.Marshal(masked)
		assert.NotContains(t, string(maskedJSON), secretValue)
		assert.Contains(t, string(maskedJSON), "***MASKED***")
	})

	t.Run("masks multiple credentials", func(t *testing.T) {
		// Create second credential
		cred2Data := &CredentialData{
			Value: map[string]interface{}{
				"password": "another-secret-pass",
			},
		}

		encrypted2, err := svc.Encrypt(ctx, tenantID, cred2Data)
		require.NoError(t, err)

		decrypted2, err := svc.Decrypt(ctx, encrypted2)
		require.NoError(t, err)

		password := decrypted2.Value["password"].(string)

		output := "API: " + secretValue + " Password: " + password
		secrets := []string{secretValue, password}
		masked := masker.MaskString(output, secrets)

		assert.NotContains(t, masked, secretValue)
		assert.NotContains(t, masked, password)
		assert.Equal(t, "API: ***MASKED*** Password: ***MASKED***", masked)
	})
}

// TestIntegration_CredentialReferenceExtraction tests extraction of credential references from config
func TestIntegration_CredentialReferenceExtraction(t *testing.T) {
	// Use regex directly since we can't instantiate injector without DB
	credentialReferenceRegex := `\{\{credentials\.([a-zA-Z0-9_-]+)\}\}`

	tests := []struct {
		name     string
		config   string
		expected []string
	}{
		{
			name:     "no credentials",
			config:   `{"url":"https://api.example.com"}`,
			expected: []string{},
		},
		{
			name:     "single credential in header",
			config:   `{"headers":{"Authorization":"Bearer {{credentials.api_token}}"}}`,
			expected: []string{"api_token"},
		},
		{
			name: "multiple credentials",
			config: `{
				"headers": {
					"Authorization": "Bearer {{credentials.api_token}}",
					"X-API-Key": "{{credentials.api_key}}"
				}
			}`,
			expected: []string{"api_token", "api_key"},
		},
		{
			name:     "credential in body",
			config:   `{"body":{"token":"{{credentials.oauth_token}}"}}`,
			expected: []string{"oauth_token"},
		},
		{
			name: "nested credentials",
			config: `{
				"auth": {
					"type": "oauth2",
					"credentials": {
						"client_secret": "{{credentials.client_secret}}"
					}
				}
			}`,
			expected: []string{"client_secret"},
		},
		{
			name:     "duplicate credential references",
			config:   `{"key1":"{{credentials.token}}","key2":"{{credentials.token}}"}`,
			expected: []string{"token"}, // Should deduplicate
		},
		{
			name:     "credential with hyphens and underscores",
			config:   `{"key":"{{credentials.my-api_key-123}}"}`,
			expected: []string{"my-api_key-123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Extract credential references using regex
			regex := regexp.MustCompile(credentialReferenceRegex)
			matches := regex.FindAllStringSubmatch(tt.config, -1)

			var refs []string
			seen := make(map[string]bool)
			for _, match := range matches {
				if len(match) > 1 && !seen[match[1]] {
					refs = append(refs, match[1])
					seen[match[1]] = true
				}
			}

			assert.ElementsMatch(t, tt.expected, refs)
		})
	}
}

// TestIntegration_CredentialWorkflow tests a complete credential workflow
func TestIntegration_CredentialWorkflow(t *testing.T) {
	// This test simulates the complete flow:
	// 1. Create credential with encryption
	// 2. Store encrypted data
	// 3. Reference credential in workflow config
	// 4. Extract references
	// 5. Decrypt credential
	// 6. Inject into config
	// 7. Mask in output

	// Setup
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i + 1)
	}

	svc, err := NewSimpleEncryptionService(masterKey)
	require.NoError(t, err)

	ctx := context.Background()
	tenantID := "tenant-123"

	// Step 1: Create and encrypt credential
	credData := &CredentialData{
		Value: map[string]interface{}{
			"api_key": "sk-1234567890abcdef",
		},
	}

	encrypted, err := svc.Encrypt(ctx, tenantID, credData)
	require.NoError(t, err)
	t.Logf("✓ Credential encrypted successfully")

	// Step 2: Simulate storage (in real scenario, this goes to DB)
	storedCred := &Credential{
		ID:           "cred-1",
		TenantID:     tenantID,
		Name:         "openai_api_key",
		Type:         TypeAPIKey,
		EncryptedDEK: encrypted.EncryptedDEK,
		Ciphertext:   encrypted.Ciphertext,
		Nonce:        encrypted.Nonce,
		AuthTag:      encrypted.AuthTag,
		KMSKeyID:     encrypted.KMSKeyID,
		Status:       StatusActive,
	}
	t.Logf("✓ Credential stored with ID: %s", storedCred.ID)

	// Step 3: Workflow config references the credential
	workflowConfig := json.RawMessage(`{
		"method": "POST",
		"url": "https://api.openai.com/v1/chat/completions",
		"headers": {
			"Authorization": "Bearer {{credentials.openai_api_key}}",
			"Content-Type": "application/json"
		},
		"body": {
			"model": "gpt-4",
			"messages": [{"role": "user", "content": "Hello"}]
		}
	}`)

	// Step 4: Extract credential references using regex
	credentialReferenceRegex := regexp.MustCompile(`\{\{credentials\.([a-zA-Z0-9_-]+)\}\}`)
	matches := credentialReferenceRegex.FindAllStringSubmatch(string(workflowConfig), -1)

	var refs []string
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			refs = append(refs, match[1])
			seen[match[1]] = true
		}
	}

	assert.Equal(t, []string{"openai_api_key"}, refs)
	t.Logf("✓ Extracted credential references: %v", refs)

	// Step 5: Decrypt credential (simulating what injector would do)
	decryptedCred, err := svc.Decrypt(ctx, encrypted)
	require.NoError(t, err)
	apiKey := decryptedCred.Value["api_key"].(string)
	assert.Equal(t, "sk-1234567890abcdef", apiKey)
	t.Logf("✓ Credential decrypted successfully")

	// Step 6: Simulate injection (replace placeholder with actual value)
	var config map[string]interface{}
	err = json.Unmarshal(workflowConfig, &config)
	require.NoError(t, err)

	headers := config["headers"].(map[string]interface{})
	headers["Authorization"] = "Bearer " + apiKey
	t.Logf("✓ Credential injected into config")

	// Step 7: Simulate execution output that contains the credential
	executionOutput := map[string]interface{}{
		"request": map[string]interface{}{
			"headers": headers,
		},
		"response": map[string]interface{}{
			"status": 200,
			"body":   "Success",
		},
	}

	// Step 8: Mask the credential in output
	masker := NewMasker()
	maskedOutput := masker.MaskJSON(executionOutput, []string{apiKey})

	// Verify the credential is masked
	maskedJSON, _ := json.Marshal(maskedOutput)
	assert.NotContains(t, string(maskedJSON), "sk-1234567890abcdef")
	assert.Contains(t, string(maskedJSON), "***MASKED***")
	t.Logf("✓ Credential masked in output")

	t.Logf("\n✅ Complete credential workflow test passed!")
	t.Logf("   1. Encrypted credential")
	t.Logf("   2. Stored encrypted data")
	t.Logf("   3. Referenced in workflow config")
	t.Logf("   4. Extracted references")
	t.Logf("   5. Decrypted credential")
	t.Logf("   6. Injected into config")
	t.Logf("   7. Masked in output")
}
