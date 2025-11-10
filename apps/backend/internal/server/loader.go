package server

import (
	"context"
	"fmt"

	"github.com/detect-price-by-photo/backend/internal/adapter"
	"github.com/detect-price-by-photo/backend/internal/ai"
	aiRepo "github.com/detect-price-by-photo/backend/internal/ai/repository"
	aiServices "github.com/detect-price-by-photo/backend/internal/ai/services"
	"github.com/detect-price-by-photo/backend/internal/cache"
	"github.com/detect-price-by-photo/backend/internal/config"
	"github.com/detect-price-by-photo/backend/internal/middleware"
	"github.com/detect-price-by-photo/backend/internal/notification"
	"github.com/detect-price-by-photo/backend/internal/storage"
	"github.com/detect-price-by-photo/backend/internal/photo/services"

	"github.com/labstack/echo/v4"

	modAuth "github.com/detect-price-by-photo/backend/internal/user/auth"
	modUser "github.com/detect-price-by-photo/backend/internal/user/user"
	modSubscription "github.com/detect-price-by-photo/backend/internal/subscription"
	modPhoto "github.com/detect-price-by-photo/backend/internal/photo"
	modPayment "github.com/detect-price-by-photo/backend/internal/payment"
	modAdmin "github.com/detect-price-by-photo/backend/internal/admin"
)

// registerModules registers application modules, injects middleware and attaches routes.
// Keeps Start() concise and centralizes module wiring for easier testing/refactor.
func (s *HTTPServer) registerModules(cfg *config.Config, pg *adapter.PostgresDB, mailer *notification.Mailer, e *echo.Echo) error {
	// Register primary HTTP server routes
	serverHandler := NewServerHandler(pg.Pool, s.logger)
	serverHandler.RegisterRoutes(e)

	// Register global middleware for API
	e.Use(middleware.CORSMiddleware(cfg))
	e.Use(middleware.RateLimitMiddleware(
		cfg.App.RateLimitRequests, cfg.App.RateLimitBurstSize,
	))
	e.Use(middleware.CompressionMiddleware())

	// Create API v1 route group
	apiV1Route := e.Group("/api/v1")

	// Load subscription module first (needed for user auto-enrollment)
	subscriptionModule := modSubscription.NewModule(&modSubscription.Options{
		PgPool: pg.Pool,
		Logger: s.logger,
	})

	// Load user module with subscription service for auto-enrollment
	userModule := modUser.NewModule(&modUser.Options{
		PgPool:            pg.Pool,
		Logger:            s.logger,
		SubscriptionService: subscriptionModule.GetSubscriptionService(),
	})

	// Load auth module (requires user service)
	authModule := modAuth.NewModule(&modAuth.Options{
		PgPool:       pg.Pool,
		Logger:       s.logger,
		UserService:  userModule.GetUserService(),
		JWTSecretKey: []byte(cfg.App.JWTSecretKey),
		BaseURL:      cfg.GetAppBaseURL(),
		Mailer:       mailer,
	})

	// Inject auth middleware into user module so protected user routes use same JWT config
	userModule.Use(authModule.JWTMiddleware())

	// Register the module routes after injecting middleware
	userModule.RegisterRoutes(apiV1Route)
	authModule.RegisterRoutes(apiV1Route)

	// Inject auth middleware for protected routes
	subscriptionModule.Use(authModule.JWTMiddleware())
	subscriptionModule.RegisterRoutes(apiV1Route)

	// Initialize storage client (local filesystem for development)
	storageBaseDir := "./storage/uploads"
	storageBaseURL := fmt.Sprintf("%s/files", cfg.GetAppBaseURL())
	storageClient, err := storage.NewLocalStorageClient(storageBaseDir, storageBaseURL)
	if err != nil {
		return err
	}

	// Initialize Redis cache client (if enabled)
	var cacheClient cache.CacheClientInterface
	if cfg.Redis.Enabled {
		redisDB, err := adapter.NewRedis(adapter.RedisConfig{
			URL:     cfg.Redis.URL,
			Enabled: cfg.Redis.Enabled,
			DB:      cfg.Redis.DB,
		})
		if err != nil {
			s.logger.Warn("Failed to initialize Redis, continuing without cache", "error", err)
		} else {
			cacheClient = cache.NewCacheClient(redisDB.Client)
			s.logger.Info("Redis cache initialized", "url", cfg.Redis.URL)
		}
	}

	// Initialize AI services
	var estimatorService services.EstimatorServiceInterface
	if cfg.AI.OpenRouterAPIKey != "" {
		// Create OpenRouter client
		openRouterClient := ai.NewOpenRouterClient(cfg.AI.OpenRouterAPIKey, s.logger)

		// Create AI model repository
		aiModelRepo := aiRepo.NewAIModelRepository(pg.Pool, s.logger)

		// Create model registry from database (preferred) or fallback to config
		var modelRegistry *ai.ModelRegistry
		var err error
		modelRegistry, err = ai.NewModelRegistryFromDB(aiModelRepo, s.logger)
		if err != nil {
			s.logger.Warn("failed to load models from database, falling back to config", "error", err)
			// Fallback to config-based registry
			modelRegistry, err = ai.NewModelRegistry(cfg.AI.DefaultModel, cfg.AI.ModeRouting, s.logger)
			if err != nil {
				return fmt.Errorf("failed to create model registry: %w", err)
			}
		} else {
			s.logger.Info("AI models loaded from database")
		}

		// Create embedder service
		embedderService := ai.NewEmbedderService(openRouterClient, s.logger)

		// Create retriever service
		retrieverService := ai.NewRetrieverService(pg.Pool, s.logger)

		// Create estimator service
		aiEstimatorService := ai.NewEstimatorService(
			openRouterClient,
			embedderService,
			retrieverService,
			modelRegistry,
			cacheClient,
			s.logger,
		)

		// Create adapter to convert ai.PriceEstimateResult to services.PriceEstimateResult
		estimatorService = &estimatorAdapter{service: aiEstimatorService}
		s.logger.Info("AI services initialized", "default_model", modelRegistry.GetDefaultModel().ID)
	} else {
		s.logger.Warn("OpenRouter API key not configured, price estimation will be disabled")
	}

	// Load photo module
	photoModule := modPhoto.NewModule(&modPhoto.Options{
		PgPool:              pg.Pool,
		Logger:              s.logger,
		StorageClient:       storageClient,
		SubscriptionService: subscriptionModule.GetSubscriptionService(),
		EstimatorService:    estimatorService,
	})
	// Inject auth middleware for protected routes
	photoModule.Use(authModule.JWTMiddleware())
	photoModule.RegisterRoutes(apiV1Route)

	// Initialize Midtrans client (if configured)
	var midtransClient modPayment.MidtransClientInterface
	if cfg.Midtrans.ServerKey != "" {
		midtransClient = modPayment.NewMidtransClient(
			cfg.Midtrans.ServerKey,
			cfg.Midtrans.ClientKey,
			cfg.Midtrans.Env,
			s.logger,
		)
		s.logger.Info("Midtrans client initialized", "env", cfg.Midtrans.Env)
	} else {
		s.logger.Warn("Midtrans credentials not configured, payment features will be disabled")
	}

	// Load payment module
	if midtransClient != nil {
		paymentModule := modPayment.NewModule(&modPayment.Options{
			PgPool:              pg.Pool,
			Logger:              s.logger,
			MidtransClient:      midtransClient,
			SubscriptionService: subscriptionModule.GetSubscriptionService(),
			UserService:         userModule.GetUserService(),
		})
		// Inject auth middleware for protected routes
		paymentModule.Use(authModule.JWTMiddleware())
		paymentModule.RegisterRoutes(apiV1Route)
	}

	// Load admin module
	var adminModule *modAdmin.AdminModule
	if cfg.AI.OpenRouterAPIKey != "" {
		// Initialize AI model service if AI is configured
		aiModelRepo := aiRepo.NewAIModelRepository(pg.Pool, s.logger)
		aiModelService := aiServices.NewModelService(aiModelRepo, s.logger)
		adminModule = modAdmin.NewModule(&modAdmin.Options{
			PgPool:      pg.Pool,
			Logger:      s.logger,
			UserRepo:    userModule.GetUserRepository(),
			PlanRepo:    subscriptionModule.GetPlanRepository(),
			ModelService: aiModelService,
		})
	} else {
		adminModule = modAdmin.NewModule(&modAdmin.Options{
			PgPool:      pg.Pool,
			Logger:      s.logger,
			UserRepo:    userModule.GetUserRepository(),
			PlanRepo:    subscriptionModule.GetPlanRepository(),
			ModelService: nil,
		})
	}
	// Inject auth and admin middleware for admin routes
	adminModule.Use(authModule.JWTMiddleware())
	adminModule.Use(modAdmin.RequireAdmin(userModule.GetUserRepository()))
	adminModule.RegisterRoutes(apiV1Route)

	return nil
}

// estimatorAdapter adapts ai.EstimatorService to services.EstimatorServiceInterface.
type estimatorAdapter struct {
	service *ai.EstimatorService
}

func (a *estimatorAdapter) EstimatePrice(ctx context.Context, imageURL string, productName string, condition string, mode string) (*services.PriceEstimateResult, error) {
	result, err := a.service.EstimatePrice(ctx, imageURL, productName, condition, mode)
	if err != nil {
		return nil, err
	}

	return &services.PriceEstimateResult{
		EstimatedPriceMin:    result.EstimatedPriceMin,
		EstimatedPriceMax:    result.EstimatedPriceMax,
		EstimatedPriceMedian: result.EstimatedPriceMedian,
		ConfidenceScore:      result.ConfidenceScore,
		Reasoning:            result.Reasoning,
		ModelUsed:            result.ModelUsed,
		ProcessingTimeMs:     result.ProcessingTimeMs,
		CostInUSD:            result.CostInUSD,
		MarketData:           result.MarketData,
	}, nil
}
