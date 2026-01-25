package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/gorax/gorax/internal/integration"
)

// Validator validates plugins before loading.
type Validator interface {
	// Validate validates the plugin.
	Validate(plugin *Plugin, manifest *PluginManifest, path string) error
}

// defaultValidators returns the default set of validators.
func defaultValidators() []Validator {
	return []Validator{
		&ManifestValidator{},
		&PathValidator{},
		&PermissionValidator{},
	}
}

// ManifestValidator validates the plugin manifest.
type ManifestValidator struct{}

// Validate validates the manifest fields.
func (v *ManifestValidator) Validate(_ *Plugin, manifest *PluginManifest, _ string) error {
	if manifest == nil {
		return integration.NewValidationError("manifest", "manifest is required", nil)
	}

	if manifest.Name == "" {
		return integration.NewValidationError("name", "plugin name is required", nil)
	}

	// Validate name format (alphanumeric, hyphens, underscores)
	namePattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	if !namePattern.MatchString(manifest.Name) {
		return integration.NewValidationError("name", "invalid plugin name format", manifest.Name)
	}

	if manifest.Version == "" {
		return integration.NewValidationError("version", "plugin version is required", nil)
	}

	// Validate semver format
	semverPattern := regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?$`)
	if !semverPattern.MatchString(manifest.Version) {
		return integration.NewValidationError("version", "invalid version format (expected semver)", manifest.Version)
	}

	if manifest.Type == "" {
		return integration.NewValidationError("type", "plugin type is required", nil)
	}

	return nil
}

// PathValidator validates the plugin path for security.
type PathValidator struct{}

// Validate ensures the path is safe.
func (v *PathValidator) Validate(_ *Plugin, _ *PluginManifest, path string) error {
	// Ensure absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return integration.NewValidationError("path", "invalid path", path)
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return integration.NewValidationError("path", "path traversal detected", path)
	}

	// Verify path exists and is a directory
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return integration.NewValidationError("path", "path does not exist", path)
		}
		return fmt.Errorf("checking path: %w", err)
	}

	if !info.IsDir() {
		return integration.NewValidationError("path", "path must be a directory", path)
	}

	return nil
}

// PermissionValidator validates plugin permissions.
type PermissionValidator struct {
	// AllowedPermissions is the list of permissions that can be granted.
	AllowedPermissions []string
	// DeniedPermissions is the list of permissions that cannot be granted.
	DeniedPermissions []string
}

// Validate validates that requested permissions are allowed.
func (v *PermissionValidator) Validate(_ *Plugin, manifest *PluginManifest, _ string) error {
	if len(manifest.Permissions) == 0 {
		return nil
	}

	// Check for denied permissions
	for _, perm := range manifest.Permissions {
		if slices.Contains(v.DeniedPermissions, perm) {
			return integration.NewValidationError("permissions", "permission denied", perm)
		}
	}

	// If allowed permissions are configured, check against them
	if len(v.AllowedPermissions) > 0 {
		allowedSet := make(map[string]bool)
		for _, perm := range v.AllowedPermissions {
			allowedSet[perm] = true
		}

		for _, perm := range manifest.Permissions {
			if !allowedSet[perm] {
				return integration.NewValidationError("permissions", "permission not allowed", perm)
			}
		}
	}

	return nil
}

// DependencyValidator validates plugin dependencies.
type DependencyValidator struct {
	// AvailablePlugins is a map of available plugin names for dependency checking.
	AvailablePlugins map[string]bool
}

// Validate validates that all dependencies are available.
func (v *DependencyValidator) Validate(_ *Plugin, manifest *PluginManifest, _ string) error {
	if len(manifest.Dependencies) == 0 {
		return nil
	}

	if v.AvailablePlugins == nil {
		return nil // Skip dependency checking if not configured
	}

	for _, dep := range manifest.Dependencies {
		// Parse dependency (format: "name" or "name@version")
		depName := dep
		if idx := strings.Index(dep, "@"); idx > 0 {
			depName = dep[:idx]
		}

		if !v.AvailablePlugins[depName] {
			return integration.NewValidationError("dependencies", "missing dependency", dep)
		}
	}

	return nil
}

// SchemaValidator validates the plugin schema.
type SchemaValidator struct{}

// Validate validates the plugin schema.
func (v *SchemaValidator) Validate(_ *Plugin, manifest *PluginManifest, _ string) error {
	if manifest.Schema == nil {
		return nil // Schema is optional
	}

	// Validate config spec
	for name, field := range manifest.Schema.ConfigSpec {
		if err := validateFieldSpec(name, field); err != nil {
			return fmt.Errorf("config spec: %w", err)
		}
	}

	// Validate input spec
	for name, field := range manifest.Schema.InputSpec {
		if err := validateFieldSpec(name, field); err != nil {
			return fmt.Errorf("input spec: %w", err)
		}
	}

	// Validate output spec
	for name, field := range manifest.Schema.OutputSpec {
		if err := validateFieldSpec(name, field); err != nil {
			return fmt.Errorf("output spec: %w", err)
		}
	}

	return nil
}

// validateFieldSpec validates a single field specification.
func validateFieldSpec(name string, field integration.FieldSpec) error {
	if name == "" {
		return integration.NewValidationError("field", "field name is required", nil)
	}

	if !field.Type.Valid() {
		return integration.NewValidationError(name, "invalid field type", field.Type)
	}

	return nil
}

// CompositeValidator combines multiple validators.
type CompositeValidator struct {
	validators []Validator
}

// NewCompositeValidator creates a new composite validator.
func NewCompositeValidator(validators ...Validator) *CompositeValidator {
	return &CompositeValidator{validators: validators}
}

// Validate runs all validators.
func (v *CompositeValidator) Validate(plugin *Plugin, manifest *PluginManifest, path string) error {
	for _, validator := range v.validators {
		if err := validator.Validate(plugin, manifest, path); err != nil {
			return err
		}
	}
	return nil
}

// Add adds a validator to the composite.
func (v *CompositeValidator) Add(validator Validator) {
	v.validators = append(v.validators, validator)
}
