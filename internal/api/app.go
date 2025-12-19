package api

import (
	"context"
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
	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor"
	"github.com/gorax/gorax/internal/quota"
	"github.com/gorax/gorax/internal/schedule"
	"github.com/gorax/gorax/internal/tenant"
	"github.com/gorax/gorax/internal/webhook"
	"github.com/gorax/gorax/internal/websocket"
	"github.com/gorax/gorax/internal/workflow"
)

// App holds application dependencies
type App struct {
	config   *config.Config
	logger   *slog.Logger
	db       *sqlx.DB
	redis    *redis.Client
	router   *chi.Mux

	// Services
	tenantService     *tenant.Service
	workflowService   *workflow.Service
	webhookService    *webhook.Service
	scheduleService   *schedule.Service
	credentialService credential.Service

	// WebSocket
	wsHub *websocket.Hub

	// Handlers
	healthHandler       *handlers.HealthHandler
	workflowHandler     *handlers.WorkflowHandler
	webhookHandler      *handlers.WebhookHandler
	websocketHandler    *handlers.WebSocketHandler
	tenantAdminHandler  *handlers.TenantAdminHandler
	scheduleHandler     *handlers.ScheduleHandler
	executionHandler    *handlers.ExecutionHandler
	usageHandler        *handlers.UsageHandler
	credentialHandler   *handlers.CredentialHandler

	// Middleware
	quotaChecker *apiMiddleware.QuotaChecker

	// Quota tracking
	quotaTracker *quota.Tracker
}

// NewApp creates a new application instance
func NewApp(cfg *config.Config, logger *slog.Logger) (*App, error) {
	app := &App{
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
	workflowRepo := workflow.NewRepository(db)
	webhookRepo := webhook.NewRepository(db)
	scheduleRepo := schedule.NewRepository(db)

	// Initialize services
	app.tenantService = tenant.NewService(tenantRepo, logger)
	app.workflowService = workflow.NewService(workflowRepo, logger)
	app.webhookService = webhook.NewService(webhookRepo, logger)
	app.scheduleService = schedule.NewService(scheduleRepo, logger)

	// Initialize WebSocket hub
	app.wsHub = websocket.NewHub(logger)
	go app.wsHub.Run() // Start hub in background

	// Initialize executor with WebSocket broadcaster
	broadcaster := websocket.NewHubBroadcaster(app.wsHub)
	workflowExecutor := executor.NewWithBroadcaster(workflowRepo, logger, broadcaster)

	// Create workflow getter adapter for schedule service
	workflowGetter := &workflowServiceAdapter{workflowService: app.workflowService}

	// Wire up dependencies to avoid import cycles
	app.workflowService.SetExecutor(workflowExecutor)
	app.workflowService.SetWebhookService(app.webhookService)
	app.scheduleService.SetWorkflowService(workflowGetter)

	// Initialize handlers
	app.healthHandler = handlers.NewHealthHandler(db, app.redis)
	app.workflowHandler = handlers.NewWorkflowHandler(app.workflowService, logger)
	app.webhookHandler = handlers.NewWebhookHandler(app.workflowService, app.webhookService, logger)
	app.websocketHandler = handlers.NewWebSocketHandler(app.wsHub, logger)
	app.tenantAdminHandler = handlers.NewTenantAdminHandler(app.tenantService, logger)
	app.scheduleHandler = handlers.NewScheduleHandler(app.scheduleService, logger)
	app.executionHandler = handlers.NewExecutionHandler(app.workflowService, logger)

	// TODO: Initialize credential service and handler once service implementation is complete
	// credentialRepo := credential.NewRepository(db)
	// app.credentialService = credential.NewService(credentialRepo, kmsClient, logger)
	// app.credentialHandler = handlers.NewCredentialHandler(app.credentialService, logger)

	// Initialize quota tracker
	app.quotaTracker = quota.NewTracker(app.redis)

	// Initialize usage service and handler
	usageService := handlers.NewUsageService(app.quotaTracker, app.tenantService, logger)
	app.usageHandler = handlers.NewUsageHandler(usageService)

	// Initialize middleware
	app.quotaChecker = apiMiddleware.NewQuotaChecker(app.tenantService, app.redis, logger)

	// Setup router
	app.setupRouter()

	return app, nil
}

// Router returns the HTTP router
func (a *App) Router() http.Handler {
	return a.router
}

// Close cleans up application resources
func (a *App) Close() error {
	if a.db != nil {
		a.db.Close()
	}
	if a.redis != nil {
		a.redis.Close()
	}
	return nil
}

func (a *App) setupRouter() {
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
		// Authentication middleware
		if a.config.Server.Env == "development" {
			// Use development auth that bypasses Kratos
			r.Use(apiMiddleware.DevAuth())
		} else {
			// Use production Kratos auth
			r.Use(apiMiddleware.KratosAuth(a.config.Kratos))
		}

		// Admin routes (no tenant context, no quotas)
		r.Route("/admin", func(r chi.Router) {
			// TODO: Add admin role check middleware
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

		// Tenant context middleware (for non-admin routes)
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

				// Schedule routes for a specific workflow
				r.Route("/{workflowID}/schedules", func(r chi.Router) {
					r.Get("/", a.scheduleHandler.List)
					r.Post("/", a.scheduleHandler.Create)
				})
			})

			// Execution routes
			r.Route("/executions", func(r chi.Router) {
				r.Get("/", a.executionHandler.ListExecutionsAdvanced)
				r.Get("/stats", a.executionHandler.GetExecutionStats)
				r.Get("/{executionID}", a.workflowHandler.GetExecution)
				r.Get("/{executionID}/steps", a.executionHandler.GetExecutionWithSteps)
			})

			// Usage routes
			r.Route("/tenants/{id}/usage", func(r chi.Router) {
				r.Get("/", a.usageHandler.GetCurrentUsage)
				r.Get("/history", a.usageHandler.GetUsageHistory)
			})

			// Schedule routes (all schedules across workflows)
			r.Route("/schedules", func(r chi.Router) {
				r.Get("/", a.scheduleHandler.ListAll)
				r.Get("/{scheduleID}", a.scheduleHandler.Get)
				r.Put("/{scheduleID}", a.scheduleHandler.Update)
				r.Delete("/{scheduleID}", a.scheduleHandler.Delete)
				r.Post("/parse-cron", a.scheduleHandler.ParseCron)
			})

			// WebSocket routes
			r.Route("/ws", func(r chi.Router) {
				r.Get("/", a.websocketHandler.HandleConnection)
				r.Get("/executions/{executionID}", a.websocketHandler.HandleExecutionConnection)
				r.Get("/workflows/{workflowID}", a.websocketHandler.HandleWorkflowConnection)
			})

			// Credential routes
			// TODO: Uncomment once credential service is implemented
			// r.Route("/credentials", func(r chi.Router) {
			// 	r.Get("/", a.credentialHandler.List)
			// 	r.Post("/", a.credentialHandler.Create)
			// 	r.Get("/{credentialID}", a.credentialHandler.Get)
			// 	r.Get("/{credentialID}/value", a.credentialHandler.GetValue) // Sensitive endpoint
			// 	r.Put("/{credentialID}", a.credentialHandler.Update)
			// 	r.Delete("/{credentialID}", a.credentialHandler.Delete)
			// 	r.Post("/{credentialID}/rotate", a.credentialHandler.Rotate)
			// 	r.Get("/{credentialID}/versions", a.credentialHandler.ListVersions)
			// 	r.Get("/{credentialID}/access-log", a.credentialHandler.GetAccessLog)
			// })
		})
	})

	// Webhook endpoint (public, uses webhook-specific auth)
	r.Route("/webhooks", func(r chi.Router) {
		r.Post("/{workflowID}/{webhookID}", a.webhookHandler.Handle)
	})

	a.router = r
}

// workflowServiceAdapter adapts workflow.Service to schedule.WorkflowGetter interface
type workflowServiceAdapter struct {
	workflowService *workflow.Service
}

func (w *workflowServiceAdapter) GetByID(ctx context.Context, tenantID, id string) (interface{}, error) {
	return w.workflowService.GetByID(ctx, tenantID, id)
}
