package credential

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
)

// credentialReferenceRegex matches {{credentials.name}} patterns
var credentialReferenceRegex = regexp.MustCompile(`\{\{credentials\.([a-zA-Z0-9_-]+)\}\}`)

// RepositoryInterface defines the interface for credential repository operations
type RepositoryInterface interface {
	ValidateAndGet(ctx context.Context, tenantID, name string) (*Credential, error)
	UpdateAccessTime(ctx context.Context, tenantID, credentialID string) error
	LogAccess(ctx context.Context, log *AccessLog) error
}

// EncryptionServiceInterface defines the interface for encryption operations
type EncryptionServiceInterface interface {
	Encrypt(ctx context.Context, tenantID string, data *CredentialData) (*EncryptedSecret, error)
	Decrypt(ctx context.Context, encryptedData, encryptedKey []byte) (*CredentialData, error)
}

// Injector handles credential injection into workflow actions
type Injector struct {
	repo       RepositoryInterface
	encryption EncryptionServiceInterface
	masker     *Masker
}

// NewInjector creates a new credential injector
func NewInjector(repo RepositoryInterface, encryption EncryptionServiceInterface) *Injector {
	return &Injector{
		repo:       repo,
		encryption: encryption,
		masker:     NewMasker(),
	}
}

// InjectionContext holds context for credential injection
type InjectionContext struct {
	TenantID    string
	WorkflowID  string
	ExecutionID string
	AccessedBy  string
}

// InjectResult holds the result of credential injection
type InjectResult struct {
	Config json.RawMessage // Config with credentials injected
	Values []string        // Decrypted credential values for masking
}

// InjectCredentials extracts credential references from config and injects decrypted values
func (i *Injector) InjectCredentials(ctx context.Context, config json.RawMessage, injCtx *InjectionContext) (*InjectResult, error) {
	// Parse config to extract credential references
	credentialRefs, err := i.ExtractCredentialReferences(config)
	if err != nil {
		return nil, fmt.Errorf("failed to extract credential references: %w", err)
	}

	// If no credentials referenced, return original config
	if len(credentialRefs) == 0 {
		return &InjectResult{
			Config: config,
			Values: []string{},
		}, nil
	}

	// Fetch and decrypt all referenced credentials
	credentials := make(map[string]string)
	var values []string

	for _, credName := range credentialRefs {
		value, err := i.getCredentialValue(ctx, injCtx.TenantID, credName, injCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to get credential '%s': %w", credName, err)
		}
		credentials[credName] = value
		values = append(values, value)
	}

	// Inject credentials into config
	injectedConfig, err := i.injectValues(config, credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to inject credentials: %w", err)
	}

	return &InjectResult{
		Config: injectedConfig,
		Values: values,
	}, nil
}

// ExtractCredentialReferences extracts unique credential names from config
func (i *Injector) ExtractCredentialReferences(config json.RawMessage) ([]string, error) {
	// Convert to string to search for patterns
	configStr := string(config)

	// Find all credential references
	matches := credentialReferenceRegex.FindAllStringSubmatch(configStr, -1)

	// Extract unique credential names
	credNames := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			credNames[match[1]] = true
		}
	}

	// Convert map to slice
	var result []string
	for name := range credNames {
		result = append(result, name)
	}

	return result, nil
}

// getCredentialValue retrieves and decrypts a credential value
func (i *Injector) getCredentialValue(ctx context.Context, tenantID, name string, injCtx *InjectionContext) (string, error) {
	// Validate and get credential
	cred, err := i.repo.ValidateAndGet(ctx, tenantID, name)
	if err != nil {
		// Log failed access
		_ = i.repo.LogAccess(ctx, &AccessLog{
			TenantID:     tenantID,
			CredentialID: name,
			AccessedBy:   injCtx.AccessedBy,
			AccessType:   AccessTypeRead,
			Success:      false,
			ErrorMessage: err.Error(),
		})
		return "", err
	}

	// Decrypt the value using envelope encryption
	// The Ciphertext field contains the encrypted data (nonce + ciphertext + tag combined)
	credData, err := i.encryption.Decrypt(ctx, cred.Ciphertext, cred.EncryptedDEK)
	if err != nil {
		// Log failed decryption
		_ = i.repo.LogAccess(ctx, &AccessLog{
			TenantID:     tenantID,
			CredentialID: cred.ID,
			AccessedBy:   injCtx.AccessedBy,
			AccessType:   AccessTypeRead,
			Success:      false,
			ErrorMessage: err.Error(),
		})
		return "", fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Extract the actual credential value
	// For injection, we typically need a string representation
	// The credData.Value map might have a "key" or "token" field depending on the credential type
	value := i.extractCredentialValue(credData.Value)

	// Update access time
	_ = i.repo.UpdateAccessTime(ctx, tenantID, cred.ID)

	// Log successful access
	_ = i.repo.LogAccess(ctx, &AccessLog{
		TenantID:     tenantID,
		CredentialID: cred.ID,
		AccessedBy:   injCtx.AccessedBy,
		AccessType:   AccessTypeRead,
		Success:      true,
	})

	return value, nil
}

// extractCredentialValue extracts a usable string value from credential data
// For different credential types, the value might be in different fields
func (i *Injector) extractCredentialValue(value map[string]interface{}) string {
	// Try common field names based on credential type
	if v, ok := value["api_key"]; ok {
		return fmt.Sprintf("%v", v)
	}
	if v, ok := value["token"]; ok {
		return fmt.Sprintf("%v", v)
	}
	if v, ok := value["secret"]; ok {
		return fmt.Sprintf("%v", v)
	}
	if v, ok := value["key"]; ok {
		return fmt.Sprintf("%v", v)
	}
	if v, ok := value["password"]; ok {
		return fmt.Sprintf("%v", v)
	}

	// If none of the common fields found, marshal the entire value as JSON
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}

// injectValues replaces credential references with actual values
func (i *Injector) injectValues(config json.RawMessage, credentials map[string]string) (json.RawMessage, error) {
	// Parse config to interface{}
	var data interface{}
	if err := json.Unmarshal(config, &data); err != nil {
		return nil, err
	}

	// Recursively inject credentials
	injected := i.injectValue(data, credentials)

	// Marshal back to JSON
	result, err := json.Marshal(injected)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// injectValue recursively injects credentials into a value
func (i *Injector) injectValue(value interface{}, credentials map[string]string) interface{} {
	switch v := value.(type) {
	case string:
		// Replace credential references in strings
		return i.replaceCredentialReferences(v, credentials)

	case map[string]interface{}:
		// Recursively inject into map values
		result := make(map[string]interface{})
		for key, val := range v {
			result[key] = i.injectValue(val, credentials)
		}
		return result

	case []interface{}:
		// Recursively inject into array elements
		result := make([]interface{}, len(v))
		for idx, val := range v {
			result[idx] = i.injectValue(val, credentials)
		}
		return result

	default:
		// Return non-string types as-is
		return v
	}
}

// replaceCredentialReferences replaces {{credentials.name}} with actual values
func (i *Injector) replaceCredentialReferences(input string, credentials map[string]string) string {
	return credentialReferenceRegex.ReplaceAllStringFunc(input, func(match string) string {
		// Extract credential name
		submatch := credentialReferenceRegex.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match // Return original if no match
		}

		credName := submatch[1]

		// Look up credential value
		if value, exists := credentials[credName]; exists {
			return value
		}

		// Return original if credential not found (should not happen if extraction worked correctly)
		return match
	})
}

// MaskOutput masks credential values in output data
func (i *Injector) MaskOutput(data interface{}, credentialValues []string) interface{} {
	return i.masker.maskValue(data, credentialValues)
}

// MaskOutputJSON masks credential values in JSON output
func (i *Injector) MaskOutputJSON(data json.RawMessage, credentialValues []string) (json.RawMessage, error) {
	return i.masker.MaskRawJSON(data, credentialValues)
}
