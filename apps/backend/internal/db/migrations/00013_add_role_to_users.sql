-- +goose Up
-- +goose StatementBegin
-- Add role column to users table
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin'));

-- Create index for role lookups
CREATE INDEX IF NOT EXISTS idx_users_role ON public.users (role);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_role;
ALTER TABLE public.users DROP COLUMN IF EXISTS role;
-- +goose StatementEnd

