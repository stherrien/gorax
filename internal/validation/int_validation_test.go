package validation

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSafeInt(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue int
		maxValue     int
		want         int
		wantValid    bool
	}{
		{
			name:         "valid positive integer",
			input:        "42",
			defaultValue: 10,
			maxValue:     100,
			want:         42,
			wantValid:    true,
		},
		{
			name:         "zero value",
			input:        "0",
			defaultValue: 10,
			maxValue:     100,
			want:         0,
			wantValid:    true,
		},
		{
			name:         "negative value rejected",
			input:        "-5",
			defaultValue: 10,
			maxValue:     100,
			want:         10,
			wantValid:    false,
		},
		{
			name:         "exceeds max value",
			input:        "1000",
			defaultValue: 10,
			maxValue:     100,
			want:         10,
			wantValid:    false,
		},
		{
			name:         "empty string returns default",
			input:        "",
			defaultValue: 20,
			maxValue:     100,
			want:         20,
			wantValid:    true,
		},
		{
			name:         "invalid format returns default",
			input:        "abc",
			defaultValue: 15,
			maxValue:     100,
			want:         15,
			wantValid:    false,
		},
		{
			name:         "max int32 value",
			input:        "2147483647", // math.MaxInt32
			defaultValue: 10,
			maxValue:     math.MaxInt32,
			want:         math.MaxInt32,
			wantValid:    true,
		},
		{
			name:         "overflow protection - exceeds int32",
			input:        "9223372036854775807", // math.MaxInt64
			defaultValue: 10,
			maxValue:     math.MaxInt32,
			want:         10,
			wantValid:    false,
		},
		{
			name:         "max value at boundary",
			input:        "100",
			defaultValue: 10,
			maxValue:     100,
			want:         100,
			wantValid:    true,
		},
		{
			name:         "just over max value",
			input:        "101",
			defaultValue: 10,
			maxValue:     100,
			want:         10,
			wantValid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, valid := ParseSafeInt(tt.input, tt.defaultValue, tt.maxValue)
			assert.Equal(t, tt.want, got, "ParseSafeInt() value mismatch")
			assert.Equal(t, tt.wantValid, valid, "ParseSafeInt() validity mismatch")
		})
	}
}

func TestParsePaginationLimit(t *testing.T) {
	tests := []struct {
		name         string
		limitStr     string
		defaultLimit int
		maxLimit     int
		wantLimit    int
		wantValid    bool
	}{
		{
			name:         "valid limit within bounds",
			limitStr:     "50",
			defaultLimit: 20,
			maxLimit:     100,
			wantLimit:    50,
			wantValid:    true,
		},
		{
			name:         "empty string returns default",
			limitStr:     "",
			defaultLimit: 20,
			maxLimit:     100,
			wantLimit:    20,
			wantValid:    true,
		},
		{
			name:         "zero returns default",
			limitStr:     "0",
			defaultLimit: 20,
			maxLimit:     100,
			wantLimit:    20,
			wantValid:    true,
		},
		{
			name:         "negative value returns default",
			limitStr:     "-10",
			defaultLimit: 20,
			maxLimit:     100,
			wantLimit:    20,
			wantValid:    false,
		},
		{
			name:         "exceeds max returns default",
			limitStr:     "1000",
			defaultLimit: 20,
			maxLimit:     100,
			wantLimit:    20,
			wantValid:    false,
		},
		{
			name:         "invalid format returns default",
			limitStr:     "abc",
			defaultLimit: 20,
			maxLimit:     100,
			wantLimit:    20,
			wantValid:    false,
		},
		{
			name:         "max limit boundary",
			limitStr:     "100",
			defaultLimit: 20,
			maxLimit:     100,
			wantLimit:    100,
			wantValid:    true,
		},
		{
			name:         "overflow protection",
			limitStr:     "9999999999999",
			defaultLimit: 20,
			maxLimit:     1000,
			wantLimit:    20,
			wantValid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLimit, gotValid := ParsePaginationLimit(tt.limitStr, tt.defaultLimit, tt.maxLimit)
			assert.Equal(t, tt.wantLimit, gotLimit, "ParsePaginationLimit() limit mismatch")
			assert.Equal(t, tt.wantValid, gotValid, "ParsePaginationLimit() validity mismatch")
		})
	}
}

func TestParsePaginationOffset(t *testing.T) {
	tests := []struct {
		name       string
		offsetStr  string
		wantOffset int
		wantValid  bool
	}{
		{
			name:       "valid offset",
			offsetStr:  "100",
			wantOffset: 100,
			wantValid:  true,
		},
		{
			name:       "zero offset",
			offsetStr:  "0",
			wantOffset: 0,
			wantValid:  true,
		},
		{
			name:       "empty string returns zero",
			offsetStr:  "",
			wantOffset: 0,
			wantValid:  true,
		},
		{
			name:       "negative value rejected",
			offsetStr:  "-10",
			wantOffset: 0,
			wantValid:  false,
		},
		{
			name:       "invalid format returns zero",
			offsetStr:  "xyz",
			wantOffset: 0,
			wantValid:  false,
		},
		{
			name:       "large valid offset",
			offsetStr:  "1000000",
			wantOffset: 1000000,
			wantValid:  true,
		},
		{
			name:       "overflow protection - exceeds safe int",
			offsetStr:  "9223372036854775807",
			wantOffset: 0,
			wantValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOffset, gotValid := ParsePaginationOffset(tt.offsetStr)
			assert.Equal(t, tt.wantOffset, gotOffset, "ParsePaginationOffset() offset mismatch")
			assert.Equal(t, tt.wantValid, gotValid, "ParsePaginationOffset() validity mismatch")
		})
	}
}

func TestParseArrayIndex(t *testing.T) {
	tests := []struct {
		name      string
		indexStr  string
		arrayLen  int
		wantIndex int
		wantValid bool
	}{
		{
			name:      "valid index within bounds",
			indexStr:  "5",
			arrayLen:  10,
			wantIndex: 5,
			wantValid: true,
		},
		{
			name:      "zero index",
			indexStr:  "0",
			arrayLen:  10,
			wantIndex: 0,
			wantValid: true,
		},
		{
			name:      "negative index rejected",
			indexStr:  "-1",
			arrayLen:  10,
			wantIndex: 0,
			wantValid: false,
		},
		{
			name:      "index exceeds array length",
			indexStr:  "10",
			arrayLen:  10,
			wantIndex: 0,
			wantValid: false,
		},
		{
			name:      "invalid format",
			indexStr:  "abc",
			arrayLen:  10,
			wantIndex: 0,
			wantValid: false,
		},
		{
			name:      "empty string",
			indexStr:  "",
			arrayLen:  10,
			wantIndex: 0,
			wantValid: false,
		},
		{
			name:      "overflow protection",
			indexStr:  "9999999999999",
			arrayLen:  100,
			wantIndex: 0,
			wantValid: false,
		},
		{
			name:      "max valid index",
			indexStr:  "9",
			arrayLen:  10,
			wantIndex: 9,
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIndex, gotValid := ParseArrayIndex(tt.indexStr, tt.arrayLen)
			assert.Equal(t, tt.wantIndex, gotIndex, "ParseArrayIndex() index mismatch")
			assert.Equal(t, tt.wantValid, gotValid, "ParseArrayIndex() validity mismatch")
		})
	}
}

func TestValidateIntRange(t *testing.T) {
	tests := []struct {
		name  string
		value int
		min   int
		max   int
		want  bool
	}{
		{
			name:  "value within range",
			value: 50,
			min:   0,
			max:   100,
			want:  true,
		},
		{
			name:  "value at min boundary",
			value: 0,
			min:   0,
			max:   100,
			want:  true,
		},
		{
			name:  "value at max boundary",
			value: 100,
			min:   0,
			max:   100,
			want:  true,
		},
		{
			name:  "value below min",
			value: -1,
			min:   0,
			max:   100,
			want:  false,
		},
		{
			name:  "value above max",
			value: 101,
			min:   0,
			max:   100,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateIntRange(tt.value, tt.min, tt.max)
			assert.Equal(t, tt.want, got, "ValidateIntRange() mismatch")
		})
	}
}
