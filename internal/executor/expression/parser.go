package expression

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	// templateRegex matches {{expression}} patterns
	templateRegex = regexp.MustCompile(`\{\{([^}]+)\}\}`)
)

// Expression represents a parsed expression
type Expression struct {
	Raw        string
	IsTemplate bool   // true if wrapped in {{...}}
	Content    string // the actual expression content without {{}}
}

// Parser handles parsing of condition expressions
type Parser struct{}

// NewParser creates a new expression parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses an expression string
// Supports both template syntax {{expr}} and raw expressions
func (p *Parser) Parse(expr string) (*Expression, error) {
	if expr == "" {
		return nil, fmt.Errorf("empty expression")
	}

	expr = strings.TrimSpace(expr)

	// Check if expression is wrapped in {{...}}
	if matches := templateRegex.FindStringSubmatch(expr); matches != nil {
		return &Expression{
			Raw:        expr,
			IsTemplate: true,
			Content:    strings.TrimSpace(matches[1]),
		}, nil
	}

	// Raw expression without template syntax
	return &Expression{
		Raw:        expr,
		IsTemplate: false,
		Content:    expr,
	}, nil
}

// ExtractPaths extracts all variable paths from an expression
// Returns paths like "steps.step1.output", "trigger.body.field", etc.
func (p *Parser) ExtractPaths(expr string) []string {
	var paths []string
	seen := make(map[string]bool)

	// Find all {{...}} patterns
	matches := templateRegex.FindAllStringSubmatch(expr, -1)
	for _, match := range matches {
		if len(match) > 1 {
			content := strings.TrimSpace(match[1])
			// Extract variable paths (looking for dotted paths)
			vars := extractVariablePaths(content)
			for _, v := range vars {
				if !seen[v] {
					paths = append(paths, v)
					seen[v] = true
				}
			}
		}
	}

	// Also check raw expression if no templates found
	if len(matches) == 0 {
		vars := extractVariablePaths(expr)
		for _, v := range vars {
			if !seen[v] {
				paths = append(paths, v)
				seen[v] = true
			}
		}
	}

	return paths
}

// extractVariablePaths extracts variable paths from expression content
// Looks for patterns like: steps.step1.output, trigger.body.field, env.tenant_id
func extractVariablePaths(expr string) []string {
	var paths []string

	// Regex to match variable paths: word.word.word...
	// Handles array indexing: steps.step1.output[0].name
	pathRegex := regexp.MustCompile(`\b((?:steps|trigger|env)(?:\.[a-zA-Z_][a-zA-Z0-9_]*(?:\[\d+\])?)+)`)
	matches := pathRegex.FindAllStringSubmatch(expr, -1)

	for _, match := range matches {
		if len(match) > 1 {
			paths = append(paths, match[1])
		}
	}

	return paths
}

// ValidateExpression performs basic syntax validation
func (p *Parser) ValidateExpression(expr string) error {
	if expr == "" {
		return fmt.Errorf("expression cannot be empty")
	}

	parsed, err := p.Parse(expr)
	if err != nil {
		return err
	}

	// Check for balanced parentheses
	if !isBalanced(parsed.Content, '(', ')') {
		return fmt.Errorf("unbalanced parentheses in expression")
	}

	// Check for balanced brackets
	if !isBalanced(parsed.Content, '[', ']') {
		return fmt.Errorf("unbalanced brackets in expression")
	}

	// Check for valid operators (basic check)
	invalidChars := []string{"{{", "}}", ";;", ".."}
	for _, chars := range invalidChars {
		if strings.Contains(parsed.Content, chars) {
			return fmt.Errorf("invalid characters in expression: %s", chars)
		}
	}

	return nil
}

// isBalanced checks if brackets/parentheses are balanced
func isBalanced(expr string, open, close rune) bool {
	count := 0
	for _, ch := range expr {
		if ch == open {
			count++
		} else if ch == close {
			count--
			if count < 0 {
				return false
			}
		}
	}
	return count == 0
}

// BuildContext creates an evaluation context from execution data
func BuildContext(trigger map[string]interface{}, steps map[string]interface{}, env map[string]interface{}) map[string]interface{} {
	context := make(map[string]interface{})

	if trigger != nil {
		context["trigger"] = trigger
	} else {
		context["trigger"] = make(map[string]interface{})
	}

	if steps != nil {
		context["steps"] = steps
	} else {
		context["steps"] = make(map[string]interface{})
	}

	if env != nil {
		context["env"] = env
	} else {
		context["env"] = make(map[string]interface{})
	}

	return context
}

// ResolveTemplateVariables resolves all {{...}} variables in a string
// This is useful for simple template variable resolution without full expression evaluation
func ResolveTemplateVariables(template string, context map[string]interface{}) (string, error) {
	result := template

	matches := templateRegex.FindAllStringSubmatch(template, -1)
	for _, match := range matches {
		if len(match) > 1 {
			path := strings.TrimSpace(match[1])
			value, err := GetValueByPath(context, path)
			if err != nil {
				// Return original if path not found
				continue
			}

			// Convert value to string
			strValue := toString(value)
			result = strings.Replace(result, match[0], strValue, 1)
		}
	}

	return result, nil
}

// GetValueByPath retrieves a value from nested maps using dot notation
// Supports array indexing: "steps.step1.output[0].name"
func GetValueByPath(data map[string]interface{}, path string) (interface{}, error) {
	if path == "" {
		return data, nil
	}

	// Split path by dots, but preserve array indexes
	parts := splitPathSafe(path)
	current := interface{}(data)

	for i, part := range parts {
		// Check for array indexing
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			// Handle array access like "users[0]"
			bracketIdx := strings.Index(part, "[")
			key := part[:bracketIdx]
			indexPart := part[bracketIdx:]

			// Get the field first
			switch v := current.(type) {
			case map[string]interface{}:
				current = v[key]
			default:
				return nil, fmt.Errorf("cannot access key '%s' on non-object at position %d", key, i)
			}

			// Extract all array indices (supports multiple levels like [0][1])
			indices := extractArrayIndices(indexPart)
			for _, idx := range indices {
				switch arr := current.(type) {
				case []interface{}:
					if idx < 0 || idx >= len(arr) {
						return nil, fmt.Errorf("array index %d out of bounds at '%s'", idx, key)
					}
					current = arr[idx]
				default:
					return nil, fmt.Errorf("cannot index non-array type at '%s'", key)
				}
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

// splitPathSafe splits a path by dots while preserving array indices
func splitPathSafe(path string) []string {
	var parts []string
	var current strings.Builder
	inBracket := false

	for i := 0; i < len(path); i++ {
		char := path[i]

		switch char {
		case '[':
			inBracket = true
			current.WriteByte(char)
		case ']':
			inBracket = false
			current.WriteByte(char)
		case '.':
			if !inBracket {
				if current.Len() > 0 {
					parts = append(parts, current.String())
					current.Reset()
				}
			} else {
				current.WriteByte(char)
			}
		default:
			current.WriteByte(char)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// extractArrayIndices extracts numeric indices from [0][1] format
func extractArrayIndices(indexPart string) []int {
	var indices []int
	indexRegex := regexp.MustCompile(`\[(\d+)\]`)
	matches := indexRegex.FindAllStringSubmatch(indexPart, -1)

	for _, match := range matches {
		if len(match) > 1 {
			var idx int
			fmt.Sscanf(match[1], "%d", &idx)
			indices = append(indices, idx)
		}
	}

	return indices
}

// toString converts any value to string representation
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
