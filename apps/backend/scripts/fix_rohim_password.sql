-- Update password for rohimjoy70@gmail.com with proper Argon2 hash
DO $$
DECLARE
    admin_user_id UUID;
    hash_text TEXT := '$argon2id$v=19$m=16384,t=4,p=2$WKE9zkSgLWo$7ISEqV4cttOamLRU6OuCROmAMY1HItf0paeKSTXXGEM';
BEGIN
    -- Get admin user ID
    SELECT id INTO admin_user_id
    FROM public.users
    WHERE email = 'rohimjoy70@gmail.com';

    IF admin_user_id IS NULL THEN
        RAISE EXCEPTION 'User not found';
    END IF;

    -- Update password with proper encoding
    UPDATE public.user_passwords
    SET password_hash = convert_to(hash_text, 'UTF8'),
        updated_at = CURRENT_TIMESTAMP
    WHERE user_id = admin_user_id;

    IF NOT FOUND THEN
        -- Insert if doesn't exist
        INSERT INTO public.user_passwords (user_id, password_hash, created_at)
        VALUES (admin_user_id, convert_to(hash_text, 'UTF8'), CURRENT_TIMESTAMP);
    END IF;

    RAISE NOTICE 'Password updated for user: %', admin_user_id;
END $$;

-- Verify the update
SELECT 
    u.email,
    u.role,
    CASE WHEN up.user_id IS NOT NULL THEN 'Password set' ELSE 'No password' END as password_status,
    length(convert_from(up.password_hash::bytea, 'UTF8')) as hash_length
FROM public.users u
LEFT JOIN public.user_passwords up ON u.id = up.user_id
WHERE u.email = 'rohimjoy70@gmail.com';

