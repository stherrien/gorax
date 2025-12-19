package actions

import (
	"encoding/json"
	"testing"
)

func TestInterpolateString(t *testing.T) {
	context := map[string]interface{}{
		"trigger": map[string]interface{}{
			"event": "user.created",
			"user": map[string]interface{}{
				"id":   123,
				"name": "John Doe",
			},
		},
		"steps": map[string]interface{}{
			"http-1": map[string]interface{}{
				"body": map[string]interface{}{
					"origin": "192.168.1.1",
					"users": []interface{}{
						map[string]interface{}{"id": 1, "name": "Alice"},
						map[string]interface{}{"id": 2, "name": "Bob"},
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name:     "simple property",
			template: "Event: {{trigger.event}}",
			want:     "Event: user.created",
		},
		{
			name:     "nested property",
			template: "User: {{trigger.user.name}}",
			want:     "User: John Doe",
		},
		{
			name:     "step output",
			template: "Origin: {{steps.http-1.body.origin}}",
			want:     "Origin: 192.168.1.1",
		},
		{
			name:     "array access",
			template: "First user: {{steps.http-1.body.users[0].name}}",
			want:     "First user: Alice",
		},
		{
			name:     "multiple interpolations",
			template: "{{trigger.user.name}} from {{steps.http-1.body.origin}}",
			want:     "John Doe from 192.168.1.1",
		},
		{
			name:     "missing path",
			template: "Value: {{missing.path}}",
			want:     "Value: {{missing.path}}",
		},
		{
			name:     "no interpolation",
			template: "Just a plain string",
			want:     "Just a plain string",
		},
		{
			name:     "integer value",
			template: "ID: {{trigger.user.id}}",
			want:     "ID: 123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InterpolateString(tt.template, context)
			if got != tt.want {
				t.Errorf("InterpolateString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetValueByPath(t *testing.T) {
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"id":   123,
			"name": "John",
			"tags": []interface{}{"admin", "developer"},
		},
		"items": []interface{}{
			map[string]interface{}{"id": 1, "name": "Item 1"},
			map[string]interface{}{"id": 2, "name": "Item 2"},
		},
		"count": 42,
	}

	tests := []struct {
		name    string
		path    string
		want    interface{}
		wantErr bool
	}{
		{
			name:    "simple property",
			path:    "count",
			want:    42,
			wantErr: false,
		},
		{
			name:    "nested property",
			path:    "user.name",
			want:    "John",
			wantErr: false,
		},
		{
			name:    "array in object",
			path:    "user.tags[0]",
			want:    "admin",
			wantErr: false,
		},
		{
			name:    "nested array",
			path:    "items[1].name",
			want:    "Item 2",
			wantErr: false,
		},
		{
			name:    "empty path returns root",
			path:    "",
			want:    data,
			wantErr: false,
		},
		{
			name:    "missing key",
			path:    "missing",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid array index",
			path:    "items[999]",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "non-object traversal",
			path:    "count.invalid",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetValueByPath(data, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValueByPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !deepEqual(got, tt.want) {
				t.Errorf("GetValueByPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterpolateJSON(t *testing.T) {
	context := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
			"id":   100,
		},
	}

	tests := []struct {
		name string
		json string
		want map[string]interface{}
	}{
		{
			name: "simple interpolation",
			json: `{"name": "{{user.name}}", "static": "value"}`,
			want: map[string]interface{}{
				"name":   "Alice",
				"static": "value",
			},
		},
		{
			name: "nested object",
			json: `{"user": {"name": "{{user.name}}", "id": "{{user.id}}"}}`,
			want: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "Alice",
					"id":   "100",
				},
			},
		},
		{
			name: "array interpolation",
			json: `{"items": ["{{user.name}}", "static", "{{user.id}}"]}`,
			want: map[string]interface{}{
				"items": []interface{}{"Alice", "static", "100"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InterpolateJSON(json.RawMessage(tt.json), context)
			if !deepEqual(got, tt.want) {
				t.Errorf("InterpolateJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "simple path",
			path: "user.name",
			want: []string{"user", "name"},
		},
		{
			name: "nested path",
			path: "user.profile.address.city",
			want: []string{"user", "profile", "address", "city"},
		},
		{
			name: "single element",
			path: "user",
			want: []string{"user"},
		},
		{
			name: "empty path",
			path: "",
			want: []string{},
		},
		{
			name: "escaped dot",
			path: `user.file\.name`,
			want: []string{"user", "file.name"},
		},
		{
			name: "with array notation",
			path: "users[0]",
			want: []string{"users[0]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitPath(tt.path)
			if !equalStringSlices(got, tt.want) {
				t.Errorf("splitPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{
			name:  "string",
			value: "hello",
			want:  "hello",
		},
		{
			name:  "integer",
			value: 42,
			want:  "42",
		},
		{
			name:  "float",
			value: 3.14,
			want:  "3.14",
		},
		{
			name:  "boolean true",
			value: true,
			want:  "true",
		},
		{
			name:  "boolean false",
			value: false,
			want:  "false",
		},
		{
			name:  "nil",
			value: nil,
			want:  "",
		},
		{
			name:  "map",
			value: map[string]interface{}{"key": "value"},
			want:  `{"key":"value"}`,
		},
		{
			name:  "array",
			value: []interface{}{1, 2, 3},
			want:  "[1,2,3]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toString(tt.value)
			if got != tt.want {
				t.Errorf("toString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeepCopy(t *testing.T) {
	original := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
			"tags": []interface{}{"admin", "user"},
		},
		"count": 42,
	}

	copied, err := DeepCopy(original)
	if err != nil {
		t.Fatalf("DeepCopy() error = %v", err)
	}

	// Verify deep copy
	copiedMap := copied.(map[string]interface{})
	if !deepEqual(copiedMap, original) {
		t.Errorf("DeepCopy() = %v, want %v", copiedMap, original)
	}

	// Modify copy and verify original is unchanged
	copiedMap["count"] = 100
	if original["count"] != 42 {
		t.Errorf("Original was modified by changing copy")
	}
}

// Helper functions

func deepEqual(a, b interface{}) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
