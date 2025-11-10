-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- Create Market Knowledge table - RAG knowledge base with pgvector embeddings
-- ============================================================================

-- Ensure pgvector extension is enabled
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS public.market_knowledge (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    product_name TEXT NOT NULL CHECK (char_length(product_name) > 0),
    condition TEXT NOT NULL CHECK (condition IN ('new', 'like_new', 'good', 'fair', 'poor')),
    embedding vector(1536) NOT NULL,
    market_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    source TEXT,
    last_updated TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_market_knowledge_product_name ON public.market_knowledge (product_name);
CREATE INDEX IF NOT EXISTS idx_market_knowledge_condition ON public.market_knowledge (condition);
CREATE INDEX IF NOT EXISTS idx_market_knowledge_created_at ON public.market_knowledge (created_at);

-- HNSW index for vector similarity search (pgvector)
-- This enables fast cosine similarity searches
CREATE INDEX IF NOT EXISTS idx_market_knowledge_embedding_hnsw ON public.market_knowledge 
    USING hnsw (embedding vector_cosine_ops);

-- Composite index for product + condition lookups
CREATE INDEX IF NOT EXISTS idx_market_knowledge_product_condition ON public.market_knowledge (product_name, condition);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop indexes and table (reverse order of creation)
DROP INDEX IF EXISTS idx_market_knowledge_product_condition;
DROP INDEX IF EXISTS idx_market_knowledge_embedding_hnsw;
DROP INDEX IF EXISTS idx_market_knowledge_created_at;
DROP INDEX IF EXISTS idx_market_knowledge_condition;
DROP INDEX IF EXISTS idx_market_knowledge_product_name;
DROP TABLE IF EXISTS public.market_knowledge;

-- Note: We don't drop the vector extension as it might be used by other tables

-- +goose StatementEnd

