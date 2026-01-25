package integration

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// Registry manages the registration and discovery of integrations.
type Registry struct {
	integrations map[string]Integration
	factories    map[string]Factory
	metadata     map[string]*Metadata
	mu           sync.RWMutex
	logger       *slog.Logger
}

// Factory is a function that creates a new Integration instance.
type Factory func(config *Config) (Integration, error)

// NewRegistry creates a new integration registry.
func NewRegistry(logger *slog.Logger) *Registry {
	if logger == nil {
		logger = slog.Default()
	}
	return &Registry{
		integrations: make(map[string]Integration),
		factories:    make(map[string]Factory),
		metadata:     make(map[string]*Metadata),
		logger:       logger,
	}
}

// Register registers an integration with the registry.
func (r *Registry) Register(integration Integration) error {
	if integration == nil {
		return NewValidationError("integration", "integration cannot be nil", nil)
	}

	name := integration.Name()
	if name == "" {
		return NewValidationError("name", "integration name cannot be empty", nil)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.integrations[name]; exists {
		return fmt.Errorf("%w: %s", ErrAlreadyRegistered, name)
	}

	r.integrations[name] = integration
	r.metadata[name] = integration.GetMetadata()

	r.logger.Info("integration registered",
		"name", name,
		"type", integration.Type(),
	)

	return nil
}

// RegisterFactory registers a factory function for creating integrations.
func (r *Registry) RegisterFactory(name string, factory Factory) error {
	if name == "" {
		return NewValidationError("name", "factory name cannot be empty", nil)
	}
	if factory == nil {
		return NewValidationError("factory", "factory cannot be nil", nil)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("%w: factory %s", ErrAlreadyRegistered, name)
	}

	r.factories[name] = factory

	r.logger.Info("integration factory registered",
		"name", name,
	)

	return nil
}

// Get retrieves an integration by name.
func (r *Registry) Get(name string) (Integration, error) {
	if name == "" {
		return nil, NewValidationError("name", "integration name cannot be empty", nil)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	integration, exists := r.integrations[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
	}

	return integration, nil
}

// GetOrCreate retrieves an integration by name, or creates one using a factory.
func (r *Registry) GetOrCreate(name string, config *Config) (Integration, error) {
	// Try to get existing integration
	r.mu.RLock()
	if integration, exists := r.integrations[name]; exists {
		r.mu.RUnlock()
		return integration, nil
	}
	r.mu.RUnlock()

	// Try to create using factory
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if integration, exists := r.integrations[name]; exists {
		return integration, nil
	}

	factory, exists := r.factories[name]
	if !exists {
		return nil, fmt.Errorf("%w: no integration or factory found for %s", ErrNotFound, name)
	}

	integration, err := factory(config)
	if err != nil {
		return nil, fmt.Errorf("creating integration %s: %w", name, err)
	}

	r.integrations[name] = integration
	r.metadata[name] = integration.GetMetadata()

	r.logger.Info("integration created from factory",
		"name", name,
		"type", integration.Type(),
	)

	return integration, nil
}

// Unregister removes an integration from the registry.
func (r *Registry) Unregister(name string) error {
	if name == "" {
		return NewValidationError("name", "integration name cannot be empty", nil)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	integration, exists := r.integrations[name]
	if !exists {
		return fmt.Errorf("%w: %s", ErrNotFound, name)
	}

	// Call Shutdown if the integration implements LifecycleAware
	if lifecycle, ok := integration.(LifecycleAware); ok {
		ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()

		if err := lifecycle.Shutdown(ctx); err != nil {
			r.logger.Warn("integration shutdown error",
				"name", name,
				"error", err,
			)
		}
	}

	delete(r.integrations, name)
	delete(r.metadata, name)

	r.logger.Info("integration unregistered",
		"name", name,
	)

	return nil
}

// List returns a list of all registered integration names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.integrations))
	for name := range r.integrations {
		names = append(names, name)
	}
	return names
}

// ListByType returns a list of integration names filtered by type.
func (r *Registry) ListByType(intType IntegrationType) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0)
	for name, integration := range r.integrations {
		if integration.Type() == intType {
			names = append(names, name)
		}
	}
	return names
}

// GetMetadata returns metadata for a specific integration.
func (r *Registry) GetMetadata(name string) (*Metadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadata, exists := r.metadata[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
	}

	return metadata, nil
}

// ListMetadata returns metadata for all registered integrations.
func (r *Registry) ListMetadata() []*Metadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Metadata, 0, len(r.metadata))
	for _, meta := range r.metadata {
		result = append(result, meta)
	}
	return result
}

// Has checks if an integration is registered.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.integrations[name]
	return exists
}

// Count returns the number of registered integrations.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.integrations)
}

// Execute executes an integration by name with the given configuration and parameters.
func (r *Registry) Execute(ctx context.Context, name string, config *Config, params JSONMap) (*Result, error) {
	integration, err := r.Get(name)
	if err != nil {
		return nil, err
	}

	// Validate configuration
	if err := integration.Validate(config); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return integration.Execute(ctx, config, params)
}

// HealthCheck performs health checks on all registered integrations.
func (r *Registry) HealthCheck(ctx context.Context) map[string]error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]error)
	for name, integration := range r.integrations {
		if healthCheckable, ok := integration.(HealthCheckable); ok {
			results[name] = healthCheckable.HealthCheck(ctx)
		}
	}
	return results
}

// Initialize initializes all registered integrations that implement LifecycleAware.
func (r *Registry) Initialize(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, integration := range r.integrations {
		if lifecycle, ok := integration.(LifecycleAware); ok {
			if err := lifecycle.Initialize(ctx); err != nil {
				return fmt.Errorf("initializing integration %s: %w", name, err)
			}
			r.logger.Info("integration initialized",
				"name", name,
			)
		}
	}
	return nil
}

// Shutdown gracefully shuts down all registered integrations.
func (r *Registry) Shutdown(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var lastErr error
	for name, integration := range r.integrations {
		if lifecycle, ok := integration.(LifecycleAware); ok {
			if err := lifecycle.Shutdown(ctx); err != nil {
				r.logger.Error("integration shutdown error",
					"name", name,
					"error", err,
				)
				lastErr = err
			} else {
				r.logger.Info("integration shut down",
					"name", name,
				)
			}
		}
	}
	return lastErr
}

// Default timeout for shutdown operations.
const defaultShutdownTimeout = 30 * secondsDuration

const secondsDuration = 1000000000 // time.Second in nanoseconds

// Global registry instance.
var globalRegistry *Registry
var globalRegistryOnce sync.Once

// GlobalRegistry returns the global integration registry.
func GlobalRegistry() *Registry {
	globalRegistryOnce.Do(func() {
		globalRegistry = NewRegistry(nil)
	})
	return globalRegistry
}

// Register registers an integration with the global registry.
func Register(integration Integration) error {
	return GlobalRegistry().Register(integration)
}

// Get retrieves an integration from the global registry.
func Get(name string) (Integration, error) {
	return GlobalRegistry().Get(name)
}

// MustGet retrieves an integration from the global registry, panicking on error.
func MustGet(name string) Integration {
	integration, err := GlobalRegistry().Get(name)
	if err != nil {
		panic(fmt.Sprintf("integration not found: %s", name))
	}
	return integration
}
