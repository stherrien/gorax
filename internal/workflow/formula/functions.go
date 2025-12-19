package formula

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// String Functions

// stringUpper converts a string to uppercase
func stringUpper(s interface{}) (string, error) {
	str := toString(s)
	return strings.ToUpper(str), nil
}

// stringLower converts a string to lowercase
func stringLower(s interface{}) (string, error) {
	str := toString(s)
	return strings.ToLower(str), nil
}

// stringTrim removes leading and trailing whitespace
func stringTrim(s interface{}) (string, error) {
	str := toString(s)
	return strings.TrimSpace(str), nil
}

// stringConcat concatenates multiple values into a string
func stringConcat(args ...interface{}) (string, error) {
	var result strings.Builder
	for _, arg := range args {
		str := toString(arg)
		result.WriteString(str)
	}
	return result.String(), nil
}

// stringSubstr extracts a substring
func stringSubstr(s string, start int, length int) (string, error) {
	if start < 0 {
		return "", fmt.Errorf("start index cannot be negative")
	}
	if length < 0 {
		return "", fmt.Errorf("length cannot be negative")
	}

	if start >= len(s) {
		return "", nil
	}

	end := start + length
	if end > len(s) {
		end = len(s)
	}

	return s[start:end], nil
}

// Date Functions

// dateNow returns the current time
func dateNow() (interface{}, error) {
	return time.Now(), nil
}

// dateFormat formats a time value using the given layout
func dateFormat(t time.Time, layout string) (string, error) {
	return t.Format(layout), nil
}

// dateParse parses a time string using the given layout
func dateParse(value string, layout string) (interface{}, error) {
	t, err := time.Parse(layout, value)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time: %w", err)
	}
	return t, nil
}

// dateAddDays adds or subtracts days from a time value
func dateAddDays(t time.Time, days int) (interface{}, error) {
	return t.AddDate(0, 0, days), nil
}

// Math Functions

// mathRound rounds to the nearest integer
func mathRound(x float64) (float64, error) {
	return math.Round(x), nil
}

// mathCeil rounds up to the nearest integer
func mathCeil(x float64) (float64, error) {
	return math.Ceil(x), nil
}

// mathFloor rounds down to the nearest integer
func mathFloor(x float64) (float64, error) {
	return math.Floor(x), nil
}

// mathAbs returns the absolute value
func mathAbs(x float64) (float64, error) {
	return math.Abs(x), nil
}

// mathMin returns the minimum value from the arguments
func mathMin(args ...interface{}) (float64, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("min requires at least one argument")
	}

	min, err := toFloat(args[0])
	if err != nil {
		return 0, fmt.Errorf("invalid argument type: %w", err)
	}

	for i := 1; i < len(args); i++ {
		val, err := toFloat(args[i])
		if err != nil {
			return 0, fmt.Errorf("invalid argument type at position %d: %w", i, err)
		}
		if val < min {
			min = val
		}
	}

	return min, nil
}

// mathMax returns the maximum value from the arguments
func mathMax(args ...interface{}) (float64, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("max requires at least one argument")
	}

	max, err := toFloat(args[0])
	if err != nil {
		return 0, fmt.Errorf("invalid argument type: %w", err)
	}

	for i := 1; i < len(args); i++ {
		val, err := toFloat(args[i])
		if err != nil {
			return 0, fmt.Errorf("invalid argument type at position %d: %w", i, err)
		}
		if val > max {
			max = val
		}
	}

	return max, nil
}

// Array Functions

// arrayLength returns the length of an array
func arrayLength(arr interface{}) (int, error) {
	switch v := arr.(type) {
	case []interface{}:
		return len(v), nil
	default:
		return 0, fmt.Errorf("argument must be an array, got %T", arr)
	}
}

// lenFunc returns the length of an array or string
func lenFunc(v interface{}) (int, error) {
	switch val := v.(type) {
	case []interface{}:
		return len(val), nil
	case string:
		return len(val), nil
	case []string:
		return len(val), nil
	case []int:
		return len(val), nil
	case []float64:
		return len(val), nil
	default:
		return 0, fmt.Errorf("len() requires an array or string, got %T", v)
	}
}

// Helper Functions

// toString converts various types to string
func toString(v interface{}) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// toFloat converts various types to float64
func toFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}
