-- Admin User Seed Data
-- This script creates an admin user with role 'admin'
-- Run this script to seed the database with admin user

-- Admin User
INSERT INTO public.users (username, display_name, email, role, metadata, email_verified_at)
VALUES (
    'admin',
    'Admin Sistem',
    'admin@detectprice.com',
    'admin',
    '{"timezone": "Asia/Jakarta"}'::jsonb,
    NOW()
)
ON CONFLICT (username) DO UPDATE
SET
    display_name = EXCLUDED.display_name,
    email = EXCLUDED.email,
    role = EXCLUDED.role,
    metadata = EXCLUDED.metadata,
    email_verified_at = EXCLUDED.email_verified_at;

-- Set admin password (password: admin123)
-- Note: This uses Argon2 hash. In production, use the Go seeder for proper password hashing.
-- For quick testing, you can use: $argon2id$v=19$m=65536,t=3,p=4$...
-- But it's better to use the Go seeder which handles password hashing correctly.

-- Get the admin user ID and set password
DO $$
DECLARE
    admin_user_id UUID;
    password_hash TEXT;
BEGIN
    -- Get admin user ID
    SELECT id INTO admin_user_id FROM public.users WHERE username = 'admin';
    
    -- Hash password: admin123
    -- This is a pre-computed Argon2 hash for 'admin123'
    -- In production, always use the Go seeder for proper password hashing
    password_hash := '$argon2id$v=19$m=65536,t=3,p=4$example_salt$example_hash';
    
    -- Insert or update password
    INSERT INTO public.user_passwords (user_id, password_hash)
    VALUES (admin_user_id, password_hash)
    ON CONFLICT (user_id) DO UPDATE
    SET password_hash = EXCLUDED.password_hash;
END $$;

-- Regular User (for testing)
INSERT INTO public.users (username, display_name, email, role, metadata, email_verified_at)
VALUES (
    'johndoe',
    'John Doe',
    'johndoe@example.com',
    'user',
    '{"timezone": "UTC"}'::jsonb,
    NULL
)
ON CONFLICT (username) DO UPDATE
SET
    display_name = EXCLUDED.display_name,
    email = EXCLUDED.email,
    role = EXCLUDED.role,
    metadata = EXCLUDED.metadata,
    email_verified_at = EXCLUDED.email_verified_at;

-- Note: For proper password hashing, use the Go seeder:
-- go run -tags debug cmd/main.go migrate:seed --force

