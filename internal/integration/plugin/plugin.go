// Package plugin provides a plugin system for loading external integrations.
package plugin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorax/gorax/internal/integration"
)

// PluginState represents the state of a plugin.
type PluginState string

const (
	// StateUnloaded means the plugin is not loaded.
	StateUnloaded PluginState = "unloaded"
	// StateLoading means the plugin is being loaded.
	StateLoading PluginState = "loading"
	// StateLoaded means the plugin is loaded and ready.
	StateLoaded PluginState = "loaded"
	// StateError means the plugin failed to load.
	StateError PluginState = "error"
)

// Plugin represents a loaded integration plugin.
type Plugin struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Version     string              `json:"version"`
	Description string              `json:"description,omitempty"`
	Author      string              `json:"author,omitempty"`
	State       PluginState         `json:"state"`
	Path        string              `json:"path,omitempty"`
	Hash        string              `json:"hash,omitempty"`
	LoadedAt    *time.Time          `json:"loaded_at,omitempty"`
	Error       string              `json:"error,omitempty"`
	Metadata    integration.JSONMap `json:"metadata,omitempty"`

	integration integration.Integration
}

// PluginManifest defines the manifest file for a plugin.
type PluginManifest struct {
	Name         string              `json:"name"`
	Version      string              `json:"version"`
	Description  string              `json:"description,omitempty"`
	Author       string              `json:"author,omitempty"`
	Homepage     string              `json:"homepage,omitempty"`
	License      string              `json:"license,omitempty"`
	EntryPoint   string              `json:"entry_point"`
	Type         string              `json:"type"`
	Permissions  []string            `json:"permissions,omitempty"`
	Dependencies []string            `json:"dependencies,omitempty"`
	Schema       *integration.Schema `json:"schema,omitempty"`
	Metadata     integration.JSONMap `json:"metadata,omitempty"`
}

// Manager manages the lifecycle of integration plugins.
type Manager struct {
	plugins       map[string]*Plugin
	pluginDirs    []string
	registry      *integration.Registry
	logger        *slog.Logger
	validators    []Validator
	allowedHashes map[string]bool // Hash allowlist for security
	mu            sync.RWMutex
}

// ManagerConfig holds configuration for the plugin manager.
type ManagerConfig struct {
	PluginDirs    []string
	AllowedHashes []string
	Logger        *slog.Logger
}

// NewManager creates a new plugin manager.
func NewManager(registry *integration.Registry, config *ManagerConfig) *Manager {
	if config == nil {
		config = &ManagerConfig{}
	}

	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	allowedHashes := make(map[string]bool)
	for _, hash := range config.AllowedHashes {
		allowedHashes[hash] = true
	}

	return &Manager{
		plugins:       make(map[string]*Plugin),
		pluginDirs:    config.PluginDirs,
		registry:      registry,
		logger:        logger,
		validators:    defaultValidators(),
		allowedHashes: allowedHashes,
	}
}

// LoadPlugin loads a plugin from the specified path.
func (m *Manager) LoadPlugin(ctx context.Context, path string) (*Plugin, error) {
	m.logger.Info("loading plugin",
		"path", path,
	)

	// Read manifest
	manifest, err := m.readManifest(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}

	// Create plugin
	plugin := &Plugin{
		ID:          generatePluginID(manifest.Name, manifest.Version),
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
		State:       StateLoading,
		Path:        path,
		Metadata:    manifest.Metadata,
	}

	// Validate plugin
	if err := m.validatePlugin(plugin, manifest, path); err != nil {
		plugin.State = StateError
		plugin.Error = err.Error()
		return plugin, fmt.Errorf("validating plugin: %w", err)
	}

	// Calculate hash
	hash, err := m.calculateHash(path)
	if err != nil {
		m.logger.Warn("failed to calculate plugin hash",
			"path", path,
			"error", err,
		)
	}
	plugin.Hash = hash

	// Check hash allowlist if configured
	if len(m.allowedHashes) > 0 && hash != "" {
		if !m.allowedHashes[hash] {
			plugin.State = StateError
			plugin.Error = "plugin hash not in allowlist"
			return plugin, integration.ErrPluginInvalid
		}
	}

	// Create integration from manifest
	integ, err := m.createIntegrationFromManifest(manifest)
	if err != nil {
		plugin.State = StateError
		plugin.Error = err.Error()
		return plugin, fmt.Errorf("creating integration: %w", err)
	}

	// Initialize if lifecycle-aware
	if lifecycle, ok := integ.(integration.LifecycleAware); ok {
		if err := lifecycle.Initialize(ctx); err != nil {
			plugin.State = StateError
			plugin.Error = err.Error()
			return plugin, fmt.Errorf("initializing integration: %w", err)
		}
	}

	// Register with integration registry
	if err := m.registry.Register(integ); err != nil {
		plugin.State = StateError
		plugin.Error = err.Error()
		return plugin, fmt.Errorf("registering integration: %w", err)
	}

	plugin.integration = integ
	plugin.State = StateLoaded
	now := time.Now()
	plugin.LoadedAt = &now

	// Store plugin
	m.mu.Lock()
	m.plugins[plugin.ID] = plugin
	m.mu.Unlock()

	m.logger.Info("plugin loaded successfully",
		"name", plugin.Name,
		"version", plugin.Version,
		"id", plugin.ID,
	)

	return plugin, nil
}

// UnloadPlugin unloads a plugin by ID.
func (m *Manager) UnloadPlugin(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[id]
	if !exists {
		return fmt.Errorf("%w: %s", integration.ErrPluginNotLoaded, id)
	}

	// Shutdown if lifecycle-aware
	if plugin.integration != nil {
		if lifecycle, ok := plugin.integration.(integration.LifecycleAware); ok {
			if err := lifecycle.Shutdown(ctx); err != nil {
				m.logger.Warn("plugin shutdown error",
					"id", id,
					"error", err,
				)
			}
		}

		// Unregister from registry
		if err := m.registry.Unregister(plugin.integration.Name()); err != nil {
			m.logger.Warn("failed to unregister plugin integration",
				"id", id,
				"error", err,
			)
		}
	}

	delete(m.plugins, id)
	plugin.State = StateUnloaded

	m.logger.Info("plugin unloaded",
		"name", plugin.Name,
		"id", id,
	)

	return nil
}

// ReloadPlugin reloads a plugin by ID.
func (m *Manager) ReloadPlugin(ctx context.Context, id string) (*Plugin, error) {
	m.mu.RLock()
	plugin, exists := m.plugins[id]
	if !exists {
		m.mu.RUnlock()
		return nil, fmt.Errorf("%w: %s", integration.ErrPluginNotLoaded, id)
	}
	path := plugin.Path
	m.mu.RUnlock()

	// Unload first
	if err := m.UnloadPlugin(ctx, id); err != nil {
		return nil, fmt.Errorf("unloading plugin: %w", err)
	}

	// Load again
	return m.LoadPlugin(ctx, path)
}

// GetPlugin returns a plugin by ID.
func (m *Manager) GetPlugin(id string) (*Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[id]
	if !exists {
		return nil, fmt.Errorf("%w: %s", integration.ErrPluginNotLoaded, id)
	}

	return plugin, nil
}

// ListPlugins returns all loaded plugins.
func (m *Manager) ListPlugins() []*Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]*Plugin, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// DiscoverPlugins discovers plugins in configured directories.
func (m *Manager) DiscoverPlugins(ctx context.Context) ([]*Plugin, error) {
	var plugins []*Plugin

	for _, dir := range m.pluginDirs {
		discovered, err := m.discoverInDirectory(ctx, dir)
		if err != nil {
			m.logger.Warn("plugin discovery error",
				"directory", dir,
				"error", err,
			)
			continue
		}
		plugins = append(plugins, discovered...)
	}

	return plugins, nil
}

// discoverInDirectory discovers plugins in a single directory.
func (m *Manager) discoverInDirectory(ctx context.Context, dir string) ([]*Plugin, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var plugins []*Plugin
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginPath := filepath.Join(dir, entry.Name())
		manifestPath := filepath.Join(pluginPath, "manifest.json")

		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue
		}

		plugin, err := m.LoadPlugin(ctx, pluginPath)
		if err != nil {
			m.logger.Warn("failed to load discovered plugin",
				"path", pluginPath,
				"error", err,
			)
			if plugin != nil {
				plugin.State = StateError
				plugin.Error = err.Error()
				plugins = append(plugins, plugin)
			}
			continue
		}

		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

// readManifest reads the plugin manifest.
func (m *Manager) readManifest(path string) (*PluginManifest, error) {
	manifestPath := filepath.Join(path, "manifest.json")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("reading manifest file: %w", err)
	}

	var manifest PluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}

	return &manifest, nil
}

// validatePlugin validates a plugin.
func (m *Manager) validatePlugin(plugin *Plugin, manifest *PluginManifest, path string) error {
	for _, validator := range m.validators {
		if err := validator.Validate(plugin, manifest, path); err != nil {
			return err
		}
	}
	return nil
}

// calculateHash calculates the SHA256 hash of the plugin directory.
func (m *Manager) calculateHash(path string) (string, error) {
	manifestPath := filepath.Join(path, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// createIntegrationFromManifest creates an integration from a manifest.
func (m *Manager) createIntegrationFromManifest(manifest *PluginManifest) (integration.Integration, error) {
	// For now, create a generic plugin integration
	// In a real implementation, this would load the actual plugin code
	return &pluginIntegration{
		base:     integration.NewBaseIntegration(manifest.Name, integration.TypePlugin),
		manifest: manifest,
	}, nil
}

// AddValidator adds a custom validator.
func (m *Manager) AddValidator(v Validator) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.validators = append(m.validators, v)
}

// AddAllowedHash adds a hash to the allowlist.
func (m *Manager) AddAllowedHash(hash string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.allowedHashes[hash] = true
}

// Shutdown shuts down all loaded plugins.
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for id, plugin := range m.plugins {
		if plugin.integration != nil {
			if lifecycle, ok := plugin.integration.(integration.LifecycleAware); ok {
				if err := lifecycle.Shutdown(ctx); err != nil {
					m.logger.Error("plugin shutdown error",
						"id", id,
						"error", err,
					)
					lastErr = err
				}
			}
		}
		plugin.State = StateUnloaded
	}

	return lastErr
}

// generatePluginID generates a unique plugin ID.
func generatePluginID(name, version string) string {
	data := fmt.Sprintf("%s:%s", name, version)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

// pluginIntegration is a generic integration wrapper for plugins.
type pluginIntegration struct {
	base     *integration.BaseIntegration
	manifest *PluginManifest
}

func (p *pluginIntegration) Name() string {
	return p.base.Name()
}

func (p *pluginIntegration) Type() integration.IntegrationType {
	return p.base.Type()
}

func (p *pluginIntegration) Execute(ctx context.Context, config *integration.Config, params integration.JSONMap) (*integration.Result, error) {
	return nil, errors.New("plugin execution not implemented")
}

func (p *pluginIntegration) Validate(config *integration.Config) error {
	return p.base.ValidateConfig(config)
}

func (p *pluginIntegration) GetSchema() *integration.Schema {
	if p.manifest.Schema != nil {
		return p.manifest.Schema
	}
	return p.base.GetSchema()
}

func (p *pluginIntegration) GetMetadata() *integration.Metadata {
	return &integration.Metadata{
		Name:        p.manifest.Name,
		Version:     p.manifest.Version,
		Description: p.manifest.Description,
		Author:      p.manifest.Author,
		Homepage:    p.manifest.Homepage,
		License:     p.manifest.License,
		Permissions: p.manifest.Permissions,
		Schema:      p.manifest.Schema,
	}
}
