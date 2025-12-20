package validation

import (
	"math"
	"strconv"
)

const (
	// MaxSafeInt is the maximum safe integer value to prevent overflow
	// when converting to int on 32-bit systems
	MaxSafeInt = math.MaxInt32

	// DefaultPaginationLimit is the default pagination limit
	DefaultPaginationLimit = 20

	// MaxPaginationLimit is the maximum allowed pagination limit
	MaxPaginationLimit = 1000

	// MaxPaginationOffset is the maximum allowed pagination offset
	MaxPaginationOffset = math.MaxInt32
)

// ParseSafeInt parses a string to int with bounds checking to prevent overflow.
// Returns the parsed value and a boolean indicating if parsing was successful and within bounds.
// If parsing fails or value is out of bounds, returns defaultValue and false.
func ParseSafeInt(s string, defaultValue, maxValue int) (int, bool) {
	if s == "" {
		return defaultValue, true
	}

	// Parse as int64 first to check for overflow
	val64, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return defaultValue, false
	}

	// Check if negative
	if val64 < 0 {
		return defaultValue, false
	}

	// Check if exceeds max safe int for 32-bit systems
	if val64 > int64(MaxSafeInt) {
		return defaultValue, false
	}

	// Check against custom max value
	val := int(val64)
	if val > maxValue {
		return defaultValue, false
	}

	return val, true
}

// ParsePaginationLimit parses and validates a pagination limit parameter.
// Returns defaultLimit if the input is invalid, zero, negative, or exceeds maxLimit.
// The returned boolean indicates if the parse was successful and within valid bounds.
func ParsePaginationLimit(limitStr string, defaultLimit, maxLimit int) (int, bool) {
	if limitStr == "" {
		return defaultLimit, true
	}

	limit, valid := ParseSafeInt(limitStr, defaultLimit, maxLimit)
	if !valid {
		return defaultLimit, false
	}

	// Zero limit should return default
	if limit == 0 {
		return defaultLimit, true
	}

	return limit, true
}

// ParsePaginationOffset parses and validates a pagination offset parameter.
// Returns 0 if the input is invalid or negative.
// The returned boolean indicates if the parse was successful.
func ParsePaginationOffset(offsetStr string) (int, bool) {
	if offsetStr == "" {
		return 0, true
	}

	offset, valid := ParseSafeInt(offsetStr, 0, MaxPaginationOffset)
	if !valid {
		return 0, false
	}

	return offset, true
}

// ParseArrayIndex parses and validates an array index string.
// Returns 0 and false if the index is invalid, negative, or out of array bounds.
// The returned boolean indicates if the index is valid and within array bounds.
func ParseArrayIndex(indexStr string, arrayLen int) (int, bool) {
	if indexStr == "" {
		return 0, false
	}

	// Parse as int64 first to check for overflow
	index64, err := strconv.ParseInt(indexStr, 10, 64)
	if err != nil {
		return 0, false
	}

	// Check if negative
	if index64 < 0 {
		return 0, false
	}

	// Check if exceeds max safe int
	if index64 > int64(MaxSafeInt) {
		return 0, false
	}

	index := int(index64)

	// Check if within array bounds
	if index >= arrayLen {
		return 0, false
	}

	return index, true
}

// ValidateIntRange checks if a value is within the specified range (inclusive).
func ValidateIntRange(value, min, max int) bool {
	return value >= min && value <= max
}
