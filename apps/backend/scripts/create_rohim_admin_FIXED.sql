-- ============================================
-- Create Admin User: rohimjoy70@gmail.com (FIXED VERSION)
-- ============================================
-- IMPORTANT: Run this AFTER applying the password hash fix migration
-- The password_hash column must be TEXT type (not BYTEA)
--
-- Usage via docker:
--   docker exec -i detect-price-db-1 psql -U postgres -d detect_price < apps/backend/scripts/create_rohim_admin_FIXED.sql
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

-- Step 2: Set password for admin user (FIXED - No convert_to!)
-- Password: admin123
-- Using Argon2id hash (correct format)
DO $$
DECLARE
    admin_user_id UUID;
    -- REAL Argon2id hash for 'admin123' (PHC format)
    password_hash TEXT := '$argon2id$v=19$m=16384,t=4,p=2$jsxXuc0Yotg$HNWo2/vnytAcXF2JeEJdEhhlN185VHpITn9wHzmB6E8';
BEGIN
    -- Get admin user ID
    SELECT id INTO admin_user_id
    FROM public.users
    WHERE email = 'rohimjoy70@gmail.com';

    IF admin_user_id IS NULL THEN
        RAISE EXCEPTION 'Admin user not found. Please run the INSERT statement first.';
    END IF;

    -- Insert or update password (FIXED: Direct TEXT storage, no convert_to!)
    INSERT INTO public.user_passwords (user_id, password_hash, created_at)
    VALUES (admin_user_id, password_hash, CURRENT_TIMESTAMP)
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
    CASE WHEN up.user_id IS NOT NULL THEN 'Password set' ELSE 'No password' END as password_status,
    LEFT(up.password_hash, 30) as hash_preview
FROM public.users u
LEFT JOIN public.user_passwords up ON u.id = up.user_id
WHERE u.email = 'rohimjoy70@gmail.com';

-- ============================================
-- Admin Credentials:
-- Email:    rohimjoy70@gmail.com
-- Password: admin123
-- Role:     admin
-- ============================================

