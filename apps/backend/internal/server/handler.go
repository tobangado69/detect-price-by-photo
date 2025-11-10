package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/detect-price-by-photo/backend/internal/config"

	"github.com/alexliesenfeld/health"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// ServerHandler holds dependencies for HTTP handlers.
type ServerHandler struct {
	PGPool      *pgxpool.Pool
	Logger      *slog.Logger
	StorageDir  string // Local storage directory for file serving
}

// NewServerHandler creates a new ServerHandler.
func NewServerHandler(pgPool *pgxpool.Pool, logger *slog.Logger) *ServerHandler {
	return &ServerHandler{
		PGPool:     pgPool,
		Logger:     logger,
		StorageDir: "./storage/uploads",
	}
}

// RegisterRoutes registers all HTTP routes to the given Echo instance.
func (h *ServerHandler) RegisterRoutes(e *echo.Echo) {
	e.GET("/healthz", h.HealthCheckHandler)          // Health check endpoint
	e.GET("/api/openapi.json", h.OpenAPISpecHandler) // Placeholder API spec
	e.GET("/files/*", h.ServeFiles)                  // Serve uploaded files (local dev only)
}

// @Summary		    Service healthcheck
// @Description	    Checks the health of the service
// @Tags	        General Information
// @Router		    /healthz [get]
func (h *ServerHandler) HealthCheckHandler(c echo.Context) error {
	hc := health.NewChecker(
		health.WithCacheDuration(10*time.Second),
		health.WithTimeout(5*time.Second),
		health.WithCheck(health.Check{
			Name:    "database",
			Timeout: 2 * time.Second,
			Check: func(ctx context.Context) error {
				return h.PGPool.Ping(ctx)
			},
		}),
	)

	// Transform health.NewHandler to Echo handler
	handler := health.NewHandler(hc)
	handler.ServeHTTP(c.Response(), c.Request())

	return nil
}

// OpenAPISpecHandler serves the embedded swagger.json as application/json.
func (h *ServerHandler) OpenAPISpecHandler(c echo.Context) error {
	cfg := config.Get()

	if !cfg.IsAPIDocsEnabled() {
		return echo.NewHTTPError(http.StatusNotFound, "API docs are disabled")
	}

	response := []byte(`{"info":{"title":"Detect Price API","version":"0.0.1"},"paths":{}}`)
	return c.Blob(http.StatusOK, "application/json", response)
}

// ServeFiles serves uploaded files from local storage (development only).
func (h *ServerHandler) ServeFiles(c echo.Context) error {
	// Get file path from URL
	requestedPath := c.Param("*")
	if requestedPath == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "file path required")
	}

	// Sanitize path to prevent directory traversal
	requestedPath = filepath.Clean(requestedPath)
	if strings.Contains(requestedPath, "..") {
		return echo.NewHTTPError(http.StatusForbidden, "invalid file path")
	}

	// Build full file path
	fullPath := filepath.Join(h.StorageDir, requestedPath)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return echo.NewHTTPError(http.StatusNotFound, "file not found")
	}

	// Serve file
	return c.File(fullPath)
}
