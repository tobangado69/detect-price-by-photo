-- ============================================
-- Update Rohim's Password (After Fix Applied)
-- ============================================
-- Run this AFTER running: make fix-password
-- This updates only the password hash to the correct format
-- ============================================

DO $$
DECLARE
    admin_user_id UUID;
BEGIN
    -- Get Rohim's user ID
    SELECT id INTO admin_user_id
    FROM public.users
    WHERE email = 'rohimjoy70@gmail.com';

    IF admin_user_id IS NULL THEN
        RAISE EXCEPTION 'User rohimjoy70@gmail.com not found!';
    END IF;

    -- Update password to use correct TEXT format (no convert_to!)
    -- This is a REAL Argon2id hash for password: admin123
    UPDATE public.user_passwords
    SET 
        password_hash = '$argon2id$v=19$m=16384,t=4,p=2$jsxXuc0Yotg$HNWo2/vnytAcXF2JeEJdEhhlN185VHpITn9wHzmB6E8',
        updated_at = CURRENT_TIMESTAMP
    WHERE user_id = admin_user_id;

    RAISE NOTICE 'Password updated for rohimjoy70@gmail.com';
    RAISE NOTICE 'You can now login with: rohimjoy70@gmail.com / admin123';
END $$;

-- Verify the update
SELECT 
    u.email,
    u.username,
    u.role,
    LEFT(up.password_hash, 50) as hash_preview,
    LENGTH(up.password_hash) as hash_length
FROM public.users u
JOIN public.user_passwords up ON u.id = up.user_id
WHERE u.email = 'rohimjoy70@gmail.com';

