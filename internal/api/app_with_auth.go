package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/gorax/gorax/internal/api/handlers"
	apiMiddleware "github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/config"
	"github.com/gorax/gorax/internal/executor"
	"github.com/gorax/gorax/internal/tenant"
	"github.com/gorax/gorax/internal/user"
	"github.com/gorax/gorax/internal/webhook"
	"github.com/gorax/gorax/internal/websocket"
	"github.com/gorax/gorax/internal/workflow"
)

// AppWithAuth holds application dependencies including authentication
type AppWithAuth struct {
	config *config.Config
	logger *slog.Logger
	db     *sqlx.DB
	redis  *redis.Client
	router *chi.Mux

	// Services
	tenantService   *tenant.Service
	userService     *user.Service
	workflowService *workflow.Service
	webhookService  *webhook.Service

	// WebSocket
	wsHub *websocket.Hub

	// Handlers
	healthHandler      *handlers.HealthHandler
	authHandler        *handlers.AuthHandler
	workflowHandler    *handlers.WorkflowHandler
	webhookHandler     *handlers.WebhookHandler
	websocketHandler   *handlers.WebSocketHandler
	tenantAdminHandler *handlers.TenantAdminHandler

	// Middleware
	quotaChecker *apiMiddleware.QuotaChecker
}

// NewAppWithAuth creates a new application instance with authentication
func NewAppWithAuth(cfg *config.Config, logger *slog.Logger) (*AppWithAuth, error) {
	app := &AppWithAuth{
		config: cfg,
		logger: logger,
	}

	// Initialize database connection
	db, err := sqlx.Connect("postgres", cfg.Database.ConnectionString())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	app.db = db

	// Initialize Redis client
	app.redis = redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Initialize repositories
	tenantRepo := tenant.NewRepository(db)
	userRepo := user.NewRepository(db)
	workflowRepo := workflow.NewRepository(db)
	webhookRepo := webhook.NewRepository(db)

	// Initialize services
	app.tenantService = tenant.NewService(tenantRepo, logger)
	app.userService = user.NewService(userRepo, logger)
	app.workflowService = workflow.NewService(workflowRepo, logger)
	app.webhookService = webhook.NewService(webhookRepo, logger)

	// Initialize WebSocket hub
	app.wsHub = websocket.NewHub(logger)
	go app.wsHub.Run() // Start hub in background

	// Initialize executor with WebSocket broadcaster
	broadcaster := websocket.NewHubBroadcaster(app.wsHub)
	workflowExecutor := executor.NewWithBroadcaster(workflowRepo, logger, broadcaster)

	// Wire up dependencies to avoid import cycles
	app.workflowService.SetExecutor(workflowExecutor)
	app.workflowService.SetWebhookService(app.webhookService)

	// Initialize handlers
	app.healthHandler = handlers.NewHealthHandler(db, app.redis)
	app.authHandler = handlers.NewAuthHandler(app.userService, cfg.Kratos, logger)
	app.workflowHandler = handlers.NewWorkflowHandler(app.workflowService, logger)
	app.webhookHandler = handlers.NewWebhookHandler(app.workflowService, app.webhookService, logger)
	app.websocketHandler = handlers.NewWebSocketHandler(app.wsHub, logger)
	app.tenantAdminHandler = handlers.NewTenantAdminHandler(app.tenantService, logger)

	// Initialize middleware
	app.quotaChecker = apiMiddleware.NewQuotaChecker(app.tenantService, app.redis, logger)

	// Setup router
	app.setupRouter()

	return app, nil
}

// Router returns the HTTP router
func (a *AppWithAuth) Router() http.Handler {
	return a.router
}

// Close cleans up application resources
func (a *AppWithAuth) Close() error {
	if a.db != nil {
		a.db.Close()
	}
	if a.redis != nil {
		a.redis.Close()
	}
	return nil
}

func (a *AppWithAuth) setupRouter() {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(apiMiddleware.StructuredLogger(a.logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:5174", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoints (no auth required)
	r.Get("/health", a.healthHandler.Health)
	r.Get("/ready", a.healthHandler.Ready)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public auth routes (no auth middleware)
		r.Route("/auth", func(r chi.Router) {
			// Registration flow
			r.Get("/register", a.authHandler.InitiateRegistration)
			r.Post("/register", a.authHandler.Register)

			// Login flow
			r.Get("/login", a.authHandler.InitiateLogin)
			r.Post("/login", a.authHandler.Login)

			// Password reset flow
			r.Post("/password-reset/request", a.authHandler.RequestPasswordReset)
			r.Post("/password-reset/confirm", a.authHandler.ConfirmPasswordReset)

			// Email verification flow
			r.Post("/verification/request", a.authHandler.RequestEmailVerification)
			r.Post("/verification/confirm", a.authHandler.ConfirmEmailVerification)

			// Kratos webhook endpoint (secured with webhook secret)
			r.Post("/webhooks/kratos", a.authHandler.KratosWebhook)

			// Protected auth routes
			r.Group(func(r chi.Router) {
				if a.config.Server.Env == "development" {
					r.Use(apiMiddleware.DevAuth())
				} else {
					r.Use(apiMiddleware.KratosAuth(a.config.Kratos))
				}

				r.Post("/logout", a.authHandler.Logout)
				r.Get("/me", a.authHandler.GetCurrentUser)
			})
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			// Authentication middleware
			if a.config.Server.Env == "development" {
				r.Use(apiMiddleware.DevAuth())
			} else {
				r.Use(apiMiddleware.KratosAuth(a.config.Kratos))
			}

			// Admin routes (no tenant context, no quotas)
			r.Route("/admin", func(r chi.Router) {
				r.Route("/tenants", func(r chi.Router) {
					r.Get("/", a.tenantAdminHandler.ListTenants)
					r.Post("/", a.tenantAdminHandler.CreateTenant)
					r.Get("/{tenantID}", a.tenantAdminHandler.GetTenant)
					r.Put("/{tenantID}", a.tenantAdminHandler.UpdateTenant)
					r.Delete("/{tenantID}", a.tenantAdminHandler.DeleteTenant)
					r.Put("/{tenantID}/quotas", a.tenantAdminHandler.UpdateTenantQuotas)
					r.Get("/{tenantID}/usage", a.tenantAdminHandler.GetTenantUsage)
				})
			})

			// Tenant-scoped routes
			r.Group(func(r chi.Router) {
				r.Use(apiMiddleware.TenantContext(a.tenantService))
				r.Use(a.quotaChecker.CheckQuotas())

				// Workflow routes
				r.Route("/workflows", func(r chi.Router) {
					r.Get("/", a.workflowHandler.List)
					r.Post("/", a.workflowHandler.Create)
					r.Get("/{workflowID}", a.workflowHandler.Get)
					r.Put("/{workflowID}", a.workflowHandler.Update)
					r.Delete("/{workflowID}", a.workflowHandler.Delete)
					r.Post("/{workflowID}/execute", a.workflowHandler.Execute)
					r.Post("/{workflowID}/dry-run", a.workflowHandler.DryRun)
				})

				// Execution routes
				r.Route("/executions", func(r chi.Router) {
					r.Get("/", a.workflowHandler.ListExecutions)
					r.Get("/{executionID}", a.workflowHandler.GetExecution)
				})

				// WebSocket routes
				r.Route("/ws", func(r chi.Router) {
					r.Get("/", a.websocketHandler.HandleConnection)
					r.Get("/executions/{executionID}", a.websocketHandler.HandleExecutionConnection)
					r.Get("/workflows/{workflowID}", a.websocketHandler.HandleWorkflowConnection)
				})
			})
		})
	})

	// Webhook endpoint (public, uses webhook-specific auth)
	r.Route("/webhooks", func(r chi.Router) {
		r.Post("/{workflowID}/{webhookID}", a.webhookHandler.Handle)
	})

	a.router = r
}
