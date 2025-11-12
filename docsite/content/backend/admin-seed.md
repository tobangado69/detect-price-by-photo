---
title: Admin Seed
weight: 4
---

# Admin User Seed Data

## Overview

The seeder creates default users including an admin account for testing and development.

## Admin User Credentials

**Email:** `admin@detectprice.com`  
**Username:** `admin`  
**Password:** `admin123`  
**Role:** `admin`  
**Email Verified:** ✅ Yes

## Regular User Credentials

**Email:** `johndoe@example.com`  
**Username:** `johndoe`  
**Password:** `secure.password`  
**Role:** `user`  
**Email Verified:** ❌ No

## Running the Seeder

### Option 1: Using the migrate:seed command (Recommended)

**Local (requires database connection):**
```bash
cd apps/backend
go run -tags debug cmd/main.go migrate:seed --force
```

**Via Docker:**
```bash
# Build and run the seeder in Docker
docker compose -f compose.yaml run --rm backend \
  go run -tags debug cmd/main.go migrate:seed --force
```

### Option 2: Direct Database Access (SQL)

If you prefer SQL or need to run it directly:

```bash
# Connect to PostgreSQL
docker compose -f compose.yaml exec db psql -U postgres -d detect_price

# Then run the SQL script
\i /app/scripts/seed_admin.sql
```

Or copy the SQL file and run it:
```bash
docker compose -f compose.yaml exec -T db psql -U postgres -d detect_price < apps/backend/scripts/seed_admin.sql
```

### Option 3: Using the create_admin.go script

```bash
cd apps/backend
go run scripts/create_admin.go -email admin@detectprice.com -password admin123 -role admin
```

## Seeder Features

- ✅ **Idempotent**: Running multiple times won't create duplicates
- ✅ **Role Support**: Sets user role (`admin` or `user`)
- ✅ **Password Hashing**: Uses Argon2 for secure password storage
- ✅ **Email Verification**: Admin user is pre-verified
- ✅ **Metadata Support**: Includes timezone and other metadata
- ✅ **Transaction Safe**: All operations in a single transaction

## Code Location

The seeder is located at:
- **File**: `apps/backend/internal/db/seeders/user_factory.go`
- **Function**: `UserFactory(ctx context.Context, pool *pgxpool.Pool) error`
- **Called from**: `apps/backend/internal/db/seeder.go`

## Customization

To modify the seeded users, edit the `users` slice in `user_factory.go`:

```go
users := []UserSeed{
    {
        DisplayName:     "Your Admin Name",
        Email:           "your-admin@example.com",
        Username:        "your-admin",
        Password:        "your-secure-password",
        Role:            "admin",
        Metadata:        map[string]string{"timezone": "Asia/Jakarta"},
        EmailVerifiedAt: &now,
    },
    // Add more users...
}
```

## Security Notes

⚠️ **Important**: 
- Change the default admin password in production
- Never commit production credentials to version control
- Use environment variables for sensitive data in production
- The seeder uses `ON CONFLICT` to update existing users, so passwords can be reset by re-running the seeder

## Testing Admin Access

After seeding, test admin access:

```bash
# Login as admin
curl -X POST http://localhost:8080/api/v1/auth/signin/email \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@detectprice.com",
    "password": "admin123"
  }'

# Use the access_token to access admin endpoints
curl -X GET http://localhost:8080/api/v1/admin/users \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

