package security

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInputValidator_ValidateEmail(t *testing.T) {
	v := NewInputValidator()

	tests := []struct {
		name    string
		email   string
		wantErr bool
		errCode string
	}{
		{
			name:    "valid email",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with subdomain",
			email:   "test@mail.example.com",
			wantErr: false,
		},
		{
			name:    "valid email with plus",
			email:   "test+label@example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
			errCode: "required",
		},
		{
			name:    "missing @ symbol",
			email:   "testexample.com",
			wantErr: true,
			errCode: "invalid_format",
		},
		{
			name:    "missing domain",
			email:   "test@",
			wantErr: true,
			errCode: "invalid_format",
		},
		{
			name:    "missing TLD",
			email:   "test@example",
			wantErr: true,
			errCode: "invalid_chars",
		},
		{
			name:    "too long email",
			email:   strings.Repeat("a", 250) + "@example.com",
			wantErr: true,
			errCode: "max_length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateEmail(tt.email)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					vErr, ok := err.(*ValidationError)
					require.True(t, ok)
					assert.Equal(t, tt.errCode, vErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInputValidator_ValidatePassword(t *testing.T) {
	v := NewInputValidator()

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errCode  string
	}{
		{
			name:     "strong password",
			password: "MyStr0ng!Pass#123",
			wantErr:  false,
		},
		{
			name:     "password with 3 character types",
			password: "MyPassword123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
			errCode:  "required",
		},
		{
			name:     "too short",
			password: "Short1!",
			wantErr:  true,
			errCode:  "min_length",
		},
		{
			name:     "only lowercase",
			password: "alllowercase",
			wantErr:  true,
			errCode:  "weak_password",
		},
		{
			name:     "only uppercase",
			password: "ALLUPPERCASE",
			wantErr:  true,
			errCode:  "weak_password",
		},
		{
			name:     "only digits",
			password: "123456789012",
			wantErr:  true,
			errCode:  "weak_password",
		},
		{
			name:     "too long",
			password: strings.Repeat("Aa1!", 50), // 200 chars
			wantErr:  true,
			errCode:  "max_length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidatePassword(tt.password)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					vErr, ok := err.(*ValidationError)
					require.True(t, ok)
					assert.Equal(t, tt.errCode, vErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInputValidator_ValidateName(t *testing.T) {
	v := NewInputValidator()

	tests := []struct {
		name      string
		inputName string
		fieldName string
		wantErr   bool
		errCode   string
	}{
		{
			name:      "valid name",
			inputName: "My Workflow",
			fieldName: "name",
			wantErr:   false,
		},
		{
			name:      "name with numbers",
			inputName: "Workflow 123",
			fieldName: "name",
			wantErr:   false,
		},
		{
			name:      "name with special chars",
			inputName: "My_Workflow-Test",
			fieldName: "name",
			wantErr:   false,
		},
		{
			name:      "empty name",
			inputName: "",
			fieldName: "name",
			wantErr:   true,
			errCode:   "required",
		},
		{
			name:      "name with null byte",
			inputName: "Test\x00Name",
			fieldName: "name",
			wantErr:   true,
			errCode:   "invalid_chars",
		},
		{
			name:      "too long name",
			inputName: strings.Repeat("a", 300),
			fieldName: "name",
			wantErr:   true,
			errCode:   "max_length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateName(tt.inputName, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					vErr, ok := err.(*ValidationError)
					require.True(t, ok)
					assert.Equal(t, tt.errCode, vErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInputValidator_ValidateUUID(t *testing.T) {
	v := NewInputValidator()

	tests := []struct {
		name      string
		id        string
		fieldName string
		wantErr   bool
		errCode   string
	}{
		{
			name:      "valid UUID lowercase",
			id:        "550e8400-e29b-41d4-a716-446655440000",
			fieldName: "id",
			wantErr:   false,
		},
		{
			name:      "valid UUID uppercase",
			id:        "550E8400-E29B-41D4-A716-446655440000",
			fieldName: "id",
			wantErr:   false,
		},
		{
			name:      "empty UUID",
			id:        "",
			fieldName: "id",
			wantErr:   true,
			errCode:   "required",
		},
		{
			name:      "invalid format - no dashes",
			id:        "550e8400e29b41d4a716446655440000",
			fieldName: "id",
			wantErr:   true,
			errCode:   "invalid_format",
		},
		{
			name:      "invalid format - wrong length",
			id:        "550e8400-e29b-41d4-a716",
			fieldName: "id",
			wantErr:   true,
			errCode:   "invalid_format",
		},
		{
			name:      "invalid characters",
			id:        "550e8400-e29b-41d4-a716-44665544000g",
			fieldName: "id",
			wantErr:   true,
			errCode:   "invalid_format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateUUID(tt.id, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					vErr, ok := err.(*ValidationError)
					require.True(t, ok)
					assert.Equal(t, tt.errCode, vErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInputValidator_ValidateJSON(t *testing.T) {
	v := NewInputValidator()

	tests := []struct {
		name    string
		json    []byte
		wantErr bool
		errCode string
	}{
		{
			name:    "valid JSON object",
			json:    []byte(`{"key": "value"}`),
			wantErr: false,
		},
		{
			name:    "valid JSON array",
			json:    []byte(`[1, 2, 3]`),
			wantErr: false,
		},
		{
			name:    "empty JSON object",
			json:    []byte(`{}`),
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			json:    []byte(`{"key": }`),
			wantErr: true,
			errCode: "invalid_json",
		},
		{
			name:    "not JSON",
			json:    []byte(`not json at all`),
			wantErr: true,
			errCode: "invalid_json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateJSON(tt.json)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					vErr, ok := err.(*ValidationError)
					require.True(t, ok)
					assert.Equal(t, tt.errCode, vErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal string",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "string with null byte",
			input:    "Hello\x00World",
			expected: "HelloWorld",
		},
		{
			name:     "string with whitespace",
			input:    "  Hello World  ",
			expected: "Hello World",
		},
		{
			name:     "string with null and whitespace",
			input:    "  Hello\x00World  ",
			expected: "HelloWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal string",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "string with HTML tags",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "string with quotes",
			input:    `"quoted" & 'single'`,
			expected: "&#34;quoted&#34; &amp; &#39;single&#39;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeHTML(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "normal relative path",
			input:    "folder/file.txt",
			expected: "folder/file.txt",
			wantErr:  false,
		},
		{
			name:     "path with current dir",
			input:    "./folder/file.txt",
			expected: "folder/file.txt",
			wantErr:  false,
		},
		{
			name:     "path traversal attempt",
			input:    "../../../etc/passwd",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "hidden path traversal",
			input:    "folder/../../../etc/passwd",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "absolute path",
			input:    "/etc/passwd",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "path with null byte",
			input:    "folder\x00/file.txt",
			expected: "folder/file.txt",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizePath(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestContainsSQLInjection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "normal string",
			input:    "Hello World",
			expected: false,
		},
		{
			name:     "SQL comment",
			input:    "test--comment",
			expected: true,
		},
		{
			name:     "SQL union",
			input:    "test UNION SELECT * FROM users",
			expected: true,
		},
		{
			name:     "SQL drop",
			input:    "test; DROP TABLE users;",
			expected: true,
		},
		{
			name:     "mixed case union",
			input:    "test UnIoN select",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsSQLInjection(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsShellMetaChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "normal string",
			input:    "Hello World",
			expected: false,
		},
		{
			name:     "pipe character",
			input:    "test | cat",
			expected: true,
		},
		{
			name:     "semicolon",
			input:    "test; rm -rf /",
			expected: true,
		},
		{
			name:     "backtick",
			input:    "test`whoami`",
			expected: true,
		},
		{
			name:     "dollar sign",
			input:    "test$USER",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsShellMetaChars(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsXSSPattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "normal string",
			input:    "Hello World",
			expected: false,
		},
		{
			name:     "script tag",
			input:    "<script>alert('xss')</script>",
			expected: true,
		},
		{
			name:     "javascript protocol",
			input:    "javascript:alert('xss')",
			expected: true,
		},
		{
			name:     "event handler",
			input:    `<img src="x" onerror="alert('xss')">`,
			expected: true,
		},
		{
			name:     "iframe tag",
			input:    "<iframe src='evil.com'>",
			expected: true,
		},
		{
			name:     "svg tag",
			input:    "<svg onload='alert(1)'>",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsXSSPattern(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidationErrors(t *testing.T) {
	t.Run("empty errors", func(t *testing.T) {
		errs := &ValidationErrors{}
		assert.False(t, errs.HasErrors())
		assert.Equal(t, "validation failed", errs.Error())
	})

	t.Run("single error", func(t *testing.T) {
		errs := &ValidationErrors{}
		errs.Add("email", "invalid format", "invalid_format")
		assert.True(t, errs.HasErrors())
		assert.Equal(t, "email: invalid format", errs.Error())
	})

	t.Run("multiple errors", func(t *testing.T) {
		errs := &ValidationErrors{}
		errs.Add("email", "invalid format", "invalid_format")
		errs.Add("password", "too short", "min_length")
		assert.True(t, errs.HasErrors())
		assert.Contains(t, errs.Error(), "email: invalid format")
		assert.Contains(t, errs.Error(), "password: too short")
	})
}

func TestWebhookSignatureValidator(t *testing.T) {
	v := &WebhookSignatureValidator{}

	tests := []struct {
		name      string
		signature string
		wantErr   bool
		errCode   string
	}{
		{
			name:      "valid signature",
			signature: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			wantErr:   false,
		},
		{
			name:      "empty signature",
			signature: "",
			wantErr:   true,
			errCode:   "required",
		},
		{
			name:      "too short",
			signature: "a1b2c3d4",
			wantErr:   true,
			errCode:   "invalid_format",
		},
		{
			name:      "invalid characters",
			signature: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6ghij",
			wantErr:   true,
			errCode:   "invalid_chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateSignatureFormat(tt.signature)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					vErr, ok := err.(*ValidationError)
					require.True(t, ok)
					assert.Equal(t, tt.errCode, vErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
