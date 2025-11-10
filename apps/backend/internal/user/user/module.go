package user

import (
	"context"
	"log/slog"
	"os"

	subServices "github.com/detect-price-by-photo/backend/internal/subscription/services"
	"github.com/detect-price-by-photo/backend/internal/user/user/handler"
	"github.com/detect-price-by-photo/backend/internal/user/user/repository"
	"github.com/detect-price-by-photo/backend/internal/user/user/services"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type Options struct {
	PgPool              *pgxpool.Pool                            // PostgreSQL connection pool (required)
	Logger              *slog.Logger                             // Slog logger instance (optional)
	SubscriptionService subServices.SubscriptionServiceInterface // Optional: for auto-enrollment
}

// subscriptionServiceAdapter adapts the subscription service to match UserService's interface
type subscriptionServiceAdapter struct {
	service subServices.SubscriptionServiceInterface
}

func (a *subscriptionServiceAdapter) CreateSubscription(ctx context.Context, userID uuid.UUID, planID uuid.UUID) error {
	_, err := a.service.CreateSubscription(ctx, userID, planID)
	return err
}

func (a *subscriptionServiceAdapter) GetPlanByName(ctx context.Context, name string) (uuid.UUID, error) {
	return a.service.GetPlanByName(ctx, name)
}

// UserModule holds dependencies for user-related handlers.
type UserModule struct {
	logger      *slog.Logger
	middlewares []echo.MiddlewareFunc
	handler     *handler.Handler
	userService services.UserServiceInterface
	userRepo    repository.UserRepositoryInterface
}

// NewModule creates a new UserModule.
func NewModule(opts *Options) *UserModule {
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	}

	// Initialize required services
	var subscriptionAdapter services.SubscriptionServiceInterface
	if opts.SubscriptionService != nil {
		subscriptionAdapter = &subscriptionServiceAdapter{service: opts.SubscriptionService}
	}

	userRepo := repository.NewUserRepository(opts.PgPool, logger)
	userService := services.NewUserService(services.UserServiceOpts{
		UserRepo:            userRepo,
		SubscriptionService: subscriptionAdapter,
		Logger:              logger,
	})

	h := handler.NewHandler(&handler.HandlerOpts{
		Logger:      logger,
		UserService: userService,
	})

	return &UserModule{
		logger:      logger,
		handler:     h,
		userService: userService,
		userRepo:    userRepo,
	}
}

// Expose UserService, so it can be used by other modules
func (m *UserModule) GetUserService() services.UserServiceInterface {
	return m.userService
}

// Expose UserRepository, so it can be used by other modules
func (m *UserModule) GetUserRepository() repository.UserRepositoryInterface {
	return m.userRepo
}

// Use adds middleware(s) to the UserModule (grouped).
func (m *UserModule) Use(mw ...echo.MiddlewareFunc) {
	m.middlewares = append(m.middlewares, mw...)
}

// RegisterRoutes registers user endpoints to the given Echo group.
func (m *UserModule) RegisterRoutes(e *echo.Group) {
	g := e.Group("/users", m.middlewares...)
	g.POST("", m.handler.CreateUser)
	g.GET("", m.handler.ListUsers)
	g.GET("/:userId", m.handler.GetUser)
	g.PUT("/:userId", m.handler.UpdateUser)
	g.DELETE("/:userId", m.handler.DeleteUser)
}
