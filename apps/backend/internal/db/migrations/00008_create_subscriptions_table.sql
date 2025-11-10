-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- Create Subscriptions table - User subscription tracking with daily quota
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.subscriptions (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES public.plans(id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'paused', 'cancelled')),
    daily_photo_limit INTEGER NOT NULL CHECK (daily_photo_limit >= 0),
    current_day_usage INTEGER NOT NULL DEFAULT 0 CHECK (current_day_usage >= 0),
    usage_reset_at TIMESTAMPTZ NOT NULL DEFAULT timezone('Asia/Jakarta', date_trunc('day', CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Jakarta') + INTERVAL '1 day'),
    started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    current_period_start TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    current_period_end TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP + INTERVAL '1 month'),
    cancelled_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT NULL,
    CONSTRAINT chk_usage_not_exceed_limit CHECK (current_day_usage <= daily_photo_limit)
);

-- Subscriptions table indexes and updated_at trigger
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON public.subscriptions (user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_plan_id ON public.subscriptions (plan_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON public.subscriptions (status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_status ON public.subscriptions (user_id, status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_usage_reset_at ON public.subscriptions (usage_reset_at);
CREATE INDEX IF NOT EXISTS idx_subscriptions_created_at ON public.subscriptions (created_at);
CREATE TRIGGER trg_subscriptions_updated_at BEFORE UPDATE ON public.subscriptions FOR EACH ROW EXECUTE FUNCTION fn_updated_at_value();

-- Ensure one active subscription per user (enforced at application level, but index helps)
CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_user_active ON public.subscriptions (user_id) WHERE status = 'active';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop triggers, indexes, and table (reverse order of creation)
DROP TRIGGER IF EXISTS trg_subscriptions_updated_at ON public.subscriptions;
DROP INDEX IF EXISTS idx_subscriptions_user_active;
DROP INDEX IF EXISTS idx_subscriptions_created_at;
DROP INDEX IF EXISTS idx_subscriptions_usage_reset_at;
DROP INDEX IF EXISTS idx_subscriptions_user_status;
DROP INDEX IF EXISTS idx_subscriptions_status;
DROP INDEX IF EXISTS idx_subscriptions_plan_id;
DROP INDEX IF EXISTS idx_subscriptions_user_id;
DROP TABLE IF EXISTS public.subscriptions;

-- +goose StatementEnd

