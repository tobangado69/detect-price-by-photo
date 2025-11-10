package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/detect-price-by-photo/backend/internal/payment/models"
	"github.com/detect-price-by-photo/backend/internal/payment/services"
	"github.com/detect-price-by-photo/backend/internal/user/auth"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v4"
)

// HandlerInterface defines the contract for payment handlers.
type HandlerInterface interface {
	CreateSubscriptionPayment(c echo.Context) error
	HandleWebhook(c echo.Context) error
	ListUserInvoices(c echo.Context) error
}

// Ensure Handler implements HandlerInterface
var _ HandlerInterface = (*Handler)(nil)

// Handler holds dependencies for payment handlers.
type Handler struct {
	logger        *slog.Logger
	paymentService services.PaymentServiceInterface
	validator     *validator.Validate
}

type HandlerOpts struct {
	Logger        *slog.Logger
	PaymentService services.PaymentServiceInterface
}

// NewHandler creates a new Handler instance.
func NewHandler(opts *HandlerOpts) *Handler {
	return &Handler{
		logger:        opts.Logger,
		paymentService: opts.PaymentService,
		validator:     validator.New(),
	}
}

// @Summary      Create subscription payment
// @Description  Creates an invoice and Midtrans transaction for subscription upgrade
// @Tags         Payments
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                                    true  "Bearer {token}"
// @Param        request         body      models.CreateSubscriptionPaymentRequest  true  "Payment request"
// @Success      200             {object}  models.CreateSubscriptionPaymentResponse
// @Failure      400             {object}  map[string]string
// @Failure      401             {object}  map[string]string
// @Router       /api/v1/subscriptions/subscribe [post]
func (h *Handler) CreateSubscriptionPayment(c echo.Context) error {
	ctx := c.Request().Context()
	userIDStr, ok := auth.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid user ID"})
	}

	var req models.CreateSubscriptionPaymentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := h.validator.Struct(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	response, err := h.paymentService.CreateSubscriptionPayment(ctx, userID, req.PlanID)
	if err != nil {
		h.logger.Error("Failed to create subscription payment", slog.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create payment transaction"})
	}

	return c.JSON(http.StatusOK, response)
}

// @Summary      Handle Midtrans webhook
// @Description  Processes payment webhook notifications from Midtrans
// @Tags         Payments
// @Accept       json
// @Produce      json
// @Param        payload  body      map[string]interface{}  true  "Midtrans webhook payload"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Router       /api/v1/payments/webhook [post]
func (h *Handler) HandleWebhook(c echo.Context) error {
	ctx := c.Request().Context()

	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid webhook payload"})
	}

	if err := h.paymentService.HandleWebhook(ctx, payload); err != nil {
		h.logger.Error("Failed to process webhook", slog.String("error", err.Error()))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to process webhook"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "success", "message": "Webhook processed successfully"})
}

// @Summary      List user invoices
// @Description  Retrieves a paginated list of invoices for the authenticated user
// @Tags         Payments
// @Security     BearerAuth
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer {token}"
// @Param        page           query     int     false "Page number" default(1)
// @Param        limit          query     int     false "Items per page" default(20)
// @Success      200            {object}  models.InvoiceListResponse
// @Router       /api/v1/invoices [get]
func (h *Handler) ListUserInvoices(c echo.Context) error {
	ctx := c.Request().Context()
	userIDStr, ok := auth.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid user ID"})
	}

	// Parse pagination params
	page := 1
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	invoices, total, err := h.paymentService.ListUserInvoices(ctx, userID, page, limit)
	if err != nil {
		h.logger.Error("Failed to list invoices", slog.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve invoices"})
	}

	// Convert to response
	responses := make([]models.InvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		responses[i] = models.InvoiceResponse{
			ID:             invoice.ID.String(),
			UserID:         invoice.UserID.String(),
			SubscriptionID: invoice.SubscriptionID.String(),
			AmountIDR:     invoice.AmountIDR,
			AmountUSD:     invoice.AmountUSD,
			PeriodStart:    invoice.PeriodStart,
			PeriodEnd:      invoice.PeriodEnd,
			Status:         string(invoice.Status),
			PaymentMethod:  invoice.PaymentMethod,
			CreatedAt:      invoice.CreatedAt,
			DueAt:          invoice.DueAt,
			PaidAt:         invoice.PaidAt,
		}
	}

	return c.JSON(http.StatusOK, models.InvoiceListResponse{
		Invoices: responses,
		Total:    total,
		Page:     page,
		Limit:    limit,
	})
}

