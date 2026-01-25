package plugin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/integration"
)

func TestManifestValidator(t *testing.T) {
	validator := &ManifestValidator{}

	t.Run("validates valid manifest", func(t *testing.T) {
		manifest := &PluginManifest{
			Name:    "valid-plugin",
			Version: "1.0.0",
			Type:    "http",
		}
		err := validator.Validate(nil, manifest, "")
		require.NoError(t, err)
	})

	t.Run("fails with nil manifest", func(t *testing.T) {
		err := validator.Validate(nil, nil, "")
		require.Error(t, err)
	})

	t.Run("fails with empty name", func(t *testing.T) {
		manifest := &PluginManifest{
			Name:    "",
			Version: "1.0.0",
			Type:    "http",
		}
		err := validator.Validate(nil, manifest, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("fails with invalid name format", func(t *testing.T) {
		testCases := []string{
			"123invalid",
			"has spaces",
			"special@chars",
			"_underscore_start",
		}

		for _, name := range testCases {
			manifest := &PluginManifest{
				Name:    name,
				Version: "1.0.0",
				Type:    "http",
			}
			err := validator.Validate(nil, manifest, "")
			require.Error(t, err, "expected error for name: %s", name)
		}
	})

	t.Run("accepts valid name formats", func(t *testing.T) {
		testCases := []string{
			"valid",
			"ValidPlugin",
			"my-plugin",
			"my_plugin",
			"plugin123",
			"My-Plugin_v2",
		}

		for _, name := range testCases {
			manifest := &PluginManifest{
				Name:    name,
				Version: "1.0.0",
				Type:    "http",
			}
			err := validator.Validate(nil, manifest, "")
			require.NoError(t, err, "expected no error for name: %s", name)
		}
	})

	t.Run("fails with empty version", func(t *testing.T) {
		manifest := &PluginManifest{
			Name:    "valid-plugin",
			Version: "",
			Type:    "http",
		}
		err := validator.Validate(nil, manifest, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "version")
	})

	t.Run("fails with invalid version format", func(t *testing.T) {
		testCases := []string{
			"1.0",
			"v1.0.0",
			"1.0.0.0",
			"invalid",
		}

		for _, version := range testCases {
			manifest := &PluginManifest{
				Name:    "valid-plugin",
				Version: version,
				Type:    "http",
			}
			err := validator.Validate(nil, manifest, "")
			require.Error(t, err, "expected error for version: %s", version)
		}
	})

	t.Run("accepts valid version formats", func(t *testing.T) {
		testCases := []string{
			"1.0.0",
			"0.0.1",
			"10.20.30",
			"1.0.0-alpha",
			"2.3.4-beta",
		}

		for _, version := range testCases {
			manifest := &PluginManifest{
				Name:    "valid-plugin",
				Version: version,
				Type:    "http",
			}
			err := validator.Validate(nil, manifest, "")
			require.NoError(t, err, "expected no error for version: %s", version)
		}
	})

	t.Run("fails with empty type", func(t *testing.T) {
		manifest := &PluginManifest{
			Name:    "valid-plugin",
			Version: "1.0.0",
			Type:    "",
		}
		err := validator.Validate(nil, manifest, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "type")
	})
}

func TestPathValidator(t *testing.T) {
	validator := &PathValidator{}

	t.Run("validates valid directory path", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := validator.Validate(nil, nil, tmpDir)
		require.NoError(t, err)
	})

	t.Run("fails with non-existent path", func(t *testing.T) {
		err := validator.Validate(nil, nil, "/nonexistent/path/to/plugin")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("fails with path traversal", func(t *testing.T) {
		err := validator.Validate(nil, nil, "/some/path/../../../etc/passwd")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "path traversal")
	})

	t.Run("fails when path is a file not directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "file.txt")
		require.NoError(t, os.WriteFile(filePath, []byte("content"), 0644))

		err := validator.Validate(nil, nil, filePath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a directory")
	})
}

func TestPermissionValidator(t *testing.T) {
	t.Run("validates manifest with no permissions", func(t *testing.T) {
		validator := &PermissionValidator{}
		manifest := &PluginManifest{
			Permissions: nil,
		}
		err := validator.Validate(nil, manifest, "")
		require.NoError(t, err)
	})

	t.Run("validates allowed permissions", func(t *testing.T) {
		validator := &PermissionValidator{
			AllowedPermissions: []string{"network", "storage", "api"},
		}
		manifest := &PluginManifest{
			Permissions: []string{"network", "api"},
		}
		err := validator.Validate(nil, manifest, "")
		require.NoError(t, err)
	})

	t.Run("fails with denied permissions", func(t *testing.T) {
		validator := &PermissionValidator{
			DeniedPermissions: []string{"filesystem", "exec"},
		}
		manifest := &PluginManifest{
			Permissions: []string{"network", "filesystem"},
		}
		err := validator.Validate(nil, manifest, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})

	t.Run("fails with non-allowed permissions when allowlist set", func(t *testing.T) {
		validator := &PermissionValidator{
			AllowedPermissions: []string{"network", "api"},
		}
		manifest := &PluginManifest{
			Permissions: []string{"network", "storage"},
		}
		err := validator.Validate(nil, manifest, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "permission not allowed")
	})

	t.Run("validates when no allowlist configured", func(t *testing.T) {
		validator := &PermissionValidator{}
		manifest := &PluginManifest{
			Permissions: []string{"anything", "goes"},
		}
		err := validator.Validate(nil, manifest, "")
		require.NoError(t, err)
	})
}

func TestDependencyValidator(t *testing.T) {
	t.Run("validates manifest with no dependencies", func(t *testing.T) {
		validator := &DependencyValidator{}
		manifest := &PluginManifest{
			Dependencies: nil,
		}
		err := validator.Validate(nil, manifest, "")
		require.NoError(t, err)
	})

	t.Run("skips validation when AvailablePlugins not configured", func(t *testing.T) {
		validator := &DependencyValidator{}
		manifest := &PluginManifest{
			Dependencies: []string{"some-dep"},
		}
		err := validator.Validate(nil, manifest, "")
		require.NoError(t, err)
	})

	t.Run("validates all dependencies available", func(t *testing.T) {
		validator := &DependencyValidator{
			AvailablePlugins: map[string]bool{
				"dep1": true,
				"dep2": true,
			},
		}
		manifest := &PluginManifest{
			Dependencies: []string{"dep1", "dep2"},
		}
		err := validator.Validate(nil, manifest, "")
		require.NoError(t, err)
	})

	t.Run("fails when dependency missing", func(t *testing.T) {
		validator := &DependencyValidator{
			AvailablePlugins: map[string]bool{
				"dep1": true,
			},
		}
		manifest := &PluginManifest{
			Dependencies: []string{"dep1", "missing-dep"},
		}
		err := validator.Validate(nil, manifest, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing dependency")
	})

	t.Run("handles versioned dependencies", func(t *testing.T) {
		validator := &DependencyValidator{
			AvailablePlugins: map[string]bool{
				"dep1": true,
			},
		}
		manifest := &PluginManifest{
			Dependencies: []string{"dep1@1.0.0"},
		}
		err := validator.Validate(nil, manifest, "")
		require.NoError(t, err)
	})
}

func TestSchemaValidator(t *testing.T) {
	validator := &SchemaValidator{}

	t.Run("validates nil schema", func(t *testing.T) {
		manifest := &PluginManifest{
			Schema: nil,
		}
		err := validator.Validate(nil, manifest, "")
		require.NoError(t, err)
	})

	t.Run("validates valid schema", func(t *testing.T) {
		manifest := &PluginManifest{
			Schema: &integration.Schema{
				ConfigSpec: map[string]integration.FieldSpec{
					"url": {Name: "url", Type: integration.FieldTypeString},
				},
				InputSpec: map[string]integration.FieldSpec{
					"data": {Name: "data", Type: integration.FieldTypeObject},
				},
				OutputSpec: map[string]integration.FieldSpec{
					"result": {Name: "result", Type: integration.FieldTypeBoolean},
				},
			},
		}
		err := validator.Validate(nil, manifest, "")
		require.NoError(t, err)
	})

	t.Run("fails with invalid field type in config spec", func(t *testing.T) {
		manifest := &PluginManifest{
			Schema: &integration.Schema{
				ConfigSpec: map[string]integration.FieldSpec{
					"url": {Name: "url", Type: "invalid_type"},
				},
			},
		}
		err := validator.Validate(nil, manifest, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "config spec")
	})

	t.Run("fails with invalid field type in input spec", func(t *testing.T) {
		manifest := &PluginManifest{
			Schema: &integration.Schema{
				InputSpec: map[string]integration.FieldSpec{
					"data": {Name: "data", Type: "bad_type"},
				},
			},
		}
		err := validator.Validate(nil, manifest, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "input spec")
	})

	t.Run("fails with invalid field type in output spec", func(t *testing.T) {
		manifest := &PluginManifest{
			Schema: &integration.Schema{
				OutputSpec: map[string]integration.FieldSpec{
					"result": {Name: "result", Type: "unknown"},
				},
			},
		}
		err := validator.Validate(nil, manifest, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "output spec")
	})
}

func TestValidateFieldSpec(t *testing.T) {
	t.Run("fails with empty field name", func(t *testing.T) {
		err := validateFieldSpec("", integration.FieldSpec{Type: integration.FieldTypeString})
		require.Error(t, err)
	})

	t.Run("fails with invalid field type", func(t *testing.T) {
		err := validateFieldSpec("field", integration.FieldSpec{Type: "invalid"})
		require.Error(t, err)
	})

	t.Run("validates all valid field types", func(t *testing.T) {
		validTypes := []integration.FieldType{
			integration.FieldTypeString,
			integration.FieldTypeNumber,
			integration.FieldTypeInteger,
			integration.FieldTypeBoolean,
			integration.FieldTypeArray,
			integration.FieldTypeObject,
			integration.FieldTypeSecret,
		}

		for _, ft := range validTypes {
			err := validateFieldSpec("field", integration.FieldSpec{Type: ft})
			require.NoError(t, err, "expected no error for field type: %s", ft)
		}
	})
}

func TestCompositeValidator(t *testing.T) {
	t.Run("runs all validators", func(t *testing.T) {
		composite := NewCompositeValidator(
			&ManifestValidator{},
			&PermissionValidator{},
		)

		manifest := &PluginManifest{
			Name:    "test-plugin",
			Version: "1.0.0",
			Type:    "http",
		}
		err := composite.Validate(nil, manifest, "")
		require.NoError(t, err)
	})

	t.Run("stops on first error", func(t *testing.T) {
		composite := NewCompositeValidator(
			&ManifestValidator{},
			&PathValidator{},
		)

		manifest := &PluginManifest{
			Name:    "test-plugin",
			Version: "1.0.0",
			Type:    "http",
		}
		// Path doesn't exist
		err := composite.Validate(nil, manifest, "/nonexistent/path")
		require.Error(t, err)
	})

	t.Run("Add appends validators", func(t *testing.T) {
		composite := NewCompositeValidator()
		assert.Empty(t, composite.validators)

		composite.Add(&ManifestValidator{})
		assert.Len(t, composite.validators, 1)

		composite.Add(&PathValidator{})
		assert.Len(t, composite.validators, 2)
	})
}

func TestDefaultValidators(t *testing.T) {
	validators := defaultValidators()

	assert.Len(t, validators, 3)

	// Check types
	_, ok := validators[0].(*ManifestValidator)
	assert.True(t, ok)

	_, ok = validators[1].(*PathValidator)
	assert.True(t, ok)

	_, ok = validators[2].(*PermissionValidator)
	assert.True(t, ok)
}
