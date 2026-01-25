package plugin

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/integration"
)

func TestNewManager(t *testing.T) {
	t.Run("creates manager with default config", func(t *testing.T) {
		registry := integration.NewRegistry(slog.Default())
		manager := NewManager(registry, nil)

		require.NotNil(t, manager)
		assert.Empty(t, manager.pluginDirs)
		assert.NotNil(t, manager.plugins)
		assert.NotNil(t, manager.validators)
	})

	t.Run("creates manager with custom config", func(t *testing.T) {
		registry := integration.NewRegistry(slog.Default())
		config := &ManagerConfig{
			PluginDirs:    []string{"/plugins", "/custom"},
			AllowedHashes: []string{"abc123", "def456"},
			Logger:        slog.Default(),
		}
		manager := NewManager(registry, config)

		require.NotNil(t, manager)
		assert.Len(t, manager.pluginDirs, 2)
		assert.True(t, manager.allowedHashes["abc123"])
		assert.True(t, manager.allowedHashes["def456"])
	})
}

func TestManager_LoadPlugin(t *testing.T) {
	t.Run("loads valid plugin", func(t *testing.T) {
		// Create temp directory with manifest
		tmpDir := t.TempDir()
		pluginDir := filepath.Join(tmpDir, "test-plugin")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		manifest := PluginManifest{
			Name:        "test-plugin",
			Version:     "1.0.0",
			Description: "A test plugin",
			Author:      "Test Author",
			Type:        "http",
			EntryPoint:  "main.go",
		}
		manifestBytes, err := json.Marshal(manifest)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(pluginDir, "manifest.json"), manifestBytes, 0644))

		registry := integration.NewRegistry(slog.Default())
		manager := NewManager(registry, nil)

		plugin, err := manager.LoadPlugin(context.Background(), pluginDir)
		require.NoError(t, err)
		require.NotNil(t, plugin)

		assert.Equal(t, "test-plugin", plugin.Name)
		assert.Equal(t, "1.0.0", plugin.Version)
		assert.Equal(t, StateLoaded, plugin.State)
		assert.NotEmpty(t, plugin.ID)
		assert.NotNil(t, plugin.LoadedAt)
	})

	t.Run("fails with missing manifest", func(t *testing.T) {
		tmpDir := t.TempDir()
		registry := integration.NewRegistry(slog.Default())
		manager := NewManager(registry, nil)

		_, err := manager.LoadPlugin(context.Background(), tmpDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reading manifest")
	})

	t.Run("fails with invalid manifest", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "manifest.json"), []byte("invalid json"), 0644))

		registry := integration.NewRegistry(slog.Default())
		manager := NewManager(registry, nil)

		_, err := manager.LoadPlugin(context.Background(), tmpDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parsing manifest")
	})

	t.Run("fails with disallowed hash", func(t *testing.T) {
		tmpDir := t.TempDir()
		pluginDir := filepath.Join(tmpDir, "test-plugin")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		manifest := PluginManifest{
			Name:    "test-plugin",
			Version: "1.0.0",
			Type:    "http",
		}
		manifestBytes, _ := json.Marshal(manifest)
		require.NoError(t, os.WriteFile(filepath.Join(pluginDir, "manifest.json"), manifestBytes, 0644))

		registry := integration.NewRegistry(slog.Default())
		config := &ManagerConfig{
			AllowedHashes: []string{"only-this-hash-allowed"},
		}
		manager := NewManager(registry, config)

		plugin, err := manager.LoadPlugin(context.Background(), pluginDir)
		require.Error(t, err)
		require.NotNil(t, plugin)
		assert.Equal(t, StateError, plugin.State)
		assert.Contains(t, plugin.Error, "not in allowlist")
	})
}

func TestManager_UnloadPlugin(t *testing.T) {
	t.Run("unloads loaded plugin", func(t *testing.T) {
		tmpDir := t.TempDir()
		pluginDir := filepath.Join(tmpDir, "test-plugin")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		manifest := PluginManifest{
			Name:    "test-plugin",
			Version: "1.0.0",
			Type:    "http",
		}
		manifestBytes, _ := json.Marshal(manifest)
		require.NoError(t, os.WriteFile(filepath.Join(pluginDir, "manifest.json"), manifestBytes, 0644))

		registry := integration.NewRegistry(slog.Default())
		manager := NewManager(registry, nil)

		plugin, err := manager.LoadPlugin(context.Background(), pluginDir)
		require.NoError(t, err)

		err = manager.UnloadPlugin(context.Background(), plugin.ID)
		require.NoError(t, err)

		_, err = manager.GetPlugin(plugin.ID)
		assert.Error(t, err)
	})

	t.Run("fails for non-existent plugin", func(t *testing.T) {
		registry := integration.NewRegistry(slog.Default())
		manager := NewManager(registry, nil)

		err := manager.UnloadPlugin(context.Background(), "non-existent-id")
		require.Error(t, err)
	})
}

func TestManager_ReloadPlugin(t *testing.T) {
	t.Run("reloads existing plugin", func(t *testing.T) {
		tmpDir := t.TempDir()
		pluginDir := filepath.Join(tmpDir, "test-plugin")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		manifest := PluginManifest{
			Name:    "test-plugin",
			Version: "1.0.0",
			Type:    "http",
		}
		manifestBytes, _ := json.Marshal(manifest)
		require.NoError(t, os.WriteFile(filepath.Join(pluginDir, "manifest.json"), manifestBytes, 0644))

		registry := integration.NewRegistry(slog.Default())
		manager := NewManager(registry, nil)

		plugin, err := manager.LoadPlugin(context.Background(), pluginDir)
		require.NoError(t, err)
		originalID := plugin.ID

		// Reload
		reloaded, err := manager.ReloadPlugin(context.Background(), originalID)
		require.NoError(t, err)
		assert.Equal(t, plugin.Name, reloaded.Name)
		assert.Equal(t, StateLoaded, reloaded.State)
	})

	t.Run("fails for non-existent plugin", func(t *testing.T) {
		registry := integration.NewRegistry(slog.Default())
		manager := NewManager(registry, nil)

		_, err := manager.ReloadPlugin(context.Background(), "non-existent-id")
		require.Error(t, err)
	})
}

func TestManager_GetPlugin(t *testing.T) {
	t.Run("returns loaded plugin", func(t *testing.T) {
		tmpDir := t.TempDir()
		pluginDir := filepath.Join(tmpDir, "test-plugin")
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		manifest := PluginManifest{
			Name:    "test-plugin",
			Version: "1.0.0",
			Type:    "http",
		}
		manifestBytes, _ := json.Marshal(manifest)
		require.NoError(t, os.WriteFile(filepath.Join(pluginDir, "manifest.json"), manifestBytes, 0644))

		registry := integration.NewRegistry(slog.Default())
		manager := NewManager(registry, nil)

		loaded, err := manager.LoadPlugin(context.Background(), pluginDir)
		require.NoError(t, err)

		retrieved, err := manager.GetPlugin(loaded.ID)
		require.NoError(t, err)
		assert.Equal(t, loaded.ID, retrieved.ID)
	})

	t.Run("returns error for missing plugin", func(t *testing.T) {
		registry := integration.NewRegistry(slog.Default())
		manager := NewManager(registry, nil)

		_, err := manager.GetPlugin("missing-id")
		require.Error(t, err)
	})
}

func TestManager_ListPlugins(t *testing.T) {
	tmpDir := t.TempDir()

	for _, name := range []string{"plugin1", "plugin2", "plugin3"} {
		pluginDir := filepath.Join(tmpDir, name)
		require.NoError(t, os.MkdirAll(pluginDir, 0755))

		manifest := PluginManifest{
			Name:    name,
			Version: "1.0.0",
			Type:    "http",
		}
		manifestBytes, _ := json.Marshal(manifest)
		require.NoError(t, os.WriteFile(filepath.Join(pluginDir, "manifest.json"), manifestBytes, 0644))
	}

	registry := integration.NewRegistry(slog.Default())
	manager := NewManager(registry, nil)

	for _, name := range []string{"plugin1", "plugin2", "plugin3"} {
		_, err := manager.LoadPlugin(context.Background(), filepath.Join(tmpDir, name))
		require.NoError(t, err)
	}

	plugins := manager.ListPlugins()
	assert.Len(t, plugins, 3)
}

func TestManager_DiscoverPlugins(t *testing.T) {
	t.Run("discovers plugins in directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create two plugin directories
		for _, name := range []string{"pluginA", "pluginB"} {
			pluginDir := filepath.Join(tmpDir, name)
			require.NoError(t, os.MkdirAll(pluginDir, 0755))

			manifest := PluginManifest{
				Name:    name,
				Version: "1.0.0",
				Type:    "http",
			}
			manifestBytes, _ := json.Marshal(manifest)
			require.NoError(t, os.WriteFile(filepath.Join(pluginDir, "manifest.json"), manifestBytes, 0644))
		}

		registry := integration.NewRegistry(slog.Default())
		config := &ManagerConfig{
			PluginDirs: []string{tmpDir},
		}
		manager := NewManager(registry, config)

		plugins, err := manager.DiscoverPlugins(context.Background())
		require.NoError(t, err)
		assert.Len(t, plugins, 2)
	})

	t.Run("handles missing directory", func(t *testing.T) {
		registry := integration.NewRegistry(slog.Default())
		config := &ManagerConfig{
			PluginDirs: []string{"/nonexistent/path"},
		}
		manager := NewManager(registry, config)

		plugins, err := manager.DiscoverPlugins(context.Background())
		require.NoError(t, err)
		assert.Empty(t, plugins)
	})
}

func TestManager_AddValidator(t *testing.T) {
	registry := integration.NewRegistry(slog.Default())
	manager := NewManager(registry, nil)

	initialCount := len(manager.validators)

	manager.AddValidator(&ManifestValidator{})
	assert.Len(t, manager.validators, initialCount+1)
}

func TestManager_AddAllowedHash(t *testing.T) {
	registry := integration.NewRegistry(slog.Default())
	manager := NewManager(registry, nil)

	manager.AddAllowedHash("newhash123")
	assert.True(t, manager.allowedHashes["newhash123"])
}

func TestManager_Shutdown(t *testing.T) {
	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, "test-plugin")
	require.NoError(t, os.MkdirAll(pluginDir, 0755))

	manifest := PluginManifest{
		Name:    "test-plugin",
		Version: "1.0.0",
		Type:    "http",
	}
	manifestBytes, _ := json.Marshal(manifest)
	require.NoError(t, os.WriteFile(filepath.Join(pluginDir, "manifest.json"), manifestBytes, 0644))

	registry := integration.NewRegistry(slog.Default())
	manager := NewManager(registry, nil)

	_, err := manager.LoadPlugin(context.Background(), pluginDir)
	require.NoError(t, err)

	err = manager.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestPluginState(t *testing.T) {
	assert.Equal(t, PluginState("unloaded"), StateUnloaded)
	assert.Equal(t, PluginState("loading"), StateLoading)
	assert.Equal(t, PluginState("loaded"), StateLoaded)
	assert.Equal(t, PluginState("error"), StateError)
}

func TestGeneratePluginID(t *testing.T) {
	id1 := generatePluginID("test", "1.0.0")
	id2 := generatePluginID("test", "1.0.0")
	id3 := generatePluginID("test", "2.0.0")

	// Same input should produce same output
	assert.Equal(t, id1, id2)
	// Different version should produce different ID
	assert.NotEqual(t, id1, id3)
	// ID should be non-empty
	assert.NotEmpty(t, id1)
}

func TestPluginIntegration(t *testing.T) {
	manifest := &PluginManifest{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test plugin",
		Author:      "Test Author",
		Homepage:    "https://example.com",
		License:     "MIT",
		Permissions: []string{"http"},
		Schema: &integration.Schema{
			ConfigSpec: map[string]integration.FieldSpec{
				"url": {Name: "url", Type: integration.FieldTypeString},
			},
		},
	}

	pi := &pluginIntegration{
		base:     integration.NewBaseIntegration("test", integration.TypePlugin),
		manifest: manifest,
	}

	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, "test", pi.Name())
	})

	t.Run("Type", func(t *testing.T) {
		assert.Equal(t, integration.TypePlugin, pi.Type())
	})

	t.Run("Execute returns not implemented", func(t *testing.T) {
		_, err := pi.Execute(context.Background(), nil, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
	})

	t.Run("Validate", func(t *testing.T) {
		config := &integration.Config{
			Name:     "test-config",
			Type:     integration.TypePlugin,
			Settings: integration.JSONMap{},
		}
		err := pi.Validate(config)
		require.NoError(t, err)

		// Nil config should fail
		err = pi.Validate(nil)
		require.Error(t, err)
	})

	t.Run("GetSchema", func(t *testing.T) {
		schema := pi.GetSchema()
		require.NotNil(t, schema)
		assert.Contains(t, schema.ConfigSpec, "url")
	})

	t.Run("GetSchema returns base schema when manifest has none", func(t *testing.T) {
		pi2 := &pluginIntegration{
			base:     integration.NewBaseIntegration("test2", integration.TypePlugin),
			manifest: &PluginManifest{Name: "test2", Version: "1.0.0"},
		}
		schema := pi2.GetSchema()
		require.NotNil(t, schema)
	})

	t.Run("GetMetadata", func(t *testing.T) {
		metadata := pi.GetMetadata()
		require.NotNil(t, metadata)
		assert.Equal(t, "test", metadata.Name)
		assert.Equal(t, "1.0.0", metadata.Version)
		assert.Equal(t, "Test plugin", metadata.Description)
		assert.Equal(t, "Test Author", metadata.Author)
		assert.Equal(t, "https://example.com", metadata.Homepage)
		assert.Equal(t, "MIT", metadata.License)
		assert.Contains(t, metadata.Permissions, "http")
	})
}

func TestPluginManifest(t *testing.T) {
	manifest := PluginManifest{
		Name:         "my-plugin",
		Version:      "2.1.0",
		Description:  "A useful plugin",
		Author:       "Developer",
		Homepage:     "https://github.com/example/plugin",
		License:      "Apache-2.0",
		EntryPoint:   "main.go",
		Type:         "api",
		Permissions:  []string{"network", "storage"},
		Dependencies: []string{"other-plugin@1.0.0"},
		Metadata: integration.JSONMap{
			"custom": "value",
		},
	}

	data, err := json.Marshal(manifest)
	require.NoError(t, err)

	var decoded PluginManifest
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, manifest.Name, decoded.Name)
	assert.Equal(t, manifest.Version, decoded.Version)
	assert.Equal(t, manifest.Dependencies, decoded.Dependencies)
}

func TestPlugin(t *testing.T) {
	now := time.Now()
	plugin := Plugin{
		ID:          "plugin-123",
		Name:        "test-plugin",
		Version:     "1.0.0",
		Description: "Test description",
		Author:      "Test Author",
		State:       StateLoaded,
		Path:        "/path/to/plugin",
		Hash:        "abc123hash",
		LoadedAt:    &now,
		Metadata: integration.JSONMap{
			"key": "value",
		},
	}

	data, err := json.Marshal(plugin)
	require.NoError(t, err)

	var decoded Plugin
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, plugin.ID, decoded.ID)
	assert.Equal(t, plugin.Name, decoded.Name)
	assert.Equal(t, plugin.State, decoded.State)
}
