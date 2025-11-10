-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.ai_models (
    id TEXT NOT NULL PRIMARY KEY, -- e.g., "openai/gpt-4o-mini"
    display_name TEXT NOT NULL CHECK (char_length(display_name) > 0),
    provider TEXT NOT NULL CHECK (char_length(provider) > 0),
    cost_per_1k_tokens_usd NUMERIC(10, 6) NOT NULL DEFAULT 0,
    average_latency_ms INTEGER NOT NULL DEFAULT 0,
    modes_supported TEXT[] NOT NULL DEFAULT '{}', -- e.g., ['fast', 'accurate', 'knowledge_based']
    fallback_chain TEXT[] NOT NULL DEFAULT '{}', -- Array of model IDs to fallback to
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled', 'deprecated')),
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    max_tokens INTEGER NOT NULL DEFAULT 2000,
    temperature NUMERIC(3, 2) NOT NULL DEFAULT 0.7,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT NULL
);

-- Ensure only one default model exists
CREATE UNIQUE INDEX IF NOT EXISTS idx_ai_models_single_default ON public.ai_models (is_default) WHERE is_default = TRUE;

-- Index for status lookups
CREATE INDEX IF NOT EXISTS idx_ai_models_status ON public.ai_models (status);
CREATE INDEX IF NOT EXISTS idx_ai_models_provider ON public.ai_models (provider);
CREATE INDEX IF NOT EXISTS idx_ai_models_updated_at ON public.ai_models (updated_at) WHERE updated_at IS NOT NULL;

CREATE TRIGGER trg_ai_models_updated_at BEFORE UPDATE ON public.ai_models FOR EACH ROW EXECUTE FUNCTION fn_updated_at_value();

-- Seed initial models
INSERT INTO public.ai_models (id, display_name, provider, cost_per_1k_tokens_usd, average_latency_ms, modes_supported, fallback_chain, status, is_default, max_tokens, temperature)
VALUES
    ('openai/gpt-4o-mini', 'GPT-4o mini', 'openai', 0.150, 1800, ARRAY['fast', 'accurate'], ARRAY['anthropic/claude-3-5-haiku'], 'active', TRUE, 2000, 0.7),
    ('anthropic/claude-3-5-haiku', 'Claude 3.5 Haiku', 'anthropic', 0.250, 2200, ARRAY['accurate', 'knowledge_based'], ARRAY['openai/gpt-4o-mini'], 'active', FALSE, 2000, 0.7),
    ('meta-llama/llama-3.1-405b-instruct', 'Llama 3.1 405B', 'meta-llama', 0.100, 3500, ARRAY['knowledge_based'], ARRAY['anthropic/claude-3-5-haiku', 'openai/gpt-4o-mini'], 'active', FALSE, 2000, 0.7)
ON CONFLICT (id) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_ai_models_updated_at ON public.ai_models;
DROP INDEX IF EXISTS idx_ai_models_updated_at;
DROP INDEX IF EXISTS idx_ai_models_provider;
DROP INDEX IF EXISTS idx_ai_models_status;
DROP INDEX IF EXISTS idx_ai_models_single_default;
DROP TABLE IF EXISTS public.ai_models;
-- +goose StatementEnd

