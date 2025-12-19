package actions

import (
	"fmt"
	"sync"
)

// Registry manages action type registration and creation
type Registry struct {
	factories map[string]ActionFactory
	mu        sync.RWMutex
}

// NewRegistry creates a new action registry
func NewRegistry() *Registry {
	r := &Registry{
		factories: make(map[string]ActionFactory),
	}

	// Register built-in actions
	r.Register("action:http", func() Action { return &HTTPAction{} })
	r.Register("action:transform", func() Action { return &TransformAction{} })
	r.Register("action:formula", func() Action { return &FormulaAction{} })
	r.Register("action:code", func() Action { return &ScriptAction{} })

	return r
}

// Register registers an action factory for a given action type
func (r *Registry) Register(actionType string, factory ActionFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[actionType] = factory
}

// Create creates a new action instance for the given action type
func (r *Registry) Create(actionType string) (Action, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, exists := r.factories[actionType]
	if !exists {
		return nil, fmt.Errorf("unknown action type: %s", actionType)
	}

	return factory(), nil
}

// IsRegistered checks if an action type is registered
func (r *Registry) IsRegistered(actionType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.factories[actionType]
	return exists
}

// RegisteredTypes returns a list of all registered action types
func (r *Registry) RegisteredTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.factories))
	for actionType := range r.factories {
		types = append(types, actionType)
	}
	return types
}

// DefaultRegistry is the global action registry
var DefaultRegistry = NewRegistry()
