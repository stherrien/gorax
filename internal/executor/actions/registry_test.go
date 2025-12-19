package actions

import (
	"context"
	"testing"
)

// MockAction is a mock implementation of Action for testing
type MockAction struct {
	ExecuteCalled bool
	ExecuteError  error
	ExecuteResult *ActionOutput
}

func (m *MockAction) Execute(ctx context.Context, input *ActionInput) (*ActionOutput, error) {
	m.ExecuteCalled = true
	if m.ExecuteError != nil {
		return nil, m.ExecuteError
	}
	if m.ExecuteResult != nil {
		return m.ExecuteResult, nil
	}
	return NewActionOutput(map[string]interface{}{"mock": true}), nil
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("NewRegistry() returned nil")
	}

	// Check that built-in actions are registered
	if !registry.IsRegistered("action:http") {
		t.Error("action:http not registered")
	}
	if !registry.IsRegistered("action:transform") {
		t.Error("action:transform not registered")
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	// Register a custom action
	registry.Register("action:custom", func() Action {
		return &MockAction{}
	})

	if !registry.IsRegistered("action:custom") {
		t.Error("custom action not registered")
	}
}

func TestRegistry_Create(t *testing.T) {
	registry := NewRegistry()

	tests := []struct {
		name       string
		actionType string
		wantErr    bool
	}{
		{
			name:       "create http action",
			actionType: "action:http",
			wantErr:    false,
		},
		{
			name:       "create transform action",
			actionType: "action:transform",
			wantErr:    false,
		},
		{
			name:       "unknown action type",
			actionType: "action:unknown",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, err := registry.Create(tt.actionType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Registry.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && action == nil {
				t.Error("Registry.Create() returned nil action")
			}
		})
	}
}

func TestRegistry_IsRegistered(t *testing.T) {
	registry := NewRegistry()

	tests := []struct {
		name       string
		actionType string
		want       bool
	}{
		{
			name:       "registered action",
			actionType: "action:http",
			want:       true,
		},
		{
			name:       "unregistered action",
			actionType: "action:unknown",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.IsRegistered(tt.actionType)
			if got != tt.want {
				t.Errorf("Registry.IsRegistered() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_RegisteredTypes(t *testing.T) {
	registry := NewRegistry()

	types := registry.RegisteredTypes()

	// Should have at least the built-in actions
	if len(types) < 2 {
		t.Errorf("Registry.RegisteredTypes() returned %d types, want at least 2", len(types))
	}

	// Check that built-in types are present
	hasHTTP := false
	hasTransform := false
	for _, typ := range types {
		if typ == "action:http" {
			hasHTTP = true
		}
		if typ == "action:transform" {
			hasTransform = true
		}
	}

	if !hasHTTP {
		t.Error("action:http not in registered types")
	}
	if !hasTransform {
		t.Error("action:transform not in registered types")
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()

	// Test concurrent registration and creation
	done := make(chan bool)

	// Goroutine 1: Register actions
	go func() {
		for i := 0; i < 100; i++ {
			registry.Register("action:test", func() Action {
				return &MockAction{}
			})
		}
		done <- true
	}()

	// Goroutine 2: Create actions
	go func() {
		for i := 0; i < 100; i++ {
			_, _ = registry.Create("action:http")
		}
		done <- true
	}()

	// Goroutine 3: Check registration
	go func() {
		for i := 0; i < 100; i++ {
			_ = registry.IsRegistered("action:http")
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done
}

func TestActionInput_NewActionInput(t *testing.T) {
	config := map[string]interface{}{"key": "value"}
	context := map[string]interface{}{"trigger": "data"}

	input := NewActionInput(config, context)

	if input == nil {
		t.Fatal("NewActionInput() returned nil")
	}

	if input.Config == nil {
		t.Error("Config is nil")
	}

	if input.Context == nil {
		t.Error("Context is nil")
	}
}

func TestActionInput_NilContext(t *testing.T) {
	config := map[string]interface{}{"key": "value"}

	input := NewActionInput(config, nil)

	if input.Context == nil {
		t.Error("Context should be initialized to empty map, not nil")
	}
}

func TestActionOutput_NewActionOutput(t *testing.T) {
	data := map[string]interface{}{"result": "success"}

	output := NewActionOutput(data)

	if output == nil {
		t.Fatal("NewActionOutput() returned nil")
	}

	if output.Data == nil {
		t.Error("Data is nil")
	}

	if output.Metadata == nil {
		t.Error("Metadata should be initialized")
	}
}

func TestActionOutput_WithMetadata(t *testing.T) {
	output := NewActionOutput(map[string]interface{}{"result": "success"})

	output.WithMetadata("key1", "value1")
	output.WithMetadata("key2", 123)

	if len(output.Metadata) != 2 {
		t.Errorf("Expected 2 metadata entries, got %d", len(output.Metadata))
	}

	if output.Metadata["key1"] != "value1" {
		t.Error("Metadata key1 not set correctly")
	}

	if output.Metadata["key2"] != 123 {
		t.Error("Metadata key2 not set correctly")
	}
}

func TestMockAction_Execute(t *testing.T) {
	mock := &MockAction{
		ExecuteResult: NewActionOutput(map[string]interface{}{"mock": "data"}),
	}

	input := NewActionInput(nil, nil)
	output, err := mock.Execute(context.Background(), input)

	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if !mock.ExecuteCalled {
		t.Error("Execute was not called")
	}

	if output == nil {
		t.Fatal("Execute() returned nil output")
	}
}

func TestDefaultRegistry(t *testing.T) {
	if DefaultRegistry == nil {
		t.Fatal("DefaultRegistry is nil")
	}

	// Should have built-in actions registered
	if !DefaultRegistry.IsRegistered("action:http") {
		t.Error("action:http not registered in DefaultRegistry")
	}
}
