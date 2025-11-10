package photo

import (
	"log/slog"
	"os"

	"github.com/detect-price-by-photo/backend/internal/photo/handler"
	"github.com/detect-price-by-photo/backend/internal/photo/repository"
	"github.com/detect-price-by-photo/backend/internal/photo/services"
	"github.com/detect-price-by-photo/backend/internal/storage"
	subServices "github.com/detect-price-by-photo/backend/internal/subscription/services"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type Options struct {
	PgPool              *pgxpool.Pool
	Logger              *slog.Logger
	StorageClient       storage.StorageClientInterface
	SubscriptionService subServices.SubscriptionServiceInterface
	EstimatorService    services.EstimatorServiceInterface
}

// PhotoModule holds dependencies for photo-related handlers.
type PhotoModule struct {
	logger      *slog.Logger
	middlewares []echo.MiddlewareFunc
	handler     *handler.Handler
	photoService services.PhotoServiceInterface
}

// NewModule creates a new PhotoModule.
func NewModule(opts *Options) *PhotoModule {
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	}

	// Initialize repositories
	analysisRepo := repository.NewAnalysisRepository(opts.PgPool, logger)

	// Initialize service
	photoService := services.NewPhotoService(services.PhotoServiceOpts{
		AnalysisRepo:       analysisRepo,
		StorageClient:       opts.StorageClient,
		SubscriptionService: opts.SubscriptionService,
		EstimatorService:    opts.EstimatorService,
		Logger:              logger,
	})

	// Initialize handler
	h := handler.NewHandler(&handler.HandlerOpts{
		Logger:       logger,
		PhotoService: photoService,
	})

	return &PhotoModule{
		logger:      logger,
		handler:     h,
		photoService: photoService,
	}
}

// Expose PhotoService, so it can be used by other modules
func (m *PhotoModule) GetPhotoService() services.PhotoServiceInterface {
	return m.photoService
}

// Use adds middleware(s) to the PhotoModule (grouped).
func (m *PhotoModule) Use(mw ...echo.MiddlewareFunc) {
	m.middlewares = append(m.middlewares, mw...)
}

// RegisterRoutes registers photo endpoints to the given Echo group.
func (m *PhotoModule) RegisterRoutes(e *echo.Group) {
	// Protected routes (require auth)
	protectedGroup := e.Group("/analyses", m.middlewares...)
	protectedGroup.POST("/upload", m.handler.UploadPhoto)
	protectedGroup.POST("/:id/estimate", m.handler.EstimatePrice)
	protectedGroup.GET("/history", m.handler.ListUserAnalyses)
	protectedGroup.GET("/:id", m.handler.GetAnalysis)
	protectedGroup.DELETE("/:id", m.handler.DeleteAnalysis)
}

