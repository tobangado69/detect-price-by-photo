-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- Create Payments table - Tracks payment transactions from Midtrans
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.payments (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES public.invoices(id) ON DELETE CASCADE,
    transaction_id TEXT NOT NULL UNIQUE, -- Midtrans transaction ID
    amount_idr NUMERIC(10, 2) NOT NULL CHECK (amount_idr >= 0),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'settlement', 'capture', 'deny', 'cancel', 'expire', 'refund', 'partial_refund', 'chargeback')),
    payment_method TEXT,
    payment_channel TEXT, -- e.g., 'credit_card', 'gopay', 'qris', 'bank_transfer'
    midtrans_response JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ DEFAULT NULL,
    expires_at TIMESTAMPTZ DEFAULT NULL,
    updated_at TIMESTAMPTZ DEFAULT NULL
);

-- Indexes for efficient queries
CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_transaction_id ON public.payments (transaction_id);
CREATE INDEX IF NOT EXISTS idx_payments_invoice_id ON public.payments (invoice_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON public.payments (status);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON public.payments (created_at);

-- Trigger for updated_at
CREATE TRIGGER trg_payments_updated_at BEFORE UPDATE ON public.payments FOR EACH ROW EXECUTE FUNCTION fn_updated_at_value();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS trg_payments_updated_at ON public.payments;
DROP INDEX IF EXISTS idx_payments_created_at;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_invoice_id;
DROP INDEX IF EXISTS idx_payments_transaction_id;
DROP TABLE IF EXISTS public.payments;

-- +goose StatementEnd

