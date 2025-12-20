package integrations

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAction is a test implementation of the Action interface
type mockAction struct {
	name        string
	description string
}

func (m *mockAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func (m *mockAction) Validate(config map[string]interface{}) error {
	return nil
}

func (m *mockAction) Name() string {
	return m.name
}

func (m *mockAction) Description() string {
	return m.description
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	action := &mockAction{name: "test:action", description: "Test action"}

	err := registry.Register(action)
	assert.NoError(t, err)

	// Verify action was registered
	retrieved, err := registry.Get("test:action")
	require.NoError(t, err)
	assert.Equal(t, action, retrieved)
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	registry := NewRegistry()
	action := &mockAction{name: "test:action", description: "Test action"}

	err := registry.Register(action)
	require.NoError(t, err)

	// Try to register again
	err = registry.Register(action)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRegistry_RegisterEmptyName(t *testing.T) {
	registry := NewRegistry()
	action := &mockAction{name: "", description: "Test action"}

	err := registry.Register(action)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	action := &mockAction{name: "test:action", description: "Test action"}

	err := registry.Register(action)
	require.NoError(t, err)

	retrieved, err := registry.Get("test:action")
	assert.NoError(t, err)
	assert.Equal(t, action, retrieved)
}

func TestRegistry_GetNotFound(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.Get("nonexistent:action")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	action1 := &mockAction{name: "test:action1", description: "Test action 1"}
	action2 := &mockAction{name: "test:action2", description: "Test action 2"}

	err := registry.Register(action1)
	require.NoError(t, err)

	err = registry.Register(action2)
	require.NoError(t, err)

	names := registry.List()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "test:action1")
	assert.Contains(t, names, "test:action2")
}

func TestRegistry_ListEmpty(t *testing.T) {
	registry := NewRegistry()

	names := registry.List()
	assert.Empty(t, names)
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()
	action := &mockAction{name: "test:action", description: "Test action"}

	err := registry.Register(action)
	require.NoError(t, err)

	err = registry.Unregister("test:action")
	assert.NoError(t, err)

	// Verify action was removed
	_, err = registry.Get("test:action")
	assert.Error(t, err)
}

func TestRegistry_UnregisterNotFound(t *testing.T) {
	registry := NewRegistry()

	err := registry.Unregister("nonexistent:action")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()

	// Register actions concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			action := &mockAction{
				name:        fmt.Sprintf("test:action%d", id),
				description: "Test action",
			}
			_ = registry.Register(action)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all actions were registered
	names := registry.List()
	assert.Len(t, names, 10)
}
