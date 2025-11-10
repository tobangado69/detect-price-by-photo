package repository

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/detect-price-by-photo/backend/internal/payment/models"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Sentinel errors
var (
	ErrInvoiceNotFound = errors.New("invoice not found")
	ErrPaymentNotFound = errors.New("payment not found")
)

// InvoiceRepositoryInterface defines the contract for invoice data access.
type InvoiceRepositoryInterface interface {
	CreateInvoice(ctx context.Context, invoice *models.Invoice) error
	GetInvoice(ctx context.Context, id uuid.UUID) (*models.Invoice, error)
	GetInvoiceByID(ctx context.Context, id uuid.UUID) (*models.Invoice, error)
	UpdateInvoiceStatus(ctx context.Context, id uuid.UUID, status models.InvoiceStatus, paidAt *time.Time) error
	ListUserInvoices(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Invoice, int, error)
}

// PaymentRepositoryInterface defines the contract for payment data access.
type PaymentRepositoryInterface interface {
	CreatePayment(ctx context.Context, payment *models.Payment) error
	GetPaymentByID(ctx context.Context, id uuid.UUID) (*models.Payment, error)
	GetPaymentByTransactionID(ctx context.Context, transactionID string) (*models.Payment, error)
	UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status models.PaymentStatus, completedAt *time.Time, midtransResponse json.RawMessage) error
}

// Ensure repositories implement interfaces
var _ InvoiceRepositoryInterface = (*InvoiceRepository)(nil)
var _ PaymentRepositoryInterface = (*PaymentRepository)(nil)

// InvoiceRepository is an implementation of InvoiceRepositoryInterface using pgxpool.
type InvoiceRepository struct {
	pgPool *pgxpool.Pool
	logger *slog.Logger
}

// PaymentRepository is an implementation of PaymentRepositoryInterface using pgxpool.
type PaymentRepository struct {
	pgPool *pgxpool.Pool
	logger *slog.Logger
}

// NewInvoiceRepository creates a new InvoiceRepository.
func NewInvoiceRepository(pgPool *pgxpool.Pool, logger *slog.Logger) *InvoiceRepository {
	return &InvoiceRepository{
		pgPool: pgPool,
		logger: logger,
	}
}

// NewPaymentRepository creates a new PaymentRepository.
func NewPaymentRepository(pgPool *pgxpool.Pool, logger *slog.Logger) *PaymentRepository {
	return &PaymentRepository{
		pgPool: pgPool,
		logger: logger,
	}
}

// CreateInvoice creates a new invoice record.
func (r *InvoiceRepository) CreateInvoice(ctx context.Context, invoice *models.Invoice) error {
	if invoice.ID == uuid.Nil {
		invoice.ID = uuid.Must(uuid.NewV7())
	}
	invoice.CreatedAt = time.Now()

	query := `
		INSERT INTO ` + models.InvoiceTable + ` (
			id, user_id, subscription_id, amount_idr, amount_usd,
			period_start, period_end, status, payment_method,
			created_at, due_at, paid_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.pgPool.Exec(ctx, query,
		invoice.ID, invoice.UserID, invoice.SubscriptionID,
		invoice.AmountIDR, invoice.AmountUSD,
		invoice.PeriodStart, invoice.PeriodEnd,
		invoice.Status, invoice.PaymentMethod,
		invoice.CreatedAt, invoice.DueAt, invoice.PaidAt, invoice.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("failed to create invoice", "error", err)
		return err
	}
	return nil
}

// GetInvoice retrieves an invoice by ID.
func (r *InvoiceRepository) GetInvoice(ctx context.Context, id uuid.UUID) (*models.Invoice, error) {
	return r.GetInvoiceByID(ctx, id)
}

// GetInvoiceByID retrieves an invoice by ID.
func (r *InvoiceRepository) GetInvoiceByID(ctx context.Context, id uuid.UUID) (*models.Invoice, error) {
	query := `
		SELECT id, user_id, subscription_id, amount_idr, amount_usd,
		       period_start, period_end, status, payment_method,
		       created_at, due_at, paid_at, updated_at
		FROM ` + models.InvoiceTable + `
		WHERE id = $1
	`

	var invoice models.Invoice
	err := r.pgPool.QueryRow(ctx, query, id).Scan(
		&invoice.ID, &invoice.UserID, &invoice.SubscriptionID,
		&invoice.AmountIDR, &invoice.AmountUSD,
		&invoice.PeriodStart, &invoice.PeriodEnd,
		&invoice.Status, &invoice.PaymentMethod,
		&invoice.CreatedAt, &invoice.DueAt, &invoice.PaidAt, &invoice.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvoiceNotFound
		}
		r.logger.Error("failed to get invoice by ID", "id", id, "error", err)
		return nil, err
	}
	return &invoice, nil
}

// UpdateInvoiceStatus updates an invoice's status.
func (r *InvoiceRepository) UpdateInvoiceStatus(ctx context.Context, id uuid.UUID, status models.InvoiceStatus, paidAt *time.Time) error {
	query := `
		UPDATE ` + models.InvoiceTable + `
		SET status = $2, paid_at = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.pgPool.Exec(ctx, query, id, status, paidAt)
	if err != nil {
		r.logger.Error("failed to update invoice status", "id", id, "error", err)
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrInvoiceNotFound
	}
	return nil
}

// ListUserInvoices retrieves a paginated list of invoices for a user.
func (r *InvoiceRepository) ListUserInvoices(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Invoice, int, error) {
	// Count total
	countQuery := `SELECT COUNT(*) FROM ` + models.InvoiceTable + ` WHERE user_id = $1`
	var total int
	err := r.pgPool.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		r.logger.Error("failed to count user invoices", "user_id", userID, "error", err)
		return nil, 0, err
	}

	// Get invoices
	query := `
		SELECT id, user_id, subscription_id, amount_idr, amount_usd,
		       period_start, period_end, status, payment_method,
		       created_at, due_at, paid_at, updated_at
		FROM ` + models.InvoiceTable + `
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pgPool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		r.logger.Error("failed to list user invoices", "user_id", userID, "error", err)
		return nil, 0, err
	}
	defer rows.Close()

	var invoices []*models.Invoice
	for rows.Next() {
		var invoice models.Invoice
		if err := rows.Scan(
			&invoice.ID, &invoice.UserID, &invoice.SubscriptionID,
			&invoice.AmountIDR, &invoice.AmountUSD,
			&invoice.PeriodStart, &invoice.PeriodEnd,
			&invoice.Status, &invoice.PaymentMethod,
			&invoice.CreatedAt, &invoice.DueAt, &invoice.PaidAt, &invoice.UpdatedAt,
		); err != nil {
			r.logger.Error("failed to scan invoice row", "error", err)
			return nil, 0, err
		}
		invoices = append(invoices, &invoice)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating invoice rows", "error", err)
		return nil, 0, err
	}

	return invoices, total, nil
}

// CreatePayment creates a new payment record.
func (r *PaymentRepository) CreatePayment(ctx context.Context, payment *models.Payment) error {
	if payment.ID == uuid.Nil {
		payment.ID = uuid.Must(uuid.NewV7())
	}
	payment.CreatedAt = time.Now()

	query := `
		INSERT INTO ` + models.PaymentTable + ` (
			id, invoice_id, transaction_id, amount_idr, status,
			payment_method, payment_channel, midtrans_response,
			created_at, completed_at, expires_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.pgPool.Exec(ctx, query,
		payment.ID, payment.InvoiceID, payment.TransactionID,
		payment.AmountIDR, payment.Status,
		payment.PaymentMethod, payment.PaymentChannel,
		payment.MidtransResponse,
		payment.CreatedAt, payment.CompletedAt, payment.ExpiresAt, payment.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("failed to create payment", "error", err)
		return err
	}
	return nil
}

// GetPaymentByID retrieves a payment by ID.
func (r *PaymentRepository) GetPaymentByID(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
	query := `
		SELECT id, invoice_id, transaction_id, amount_idr, status,
		       payment_method, payment_channel, midtrans_response,
		       created_at, completed_at, expires_at, updated_at
		FROM ` + models.PaymentTable + `
		WHERE id = $1
	`

	var payment models.Payment
	err := r.pgPool.QueryRow(ctx, query, id).Scan(
		&payment.ID, &payment.InvoiceID, &payment.TransactionID,
		&payment.AmountIDR, &payment.Status,
		&payment.PaymentMethod, &payment.PaymentChannel,
		&payment.MidtransResponse,
		&payment.CreatedAt, &payment.CompletedAt, &payment.ExpiresAt, &payment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		r.logger.Error("failed to get payment by ID", "id", id, "error", err)
		return nil, err
	}
	return &payment, nil
}

// GetPaymentByTransactionID retrieves a payment by Midtrans transaction ID.
func (r *PaymentRepository) GetPaymentByTransactionID(ctx context.Context, transactionID string) (*models.Payment, error) {
	query := `
		SELECT id, invoice_id, transaction_id, amount_idr, status,
		       payment_method, payment_channel, midtrans_response,
		       created_at, completed_at, expires_at, updated_at
		FROM ` + models.PaymentTable + `
		WHERE transaction_id = $1
	`

	var payment models.Payment
	err := r.pgPool.QueryRow(ctx, query, transactionID).Scan(
		&payment.ID, &payment.InvoiceID, &payment.TransactionID,
		&payment.AmountIDR, &payment.Status,
		&payment.PaymentMethod, &payment.PaymentChannel,
		&payment.MidtransResponse,
		&payment.CreatedAt, &payment.CompletedAt, &payment.ExpiresAt, &payment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		r.logger.Error("failed to get payment by transaction ID", "transaction_id", transactionID, "error", err)
		return nil, err
	}
	return &payment, nil
}

// UpdatePaymentStatus updates a payment's status.
func (r *PaymentRepository) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status models.PaymentStatus, completedAt *time.Time, midtransResponse json.RawMessage) error {
	query := `
		UPDATE ` + models.PaymentTable + `
		SET status = $2, completed_at = $3, midtrans_response = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.pgPool.Exec(ctx, query, id, status, completedAt, midtransResponse)
	if err != nil {
		r.logger.Error("failed to update payment status", "id", id, "error", err)
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrPaymentNotFound
	}
	return nil
}

