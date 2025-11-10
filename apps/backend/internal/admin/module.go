package admin

import (
	"log/slog"
	"os"

	"github.com/detect-price-by-photo/backend/internal/admin/handler"
	"github.com/detect-price-by-photo/backend/internal/admin/repository"
	"github.com/detect-price-by-photo/backend/internal/admin/services"
	aiServices "github.com/detect-price-by-photo/backend/internal/ai/services"
	userRepo "github.com/detect-price-by-photo/backend/internal/user/user/repository"
	subRepo "github.com/detect-price-by-photo/backend/internal/subscription/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// AdminModule holds dependencies for admin-related handlers.
type AdminModule struct {
	logger           *slog.Logger
	middlewares      []echo.MiddlewareFunc
	userHandler      *handler.UserHandler
	analyticsHandler *handler.AnalyticsHandler
	planHandler      *handler.PlanHandler
	modelHandler     *handler.ModelHandler
}

// Options for creating AdminModule.
type Options struct {
	PgPool      *pgxpool.Pool
	Logger      *slog.Logger
	UserRepo    userRepo.UserRepositoryInterface
	PlanRepo    subRepo.PlanRepositoryInterface
	ModelService aiServices.ModelServiceInterface
}

// NewModule creates a new AdminModule.
func NewModule(opts *Options) *AdminModule {
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	}

	// Initialize repositories
	auditLogRepo := repository.NewAuditLogRepository(opts.PgPool, logger)

	// Initialize services
	analyticsService := services.NewAnalyticsService(opts.PgPool, logger)

	// Initialize handlers
	userHandler := handler.NewUserHandler(opts.UserRepo, auditLogRepo, logger)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService)
	planHandler := handler.NewPlanHandler(opts.PlanRepo, auditLogRepo, logger)
	
	var modelHandler *handler.ModelHandler
	if opts.ModelService != nil {
		modelHandler = handler.NewModelHandler(opts.ModelService, auditLogRepo, logger)
	}

	return &AdminModule{
		logger:           logger,
		userHandler:      userHandler,
		analyticsHandler: analyticsHandler,
		planHandler:      planHandler,
		modelHandler:     modelHandler,
	}
}

// Use adds middleware(s) to the AdminModule.
func (m *AdminModule) Use(mw ...echo.MiddlewareFunc) {
	m.middlewares = append(m.middlewares, mw...)
}

// RegisterRoutes registers admin endpoints to the given Echo group.
func (m *AdminModule) RegisterRoutes(e *echo.Group) {
	// All admin routes require admin middleware (injected via Use())
	adminGroup := e.Group("/admin", m.middlewares...)

	// User management
	adminGroup.GET("/users", m.userHandler.ListUsers)
	adminGroup.GET("/users/:id", m.userHandler.GetUser)
	adminGroup.PUT("/users/:id", m.userHandler.UpdateUser)

	// Analytics
	adminGroup.GET("/analytics/dashboard", m.analyticsHandler.GetDashboard)
	adminGroup.GET("/analytics/revenue", m.analyticsHandler.GetRevenueMetrics)

	// Plan management
	adminGroup.GET("/plans", m.planHandler.ListPlans)
	adminGroup.POST("/plans", m.planHandler.CreatePlan)
	adminGroup.PUT("/plans/:id", m.planHandler.UpdatePlan)
	adminGroup.DELETE("/plans/:id", m.planHandler.DeletePlan)

	// AI Model management (only if ModelService is provided)
	if m.modelHandler != nil {
		adminGroup.GET("/models", m.modelHandler.ListModels)
		adminGroup.PUT("/models/default", m.modelHandler.SetDefaultModel)
		adminGroup.PUT("/models/:id", m.modelHandler.UpdateModel)
	}
}

