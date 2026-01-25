package integration

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/api"
	"github.com/gorax/gorax/internal/config"
)

// TestServer represents a fully configured test server with all dependencies
type TestServer struct {
	App        *api.App
	DB         *sqlx.DB
	Redis      *redis.Client
	HTTPServer *httptest.Server
	Client     *http.Client
	BaseURL    string
	Config     *config.Config
	Logger     *slog.Logger
	ctx        context.Context
	cancel     context.CancelFunc
}

// SetupTestServer initializes a complete test environment with database, Redis, and HTTP server
func SetupTestServer(t *testing.T) *TestServer {
	t.Helper()

	// Create context with timeout for the entire test
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	// Load test configuration
	cfg := loadTestConfig(t)

	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Reduce noise in tests
	}))

	// Setup database connection
	db := setupTestDatabase(t, cfg)

	// Setup Redis connection
	redisClient := setupTestRedis(t, cfg)

	// Create application instance
	app, err := api.NewApp(cfg, logger)
	require.NoError(t, err)

	// Create HTTP test server
	httpServer := httptest.NewServer(app.Router())

	// Create HTTP client with reasonable timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	ts := &TestServer{
		App:        app,
		DB:         db,
		Redis:      redisClient,
		HTTPServer: httpServer,
		Client:     client,
		BaseURL:    httpServer.URL,
		Config:     cfg,
		Logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}

	// Register cleanup
	t.Cleanup(func() {
		ts.Cleanup()
	})

	return ts
}

// Cleanup closes all resources
func (ts *TestServer) Cleanup() {
	if ts.cancel != nil {
		ts.cancel()
	}
	if ts.HTTPServer != nil {
		ts.HTTPServer.Close()
	}
	if ts.App != nil {
		ts.App.Close()
	}
	// Note: DB and Redis are closed by App.Close()
}

// loadTestConfig creates a test configuration
func loadTestConfig(t *testing.T) *config.Config {
	t.Helper()

	// Get database URL from environment or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://gorax:gorax_test@localhost:5432/gorax_test?sslmode=disable"
	}

	// Get Redis URL from environment or use default
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	// Create test master key (32 bytes base64 encoded)
	masterKey := base64.StdEncoding.EncodeToString([]byte("test-master-key-32-bytes-long!!!"))

	cfg := &config.Config{
		Server: config.ServerConfig{
			Address: ":0", // Random port
			Env:     "test",
		},
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "gorax",
			Password:        "gorax_test",
			DBName:          "gorax_test",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 10 * time.Minute,
		},
		Redis: config.RedisConfig{
			Address:  redisURL,
			Password: "",
			DB:       0,
		},
		Kratos: config.KratosConfig{
			PublicURL: "http://localhost:4433",
			AdminURL:  "http://localhost:4434",
		},
		Credential: config.CredentialConfig{
			MasterKey: masterKey,
			UseKMS:    false,
		},
		Queue: config.QueueConfig{
			Enabled: false,
		},
		Cleanup: config.CleanupConfig{
			Enabled: false,
		},
		Retention: config.RetentionConfig{
			Enabled: false,
		},
		Observability: config.ObservabilityConfig{
			MetricsEnabled: false,
			TracingEnabled: false,
			SentryEnabled:  false,
		},
		Notification: config.NotificationConfig{
			EnableEmail: false,
			EnableSlack: false,
			EnableInApp: true,
		},
		CORS: config.CORSConfig{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
			MaxAge:           300,
		},
		SecurityHeader: config.SecurityHeaderConfig{
			EnableHSTS:    false,
			CSPDirectives: "",
			FrameOptions:  "SAMEORIGIN",
		},
		AIBuilder: config.AIBuilderConfig{
			Enabled: false,
		},
		WebSocket: config.WebSocketConfig{
			AllowedOrigins:                []string{"*"},
			MaxMessageSize:                512 * 1024,
			MaxConnectionsPerWorkflow:     50,
			ConnectionsPerTenantPerMinute: 60,
		},
		SSRF: config.SSRFConfig{
			Enabled: true,
		},
		FormulaCache: config.FormulaCacheConfig{
			Enabled: true,
			Size:    100,
		},
		OAuth: config.OAuthConfig{
			BaseURL:               "http://localhost:8080",
			GitHubClientID:        "test-github-client",
			GitHubClientSecret:    "test-github-secret",
			GoogleClientID:        "test-google-client",
			GoogleClientSecret:    "test-google-secret",
			SlackClientID:         "test-slack-client",
			SlackClientSecret:     "test-slack-secret",
			MicrosoftClientID:     "test-microsoft-client",
			MicrosoftClientSecret: "test-microsoft-secret",
		},
		Audit: config.AuditConfig{
			Enabled:           true,
			BufferSize:        10,
			FlushInterval:     1 * time.Second,
			HotRetentionDays:  90,
			WarmRetentionDays: 365,
			ColdRetentionDays: 2555,
			ArchiveEnabled:    false,
			PurgeEnabled:      false,
		},
	}

	return cfg
}

// setupTestDatabase creates and migrates a test database
func setupTestDatabase(t *testing.T, cfg *config.Config) *sqlx.DB {
	t.Helper()

	connStr := cfg.Database.ConnectionString()
	db, err := sqlx.Connect("postgres", connStr)
	require.NoError(t, err, "failed to connect to test database")

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	require.NoError(t, err, "failed to ping test database")

	t.Logf("Connected to test database: %s", cfg.Database.DBName)

	return db
}

// setupTestRedis creates a Redis client for testing
func setupTestRedis(t *testing.T, cfg *config.Config) *redis.Client {
	t.Helper()

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	require.NoError(t, err, "failed to ping Redis")

	t.Logf("Connected to Redis: %s", cfg.Redis.Address)

	// Cleanup: Flush test database
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		client.FlushDB(ctx)
	})

	return client
}

// Helper methods for common test operations

// CreateTestTenant creates a test tenant
func (ts *TestServer) CreateTestTenant(t *testing.T, name string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(ts.ctx, 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO tenants (id, name, slug, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, NOW(), NOW())
		RETURNING id
	`

	var tenantID string
	err := ts.DB.QueryRowContext(ctx, query, name, slugify(name)).Scan(&tenantID)
	require.NoError(t, err, "failed to create test tenant")

	t.Logf("Created test tenant: %s (ID: %s)", name, tenantID)

	return tenantID
}

// CreateTestUser creates a test user
func (ts *TestServer) CreateTestUser(t *testing.T, tenantID, email, role string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(ts.ctx, 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO users (id, tenant_id, email, role, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW())
		RETURNING id
	`

	var userID string
	err := ts.DB.QueryRowContext(ctx, query, tenantID, email, role).Scan(&userID)
	require.NoError(t, err, "failed to create test user")

	t.Logf("Created test user: %s (ID: %s, Role: %s)", email, userID, role)

	return userID
}

// ExecuteSQL executes a SQL statement for test setup
func (ts *TestServer) ExecuteSQL(t *testing.T, query string, args ...any) sql.Result {
	t.Helper()

	ctx, cancel := context.WithTimeout(ts.ctx, 5*time.Second)
	defer cancel()

	result, err := ts.DB.ExecContext(ctx, query, args...)
	require.NoError(t, err, "failed to execute SQL")

	return result
}

// GetFromRedis gets a value from Redis
func (ts *TestServer) GetFromRedis(t *testing.T, key string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(ts.ctx, 5*time.Second)
	defer cancel()

	val, err := ts.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return ""
	}
	require.NoError(t, err, "failed to get from Redis")

	return val
}

// SetInRedis sets a value in Redis
func (ts *TestServer) SetInRedis(t *testing.T, key, value string, expiration time.Duration) {
	t.Helper()

	ctx, cancel := context.WithTimeout(ts.ctx, 5*time.Second)
	defer cancel()

	err := ts.Redis.Set(ctx, key, value, expiration).Err()
	require.NoError(t, err, "failed to set in Redis")
}

// MakeRequest makes an HTTP request to the test server
func (ts *TestServer) MakeRequest(t *testing.T, method, path string, body any, headers map[string]string) *http.Response {
	t.Helper()

	return MakeHTTPRequest(t, ts.Client, ts.BaseURL, method, path, body, headers)
}

// MakeRequestWithBody makes an HTTP request with a raw byte body
func (ts *TestServer) MakeRequestWithBody(t *testing.T, method, path string, body []byte, headers map[string]string) *http.Response {
	t.Helper()

	return MakeRawRequest(t, ts.Client, ts.BaseURL, method, path, body, headers)
}

// Helper function to slugify strings
func slugify(s string) string {
	// Simple slugification for testing
	return fmt.Sprintf("%s-%d", s, time.Now().Unix())
}

// WaitForCondition waits for a condition to be true or times out
func WaitForCondition(t *testing.T, timeout time.Duration, condition func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for {
		if condition() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("timeout waiting for condition")
		}
		time.Sleep(100 * time.Millisecond)
	}
}
