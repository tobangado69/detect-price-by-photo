-- ============================================
-- Create Admin User: rohimjoy70@gmail.com
-- ============================================
-- Run this SQL script directly in your PostgreSQL database
-- You can run it via: psql, pgAdmin, or docker exec
--
-- Usage via docker:
--   docker exec -i detect-price-db-1 psql -U postgres -d detect_price < apps/backend/scripts/create_rohim_admin.sql
-- ============================================

-- Step 1: Create admin user (or update existing to admin)
INSERT INTO public.users (
    id,
    display_name,
    email,
    username,
    role,
    metadata,
    created_at,
    email_verified_at
) VALUES (
    gen_random_uuid(),
    'Rohim Joy',
    'rohimjoy70@gmail.com',
    'rohimjoy70',
    'admin',
    '{"timezone": "Asia/Jakarta"}'::jsonb,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (email) DO UPDATE
SET 
    role = 'admin',
    email_verified_at = COALESCE(users.email_verified_at, CURRENT_TIMESTAMP),
    display_name = EXCLUDED.display_name
RETURNING id, email, role, display_name;

-- Step 2: Set password for admin user
-- Password hash for 'admin123' (bcrypt, cost 10)
DO $$
DECLARE
    admin_user_id UUID;
    password_hash TEXT := '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'; -- admin123
BEGIN
    -- Get admin user ID
    SELECT id INTO admin_user_id
    FROM public.users
    WHERE email = 'rohimjoy70@gmail.com';

    IF admin_user_id IS NULL THEN
        RAISE EXCEPTION 'Admin user not found. Please run the INSERT statement first.';
    END IF;

    -- Insert or update password
    INSERT INTO public.user_passwords (user_id, password_hash, created_at)
    VALUES (admin_user_id, convert_to(password_hash, 'UTF8'), CURRENT_TIMESTAMP)
    ON CONFLICT (user_id) DO UPDATE
    SET password_hash = EXCLUDED.password_hash,
        updated_at = CURRENT_TIMESTAMP;

    RAISE NOTICE 'Password set for admin user: %', admin_user_id;
END $$;

-- Step 3: Verify the admin user was created successfully
SELECT 
    u.id,
    u.email,
    u.display_name,
    u.username,
    u.role,
    u.email_verified_at IS NOT NULL as email_verified,
    u.created_at,
    CASE WHEN up.user_id IS NOT NULL THEN 'Password set' ELSE 'No password' END as password_status
FROM public.users u
LEFT JOIN public.user_passwords up ON u.id = up.user_id
WHERE u.email = 'rohimjoy70@gmail.com';

-- ============================================
-- Admin Credentials:
-- Email:    rohimjoy70@gmail.com
-- Password: admin123
-- Role:     admin
-- ============================================

