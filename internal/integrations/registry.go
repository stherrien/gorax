package integrations

import (
	"fmt"
	"sync"
)

// Registry manages all available integration actions
type Registry struct {
	actions map[string]Action
	mu      sync.RWMutex
}

// NewRegistry creates a new action registry
func NewRegistry() *Registry {
	return &Registry{
		actions: make(map[string]Action),
	}
}

// Register adds an action to the registry
func (r *Registry) Register(action Action) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := action.Name()
	if name == "" {
		return fmt.Errorf("action name cannot be empty")
	}

	if _, exists := r.actions[name]; exists {
		return fmt.Errorf("action %s already registered", name)
	}

	r.actions[name] = action
	return nil
}

// Get retrieves an action by name
func (r *Registry) Get(name string) (Action, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	action, exists := r.actions[name]
	if !exists {
		return nil, fmt.Errorf("action %s not found", name)
	}

	return action, nil
}

// List returns all registered action names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.actions))
	for name := range r.actions {
		names = append(names, name)
	}

	return names
}

// Unregister removes an action from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.actions[name]; !exists {
		return fmt.Errorf("action %s not found", name)
	}

	delete(r.actions, name)
	return nil
}

// GlobalRegistry is the default registry instance
var GlobalRegistry = NewRegistry()
