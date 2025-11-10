-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- Create Invoices table - Tracks subscription billing invoices
-- ============================================================================

CREATE TABLE IF NOT EXISTS public.invoices (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    subscription_id UUID NOT NULL REFERENCES public.subscriptions(id) ON DELETE CASCADE,
    amount_idr NUMERIC(10, 2) NOT NULL CHECK (amount_idr >= 0),
    amount_usd NUMERIC(10, 2) NOT NULL CHECK (amount_usd >= 0),
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'cancelled', 'refunded')),
    payment_method TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    due_at TIMESTAMPTZ NOT NULL,
    paid_at TIMESTAMPTZ DEFAULT NULL,
    updated_at TIMESTAMPTZ DEFAULT NULL
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_invoices_user_id ON public.invoices (user_id);
CREATE INDEX IF NOT EXISTS idx_invoices_subscription_id ON public.invoices (subscription_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON public.invoices (status);
CREATE INDEX IF NOT EXISTS idx_invoices_created_at ON public.invoices (created_at);
CREATE INDEX IF NOT EXISTS idx_invoices_due_at ON public.invoices (due_at);

-- Trigger for updated_at
CREATE TRIGGER trg_invoices_updated_at BEFORE UPDATE ON public.invoices FOR EACH ROW EXECUTE FUNCTION fn_updated_at_value();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS trg_invoices_updated_at ON public.invoices;
DROP INDEX IF EXISTS idx_invoices_due_at;
DROP INDEX IF EXISTS idx_invoices_created_at;
DROP INDEX IF EXISTS idx_invoices_status;
DROP INDEX IF EXISTS idx_invoices_subscription_id;
DROP INDEX IF EXISTS idx_invoices_user_id;
DROP TABLE IF EXISTS public.invoices;

-- +goose StatementEnd

