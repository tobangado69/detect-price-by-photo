-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- Create Price Analyses table - Stores photo uploads and price estimation results
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.price_analyses (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    product_name TEXT NOT NULL CHECK (char_length(product_name) > 0),
    condition TEXT NOT NULL DEFAULT 'good' CHECK (condition IN ('new', 'like_new', 'good', 'fair', 'poor')),
    mode TEXT NOT NULL DEFAULT 'fast' CHECK (mode IN ('fast', 'accurate', 'knowledge_based')),
    
    -- Price estimation results
    estimated_price_min NUMERIC(12, 2),
    estimated_price_max NUMERIC(12, 2),
    estimated_price_median NUMERIC(12, 2),
    confidence_score NUMERIC(5, 2) CHECK (confidence_score >= 0 AND confidence_score <= 100),
    reasoning TEXT,
    
    -- Processing metadata
    processing_time_ms INTEGER,
    cost_in_usd NUMERIC(10, 6),
    model_used TEXT,
    
    -- Status tracking
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    error_message TEXT,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT NULL,
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_price_analyses_user_id ON public.price_analyses (user_id);
CREATE INDEX IF NOT EXISTS idx_price_analyses_created_at ON public.price_analyses (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_price_analyses_status ON public.price_analyses (status) WHERE status != 'deleted';
CREATE INDEX IF NOT EXISTS idx_price_analyses_user_created ON public.price_analyses (user_id, created_at DESC);

-- Updated_at trigger
CREATE TRIGGER trg_price_analyses_updated_at BEFORE UPDATE ON public.price_analyses FOR EACH ROW EXECUTE FUNCTION fn_updated_at_value();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop triggers, indexes, and table (reverse order of creation)
DROP TRIGGER IF EXISTS trg_price_analyses_updated_at ON public.price_analyses;
DROP INDEX IF EXISTS idx_price_analyses_user_created;
DROP INDEX IF EXISTS idx_price_analyses_status;
DROP INDEX IF EXISTS idx_price_analyses_created_at;
DROP INDEX IF EXISTS idx_price_analyses_user_id;
DROP TABLE IF EXISTS public.price_analyses;

-- +goose StatementEnd

