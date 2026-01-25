package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/gorax/gorax/internal/aibuilder"
	"github.com/gorax/gorax/internal/analytics"
	"github.com/gorax/gorax/internal/api/handlers"
	apiMiddleware "github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/audit"
	"github.com/gorax/gorax/internal/collaboration"
	"github.com/gorax/gorax/internal/config"
	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/errortracking"
	"github.com/gorax/gorax/internal/eventtypes"
	"github.com/gorax/gorax/internal/executor"
	"github.com/gorax/gorax/internal/llm"
	"github.com/gorax/gorax/internal/llm/providers/anthropic"
	"github.com/gorax/gorax/internal/llm/providers/bedrock"
	"github.com/gorax/gorax/internal/llm/providers/openai"
	"github.com/gorax/gorax/internal/marketplace"
	"github.com/gorax/gorax/internal/metrics"
	"github.com/gorax/gorax/internal/oauth"
	oauthProviders "github.com/gorax/gorax/internal/oauth/providers"
	"github.com/gorax/gorax/internal/quota"
	"github.com/gorax/gorax/internal/schedule"
	"github.com/gorax/gorax/internal/sso"
	"github.com/gorax/gorax/internal/suggestions"
	"github.com/gorax/gorax/internal/template"
	"github.com/gorax/gorax/internal/tenant"
	"github.com/gorax/gorax/internal/tracing"
	"github.com/gorax/gorax/internal/webhook"
	"github.com/gorax/gorax/internal/websocket"
	"github.com/gorax/gorax/internal/workflow"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	goraxGraphQL "github.com/gorax/gorax/internal/graphql"
	"github.com/gorax/gorax/internal/graphql/generated"
)

var llmProvidersOnce sync.Once

// registerLLMProviders registers all LLM providers with the global registry.
// This is called once on application startup.
func registerLLMProviders() {
	llmProvidersOnce.Do(func() {
		// Register OpenAI provider
		_ = llm.RegisterProvider(llm.ProviderOpenAI, func(cfg *llm.ProviderConfig) (llm.Provider, error) {
			return openai.NewClient(cfg)
		})

		// Register Anthropic provider
		_ = llm.RegisterProvider(llm.ProviderAnthropic, func(cfg *llm.ProviderConfig) (llm.Provider, error) {
			return anthropic.NewClient(cfg)
		})

		// Register AWS Bedrock provider
		_ = llm.RegisterProvider(llm.ProviderBedrock, func(cfg *llm.ProviderConfig) (llm.Provider, error) {
			return bedrock.NewClient(cfg)
		})
	})
}

// App holds application dependencies
type App struct {
	config *config.Config
	logger *slog.Logger
	db     *sqlx.DB
	redis  *redis.Client
	router *chi.Mux

	// Error tracking
	errorTracker *errortracking.Tracker

	// Metrics
	metrics          *metrics.Metrics
	metricsRegistry  *prometheus.Registry
	dbStatsCollector *metrics.DBStatsCollector
	metricsStopCtx   context.Context
	metricsStopFunc  context.CancelFunc

	// Services
	tenantService       *tenant.Service
	workflowService     *workflow.Service
	workflowBulkService *workflow.BulkService
	webhookService      *webhook.Service
	scheduleService     *schedule.Service
	eventTypeService    *eventtypes.Service
	credentialService   credential.Service
	templateService     *template.Service
	marketplaceService  *marketplace.Service
	collabService       *collaboration.Service
	oauthService        *oauth.Service
	ssoService          *sso.Service
	auditService        *audit.Service

	// WebSocket
	wsHub     *websocket.Hub
	collabHub *collaboration.Hub

	// Handlers
	healthHandler            *handlers.HealthHandler
	workflowHandler          *handlers.WorkflowHandler
	workflowBulkHandler      *handlers.WorkflowBulkHandler
	webhookHandler           *handlers.WebhookHandler
	webhookManagementHandler *handlers.WebhookManagementHandler
	webhookReplayHandler     *handlers.WebhookReplayHandler
	webhookFilterHandler     *handlers.WebhookFilterHandler
	websocketHandler         *handlers.WebSocketHandler
	tenantAdminHandler       *handlers.TenantAdminHandler
	tenantHandler            *handlers.TenantHandler
	scheduleHandler          *handlers.ScheduleHandler
	executionHandler         *handlers.ExecutionHandler
	usageHandler             *handlers.UsageHandler
	credentialHandler        *handlers.CredentialHandler
	metricsHandler           *handlers.MetricsHandler
	eventTypesHandler        *handlers.EventTypesHandler
	suggestionsHandler       *handlers.SuggestionsHandler
	aiBuilderHandler         *handlers.AIBuilderHandler
	marketplaceHandler       *handlers.MarketplaceHandler
	analyticsHandler         *handlers.AnalyticsHandler
	collaborationHandler     *handlers.CollaborationHandler
	oauthHandler             *handlers.OAuthHandler
	ssoHandler               *handlers.SSOHandler
	auditHandler             *handlers.AuditHandler

	// Middleware
	quotaChecker *apiMiddleware.QuotaChecker

	// Quota tracking
	quotaTracker *quota.Tracker
}

// NewApp creates a new application instance
func NewApp(cfg *config.Config, logger *slog.Logger) (*App, error) {
	// Register LLM providers once at startup
	registerLLMProviders()

	app := &App{
		config: cfg,
		logger: logger,
	}

	// Initialize database connection
	db, err := sqlx.Connect("postgres", cfg.Database.ConnectionString())
	if err != nil {
		return nil, err
	}

	// Configure connection pool for optimal performance
	// See docs/POST_DEPLOYMENT_CHECKLIST.md for calculation details
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)
	app.db = db

	// Initialize metrics and Prometheus registry
	app.metrics = metrics.NewMetrics()
	app.metricsRegistry = prometheus.NewRegistry()
	if err := app.metrics.Register(app.metricsRegistry); err != nil {
		return nil, fmt.Errorf("failed to register metrics: %w", err)
	}
	logger.Info("Metrics initialized")

	// Initialize and start DB stats collector
	app.metricsStopCtx, app.metricsStopFunc = context.WithCancel(context.Background())
	app.dbStatsCollector = metrics.NewDBStatsCollector(app.metrics, db.DB, "main", logger)
	go app.dbStatsCollector.Start(app.metricsStopCtx, 15*time.Second)
	logger.Info("DB stats collector started", "interval", "15s")

	// Initialize Redis client
	app.redis = redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Initialize error tracking (Sentry)
	errorTracker, err := errortracking.Initialize(cfg.Observability)
	if err != nil {
		logger.Warn("failed to initialize Sentry", "error", err)
		// Continue without error tracking rather than failing
	}
	app.errorTracker = errorTracker

	// Initialize repositories
	tenantRepo := tenant.NewRepository(db)
	workflowRepo := workflow.NewRepository(db)
	webhookRepo := webhook.NewRepository(db)
	scheduleRepo := schedule.NewRepository(db)
	eventTypeRepo := eventtypes.NewRepository(db)
	templateRepo := template.NewRepository(db)
	marketplaceRepo := marketplace.NewRepository(db)
	categoryRepo := marketplace.NewCategoryRepository(db)

	// Initialize services
	app.tenantService = tenant.NewService(tenantRepo, logger)
	app.workflowService = workflow.NewService(workflowRepo, logger)
	app.webhookService = webhook.NewService(webhookRepo, logger)
	app.workflowBulkService = workflow.NewBulkService(workflowRepo, app.webhookService, logger)
	app.scheduleService = schedule.NewService(scheduleRepo, logger)
	app.eventTypeService = eventtypes.NewService(eventTypeRepo, logger)
	app.templateService = template.NewService(templateRepo, logger)

	// Initialize marketplace service with workflow service adapter
	workflowServiceForMarketplace := &workflowServiceMarketplaceAdapter{workflowService: app.workflowService}
	app.marketplaceService = marketplace.NewService(marketplaceRepo, workflowServiceForMarketplace, logger)

	// Initialize category service
	categoryService := marketplace.NewCategoryService(categoryRepo)

	// Initialize WebSocket hub
	app.wsHub = websocket.NewHub(logger)
	go app.wsHub.Run() // Start hub in background

	// Initialize collaboration service and hub
	app.collabService = collaboration.NewService()
	app.collabHub = collaboration.NewHub(app.collabService, app.wsHub, logger)

	// Initialize executor with WebSocket broadcaster and metrics
	broadcaster := websocket.NewHubBroadcaster(app.wsHub)
	workflowExecutor := executor.NewWithBroadcaster(workflowRepo, logger, broadcaster)
	workflowExecutor.SetMetrics(app.metrics)

	// Create workflow getter adapter for schedule service
	workflowGetter := &workflowServiceAdapter{workflowService: app.workflowService}

	// Wire up dependencies to avoid import cycles
	app.workflowService.SetExecutor(workflowExecutor)
	app.workflowService.SetWebhookService(app.webhookService)
	app.scheduleService.SetWorkflowService(workflowGetter)

	// Initialize handlers
	app.healthHandler = handlers.NewHealthHandler(db, app.redis)
	app.workflowHandler = handlers.NewWorkflowHandler(app.workflowService, logger)
	app.workflowBulkHandler = handlers.NewWorkflowBulkHandler(app.workflowBulkService, logger)
	app.webhookHandler = handlers.NewWebhookHandler(app.workflowService, app.webhookService, logger)
	app.webhookManagementHandler = handlers.NewWebhookManagementHandler(app.webhookService, logger)

	// Initialize replay service and handler
	workflowExecutorForReplay := &workflowExecutorAdapter{workflowService: app.workflowService}
	replayService := webhook.NewReplayService(webhookRepo, workflowExecutorForReplay, logger)
	app.webhookReplayHandler = handlers.NewWebhookReplayHandler(replayService, logger)

	// Initialize filter handler
	app.webhookFilterHandler = handlers.NewWebhookFilterHandler(app.webhookService, logger)

	app.websocketHandler = handlers.NewWebSocketHandler(app.wsHub, logger)
	app.tenantAdminHandler = handlers.NewTenantAdminHandler(app.tenantService, logger)
	app.tenantHandler = handlers.NewTenantHandler(app.tenantService, logger)
	app.scheduleHandler = handlers.NewScheduleHandler(app.scheduleService, logger)
	app.executionHandler = handlers.NewExecutionHandler(app.workflowService, logger)
	app.metricsHandler = handlers.NewMetricsHandler(workflowRepo)
	app.eventTypesHandler = handlers.NewEventTypesHandler(app.eventTypeService, logger)

	// Initialize credential service
	credentialRepo := credential.NewRepository(db)

	// Create encryption service (KMS for production, SimpleEncryption for dev)
	var encryptionService credential.EncryptionServiceInterface
	if cfg.Credential.UseKMS {
		// Production: Use AWS KMS for envelope encryption
		if cfg.Credential.KMSKeyID == "" {
			return nil, fmt.Errorf("CREDENTIAL_KMS_KEY_ID is required when USE_KMS is true")
		}

		// Load AWS config with region override if KMSRegion is set
		awsCfg, err := awsConfig.LoadDefaultConfig(context.Background(), awsConfig.WithRegion(cfg.Credential.KMSRegion))
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config for KMS: %w", err)
		}

		// Create KMS client
		kmsClient := kms.NewFromConfig(awsCfg)

		// Create KMS encryption service
		kmsEncryptionService, err := credential.NewKMSEncryptionService(kmsClient, cfg.Credential.KMSKeyID)
		if err != nil {
			return nil, fmt.Errorf("failed to create KMS encryption service: %w", err)
		}

		encryptionService = credential.NewKMSEncryptionAdapter(kmsEncryptionService)
		logger.Info("Credential encryption initialized", "mode", "KMS", "key_id", cfg.Credential.KMSKeyID, "region", cfg.Credential.KMSRegion)
	} else {
		// Development: Use simple encryption with master key
		masterKey, err := base64.StdEncoding.DecodeString(cfg.Credential.MasterKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode credential master key: %w", err)
		}

		simpleEncryption, err := credential.NewSimpleEncryptionService(masterKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create simple encryption service: %w", err)
		}

		encryptionService = credential.NewSimpleEncryptionAdapter(simpleEncryption)
		logger.Warn("Credential encryption initialized", "mode", "simple", "warning", "Use KMS in production")
	}

	app.credentialService = credential.NewServiceImpl(credentialRepo, encryptionService, logger)
	app.credentialHandler = handlers.NewCredentialHandler(app.credentialService, logger)

	// Initialize quota tracker
	app.quotaTracker = quota.NewTracker(app.redis)

	// Initialize usage service and handler
	usageService := handlers.NewUsageService(app.quotaTracker, app.tenantService, logger)
	app.usageHandler = handlers.NewUsageHandler(usageService)

	// Initialize LLM provider (shared by suggestions and AI builder)
	var llmProvider llm.Provider
	if cfg.AIBuilder.Enabled && cfg.AIBuilder.APIKey != "" {
		llmConfig := &llm.ProviderConfig{
			APIKey: cfg.AIBuilder.APIKey,
		}

		// For AWS Bedrock, use AWS credentials from main AWS config
		if cfg.AIBuilder.Provider == "bedrock" {
			llmConfig.Region = cfg.AWS.Region
			llmConfig.AWSAccessKeyID = cfg.AWS.AccessKeyID
			llmConfig.AWSSecretAccessKey = cfg.AWS.SecretAccessKey
		}

		var err error
		llmProvider, err = llm.GlobalProviderRegistry.GetProvider(cfg.AIBuilder.Provider, llmConfig)
		if err != nil {
			logger.Warn("Failed to initialize LLM provider", "error", err, "provider", cfg.AIBuilder.Provider)
		} else {
			logger.Info("LLM provider initialized",
				"provider", cfg.AIBuilder.Provider,
				"model", cfg.AIBuilder.Model,
			)
		}
	}

	// Initialize suggestions service and handler
	suggestionsRepo := suggestions.NewPostgresRepository(db)
	patternAnalyzer := suggestions.NewPatternAnalyzer(nil) // Uses default patterns

	var llmAnalyzer suggestions.Analyzer
	useLLMForUnmatched := false
	if llmProvider != nil {
		// Create LLM analyzer for suggestions using the shared provider
		llmAnalyzerConfig := suggestions.LLMAnalyzerConfig{
			Model:     cfg.AIBuilder.Model,
			MaxTokens: 1024, // Suggestions need less tokens
		}
		llmAnalyzer = suggestions.NewLLMAnalyzer(llmProvider, llmAnalyzerConfig)
		useLLMForUnmatched = true
		logger.Info("Smart Suggestions LLM analyzer initialized")
	}

	suggestionsService := suggestions.NewSuggestionService(suggestions.SuggestionServiceConfig{
		Repository:         suggestionsRepo,
		PatternAnalyzer:    patternAnalyzer,
		LLMAnalyzer:        llmAnalyzer,
		UseLLMForUnmatched: useLLMForUnmatched,
		Logger:             logger,
	})
	app.suggestionsHandler = handlers.NewSuggestionsHandler(suggestionsService, logger)

	// Initialize AI builder service and handler
	aibuilderRepo := aibuilder.NewPostgresRepository(db)
	nodeRegistry := aibuilder.NewNodeRegistry()

	var aibuilderGenerator *aibuilder.WorkflowGenerator
	if llmProvider != nil {
		// Create generator with LLM provider
		generatorConfig := &aibuilder.GeneratorConfig{
			Model:       cfg.AIBuilder.Model,
			MaxTokens:   cfg.AIBuilder.MaxTokens,
			Temperature: cfg.AIBuilder.Temperature,
		}
		aibuilderGenerator = aibuilder.NewWorkflowGenerator(llmProvider, nodeRegistry, generatorConfig)
		logger.Info("AI Builder generator initialized",
			"model", cfg.AIBuilder.Model,
		)
	} else {
		logger.Info("AI Builder initialized without LLM provider",
			"enabled", cfg.AIBuilder.Enabled,
			"api_key_set", cfg.AIBuilder.APIKey != "",
			"note", "Set AI_BUILDER_ENABLED=true and AI_BUILDER_API_KEY to enable",
		)
	}

	aibuilderService := aibuilder.NewAIBuilderService(
		aibuilder.NewContextRepositoryAdapter(aibuilderRepo),
		aibuilderGenerator,
		&workflowCreatorAdapter{workflowService: app.workflowService},
	)
	app.aiBuilderHandler = handlers.NewAIBuilderHandler(aibuilderService)

	// Initialize marketplace handler
	app.marketplaceHandler = handlers.NewMarketplaceHandler(app.marketplaceService, categoryService, logger)

	// Initialize collaboration handler
	app.collaborationHandler = handlers.NewCollaborationHandler(app.collabHub, app.wsHub, cfg.WebSocket, logger)

	// Initialize analytics service and handler
	analyticsRepo := analytics.NewRepository(db)
	analyticsService := analytics.NewService(analyticsRepo)
	app.analyticsHandler = handlers.NewAnalyticsHandler(analyticsService, logger)

	// Initialize OAuth service and handler
	oauthRepo := oauth.NewRepository(db)

	// Determine Salesforce environment (sandbox or production)
	salesforceIsSandbox := cfg.OAuth.SalesforceEnvironment == "sandbox"

	oauthProviderRegistry := map[string]oauth.Provider{
		"github":     oauthProviders.NewGitHubProvider(),
		"google":     oauthProviders.NewGoogleProvider(),
		"slack":      oauthProviders.NewSlackProvider(),
		"microsoft":  oauthProviders.NewMicrosoftProvider(),
		"twitter":    oauthProviders.NewTwitterProvider(),
		"linkedin":   oauthProviders.NewLinkedInProvider(),
		"salesforce": oauthProviders.NewSalesforceProvider(salesforceIsSandbox),
		"auth0":      oauthProviders.NewAuth0Provider(cfg.OAuth.Auth0Domain),
	}
	// Create an OAuth encryption adapter from the credential encryption service
	oauthEncryptionAdapter := &oauthEncryptionAdapter{encryptionSvc: encryptionService}
	app.oauthService = oauth.NewService(oauthRepo, oauthEncryptionAdapter, oauthProviderRegistry, cfg.OAuth.BaseURL)
	app.oauthHandler = handlers.NewOAuthHandler(app.oauthService)
	logger.Info("OAuth service initialized", "providers", len(oauthProviderRegistry))

	// Initialize SSO service and handler
	// TODO: SSO service requires refactoring to avoid import cycles
	// For now, initialize with nil to allow compilation
	// ssoRepo := sso.NewRepository(db)
	// ssoFactory := sso.NewProviderFactory()
	// app.ssoService = sso.NewService(ssoRepo, nil, ssoFactory)
	// app.ssoHandler = handlers.NewSSOHandler(app.ssoService)
	logger.Info("SSO service initialization skipped - requires refactoring")

	// Initialize Audit service and handler
	auditRepo := audit.NewRepository(db)
	app.auditService = audit.NewService(auditRepo, cfg.Audit.BufferSize, cfg.Audit.FlushInterval)
	app.auditHandler = handlers.NewAuditHandler(app.auditService, logger)
	logger.Info("Audit service initialized",
		"buffer_size", cfg.Audit.BufferSize,
		"flush_interval", cfg.Audit.FlushInterval,
	)

	// Initialize middleware
	app.quotaChecker = apiMiddleware.NewQuotaChecker(app.tenantService, app.redis, logger)

	// Start collaboration cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			cleaned := app.collabService.CleanupInactiveSessions(30 * time.Minute)
			if cleaned > 0 {
				logger.Info("cleaned up inactive collaboration sessions", "count", cleaned)
			}
		}
	}()

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
	// Stop metrics collection
	if a.metricsStopFunc != nil {
		a.metricsStopFunc()
	}
	if a.dbStatsCollector != nil {
		a.dbStatsCollector.Stop()
	}

	// Close audit service (flush buffered events)
	if a.auditService != nil {
		a.auditService.Close()
	}

	if a.errorTracker != nil {
		a.errorTracker.Close()
	}
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

	// HTTP logging with configured level
	httpLogLevel := parseHTTPLogLevel(a.config.Log.HTTPLogLevel)
	r.Use(apiMiddleware.StructuredLoggerWithConfig(a.logger, apiMiddleware.HTTPLoggerConfig{
		LogLevel: httpLogLevel,
	}))

	// Security headers middleware
	securityHeadersConfig := apiMiddleware.SecurityHeadersConfig{
		EnableHSTS:    a.config.SecurityHeader.EnableHSTS,
		HSTSMaxAge:    a.config.SecurityHeader.HSTSMaxAge,
		CSPDirectives: a.config.SecurityHeader.CSPDirectives,
		FrameOptions:  a.config.SecurityHeader.FrameOptions,
	}
	r.Use(apiMiddleware.SecurityHeaders(securityHeadersConfig))

	// Add distributed tracing middleware if enabled
	if a.config.Observability.TracingEnabled {
		r.Use(tracing.HTTPMiddleware())
	}

	// Add Sentry middleware if error tracking is enabled
	if a.errorTracker != nil {
		r.Use(apiMiddleware.SentryMiddleware(a.errorTracker))
	}

	// Add audit middleware if enabled
	if a.config.Audit.Enabled && a.auditService != nil {
		r.Use(apiMiddleware.AuditMiddleware(a.auditService, a.logger))
	}

	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// CORS middleware with environment-aware validation
	corsMiddleware, err := apiMiddleware.NewCORSMiddleware(a.config.CORS, a.config.Server.Env)
	if err != nil {
		a.logger.Error("failed to create CORS middleware", "error", err)
		// Fall back to restrictive CORS in case of configuration error
	} else {
		r.Use(corsMiddleware)
	}

	// Health check endpoints (no auth required)
	r.Get("/health", a.healthHandler.Health)
	r.Get("/ready", a.healthHandler.Ready)

	// Prometheus metrics endpoint (no auth required for scraping)
	if a.config.Observability.MetricsEnabled {
		r.Handle("/metrics", promhttp.HandlerFor(a.metricsRegistry, promhttp.HandlerOpts{}))
	}

	// Swagger API documentation (no auth required)
	r.Get("/api/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/docs/api/swagger.json"),
	))

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
			// Require admin role for all admin routes
			r.Use(apiMiddleware.RequireAdmin())

			// Tenant switching for admins
			r.Post("/switch-tenant", a.tenantAdminHandler.SwitchTenant)

			r.Route("/tenants", func(r chi.Router) {
				r.Get("/", a.tenantAdminHandler.ListTenants)
				r.Post("/", a.tenantAdminHandler.CreateTenant)
				r.Get("/{tenantID}", a.tenantAdminHandler.GetTenant)
				r.Put("/{tenantID}", a.tenantAdminHandler.UpdateTenant)
				r.Delete("/{tenantID}", a.tenantAdminHandler.DeleteTenant)
				r.Put("/{tenantID}/quotas", a.tenantAdminHandler.UpdateTenantQuotas)
				r.Get("/{tenantID}/usage", a.tenantAdminHandler.GetTenantUsage)
				r.Put("/{tenantID}/status", a.tenantAdminHandler.SetTenantStatus)
				r.Post("/{tenantID}/activate", a.tenantAdminHandler.ActivateTenant)
				r.Post("/{tenantID}/suspend", a.tenantAdminHandler.SuspendTenant)
			})

			// SSO provider management routes (admin only)
			// TODO: Re-enable when SSO service is properly initialized
			/* r.Route("/sso", func(r chi.Router) {
				r.Post("/providers", a.ssoHandler.CreateProvider)
				r.Get("/providers", a.ssoHandler.ListProviders)
				r.Get("/providers/{id}", a.ssoHandler.GetProvider)
				r.Put("/providers/{id}", a.ssoHandler.UpdateProvider)
				r.Delete("/providers/{id}", a.ssoHandler.DeleteProvider)
			}) */

			// Audit log routes (admin only)
			r.Route("/audit", func(r chi.Router) {
				r.Get("/events", a.auditHandler.QueryEvents)
				r.Get("/events/{id}", a.auditHandler.GetEvent)
				r.Get("/stats", a.auditHandler.GetStats)
				r.Post("/export", a.auditHandler.ExportEvents)
			})
		})

		// Tenant context middleware (for non-admin routes)
		r.Group(func(r chi.Router) {
			// Configure tenant middleware with single/multi tenant mode support
			tenantMiddlewareCfg := apiMiddleware.TenantMiddlewareConfig{
				TenantConfig: a.config.Tenant,
			}
			r.Use(apiMiddleware.TenantContextWithConfig(a.tenantService, tenantMiddlewareCfg))
			r.Use(a.quotaChecker.CheckQuotas())

			// Current tenant info routes (available to all authenticated users)
			r.Route("/tenant", func(r chi.Router) {
				r.Get("/info", a.tenantHandler.GetCurrentTenant)
				r.Get("/settings", a.tenantHandler.GetTenantSettings)
				r.Get("/quotas", a.tenantHandler.GetTenantQuotas)
			})

			// Workflow routes
			r.Route("/workflows", func(r chi.Router) {
				r.Get("/", a.workflowHandler.List)
				r.Post("/", a.workflowHandler.Create)
				r.Get("/{workflowID}", a.workflowHandler.Get)
				r.Put("/{workflowID}", a.workflowHandler.Update)
				r.Delete("/{workflowID}", a.workflowHandler.Delete)
				r.Post("/{workflowID}/execute", a.workflowHandler.Execute)
				r.Post("/{workflowID}/dry-run", a.workflowHandler.DryRun)

				// Bulk operations
				r.Route("/bulk", func(r chi.Router) {
					r.Post("/delete", a.workflowBulkHandler.BulkDelete)
					r.Post("/enable", a.workflowBulkHandler.BulkEnable)
					r.Post("/disable", a.workflowBulkHandler.BulkDisable)
					r.Post("/export", a.workflowBulkHandler.BulkExport)
					r.Post("/clone", a.workflowBulkHandler.BulkClone)
				})

				// Version routes for a specific workflow
				r.Route("/{workflowID}/versions", func(r chi.Router) {
					r.Get("/", a.workflowHandler.ListVersions)
					r.Get("/{version}", a.workflowHandler.GetVersion)
					r.Post("/{version}/restore", a.workflowHandler.RestoreVersion)
				})

				// Schedule routes for a specific workflow
				r.Route("/{workflowID}/schedules", func(r chi.Router) {
					r.Get("/", a.scheduleHandler.List)
					r.Post("/", a.scheduleHandler.Create)
				})

				// Collaboration WebSocket route for a specific workflow
				r.Get("/{id}/collaborate", a.collaborationHandler.HandleWorkflowCollaboration)
			})

			// Execution routes
			r.Route("/executions", func(r chi.Router) {
				r.Get("/", a.executionHandler.ListExecutionsAdvanced)
				r.Get("/stats", a.executionHandler.GetExecutionStats)
				r.Get("/{executionID}", a.workflowHandler.GetExecution)
				r.Get("/{executionID}/steps", a.executionHandler.GetExecutionWithSteps)
			})

			// Metrics routes
			r.Route("/metrics", func(r chi.Router) {
				r.Get("/trends", a.metricsHandler.GetExecutionTrends)
				r.Get("/duration", a.metricsHandler.GetDurationStats)
				r.Get("/failures", a.metricsHandler.GetTopFailures)
				r.Get("/trigger-breakdown", a.metricsHandler.GetTriggerBreakdown)
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
				r.Post("/preview", a.scheduleHandler.PreviewSchedule)

				// Execution history routes
				r.Get("/{scheduleID}/executions", a.scheduleHandler.ListExecutionHistory)
				r.Get("/{scheduleID}/executions/{logID}", a.scheduleHandler.GetExecutionLog)
			})

			// Webhook management routes
			r.Route("/webhooks", func(r chi.Router) {
				r.Get("/", a.webhookManagementHandler.List)
				r.Post("/", a.webhookManagementHandler.Create)
				r.Get("/{id}", a.webhookManagementHandler.Get)
				r.Put("/{id}", a.webhookManagementHandler.Update)
				r.Delete("/{id}", a.webhookManagementHandler.Delete)
				r.Post("/{id}/regenerate-secret", a.webhookManagementHandler.RegenerateSecret)
				r.Post("/{id}/test", a.webhookManagementHandler.TestWebhook)
				r.Get("/{id}/events", a.webhookManagementHandler.GetEventHistory)
				r.Post("/{webhookID}/events/replay", a.webhookReplayHandler.BatchReplayEvents)

				// Filter routes
				r.Route("/{id}/filters", func(r chi.Router) {
					r.Get("/", a.webhookFilterHandler.List)
					r.Post("/", a.webhookFilterHandler.Create)
					r.Get("/{filterID}", a.webhookFilterHandler.Get)
					r.Put("/{filterID}", a.webhookFilterHandler.Update)
					r.Delete("/{filterID}", a.webhookFilterHandler.Delete)
					r.Post("/test", a.webhookFilterHandler.Test)
				})
			})

			// Webhook event replay routes
			r.Route("/events", func(r chi.Router) {
				r.Post("/{eventID}/replay", a.webhookReplayHandler.ReplayEvent)
			})

			// Event types registry routes
			r.Route("/event-types", func(r chi.Router) {
				r.Get("/", a.eventTypesHandler.List)
			})

			// WebSocket routes
			r.Route("/ws", func(r chi.Router) {
				r.Get("/", a.websocketHandler.HandleConnection)
				r.Get("/executions/{executionID}", a.websocketHandler.HandleExecutionConnection)
				r.Get("/workflows/{workflowID}", a.websocketHandler.HandleWorkflowConnection)
			})

			// Credential routes
			r.Route("/credentials", func(r chi.Router) {
				r.Get("/", a.credentialHandler.List)
				r.Post("/", a.credentialHandler.Create)
				r.Get("/{credentialID}", a.credentialHandler.Get)
				r.Get("/{credentialID}/value", a.credentialHandler.GetValue) // Sensitive endpoint
				r.Put("/{credentialID}", a.credentialHandler.Update)
				r.Delete("/{credentialID}", a.credentialHandler.Delete)
				r.Post("/{credentialID}/rotate", a.credentialHandler.Rotate)
				r.Get("/{credentialID}/versions", a.credentialHandler.ListVersions)
				r.Get("/{credentialID}/access-log", a.credentialHandler.GetAccessLog)
			})

			// Suggestions routes (Smart error analysis)
			r.Route("/suggestions", func(r chi.Router) {
				r.Get("/{suggestionID}", a.suggestionsHandler.Get)
				r.Post("/{suggestionID}/apply", a.suggestionsHandler.Apply)
				r.Post("/{suggestionID}/dismiss", a.suggestionsHandler.Dismiss)
			})

			// Execution suggestions (nested under executions for context)
			r.Route("/executions/{executionID}/suggestions", func(r chi.Router) {
				r.Get("/", a.suggestionsHandler.List)
				r.Post("/analyze", a.suggestionsHandler.Analyze)
			})

			// AI Workflow Builder routes
			r.Route("/ai/workflows", func(r chi.Router) {
				r.Post("/generate", a.aiBuilderHandler.Generate)
				r.Post("/refine", a.aiBuilderHandler.Refine)
				r.Route("/conversations", func(r chi.Router) {
					r.Get("/", a.aiBuilderHandler.ListConversations)
					r.Get("/{id}", a.aiBuilderHandler.GetConversation)
					r.Post("/{id}/apply", a.aiBuilderHandler.Apply)
					r.Post("/{id}/abandon", a.aiBuilderHandler.Abandon)
				})
			})

			// Marketplace routes
			r.Route("/marketplace", func(r chi.Router) {
				// Public template routes
				r.Get("/templates", a.marketplaceHandler.ListTemplates)
				r.Get("/templates/{id}", a.marketplaceHandler.GetTemplate)
				r.Post("/templates", a.marketplaceHandler.PublishTemplate)
				r.Post("/templates/{id}/install", a.marketplaceHandler.InstallTemplate)
				r.Get("/trending", a.marketplaceHandler.GetTrending)
				r.Get("/popular", a.marketplaceHandler.GetPopular)

				// Category routes
				r.Get("/categories", a.marketplaceHandler.GetCategories)
				r.Get("/categories/{id}", a.marketplaceHandler.GetCategory)

				// Review routes
				r.Get("/templates/{id}/reviews", a.marketplaceHandler.GetReviews)
				r.Get("/templates/{id}/rating-distribution", a.marketplaceHandler.GetRatingDistribution)
				r.Post("/templates/{id}/rate", a.marketplaceHandler.RateTemplate)
				r.Delete("/templates/{id}/reviews/{reviewId}", a.marketplaceHandler.DeleteReview)

				// Review helpful votes
				r.Post("/reviews/{reviewId}/helpful", a.marketplaceHandler.VoteReviewHelpful)
				r.Delete("/reviews/{reviewId}/helpful", a.marketplaceHandler.UnvoteReviewHelpful)

				// Review reporting
				r.Post("/reviews/{reviewId}/report", a.marketplaceHandler.ReportReview)

				// Admin routes (marketplace moderation)
				r.Route("/admin", func(r chi.Router) {
					r.Use(apiMiddleware.RequireAdmin())

					// Category management (admin only)
					r.Post("/categories", a.marketplaceHandler.CreateCategory)

					// Review moderation
					r.Get("/review-reports", a.marketplaceHandler.GetReviewReports)
					r.Put("/review-reports/{reportId}", a.marketplaceHandler.ResolveReviewReport)
					r.Put("/reviews/{reviewId}/hide", a.marketplaceHandler.HideReview)
				})
			})

			// Analytics routes
			r.Route("/analytics", func(r chi.Router) {
				r.Get("/overview", a.analyticsHandler.GetTenantOverview)
				r.Get("/workflows/{workflowID}", a.analyticsHandler.GetWorkflowStats)
				r.Get("/trends", a.analyticsHandler.GetExecutionTrends)
				r.Get("/top-workflows", a.analyticsHandler.GetTopWorkflows)
				r.Get("/errors", a.analyticsHandler.GetErrorBreakdown)
				r.Get("/workflows/{workflowID}/nodes", a.analyticsHandler.GetNodePerformance)
			})

			// OAuth routes
			r.Route("/oauth", func(r chi.Router) {
				r.Get("/providers", a.oauthHandler.ListProviders)
				r.Get("/authorize/{provider}", a.oauthHandler.Authorize)
				r.Get("/callback/{provider}", a.oauthHandler.Callback)
				r.Get("/connections", a.oauthHandler.ListConnections)
				r.Get("/connections/{id}", a.oauthHandler.GetConnection)
				r.Delete("/connections/{id}", a.oauthHandler.RevokeConnection)
				r.Post("/connections/{id}/test", a.oauthHandler.TestConnection)
			})
		})

		// GraphQL API endpoint (with authentication and tenant context)
		r.Group(func(r chi.Router) {
			// Create GraphQL resolver with services
			resolver := &goraxGraphQL.Resolver{
				WorkflowService: a.workflowService,
				WebhookService:  a.webhookService,
				ScheduleService: a.scheduleService,
				TemplateService: a.templateService,
				Logger:          a.logger,
			}

			// Create GraphQL server
			graphqlServer := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
				Resolvers: resolver,
			}))

			// GraphQL endpoint
			r.Handle("/graphql", graphqlServer)

			// GraphQL Playground (only in development)
			if a.config.Server.Env == "development" {
				r.Handle("/graphql/playground", playground.Handler("GraphQL Playground", "/api/v1/graphql"))
			}
		})
	})

	// Webhook endpoint (public, uses webhook-specific auth)
	r.Route("/webhooks", func(r chi.Router) {
		r.Post("/{workflowID}/{webhookID}", a.webhookHandler.Handle)
	})

	// SSO authentication endpoints (public)
	// TODO: Re-enable when SSO service is properly initialized
	/* r.Route("/sso", func(r chi.Router) {
		r.Get("/login/{id}", a.ssoHandler.InitiateLogin)
		r.Get("/callback/{id}", a.ssoHandler.HandleCallback)
		r.Post("/callback/{id}", a.ssoHandler.HandleCallback)
		r.Post("/acs", a.ssoHandler.HandleSAMLAssertion) // SAML Assertion Consumer Service
		r.Get("/metadata/{id}", a.ssoHandler.GetMetadata)
		r.Get("/discover", a.ssoHandler.DiscoverProvider)
	}) */

	a.router = r
}

// workflowServiceAdapter adapts workflow.Service to schedule.WorkflowGetter interface
type workflowServiceAdapter struct {
	workflowService *workflow.Service
}

func (w *workflowServiceAdapter) GetByID(ctx context.Context, tenantID, id string) (interface{}, error) {
	return w.workflowService.GetByID(ctx, tenantID, id)
}

// workflowExecutorAdapter adapts workflow.Service to webhook.WorkflowExecutor interface
type workflowExecutorAdapter struct {
	workflowService *workflow.Service
}

func (w *workflowExecutorAdapter) Execute(ctx context.Context, tenantID, workflowID, triggerType string, triggerData []byte) (string, error) {
	execution, err := w.workflowService.Execute(ctx, tenantID, workflowID, triggerType, triggerData)
	if err != nil {
		return "", err
	}
	return execution.ID, nil
}

// workflowCreatorAdapter adapts workflow.Service to aibuilder.WorkflowCreator interface
type workflowCreatorAdapter struct {
	workflowService *workflow.Service
}

func (w *workflowCreatorAdapter) CreateWorkflow(ctx context.Context, tenantID, userID string, generated *aibuilder.GeneratedWorkflow) (string, error) {
	// Convert generated workflow definition to JSON
	var definitionJSON json.RawMessage
	var err error
	if generated.Definition != nil {
		// Build the definition structure
		def := workflow.WorkflowDefinition{
			Nodes: make([]workflow.Node, len(generated.Definition.Nodes)),
			Edges: make([]workflow.Edge, len(generated.Definition.Edges)),
		}

		// Convert nodes
		for i, gn := range generated.Definition.Nodes {
			var position workflow.Position
			if gn.Position != nil {
				position = workflow.Position{X: gn.Position.X, Y: gn.Position.Y}
			}

			// Marshal config to json.RawMessage
			var configJSON json.RawMessage
			if gn.Config != nil {
				configJSON, err = json.Marshal(gn.Config)
				if err != nil {
					return "", fmt.Errorf("failed to marshal node config: %w", err)
				}
			}

			def.Nodes[i] = workflow.Node{
				ID:       gn.ID,
				Type:     gn.Type,
				Position: position,
				Data: workflow.NodeData{
					Name:   gn.Name,
					Config: configJSON,
				},
			}
		}

		// Convert edges
		if generated.Definition.Edges != nil {
			for i, ge := range generated.Definition.Edges {
				def.Edges[i] = workflow.Edge{
					ID:       ge.ID,
					Source:   ge.Source,
					Target:   ge.Target,
					SourceID: ge.SourceHandle,
					TargetID: ge.TargetHandle,
				}
			}
		}

		definitionJSON, err = json.Marshal(def)
		if err != nil {
			return "", fmt.Errorf("failed to marshal workflow definition: %w", err)
		}
	}

	// Create workflow using the service's CreateWorkflowInput
	input := workflow.CreateWorkflowInput{
		Name:        generated.Name,
		Description: generated.Description,
		Definition:  definitionJSON,
	}

	// Create the workflow
	created, err := w.workflowService.Create(ctx, tenantID, userID, input)
	if err != nil {
		return "", err
	}
	return created.ID, nil
}

// oauthEncryptionAdapter adapts credential.EncryptionServiceInterface to oauth.EncryptionService
type oauthEncryptionAdapter struct {
	encryptionSvc credential.EncryptionServiceInterface
}

func (a *oauthEncryptionAdapter) Encrypt(ctx context.Context, tenantID string, data *credential.CredentialData) (*credential.EncryptedSecret, error) {
	return a.encryptionSvc.Encrypt(ctx, tenantID, data)
}

func (a *oauthEncryptionAdapter) Decrypt(ctx context.Context, encrypted *credential.EncryptedSecret) (*credential.CredentialData, error) {
	// Convert EncryptedSecret to byte array format expected by EncryptionServiceInterface
	// encryptedData format: nonce (12 bytes) + ciphertext + authTag (16 bytes)
	const nonceSize = 12
	encryptedData := make([]byte, 0, nonceSize+len(encrypted.Ciphertext)+len(encrypted.AuthTag))
	encryptedData = append(encryptedData, encrypted.Nonce...)
	encryptedData = append(encryptedData, encrypted.Ciphertext...)
	encryptedData = append(encryptedData, encrypted.AuthTag...)

	// encryptedKey is the encrypted DEK
	encryptedKey := encrypted.EncryptedDEK

	return a.encryptionSvc.Decrypt(ctx, encryptedData, encryptedKey)
}

// workflowServiceMarketplaceAdapter adapts workflow.Service to marketplace.WorkflowService interface
type workflowServiceMarketplaceAdapter struct {
	workflowService *workflow.Service
}

func (w *workflowServiceMarketplaceAdapter) CreateFromTemplate(ctx context.Context, tenantID, userID, templateID, workflowName string, definition json.RawMessage) (string, error) {
	input := workflow.CreateWorkflowInput{
		Name:       workflowName,
		Definition: definition,
	}

	created, err := w.workflowService.Create(ctx, tenantID, userID, input)
	if err != nil {
		return "", err
	}
	return created.ID, nil
}

// parseHTTPLogLevel converts string log level to slog.Level for HTTP access logs
func parseHTTPLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		// Default to debug for HTTP logs to reduce noise
		return slog.LevelDebug
	}
}
