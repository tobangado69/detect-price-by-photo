package models

import (
	"encoding/json"
	"time"

	"github.com/gofrs/uuid/v5"
)

// InvoiceStatus represents invoice status.
type InvoiceStatus string

const (
	InvoiceStatusPending  InvoiceStatus = "pending"
	InvoiceStatusPaid     InvoiceStatus = "paid"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
	InvoiceStatusRefunded InvoiceStatus = "refunded"
)

// PaymentStatus represents payment status from Midtrans.
type PaymentStatus string

const (
	PaymentStatusPending      PaymentStatus = "pending"
	PaymentStatusSettlement   PaymentStatus = "settlement"
	PaymentStatusCapture      PaymentStatus = "capture"
	PaymentStatusDeny         PaymentStatus = "deny"
	PaymentStatusCancel       PaymentStatus = "cancel"
	PaymentStatusExpire       PaymentStatus = "expire"
	PaymentStatusRefund       PaymentStatus = "refund"
	PaymentStatusPartialRefund PaymentStatus = "partial_refund"
	PaymentStatusChargeback   PaymentStatus = "chargeback"
)

// InvoiceTable is the table name for invoices.
const InvoiceTable = "public.invoices"

// PaymentTable is the table name for payments.
const PaymentTable = "public.payments"

// Invoice represents a subscription billing invoice.
type Invoice struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	SubscriptionID uuid.UUID
	AmountIDR     float64
	AmountUSD     float64
	PeriodStart    time.Time
	PeriodEnd      time.Time
	Status         InvoiceStatus
	PaymentMethod  *string
	CreatedAt      time.Time
	DueAt          time.Time
	PaidAt         *time.Time
	UpdatedAt      *time.Time
}

// IsPaid returns true if the invoice is paid.
func (i *Invoice) IsPaid() bool {
	return i.Status == InvoiceStatusPaid
}

// IsOverdue returns true if the invoice is overdue.
func (i *Invoice) IsOverdue() bool {
	return i.Status == InvoiceStatusPending && time.Now().After(i.DueAt)
}

// Payment represents a payment transaction from Midtrans.
type Payment struct {
	ID              uuid.UUID
	InvoiceID       uuid.UUID
	TransactionID   string // Midtrans transaction ID (unique)
	AmountIDR       float64
	Status          PaymentStatus
	PaymentMethod   *string
	PaymentChannel  *string
	MidtransResponse json.RawMessage // JSONB from database
	CreatedAt       time.Time
	CompletedAt     *time.Time
	ExpiresAt       *time.Time
	UpdatedAt       *time.Time
}

// IsSuccessful returns true if the payment is successful.
func (p *Payment) IsSuccessful() bool {
	return p.Status == PaymentStatusSettlement || p.Status == PaymentStatusCapture
}

// IsFailed returns true if the payment failed.
func (p *Payment) IsFailed() bool {
	return p.Status == PaymentStatusDeny || p.Status == PaymentStatusCancel || p.Status == PaymentStatusExpire
}

// CreateSubscriptionPaymentRequest represents the request to create a subscription payment.
type CreateSubscriptionPaymentRequest struct {
	PlanID uuid.UUID `json:"plan_id" validate:"required"`
}

// CreateSubscriptionPaymentResponse represents the response for creating a subscription payment.
type CreateSubscriptionPaymentResponse struct {
	SnapToken     string    `json:"snap_token"`
	RedirectURL   string    `json:"redirect_url"`
	TransactionID string    `json:"transaction_id"`
	InvoiceID     string    `json:"invoice_id"`
	DueAt         time.Time `json:"due_at"`
}

// InvoiceListResponse represents paginated invoice list.
type InvoiceListResponse struct {
	Invoices []InvoiceResponse `json:"invoices"`
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	Limit    int                `json:"limit"`
}

// InvoiceResponse represents an invoice in API responses.
type InvoiceResponse struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	SubscriptionID string     `json:"subscription_id"`
	AmountIDR     float64    `json:"amount_idr"`
	AmountUSD     float64    `json:"amount_usd"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
	Status         string    `json:"status"`
	PaymentMethod  *string   `json:"payment_method,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	DueAt          time.Time `json:"due_at"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
}

