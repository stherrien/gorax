package credential

import (
	"encoding/json"
	"strings"
)

const DefaultMask = "***MASKED***"

// Masker handles masking of sensitive credential values in logs and outputs
type Masker struct {
	mask string
}

// NewMasker creates a new masker with default mask string
func NewMasker() *Masker {
	return &Masker{
		mask: DefaultMask,
	}
}

// NewMaskerWithMask creates a new masker with custom mask string
func NewMaskerWithMask(mask string) *Masker {
	return &Masker{
		mask: mask,
	}
}

// MaskString replaces all occurrences of secrets in the input string with the mask
func (m *Masker) MaskString(input string, secrets []string) string {
	if input == "" || len(secrets) == 0 {
		return input
	}

	result := input
	for _, secret := range secrets {
		if secret != "" {
			result = strings.ReplaceAll(result, secret, m.mask)
		}
	}

	return result
}

// MaskJSON recursively masks secrets in a JSON structure
func (m *Masker) MaskJSON(data map[string]interface{}, secrets []string) map[string]interface{} {
	if len(secrets) == 0 {
		return data
	}

	result := make(map[string]interface{})
	for key, value := range data {
		result[key] = m.maskValue(value, secrets)
	}

	return result
}

// maskValue recursively masks a value
func (m *Masker) maskValue(value interface{}, secrets []string) interface{} {
	switch v := value.(type) {
	case string:
		return m.MaskString(v, secrets)
	case map[string]interface{}:
		return m.MaskJSON(v, secrets)
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = m.maskValue(item, secrets)
		}
		return result
	default:
		// Preserve non-string types as-is
		return v
	}
}

// MaskRawJSON masks secrets in raw JSON data
func (m *Masker) MaskRawJSON(data json.RawMessage, secrets []string) (json.RawMessage, error) {
	if len(data) == 0 || len(secrets) == 0 {
		return data, nil
	}

	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}

	masked := m.maskValue(parsed, secrets)

	result, err := json.Marshal(masked)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ExtractSecrets recursively extracts all string values from a data structure
// This is useful for extracting credential values that need to be masked
func (m *Masker) ExtractSecrets(value interface{}) []string {
	var secrets []string

	switch v := value.(type) {
	case string:
		if v != "" {
			secrets = append(secrets, v)
		}
	case map[string]interface{}:
		for _, val := range v {
			secrets = append(secrets, m.ExtractSecrets(val)...)
		}
	case []interface{}:
		for _, item := range v {
			secrets = append(secrets, m.ExtractSecrets(item)...)
		}
	}

	return secrets
}
