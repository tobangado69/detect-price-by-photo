# Create Admin Account

This guide shows you how to create an admin account in the database.

## Method 1: Using SQL Script (Recommended)

### Via Docker:
```bash
docker exec -i detect-price-db-1 psql -U postgres -d detect_price < apps/backend/scripts/create_admin.sql
```

### Via pgweb (Web UI):
1. Open pgweb: http://localhost:8081
2. Navigate to SQL Editor
3. Copy and paste the contents of `apps/backend/scripts/create_admin.sql`
4. Click "Run"

### Via psql:
```bash
psql -U postgres -d detect_price -f apps/backend/scripts/create_admin.sql
```

## Method 2: Using Go Script

Make sure your database is running and `.env` is configured:

```bash
cd apps/backend
go run scripts/create_admin.go -email admin@detectprice.com -password yourpassword -name "Admin Name"
```

## Method 3: Manual SQL (Quick)

Connect to your database and run:

```sql
-- Create admin user
INSERT INTO public.users (id, display_name, email, username, role, metadata, created_at, email_verified_at)
VALUES (
    gen_random_uuid(),
    'Admin User',
    'admin@detectprice.com',
    'admin',
    'admin',
    '{"timezone": "Asia/Jakarta"}'::jsonb,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (email) DO UPDATE SET role = 'admin';

-- Set password (hash for 'admin123')
-- Replace the hash below with your own bcrypt hash if you want a different password
INSERT INTO public.user_passwords (user_id, password_hash, created_at)
SELECT id, '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'::bytea, CURRENT_TIMESTAMP
FROM public.users WHERE email = 'admin@detectprice.com'
ON CONFLICT (user_id) DO UPDATE SET password_hash = EXCLUDED.password_hash;
```

## Default Credentials

After running the script:
- **Email:** `admin@detectprice.com`
- **Password:** `admin123`
- **Role:** `admin`

⚠️ **Important:** Change the password after first login!

## Verify Admin Account

```sql
SELECT id, email, display_name, role, email_verified_at
FROM public.users
WHERE email = 'admin@detectprice.com';
```

## Generate Custom Password Hash

If you want to use a different password, generate a bcrypt hash:

**Using Go:**
```bash
go run -c 'import "golang.org/x/crypto/bcrypt"; import "fmt"; hash, _ := bcrypt.GenerateFromPassword([]byte("your_password"), 10); fmt.Println(string(hash))'
```

**Online tool:**
- Visit: https://bcrypt-generator.com/
- Enter your password
- Copy the hash
- Replace the hash in the SQL script

## Test Admin Login

After creating the admin account, test it:

```bash
curl -X POST http://localhost:8000/api/v1/auth/signin/email \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@detectprice.com",
    "password": "admin123"
  }'
```

Save the `access_token` from the response and use it for admin endpoints:

```bash
curl -X GET http://localhost:8000/api/v1/admin/users \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

