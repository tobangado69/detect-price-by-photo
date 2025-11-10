package subscription

import (
	"log/slog"
	"os"

	"github.com/detect-price-by-photo/backend/internal/subscription/handler"
	"github.com/detect-price-by-photo/backend/internal/subscription/repository"
	"github.com/detect-price-by-photo/backend/internal/subscription/services"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type Options struct {
	PgPool *pgxpool.Pool // PostgreSQL connection pool (required)
	Logger *slog.Logger  // Slog logger instance (optional)
}

// SubscriptionModule holds dependencies for subscription-related handlers.
type SubscriptionModule struct {
	logger            *slog.Logger
	middlewares       []echo.MiddlewareFunc
	handler           *handler.Handler
	subscriptionService services.SubscriptionServiceInterface
	planRepo          repository.PlanRepositoryInterface
}

// NewModule creates a new SubscriptionModule.
func NewModule(opts *Options) *SubscriptionModule {
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	}

	// Initialize repositories
	planRepo := repository.NewPlanRepository(opts.PgPool, logger)
	subscriptionRepo := repository.NewSubscriptionRepository(opts.PgPool, logger)

	// Initialize service
	subscriptionService := services.NewSubscriptionService(services.SubscriptionServiceOpts{
		PlanRepo:         planRepo,
		SubscriptionRepo: subscriptionRepo,
	})

	// Initialize handler
	h := handler.NewHandler(&handler.HandlerOpts{
		Logger:            logger,
		SubscriptionService: subscriptionService,
	})

	return &SubscriptionModule{
		logger:            logger,
		handler:           h,
		subscriptionService: subscriptionService,
		planRepo:          planRepo,
	}
}

// Expose SubscriptionService, so it can be used by other modules
func (m *SubscriptionModule) GetSubscriptionService() services.SubscriptionServiceInterface {
	return m.subscriptionService
}

// Expose PlanRepository, so it can be used by other modules
func (m *SubscriptionModule) GetPlanRepository() repository.PlanRepositoryInterface {
	return m.planRepo
}

// Use adds middleware(s) to the SubscriptionModule (grouped).
func (m *SubscriptionModule) Use(mw ...echo.MiddlewareFunc) {
	m.middlewares = append(m.middlewares, mw...)
}

// RegisterRoutes registers subscription endpoints to the given Echo group.
func (m *SubscriptionModule) RegisterRoutes(e *echo.Group) {
	// Public route - list plans (no auth required)
	publicGroup := e.Group("/subscriptions", m.middlewares...)
	publicGroup.GET("/plans", m.handler.ListPlans)

	// Protected route - get current subscription (requires auth)
	protectedGroup := publicGroup.Group("", m.middlewares...)
	protectedGroup.GET("/current", m.handler.GetCurrentSubscription)
}

