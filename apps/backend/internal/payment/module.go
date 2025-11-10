package payment

import (
	"context"
	"log/slog"
	"os"

	"github.com/detect-price-by-photo/backend/internal/payment/handler"
	"github.com/detect-price-by-photo/backend/internal/payment/repository"
	"github.com/detect-price-by-photo/backend/internal/payment/services"
	subServices "github.com/detect-price-by-photo/backend/internal/subscription/services"
	userServices "github.com/detect-price-by-photo/backend/internal/user/user/services"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// MidtransClientInterface defines the interface for Midtrans operations (to avoid circular dependency).
type MidtransClientInterface interface {
	CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*CreateTransactionResponse, error)
	ValidateWebhook(orderID, statusCode, signatureKey string) bool
	GetTransactionStatus(ctx context.Context, orderID string) (map[string]interface{}, error)
}

type Options struct {
	PgPool              *pgxpool.Pool
	Logger              *slog.Logger
	MidtransClient      MidtransClientInterface
	SubscriptionService subServices.SubscriptionServiceInterface
	UserService         userServices.UserServiceInterface
}

// PaymentModule holds dependencies for payment-related handlers.
type PaymentModule struct {
	logger        *slog.Logger
	middlewares   []echo.MiddlewareFunc
	handler       *handler.Handler
	paymentService services.PaymentServiceInterface
}

// NewModule creates a new PaymentModule.
func NewModule(opts *Options) *PaymentModule {
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	}

	// Initialize repositories
	invoiceRepo := repository.NewInvoiceRepository(opts.PgPool, logger)
	paymentRepo := repository.NewPaymentRepository(opts.PgPool, logger)

	// Initialize service
	paymentService := services.NewPaymentService(services.PaymentServiceOpts{
		InvoiceRepo:         invoiceRepo,
		PaymentRepo:         paymentRepo,
		MidtransClient:      &midtransAdapter{client: opts.MidtransClient},
		SubscriptionService: opts.SubscriptionService,
		UserService:         opts.UserService,
		Logger:              logger,
	})

	// Initialize handler
	h := handler.NewHandler(&handler.HandlerOpts{
		Logger:        logger,
		PaymentService: paymentService,
	})

	return &PaymentModule{
		logger:        logger,
		handler:       h,
		paymentService: paymentService,
	}
}

// Expose PaymentService, so it can be used by other modules
func (m *PaymentModule) GetPaymentService() services.PaymentServiceInterface {
	return m.paymentService
}

// Use adds middleware(s) to the PaymentModule (grouped).
func (m *PaymentModule) Use(mw ...echo.MiddlewareFunc) {
	m.middlewares = append(m.middlewares, mw...)
}

// RegisterRoutes registers payment endpoints to the given Echo group.
func (m *PaymentModule) RegisterRoutes(e *echo.Group) {
	// Protected routes (require auth)
	protectedGroup := e.Group("/subscriptions", m.middlewares...)
	protectedGroup.POST("/subscribe", m.handler.CreateSubscriptionPayment)

	protectedGroup = e.Group("/invoices", m.middlewares...)
	protectedGroup.GET("", m.handler.ListUserInvoices)

	// Public webhook endpoint (no auth required, signature validation instead)
	e.POST("/payments/webhook", m.handler.HandleWebhook)
}

// midtransAdapter adapts payment.MidtransClientInterface to services.MidtransClientInterface.
type midtransAdapter struct {
	client MidtransClientInterface
}

func (a *midtransAdapter) CreateTransaction(ctx context.Context, req services.CreateTransactionRequest) (*services.CreateTransactionResponse, error) {
	// Convert from services types to payment types
	paymentReq := CreateTransactionRequest{
		OrderID:     req.OrderID,
		Amount:      req.Amount,
		Email:       req.Email,
		Phone:       req.Phone,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Description: req.Description,
	}

	resp, err := a.client.CreateTransaction(ctx, paymentReq)
	if err != nil {
		return nil, err
	}

	// Convert back to services types
	return &services.CreateTransactionResponse{
		Token:         resp.Token,
		RedirectURL:   resp.RedirectURL,
		TransactionID: resp.TransactionID,
	}, nil
}

func (a *midtransAdapter) ValidateWebhook(orderID, statusCode, signatureKey string) bool {
	return a.client.ValidateWebhook(orderID, statusCode, signatureKey)
}

func (a *midtransAdapter) GetTransactionStatus(ctx context.Context, orderID string) (map[string]interface{}, error) {
	return a.client.GetTransactionStatus(ctx, orderID)
}

