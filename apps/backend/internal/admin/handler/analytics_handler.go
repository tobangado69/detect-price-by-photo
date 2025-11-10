package handler

import (
	"net/http"

	"github.com/detect-price-by-photo/backend/internal/admin/services"
	"github.com/labstack/echo/v4"
)

// AnalyticsHandler handles admin analytics endpoints.
type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
}

// NewAnalyticsHandler creates a new AnalyticsHandler.
func NewAnalyticsHandler(analyticsService *services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GetDashboard handles GET /api/v1/admin/analytics/dashboard
func (h *AnalyticsHandler) GetDashboard(c echo.Context) error {
	metrics, err := h.analyticsService.GetDashboard(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to get dashboard metrics",
		})
	}

	return c.JSON(http.StatusOK, metrics)
}

// GetRevenueMetrics handles GET /api/v1/admin/analytics/revenue?period=daily|weekly|monthly
func (h *AnalyticsHandler) GetRevenueMetrics(c echo.Context) error {
	period := c.QueryParam("period")
	if period == "" {
		period = "monthly"
	}

	metrics, err := h.analyticsService.GetRevenueMetrics(c.Request().Context(), period)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to get revenue metrics",
		})
	}

	return c.JSON(http.StatusOK, metrics)
}

