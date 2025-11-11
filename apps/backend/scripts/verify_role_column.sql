-- Verify Role Column Exists and Works
-- Run this to verify the role column is properly set up

-- 1. Check if role column exists
SELECT 
    column_name, 
    data_type, 
    column_default, 
    is_nullable
FROM information_schema.columns 
WHERE table_schema = 'public' 
  AND table_name = 'users' 
  AND column_name = 'role';

-- 2. Check role constraint
SELECT 
    constraint_name, 
    constraint_type,
    check_clause
FROM information_schema.check_constraints 
WHERE constraint_name = 'users_role_check';

-- 3. Check role index
SELECT 
    indexname, 
    indexdef
FROM pg_indexes 
WHERE tablename = 'users' 
  AND indexname = 'idx_users_role';

-- 4. Show all users with their roles
SELECT 
    id,
    email,
    username,
    role,
    display_name,
    email_verified_at IS NOT NULL as email_verified
FROM public.users
ORDER BY created_at DESC;

-- 5. Count users by role
SELECT 
    role,
    COUNT(*) as count
FROM public.users
GROUP BY role;

-- 6. Verify migration was applied
SELECT 
    version_id,
    is_applied,
    tstamp
FROM app_migrations
WHERE version_id = 13;

