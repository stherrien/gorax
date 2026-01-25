package integration

import (
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockIntegration is a test integration implementation.
type mockIntegration struct {
	*BaseIntegration
	executeFunc    func(ctx context.Context, config *Config, params JSONMap) (*Result, error)
	validateFunc   func(config *Config) error
	initCalled     bool
	shutdownCalled bool
}

func newMockIntegration(name string, intType IntegrationType) *mockIntegration {
	return &mockIntegration{
		BaseIntegration: NewBaseIntegration(name, intType),
	}
}

func (m *mockIntegration) Execute(ctx context.Context, config *Config, params JSONMap) (*Result, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, config, params)
	}
	return NewSuccessResult(nil, 0), nil
}

func (m *mockIntegration) Validate(config *Config) error {
	if m.validateFunc != nil {
		return m.validateFunc(config)
	}
	return m.BaseIntegration.ValidateConfig(config)
}

func (m *mockIntegration) Initialize(ctx context.Context) error {
	m.initCalled = true
	return nil
}

func (m *mockIntegration) Shutdown(ctx context.Context) error {
	m.shutdownCalled = true
	return nil
}

func (m *mockIntegration) HealthCheck(ctx context.Context) error {
	return nil
}

func TestRegistry_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		registry := NewRegistry(slog.Default())
		integ := newMockIntegration("test", TypeHTTP)

		err := registry.Register(integ)
		require.NoError(t, err)

		assert.True(t, registry.Has("test"))
		assert.Equal(t, 1, registry.Count())
	})

	t.Run("duplicate registration fails", func(t *testing.T) {
		registry := NewRegistry(slog.Default())
		integ1 := newMockIntegration("test", TypeHTTP)
		integ2 := newMockIntegration("test", TypeWebhook)

		err := registry.Register(integ1)
		require.NoError(t, err)

		err = registry.Register(integ2)
		assert.ErrorIs(t, err, ErrAlreadyRegistered)
	})

	t.Run("nil integration fails", func(t *testing.T) {
		registry := NewRegistry(slog.Default())

		err := registry.Register(nil)
		assert.Error(t, err)
	})

	t.Run("empty name fails", func(t *testing.T) {
		registry := NewRegistry(slog.Default())
		integ := newMockIntegration("", TypeHTTP)

		err := registry.Register(integ)
		assert.Error(t, err)
	})
}

func TestRegistry_Get(t *testing.T) {
	t.Run("get existing integration", func(t *testing.T) {
		registry := NewRegistry(slog.Default())
		integ := newMockIntegration("test", TypeHTTP)
		_ = registry.Register(integ)

		got, err := registry.Get("test")
		require.NoError(t, err)
		assert.Equal(t, "test", got.Name())
	})

	t.Run("get non-existing integration", func(t *testing.T) {
		registry := NewRegistry(slog.Default())

		_, err := registry.Get("nonexistent")
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("get with empty name", func(t *testing.T) {
		registry := NewRegistry(slog.Default())

		_, err := registry.Get("")
		assert.Error(t, err)
	})
}

func TestRegistry_Unregister(t *testing.T) {
	t.Run("unregister existing", func(t *testing.T) {
		registry := NewRegistry(slog.Default())
		integ := newMockIntegration("test", TypeHTTP)
		_ = registry.Register(integ)

		err := registry.Unregister("test")
		require.NoError(t, err)

		assert.False(t, registry.Has("test"))
		assert.True(t, integ.shutdownCalled)
	})

	t.Run("unregister non-existing", func(t *testing.T) {
		registry := NewRegistry(slog.Default())

		err := registry.Unregister("nonexistent")
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry(slog.Default())
	_ = registry.Register(newMockIntegration("http", TypeHTTP))
	_ = registry.Register(newMockIntegration("webhook", TypeWebhook))
	_ = registry.Register(newMockIntegration("api", TypeAPI))

	names := registry.List()
	assert.Len(t, names, 3)
	assert.Contains(t, names, "http")
	assert.Contains(t, names, "webhook")
	assert.Contains(t, names, "api")
}

func TestRegistry_ListByType(t *testing.T) {
	registry := NewRegistry(slog.Default())
	_ = registry.Register(newMockIntegration("http1", TypeHTTP))
	_ = registry.Register(newMockIntegration("http2", TypeHTTP))
	_ = registry.Register(newMockIntegration("webhook1", TypeWebhook))

	httpIntegrations := registry.ListByType(TypeHTTP)
	assert.Len(t, httpIntegrations, 2)

	webhookIntegrations := registry.ListByType(TypeWebhook)
	assert.Len(t, webhookIntegrations, 1)

	apiIntegrations := registry.ListByType(TypeAPI)
	assert.Len(t, apiIntegrations, 0)
}

func TestRegistry_Factory(t *testing.T) {
	t.Run("register and create from factory", func(t *testing.T) {
		registry := NewRegistry(slog.Default())

		factory := func(config *Config) (Integration, error) {
			return newMockIntegration("factory-created", TypeHTTP), nil
		}

		err := registry.RegisterFactory("test-factory", factory)
		require.NoError(t, err)

		integ, err := registry.GetOrCreate("test-factory", nil)
		require.NoError(t, err)
		assert.Equal(t, "factory-created", integ.Name())

		// Second call should return cached instance
		integ2, err := registry.GetOrCreate("test-factory", nil)
		require.NoError(t, err)
		assert.Same(t, integ, integ2)
	})

	t.Run("GetOrCreate returns existing", func(t *testing.T) {
		registry := NewRegistry(slog.Default())
		existing := newMockIntegration("existing", TypeHTTP)
		_ = registry.Register(existing)

		integ, err := registry.GetOrCreate("existing", nil)
		require.NoError(t, err)
		assert.Same(t, existing, integ)
	})

	t.Run("GetOrCreate fails without factory", func(t *testing.T) {
		registry := NewRegistry(slog.Default())

		_, err := registry.GetOrCreate("nonexistent", nil)
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestRegistry_Execute(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		registry := NewRegistry(slog.Default())
		integ := newMockIntegration("test", TypeHTTP)
		integ.executeFunc = func(ctx context.Context, config *Config, params JSONMap) (*Result, error) {
			return NewSuccessResult(JSONMap{"result": "ok"}, 10), nil
		}
		_ = registry.Register(integ)

		config := &Config{Name: "test-config", Type: TypeHTTP}
		result, err := registry.Execute(context.Background(), "test", config, nil)

		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("execute non-existing integration", func(t *testing.T) {
		registry := NewRegistry(slog.Default())

		_, err := registry.Execute(context.Background(), "nonexistent", nil, nil)
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestRegistry_Metadata(t *testing.T) {
	registry := NewRegistry(slog.Default())
	integ := newMockIntegration("test", TypeHTTP)
	integ.SetMetadata(&Metadata{
		Name:        "test",
		DisplayName: "Test Integration",
		Version:     "1.0.0",
	})
	_ = registry.Register(integ)

	t.Run("GetMetadata", func(t *testing.T) {
		metadata, err := registry.GetMetadata("test")
		require.NoError(t, err)
		assert.Equal(t, "Test Integration", metadata.DisplayName)
	})

	t.Run("GetMetadata non-existing", func(t *testing.T) {
		_, err := registry.GetMetadata("nonexistent")
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("ListMetadata", func(t *testing.T) {
		metadata := registry.ListMetadata()
		assert.Len(t, metadata, 1)
	})
}

func TestRegistry_Lifecycle(t *testing.T) {
	t.Run("Initialize all", func(t *testing.T) {
		registry := NewRegistry(slog.Default())
		integ1 := newMockIntegration("test1", TypeHTTP)
		integ2 := newMockIntegration("test2", TypeHTTP)
		_ = registry.Register(integ1)
		_ = registry.Register(integ2)

		err := registry.Initialize(context.Background())
		require.NoError(t, err)

		assert.True(t, integ1.initCalled)
		assert.True(t, integ2.initCalled)
	})

	t.Run("Shutdown all", func(t *testing.T) {
		registry := NewRegistry(slog.Default())
		integ1 := newMockIntegration("test1", TypeHTTP)
		integ2 := newMockIntegration("test2", TypeHTTP)
		_ = registry.Register(integ1)
		_ = registry.Register(integ2)

		err := registry.Shutdown(context.Background())
		require.NoError(t, err)

		assert.True(t, integ1.shutdownCalled)
		assert.True(t, integ2.shutdownCalled)
	})
}

func TestRegistry_HealthCheck(t *testing.T) {
	registry := NewRegistry(slog.Default())
	integ := newMockIntegration("test", TypeHTTP)
	_ = registry.Register(integ)

	results := registry.HealthCheck(context.Background())
	assert.Len(t, results, 1)
	assert.NoError(t, results["test"])
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry(slog.Default())

	// Register some initial integrations
	for i := 0; i < 10; i++ {
		name := "initial-" + string(rune('a'+i))
		_ = registry.Register(newMockIntegration(name, TypeHTTP))
	}

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent reads and writes
	for i := 0; i < 10; i++ {
		wg.Add(3)

		// Reader
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = registry.List()
				_ = registry.Count()
				_, _ = registry.Get("initial-a")
			}
		}()

		// Writer (register)
		go func(idx int) {
			defer wg.Done()
			name := "concurrent-" + string(rune('0'+idx))
			if err := registry.Register(newMockIntegration(name, TypeHTTP)); err != nil {
				// Duplicate registration errors are expected
				if err != ErrAlreadyRegistered {
					errors <- err
				}
			}
		}(i)

		// Execute
		go func() {
			defer wg.Done()
			config := &Config{Name: "test", Type: TypeHTTP}
			for j := 0; j < 10; j++ {
				_, _ = registry.Execute(context.Background(), "initial-a", config, nil)
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("unexpected error during concurrent access: %v", err)
	}
}

func TestGlobalRegistry(t *testing.T) {
	// Test that GlobalRegistry returns the same instance
	r1 := GlobalRegistry()
	r2 := GlobalRegistry()
	assert.Same(t, r1, r2)
}
