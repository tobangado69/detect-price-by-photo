# Password Hash Validation Fix

## Problem

When attempting to sign in via `/api/v1/auth/signin/email` endpoint, the following error occurred:

```
ERR failed to validate password hash op=ValidateUserPassword user_id=9057d65e-b074-4f3c-b656-271ba4ad1a27 error="invalid hash format"
ERR HTTP Request status=500 method=POST request_id=req_01k9rqvsjtej7ttjqazjnha89x path=/api/v1/auth/signin/email
```

## Root Cause

The `password_hash` column in the `user_passwords` table was defined as `BYTEA` type, and the code was using PostgreSQL's `convert_to()` and `convert_from()` functions to store and retrieve password hashes.

**The issue**: Argon2 password hashes use the PHC string format (e.g., `$argon2id$v=19$m=16384,t=4,p=2$...`) which contains Base64-encoded data. When converting this string to BYTEA and back, the encoding was corrupted, causing the hash validation to fail with "invalid hash format" error.

## Solution

Changed the `password_hash` column from `BYTEA` to `TEXT` and removed the `convert_to()`/`convert_from()` conversion functions.

### Changes Made

1. **Migration**: Created `00016_alter_password_hash_to_text.sql`
   - Converts `password_hash` column from `BYTEA` to `TEXT`
   - Preserves existing data using `convert_from(password_hash, 'UTF8')`

2. **Repository**: Updated `apps/backend/internal/user/auth/repository/repo_password.go`
   - Removed `convert_to()` from `SetUserPassword()`
   - Removed `convert_to()` from `UpdateUserPassword()`
   - Removed `convert_from()` from `ValidateUserPassword()`

3. **Password Utils**: Enhanced `apps/backend/internal/utils/password.go`
   - Added whitespace trimming in `Validate()`
   - Improved error messages for debugging

## How to Apply the Fix

### Step 1: Stop the Backend Service

```bash
cd docker
docker-compose down backend
```

### Step 2: Run the Migration

```bash
# Navigate to backend directory
cd apps/backend

# Run migration
go run cmd/server/main.go migrate up
```

Or if using Docker:

```bash
# Start only the database
cd docker
docker-compose up -d postgres

# Run migration inside backend container or locally
docker-compose run --rm backend ./backend migrate up
```

### Step 3: Re-seed User Data (Important!)

After the migration, you need to re-create users because existing password hashes may be corrupted:

```bash
# Clear existing users and re-seed
cd apps/backend

# Option 1: Using the seed command
go run cmd/server/main.go seed

# Option 2: Using the create_admin script
go run scripts/create_admin.go
```

Or via Docker:

```bash
cd docker
docker-compose run --rm backend ./backend seed
```

### Step 4: Restart Services

```bash
cd docker
docker-compose up -d
```

## Testing

Test the fix using the default seeded users:

### Admin User
- **Username**: `admin`
- **Email**: `admin@detectprice.com`
- **Password**: `admin123`

### Regular User
- **Username**: `johndoe`
- **Email**: `johndoe@example.com`
- **Password**: `secure.password`

### Postman Test

```bash
POST http://localhost:9871/api/v1/auth/signin/email
Content-Type: application/json

{
  "email": "admin@detectprice.com",
  "password": "admin123"
}
```

Expected response: `200 OK` with access and refresh tokens.

## Technical Details

### Argon2 PHC Format

The Argon2id algorithm produces hashes in PHC (Password Hashing Competition) string format:

```
$argon2id$v=19$m=16384,t=4,p=2$[salt]$[hash]
```

Where:
- `$argon2id$` - Algorithm identifier
- `v=19` - Argon2 version
- `m=16384` - Memory cost (16 MB)
- `t=4` - Time cost (iterations)
- `p=2` - Parallelism factor
- `[salt]` - Base64-encoded salt
- `[hash]` - Base64-encoded hash

This format is a **string**, not binary data, so it should be stored as `TEXT`, not `BYTEA`.

### Why TEXT is Better Than BYTEA

1. **Simplicity**: No conversion needed, store and retrieve as-is
2. **Compatibility**: PHC format is designed to be a string
3. **Debugging**: Easier to inspect in database tools
4. **Portability**: Can be copied/pasted without encoding issues
5. **Standard Practice**: Most password libraries expect string storage

## Verification

After applying the fix, you can verify the password hashes in the database:

```sql
-- Check password hash format
SELECT 
    user_id,
    LEFT(password_hash, 50) AS hash_preview,
    LENGTH(password_hash) AS hash_length
FROM public.user_passwords;
```

Expected output:
```
user_id                               | hash_preview                                              | hash_length
--------------------------------------|-----------------------------------------------------------|------------
9057d65e-b074-4f3c-b656-271ba4ad1a27 | $argon2id$v=19$m=16384,t=4,p=2$...                       | 93-97
```

The hash should:
- Start with `$argon2id$v=19$`
- Be approximately 90-100 characters long
- Contain only printable ASCII characters

## Rollback (if needed)

If you need to rollback the migration:

```bash
# Rollback the migration
go run cmd/server/main.go migrate down

# Or via Docker
docker-compose run --rm backend ./backend migrate down
```

**Note**: After rollback, you'll need to re-hash all passwords again since the conversion is not reversible without data loss.

## Prevention

To prevent similar issues in the future:

1. **Always use TEXT for password hashes** - They are already encoded strings
2. **Test authentication** immediately after seeding users
3. **Document hash format** in migration comments
4. **Add integration tests** for authentication flow

## Related Files

- `apps/backend/internal/db/migrations/00016_alter_password_hash_to_text.sql` - Migration
- `apps/backend/internal/user/auth/repository/repo_password.go` - Repository methods
- `apps/backend/internal/utils/password.go` - Password hashing utilities
- `apps/backend/internal/db/seeders/user_factory.go` - User seeder

