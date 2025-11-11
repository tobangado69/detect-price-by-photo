-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- Change password_hash column from BYTEA to TEXT
-- This fixes the "invalid hash format" error when validating passwords
-- ============================================================================

-- First, convert existing BYTEA data to TEXT if any exists
-- This handles any existing password hashes that were stored as BYTEA
ALTER TABLE public.user_passwords 
    ALTER COLUMN password_hash TYPE TEXT 
    USING convert_from(password_hash, 'UTF8');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revert back to BYTEA (this will require re-hashing all passwords in production)
ALTER TABLE public.user_passwords 
    ALTER COLUMN password_hash TYPE BYTEA 
    USING convert_to(password_hash, 'UTF8');

-- +goose StatementEnd

