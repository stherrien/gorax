package actions

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/gorax/gorax/internal/validation"
)

var (
	// interpolationRegex matches {{expression}} patterns
	interpolationRegex = regexp.MustCompile(`\{\{([^}]+)\}\}`)
	// arrayIndexRegex matches array[index] patterns
	arrayIndexRegex = regexp.MustCompile(`^(.+)\[(\d+)\]$`)
)

// InterpolateString replaces {{path.to.value}} with actual values from context
// Supports JSONPath-like syntax: steps.http-1.body.users[0].name
func InterpolateString(template string, context map[string]interface{}) string {
	return interpolationRegex.ReplaceAllStringFunc(template, func(match string) string {
		// Extract expression from {{...}}
		expression := strings.TrimSpace(match[2 : len(match)-2])

		value, err := GetValueByPath(context, expression)
		if err != nil {
			// Return original if path not found
			return match
		}

		// Convert value to string
		return toString(value)
	})
}

// InterpolateJSON recursively interpolates values in a JSON structure
func InterpolateJSON(data json.RawMessage, context map[string]interface{}) interface{} {
	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil
	}
	return interpolateValue(parsed, context)
}

// interpolateValue recursively interpolates a value
func interpolateValue(value interface{}, context map[string]interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return InterpolateString(v, context)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			result[key] = interpolateValue(val, context)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = interpolateValue(val, context)
		}
		return result
	default:
		return v
	}
}

// GetValueByPath retrieves a value from a nested map using dot notation
// Supports array indexing: "steps.node1.body.users[0].name"
func GetValueByPath(data map[string]interface{}, path string) (interface{}, error) {
	if path == "" {
		return data, nil
	}

	current := interface{}(data)
	parts := splitPath(path)

	for i, part := range parts {
		// Check for array indexing
		if matches := arrayIndexRegex.FindStringSubmatch(part); matches != nil {
			// Handle array access like "users[0]"
			arrayKey := matches[1]
			indexStr := matches[2]

			// Get the array first
			switch v := current.(type) {
			case map[string]interface{}:
				current = v[arrayKey]
			default:
				return nil, fmt.Errorf("cannot access key '%s' on non-object type", arrayKey)
			}

			// Then access the index with overflow protection
			switch arr := current.(type) {
			case []interface{}:
				index, valid := validation.ParseArrayIndex(indexStr, len(arr))
				if !valid {
					return nil, fmt.Errorf("invalid or out of bounds array index '%s'", indexStr)
				}
				current = arr[index]
			default:
				return nil, fmt.Errorf("cannot index non-array type at '%s'", arrayKey)
			}
			continue
		}

		// Regular object property access
		switch v := current.(type) {
		case map[string]interface{}:
			var exists bool
			current, exists = v[part]
			if !exists {
				return nil, fmt.Errorf("key '%s' not found at path position %d", part, i)
			}
		default:
			return nil, fmt.Errorf("cannot traverse into non-object type at '%s'", part)
		}
	}

	return current, nil
}

// splitPath splits a path string by dots, handling escaped dots
func splitPath(path string) []string {
	var parts []string
	var current strings.Builder
	escaped := false

	for i := 0; i < len(path); i++ {
		char := path[i]

		if char == '\\' && i+1 < len(path) && path[i+1] == '.' {
			// Escaped dot
			current.WriteByte('.')
			i++ // Skip next char
			escaped = true
			continue
		}

		if char == '.' && !escaped {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(char)
		}
		escaped = false
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// toString converts a value to its string representation
func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case nil:
		return ""
	default:
		// For complex types, marshal to JSON
		if bytes, err := json.Marshal(v); err == nil {
			return string(bytes)
		}
		return fmt.Sprintf("%v", v)
	}
}

// DeepCopy creates a deep copy of a value
func DeepCopy(src interface{}) (interface{}, error) {
	if src == nil {
		return nil, nil
	}

	// Use JSON marshaling for deep copy
	data, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}

	var dst interface{}
	if err := json.Unmarshal(data, &dst); err != nil {
		return nil, err
	}

	return dst, nil
}
