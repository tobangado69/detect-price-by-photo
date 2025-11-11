-- Ensure Role Column Exists (Idempotent)
-- This script will add the role column if it doesn't exist
-- Safe to run multiple times

DO $$
BEGIN
    -- Add role column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' 
          AND table_name = 'users' 
          AND column_name = 'role'
    ) THEN
        ALTER TABLE public.users 
        ADD COLUMN role TEXT NOT NULL DEFAULT 'user' 
        CHECK (role IN ('user', 'admin'));
        
        RAISE NOTICE '✅ Role column added successfully';
    ELSE
        RAISE NOTICE '✅ Role column already exists';
    END IF;
    
    -- Create index if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE tablename = 'users' 
          AND indexname = 'idx_users_role'
    ) THEN
        CREATE INDEX idx_users_role ON public.users (role);
        RAISE NOTICE '✅ Role index created successfully';
    ELSE
        RAISE NOTICE '✅ Role index already exists';
    END IF;
END $$;

-- Verify the role column
SELECT 
    'Role Column Status' as check_type,
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_schema = 'public' 
              AND table_name = 'users' 
              AND column_name = 'role'
        ) THEN '✅ EXISTS'
        ELSE '❌ MISSING'
    END as status;

-- Show current users and their roles
SELECT 
    email,
    username,
    role,
    display_name
FROM public.users
ORDER BY created_at DESC;

