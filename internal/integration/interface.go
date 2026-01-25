package integration

import (
	"context"
)

// Integration defines the interface for all integrations.
// Each integration must implement these methods to be registered
// and executed within the Gorax workflow engine.
type Integration interface {
	// Name returns the unique name identifier for the integration.
	Name() string

	// Type returns the type of integration (http, webhook, api, etc.).
	Type() IntegrationType

	// Execute performs the integration operation with the given configuration and parameters.
	// Returns an IntegrationResult containing success/failure status and data.
	Execute(ctx context.Context, config *Config, params JSONMap) (*Result, error)

	// Validate validates the integration configuration.
	// Returns an error if the configuration is invalid.
	Validate(config *Config) error

	// GetSchema returns the schema specification for this integration.
	GetSchema() *Schema

	// GetMetadata returns metadata about the integration.
	GetMetadata() *Metadata
}

// LifecycleAware is an optional interface for integrations that need
// lifecycle management (initialization and cleanup).
type LifecycleAware interface {
	// Initialize is called when the integration is loaded.
	Initialize(ctx context.Context) error

	// Shutdown is called when the integration is being unloaded.
	Shutdown(ctx context.Context) error
}

// HealthCheckable is an optional interface for integrations that support
// health checking.
type HealthCheckable interface {
	// HealthCheck performs a health check on the integration.
	// Returns nil if healthy, error otherwise.
	HealthCheck(ctx context.Context) error
}

// Refreshable is an optional interface for integrations that support
// credential refresh (e.g., OAuth2 token refresh).
type Refreshable interface {
	// RefreshCredentials refreshes the integration's credentials.
	RefreshCredentials(ctx context.Context, creds *Credentials) (*Credentials, error)
}

// Configurable is an optional interface for integrations that support
// dynamic configuration updates.
type Configurable interface {
	// UpdateConfig updates the integration's configuration at runtime.
	UpdateConfig(ctx context.Context, config *Config) error
}

// BaseIntegration provides common functionality for integrations.
// Embed this struct in custom integrations to get default implementations.
type BaseIntegration struct {
	name     string
	intType  IntegrationType
	metadata *Metadata
	schema   *Schema
}

// NewBaseIntegration creates a new BaseIntegration.
func NewBaseIntegration(name string, intType IntegrationType) *BaseIntegration {
	return &BaseIntegration{
		name:    name,
		intType: intType,
		metadata: &Metadata{
			Name:    name,
			Version: "1.0.0",
		},
		schema: &Schema{
			ConfigSpec: make(map[string]FieldSpec),
			InputSpec:  make(map[string]FieldSpec),
			OutputSpec: make(map[string]FieldSpec),
		},
	}
}

// Name returns the integration name.
func (b *BaseIntegration) Name() string {
	return b.name
}

// Type returns the integration type.
func (b *BaseIntegration) Type() IntegrationType {
	return b.intType
}

// GetSchema returns the integration schema.
func (b *BaseIntegration) GetSchema() *Schema {
	return b.schema
}

// SetSchema sets the integration schema.
func (b *BaseIntegration) SetSchema(schema *Schema) {
	b.schema = schema
}

// GetMetadata returns the integration metadata.
func (b *BaseIntegration) GetMetadata() *Metadata {
	return b.metadata
}

// SetMetadata sets the integration metadata.
func (b *BaseIntegration) SetMetadata(metadata *Metadata) {
	b.metadata = metadata
}

// ValidateConfig provides basic configuration validation.
// Override this in concrete integrations for custom validation.
func (b *BaseIntegration) ValidateConfig(config *Config) error {
	if config == nil {
		return NewValidationError("config", "configuration is required", nil)
	}
	if config.Name == "" {
		return NewValidationError("name", "name is required", nil)
	}
	if !config.Type.Valid() {
		return NewValidationError("type", "invalid integration type", config.Type)
	}
	return nil
}

// ValidateSchema validates that the given data matches the schema specification.
func ValidateSchema(data JSONMap, spec map[string]FieldSpec) error {
	for name, field := range spec {
		value, exists := data[name]

		if field.Required && !exists {
			return NewValidationError(name, "required field is missing", nil)
		}

		if !exists {
			continue
		}

		if err := validateFieldType(name, value, field); err != nil {
			return err
		}
	}
	return nil
}

// validateFieldType validates that a value matches the expected field type.
func validateFieldType(name string, value any, spec FieldSpec) error {
	if value == nil {
		if spec.Required {
			return NewValidationError(name, "required field cannot be null", nil)
		}
		return nil
	}

	switch spec.Type {
	case FieldTypeString, FieldTypeSecret:
		if _, ok := value.(string); !ok {
			return NewValidationError(name, "expected string value", value)
		}
	case FieldTypeNumber:
		switch value.(type) {
		case float64, float32, int, int64, int32:
			// Valid number types
		default:
			return NewValidationError(name, "expected number value", value)
		}
	case FieldTypeInteger:
		switch v := value.(type) {
		case int, int64, int32:
			// Valid integer types
		case float64:
			// Check if it's a whole number
			if v != float64(int64(v)) {
				return NewValidationError(name, "expected integer value", value)
			}
		default:
			return NewValidationError(name, "expected integer value", value)
		}
	case FieldTypeBoolean:
		if _, ok := value.(bool); !ok {
			return NewValidationError(name, "expected boolean value", value)
		}
	case FieldTypeArray:
		if _, ok := value.([]any); !ok {
			return NewValidationError(name, "expected array value", value)
		}
	case FieldTypeObject:
		switch value.(type) {
		case map[string]any, JSONMap:
			// Valid object types
		default:
			return NewValidationError(name, "expected object value", value)
		}
	}

	return nil
}
