package handler

import (
	"log/slog"
	"net/http"

	"github.com/detect-price-by-photo/backend/internal/subscription/services"
	"github.com/detect-price-by-photo/backend/internal/user/auth"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/gofrs/uuid/v5"
)

// HandlerInterface defines the contract for subscription handlers.
type HandlerInterface interface {
	ListPlans(c echo.Context) error
	GetCurrentSubscription(c echo.Context) error
}

// Ensure Handler implements HandlerInterface
var _ HandlerInterface = (*Handler)(nil)

// Handler holds dependencies for subscription handlers.
type Handler struct {
	logger            *slog.Logger
	subscriptionService services.SubscriptionServiceInterface
	validator         *validator.Validate
}

type HandlerOpts struct {
	Logger            *slog.Logger
	SubscriptionService services.SubscriptionServiceInterface
}

// NewHandler creates a new Handler instance.
func NewHandler(opts *HandlerOpts) *Handler {
	return &Handler{
		logger:            opts.Logger,
		subscriptionService: opts.SubscriptionService,
		validator:         validator.New(),
	}
}

// @Summary      List subscription plans
// @Description  Retrieves all available subscription plans
// @Tags         Subscriptions
// @Produce      json
// @Success      200  {array}   models.Plan
// @Router       /api/v1/subscriptions/plans [get]
func (h *Handler) ListPlans(c echo.Context) error {
	// Get active plans only (can be made configurable via query param)
	activeOnly := true
	plans, err := h.subscriptionService.ListPlans(c.Request().Context(), activeOnly)
	if err != nil {
		h.logger.Error("Failed to list plans", slog.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve plans"})
	}

	return c.JSON(http.StatusOK, plans)
}

// @Summary      Get current subscription
// @Description  Retrieves the current active subscription for the authenticated user
// @Tags         Subscriptions
// @Security     BearerAuth
// @Param        Authorization  header    string                      true  "Bearer {token}"
// @Produce      json
// @Success      200  {object}  models.Subscription
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/subscriptions/current [get]
func (h *Handler) GetCurrentSubscription(c echo.Context) error {
	// Get user ID from context (set by auth middleware)
	userIDStr, ok := auth.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID format"})
	}

	subscription, err := h.subscriptionService.GetCurrentSubscription(c.Request().Context(), userID)
	if err != nil {
		if err.Error() == "no active subscription found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "No active subscription found"})
		}
		h.logger.Error("Failed to get subscription", slog.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve subscription"})
	}

	// Get plan details
	plans, err := h.subscriptionService.ListPlans(c.Request().Context(), false)
	if err == nil {
		for _, plan := range plans {
			if plan.ID == subscription.PlanID {
				response := map[string]interface{}{
					"plan": plan.Name,
					"daily_limit": subscription.DailyPhotoLimit,
					"current_day_usage": subscription.CurrentDayUsage,
					"usage_reset_at": subscription.UsageResetAt,
					"current_period_end": subscription.CurrentPeriodEnd,
					"auto_renew": subscription.Status == "active" && subscription.CancelledAt == nil,
				}
				return c.JSON(http.StatusOK, response)
			}
		}
	}

	// Fallback to full subscription object
	return c.JSON(http.StatusOK, subscription)
}

