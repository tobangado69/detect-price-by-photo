package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	paymentModels "github.com/detect-price-by-photo/backend/internal/payment/models"
	"github.com/detect-price-by-photo/backend/internal/payment/repository"
	subModels "github.com/detect-price-by-photo/backend/internal/subscription/models"
	userModels "github.com/detect-price-by-photo/backend/internal/user/user/models"
	"github.com/gofrs/uuid/v5"
)

// SubscriptionServiceInterface defines the interface for subscription operations needed by PaymentService.
type SubscriptionServiceInterface interface {
	GetPlanByID(ctx context.Context, planID uuid.UUID) (*subModels.Plan, error)
	GetCurrentSubscription(ctx context.Context, userID uuid.UUID) (*subModels.Subscription, error)
	CreateSubscription(ctx context.Context, userID uuid.UUID, planID uuid.UUID) (*subModels.Subscription, error)
	UpdateSubscription(ctx context.Context, subscription *subModels.Subscription) error
}

// UserServiceInterface defines the interface for user operations needed by PaymentService.
type UserServiceInterface interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*userModels.User, error)
}

// MidtransClientInterface defines the interface for Midtrans operations.
type MidtransClientInterface interface {
	CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*CreateTransactionResponse, error)
	ValidateWebhook(orderID, statusCode, signatureKey string) bool
	GetTransactionStatus(ctx context.Context, orderID string) (map[string]interface{}, error)
}

// CreateTransactionRequest represents a transaction creation request.
type CreateTransactionRequest struct {
	OrderID     string
	Amount      int64  // Amount in IDR (rupiah)
	Email       string
	Phone       string
	FirstName   string
	LastName    string
	Description string
}

// CreateTransactionResponse represents the response from Midtrans.
type CreateTransactionResponse struct {
	Token         string `json:"token"`
	RedirectURL   string `json:"redirect_url"`
	TransactionID string `json:"transaction_id"`
}

// PaymentServiceInterface defines the contract for payment business logic.
type PaymentServiceInterface interface {
	CreateSubscriptionPayment(ctx context.Context, userID uuid.UUID, planID uuid.UUID) (*paymentModels.CreateSubscriptionPaymentResponse, error)
	HandleWebhook(ctx context.Context, payload map[string]interface{}) error
	ActivateSubscription(ctx context.Context, paymentID uuid.UUID) error
	ListUserInvoices(ctx context.Context, userID uuid.UUID, page, limit int) ([]*paymentModels.Invoice, int, error)
}

// Ensure PaymentService implements PaymentServiceInterface
var _ PaymentServiceInterface = (*PaymentService)(nil)

// PaymentService implements payment business logic.
type PaymentService struct {
	invoiceRepo         repository.InvoiceRepositoryInterface
	paymentRepo         repository.PaymentRepositoryInterface
	midtransClient      MidtransClientInterface
	subscriptionService SubscriptionServiceInterface
	userService         UserServiceInterface
	logger              *slog.Logger
}

type PaymentServiceOpts struct {
	InvoiceRepo         repository.InvoiceRepositoryInterface
	PaymentRepo         repository.PaymentRepositoryInterface
	MidtransClient      MidtransClientInterface
	SubscriptionService SubscriptionServiceInterface
	UserService         UserServiceInterface
	Logger              *slog.Logger
}

// NewPaymentService creates a new PaymentService.
func NewPaymentService(opts PaymentServiceOpts) *PaymentService {
	return &PaymentService{
		invoiceRepo:         opts.InvoiceRepo,
		paymentRepo:         opts.PaymentRepo,
		midtransClient:      opts.MidtransClient,
		subscriptionService: opts.SubscriptionService,
		userService:         opts.UserService,
		logger:              opts.Logger,
	}
}

// CreateSubscriptionPayment creates an invoice and Midtrans transaction for subscription upgrade.
func (s *PaymentService) CreateSubscriptionPayment(ctx context.Context, userID uuid.UUID, planID uuid.UUID) (*paymentModels.CreateSubscriptionPaymentResponse, error) {
	// Get plan details
	plan, err := s.subscriptionService.GetPlanByID(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	// Get user details
	user, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get current subscription
	currentSub, err := s.subscriptionService.GetCurrentSubscription(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current subscription: %w", err)
	}

	// Calculate billing period (1 month from now)
	now := time.Now()
	periodStart := now
	periodEnd := now.AddDate(0, 1, 0) // 1 month
	dueAt := now.AddDate(0, 0, 7)    // 7 days to pay

	// Create invoice
	invoice := &paymentModels.Invoice{
		ID:             uuid.Must(uuid.NewV7()),
		UserID:         userID,
		SubscriptionID: currentSub.ID,
		AmountIDR:     float64(plan.PriceIDR),
		AmountUSD:     plan.PriceUSD,
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
		Status:         paymentModels.InvoiceStatusPending,
		DueAt:          dueAt,
	}

	if err := s.invoiceRepo.CreateInvoice(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Create Midtrans transaction
	orderID := fmt.Sprintf("INV-%s", invoice.ID.String())
	midtransReq := CreateTransactionRequest{
		OrderID:     orderID,
		Amount:      plan.PriceIDR,
		Email:       user.Email,
		Phone:       "", // Add phone to user model if needed
		FirstName:   user.DisplayName,
		LastName:    "",
		Description: fmt.Sprintf("Subscription: %s", plan.Name),
	}

	midtransResp, err := s.midtransClient.CreateTransaction(ctx, midtransReq)
	if err != nil {
		// Update invoice status to cancelled if Midtrans fails
		_ = s.invoiceRepo.UpdateInvoiceStatus(ctx, invoice.ID, paymentModels.InvoiceStatusCancelled, nil)
		return nil, fmt.Errorf("failed to create Midtrans transaction: %w", err)
	}

	// Create payment record
	payment := &paymentModels.Payment{
		ID:            uuid.Must(uuid.NewV7()),
		InvoiceID:     invoice.ID,
		TransactionID: midtransResp.TransactionID,
		AmountIDR:     float64(plan.PriceIDR),
		Status:        paymentModels.PaymentStatusPending,
		ExpiresAt:     &dueAt,
	}

	if err := s.paymentRepo.CreatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	s.logger.Info("subscription payment created", "user_id", userID, "plan_id", planID, "invoice_id", invoice.ID, "transaction_id", midtransResp.TransactionID)

	return &paymentModels.CreateSubscriptionPaymentResponse{
		SnapToken:     midtransResp.Token,
		RedirectURL:   midtransResp.RedirectURL,
		TransactionID: midtransResp.TransactionID,
		InvoiceID:     invoice.ID.String(),
		DueAt:         dueAt,
	}, nil
}

// HandleWebhook processes Midtrans webhook notifications.
func (s *PaymentService) HandleWebhook(ctx context.Context, payload map[string]interface{}) error {
	// Extract webhook data
	transactionID, _ := payload["transaction_id"].(string)
	orderID, _ := payload["order_id"].(string)
	statusCode, _ := payload["transaction_status"].(string)
	signatureKey, _ := payload["signature_key"].(string)

	if transactionID == "" || orderID == "" {
		return fmt.Errorf("missing required webhook fields")
	}

	// Validate webhook signature
	if !s.midtransClient.ValidateWebhook(orderID, statusCode, signatureKey) {
		s.logger.Warn("invalid webhook signature", "order_id", orderID, "transaction_id", transactionID)
		return fmt.Errorf("invalid webhook signature")
	}

	// Check for duplicate processing (idempotency)
	existingPayment, err := s.paymentRepo.GetPaymentByTransactionID(ctx, transactionID)
	if err == nil && existingPayment != nil {
		// Payment already processed
		if existingPayment.IsSuccessful() {
			s.logger.Info("webhook already processed", "transaction_id", transactionID)
			return nil // Idempotent - return success
		}
	}

	// Extract invoice ID from order ID (format: INV-{uuid})
	if len(orderID) < 4 || orderID[:4] != "INV-" {
		return fmt.Errorf("invalid order ID format: %s", orderID)
	}
	invoiceIDStr := orderID[4:]
	invoiceID, err := uuid.FromString(invoiceIDStr)
	if err != nil {
		return fmt.Errorf("invalid invoice ID in order ID: %w", err)
	}

	// Get invoice
	invoice, err := s.invoiceRepo.GetInvoiceByID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	// Get or create payment
	payment, err := s.paymentRepo.GetPaymentByTransactionID(ctx, transactionID)
	if err != nil {
		// Payment not found, create it
		payment = &paymentModels.Payment{
			ID:            uuid.Must(uuid.NewV7()),
			InvoiceID:     invoiceID,
			TransactionID: transactionID,
			AmountIDR:     invoice.AmountIDR,
			Status:        paymentModels.PaymentStatusPending,
		}
		if err := s.paymentRepo.CreatePayment(ctx, payment); err != nil {
			return fmt.Errorf("failed to create payment: %w", err)
		}
	}

	// Map Midtrans status to our PaymentStatus
	paymentStatus := mapMidtransStatus(statusCode)
	
	// Marshal payload to JSONB
	midtransResponseJSON, _ := json.Marshal(payload)

	// Update payment status
	var completedAt *time.Time
	if paymentStatus == paymentModels.PaymentStatusSettlement || paymentStatus == paymentModels.PaymentStatusCapture {
		now := time.Now()
		completedAt = &now
	}

	if err := s.paymentRepo.UpdatePaymentStatus(ctx, payment.ID, paymentStatus, completedAt, midtransResponseJSON); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// If payment successful, activate subscription
	if paymentStatus == paymentModels.PaymentStatusSettlement || paymentStatus == paymentModels.PaymentStatusCapture {
		if err := s.ActivateSubscription(ctx, payment.ID); err != nil {
			s.logger.Error("failed to activate subscription after payment", "payment_id", payment.ID, "error", err)
			// Don't fail the webhook - we can retry activation later
		}
	}

	s.logger.Info("webhook processed", "transaction_id", transactionID, "status", statusCode)
	return nil
}

// ActivateSubscription activates a subscription after successful payment.
func (s *PaymentService) ActivateSubscription(ctx context.Context, paymentID uuid.UUID) error {
	// Get payment
	payment, err := s.paymentRepo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	if !payment.IsSuccessful() {
		return fmt.Errorf("payment is not successful, cannot activate subscription")
	}

	// Get invoice
	invoice, err := s.invoiceRepo.GetInvoiceByID(ctx, payment.InvoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	// Update invoice status
	if err := s.invoiceRepo.UpdateInvoiceStatus(ctx, invoice.ID, paymentModels.InvoiceStatusPaid, payment.CompletedAt); err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	// Get subscription and update it
	subscription, err := s.subscriptionService.GetCurrentSubscription(ctx, invoice.UserID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Update subscription period
	subscription.CurrentPeriodStart = invoice.PeriodStart
	subscription.CurrentPeriodEnd = invoice.PeriodEnd
	subscription.Status = "active"

	if err := s.subscriptionService.UpdateSubscription(ctx, subscription); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	s.logger.Info("subscription activated", "user_id", invoice.UserID, "subscription_id", subscription.ID, "payment_id", paymentID)
	return nil
}

// ListUserInvoices retrieves a paginated list of invoices for a user.
func (s *PaymentService) ListUserInvoices(ctx context.Context, userID uuid.UUID, page, limit int) ([]*paymentModels.Invoice, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	return s.invoiceRepo.ListUserInvoices(ctx, userID, limit, offset)
}

// mapMidtransStatus maps Midtrans transaction status to our PaymentStatus.
func mapMidtransStatus(midtransStatus string) paymentModels.PaymentStatus {
	switch midtransStatus {
	case "settlement":
		return paymentModels.PaymentStatusSettlement
	case "capture":
		return paymentModels.PaymentStatusCapture
	case "deny":
		return paymentModels.PaymentStatusDeny
	case "cancel":
		return paymentModels.PaymentStatusCancel
	case "expire":
		return paymentModels.PaymentStatusExpire
	case "refund":
		return paymentModels.PaymentStatusRefund
	case "partial_refund":
		return paymentModels.PaymentStatusPartialRefund
	case "chargeback":
		return paymentModels.PaymentStatusChargeback
	default:
		return paymentModels.PaymentStatusPending
	}
}

