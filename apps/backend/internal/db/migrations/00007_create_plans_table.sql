-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- Create Plans table - Subscription plans (Free, Premium, Enterprise)
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.plans (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE CHECK (char_length(name) > 0),
    daily_photo_limit INTEGER NOT NULL CHECK (daily_photo_limit >= 0),
    price_idr BIGINT NOT NULL DEFAULT 0 CHECK (price_idr >= 0),
    price_usd NUMERIC(10, 4) NOT NULL DEFAULT 0 CHECK (price_usd >= 0),
    description TEXT,
    features JSONB DEFAULT '[]'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT NULL
);

-- Plans table indexes and updated_at trigger
CREATE INDEX IF NOT EXISTS idx_plans_name ON public.plans (name);
CREATE INDEX IF NOT EXISTS idx_plans_is_active ON public.plans (is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_plans_created_at ON public.plans (created_at);
CREATE TRIGGER trg_plans_updated_at BEFORE UPDATE ON public.plans FOR EACH ROW EXECUTE FUNCTION fn_updated_at_value();

-- Seed initial plans
INSERT INTO public.plans (name, daily_photo_limit, price_idr, price_usd, description, features) VALUES
    ('free', 10, 0, 0, 'Free tier for casual traders', '["10 daily analyses", "7-day history", "basic stats"]'::jsonb),
    ('premium', 1000, 65000, 4.50, 'Premium tier for active sellers', '["1K daily analyses", "90-day history", "batch API", "team features"]'::jsonb),
    ('enterprise', 999999, 0, 0, 'Enterprise tier with unlimited access', '["Unlimited analyses", "1-year history", "advanced API", "dedicated support"]'::jsonb)
ON CONFLICT (name) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop triggers, indexes, and table (reverse order of creation)
DROP TRIGGER IF EXISTS trg_plans_updated_at ON public.plans;
DROP INDEX IF EXISTS idx_plans_created_at;
DROP INDEX IF EXISTS idx_plans_is_active;
DROP INDEX IF EXISTS idx_plans_name;
DROP TABLE IF EXISTS public.plans;

-- +goose StatementEnd

