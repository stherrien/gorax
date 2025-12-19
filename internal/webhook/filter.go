package webhook

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// FilterRepository defines the interface for filter-related data access
type FilterRepository interface {
	GetFiltersByWebhookID(ctx context.Context, webhookID string) ([]*WebhookFilter, error)
}

// FilterEvaluator evaluates webhook filters against payloads
type FilterEvaluator interface {
	// Evaluate checks if payload matches all filters for a webhook
	Evaluate(ctx context.Context, webhookID string, payload map[string]interface{}) (*FilterResult, error)

	// EvaluateSingle checks a single filter against payload
	EvaluateSingle(filter *WebhookFilter, payload map[string]interface{}) (bool, error)
}

type filterEvaluator struct {
	repo FilterRepository
}

// NewFilterEvaluator creates a new filter evaluator
func NewFilterEvaluator(repo FilterRepository) FilterEvaluator {
	return &filterEvaluator{repo: repo}
}

// Evaluate checks if payload matches all filters for a webhook
func (e *filterEvaluator) Evaluate(ctx context.Context, webhookID string, payload map[string]interface{}) (*FilterResult, error) {
	// Get all filters for this webhook
	filters, err := e.repo.GetFiltersByWebhookID(ctx, webhookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get filters: %w", err)
	}

	// No filters means pass by default
	if len(filters) == 0 {
		return &FilterResult{
			Passed:  true,
			Reason:  "no filters configured",
			Details: map[string]interface{}{},
		}, nil
	}

	// Group filters by logic group (for OR logic between groups)
	groupedFilters := make(map[int][]*WebhookFilter)
	for _, filter := range filters {
		if !filter.Enabled {
			continue
		}
		groupedFilters[filter.LogicGroup] = append(groupedFilters[filter.LogicGroup], filter)
	}

	// If all filters are disabled
	if len(groupedFilters) == 0 {
		return &FilterResult{
			Passed:  true,
			Reason:  "all filters disabled",
			Details: map[string]interface{}{},
		}, nil
	}

	// Evaluate each group (OR between groups, AND within group)
	details := make(map[string]interface{})
	var failedGroups []string

	for groupID, groupFilters := range groupedFilters {
		groupPassed := true
		var failedFilters []string

		// All filters in a group must pass (AND logic)
		for _, filter := range groupFilters {
			passed, err := e.EvaluateSingle(filter, payload)
			if err != nil {
				return nil, fmt.Errorf("filter evaluation error: %w", err)
			}

			if !passed {
				groupPassed = false
				failedFilters = append(failedFilters, fmt.Sprintf("%s %s %v", filter.FieldPath, filter.Operator, filter.Value))
			}
		}

		if groupPassed {
			// At least one group passed, so overall result is pass
			return &FilterResult{
				Passed:  true,
				Reason:  fmt.Sprintf("logic group %d passed", groupID),
				Details: details,
			}, nil
		}

		failedGroups = append(failedGroups, fmt.Sprintf("group %d", groupID))
		details[fmt.Sprintf("group_%d_failed_filters", groupID)] = failedFilters
	}

	// No groups passed
	return &FilterResult{
		Passed:  false,
		Reason:  fmt.Sprintf("no logic groups passed: %s", strings.Join(failedGroups, ", ")),
		Details: details,
	}, nil
}

// EvaluateSingle checks a single filter against payload
func (e *filterEvaluator) EvaluateSingle(filter *WebhookFilter, payload map[string]interface{}) (bool, error) {
	// Disabled filters always pass
	if !filter.Enabled {
		return true, nil
	}

	// Extract value from payload using JSON path
	value, exists := extractValue(filter.FieldPath, payload)

	// Handle exists/not exists operators
	if filter.Operator == OpExists {
		return exists, nil
	}
	if filter.Operator == OpNotExists {
		return !exists, nil
	}

	// For other operators, value must exist
	if !exists {
		return false, nil
	}

	// Evaluate based on operator
	switch filter.Operator {
	case OpEquals:
		return evaluateEquals(value, filter.Value)
	case OpNotEquals:
		result, err := evaluateEquals(value, filter.Value)
		return !result, err
	case OpContains:
		return evaluateContains(value, filter.Value)
	case OpNotContains:
		result, err := evaluateContains(value, filter.Value)
		return !result, err
	case OpStartsWith:
		return evaluateStartsWith(value, filter.Value)
	case OpEndsWith:
		return evaluateEndsWith(value, filter.Value)
	case OpRegex:
		return evaluateRegex(value, filter.Value)
	case OpGreaterThan:
		return evaluateGreaterThan(value, filter.Value)
	case OpLessThan:
		return evaluateLessThan(value, filter.Value)
	case OpIn:
		return evaluateIn(value, filter.Value)
	case OpNotIn:
		result, err := evaluateIn(value, filter.Value)
		return !result, err
	default:
		return false, fmt.Errorf("unknown operator: %s", filter.Operator)
	}
}

// extractValue extracts a value from a payload using a JSON path
func extractValue(path string, payload map[string]interface{}) (interface{}, bool) {
	// Remove leading $ if present
	path = strings.TrimPrefix(path, "$.")
	path = strings.TrimPrefix(path, "$")

	if path == "" {
		return payload, true
	}

	// Split path into parts
	parts := strings.Split(path, ".")
	current := interface{}(payload)

	for _, part := range parts {
		// Check if this is an array index
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			// Handle array[index] format
			openBracket := strings.Index(part, "[")
			closeBracket := strings.Index(part, "]")

			fieldName := part[:openBracket]
			indexStr := part[openBracket+1 : closeBracket]

			// Get the field first
			if fieldName != "" {
				currentMap, ok := current.(map[string]interface{})
				if !ok {
					return nil, false
				}
				current, ok = currentMap[fieldName]
				if !ok {
					return nil, false
				}
			}

			// Parse index
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, false
			}

			// Get array element
			arr, ok := current.([]interface{})
			if !ok {
				return nil, false
			}

			if index < 0 || index >= len(arr) {
				return nil, false
			}

			current = arr[index]
			continue
		}

		// Check if part is a numeric string (array index without brackets)
		if index, err := strconv.Atoi(part); err == nil {
			// This is an array index
			arr, ok := current.([]interface{})
			if !ok {
				return nil, false
			}

			if index < 0 || index >= len(arr) {
				return nil, false
			}

			current = arr[index]
			continue
		}

		// Regular field access
		currentMap, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}

		value, ok := currentMap[part]
		if !ok {
			return nil, false
		}

		current = value
	}

	return current, true
}

// evaluateEquals checks if two values are equal
func evaluateEquals(actual, expected interface{}) (bool, error) {
	return compareValues(actual, expected), nil
}

// compareValues compares two values with type coercion
func compareValues(a, b interface{}) bool {
	// Handle nil
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Try direct equality first
	if reflect.DeepEqual(a, b) {
		return true
	}

	// Handle numeric comparisons with type coercion
	aNum, aIsNum := toFloat64(a)
	bNum, bIsNum := toFloat64(b)
	if aIsNum && bIsNum {
		return aNum == bNum
	}

	// String comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return aStr == bStr
}

// toFloat64 attempts to convert a value to float64
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int8:
		return float64(val), true
	case int16:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint8:
		return float64(val), true
	case uint16:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	default:
		// Try string conversion
		if str, ok := v.(string); ok {
			if num, err := strconv.ParseFloat(str, 64); err == nil {
				return num, true
			}
		}
		return 0, false
	}
}

// evaluateContains checks if a string contains a substring
func evaluateContains(actual, expected interface{}) (bool, error) {
	actualStr, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("contains operator requires string value, got %T", actual)
	}

	expectedStr, ok := expected.(string)
	if !ok {
		return false, fmt.Errorf("contains operator requires string comparison value, got %T", expected)
	}

	return strings.Contains(actualStr, expectedStr), nil
}

// evaluateStartsWith checks if a string starts with a prefix
func evaluateStartsWith(actual, expected interface{}) (bool, error) {
	actualStr, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("starts_with operator requires string value, got %T", actual)
	}

	expectedStr, ok := expected.(string)
	if !ok {
		return false, fmt.Errorf("starts_with operator requires string comparison value, got %T", expected)
	}

	return strings.HasPrefix(actualStr, expectedStr), nil
}

// evaluateEndsWith checks if a string ends with a suffix
func evaluateEndsWith(actual, expected interface{}) (bool, error) {
	actualStr, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("ends_with operator requires string value, got %T", actual)
	}

	expectedStr, ok := expected.(string)
	if !ok {
		return false, fmt.Errorf("ends_with operator requires string comparison value, got %T", expected)
	}

	return strings.HasSuffix(actualStr, expectedStr), nil
}

// evaluateRegex checks if a string matches a regex pattern
func evaluateRegex(actual, expected interface{}) (bool, error) {
	actualStr, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("regex operator requires string value, got %T", actual)
	}

	patternStr, ok := expected.(string)
	if !ok {
		return false, fmt.Errorf("regex operator requires string pattern, got %T", expected)
	}

	regex, err := regexp.Compile(patternStr)
	if err != nil {
		return false, fmt.Errorf("invalid regex pattern: %w", err)
	}

	return regex.MatchString(actualStr), nil
}

// evaluateGreaterThan checks if a number is greater than another
func evaluateGreaterThan(actual, expected interface{}) (bool, error) {
	actualNum, ok := toFloat64(actual)
	if !ok {
		return false, fmt.Errorf("greater than operator requires numeric value, got %T", actual)
	}

	expectedNum, ok := toFloat64(expected)
	if !ok {
		return false, fmt.Errorf("greater than operator requires numeric comparison value, got %T", expected)
	}

	return actualNum > expectedNum, nil
}

// evaluateLessThan checks if a number is less than another
func evaluateLessThan(actual, expected interface{}) (bool, error) {
	actualNum, ok := toFloat64(actual)
	if !ok {
		return false, fmt.Errorf("less than operator requires numeric value, got %T", actual)
	}

	expectedNum, ok := toFloat64(expected)
	if !ok {
		return false, fmt.Errorf("less than operator requires numeric comparison value, got %T", expected)
	}

	return actualNum < expectedNum, nil
}

// evaluateIn checks if a value is in an array
func evaluateIn(actual, expected interface{}) (bool, error) {
	expectedArr, ok := expected.([]interface{})
	if !ok {
		return false, fmt.Errorf("in operator requires array comparison value, got %T", expected)
	}

	for _, item := range expectedArr {
		if compareValues(actual, item) {
			return true, nil
		}
	}

	return false, nil
}
