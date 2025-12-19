package formula

import (
	"testing"
	"time"
)

// TestStringFunctions tests all string manipulation functions
func TestStringFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function string
		input    interface{}
		want     interface{}
		wantErr  bool
	}{
		// Upper function tests
		{
			name:     "upper converts to uppercase",
			function: "upper",
			input:    "hello world",
			want:     "HELLO WORLD",
			wantErr:  false,
		},
		{
			name:     "upper handles empty string",
			function: "upper",
			input:    "",
			want:     "",
			wantErr:  false,
		},
		{
			name:     "upper handles mixed case",
			function: "upper",
			input:    "Hello World",
			want:     "HELLO WORLD",
			wantErr:  false,
		},

		// Lower function tests
		{
			name:     "lower converts to lowercase",
			function: "lower",
			input:    "HELLO WORLD",
			want:     "hello world",
			wantErr:  false,
		},
		{
			name:     "lower handles empty string",
			function: "lower",
			input:    "",
			want:     "",
			wantErr:  false,
		},
		{
			name:     "lower handles mixed case",
			function: "lower",
			input:    "Hello World",
			want:     "hello world",
			wantErr:  false,
		},

		// Trim function tests
		{
			name:     "trim removes leading and trailing whitespace",
			function: "trim",
			input:    "  hello world  ",
			want:     "hello world",
			wantErr:  false,
		},
		{
			name:     "trim handles no whitespace",
			function: "trim",
			input:    "hello",
			want:     "hello",
			wantErr:  false,
		},
		{
			name:     "trim handles empty string",
			function: "trim",
			input:    "",
			want:     "",
			wantErr:  false,
		},
		{
			name:     "trim handles only whitespace",
			function: "trim",
			input:    "   ",
			want:     "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			var err error

			switch tt.function {
			case "upper":
				got, err = stringUpper(tt.input)
			case "lower":
				got, err = stringLower(tt.input)
			case "trim":
				got, err = stringTrim(tt.input)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("function() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("function() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestStringConcat tests the concat function
func TestStringConcat(t *testing.T) {
	tests := []struct {
		name    string
		args    []interface{}
		want    string
		wantErr bool
	}{
		{
			name: "concat multiple strings",
			args: []interface{}{"hello", " ", "world"},
			want: "hello world",
		},
		{
			name: "concat with empty strings",
			args: []interface{}{"hello", "", "world"},
			want: "helloworld",
		},
		{
			name: "concat single string",
			args: []interface{}{"hello"},
			want: "hello",
		},
		{
			name: "concat empty args",
			args: []interface{}{},
			want: "",
		},
		{
			name: "concat with non-string types",
			args: []interface{}{"hello", 123, true},
			want: "hello123true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := stringConcat(tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("stringConcat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("stringConcat() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestStringSubstr tests the substr function
func TestStringSubstr(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		start   int
		length  int
		want    string
		wantErr bool
	}{
		{
			name:   "substr from beginning",
			str:    "hello world",
			start:  0,
			length: 5,
			want:   "hello",
		},
		{
			name:   "substr from middle",
			str:    "hello world",
			start:  6,
			length: 5,
			want:   "world",
		},
		{
			name:   "substr to end",
			str:    "hello world",
			start:  6,
			length: 100,
			want:   "world",
		},
		{
			name:   "substr with zero length",
			str:    "hello",
			start:  0,
			length: 0,
			want:   "",
		},
		{
			name:    "substr with negative start",
			str:     "hello",
			start:   -1,
			length:  5,
			want:    "",
			wantErr: true,
		},
		{
			name:    "substr with negative length",
			str:     "hello",
			start:   0,
			length:  -1,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := stringSubstr(tt.str, tt.start, tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("stringSubstr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("stringSubstr() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDateFunctions tests date manipulation functions
func TestDateFunctions(t *testing.T) {
	// Fixed time for testing
	fixedTime := time.Date(2025, 12, 17, 10, 30, 0, 0, time.UTC)

	t.Run("dateNow returns current time", func(t *testing.T) {
		before := time.Now()
		result, err := dateNow()
		after := time.Now()

		if err != nil {
			t.Errorf("dateNow() error = %v", err)
			return
		}

		resultTime, ok := result.(time.Time)
		if !ok {
			t.Errorf("dateNow() did not return time.Time, got %T", result)
			return
		}

		if resultTime.Before(before) || resultTime.After(after) {
			t.Errorf("dateNow() = %v, expected time between %v and %v", resultTime, before, after)
		}
	})

	t.Run("dateFormat formats time correctly", func(t *testing.T) {
		tests := []struct {
			name   string
			time   time.Time
			format string
			want   string
		}{
			{
				name:   "RFC3339 format",
				time:   fixedTime,
				format: time.RFC3339,
				want:   "2025-12-17T10:30:00Z",
			},
			{
				name:   "custom format",
				time:   fixedTime,
				format: "2006-01-02",
				want:   "2025-12-17",
			},
			{
				name:   "time only",
				time:   fixedTime,
				format: "15:04:05",
				want:   "10:30:00",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := dateFormat(tt.time, tt.format)
				if err != nil {
					t.Errorf("dateFormat() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("dateFormat() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("dateParse parses time string", func(t *testing.T) {
		tests := []struct {
			name    string
			value   string
			format  string
			want    time.Time
			wantErr bool
		}{
			{
				name:   "parse RFC3339",
				value:  "2025-12-17T10:30:00Z",
				format: time.RFC3339,
				want:   fixedTime,
			},
			{
				name:   "parse custom format",
				value:  "2025-12-17",
				format: "2006-01-02",
				want:   time.Date(2025, 12, 17, 0, 0, 0, 0, time.UTC),
			},
			{
				name:    "parse invalid format",
				value:   "invalid",
				format:  time.RFC3339,
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := dateParse(tt.value, tt.format)
				if (err != nil) != tt.wantErr {
					t.Errorf("dateParse() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr {
					gotTime, ok := got.(time.Time)
					if !ok {
						t.Errorf("dateParse() did not return time.Time")
						return
					}
					if !gotTime.Equal(tt.want) {
						t.Errorf("dateParse() = %v, want %v", gotTime, tt.want)
					}
				}
			})
		}
	})

	t.Run("dateAddDays adds days to time", func(t *testing.T) {
		tests := []struct {
			name string
			time time.Time
			days int
			want time.Time
		}{
			{
				name: "add positive days",
				time: fixedTime,
				days: 5,
				want: fixedTime.AddDate(0, 0, 5),
			},
			{
				name: "add negative days",
				time: fixedTime,
				days: -3,
				want: fixedTime.AddDate(0, 0, -3),
			},
			{
				name: "add zero days",
				time: fixedTime,
				days: 0,
				want: fixedTime,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := dateAddDays(tt.time, tt.days)
				if err != nil {
					t.Errorf("dateAddDays() error = %v", err)
					return
				}
				gotTime, ok := got.(time.Time)
				if !ok {
					t.Errorf("dateAddDays() did not return time.Time")
					return
				}
				if !gotTime.Equal(tt.want) {
					t.Errorf("dateAddDays() = %v, want %v", gotTime, tt.want)
				}
			})
		}
	})
}

// TestMathFunctions tests mathematical functions
func TestMathFunctions(t *testing.T) {
	t.Run("mathRound rounds to nearest integer", func(t *testing.T) {
		tests := []struct {
			name  string
			value float64
			want  float64
		}{
			{name: "round up", value: 4.6, want: 5.0},
			{name: "round down", value: 4.4, want: 4.0},
			{name: "round exact half up", value: 4.5, want: 5.0},
			{name: "round negative", value: -4.6, want: -5.0},
			{name: "round zero", value: 0.0, want: 0.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := mathRound(tt.value)
				if err != nil {
					t.Errorf("mathRound() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("mathRound() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("mathCeil rounds up", func(t *testing.T) {
		tests := []struct {
			value float64
			want  float64
		}{
			{value: 4.1, want: 5.0},
			{value: 4.9, want: 5.0},
			{value: -4.1, want: -4.0},
			{value: 5.0, want: 5.0},
		}

		for _, tt := range tests {
			got, err := mathCeil(tt.value)
			if err != nil {
				t.Errorf("mathCeil() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("mathCeil(%v) = %v, want %v", tt.value, got, tt.want)
			}
		}
	})

	t.Run("mathFloor rounds down", func(t *testing.T) {
		tests := []struct {
			value float64
			want  float64
		}{
			{value: 4.1, want: 4.0},
			{value: 4.9, want: 4.0},
			{value: -4.1, want: -5.0},
			{value: 5.0, want: 5.0},
		}

		for _, tt := range tests {
			got, err := mathFloor(tt.value)
			if err != nil {
				t.Errorf("mathFloor() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("mathFloor(%v) = %v, want %v", tt.value, got, tt.want)
			}
		}
	})

	t.Run("mathAbs returns absolute value", func(t *testing.T) {
		tests := []struct {
			value float64
			want  float64
		}{
			{value: 5.0, want: 5.0},
			{value: -5.0, want: 5.0},
			{value: 0.0, want: 0.0},
		}

		for _, tt := range tests {
			got, err := mathAbs(tt.value)
			if err != nil {
				t.Errorf("mathAbs() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("mathAbs(%v) = %v, want %v", tt.value, got, tt.want)
			}
		}
	})

	t.Run("mathMin returns minimum value", func(t *testing.T) {
		tests := []struct {
			name string
			args []interface{}
			want float64
		}{
			{name: "two values", args: []interface{}{5.0, 3.0}, want: 3.0},
			{name: "multiple values", args: []interface{}{5.0, 3.0, 7.0, 1.0}, want: 1.0},
			{name: "negative values", args: []interface{}{-5.0, -3.0, -7.0}, want: -7.0},
			{name: "single value", args: []interface{}{5.0}, want: 5.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := mathMin(tt.args...)
				if err != nil {
					t.Errorf("mathMin() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("mathMin() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("mathMax returns maximum value", func(t *testing.T) {
		tests := []struct {
			name string
			args []interface{}
			want float64
		}{
			{name: "two values", args: []interface{}{5.0, 3.0}, want: 5.0},
			{name: "multiple values", args: []interface{}{5.0, 3.0, 7.0, 1.0}, want: 7.0},
			{name: "negative values", args: []interface{}{-5.0, -3.0, -7.0}, want: -3.0},
			{name: "single value", args: []interface{}{5.0}, want: 5.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := mathMax(tt.args...)
				if err != nil {
					t.Errorf("mathMax() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("mathMax() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("mathMin requires at least one argument", func(t *testing.T) {
		_, err := mathMin()
		if err == nil {
			t.Error("mathMin() with no args should return error")
		}
	})

	t.Run("mathMax requires at least one argument", func(t *testing.T) {
		_, err := mathMax()
		if err == nil {
			t.Error("mathMax() with no args should return error")
		}
	})
}

// TestArrayFunctions tests array manipulation functions
func TestArrayFunctions(t *testing.T) {
	t.Run("arrayLength returns array length", func(t *testing.T) {
		tests := []struct {
			name  string
			input interface{}
			want  int
		}{
			{
				name:  "array with elements",
				input: []interface{}{1, 2, 3, 4},
				want:  4,
			},
			{
				name:  "empty array",
				input: []interface{}{},
				want:  0,
			},
			{
				name:  "string slice",
				input: []interface{}{"a", "b", "c"},
				want:  3,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := arrayLength(tt.input)
				if err != nil {
					t.Errorf("arrayLength() error = %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("arrayLength() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("arrayLength with non-array returns error", func(t *testing.T) {
		_, err := arrayLength("not an array")
		if err == nil {
			t.Error("arrayLength() with non-array should return error")
		}
	})
}

// TestTypeConversions tests type conversion helpers
func TestTypeConversions(t *testing.T) {
	t.Run("toString converts values correctly", func(t *testing.T) {
		tests := []struct {
			name  string
			input interface{}
			want  string
		}{
			{name: "string", input: "hello", want: "hello"},
			{name: "int", input: 123, want: "123"},
			{name: "float", input: 123.45, want: "123.45"},
			{name: "bool true", input: true, want: "true"},
			{name: "bool false", input: false, want: "false"},
			{name: "nil", input: nil, want: ""},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := toString(tt.input)
				if got != tt.want {
					t.Errorf("toString() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("toFloat converts numeric values", func(t *testing.T) {
		tests := []struct {
			name    string
			input   interface{}
			want    float64
			wantErr bool
		}{
			{name: "float64", input: 123.45, want: 123.45},
			{name: "int", input: 123, want: 123.0},
			{name: "int64", input: int64(123), want: 123.0},
			{name: "string number", input: "123.45", want: 123.45},
			{name: "string int", input: "123", want: 123.0},
			{name: "invalid string", input: "not a number", wantErr: true},
			{name: "bool", input: true, wantErr: true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := toFloat(tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("toFloat() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && got != tt.want {
					t.Errorf("toFloat() = %v, want %v", got, tt.want)
				}
			})
		}
	})
}
